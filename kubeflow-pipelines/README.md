# Kubeflow Pipelines

This Promise provides [Kubeflow Pipelines](https://www.kubeflow.org/docs/components/pipelines/v1/introduction/).

To install:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/kubeflow-pipelines/promise.yaml
```

To verify the installation:

```shell-session
kubectl wait crd/applications.app.k8s.io --for condition=established --timeout=60s
```

To make a resource request:
```shell-session
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/kubeflow-pipelines/resource-request.yaml
```

This will create a namespace with the name `metadata.name` of the resource
request, prefixed with `kf`. The kubeflow pipelines will be deployed at that
namespace.

To verify the deployment (it can take a few minutes):

```
kubectl wait pods -l application-crd-id=kubeflow-pipelines -n kf-example --for condition=Ready --timeout=600s
```

To access the Kubeflow Pipelines UI on http://localhost:8080, run:

```
kubectl port-forward -n kf-example svc/ml-pipeline-ui 8080:80
```

## Development

For development see [README.md](./internal/README.md)

## Questions? Feedback?

We are always looking for ways to improve Kratix and the Marketplace. If you run into issues or have ideas for us, please let us know. Feel free to [open an issue](https://github.com/syntasso/kratix-marketplace/issues/new/choose) or [put time on our calendar](https://www.syntasso.io/contact-us). We'd love to hear from you.
