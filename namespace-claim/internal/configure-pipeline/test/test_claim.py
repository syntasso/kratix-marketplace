#!/usr/bin/env python3
"""Unit tests for the pure ref-counting logic in claim.py.

Imports only the dependency-free helpers (count_live_claims, _get_path,
target_field), so it runs with plain python3 -- no kratix_sdk or kubernetes
needed. The Kubernetes I/O and SDK wiring are covered by the integration test
(internal/scripts/test).
"""
import importlib.util
import os
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
CLAIM_PY = os.path.join(HERE, "..", "claim.py")

spec = importlib.util.spec_from_file_location("claim", CLAIM_PY)
claim = importlib.util.module_from_spec(spec)
spec.loader.exec_module(claim)

FIELD = "spec.namespaceName"


def item(name, ns, terminating=False):
    meta = {"name": name}
    if terminating:
        meta["deletionTimestamp"] = "2026-06-17T00:00:00Z"
    return {"metadata": meta, "spec": {"namespaceName": ns}}


failures = []


def check(name, cond):
    print(("PASS: " if cond else "FAIL: ") + name)
    if not cond:
        failures.append(name)


# A live sibling claim keeps the count > 0 (terminating self excluded).
check(
    "keeps count when a sibling claim is live",
    claim.count_live_claims(
        [item("claim-a", "shared-ns", terminating=True), item("claim-b", "shared-ns")],
        FIELD, "shared-ns") == 1,
)

# Only the terminating claim remains -> zero live.
check(
    "zero live when last claim is terminating",
    claim.count_live_claims(
        [item("claim-a", "shared-ns", terminating=True)], FIELD, "shared-ns") == 0,
)

# Claims for a different namespace are not counted (so they don't keep ours alive).
check(
    "ignores claims targeting a different namespace",
    claim.count_live_claims(
        [item("claim-a", "shared-ns", terminating=True), item("claim-x", "other-ns")],
        FIELD, "shared-ns") == 0,
)

# Empty list -> zero.
check("empty list yields zero", claim.count_live_claims([], FIELD, "shared-ns") == 0)

# _get_path handles missing keys without raising.
check("_get_path returns None for missing path", claim._get_path({"spec": {}}, "spec.namespaceName") is None)

# target_field strips a leading dot (tolerates legacy ".spec.x" env values).
os.environ["TARGET_FIELD"] = ".spec.namespaceName"
check("target_field strips leading dot", claim.target_field() == "spec.namespaceName")
del os.environ["TARGET_FIELD"]
check("target_field default", claim.target_field() == "spec.namespaceName")

if failures:
    print(f"\n{len(failures)} FAILED")
    sys.exit(1)
print("\nALL TESTS PASSED")
