FROM vault:1.13.3 AS vault
FROM alpine

COPY --from=vault /bin/vault /usr/bin/vault

LABEL org.opencontainers.image.authors="kratix@syntasso.io"
LABEL org.opencontainers.image.source=https://github.com/syntasso/kratix-marketplace

RUN [ "mkdir", "/tmp/transfer" ]
RUN apk update && apk add --no-cache yq curl libcap

# NOTE requires buildx
ARG TARGETARCH
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.25.1/bin/linux/${TARGETARCH}/kubectl
RUN chmod u+x kubectl && mv kubectl /bin/kubectl

ADD execute-pipeline execute-pipeline
ADD readonly.hcl readonly.hcl

CMD [ "sh", "-c", "./execute-pipeline" ]
ENTRYPOINT []
