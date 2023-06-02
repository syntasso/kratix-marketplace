# vcluster promise

> **Warning**
>
> **To use this Promise, the Kubernetes cluster running Kratix must be registered
as a Worker Cluster and labelled with `environment: platform`**
>
> Check out the [Compound Promises
guide](https://kratix.io/docs/main/guides/compound-promises) on the Kratix
documentation for details on how to setup your platform.

This Promise provides vcluster-as-a-Service. The Promise exposes no optional
fields. The name of any resource requests become the name of the vcluster namespace.

All vclusters will be created on the same cluster that Kratix is running.

To install:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/vcluster/promise.yaml
```

To make a resource request:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/vcluster/resource-request.yaml
```

## Accessing vcluster

You can use the [vcluster CLI to connect](https://www.vcluster.com/docs/getting-started/connect)
to the cluster. In addition, a kubeconfig file will be created and stored as a
secret in the newly created vcluster namespace.

## Adding more on demand software to your vcluster

This promise creates an empty vcluster. and registers it to Kratix allowing you
to request any other platform offerings to your cluster.

Kratix scheduling is based on [labels](https://kratix.io/docs/main/reference/multicluster-management#promises).
This vcluster is created with two labels, one static (`type: vcluster`) and one dynamic
(`clusterName` set to your vcluster name). You can use these labels to schedule
additional resource requests to this cluster or ALL vclusters on your platform.

## Development

For development see [README.md](./internal/README.md)

## Questions? Feedback?

We are always looking for ways to improve Kratix and the Marketplace. If you run into issues or have ideas for us, please let us know. Feel free to [open an issue](https://github.com/syntasso/kratix-marketplace/issues/new/choose) or [put time on our calendar](https://www.syntasso.io/contact-us). We'd love to hear from you.
