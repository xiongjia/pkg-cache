FROM golang:1.21-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o pkg_cache ./cmd/pkgcache

FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /build/pkg_cache .
COPY --from=builder /build/config.yaml .

RUN mkdir -p /cache && chown -R nobody:nobody /cache

USER nobody

CMD ["./pkg_cache"]
