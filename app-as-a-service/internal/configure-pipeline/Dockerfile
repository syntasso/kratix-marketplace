FROM golang:1.25-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git
COPY ./scripts/go.mod ./scripts/go.sum ./scripts/main.go ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
RUN go build -o /app/main .

FROM alpine:3.20
RUN mkdir -p /kratix/output
COPY --from=builder /app/main /main
ENTRYPOINT ["/main"]

