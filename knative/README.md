# Knative

This Promise provides Knative-as-a-Service. The Promise has 1 field `.spec.env`
which can be `dev` or `prod`: 
  * `dev` [Knative Autoscaling-HPA](https://knative.dev/docs/serving/autoscaling/autoscaler-types/#horizontal-pod-autoscaler-hpa) is switched off.
  * `prod` [Knative Autoscaling-HPA](https://knative.dev/docs/serving/autoscaling/autoscaler-types/#horizontal-pod-autoscaler-hpa) is switched on.


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
