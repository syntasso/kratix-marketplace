# sql

This Promise provides sql-as-a-Service. The Promise has 1 field `.spec.size`
which can be `small` or `large`.

## Prerequisites

Requires a secret for connecting to azure:

```
kubectl create secret generic azure-credentials --from-file=key=key.pem \
    --from-literal=servicePrincipalID=<servicePrincipalID> \
    --from-literal=tenantID=<tenantID>
```

## Install Promise

To install:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/sql/azure/promise.yaml
```

To make a resource request (small by default):
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/sql/azure/resource-request.yaml
```

## Development

For development see [README.md](./internal/README.md)

## Questions? Feedback?

We are always looking for ways to improve Kratix and the Marketplace. If you run into issues or have ideas for us, please let us know. Feel free to [open an issue](https://github.com/syntasso/kratix-marketplace/issues/new/choose) or [put time on our calendar](https://www.syntasso.io/contact-us). We'd love to hear from you.
