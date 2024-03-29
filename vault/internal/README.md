# Development

## Build promise.yaml
The `promise.yaml` is generated by interpolating the contents of `resources/` into
the `dependencies`. To build run:

```
./scripts/inject-deps
```

## Pipeline image
To build the image:
```
./scripts/pipeline-image build
```

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
## Deploying an Application in Dev mode

Once deployed we need to setup Vault: 

1. Get vault root token: `kubectl --context kind-worker get secret -n vault-autounseal vault-root-token \
    -o 'jsonpath={.data.root_token}' | base64 -d`
2. `kubectl exec -ti vault-example-0 -- /bin/sh`
3. `vault login` (Use token from step 1.)
4. ```
    cat <<EOF > /home/vault/app-policy.hcl
    path "secret*" {
      capabilities = ["read"]
    }
    EOF
    
    vault policy write app /home/vault/app-policy.hcl
    ```
5. `vault auth enable kubernetes`
6. ```
   vault write auth/kubernetes/config \
      token_reviewer_jwt="$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
      kubernetes_host=https://${KUBERNETES_PORT_443_TCP_ADDR}:443 \
      kubernetes_ca_cert=@/var/run/secrets/kubernetes.io/serviceaccount/ca.crt
   
   vault write auth/kubernetes/role/myapp \
      bound_service_account_names=app \
      bound_service_account_namespaces=default \
      policies=app \
      ttl=1h
    ```
7. `vault secrets enable -path=secret kv`
8. `vault kv put secret/helloworld username=foobaruser password=foobarbazpass`

Exit the Vault server pod. 

On host run:

9. `kubectl --context kind-worker apply -f example/app.yaml`
10. kubectl --context kind-worker exec -ti app-XXX -c app -- cat /vault/secrets/helloworld

Thanks to the vault team for this blog: https://www.hashicorp.com/blog/injecting-vault-secrets-into-kubernetes-pods-via-a-sidecar
