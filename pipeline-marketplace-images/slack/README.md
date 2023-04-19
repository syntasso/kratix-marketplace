# Slack Pipeline Image

```yaml
xaasRequestPipeline:
- # images
- ghcr.io/syntasso/kratix-marketplace/pipeline-slack-image:v0.1.0
```

## Pre-requisites

This image uses [Slack Incoming Webhooks](https://api.slack.com/messaging/webhooks) to
send notifications to Slack. To enable this image to work properly, you will first need
to [setup the webhook](https://api.slack.com/messaging/webhooks) and store the URL in a
Kubernetes Secret in your platform cluster:

```shell
kubectl --namespace <NAMESPACE> create secret generic \
  slack-channel-hook --from-literal=url=<SLACK HOOK URL>
```

Furthermore, this image is intented to be used alongide with other container images in a
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

## Usage

* Configure your platform cluster as the steps above
* Add a step in your pipeline that generates the document specified above
* Somewhere after that step, includes the Slack image (see above for image name)
* Watch the Slack channel for the notification on resource requests
