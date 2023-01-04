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

## Questions? Feedback?

We are always looking for ways to improve Kratix and the Marketplace. If you run into issues or have ideas for us, please let us know. Feel free to [open an issue](https://github.com/syntasso/kratix-marketplace/issues/new/choose) or [put time on our calendar](https://www.syntasso.io/contact-us). We'd love to hear from you.