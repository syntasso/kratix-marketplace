# Crossplane

This Promise installs [Crossplane](https://www.crossplane.io/) on your clusters.

To install, run the following command while targeting your Platform cluster:
```
kubectl create -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/crossplane/promise.yaml
```

This will install Crossplane into your worker clusters. To verify Crossplane is installed,
run the following command while targeting a worker cluster:
```
kubectl -n crossplane-system get deployments
NAME                      READY   UP-TO-DATE   AVAILABLE   AGE
crossplane                1/1     1            1           4m36s
crossplane-rbac-manager   1/1     1            1           4m36s
```

## Development

For development see [README.md](./internal/README.md)
