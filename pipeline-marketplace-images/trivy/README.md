# Trivy

```yaml
xaasRequestPipeline:
- # images
- ghcr.io/syntasso/kratix-marketplace/pipeline-trivy-image:v0.1.0
```

This image finds all container images in the documents in `/input` and run a
scan with trivy. It fails if any HIGH or CRITICAL vulnerabilities are found.

## Usage in the Pipeline

Add the image to the `xaasRequestPipeline` definition in your Promise and make
sure the `/input` contains the document you want to scan prior to the execution
of this image.

## Limitations

* At this moment, there's no way to control `trivy` scanning options
* This image was built with Trivy offline db. To update the database, the image
  needs to be rebuilt.

