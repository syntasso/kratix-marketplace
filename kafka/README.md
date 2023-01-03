# Kafka

This Promise provides [Kafka](https://kafka.apache.org/)-as-a-Service. The Promise has 1 field `.spec.size`
which can be `small` or `large`.

To install, run the following command while targeting your Platform cluster:
```
kubectl create -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/kafka/promise.yaml
```

This will install the Kafka Operator into your clusters. To verify the Kafka Operator is
installed, run the following command while targeting a worker cluster:
```
kubectl --namespace kafka get deployments
NAME                       READY   UP-TO-DATE   AVAILABLE   AGE
strimzi-cluster-operator   1/1     1            1           89s
```

To get a Kafka instance make a resource request (small by default), run the
following command while targeting your Platform cluster:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/kafka/resource-request.yaml
```

This will create an instance of Kafka on the targeted worker cluster.

To test your Kafka instance is working, you can run the following while targeting
your worker cluster publish some messages (replace `example` with the name of your resource request):
```
kubectl -n kafka run kafka-producer -ti --image=quay.io/strimzi/kafka:0.32.0-kafka-3.3.1 --rm=true --restart=Never -- bin/kafka-console-producer.sh --bootstrap-server example-kafka-bootstrap:9092 --topic my-topic
```

and the following to consume them:
```
kubectl -n kafka run kafka-consumer -ti --image=quay.io/strimzi/kafka:0.32.0-kafka-3.3.1 --rm=true --restart=Never -- bin/kafka-console-consumer.sh --bootstrap-server example-kafka-bootstrap:9092 --topic my-topic --from-beginning
```

## Development

For development see [README.md](./internal/README.md)
