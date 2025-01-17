# Redis Multi Cluster Replication Promise
This promise provides an example of deploying Redis across three clusters, in a
primary/replica configuration. The primary cluster is the only one that can be
written to, and the replicas are read-only. The primary cluster is also
responsible for replicating data to the replicas.

This Promise relies on the existence of three Kubernetes Destination clusters:
- europe (primary)
- america (replica)
- asia (replica)

The `./setup.bash` script will create these clusters for you, and set up the
clusters to sync using ArgoCD from an in-cluster Gitea instead on the platform
cluster.

## Usage
To use this Promise, you can run the `./setup.bash` script. This script will
create the necessary clusters and deploy the Promise to the platform cluster.
The script requires you have the following tools installed:
- `kubectl`
- `kind`
- `git`
- `yq`
- `kratix` https://github.com/syntasso/kratix-cli

Once the Promise is deployed, you can create a Redis cluster by creating a
`RedisMultiClusterReplication` object in the platform cluster. An example object
is provided in `example-request.yaml`. Install it onto the platform cluster by
running `kubectl --context kind-platform apply -f example-request.yaml`.

You can then observe the deployment of the Redis cluster by running:
```
kubectl --context kind-platform get RedisMultiClusterReplication
kubectl --context kind-worker-1 get pods
kubectl --context kind-worker-2 get pods
kubectl --context kind-worker-3 get pods
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
