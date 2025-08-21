# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Путь к пакету с main (относительно контекста сборки)
ARG CMD_DIR=./

# 🔎 DEBUG: выводим версии/пути и сразу собираем
RUN set -eux; \
    go version; \
    pwd; \
    ls -la; \
    echo "Building ${CMD_DIR}"; \
    ls -la "${CMD_DIR}" || true; \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/app "${CMD_DIR}"

FROM alpine:3.20
RUN adduser -D -g '' appuser && apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/app /usr/local/bin/app
COPY config ./config
USER appuser
EXPOSE 8080
ENTRYPOINT ["app"]