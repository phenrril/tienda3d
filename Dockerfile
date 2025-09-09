# Multi-stage build
FROM golang:1.22 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o tienda3d ./cmd/tienda3d

FROM alpine:3.20 AS runtime
WORKDIR /app
RUN adduser -D appuser
COPY --from=builder /app/tienda3d /app/
COPY --from=builder /app/internal/views /app/internal/views
COPY --from=builder /app/public /app/public
ENV PORT=8080
EXPOSE 8080
USER appuser
ENTRYPOINT ["/app/tienda3d"]

