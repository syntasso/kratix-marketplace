FROM "mikefarah/yq:4"

MAINTAINER "kratix@syntasso.io"
LABEL org.opencontainers.image.authors "kratix@syntasso.io"
LABEL org.opencontainers.image.source https://github.com/syntasso/kratix-marketplace
LABEL org.opencontainers.image.description="Pipeline image for the Jenkins Promise"

RUN [ "mkdir", "/tmp/transfer" ]

ADD resources/* /tmp/transfer/
ADD execute-pipeline.sh execute-pipeline.sh

CMD [ "sh", "-c", "./execute-pipeline.sh" ]
ENTRYPOINT []
