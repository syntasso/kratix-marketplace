# App Promise

This Promise is the result of following [Kratix
Workshop](https://docs.kratix.io/workshop/part-ii/intro) on Promise writing.

## Setup

* Run Minio in your Platform cluster, and make sure a secret with the MinIO
  credentials exists. See [./secret-example.yaml](./secret-example.yaml) for an
  example.

* Install the Postgresql Promise

  ```bash kubectl --context $PLATFORM apply
  --filename
  https://raw.githubusercontent.com/syntasso/promise-postgresql/main/promise-release.yaml
  ```

* Configure the Platform as a destination, and label the destinations
  accordingly:

  ```bash
  $> kubectl --context $PLATFORM get destinations --show-labels
  NAME               AGE     LABELS
  platform-cluster   8m13s   environment=platform
  worker-1           10m     environment=dev
  ```

* Build and Load the Aspects into your KinD cluster

  ```
  docker build --tag kratix-workshop/app-promise-pipeline:v0.1.0 workflows/promise/configure/dependencies/configure-deps
  docker build --tag kratix-workshop/app-pipeline-image:v1.0.0 workflows/resource/configure/mypipeline/kratix-workshop-app-pipeline-image

  kind load docker-image kratix-workshop/app-promise-pipeline:v0.1.0 --name platform
  kind load docker-image kratix-workshop/app-pipeline-image:v1.0.0 --name platform
  ```

For detailed instructions on environment setup, see the [Kratix
Workshop](https://docs.kratix.io/workshop/intro).

