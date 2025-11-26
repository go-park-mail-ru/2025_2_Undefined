# Build stage
FROM golang:1.24.6 AS builder

WORKDIR /app

# Копируем go mod файлы
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект
COPY . .

# Собираем chats-service
RUN CGO_ENABLED=0 GOOS=linux go build -o chats_service ./cmd/chats_service/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Добавляем необходимые пакеты
RUN apk --no-cache add ca-certificates tzdata

# Копируем бинарник из builder
COPY --from=builder /app/chats_service .

# Устанавливаем права на выполнение
RUN chmod +x chats_service

# Открываем порт для gRPC
EXPOSE 50053

# Запускаем сервис
CMD ["./chats_service"]
