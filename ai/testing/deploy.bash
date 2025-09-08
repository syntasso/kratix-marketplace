#!/usr/bin/env bash

set -euo pipefail

ROOT=$(dirname "$(realpath "$0")")
echo $ROOT
echo "This script is intended to be used to test the promise against a Kratix
instance created using the \"./scripts/quick-start.sh --single-cluster
--recreate\" script in the Kratix repo. 

Its also goig to produce an ollama-tiny-dolphin image, to provide a simple LLM API for testing"

pushd $ROOT/ollama/
  docker build -t ollama-tiny-dolphin .
  kind load docker-image ollama-tiny-dolphin --name platform
  kubectl apply -f ollama.yaml
popd
  
kubectl apply -f https://raw.githubusercontent.com/syntasso/promise-postgresql/refs/heads/main/promise.yaml
echo "Marking platform cluster as dev environment, to ensure everything is deployed to it"
kubectl label destinations.platform.kratix.io platform-cluster environment=dev --overwrite || true
kubectl label destinations.platform.kratix.io platform environment=dev --overwrite || true
docker pull ghcr.io/open-webui/open-webui:0.6.25
docker pull ghcr.io/berriai/litellm:v1.75.8-stable
kind load docker-image ghcr.io/open-webui/open-webui:0.6.25 --name platform
kind load docker-image ghcr.io/berriai/litellm:v1.75.8-stable --name platform

kubectl apply -f $ROOT/secrets.yaml
