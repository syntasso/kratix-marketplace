# Destination

```yaml
metadata:
  annotations:
    destinationSelector: foo
workflows:
  resource:
    configure:
      - apiVersion: platform.kratix.io/v1alpha1
        kind: Pipeline
        metadata:
          name: instance-configure
        spec:
          containers:
            - image: ...
              name: ...
            - image: http://ghcr.io/syntasso/kratix-marketplace/pipeline-destination-image:v0.1.0
              name: destination
```

This image creates the destinations selectors file based on annotations under `.metadata.annotations.destinationSelector`

