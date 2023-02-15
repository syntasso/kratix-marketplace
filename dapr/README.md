# Dapr


## Generating worker cluster resources
helm template dapr dapr/dapr -n dapr-system --create-namespace --version 1.9.6 > helm.yaml

helm show crds dapr/dapr --version 1.9.6

This Promise provides install Dapr on all clusters

To install:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/dapr/dapr/promise.yaml
```
## Development

For development see [README.md](./internal/README.md)

## Questions? Feedback?

We are always looking for ways to improve Kratix and the Marketplace. If you run into issues or have ideas for us, please let us know. Feel free to [open an issue](https://github.com/syntasso/kratix-marketplace/issues/new/choose) or [put time on our calendar](https://www.syntasso.io/contact-us). We'd love to hear from you.
