FROM "alpine"

LABEL org.opencontainers.image.authors="kratix@syntasso.io"
LABEL org.opencontainers.image.source=https://github.com/syntasso/kratix-marketplace

RUN [ "mkdir", "/tmp/transfer" ]
RUN apk update && apk add --no-cache yq

ADD resources/* /tmp/transfer/
ADD execute-pipeline execute-pipeline
ADD bootstrapper bootstrapper

RUN apk add curl openssl bash --no-cache
RUN curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 \
    && chmod +x get_helm.sh && ./get_helm.sh

# NOTE requires buildx
ARG TARGETARCH
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.25.1/bin/linux/${TARGETARCH}/kubectl
RUN chmod u+x kubectl && mv kubectl /bin/kubectl

CMD [ "sh", "-c", "./execute-pipeline" ]
ENTRYPOINT []
