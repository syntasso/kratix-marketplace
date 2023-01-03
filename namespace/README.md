# Namespace

This Promise provides Namespaces-as-a-Service. It has a single field: `namespaceName`, which is the name of the namespace to be created.

To install, run the following command while targeting your Platform cluster:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/namespace/promise.yaml
```

To verify that the Promise is installed, run the following on your Platform cluster:
```
$ kubectl get promises.platform.kratix.io
NAME        AGE
namespace   1m
```

To create a namespace in the worker cluster, run the following command while targeting the Platform cluster:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/namespace/resource-request.yaml
```

To verify that the namespace is created, run the following command while targeting the Worker cluster:
```shell-session
$ kubectl get namespaces promised-namespace
NAME                STATUS   AGE
promised-namespace  Active   1m
```

## Development

For development see [README.md](./internal/README.md)
