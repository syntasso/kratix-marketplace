# Namespace

This Promise provides Namespaces-as-a-Service. It has a single field: `namespaceName`, which is the name of the namespace to be created.

To install:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/namespace/promise.yaml
```

To make a resource request (small by default):
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/namespace/resource-request.yaml
```

## Development

For development see [README.md](./internal/README.md)
