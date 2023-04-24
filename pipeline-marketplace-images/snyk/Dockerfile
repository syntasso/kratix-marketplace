FROM "ghcr.io/syntasso/kratix-pipeline-utility:v0.0.1"

LABEL org.opencontainers.image.authors "kratix@syntasso.io"
LABEL org.opencontainers.image.source https://github.com/syntasso/kratix-marketplace

ADD execute-pipeline execute-pipeline

RUN curl -Lo snyk https://static.snyk.io/cli/latest/snyk-alpine
RUN mv snyk /bin/snyk && chmod u+x /bin/snyk

CMD [ "sh", "-c", "./execute-pipeline" ]
ENTRYPOINT []
