FROM golang:1.27.3-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/seed ./cmd/seed \
 && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/gameserver ./cmd/gameserver \
 && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/gateway ./cmd/gateway
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends bash ca-certificates curl && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /out/seed /usr/local/bin/seed
COPY --from=build /out/gameserver /usr/local/bin/gameserver
COPY --from=build /out/gateway /usr/local/bin/gateway
COPY config.example.yaml /app/config.example.yaml
COPY server_data/static.sqlite3 /app/server_data/static.sqlite3
COPY docker/entrypoint.sh /usr/local/bin/rf4-entrypoint
RUN chmod 0755 /usr/local/bin/rf4-entrypoint && mkdir -p /app/data /app/certs
EXPOSE 443/tcp 9341/tcp
VOLUME ["/app/data", "/app/certs"]
ENTRYPOINT ["/usr/local/bin/rf4-entrypoint"]
