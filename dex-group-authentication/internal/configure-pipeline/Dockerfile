FROM "alpine"

LABEL org.opencontainers.image.authors="kratix@syntasso.io"
LABEL org.opencontainers.image.source=https://github.com/syntasso/kratix-marketplace

RUN [ "mkdir", "/tmp/transfer" ]
RUN apk update && apk add --no-cache yq

ADD resources/* /tmp/transfer/
ADD resource-configure /usr/local/bin/resource-configure
ADD promise-configure /usr/local/bin/promise-configure

RUN chmod +x /usr/local/bin/*
