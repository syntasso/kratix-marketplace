# istio

This Promise install Istio on all clusters.

To install:
```
kubectl create -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/istio/promise.yaml
```

To enable Istio sidecar inject in pods in your namespace:
```
kubectl label namespace <namespace> istio-injection=enabled
```

To access the Kiali dashboard open a port-forward
```
kubectl port-forward svc/kiali -n istio-system 20001:20001
```

access in your browser at localhost:20001

## Development

For development see [README.md](./internal/README.md)
