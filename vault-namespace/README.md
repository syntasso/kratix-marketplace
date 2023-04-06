# BETA- DO NOT USE IN PRODUCTION
# Vault Namespaces

This Promise provides Vault-Namespaces-as-a-Service for Vault Enterprise. The
Promise has 3 fields:
* `.spec.team` the name of the team (used to create the namespace)
* `.spec.kubernetes.host` and `.spec.kubernetes.caCert`. The cluster to grant
  access to vault. Only service account in the team namespace within this
  cluster will have access.

Each resource request will provision a Vault namespace (under the parent namespace
configured below), create a KV-2 Engine `kv` inside that namespace, and enable
Kubernetes auth to the vault cluster from the Kubernetes namespace with the same name.
All service account in the namespace will have read-only access to the `kv` engine.


To install:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/vault-namespace/promise.yaml
```

In order for the Vault namepaces to be provisioned, you need to ensure you have setup your
platform cluster to have access to Vault using the [Kubernetes auth](https://developer.hashicorp.com/vault/docs/auth/kubernetes) method.
Once auth is setup, create a Role within Vault for Kratix to use to provision the namespaces:
```
vault write auth/kubernetes/role/kratix \
    bound_service_account_names=vaultnamespace-default-promise-pipeline \
    bound_service_account_namespaces=default \
    policies=NAMESPACE_POLICY \
    ttl=1h
```

The pipeline will use the `vaultnamespace-default-promise-pipeline` Service
Account to authenticate with vault. Replace `NAMESPACE_POLICY` with a policy that
can create Vault Namespaces, KV and Auth engines.

Note: the default endpoint is `auth/kubernetes/login`. If this auth method was enabled
at a different path, use that value instead of kubernetes.

Create a config map `vault` with the address of your vault instance:

```
kubectl create configmap vault \ 
    --from-literal=url=$VAULT_ADDR \
    --from-literal=role=kratix \
    --from-literal=parentNamespace=admin \
    --from-literal=authpath=kubernetes
```

Replacing
* `VAULT_ADDR` with your Vault instance URL. If you are running both Kratix on KinD and Vault locally, use `http://host.docker.internal:PORT'.

To make a resource request download the template:
```
wget https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/vault-namespace/resource-request.yaml
```

update the resource request with your teams details and make the request to the platform clusteR:
```
kubectl apply -f resource-request.yaml
```

Once the pipeline is executed, the namespace is ready to be used and the Kubernetes cluster
is configured to authenticate with the Vault namespace.

## Development

For development see [README.md](./internal/README.md)

## Questions? Feedback?

We are always looking for ways to improve Kratix and the Marketplace. If you
run into issues or have ideas for us, please let us know. Feel free to [open an
issue](https://github.com/syntasso/kratix-marketplace/issues/new/choose) or
[put time on our calendar](https://www.syntasso.io/contact-us). We'd love to
hear from you.
