#!/usr/bin/env bash
set -euo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
PIPELINE_DIR="$(cd "$HERE/.." && pwd)"
CLAIM="$PIPELINE_DIR/claim"

fail() { echo "FAIL: $1"; exit 1; }
pass() { echo "PASS: $1"; }

# Isolated workspace with a PATH-stubbed kubectl that records what it's asked to
# do: `get -o json` echoes a fixture, `apply -f -` captures stdin, `delete`
# appends its args to a log. yq is the real binary (parses the fixture JSON).
setup() {
  WORK="$(mktemp -d)"
  mkdir -p "$WORK/bin"
  DELETE_LOG="$WORK/deleted"
  APPLY_CAPTURE="$WORK/applied.yaml"
  KUBECTL_GET_JSON="$WORK/get.json"
  KRATIX_INPUT="$WORK/object.yaml"
  : > "$DELETE_LOG"

  cat > "$WORK/bin/kubectl" <<EOF
#!/usr/bin/env bash
mode=""
for a in "\$@"; do
  [ "\$a" = "get" ] && mode=get
  [ "\$a" = "apply" ] && mode=apply
  [ "\$a" = "delete" ] && mode=delete
done
[ "\$mode" = "get" ] && { cat "$KUBECTL_GET_JSON"; exit 0; }
[ "\$mode" = "apply" ] && { cat > "$APPLY_CAPTURE"; exit 0; }
[ "\$mode" = "delete" ] && { echo "\$*" >> "$DELETE_LOG"; exit 0; }
exit 0
EOF
  chmod +x "$WORK/bin/kubectl"

  cat > "$KRATIX_INPUT" <<'EOF'
apiVersion: marketplace.kratix.io/v1alpha1
kind: NamespaceClaim
metadata:
  name: claim-a
spec:
  namespaceName: shared-ns
EOF
}

teardown() { rm -rf "$WORK"; }

run_claim() { # $1 = flag; any further args are extra VAR=val env assignments
  flag="$1"; shift
  env PATH="$WORK/bin:$PATH" \
    CLAIM_KIND="namespaceclaims.marketplace.kratix.io" \
    TARGET_FIELD=".spec.namespaceName" \
    KRATIX_INPUT="$KRATIX_INPUT" \
    "$@" \
    sh "$CLAIM" "$flag"
}

# Case 1: another live claim exists -> nothing deleted
setup
cat > "$KUBECTL_GET_JSON" <<'EOF'
{"items":[
  {"metadata":{"name":"claim-a","deletionTimestamp":"2026-06-16T00:00:00Z"},"spec":{"namespaceName":"shared-ns"}},
  {"metadata":{"name":"claim-b","deletionTimestamp":null},"spec":{"namespaceName":"shared-ns"}}
]}
EOF
run_claim --release
[ -s "$DELETE_LOG" ] && fail "deleted shared resource while a live claim remains" || pass "keeps namespace when a sibling claim is live"
teardown

# Case 2: only this (terminating) claim remains -> delete provider RR by target
setup
cat > "$KUBECTL_GET_JSON" <<'EOF'
{"items":[
  {"metadata":{"name":"claim-a","deletionTimestamp":"2026-06-16T00:00:00Z"},"spec":{"namespaceName":"shared-ns"}}
]}
EOF
run_claim --release
grep -q "namespace.marketplace.kratix.io" "$DELETE_LOG" || fail "did not delete the provider RR type"
grep -q "shared-ns" "$DELETE_LOG" || fail "deleted wrong target"
pass "deletes namespace when last claim is released"
teardown

# Case 3: claims for a DIFFERENT namespace do not keep this one alive
setup
cat > "$KUBECTL_GET_JSON" <<'EOF'
{"items":[
  {"metadata":{"name":"claim-a","deletionTimestamp":"2026-06-16T00:00:00Z"},"spec":{"namespaceName":"shared-ns"}},
  {"metadata":{"name":"claim-x","deletionTimestamp":null},"spec":{"namespaceName":"other-ns"}}
]}
EOF
run_claim --release
grep -q "shared-ns" "$DELETE_LOG" || fail "did not delete; unrelated namespace miscounted"
pass "ignores claims targeting a different namespace"
teardown

# Case 4: --ensure applies a provider Namespace RR named after the target
setup
echo '{"items":[]}' > "$KUBECTL_GET_JSON"
run_claim --ensure
grep -q "kind: namespace" "$APPLY_CAPTURE" || fail "ensure did not apply a namespace provider RR"
[ "$(yq e '.metadata.name' "$APPLY_CAPTURE")" = "shared-ns" ] || fail "provider RR not named after target"
[ "$(yq e '.spec.namespaceName' "$APPLY_CAPTURE")" = "shared-ns" ] || fail "provider RR missing namespaceName"
pass "ensure applies a deterministically-named provider RR"
teardown

# Case 5: --ensure passes clusterName through when provided
setup
echo '{"items":[]}' > "$KUBECTL_GET_JSON"
run_claim --ensure CLUSTER_NAME=worker-1
[ "$(yq e '.spec.clusterName' "$APPLY_CAPTURE")" = "worker-1" ] || fail "clusterName not passed through"
pass "ensure passes clusterName through"
teardown

echo "ALL TESTS PASSED"
