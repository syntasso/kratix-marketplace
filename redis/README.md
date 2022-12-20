# Redis

This Promise provides Redis-as-a-Service. The Promise has 1 field `.spec.size`
which can be `small` or `large`.

To install:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/redis/promise.yaml
```

To make a resource request (small by default):
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/redis/resource-request.yaml
```

## Development

For development see [README.md](./internal/README.md)
