# Kubeflow Pipelines

This Promise provides Kubeflow Pipelines.

To install:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/kubeflow-pipelines/promise.yaml
```

To verify the installation:

```shell-session
kubectl wait crd/applications.app.k8s.io --for condition=established --timeout=60s
```

# Cert-manager

This Promise provides [cert-manager](https://cert-manager.io/docs/) to a
Cluster. Installing the Promise will install cert-manager under the
`cert-manager` namespace. There's no need to request individual resources once
the Promise is installed.

To install, run the following command while targeting your Platform cluster:
```
kubectl create -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/cert-manager/promise.yaml
```

To verify the Promise is installed, you can run the following command while
targeting a worker cluster:

```
kubectl wait crd/certificates.cert-manager.io --for condition=established --timeout=60s
kubectl wait pods -l app.kubernetes.io/instance=cert-manager -n cert-manager --for condition=Ready --timeout=60s
```

## Development

For development see [README.md](./internal/README.md)

## Questions? Feedback?

We are always looking for ways to improve Kratix and the Marketplace. If you run into issues or have ideas for us, please let us know. Feel free to [open an issue](https://github.com/syntasso/kratix-marketplace/issues/new/choose) or [put time on our calendar](https://www.syntasso.io/contact-us). We'd love to hear from you.
