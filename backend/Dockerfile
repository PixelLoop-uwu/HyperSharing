# Build stage
FROM golang:1.27-alpine AS build
WORKDIR /src

# Cache module downloads before copying the source, so rebuilds are cheap.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/hypersharing ./cmd/app
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/hypersharing-worker ./cmd/worker
# goose CLI для one-shot migrate-сервиса (самодостаточно, без внешнего образа).
RUN GOBIN=/out CGO_ENABLED=0 go install github.com/pressly/goose/v3/cmd/goose@v3.29.0

# Runtime stage
FROM alpine:3.21
# ca-certificates для будущих HTTPS-подключений к S3; wget (busybox) — для healthcheck.
RUN apk add --no-cache ca-certificates \
  && adduser -D -H -u 10001 app
COPY --from=build /out/hypersharing /usr/local/bin/hypersharing
COPY --from=build /out/hypersharing-worker /usr/local/bin/hypersharing-worker
COPY --from=build /out/goose /usr/local/bin/goose
USER app
EXPOSE 8000
ENTRYPOINT ["hypersharing"]
