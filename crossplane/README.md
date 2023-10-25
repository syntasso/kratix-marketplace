# Crossplane

This Promise installs [Crossplane](https://www.crossplane.io/) on your clusters.

To install, run the following command while targeting your Platform cluster:

```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/crossplane/promise.yaml
```

This will install Crossplane into your worker clusters. To verify Crossplane is installed,
run the following command while targeting a worker cluster:

```
kubectl -n crossplane-system get deployments
NAME                      READY   UP-TO-DATE   AVAILABLE   AGE
crossplane                1/1     1            1           4m36s
crossplane-rbac-manager   1/1     1            1           4m36s
```

Crossplane is not being provided as-a-Service with this Promise. Therefore, there's
no Resource definition: installing the Promise suffice to get Crossplane installed.

## Development

For development see [README.md](./internal/README.md)

## Questions? Feedback?

We are always looking for ways to improve Kratix and the Marketplace. If you run into issues or have ideas for us, please let us know. Feel free to [open an issue](https://github.com/syntasso/kratix-marketplace/issues/new/choose) or [put time on our calendar](https://www.syntasso.io/contact-us). We'd love to hear from you.
