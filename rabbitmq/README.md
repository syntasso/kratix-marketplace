# RabbitMQ

This Promise provides RabbitMQ-as-a-Service. The Promise has 3 fields:
* `.spec.env`
* `.spec.plugins`

Check the CRD documentation for more information.


To install:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/rabbitmq/promise.yaml
```

To make a resource request:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/rabbitmq/resource-request.yaml
```

## Development

For development see [README.md](./internal/README.md)
