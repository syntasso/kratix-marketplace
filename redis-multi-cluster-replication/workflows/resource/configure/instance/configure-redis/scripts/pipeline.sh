#!/usr/bin/env sh

set -xe

export name="$(yq eval '.metadata.name' /kratix/input/object.yaml)"

mkdir -p /kratix/output/primary/
mkdir -p /kratix/output/replica-1/
mkdir -p /kratix/output/replica-2/

helm repo add bitnami https://charts.bitnami.com/bitnami

cat << EOF > values.yaml
architecture: standalone
auth:
  enabled: false  # Disable authentication for simplicity (optional; enable if needed)
master:
  service:
    type: NodePort
    nodePorts:
      redis: "31341"
EOF
helm template redis-primary bitnami/redis --version 20.6.2 -f values.yaml > /kratix/output/primary/redis-primary.yaml

cat << EOF > values.yaml
architecture: replication
master:
  count: 0  # Ensure no master pods are deployed
replica:
  replicaCount: 1
  externalMaster:
    enabled: true
    # In this case we know the address of the primary, this could be fetched dynamically instead
    host: host.docker.internal
    port: 31341
  command:
    - redis-server
  args:
    - --replicaof
    - host.docker.internal
    - "31341"
    - --replica-announce-ip
    - ${name}-replica-1
  service:
    type: NodePort
    nodePorts:
      redis: "31342"

auth:
  enabled: false  # Disable authentication for simplicity

EOF
helm template redis-replica bitnami/redis --version 20.6.2 -f values.yaml > /kratix/output/replica-1/redis-replica.yaml

cat << EOF > values.yaml
architecture: replication
master:
  count: 0  # Ensure no master pods are deployed
replica:
  replicaCount: 1
  externalMaster:
    enabled: true
    # In this case we know the address of the primary, this could be fetched dynamically instead
    host: host.docker.internal
    port: 31341
  command:
    - redis-server
  args:
    - --replicaof
    - host.docker.internal
    - "31341"
    - --replica-announce-ip
    - ${name}-replica-2
  service:
    type: NodePort
    nodePorts:
      redis: "31343"

auth:
  enabled: false  # Disable authentication for simplicity
EOF
helm template redis-replica bitnami/redis --version 20.6.2 -f values.yaml > /kratix/output/replica-2/redis-replica.yaml


cat << EOF > /kratix/metadata/destination-selectors.yaml
- directory: primary
  matchLabels:
    region: europe
- directory: replica-1
  matchLabels:
    region: asia
- directory: replica-2
  matchLabels:
    region: america
EOF
