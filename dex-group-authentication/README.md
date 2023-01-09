# Dex

> **Warning**
> 
> **This repository requires a more advanced setup and the promise cannot be used directly on a cluster without prior work**
> 
> **This should NOT be used in production as is and contains local example credentials that are not to be used in real environments**

This demo Promise that installs [Dex](https://dexidp.io/) with a basic configuration for authorizing through the [GitHub connector](https://dexidp.io/docs/connectors/github/).

Authorization is a complex process, and requires a lot of domain understanding for each team.
This solution bypasses a lot of those challenges by using locally generated certificates in this directory and broad RBAC.

This Promise can be used to authenticate users to Kubernetes clusters. To do this, you will need:

1. A cluster started with OIDC configurations set including a valid Certificate Authority using. Use the script `./internal/scripts/setup` to get setup locally
1. A valid [GitHub OAuth application](https://github.com/settings/applications/new)
(for testing locally you can use `http://127.0.0.1:5555` as Homepage URL and `https://localhost:32000/callback` as the callback)
1. A secret which references the created GitHub OAuth application, run the following while targeting the worker cluster:
    ```
    kubectl -n dex create secret \
        generic github-client \
        --from-literal=client-id=<valid_github_oauth_client_id> \
        --from-literal=client-secret=<valid_github_oauth_client_secret>
    ```

Once this is in place, you can use resource requests for this Promise to allocate permissions to the required groups.
Users within these groups can then use [KubeLogin plugin](https://github.com/int128/kubelogin) to authenticate and use the `kubectl` commandline tool.

Once you've setup the prerequisites above you can install the promise by applying the following while targeting
the Platform cluster:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/dex-group-authentication/promise.yaml
```

To verify its correctly installed, run the following while targeting the worker cluster:
```
kubectl -n dex get deployments
NAME   READY   UP-TO-DATE   AVAILABLE   AGE
dex    3/3     3            3           17m
```

The kind cluster created in the earlier steps is now setup with Dex installed onto it. The final step is to create a resource
request stating what github `userGroups` should be allowed to access the Kubernetes environment. For example `userGroups: syntasso` would
allow all users in the `syntasso` org to have access to the cluster. `userGroups: syntasso:my-team` would limit it to only members of the `my-team` team
in the `syntasso` org.

To make a resource request modify the local `resource-request.yaml` with your desired group and run the following while targeting the platform cluster:
```
kubectl apply -f resource-request.yaml
```

You've now setup permissions for a userGroup specified in the `resource-request.yaml` to have access to login to the worker
cluster via OIDC. Use [KubeLogin](https://github.com/int128/kubelogin#setup) to verify this works (ensure `DEX_PATH` env var
is exported, for example `export DEX_PATH=~/workspace/kratix-marketplace/dex-group-authentication`:

1. Set up KubeLogin
    ```
    kubectl oidc-login setup \
    --oidc-issuer-url=https://localhost:32000 \
    --oidc-extra-scope=email \
    --oidc-extra-scope=groups \
    --oidc-client-id=kube \
    --oidc-client-secret=ZXhhbXBsZS1hcHAtc2VjcmV0 \
    --certificate-authority=$DEX_PATH/internal/scripts/config/ssl/ca.pem
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
	  --exec-arg=--certificate-authority=$DEX_PATH/internal/scripts/config/ssl/ca.pem
    ```

You should now be able to issue kubectl commands via the OIDC session. When the users attempts
to run a kubectl command they will be prompted to login via the GitHub auth flow, and will need to ensure they
grant access to the related org when prompted.


## Development

For development see [README.md](./internal/README.md)

## Questions? Feedback?

We are always looking for ways to improve Kratix and the Marketplace. If you run into issues or have ideas for us, please let us know. Feel free to [open an issue](https://github.com/syntasso/kratix-marketplace/issues/new/choose) or [put time on our calendar](https://www.syntasso.io/contact-us). We'd love to hear from you.
