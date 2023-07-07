# ArgoCD Application

This Promise provides [ArgoCD](https://argo-cd.readthedocs.io/en/stable/) [Applications](https://argo-cd.readthedocs.io/en/stable/operator-manual/declarative-setup/#applications)-as-a-Service. The Promise will install an ArgoCD Server and then on each request for a Resource create a new ArgoCD Application within that server.

The Resource definition can configure the following fields:

- `source.repoURL` [required]: Must be a valid URL for a public git repository that contains Kubernetes resources.
- `source.path` [optional]: Is the directory within the git repository for ArgoCD to sync. Defaults to root of repository.
- `source.targetRevision` [optional]: Is the git revision for ArgoCD to sync. Defaults to `HEAD`.

To install, run the following command while targeting your Platform cluster:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/argocd-application/promise.yaml
```

To verify the ArgoCD Application controller is installed, run the following command
while targeting a Worker cluster:
```
kubectl get --namespace argocd deployment/argocd-server
```

To make a resource request, run the following command while targeting your Platform cluster:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/argocd-application/resource-request.yaml
```

This example deploys the contents of the `guestbook` directory in the
[argocd-example-app repository](https://github.com/argoproj/argocd-example-apps.git)

To verify the Resource is applied, you can see the corresponding Application in the
ArgoCD UI or resources on the cluster by run the following command while targeting the worker cluster:
```
kubectl get --namespace default deployment/guestbook-ui
kubectl get --namespace argocd applications.argoproj.io
```

## Development

For development see [README.md](./internal/README.md)

## Questions? Feedback?

We are always looking for ways to improve Kratix and the Marketplace. If you run into issues or have ideas for us, please let us know. Feel free to [open an issue](https://github.com/syntasso/kratix-marketplace/issues/new/choose) or [put time on our calendar](https://www.syntasso.io/contact-us). We'd love to hear from you.
