# AWS DB Promise

This Promise provides RDS-as-a-Service. The following parameters are available:

* `spec.dbName`: The name of the database to create.
* `spec.engine`: The database engine to use. Supported values are `mysql`, `postgres`, and `mariadb`.
* `spec.size`: The size of this deployment. Supported values are `micro`, `small`, `medium`, and `large`.

## Pre-requisites

To use this promise, you must have:

* A Access Key and Secret Access Key for your AWS account with permission to manage RDS instances
    You can follow AWS documentation, or review the following commands via the AWS CLI:

    ```bash
    export IAM_USER="kratix-$(whoami)-${RANDOM}-rds"
    aws iam create-user --user-name ${IAM_USER} --no-cli-pager 
    aws iam attach-user-policy --user-name ${IAM_USER} --policy-arn arn:aws:iam::aws:policy/AmazonRDSFullAccess --no-cli-pager 
    eval $(aws iam create-access-key --user-name ${IAM_USER} --output text --no-cli-pager --query 'AccessKey.[AccessKeyId,SecretAccessKey]' | \
        awk '{print "export RDS_KEY_ID=\"" $1 "\"\nexport RDS_SECRET_ACCESS_KEY=\"" $2 "\""}')
    ```
* A secret with a key for that service account in your cluster
    ```bash
    kubectl create secret generic aws-rds \
        --namespace default \
        --from-literal=accessKeyID=${RDS_KEY_ID} \
        --from-literal=secretAccessKey=${RDS_SECRET_ACCESS_KEY}
    ```

## Install Promise

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
