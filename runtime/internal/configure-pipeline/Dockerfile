FROM "alpine"

LABEL org.opencontainers.image.authors="kratix@syntasso.io"
LABEL org.opencontainers.image.source=https://github.com/syntasso/kratix-marketplace

RUN apk update && apk add --no-cache yq kubectl

ADD resources /resources
ADD resource-configure /usr/local/bin/resource-configure
ADD promise-configure /usr/local/bin/promise-configure

RUN chmod +x /usr/local/bin/*
