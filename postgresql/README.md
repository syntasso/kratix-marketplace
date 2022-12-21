# PostgreSQL

This Promise provides Postgresql-as-a-Service. The Promise has 3 fields:
* `.spec.env`
* `.spec.teamID`
* `.spec.dbName`

Check the CRD documentation for more information.


To install:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/postgresql/promise.yaml
```

To make a resource request (small by default):
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/postgresql/resource-request.yaml
```

## Development

For development see [README.md](./internal/README.md)
