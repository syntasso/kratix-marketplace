#!/usr/bin/env bash
set -euo pipefail


setup_platform_cluster() {
	rm -rf kratix/ || true
	git clone git@github.com:syntasso/kratix.git
	pushd kratix
	  ./scripts/quick-start.sh --single-cluster --recreate --git
	popd
}

create_kind_clusters() {
	echo "Creating worker clusters"
	kind create cluster --name worker-1 --config=assets/worker-1.yaml
	kind create cluster --name worker-2 --config=assets/worker-2.yaml
	kind create cluster --name worker-3 --config=assets/worker-3.yaml
	kubectl config use-context kind-platform
}

register_destinations() {
	echo "Registering destinations"
	cat <<EOF | kubectl apply --context kind-platform -f -
---
apiVersion: platform.kratix.io/v1alpha1
kind: Destination
metadata:
  name: europe
  labels:
    clusterName: kind-worker-1
    region: europe
spec:
  stateStoreRef:
    kind: GitStateStore
    name: default
---
apiVersion: platform.kratix.io/v1alpha1
kind: Destination
metadata:
  name: asia
  labels:
    clusterName: kind-worker-2
    region: asia
spec:
  stateStoreRef:
    kind: GitStateStore
    name: default
---
apiVersion: platform.kratix.io/v1alpha1
kind: Destination
metadata:
  name: america
  labels:
    clusterName: kind-worker-3
    region: america
spec:
  stateStoreRef:
    kind: GitStateStore
    name: default
EOF

	pushd kratix
		./scripts/install-gitops --context kind-worker-1 --path europe --git --kustomization-name europe --git --gitops-provider argo
		./scripts/install-gitops --context kind-worker-2 --path asia --git --kustomization-name asia --git --gitops-provider argo
		./scripts/install-gitops --context kind-worker-3 --path america --git --kustomization-name america --git --gitops-provider argo
	popd
}

deploy_redis() {
	echo "Installing Promise"
	# if arg is --build then build image
	if [ "${1:-}" == "--build-and-push" ]; then
		docker build -t ghcr.io/syntasso/kratix-marketplace/redis-multi-cluster-replication-configure-pipeline:v0.1.0 ./workflows/resource/configure/instance/configure-redis/
		kind load docker-image --name platform ghcr.io/syntasso/kratix-marketplace/redis-multi-cluster-replication-configure-pipeline:v0.1.0
	fi
	export WORKER_1_IP=$(docker inspect worker-1-control-plane | yq ".[0].NetworkSettings.Networks.kind.IPAddress")
	kubectl --context kind-platform create configmap redis-multi-cluster-replication-promise-data --from-literal=host=$WORKER_1_IP --from-literal=port="31341"
	kubectl --context kind-platform apply -f redis-multi-cluster-replication-promise.yaml
}

setup_platform_cluster
create_kind_clusters
register_destinations
deploy_redis $@
echo ""
echo "Environment setup complete, to make a request run \`kubectl apply -f example-resource.yaml\`"
