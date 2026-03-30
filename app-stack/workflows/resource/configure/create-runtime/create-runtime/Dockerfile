FROM alpine:3.20

RUN apk update && apk add --no-cache yq

COPY scripts/* /usr/local/bin/

RUN chmod +x /usr/local/bin/*

CMD [ "create-runtime" ]
ENTRYPOINT []