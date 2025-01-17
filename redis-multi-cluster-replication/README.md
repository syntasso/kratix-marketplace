# Redis Multi-Cluster Replication Promise

This Promise demonstrates deploying Redis across three Kubernetes clusters in a **primary/replica configuration**:
- The **primary cluster** is write-enabled.
- The **replica clusters** are read-only and sync data from the primary.

The primary cluster is responsible for replicating data to the replicas, ensuring consistency across the clusters.

## Prerequisites

This Promise assumes the existence of three Kubernetes clusters:
- **europe** (primary)
- **america** (replica)
- **asia** (replica)

The provided `./setup.bash` script will:
1. Create these clusters locally using `kind` and register them with Kratix.
2. Configure them to sync using ArgoCD, using the in-cluster Gitea on the platform cluster.

### Required Tools

Ensure the following tools are installed on your system:
- [`kubectl`](https://kubernetes.io/docs/tasks/tools/)
- [`kind`](https://kind.sigs.k8s.io/)
- [`git`](https://git-scm.com/)
- [`yq`](https://mikefarah.gitbook.io/yq/)
- [`kratix`](https://github.com/syntasso/kratix-cli)

## Setup Instructions

1. Run the `./setup.bash` script to:
   - Create the necessary clusters.
   - Deploy the Promise to the platform cluster.

2. Once deployed, create a Redis cluster by applying the `RedisMultiClusterReplication` object.
  ```bash
  kubectl --context kind-platform apply -f example-request.yaml
  ```

### Example Deployment
You can watch the rollout of the Redis cluster across the three clusters by running the following command:
```bash
kubectl --context kind-platform get pods  # Pipeline pod orchestrating the deployment
kubectl --context kind-worker-1 get pods  # Pods in the primary cluster
kubectl --context kind-worker-2 get pods  # Pods in the first replica cluster
kubectl --context kind-worker-3 get pods  # Pods in the second replica cluster
```

Eventually you will see something like:
```
=================================
           PLATFORM
=================================
NAMESPACE                NAME                                                      READY   STATUS      RESTARTS   AGE
default                  kratix-redis-multi-cluster-example-instance-c7733-x8qpn   0/1     Completed   0          38m
kratix-platform-system   kratix-platform-controller-manager-7df98c89c9-cxffg       1/1     Running     0          66m

=================================
           WORKER-1
=================================
NAMESPACE            NAME                                                READY   STATUS    RESTARTS   AGE
argocd               argocd-application-controller-0                     1/1     Running   0          65m
argocd               argocd-applicationset-controller-697b7748b9-92lpn   1/1     Running   0          65m
argocd               argocd-dex-server-5dfd67bb67-rc5wb                  1/1     Running   0          65m
argocd               argocd-notifications-controller-79c7b96c58-f4nnw    1/1     Running   0          65m
argocd               argocd-redis-648d946dd-mjhlq                        1/1     Running   0          65m
argocd               argocd-repo-server-846b58b8d7-cqgzp                 1/1     Running   0          65m
argocd               argocd-server-6767446cb9-lkblf                      1/1     Running   0          65m
default              redis-primary-master-0                              1/1     Running   0          38m

=================================
           WORKER-2
=================================
NAMESPACE            NAME                                                READY   STATUS    RESTARTS   AGE
argocd               argocd-application-controller-0                     1/1     Running   0          64m
argocd               argocd-applicationset-controller-6bc5667469-n9vf4   1/1     Running   0          64m
argocd               argocd-dex-server-5dfd67bb67-w92gp                  1/1     Running   0          64m
argocd               argocd-notifications-controller-79c7b96c58-qjj9v    1/1     Running   0          64m
argocd               argocd-redis-648d946dd-st462                        1/1     Running   0          64m
argocd               argocd-repo-server-79f85747fd-l2vwt                 1/1     Running   0          64m
argocd               argocd-server-6767446cb9-q7zk5                      1/1     Running   0          64m
default              redis-replica-replicas-0                            1/1     Running   0          38m

=================================
           WORKER-3
=================================
NAMESPACE            NAME                                                READY   STATUS    RESTARTS   AGE
argocd               argocd-application-controller-0                     1/1     Running   0          63m
argocd               argocd-applicationset-controller-6d6c9f59df-t8ngf   1/1     Running   0          63m
argocd               argocd-dex-server-5dfd67bb67-lwl9c                  1/1     Running   0          63m
argocd               argocd-notifications-controller-79c7b96c58-7t879    1/1     Running   0          63m
argocd               argocd-redis-648d946dd-wlpdv                        1/1     Running   0          63m
argocd               argocd-repo-server-554876df8b-lxt9n                 1/1     Running   0          63m
argocd               argocd-server-6767446cb9-2t84b                      1/1     Running   0          63m
default              redis-replica-replicas-0                            1/1     Running   0          38m
```

Inspecting the logs of the `redis-primary-master-0` pod will show the Redis cluster
replicating data to the replicas.
```bash
kubectl --context kind-worker-1 logs redis-primary-master-0
```

# Development

## Promise Template
This Promise was generated with:

```
kratix init promise redis-multi-cluster-replication --group marketplace.kratix.io --kind RedisMultiClusterReplication
```

## Updating API properties

To update the Promise API, you can use the `kratix update api` command:

```
kratix update api --property name:string --property region- --kind RedisMultiClusterReplication
```

## Updating Workflows

To add workflow containers, you can use the `kratix add container` command:

```
kratix add container resource/configure/pipeline0 --image syntasso/postgres-resource:v1.0.0
```

## Updating Dependencies

TBD
