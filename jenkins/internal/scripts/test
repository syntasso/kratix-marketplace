#!/usr/bin/env bash
set -e

test_promise() {
  kubectl get crds jenkins.jenkins.io
  kubectl wait --for=condition=Available --timeout=5s deployment/jenkins-operator
}

test_resource_request() {
  kubectl wait --for=condition=ready --timeout=5s pod/jenkins-dev-example
}

if [ "$1" = "promise" ]; then
  test_promise
else
  test_resource_request
fi
