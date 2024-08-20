FROM mcr.microsoft.com/azure-cli

LABEL org.opencontainers.image.authors="kratix@syntasso.io"
LABEL org.opencontainers.image.source=https://github.com/syntasso/kratix-marketplace

RUN apk update && apk add --no-cache yq jq python3 curl bash kubectl uuidgen

ADD execute-pipeline execute-pipeline

CMD [ "sh", "-c", "./execute-pipeline" ]
ENTRYPOINT []
