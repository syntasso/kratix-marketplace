FROM "alpine"

LABEL org.opencontainers.image.authors "kratix@syntasso.io"
LABEL org.opencontainers.image.source https://github.com/syntasso/kratix-marketplace

RUN [ "mkdir", "/tmp/transfer" ]
RUN apk update && apk add --no-cache yq curl vault libcap
RUN setcap -r /usr/sbin/vault

RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.25.1/bin/linux/amd64/kubectl
RUN chmod u+x kubectl && mv kubectl /bin/kubectl

ADD execute-pipeline execute-pipeline
ADD readonly.hcl readonly.hcl

CMD [ "sh", "-c", "./execute-pipeline" ]
ENTRYPOINT []
