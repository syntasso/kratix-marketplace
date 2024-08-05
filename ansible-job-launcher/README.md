# Ansible Job Launcher

This promise allows users to launch job templates on an existing AWX Tower via requests to Kratix.

## Setup

It is common for users to have a Tower already running and this Promise is easily adaptable for use with any existing Tower. For testing, you can begin by creating a local cluster with both Kratix and an AWX instance using Helm:
1. Start a cluster from the root of the [kratix repo](https://github.com/syntasso/kratix):
    ```
    ./scripts/quick-start.sh --git --single-cluster --recreate
    ```
1. Install the AWX operator:
    ```
    helm repo add awx-operator https://ansible.github.io/awx-operator/
    helm repo update
    helm install --version 2.19.1 -n awx --create-namespace my-awx-operator awx-operator/awx-operator
    ```
1. Create an instance of AWX using the operator:
    ```
    kubectl apply -f - <<EOF
    apiVersion: awx.ansible.com/v1beta1
    kind: AWX
    metadata:
        name: awx-demo
        namespace: awx
    spec:
        service_type: nodeport
        nodeport_port: 31340
    EOF
    ```
1. Login to verify AWX instance is healthy:
    The instance may take 5-10 minutes to become ready for use while showing an "upgrading" or "migration failed" page. This is normal startup process for AWX.
    
    If you used the quick start script to start your cluster the url localhost:31340 should be directly accessible without port-forwarding.
    * URL: http://localhost:31340/
    * Username: admin
Password:
    * Password:
        ```
        kubectl get secrets -n awx awx-demo-admin-password -ogo-template='{{.data.password|base64decode}}'
        ```

## Using the Promise

> Warning
> The Promise is set to work with the AWX instance installed in the setup commands by default. Changing this to another AWX instance is as easy as editing the environment variables set in the `./internal/configure-pipeline/execute-pipeline` script and rebuilding the pipeline images.

From this current directory, to install:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/ansible-job-launcher/promise.yaml
```

From this current directory, to make a resource request:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/ansible-job-launcher/resource-request.yaml
```

## Development

For development see [README.md](./internal/README.md)

## Questions? Feedback?

Further extensions are available to generate new playbooks, job templates and other admin tasks via Promises. Please reach out to learn more.

We are always looking for ways to improve Kratix and the Marketplace. If you run into issues or have ideas for us, please let us know. Feel free to [open an issue](https://github.com/syntasso/kratix-marketplace/issues/new/choose) or [put time on our calendar](https://www.syntasso.io/contact-us). We'd love to hear from you.
