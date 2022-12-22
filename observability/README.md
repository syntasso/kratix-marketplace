# Observability

This Promise provides Observability-as-a-Service by deploying [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus). The Promise has 2 field `.spec.namespace`,
which is the namespace the Grafana and Prometheus installation and `.spec.env` which can be `dev` or `prod`.

To install:
```
kubectl create -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/observability/promise.yaml
```

To make a resource request (dev by default):
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/observability/resource-request.yaml
```

## Development

For development see [README.md](./internal/README.md)
