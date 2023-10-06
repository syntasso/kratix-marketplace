# MongoDB

This Promise provides MongoDB-as-a-Service. The Promise has 1 field
`.spec.majorVersion` which can be 4, 5 or 6.

`prod` MongoDB comes with Backups pre-configured.

To install:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/mongodb/promise.yaml
```

To make a resource request (small by default):
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/mongodb/resource-request.yaml
```

> **Warning**
>
> **This Promise uses a hard-coded password for demo purposes**
>

## Development

For development see [README.md](./internal/README.md)

## Questions? Feedback?

We are always looking for ways to improve Kratix and the Marketplace. If you run into issues or have ideas for us, please let us know. Feel free to [open an issue](https://github.com/syntasso/kratix-marketplace/issues/new/choose) or [put time on our calendar](https://www.syntasso.io/contact-us). We'd love to hear from you.
