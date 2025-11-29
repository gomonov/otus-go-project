FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o application ./cmd

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/application .
COPY --from=builder /app/configs/config.toml ./config.toml
COPY --from=builder /app/migrations ./migrations

RUN adduser -D -s /bin/sh appuser && chown -R appuser:appuser /root/

USER appuser

EXPOSE 8080

CMD ["./application", "-config", "./config.toml"]