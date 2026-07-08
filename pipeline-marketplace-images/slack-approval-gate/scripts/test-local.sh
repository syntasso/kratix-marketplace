#!/bin/bash
# Simulates what Kratix does: runs the approval-gate stage in a loop,
# re-reading state from the backing service until approved (exit) or denied (exit).
#
# Requires the stage's Python deps locally:
#   pip install -r approval-gate/requirements.txt
#
# Usage:
#   source .envrc
#   ./scripts/test-local.sh
#
# Reset first with: ./scripts/reset-approval.sh

set -euo pipefail

RETRY_INTERVAL=${RETRY_INTERVAL:-10} # seconds between retries
MAX_RETRIES=${MAX_RETRIES:-60}       # ~10 minutes total

required_vars=(APPROVAL_SERVER_URL APPROVAL_SERVER_TOKEN KRATIX_PROMISE_NAME KRATIX_OBJECT_NAME REQUESTER_EMAIL)
for var in "${required_vars[@]}"; do
    if [[ -z "${!var:-}" ]]; then
        echo "ERROR: $var is not set. Source your env file first:"
        echo "  source .envrc"
        exit 1
    fi
done

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GATE="$SCRIPT_DIR/../approval-gate/approval-gate.py"
REQUEST_ID="approval:${KRATIX_PROMISE_NAME}:${KRATIX_OBJECT_NAME}"

echo "==> Starting approval gate loop"
echo "    Request:  $REQUEST_ID"
echo "    Worker:   $APPROVAL_SERVER_URL"
echo "    Retrying every ${RETRY_INTERVAL}s, max ${MAX_RETRIES} attempts"
echo ""

attempt=0
while [[ $attempt -lt $MAX_RETRIES ]]; do
    attempt=$((attempt + 1))
    echo "--- Attempt $attempt / $MAX_RETRIES ---"

    python3 "$GATE"

    STATE=$(curl -s -H "Authorization: Bearer ${APPROVAL_SERVER_TOKEN}" \
        "${APPROVAL_SERVER_URL}/approvals/${REQUEST_ID}" | jq -r '.state // "null"')

    if [[ "$STATE" == "approved" ]]; then
        echo ""
        echo "==> Pipeline approved. Continuing."
        exit 0
    fi
    if [[ "$STATE" == "denied" ]]; then
        echo ""
        echo "==> Pipeline denied. Stopping."
        exit 1
    fi

    echo "    State is '${STATE}'. Waiting ${RETRY_INTERVAL}s before next attempt..."
    sleep "$RETRY_INTERVAL"
done

echo ""
echo "==> Max retries reached without a decision. Exiting."
exit 1
