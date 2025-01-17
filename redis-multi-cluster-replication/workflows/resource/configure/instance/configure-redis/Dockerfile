FROM "alpine"

RUN apk update && apk add --no-cache yq helm

ADD scripts/pipeline.sh /usr/bin/pipeline.sh
ADD resources resources

RUN chmod +x /usr/bin/pipeline.sh

CMD [ "sh", "-c", "pipeline.sh" ]
ENTRYPOINT []
