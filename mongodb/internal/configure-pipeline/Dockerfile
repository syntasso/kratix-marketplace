FROM python:3.12-alpine

LABEL org.opencontainers.image.authors="kratix@syntasso.io"
LABEL org.opencontainers.image.source=https://github.com/syntasso/kratix-marketplace

RUN apk update && apk add --no-cache yq kubectl curl git
RUN python3 -m pip install --no-cache-dir \
  "kratix-sdk @ git+https://github.com/syntasso/kratix-python.git@v0.2.0" \
  PyYAML

ADD execute-pipeline.py execute-pipeline.py

ENTRYPOINT ["python", "/execute-pipeline.py"]