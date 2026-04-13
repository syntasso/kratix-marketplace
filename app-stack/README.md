# Compound Promise with Waits and Retries

This Promise provides an App-as-a-Service, with waits and retries. The Promise has 2 fields `spec.database.driver` which can be `postgresql`
and `spec.image` which can be any application image.

## Promise Resource Workflows

The promise uses three `resource.configure` Pipelines:

- `create-dependencies`: Creates the dependencies
- `wait-for-dependencies`: Waits for dependencies
- `create-runtime`: Creates the application once the dependencies are met

Together these Pipelines can output up to two sub-requests:
- a `postgresql` request for the PostgreSQL Promise when `spec.database.driver` is `postgresql`
- a `Runtime` request for an application image after requested dependencies are ready

To install, run the following command while targeting your Platform cluster:

```sh
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/app-stack/promise.yaml
```

These prerequisite promises can be installed via the `PromiseRelease`s in this repo:

```bash
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/app-stack/promises/runtime-release.yaml
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/app-stack/promises/postgres-release.yaml
```

To make a resource request:

```sh
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/app-stack/resource-request.yaml
```