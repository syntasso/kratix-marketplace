# Approval Server: example backing service

The [`slack-approval-gate` stage](../README.md) doesn't talk to Slack or store any state itself. It delegates all of that to a small **backing service** that it reaches over HTTP. This directory is a **reference implementation** of that service, built as a **Cloudflare Worker** with **Workers KV** for state.

You are **not required to use Cloudflare**: the stage only depends on the HTTP contract below. Any service that honours it (a Worker, a container behind an Ingress, a Lambda, a plain Node/Go server, …) will work. Cloudflare Workers just happen to give you a free, always-public HTTPS endpoint (which Slack's button callbacks need) with no server to run.

## The contract

The stage calls two endpoints, authenticated with a shared bearer token (`APPROVAL_SERVER_TOKEN`):

| Method & path | Auth | Purpose |
|---|---|---|
| `POST /approvals` | `Authorization: Bearer <token>` | Create a request and kick off the human approval (here: post the Slack message). Body: `{ "promiseName", "objectName", "spec" }`, where `spec` is the resource's whole spec object (rendered generically into the message). **Idempotent**: if the request already exists, return its current state without re-notifying. Returns `{ "state", "created" }`. |
| `GET /approvals/{id}` | `Authorization: Bearer <token>` | Return `{ "state": "pending" \| "approved" \| "denied" }`, or `404` if unknown. |
| `DELETE /approvals/{id}` | `Authorization: Bearer <token>` | Delete the record (used by `scripts/reset-approval.sh`). |

The `id` is `approval:{promiseName}:{objectName}`.

Beyond serving the stage, the service is responsible for the **human decision loop** (however it chooses to implement it) and for flipping a request's state to `approved`/`denied`. In this implementation that's a Slack endpoint:

| Method & path | Auth | Purpose |
|---|---|---|
| `POST /slack/interactions` | Slack request signature | Slack button callback. Verifies the signature, writes the decision to KV (only if still `pending`), updates the original message, and DMs the requester. |

**If you implement your own service**, you only need the three `/approvals` endpoints plus whatever decision mechanism you like; the stage never knows or cares how the decision was made.

## How this implementation works

- **State:** one JSON record per key in Workers KV, `{ state, resource, promise, spec, sentAt, decidedBy?, decidedAt? }`, with a 7-day TTL. No separate database.
- **Generic rendering:** the approval message lists every field in the resource `spec`. If `spec.requester` is an email, the decision DM goes to that user; otherwise no DM is sent.
- **Security:** `/approvals*` require the bearer token; `/slack/interactions` is authenticated by Slack's HMAC request signature (with a 5-minute replay window) instead.
- **Correctness:** the Slack message is sent *before* `pending` is persisted (a Slack failure returns 502 and writes nothing, so the stage retries cleanly); decisions only overwrite a `pending` record (double-click / already-decided guard); the requester DM is best-effort and never fails the callback.

Source is a single file: [`src/index.js`](src/index.js).

## Deploy on Cloudflare

### 1. Slack app

At [api.slack.com/apps](https://api.slack.com/apps) → **Create New App → From scratch**. Under **OAuth & Permissions → Bot Token Scopes** add:

| Scope | Purpose |
|---|---|
| `chat:write` | Send approval messages and DMs |
| `chat:write.public` | Send to channels the bot hasn't joined |
| `users:read.email` | Resolve the requester's email to a Slack user for the DM |

Install to your workspace, then copy the **Bot User OAuth Token** (`xoxb-…`) and the **Signing Secret** (Basic Information). Invite the bot to your approval channel: `/invite @<bot-name>`.

### 2. Worker

From the repo root, fill in `.envrc` (see `.envrc.example`) with the Slack values and a random `APPROVAL_SERVER_TOKEN`, then:

```bash
source .envrc
cd approval-server
wrangler login
wrangler kv namespace create APPROVALS
# paste the printed id into wrangler.toml under [[kv_namespaces]] -> id
./deploy.sh
```

`deploy.sh` sets four Worker secrets and deploys:

| Secret | Source |
|---|---|
| `SLACK_SIGNING_SECRET` | Slack app → Basic Information |
| `SLACK_BOT_TOKEN` | Slack app → OAuth & Permissions (`xoxb-…`) |
| `SLACK_APPROVAL_CHANNEL` | e.g. `#platform-approvals` |
| `APPROVAL_SERVER_TOKEN` | the shared bearer token (must match what the stage sends) |

plus the `APPROVALS` KV namespace binding from `wrangler.toml`.

Copy the printed Worker URL; it's your `APPROVAL_SERVER_URL`.

### 3. Wire Slack to the Worker

In the Slack app → **Interactivity & Shortcuts** → enable → set **Request URL** to `<worker-url>/slack/interactions` → Save.

## Local development

`wrangler dev --local` runs the Worker with a local KV simulation, so you can exercise `create → get → delete` and a signed interaction without deploying. Put local values in `.dev.vars` (git-ignored):

```
SLACK_SIGNING_SECRET=dev-signing-secret
SLACK_BOT_TOKEN=xoxb-your-test-bot-token
SLACK_APPROVAL_CHANNEL=#your-test-channel
APPROVAL_SERVER_TOKEN=local-dev-token
```

The Slack *send* and DM paths still call the real Slack API, so they need a valid bot token and channel.

## Notes / limitations

- **Workers KV is eventually consistent:** after a decision, a `GET` can keep returning the previous state for up to ~60s (edge read cache + write propagation). Fine when the caller polls on an interval; if you need read-your-writes consistency, use Durable Objects or a strongly-consistent store instead.
- **State is transient:** the 7-day KV TTL is not an audit trail. For durable history, write decisions to a real datastore.
- **Single approver:** first click wins; there's no multi-approver flow.
