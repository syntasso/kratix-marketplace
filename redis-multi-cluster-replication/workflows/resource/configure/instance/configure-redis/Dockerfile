FROM "alpine"

LABEL org.opencontainers.image.authors="kratix@syntasso.io"
LABEL org.opencontainers.image.source=https://github.com/syntasso/kratix-marketplace

RUN apk update && apk add --no-cache yq helm kubectl

ADD scripts/pipeline.sh /usr/bin/pipeline.sh
ADD resources resources

RUN chmod +x /usr/bin/pipeline.sh

CMD [ "sh", "-c", "pipeline.sh" ]
ENTRYPOINT []
