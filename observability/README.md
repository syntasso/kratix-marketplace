# Observability

This Promise provides Observability-as-a-Service by deploying [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus). The Promise has 2 field `.spec.namespace`,
which is the namespace the Grafana and Prometheus installation and `.spec.env` which can be `dev` or `prod`.

To install:
```
kubectl create -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/observability/promise.yaml
```

This will install the Prometheus Operator and related components into the clusters. To get an
instance of Promethues and Grafana you need to make a resource request.

To make a resource request (dev by default):
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/observability/resource-request.yaml
```

This will create an instance of Promethues and Grafana on the targeted cluster.

## Accessing UI

Prometheus and Grafana dashboards can be accessed quickly using `kubectl port-forward`

### Prometheus

```shell
$ kubectl --namespace monitoring port-forward svc/prometheus-k8s 9090
```

Then access via [http://localhost:9090](http://localhost:9090)

### Grafana

```shell
$ kubectl --namespace monitoring port-forward svc/grafana 3000
```

Then access via [http://localhost:3000](http://localhost:3000) and use the default grafana user:password of `admin:admin`.

## Development

For development see [README.md](./internal/README.md)
