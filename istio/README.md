# Istio

This Promise installs [Istio](https://istio.io/) on your clusters.

To install, run the following command while targeting your Platform cluster:
```
kubectl create -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/istio/promise.yaml
```

This will install Istio into your worker clusters. To verify Istio is installed,
run the following command while targeting a worker cluster:
```
kubectl -n istio-system get deployments
NAME         READY   UP-TO-DATE   AVAILABLE   AGE
istiod       1/1     1            1           3m18s
jaeger       1/1     1            1           3m18s
kiali        1/1     1            1           3m18s
prometheus   1/1     1            1           3m18s

```

Istio is not being provided as-a-Service with this Promise. Therefore, there's no Resource Request: installing the Promise suffice to get Istio installed.

To enable Istio sidecar inject in pods in your namespace, run the following
command while targeting the worker cluster:
```
kubectl label namespace <namespace> istio-injection=enabled
```

To access the Kiali dashboard open a port-forward
```
kubectl port-forward svc/kiali -n istio-system 20001:20001
```

Access in your browser at localhost:20001

## Development

For development see [README.md](./internal/README.md)

## Questions? Feedback?

We are always looking for ways to improve Kratix and the Marketplace. If you run into issues or have ideas for us, please let us know. Feel free to [open an issue](https://github.com/syntasso/kratix-marketplace/issues/new/choose) or [put time on our calendar](https://www.syntasso.io/contact-us). We'd love to hear from you.