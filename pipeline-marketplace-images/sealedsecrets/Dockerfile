FROM "ghcr.io/syntasso/kratix-pipeline-utility:v0.0.1"

LABEL org.opencontainers.image.authors="kratix@syntasso.io"
LABEL org.opencontainers.image.source=https://github.com/syntasso/kratix-marketplace

ADD execute-pipeline execute-pipeline

ARG KUBESEAL_VERSION=0.24.1
# NOTE requires buildx
ARG TARGETARCH
RUN curl -Lo kubeseal.tar.gz https://github.com/bitnami-labs/sealed-secrets/releases/download/v${KUBESEAL_VERSION}/kubeseal-${KUBESEAL_VERSION}-linux-${TARGETARCH}.tar.gz
RUN tar -xvzf kubeseal.tar.gz kubeseal && chmod u+x kubeseal && mv kubeseal /bin/kubeseal

ENV KUBESEAL_CONFIGMAP=sealed-secrets
ENV KUBESEAL_NAMESPACE=default

CMD [ "sh", "-c", "./execute-pipeline" ]
ENTRYPOINT []
