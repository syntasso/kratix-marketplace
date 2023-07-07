# Slack Pipeline Image

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
          - image: ghcr.io/syntasso/kratix-marketplace/pipeline-slack-image:v0.1.0
            name: slack
```

This image uses [Slack Incoming Webhooks](https://api.slack.com/messaging/webhooks) to
send notifications to Slack.

## Pre-requisites

For this image to work, you will first need to [setup the
webhook](https://api.slack.com/messaging/webhooks) and store the URL in a
Kubernetes Secret in your platform cluster:

```shell
kubectl --namespace <NAMESPACE> create secret generic \
  slack-channel-hook --from-literal=url=<SLACK HOOK URL>
```

## Usage in the Pipeline

Add the image to the workflow definition in your Promise.

This image is intented to be used alongide with other container images in a
Promise Pipeline. It relies on the existence of one or more YAML files in a
`/metadata/notifications` with the following format:

```yaml
message: The message to be sent
slackHook:
    secretName: slack-channel-hook
    namespace: default
    # OR
    url: https://slack.hook.example.url
```

If `slackHook.url` is provided, the pipeline image will send the `message` to that URL.

If `slackHook.secretName` is provided, the pipeline image will first fetch the URL from
the Kubernetes Secret in the specified namespace. You must ensure the Service
Account associated with the Promise that includes this image has _read_ access
to the Secret. Check [Passing secrets to the
Pipeline](https://kratix.io/docs/main/reference/resource-requests/pipelines#passing-secrets-to-the-pipeline)
for further details.

### Avoiding multiple notifications

This image makes use of `/metadata/status.yaml` to determine if the
notifications were successfully sent in a previous pipeline run. To ensure the
image does not send a new notification with every run, ensure to store the
initial `object.yaml` status in `/metadata/status.yaml` in a previous pipeline
stage.
