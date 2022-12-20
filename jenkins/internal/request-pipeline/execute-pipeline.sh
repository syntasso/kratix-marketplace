#!/usr/bin/env sh

set -x

# Read current values from the provided resource request
export name="$(yq eval '.metadata.name' /input/object.yaml)"

env_type="$(yq eval '.spec.env // "dev"' /input/object.yaml)"
base_instance="/tmp/transfer/${env_type}-jenkins-instance.yaml"

# Replace defaults with user provided values
sed "s/NAME/${name}/g" "${base_instance}" > /output/jenkins-instance.yaml
