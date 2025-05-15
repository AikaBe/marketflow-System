# Dockerfile

FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Собираем бинарник из main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o marketflow ./cmd/main.go

# Финальный образ
FROM alpine:latest

WORKDIR /app

# Копируем бинарник
COPY --from=builder /app/marketflow .

# Копируем конфиг
COPY configs/config.yaml ./configs/

EXPOSE 8080

ENTRYPOINT ["./marketflow"]
