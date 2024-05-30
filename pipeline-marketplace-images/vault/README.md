# Vault

```yaml
workflows:
  resource:
    configure:
      - apiVersion: platform.kratix.io/v1alpha1
        kind: Pipeline
        metadata:
          name: instance-configure
          namespace: default
        spec:
          volumes:
            - name: vault-config
              configMap:
                name: vault-pipeline-image
          containers:
            - image: ...
              name: ...
            - image: ghcr.io/syntasso/kratix-marketplace/pipeline-vault-image:v0.2.0
              name: vault
              volumeMounts:
                - name: vault-config
                  mountPath: /vault/config
```

This image finds all `kind: Secret` documents in `/kratix/output`, stores them
in Vault, and remove the secrets from the output directory. All other documents
are left untouched.

If the original resource request is available on `/kratix/input`, the secrets will be stored
under `/secret/NAMESPACE/RESOURCE_NAME`. Otherwise, it will be stored under
`/secret/default/request-RANDOM`.

The path to the secrets will be available in the resource request status at the end of the
pipeline.

## Pre-requisites

### Vault Kubernetes Auth enabled

In order for the pipeline image to store the secrets, you need to setup your platform
cluster to have access to Vault using the Kubernetes auth method. Follow the commands
below, replacing the values with your Vault instance and Platform cluster information:

```bash
vault auth enable kubernetes

vault write auth/kubernetes/config \
    token_reviewer_jwt="${TOKEN}" \
    kubernetes_host="${K8S_HOST}" \
    kubernetes_ca_cert=@ca.crt
```

For further
documentation, please follow the [Vault
docs](https://developer.hashicorp.com/vault/docs/auth/kubernetes).

<details>
<summary>Running Vault and Kubernetes with kind?</summary>
<br />

For the JWT Token Reviewer, you can:

- Create a ServiceAccount for this pipeline stage:
  ```
  kubectl create serviceaccount vault-auth-delegator
  ```
- Create a ClusterRoleBinding binding the `system:auth-deletagor` ClusterRole to the ServiceAccount
  ```
  kubectl create clusterrolebinding role-tokenreview-binding \
      --clusterrole=system:auth-delegator \
      --serviceaccount=default:vault-auth-delegator
  ```
- Create a Secret and attach it to the ServiceAccount:
  ```
  kubectl apply -f - <<EOF
  apiVersion: v1
  kind: Secret
  metadata:
    name: vault-auth-token
    annotations:
      kubernetes.io/service-account.name: vault-auth-delegator
  type: kubernetes.io/service-account-token
  EOF
  ```
- Extract the JWT token:
  ```
  kubectl describe secrets/vault-auth-token
  ```

For the Kubernetes Host, you can run:

```bash
kubectl cluster-info
```

For the Kubernetes CA Certificate, run:

```bash
kubectl config view --raw --minify --flatten -o jsonpath='{.clusters[].cluster}' | yq '."certificate-authority-data"' | base64 -d
```

</details>

### Permission to write to /secret

The pipeline image will write all secrets to `/secret` in Vault. That means the
ServiceAccount for the Promise must have `create` access to `/secret`. For example:

```bash
cat > policy.hcl <<EOF
path "secret/*" {
  capabilities = ["create", "read", "update", "patch", "delete", "list"]
}
EOF
vault policy write secret-writer policy.hcl
```

Next step is to create a role attaching the policy to the ServiceAccount of the Promise
you intend to include this pipeline image to:

```bash
PROMISE_SA=<ServiceAccount for your Promise>
vault write auth/kubernetes/role/vault-pipeline-image \
    bound_service_account_names=${PROMISE_SA} \
    bound_service_account_namespaces=default \
    policies=secret-writer \
    ttl=1h
```

### Create the ConfigMap

Next, create a ConfigMap to inform the container of where is your Vault instance running:

```bash
VAULT_ADDR=<Your Vault address>
kubectl create configmap vault-pipeline-image \
    --from-literal=url=$VAULT_ADDR \
    --from-literal=role=vault-pipeline-image \
    --from-literal=authpath=kubernetes
```

Setting

- `VAULT_ADDR` with your Vault instance URL. If you are running both Kratix on KinD and
  Vault locally, use `http://host.docker.internal:PORT`.

The ConfigMap must be mounted on the container at `/vault/config`. See the
example on the top of this README as a reference.

## Usage in the Pipeline

Add the image to the workflow definition in your Promise. The image will
fetch the Vault config the ConfigMap and store the Secrets in Vault.

## Limitations

- This image won't parse `kind: List`, even if the list items are of `kind: Secret`.
- If any other document in `/kratix/output` refer to the Secret (like via a `volumeMount` in a
  `Pod`), this image won't remove those references, nor will it add any Vault-agent
  annotations. Please add an extra job in the pipeline to do that.
- Only keys in the `data` and `stringData` part of the Secret will be parsed
  and stored in Vault.
