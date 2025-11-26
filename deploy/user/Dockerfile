# Build stage
FROM golang:1.24.6 AS builder

WORKDIR /app

# Копируем go mod файлы
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект
COPY . .

# Собираем user-service
RUN CGO_ENABLED=0 GOOS=linux go build -o user_service ./cmd/user_service/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Добавляем необходимые пакеты
RUN apk --no-cache add ca-certificates tzdata

# Копируем бинарник из builder
COPY --from=builder /app/user_service .

# Устанавливаем права на выполнение
RUN chmod +x user_service

# Открываем порт для gRPC
EXPOSE 50052

# Запускаем сервис
CMD ["./user_service"]
