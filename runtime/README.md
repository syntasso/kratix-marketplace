# Runtime Promise

Runs an Application on top of Kubernetes resources. It will deploy:

- A Deployment
- A Service
- An ingress rule

The domain used in the ingress rule can be controlled by a ConfigMap called `runtime-domain` with the following structure:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: runtime-domain
data:
  domain: example.com
  port: 80 # optional
```

To install, run the following command while targeting your Platform cluster:

```bash
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/runtime/promise.yaml
```

To make a resource request:

```bash
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/runtime/resource-request.yaml
```