# NGINX Ingress

This Promise provides access to the [NGINX Ingress Controller (NIC)](https://docs.nginx.com/nginx-ingress-controller/) in the global deployment configuration. Since this configuration only deploys a single instance per cluster, there is no need to request individual resources once the Promise is installed. For more details about the difference, see the NGINX documents [here](https://docs.nginx.com/nginx-ingress-controller/installation/running-multiple-ingress-controllers/).

Once the Promise is installed, you can see the ingress controller running on any applicable worker clusters with the following command:
```
kubectl get --namespace default deployment/nginx-nginx-ingress
```

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
