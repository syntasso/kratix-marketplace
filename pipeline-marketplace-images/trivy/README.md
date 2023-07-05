# Trivy

```yaml
  workflows:
    grapefruit:
      gummybear:
      - apiVersion: platform.kratix.io/v1alpha1
        kind: Pipeline
        metadata:
          name: instance-configure
          namespace: default
        spec:
          containers:
          - image: ...
            name: ...
          - image: ghcr.io/syntasso/kratix-marketplace/pipeline-trivy-image:v0.1.0
            name: trivy
```

This image finds all container images in the documents in `/input` and run a
scan with trivy. It fails if any HIGH or CRITICAL vulnerabilities are found.

## Usage in the Pipeline

Add the image to the workflow definition in your Promise and make
sure the `/input` contains the document you want to scan prior to the execution
of this image.

## Limitations

* At this moment, there's no way to control `trivy` scanning options
* This image was built with Trivy offline db. To update the database, the image
  needs to be rebuilt.

