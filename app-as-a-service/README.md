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
- [Postgres](https://github.com/syntasso/promise-postgresql)

These prerequisite promises can be installed via the `PromiseRelease`s in this repo:

```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/app-as-a-service/promises/nginx-promise-release.yaml
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/app-as-a-service/promises/postgresql-promise-release.yaml
```

When `dbDriver: postgresql` is set, the configure pipeline provisions a `postgresql` resource request.
Vault integration is optional and described in the [Vault Integration](#vault-integration) section.

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

To test the sample app once it is successfully deployed, access it at `http://todoer.local.gd:31338`.

## Vault Integration

Vault is opt-in for each app resource. Without the Vault label, the app still uses PostgreSQL and no Vault-specific wiring is applied.

To opt into Vault for a specific app, add this label to the resource metadata:
```
app-as-a-service.marketplace.kratix.io/vault: "true"
```

Tip: you can set this label directly in the resource YAML before applying it, or add it later:
```
kubectl label app <app-name> app-as-a-service.marketplace.kratix.io/vault=true --overwrite
```

If you want this to be automatic, use an admission webhook (for example, a mutating webhook) to inject
`app-as-a-service.marketplace.kratix.io/vault: "true"` into matching app resources.

When the Vault label is set to `true`, a dedicated `vault-configure` workflow step:

- creates a dedicated workload `ServiceAccount` for the app in the target namespace
- runs the app `Deployment` using that service account
- updates the `postgresql` resource request to enable Vault and bind it to that service account
- adds Vault Agent injector annotations so credentials are rendered to `/vault/secrets/pg-db.env`
- wraps application startup with a launcher that reads those credentials into env vars before starting the original image entrypoint/cmd
- runs a small reloader sidecar that restarts the app container when the credential file rotates

The configure pipeline resolves the image entrypoint/cmd from the registry so the wrapper can preserve normal startup behavior.

Or apply the example request with Vault enabled:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/app-as-a-service/resource-request-vault.yaml
```

## Development

For development see [README.md](./internal/README.md)

## Questions? Feedback?

We are always looking for ways to improve Kratix and the [Marketplace](kratix.io/marketplace). If you run into issues or have ideas for us, please let us know. Feel free to [open an issue](https://github.com/syntasso/kratix-marketplace/issues/new/choose) or [put time on our calendar](https://www.syntasso.io/contact-us). We'd love to hear from you.
