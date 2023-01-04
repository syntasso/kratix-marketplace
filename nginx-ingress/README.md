# NGINX Ingress

This Promise provides access to the [NGINX Ingress controller](https://docs.nginx.com/nginx-ingress-controller/) for a given cluster by deploying the [NGINX Ingress Operator](https://github.com/nginxinc/nginx-ingress-helm-operator) and configuring a Cluster-wide Ingress Controller.

Given the controller is available cluster wide, there is no additional requirement to make a resource request with this promise.

To install:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/nginx-ingress/promise.yaml
```

To make a resource request (small by default):
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/nginx-ingress/resource-request.yaml
```

## Development

For development see [README.md](./internal/README.md)
