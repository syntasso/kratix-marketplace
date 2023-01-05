# Knative

This Promise provides [Knative](https://knative.dev/docs/)-as-a-Service. The Promise has 1 field `.spec.env`
which can be `dev` or `prod`: 
  * `dev` [Knative Autoscaling-HPA](https://knative.dev/docs/serving/autoscaling/autoscaler-types/#horizontal-pod-autoscaler-hpa) is switched off.
  * `prod` [Knative Autoscaling-HPA](https://knative.dev/docs/serving/autoscaling/autoscaler-types/#horizontal-pod-autoscaler-hpa) is switched on.

Once the Promise is installed, you can see Knative prerequisite CRDs in a worker cluster with the following command:
```
kubectl get crds | grep knative.dev
```

Once a Resource Request is made, you can see the controller running in the worker cluster with the following command:
```
kubectl get --namespace knative-serving deployment/controller
```

And you can apply Knative resources to deploy application, for example:
```
kubectl apply --filename https://raw.githubusercontent.com/syntasso/sample-golang-app/main/k8s/serving.yaml
```

To install:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/knative/promise.yaml
```

To make a resource request (small by default):
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/knative/resource-request.yaml
```

## Development

For development see [README.md](./internal/README.md)
