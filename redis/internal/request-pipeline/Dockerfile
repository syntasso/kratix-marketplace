FROM "mikefarah/yq:4"
RUN [ "mkdir", "/tmp/transfer" ]

LABEL org.opencontainers.image.authors "Syntasso"
LABEL org.opencontainers.image.source https://github.com/syntasso/kratix-marketplace
LABEL org.opencontainers.image.description="Pipeline image for the Redis Promise"

ADD resources/redis-instance.yaml /tmp/transfer/redis-instance.yaml
ADD execute-pipeline.sh execute-pipeline.sh

CMD [ "sh", "-c", "./execute-pipeline.sh" ]
ENTRYPOINT []
