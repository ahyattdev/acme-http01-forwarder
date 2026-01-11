FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /acme-http01-forwarder .

FROM alpine:3.23

LABEL org.opencontainers.image.source=https://github.com/ahyattdev/acme-http01-forwarder

COPY --from=builder /acme-http01-forwarder /usr/local/bin/acme-http01-forwarder

EXPOSE 80

ENTRYPOINT ["/usr/local/bin/acme-http01-forwarder"]
