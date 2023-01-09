# Dex

This demo Promise that installs [Dex](https://dexidp.io/) with a basic configuration for authorizing through the [GitHub connector](https://dexidp.io/docs/connectors/github/).

Authorization is a complex process, and requires a lot of domain understanding for each team. This solution bypasses a lot of those challenges by using hardcoded certificates in this directory and broad RBAC.

This should NOT be used in production as is!

This Promise can be used to authenticate users to Kubernetes clusters. To do this, you will need:

1. A cluster started with OIDC configurations set including a valid Certificate Authority using `./scripts/setup`
1. A valid [GitHub OAuth application](https://github.com/settings/applications/new)
1. A secret which references the created GitHub OAuth application:
    ```
    kubectl -n dex create secret \
        generic github-client \
        --from-literal=client-id=<valid_github_oauth_client_id> \
        --from-literal=client-secret=<valid_github_oauth_client_secret>
    ```

Once this is in place, you can use resource requests for this Promise to allocate permissions to the required groups. Users within these groups can then use [KubeLogin plugin](https://github.com/int128/kubelogin) to authenticate and use the `kubectl` commandline tool.

Below is an example workflow for a user through KubeLogin:

1. Set up KubeLogin
    ```
    kubectl oidc-login setup \
    --oidc-issuer-url=https://localhost:32000 \
    --oidc-extra-scope=email \
    --oidc-extra-scope=groups \
    --oidc-client-id=kube \
    --oidc-client-secret=ZXhhbXBsZS1hcHAtc2VjcmV0 \
    --certificate-authority=./internal/scripts/config/ssl/ca.pem
    ```
2. Set up kube config:
    ```
    kubectl config set-credentials oidc \
	  --exec-api-version=client.authentication.k8s.io/v1beta1 \
	  --exec-command=kubectl \
	  --exec-arg=oidc-login \
	  --exec-arg=get-token \
	  --exec-arg=--oidc-issuer-url=https://localhost:32000 \
	  --exec-arg=--oidc-client-id=kube \
	  --exec-arg=--oidc-client-secret=ZXhhbXBsZS1hcHAtc2VjcmV0 \
	  --exec-arg=--oidc-extra-scope=email \
	  --exec-arg=--oidc-extra-scope=groups \
	  --exec-arg=--certificate-authority=./internal/scripts/config/ssl/ca.pem
    ```


To install:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/dex-group-authentication/promise.yaml
```

To make a resource request (small by default):
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/dex-group-authentication/resource-request.yaml
```

## Development

For development see [README.md](./internal/README.md)

## Questions? Feedback?

We are always looking for ways to improve Kratix and the Marketplace. If you run into issues or have ideas for us, please let us know. Feel free to [open an issue](https://github.com/syntasso/kratix-marketplace/issues/new/choose) or [put time on our calendar](https://www.syntasso.io/contact-us). We'd love to hear from you.
