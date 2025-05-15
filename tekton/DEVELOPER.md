# Developer

## Fetch external resources

To fetch the Tekton & Dashboard manifests, execute the following command
```shell
./scripts/fetch-deps <TEKTON_VERSION> <DASHBOARD_VERSION>
```
If no versions are passed as argument, then we download the latest release file for Tekton and the Dashboard. If you prefer to download a specific vesion of Tekton, then pass its version to the script
```shell
./scripts/fetch-deps v0.68.1
```
**NOTE**: The Tekton file downloaded will be stored under the folder `./dependencies` and its name will include the version: `tekton-v0.68.1.yaml`, `tekton-la  test.yaml`, etc. 

## Pipeline image

To build the image used by the `workflows/resource/configure`, execute the following command
```
cd workflows/resource-configure
podman build -t tekton-configure:0.1.0 .
```

To load the image to a local kind platform cluster:
```
kind load docker-image tekton-configure:0.1.0 -n <KIND_NODE_NAME>
```

To push the image to `ghcr.io`:
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
