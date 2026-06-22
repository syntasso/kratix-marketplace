FROM python:3.12-alpine

LABEL org.opencontainers.image.authors="kratix@syntasso.io"
LABEL org.opencontainers.image.source=https://github.com/syntasso/kratix-marketplace

# git is needed to pip-install the SDK from its git tag; kubernetes + PyYAML
# come transitively with kratix-sdk.
RUN apk add --no-cache git
RUN python3 -m pip install --no-cache-dir \
  "kratix-sdk @ git+https://github.com/syntasso/kratix-python.git@v0.2.0"

ADD claim.py claim.py

ENTRYPOINT ["python", "/claim.py"]
