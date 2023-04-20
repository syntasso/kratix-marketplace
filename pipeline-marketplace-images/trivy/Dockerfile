FROM "ghcr.io/syntasso/kratix-pipeline-utility:v0.0.1"

LABEL org.opencontainers.image.authors "kratix@syntasso.io"
LABEL org.opencontainers.image.source https://github.com/syntasso/kratix-marketplace

ADD execute-pipeline execute-pipeline

ENV TRIVY_VERSION 0.40.0
RUN curl -Lo trivy.tar.gz https://github.com/aquasecurity/trivy/releases/download/v${TRIVY_VERSION}/trivy_${TRIVY_VERSION}_Linux-64bit.tar.gz
RUN tar -xvzf trivy.tar.gz && chmod u+x trivy && mv trivy /bin/

RUN mkdir -p /trivy
RUN trivy --cache-dir /trivy image --download-db-only
RUN trivy --cache-dir /trivy image --download-java-db-only

CMD [ "sh", "-c", "./execute-pipeline" ]
ENTRYPOINT []
