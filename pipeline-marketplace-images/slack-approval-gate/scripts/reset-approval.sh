#!/bin/bash
# Deletes the approval key from the Worker so you can re-run the test fresh.
#
# Usage:
#   source .envrc
#   ./scripts/reset-approval.sh

set -euo pipefail

required_vars=(APPROVAL_SERVER_URL APPROVAL_SERVER_TOKEN KRATIX_PROMISE_NAME KRATIX_OBJECT_NAME)
for var in "${required_vars[@]}"; do
    if [[ -z "${!var:-}" ]]; then
        echo "ERROR: $var is not set. Source your env file first."
        exit 1
    fi
done

REQUEST_ID="approval:${KRATIX_PROMISE_NAME}:${KRATIX_OBJECT_NAME}"

echo "==> Deleting ${REQUEST_ID} on the Worker..."
curl -s -X DELETE -H "Authorization: Bearer ${APPROVAL_SERVER_TOKEN}" \
    "${APPROVAL_SERVER_URL}/approvals/${REQUEST_ID}" >/dev/null

echo "Done. State reset — next test-local.sh run will send a fresh Slack message."
