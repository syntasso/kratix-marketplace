# NGINX Ingress

This Promise installs the NGINX Ingress controller onto a requested cluster. Given the controller is available cluster wide, there is no additional requirement to make a resource request with this promise.

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
