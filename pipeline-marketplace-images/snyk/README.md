# Snyk

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
          rbac:
            permissions:
            - apiGroups:
              - ""
              resources:
              - secrets
              verbs:
              - get
              resourceNames:
              - snyk-token # change this to the secret name if different
              resourceNamespace: default
          containers:
            - image: ...
              name: ...
            - image: ghcr.io/syntasso/kratix-marketplace/pipeline-snyk-image:v0.1.0
              name: snyk
```

This image finds all container images in the documents in `/kratix/output` and run a
scan with snyk.

## Pre-requisites

The pipeline requires a Snyk token token to authenticate with. Create a secret called
`snyk-token` in the `default` namespace with the field `token` set with your snyk token.

```
kubectl --namespace default create secret generic \
  snyk-token --from-literal=token=${SNYK_TOKEN}
```

## Usage in the Pipeline

Add the image to the workflow definition in your Promise and make
sure the `/kratix/output` contains the document you want to scan prior to the execution
of this image.

## Limitations

- At this moment, there's no way to control `snyk` scanning options
