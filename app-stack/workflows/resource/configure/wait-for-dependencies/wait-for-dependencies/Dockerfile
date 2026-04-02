FROM alpine:3.20

RUN apk update && apk add --no-cache kubectl yq

COPY scripts/* /usr/local/bin/

RUN chmod +x /usr/local/bin/*

CMD [ "wait-for-dependencies" ]

ENTRYPOINT []