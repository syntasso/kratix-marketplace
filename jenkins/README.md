# Jenkins

This Promise provides Jenkins-as-a-Service. The Promise has 1 field `.spec.env`
which can be `dev` or `prod`.

To install:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/jenkins/promise.yaml
```

To make a resource request (small by default):
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/jenkins/resource-request.yaml
```

## Development

For development see [README.md](./internal/README.md)
