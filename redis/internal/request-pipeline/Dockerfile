FROM "mikefarah/yq:4"
RUN [ "mkdir", "/tmp/transfer" ]

ADD resources/redis-instance.yaml /tmp/transfer/redis-instance.yaml
ADD execute-pipeline.sh execute-pipeline.sh

CMD [ "sh", "-c", "./execute-pipeline.sh" ]
ENTRYPOINT []
