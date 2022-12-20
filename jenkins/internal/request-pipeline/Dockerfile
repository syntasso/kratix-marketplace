FROM "alpine"

LABEL org.opencontainers.image.authors "kratix@syntasso.io"
LABEL org.opencontainers.image.source https://github.com/syntasso/kratix-marketplace

RUN [ "mkdir", "/tmp/transfer" ]
RUN apk update && apk add --no-cache yq

ADD resources/* /tmp/transfer/
ADD execute-pipeline.sh execute-pipeline.sh

CMD [ "sh", "-c", "./execute-pipeline.sh" ]
ENTRYPOINT []
