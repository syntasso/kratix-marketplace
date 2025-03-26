FROM "ghcr.io/syntasso/kratix-pipeline-utility:v0.0.1"

LABEL org.opencontainers.image.authors="kratix@syntasso.io"
LABEL org.opencontainers.image.source=https://github.com/syntasso/kratix-marketplace

ADD execute-pipeline execute-pipeline
# NOTE requires buildx
ARG TARGETARCH
RUN curl https://releases.hashicorp.com/vault/1.16.2/vault_1.16.2_linux_${TARGETARCH}.zip -o vault.zip && unzip vault.zip && mv vault /usr/sbin/vault && rm vault.zip

CMD [ "sh", "-c", "./execute-pipeline" ]
ENTRYPOINT []
