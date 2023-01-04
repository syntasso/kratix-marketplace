# Elasitc Cloud on Kubernetes

This Promise deploys an Elasticsearch and a Kibana instance, using [Elastic Cloud on
Kubernetes
(ECK)](https://www.elastic.co/guide/en/cloud-on-k8s/current/k8s-overview.html). The
promise has a single field:

- `spec.env`

Check the CRD documentation for more information.

To install, run the following command while targeting your Platform cluster:

```
kubectl create -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/elasticcloud/promise.yaml
```

This will start the Elastic Operator on Worker Cluster. To verify the installation, run the following
command while targeting the Worker cluster (it may take a couple of minutes):

```shell-session
$ kubectl get all --namespace elastic-system
NAME                     READY   STATUS    RESTARTS   AGE
pod/elastic-operator-0   1/1     Running   0          49s

NAME                             TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)   AGE
service/elastic-webhook-server   ClusterIP   10.96.99.51   <none>        443/TCP   49s

NAME                                READY   AGE
statefulset.apps/elastic-operator   1/1     49s
```

To create an instance of Elastic Search with Kibana, run the following command while targeting the Platform cluster:

```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/elasticcloud/resource-request.yaml
```

To verify that the Elastic Stack is created, run the following command while targeting the Worker cluster:

```shell-session
$ kubectl get elasticsearches.elasticsearch.k8s.elastic.co
NAME      HEALTH   NODES   VERSION   PHASE   AGE
example   yellow   1       8.5.3     Ready   4m15s
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
kubectl get secret example-es-elastic-user -o=jsonpath='{.data.elastic}' | base64 --decode
```

## Development

For development see [README.md](./internal/README.md)

## Questions? Feedback?

We are always looking for ways to improve Kratix and the Marketplace. If you run into issues or have ideas for us, please let us know. Feel free to [open an issue](https://github.com/syntasso/kratix-marketplace/issues/new/choose) or [put time on our calendar](https://www.syntasso.io/contact-us). We'd love to hear from you.
