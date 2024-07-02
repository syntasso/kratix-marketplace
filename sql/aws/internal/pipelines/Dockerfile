FROM alpine:3.20.1

LABEL org.opencontainers.image.authors "kratix@syntasso.io"
LABEL org.opencontainers.image.source https://github.com/syntasso/kratix-marketplace

RUN apk update && apk add --no-cache jq yq bash kubectl aws-cli

ADD resource-lifecycle resource-lifecycle
ADD wait-instance wait-instance

ENV PATH="/:${PATH}"

CMD [ "sh", "-c", "./resource-lifecycle" ]

ENTRYPOINT []
