FROM "ghcr.io/syntasso/kratix-pipeline-utility:v0.0.1"

LABEL org.opencontainers.image.authors "kratix@syntasso.io"
LABEL org.opencontainers.image.source https://github.com/syntasso/kratix-marketplace

ADD execute-pipeline execute-pipeline

RUN apk update && apk add --no-cache vault libcap
RUN setcap -r /usr/sbin/vault

CMD [ "sh", "-c", "./execute-pipeline" ]
ENTRYPOINT []
