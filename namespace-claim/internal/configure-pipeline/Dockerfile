FROM alpine:3.20

LABEL org.opencontainers.image.authors="kratix@syntasso.io"
LABEL org.opencontainers.image.source=https://github.com/syntasso/kratix-marketplace

RUN apk update && apk add --no-cache kubectl yq

COPY claim /usr/local/bin/
RUN chmod +x /usr/local/bin/claim

CMD [ "claim", "--ensure" ]
ENTRYPOINT []
