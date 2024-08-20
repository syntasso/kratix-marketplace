FROM gcr.io/google.com/cloudsdktool/google-cloud-cli:alpine

LABEL org.opencontainers.image.authors="kratix@syntasso.io"
LABEL org.opencontainers.image.source=https://github.com/syntasso/kratix-marketplace

RUN apk update && apk add --no-cache yq jq python3 curl bash kubectl

RUN curl -sSL https://sdk.cloud.google.com | bash

ADD execute-pipeline execute-pipeline

CMD [ "sh", "-c", "./execute-pipeline" ]
ENTRYPOINT []
