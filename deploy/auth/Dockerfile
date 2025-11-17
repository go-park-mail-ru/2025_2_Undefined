# Build stage
FROM golang:1.24.6 AS builder

WORKDIR /app

# Копируем go mod файлы
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект
COPY . .

# Собираем auth-service
RUN CGO_ENABLED=0 GOOS=linux go build -o auth_service ./cmd/auth_service/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Добавляем необходимые пакеты
RUN apk --no-cache add ca-certificates tzdata

# Копируем бинарник из builder
COPY --from=builder /app/auth_service .

# Устанавливаем права на выполнение
RUN chmod +x auth_service

# Открываем порт для gRPC
EXPOSE 50051

# Запускаем сервис
CMD ["./auth_service"]
