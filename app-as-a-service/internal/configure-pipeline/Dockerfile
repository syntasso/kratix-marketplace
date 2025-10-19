FROM golang:1.25-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git kubectl
COPY ./scripts/go.mod ./scripts/go.sum ./scripts/main.go ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
RUN go build -o /app/main .

ENTRYPOINT ["/app/main"]

