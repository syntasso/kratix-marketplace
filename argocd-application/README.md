# ArgoCD Application

This Promise provides ArgoCD Applications-as-a-Service. The Promise will install an ArgoCD Server and then on each Resource Request create a new ArgoCD Application within that server.

The Resource Request can configure the following fields:
  * `source.repoURL` [required]: Must be a valid URL for a public git repository that contains Kubernetes resources.
  * `source.path` [optional]: Is the directory within the git repository for ArgoCD to sync. Defaults to root of repository.
  * `source.targetRevision` [optional]: Is the git revision for ArgoCD to sync. Defaults to `HEAD`.


To install:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/argocd-application/promise.yaml
```

To make a resource request (small by default):
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/argocd-application/resource-request.yaml
```

## Development

For development see [README.md](./internal/README.md)
