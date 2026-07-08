# Slack Approval Gate: a Kratix pipeline stage

A reusable **pipeline stage** that pauses any Kratix Promise's resource workflow for a human approval in Slack. Drop it in as the last container of your `resource.configure` pipeline: when someone requests a resource, an approval message with **Approve** / **Deny** buttons is posted to Slack, and nothing is provisioned until someone clicks. The requester gets a Slack DM with the decision.

The stage itself is thin. It holds no Slack or state-store credentials, and talks only to a **backing service** over HTTP: "create this approval request" and, on each retry, "has it been decided yet?". Everything else (posting to Slack, storing state, handling the button click) lives in that service.

This repo ships an **example backing service** (a Cloudflare Worker), but it's **not required**: any HTTP service that honours the [contract](#the-backing-service) works. See [`approval-server/README.md`](approval-server/README.md).

## What's in this repo

| Path | What it is |
|---|---|
| `approval-gate/` | The stage: a small Python script (built on the [Kratix Python SDK](https://github.com/syntasso/kratix-python)) + Dockerfile, published as `ghcr.io/syntasso/kratix-marketplace/slack-approval-gate:v0.1.0`. |
| `approval-server/` | An **example** backing service (Cloudflare Worker; state in KV, all Slack interaction). Swap it for your own; see its [README](approval-server/README.md). |
| `promise/` | A worked example: a `redis` Promise with the gate wired in, plus sample resource requests. |
| `scripts/` | Helpers to deploy the example on KinD and to test the stage locally. |

---

## Using the stage in your Promise

This is the part you care about if you found this in the marketplace. Two prerequisites first:

1. A **backing service is running**. See [The backing service](#the-backing-service) below (deployed once, then reused).
2. The **`approval-gate-secrets` Secret exists** in the namespace where your pipelines run:

   ```yaml
   apiVersion: v1
   kind: Secret
   metadata:
     name: approval-gate-secrets
     namespace: kratix-platform-system   # your pipeline namespace
   type: Opaque
   stringData:
     server.url: "https://<your-backing-service>"
     server.token: "<the shared bearer token>"
   ```

Then add the gate as the **last container** in your Promise's `spec.workflows.resource.configure` pipeline, after the containers that write your resource to `/kratix/output`:

```yaml
containers:
  # ... your existing configure container(s) ...

  - image: ghcr.io/syntasso/kratix-marketplace/slack-approval-gate:v0.1.0
    name: slack-approval-gate
    env:
      - name: RETRY_INTERVAL
        value: "20s"                     # how often Kratix re-checks for a decision
      - name: APPROVAL_SERVER_URL
        valueFrom:
          secretKeyRef:
            name: approval-gate-secrets
            key: server.url
      - name: APPROVAL_SERVER_TOKEN
        valueFrom:
          secretKeyRef:
            name: approval-gate-secrets
            key: server.token
```

That's the whole integration. The stage holds **no** Slack or state-store credentials, only the backing-service URL and a shared bearer token.

### Configuration reference

| Env var | Required | Purpose |
|---|---|---|
| `APPROVAL_SERVER_URL` | yes | Base URL of the backing service. |
| `APPROVAL_SERVER_TOKEN` | yes | Shared bearer token; must match the backing service's token. |
| `RETRY_INTERVAL` | yes | Delay Kratix waits before re-running the pipeline while pending (e.g. `20s`). |

### What the stage sends

The stage forwards the resource's **entire `spec`** to the backing service, which renders every field into the approval message. It hardcodes no promise-specific fields, so it works with any CRD unchanged.

One optional convention: if the spec contains a `requester` field holding an email address, the backing service DMs that user when a decision is made (skipped silently otherwise). Add it to your CRD if you want requester notifications.

---

## How it works

The gate runs **after** your resource config has been written to `/kratix/output`, and it controls whether that output is ever scheduled to a destination:

- **first run**: POSTs a "create" to the backing service (which starts the approval and records `pending`), clears `/kratix/output` and `/kratix/metadata`, writes a `workflow-control.yaml` with `retryAfter`, and exits 0. Nothing is scheduled.
- **pending**: clears `/kratix/output` and `/kratix/metadata`, re-writes `workflow-control.yaml`, exits 0. Kratix reschedules after `RETRY_INTERVAL`.
- **approved**: leaves `/kratix/output` intact and exits 0. Kratix schedules the resources normally.
- **denied**: clears `/kratix/output` and `/kratix/metadata`, exits 0. Nothing is scheduled.

The gate always exits 0. It gates provisioning by *clearing or keeping the output*, not by failing.

### The Kratix retry mechanism

While waiting, the gate writes:

```yaml
# /kratix/metadata/workflow-control.yaml
message: "Waiting for approval"
retryAfter: <RETRY_INTERVAL>
```

Kratix reads this after the pipeline exits and reschedules the next run accordingly.

### The decision round-trip

```
gate → POST /approvals (create)  → service starts the approval (e.g. posts to Slack),
                                    stores "pending"
approver decides                 → service records the decision (approved/denied),
                                    notifies the requester
gate → GET /approvals/{id}       → reads the decision on the next retry → continues or stops
```

---

## The backing service

The stage depends on a small HTTP service that:

- accepts `POST /approvals` (create + start the approval) and `GET /approvals/{id}` (return `pending`/`approved`/`denied`), both authenticated with the shared bearer token;
- drives whatever human-decision mechanism you want and flips the request to `approved`/`denied`.

This repo ships a ready-to-use implementation as a **Cloudflare Worker** (state in Workers KV, decisions via Slack buttons). It's the recommended starting point and what the example and local tests use. **You are not tied to Cloudflare**: reimplement the contract in any runtime if you prefer.

**→ See [`approval-server/README.md`](approval-server/README.md)** for the full contract, the architecture, and how to deploy the example on Cloudflare (Slack app, KV namespace, secrets, Interactivity URL).

Once it's deployed, put its URL and token in the `approval-gate-secrets` Secret shown in [Using the stage](#using-the-stage-in-your-promise).

---

## Example: the `redis` Promise

`promise/promise.yaml` is a complete Promise with the gate wired in as the last container of `resource.configure`. To try it on a local KinD cluster (with Kratix installed):

```bash
source .envrc
./scripts/deploy.sh                              # build+load the gate image, create the Secret, apply the Promise
kubectl apply -f promise/resource-request.yaml   # request a redis instance
```

The pipeline holds until you Approve in Slack, then provisions.

## Testing the stage locally (no cluster)

```bash
source .envrc
./scripts/test-local.sh
```

This simulates the Kratix retry loop: on the first run it starts the approval and polls every 10 seconds. Click Approve or Deny in Slack, and the loop detects the decision and exits.

Note: with the example Cloudflare Worker, after you click, the loop can keep printing `pending` for up to ~60 seconds before it flips (Workers KV eventual consistency, see the [server README](approval-server/README.md#notes--limitations)), not a hang.

Reset state and re-run:

```bash
./scripts/reset-approval.sh
./scripts/test-local.sh
```

### Troubleshooting local testing

- **"approval server unreachable" / connection refused**: `APPROVAL_SERVER_URL` is wrong or the backing service isn't reachable. Confirm the service is deployed and copy its exact URL into `.envrc`.
- **Decision takes a while to register**: with the example Worker this is Workers KV read caching (up to ~60s). See the [server README](approval-server/README.md#notes--limitations).

## Limitations (prototype)

- **Single approver**: the first person to decide wins. No multi-approver workflow.
- **Bot must be invited to the approval channel** (example Slack service): run `/invite @<bot-name>` before testing.
- **Backing-service URL must be reachable before decisions work**: with the example Worker, the Slack Interactivity Request URL must be configured, or clicking a button shows a generic Slack error.
- Additional properties of the example backing service (state TTL, KV consistency) are documented in [`approval-server/README.md`](approval-server/README.md#notes--limitations).
