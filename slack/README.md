# Slack

This Promise provides Slack-Messages-as-a-Service. It has a single field: `.spec.message`, which is the message to be sent.

To install, run the following commands while targeting your Platform cluster:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/slack/promise.yaml
```

To provide credentials to slack create a secret in `default` namespace called
`slack-channel-hook` with a `.data.url` field. You can create it using
the following command (ensure you have SLACK_HOOK_URL env var exported):
```
kubectl --namespace default create secret generic \
  slack-channel-hook --from-literal=url=${SLACK_HOOK_URL}
```

Then apply the required RBAC:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/slack/rbac.yaml
```

To verify that the Promise is installed, run the following on your Platform cluster:
```
$ kubectl get promises.platform.kratix.io
NAME        AGE
slack       1m
```

To create a slack message, run the following command while targeting the Platform cluster:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/slack/resource-request.yaml
```

You should see a slack message appear shortly afterwards.

## Development

For development see [README.md](./internal/README.md)
