#!/usr/bin/env bash
set -euo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
P="$(cd "$HERE/../../.." && pwd)/promise.yaml"
fail() { echo "FAIL: $1"; exit 1; }

[ "$(yq e '.spec.api.spec.names.kind' "$P")" = "NamespaceClaim" ] || fail "wrong CRD kind"
[ "$(yq e '.spec.api.metadata.name' "$P")" = "namespaceclaims.marketplace.kratix.io" ] || fail "wrong CRD name"
[ "$(yq e '.spec.api.spec.versions[0].schema.openAPIV3Schema.properties.spec.required[0]' "$P")" = "namespaceName" ] || fail "namespaceName not required"
# Both configure and delete workflows exist; the script dispatches on the
# Kratix workflow action, so neither needs a command override.
[ "$(yq e '.spec.workflows.resource.configure | length' "$P")" -ge 1 ] || fail "no configure workflow"
[ "$(yq e '.spec.workflows.resource.delete | length' "$P")" -ge 1 ] || fail "no delete workflow"

# Both workflows run the same pipeline image.
configure_img="$(yq e '.spec.workflows.resource.configure[0].spec.containers[0].image' "$P")"
delete_img="$(yq e '.spec.workflows.resource.delete[0].spec.containers[0].image' "$P")"
[[ "$configure_img" == *namespace-claim-configure-pipeline* ]] || fail "configure image wrong"
[ "$configure_img" = "$delete_img" ] || fail "configure and delete must use the same image"

rbac_resources="$(yq e '.. | select(has("permissions")) | .permissions[].resources[]' "$P")"
grep -qx namespaces <<<"$rbac_resources" || fail "rbac missing namespaces (provider RR)"
grep -qx namespaceclaims <<<"$rbac_resources" || fail "rbac missing namespaceclaims (sibling list)"

echo "ALL TESTS PASSED"
