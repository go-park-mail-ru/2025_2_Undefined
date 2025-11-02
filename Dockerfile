# Этап 1: Сборка приложения
FROM golang:1.24.6 AS builder

WORKDIR /app

# Копируем только файлы, необходимые для загрузки зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальные файлы
COPY . .

# Собираем оба приложения
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/app/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -o migrate ./cmd/migration/main.go

# Этап 2: Финальный образ
FROM alpine:3.18

WORKDIR /app

# Устанавливаем зависимости для миграций
RUN apk add --no-cache postgresql-client

# Копируем только необходимые артефакты
COPY --from=builder /app/main .
COPY --from=builder /app/migrate .
COPY --from=builder /app/db/migrations ./db/migrations
COPY --from=builder /app/config.yml .
COPY --from=builder /app/.env ./.env

EXPOSE 8080

# Запускаем миграции и приложение
CMD sh -c "./migrate && ./main"