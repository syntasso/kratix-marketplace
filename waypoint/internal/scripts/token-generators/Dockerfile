FROM alpine/k8s:1.25.5 

COPY --from=docker.io/hashicorp/waypoint:0.10.5 /usr/bin/waypoint /usr/local/bin/waypoint

RUN mkdir -p /.config/waypoint && mkdir -p /etc/xdg/waypoint
RUN apk update && apk add jq

COPY ./generate-token /usr/local/bin/generate-token

CMD [ "generate-token" ]
