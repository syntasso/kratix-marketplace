# Kubevela global instance

This Promise will install global instance of Kubevela onto a Worker cluster.

To install:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/kubevela/promise.yaml
```

You can deploy Kubevela applications directly to the Worker cluster. There is a sample app `application.yaml`:

```bash
# targeting the Worker Cluster
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/kubevela/application.yaml
```

To validate the application, you can use the [`vela` CLI](https://kubevela.io/docs/cli/vela):

```bash
vela port-forward helm-test-vela-app 8000:8000 -n vela-system
```

The app will be available on [http://localhost:8000](http://localhost:8000).

## Development

For development see [README.md](./internal/README.md)

## Questions? Feedback?

We are always looking for ways to improve Kratix and the Marketplace. If you run into issues or have ideas for us, please let us know. Feel free to [open an issue](https://github.com/syntasso/kratix-marketplace/issues/new/choose) or [put time on our calendar](https://www.syntasso.io/contact-us). We'd love to hear from you.
