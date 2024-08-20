FROM "ghcr.io/syntasso/kratix-pipeline-utility:v0.0.1"

LABEL org.opencontainers.image.authors="kratix@syntasso.io"
LABEL org.opencontainers.image.source=https://github.com/syntasso/kratix-marketplace

WORKDIR /gh
ENV PATH="/gh:${PATH}"
ADD create-issue create-issue
ADD wait-approval wait-approval

ENTRYPOINT []
