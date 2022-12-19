#!/usr/bin/env bash

set -x

PWD="$( cd $(dirname $0)/.. && pwd)"

promise_name="$(basename "$(dirname "${PWD}")")"
pipeline_image="ghcr.io/syntasso/marketplace/${promise_name}-request-pipeline:v0.1.0"

case "$1" in
  build)
    docker build \
      --tag "${pipeline_image}" \
      --platform linux/amd64 \
      "${PWD}/request-pipeline" ;;

  load)
    kind load docker-image "${pipeline_image}" --name platform ;;

  push)
    docker push "${pipeline_image}" ;;

  *)
    echo "unknown command $1"
    exit 1
    ;;
esac
