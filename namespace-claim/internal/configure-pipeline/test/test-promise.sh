#!/usr/bin/env bash
set -euo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
P="$(cd "$HERE/../../.." && pwd)/promise.yaml"
fail() { echo "FAIL: $1"; exit 1; }

[ "$(yq e '.spec.api.spec.names.kind' "$P")" = "NamespaceClaim" ] || fail "wrong CRD kind"
[ "$(yq e '.spec.api.metadata.name' "$P")" = "namespaceclaims.marketplace.kratix.io" ] || fail "wrong CRD name"
[ "$(yq e '.spec.api.spec.versions[0].schema.openAPIV3Schema.properties.spec.required[0]' "$P")" = "namespaceName" ] || fail "namespaceName not required"
[ "$(yq e '.spec.workflows.resource.configure | length' "$P")" -ge 1 ] || fail "no configure workflow"
[ "$(yq e '.spec.workflows.resource.delete | length' "$P")" -ge 1 ] || fail "no delete workflow"

delete_cmd="$(yq e '.spec.workflows.resource.delete[0].spec.containers[0].command[]' "$P")"
[[ "$delete_cmd" == *claim* && "$delete_cmd" == *--release* ]] || fail "delete does not run claim --release"

rbac_resources="$(yq e '.. | select(has("permissions")) | .permissions[].resources[]' "$P")"
grep -qx namespaces <<<"$rbac_resources" || fail "rbac missing namespaces"
grep -qx namespaceclaims <<<"$rbac_resources" || fail "rbac missing namespaceclaims"

# both pipeline containers must export the env the generic claim.sh requires
env_names="$(yq e '.. | select(has("containers")) | .containers[].env[].name' "$P")"
[ "$(grep -cx CLAIM_KIND <<<"$env_names")" -ge 2 ] || fail "CLAIM_KIND env missing from a pipeline container"
[ "$(grep -cx TARGET_FIELD <<<"$env_names")" -ge 2 ] || fail "TARGET_FIELD env missing from a pipeline container"

echo "ALL TESTS PASSED"
