# Sealed Secrets

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
          containers:
            - image: ...
              name: ...
            - image: ghcr.io/syntasso/kratix-marketplace/pipeline-sealedsecrets-image:v0.1.0
              name: sealed-secrets
```

This image finds all `kind: Secret` documents in `/kratix/input`, encrypts them with
[kubeseal](https://github.com/bitnami-labs/sealed-secrets), and creates the sealed
documents in `/kratix/output`. All non-secret documents are copied over.

## Pre-requisites

`kubeseal` uses a key certificate (public key portion) for sealing secrets. To use this
image in your pipeline, you must ensure that:

- The Sealed Secrets controller is installed on the Worker clusters
- On all Worker Clusters, the Sealed Secrets controller is configured with the same
  certificate. See [Bring your own
  certificate](https://github.com/bitnami-labs/sealed-secrets/blob/main/docs/bring-your-own-certificates.md)
  for details.
- On the Platform Cluster, there's a `sealed-secrets` ConfigMap on the `default`
  namespace with a key `certificate` containing the public part of the certificate.
- The ServiceAccount for the Promise that will include this image has access to the
  `sealed-secrets` ConfigMap.

The commands below are similar to what you'll need to run:

```bash
# Assuming a mytls.crt file exists on the local directory

# On the platform cluster
kubectl --namespace default create configmap sealed-secrets \
    --from-literal=certificate="$(cat mytls.crt)"

kubectl create clusterrole sealed-secrets-reader \
    --verb=get \
    --resource=configmap \
    --resource-name=sealed-secrets

# Replace PROMISE with the name of your promise
kubectl create clusterrolebinding PROMISE-sealed-secret \
    --clusterrole=sealed-secrets-reader \
    --serviceaccount=default:PROMISE-default-promise-pipeline
```

## Usage in the Pipeline

Add the image to the workflow definition in your Promise. The image will
fetch the certificate from the ConfigMap and replace any Secret with the SealedSecret
equivalent.

## Limitations

- This image won't parse `kind: List`, even if the list items are of `kind: Secret`.
