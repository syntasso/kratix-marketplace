FROM alpine

COPY --from=docker.io/bitnami/kubectl:1.25.5 /opt/bitnami/kubectl/bin/kubectl /usr/local/bin/kubectl
COPY --from=docker.io/hashicorp/waypoint:0.10.5 /usr/bin/waypoint /usr/local/bin/waypoint

RUN mkdir -p /.config/waypoint /etc/xdg/waypoint ${HOME}/.kube
RUN apk update && apk add jq

COPY ./generate-token /usr/local/bin/generate-token

CMD [ "generate-token" ]
