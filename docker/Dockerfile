# Dockerfile

FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o marketflow ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/marketflow .

COPY configs/config.yaml ./configs/

EXPOSE 8080

ENTRYPOINT ["./marketflow"]
