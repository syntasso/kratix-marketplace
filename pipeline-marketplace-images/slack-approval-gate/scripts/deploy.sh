#!/usr/bin/env bash

set -euo pipefail

rootdir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "==> Building approval gate stage"
docker build -t ghcr.io/syntasso/kratix-marketplace/slack-approval-gate:v0.1.0 "$rootdir/approval-gate"
kind load docker-image ghcr.io/syntasso/kratix-marketplace/slack-approval-gate:v0.1.0 --name platform

echo "==> Creating approval gate secrets"
kubectl delete secret approval-gate-secrets --ignore-not-found
kubectl create secret generic approval-gate-secrets \
  --from-literal=server.url=${APPROVAL_SERVER_URL} \
  --from-literal=server.token=${APPROVAL_SERVER_TOKEN}

echo "==> Applying promise"
kubectl apply -f "$rootdir/promise/promise.yaml"

echo "==> Done!"