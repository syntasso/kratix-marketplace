# Sealed Secrets

```yaml
workflows:
  resource:
    configure:
      - apiVersion: platform.kratix.io/v1alpha1
        kind: Pipeline
        metadata:
          name: instance-configure
        spec:
          rbac:
            permissions:
            - apiGroups:
              - ""
              resources:
              - configmaps
              verbs:
              - get
              resourceNames:
              - sealed-secrets
              resourceNamespace: default
          containers:
            - image: ...
              name: ...
            - image: ghcr.io/syntasso/kratix-marketplace/pipeline-sealedsecrets-image:v0.2.0
              name: sealed-secrets
```

This image finds all `kind: Secret` documents in `/kratix/output`, encrypts them with
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
- On the Platform Cluster, there's a ConfigMap (named `sealed-secrets` on the
  `default` namespace by default) with a `certificate` key containing the
  public part of the certificate.

The commands below are similar to what you'll need to run:

```bash
# Assuming a mytls.crt file exists on the local directory

# On the platform cluster
kubectl --namespace default create configmap sealed-secrets \
    --from-literal=certificate="$(cat mytls.crt)"
```

## Usage in the Pipeline

Add the image to the workflow definition in your Promise. The image will
fetch the certificate from the ConfigMap and replace any Secret with the SealedSecret
equivalent.

### Override the kubeseal configmap name

To use a different configmap name, amend the container as follow:
```
- image: ghcr.io/syntasso/kratix-marketplace/pipeline-sealedsecrets-image:v0.2.0
  name: sealed-secrets
  env:
  - name: KUBESEAL_CONFIGMAP
    value: sealed-secrets-dev
  - name: KUBESEAL_NAMESPACE
    value: kratix
```

## Limitations

- This image won't parse `kind: List`, even if the list items are of `kind: Secret`.
