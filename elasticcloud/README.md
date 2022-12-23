# Elasitc Cloud on Kubernetes

This Promise deploys an Elasticsearch and a Kibana instance, using [Elastic Cloud on Kubernetes (ECK)](https://www.elastic.co/guide/en/cloud-on-k8s/current/k8s-overview.html).

To install:
```
kubectl create -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/elasticcloud/promise.yaml
```

To make a resource request (small by default):
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/elasticcloud/resource-request.yaml
```

## Accessing Kibana

A ClusterIP Service is automatically created for Kibana:

```
kubectl get service example-kb-http
```

Use kubectl port-forward to access Kibana from your local workstation:

```
kubectl port-forward service/example-kb-http 5601
```

Open https://localhost:5601 in your browser. Login as the `elastic` user. The password can be obtained with the following command:

```
kubectl get secret quickstart-es-elastic-user -o=jsonpath='{.data.elastic}' | base64 --decode
```

## Development

For development see [README.md](./internal/README.md)
