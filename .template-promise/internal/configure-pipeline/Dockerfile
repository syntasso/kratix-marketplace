FROM ghcr.io/syntasso/kratix-pipeline-utility:v0.0.1

LABEL org.opencontainers.image.authors="kratix@syntasso.io"
LABEL org.opencontainers.image.source=https://github.com/syntasso/kratix-marketplace

RUN apk update && apk add --no-cache yq

ADD resources/* /tmp/transfer/
ADD execute-pipeline execute-pipeline

CMD [ "sh", "-c", "./execute-pipeline" ]
ENTRYPOINT []
