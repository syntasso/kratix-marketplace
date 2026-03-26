FROM golang:1.25-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git kubectl

COPY ./scripts/go.mod ./scripts/go.sum /src/scripts/
WORKDIR /src/scripts
RUN go mod download

COPY . /src
ENV CGO_ENABLED=0
RUN go build -o /app/main /src/scripts && \
    go build -o /app/vault-env-launcher /src/scripts/cmd/vault-env-launcher

ENTRYPOINT ["/app/main"]
