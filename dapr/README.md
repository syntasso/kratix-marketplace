# WARNING- dapr helm chart generates static pub/priv key pair that are published in the repository.
# This promise should only be used locally for demo purposes
# If you wish to use this promise for more than demo purposes you should manually
# update all the secrets with keys in the promise with your own credentials

# Dapr

This Promise provides [Dapr](https://docs.dapr.io/).
Installing the Promise will install Dapr on all matching clusters.
There's no need to request individual resources once the Promise is installed.

To install:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/dapr/promise.yaml
```

To verify Dapr dashboard, you can use the [Dapr CLI](https://docs.dapr.io/getting-started/install-dapr-cli/) while targeting the Worker Cluster:

```
dapr dashboard -k
```

Check [Dapr docs](https://docs.dapr.io/getting-started/quickstarts/) for more use-cases of Dapr.

Thanks to [@salaboy](https://github.com/salaboy/) for helping build the Promise!

## Development

For development see [README.md](./internal/README.md)

## Questions? Feedback?

We are always looking for ways to improve Kratix and the Marketplace. If you run into issues or have ideas for us, please let us know. Feel free to [open an issue](https://github.com/syntasso/kratix-marketplace/issues/new/choose) or [put time on our calendar](https://www.syntasso.io/contact-us). We'd love to hear from you.
