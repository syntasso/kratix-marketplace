# App as a Service

> **Warning**
>
> **To use this Promise, the Kubernetes cluster running Kratix must be registered
as a Destination with the label `environment=platform`**
>
> Check out the [Compound Promises
guide](https://kratix.io/docs/main/guides/compound-promises) on the Kratix
documentation for details on how to setup your platform.

This Promise provides Compound App-as-a-Service promise with two `requiredPromises`. This is a compound Promise that installs the following `Deployment` Promise.

This Promise has two `requiredPromise` that must be installed to fulfil App-As-A-Service resource requests:

- [Nginx-ingress](https://github.com/syntasso/kratix-marketplace/tree/main/nginx-ingress)
- [Postgres](https://github.com/syntasso/kratix-marketplace/tree/main/postgresql)

These prerequisite promises can be installed via the `PromiseRelease`s in this repo:

```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/app-as-a-service/promises/nginx-promise-release.yaml
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/app-as-a-service/promises/postgresql-promise-release.yaml
```

The following fields are configurable:

- name: application name
- image: application image
- dbDriver: db type, only postgresql or none are valid options currently

To install:
```
kubectl create -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/app-as-a-service/promise.yaml
```

To make a resource request:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/app-as-a-service/resource-request.yaml
```

This resource request deploys the Kratix [sample Golang app](https://github.com/syntasso/sample-todo-list-app).

To test the sample app once it is successfully deployed, access it at `http://todo.local.gd:31338`.
```

## Development

For development see [README.md](./internal/README.md)

## Questions? Feedback?

We are always looking for ways to improve Kratix and the [Marketplace](kratix.io/marketplace). If you run into issues or have ideas for us, please let us know. Feel free to [open an issue](https://github.com/syntasso/kratix-marketplace/issues/new/choose) or [put time on our calendar](https://www.syntasso.io/contact-us). We'd love to hear from you.
