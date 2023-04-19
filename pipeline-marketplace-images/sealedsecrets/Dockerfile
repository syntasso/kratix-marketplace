FROM "ghcr.io/syntasso/kratix-pipeline-utility:v0.0.1"

LABEL org.opencontainers.image.authors "kratix@syntasso.io"
LABEL org.opencontainers.image.source https://github.com/syntasso/kratix-marketplace

ADD execute-pipeline execute-pipeline

ENV KUBESEAL_VERSION 0.20.5
RUN curl -Lo kubeseal.tar.gz https://github.com/bitnami-labs/sealed-secrets/releases/download/v${KUBESEAL_VERSION}/kubeseal-${KUBESEAL_VERSION}-linux-amd64.tar.gz
RUN tar -xvzf kubeseal.tar.gz kubeseal && chmod u+x kubeseal && mv kubeseal /bin/kubeseal

CMD [ "sh", "-c", "./execute-pipeline" ]
ENTRYPOINT []
