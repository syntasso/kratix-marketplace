# Developer

TODO

## Fetch external resources

The script and path to the dir should be reviewed, depending on if we will add/include the generated resources part of the workflow `promise` or `resource`
and corresponding oci image !

To fetch the Tekton & Dashboard manifests, execute the following command
```shell
./scripts/fetch-deps <TEKTON_VERSION> <DASHBOARD_VERSION>
```

## Pipeline image

To build the image used by the workflows/resource/configure:
```
cd workflows/resource-configure
podman build -t tekton-configure:0.1.0 . && kind load docker-image tekton-configure:0.1.0 -n kratix
```

TODO: to be reviewed

To load the image to the local kind platform cluster:
```
./scripts/pipeline-image load
```

To push the image to ghcr.io:
```
./scripts/pipeline-image push
```


## Testing
To test the promise install kratix, and then:
```
kubectl apply -f promise.yaml
./scripts/test promise
```

This asserts the Promise is installed correctly.

To test the resource request:
```
kubectl apply -f resource-request.yaml
./scripts/test resource-request
```
