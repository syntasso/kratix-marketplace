# Namespace claim (ref-counted Namespace-as-a-Service)

A claim-based promise: many consumers request the *same* shared namespace, and
the namespace is created once and deleted only when the **last** claim is gone.

## How it works

- Each consumer creates a `NamespaceClaim` naming `spec.namespaceName`.
- The configure pipeline idempotently creates a provider request to the existing
  `namespace` promise (named after `namespaceName`), which materialises the real
  Namespace on the worker. The provider request is owned by nothing, so it
  survives any single claim being deleted.
- The delete pipeline counts live claims for that namespace; when none remain it
  deletes the provider request, which tears the Namespace down.

## Try it

    kubectl apply -f resource-request.yaml          # first claim -> namespace created
    kubectl apply -f - <<EOF                         # second claim, same namespace
    apiVersion: marketplace.kratix.io/v1alpha1
    kind: NamespaceClaim
    metadata: { name: team-b-claim, namespace: default }
    spec: { namespaceName: shared-team-namespace }
    EOF
    kubectl delete namespaceclaim team-a-claim       # namespace stays (team-b still claims it)
    kubectl delete namespaceclaim team-b-claim       # last claim gone -> namespace removed

## Reusing this pattern

The ref-counting in `internal/configure-pipeline/claim` is generic. To make
another resource claimable (DB server, Kafka cluster), copy the `claim` script,
adjust the `CLAIM_KIND` / `TARGET_FIELD` defaults, and reimplement its
`ensure_shared` / `destroy_shared` functions.

## Known limitation

A last-claim release racing a brand-new claim (on the same namespace name) can
delete the namespace just after the new claim's pipeline already ran. Recovery
is not automatic: Kratix re-runs a claim's pipeline on claim/promise change, not
when the provider request is deleted out-of-band, so the namespace stays gone
until the claim is re-applied. A finalizer/controller guard is the hardening
path — see the design spec.
