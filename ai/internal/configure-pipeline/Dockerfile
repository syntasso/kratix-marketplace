FROM golang:1.25-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
RUN go build -o /app/main .

FROM alpine:3.20
LABEL org.opencontainers.image.authors="kratix@syntasso.io"
LABEL org.opencontainers.image.source="https://github.com/syntasso/kratix-marketplace"
RUN mkdir -p /kratix/output
COPY --from=builder /app/main /main
ENTRYPOINT ["/main"]
