# AWS DB Promise

## Pre-requisites

### Credentials

To use this promise, you must have a Access Key and Secret Access Key for your
AWS account. The IAM user associated with these keys must have the permission to
create and delete RDS instances.

Create a Kubernetes Secret with the credentials:

```bash
kubectl create secret generic aws-rds \
    --namespace default \
    --from-literal=accessKeyID=<your access key> \
    --from-literal=secretAccessKey=<your secret access key>
```

## Usage

This Promise provides RDS-as-a-Service. The following parameters are available:

* `spec.dbName`: The name of the database to create.
* `spec.engine`: The database engine to use. Supported values are `mysql`, `postgres`, and `mariadb`.
* `spec.size`: The size of this deployment. Supported values are `micro`, `small`, `medium`, and `large`.

To install:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/sql/aws/promise.yaml
```

To make a resource request :
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/sql/aws/resource-request.yaml
```

## Development

For development see [README.md](./internal/README.md)

## Questions? Feedback?

We are always looking for ways to improve Kratix and the Marketplace. If you run into issues or have ideas for us, please let us know. Feel free to [open an issue](https://github.com/syntasso/kratix-marketplace/issues/new/choose) or [put time on our calendar](https://www.syntasso.io/contact-us). We'd love to hear from you.
