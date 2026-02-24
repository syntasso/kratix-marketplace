FROM golang:1.25-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git kubectl curl jq postgresql-client

COPY ./scripts/go.mod ./scripts/go.sum /src/scripts/
WORKDIR /src/scripts
RUN go mod download

COPY . /src
RUN mkdir -p /tmp/transfer/dependencies && cp -R /src/dependencies/. /tmp/transfer/dependencies/

ENV CGO_ENABLED=0
RUN go build -o /app/main /src/scripts

ENTRYPOINT ["/app/main"]
