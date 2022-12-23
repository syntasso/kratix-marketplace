# Kafka

This Promise provides Kafka-as-a-Service. The Promise has 1 field `.spec.size`
which can be `small` or `large`.

To install:
```
kubectl create -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/kafka/promise.yaml
```

To make a resource request (small by default):
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/kafka/resource-request.yaml
```

To test your kafka instance is working, you can run the following to publish some messages (replace `example` with the name of your resource request):
```
kubectl -n kafka run kafka-producer -ti --image=quay.io/strimzi/kafka:0.32.0-kafka-3.3.1 --rm=true --restart=Never -- bin/kafka-console-producer.sh --bootstrap-server example-kafka-bootstrap:9092 --topic my-topic
```

and the following to consume them:
```
kubectl -n kafka run kafka-consumer -ti --image=quay.io/strimzi/kafka:0.32.0-kafka-3.3.1 --rm=true --restart=Never -- bin/kafka-console-consumer.sh --bootstrap-server example-kafka-bootstrap:9092 --topic my-topic --from-beginning
```

## Development

For development see [README.md](./internal/README.md)
