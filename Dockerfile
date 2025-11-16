# ------ build ------
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# si tu main está en ./cmd/tienda3d, perfecto:
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/tienda3d ./cmd/tienda3d

# ------ runtime ------
FROM alpine:3.20
# certificados TLS, herramientas mínimas para healthcheck y postgresql-client para backups
RUN apk add --no-cache ca-certificates tzdata wget postgresql-client
WORKDIR /app
# usuario no-root
RUN adduser -D -H -s /sbin/nologin appuser
# Crear directorio de backups con permisos adecuados
RUN mkdir -p /app/backups && chmod 777 /app/backups
COPY --from=build /out/tienda3d /app/tienda3d
COPY --from=build /src/internal/views /app/internal/views
COPY --from=build /src/public /app/public
USER appuser
EXPOSE 8080
ENTRYPOINT ["/app/tienda3d"]
