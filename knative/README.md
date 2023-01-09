# Knative

This Promise provides [Knative](https://knative.dev/docs/)-as-a-Service. The Promise has 1 field `.spec.env`
which can be `dev` or `prod`: 
  * `dev` [Knative Autoscaling-HPA](https://knative.dev/docs/serving/autoscaling/autoscaler-types/#horizontal-pod-autoscaler-hpa) is switched off.
  * `prod` [Knative Autoscaling-HPA](https://knative.dev/docs/serving/autoscaling/autoscaler-types/#horizontal-pod-autoscaler-hpa) is switched on.


To install, run the following command while targeting your Platform cluster:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/knative/promise.yaml
```

To verify Knatives prerequisite CRDs are installed, run the following command while targeting a worker cluster:
```
kubectl get crds | grep knative.dev
```

To make a resource request, run the following command while targeting your Platform cluster:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/knative/resource-request.yaml
```

Once a Resource Request is made, you can see the controller is running by running the
following command while targeting a worker cluster:
```
kubectl get --namespace knative-serving deployment/controller
```

You can now provision Knative servings on your worker cluster, for example apply
the following yaml while targeting a worker cluster:
```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: helloworld-go
  namespace: default
spec:
  template:
    spec:
      containers:
        - image: gcr.io/knative-samples/helloworld-go
          env:
            - name: TARGET
              value: "Go Sample v1"
```


## Development

For development see [README.md](./internal/README.md)

## Questions? Feedback?

We are always looking for ways to improve Kratix and the Marketplace. If you run into issues or have ideas for us, please let us know. Feel free to [open an issue](https://github.com/syntasso/kratix-marketplace/issues/new/choose) or [put time on our calendar](https://www.syntasso.io/contact-us). We'd love to hear from you.