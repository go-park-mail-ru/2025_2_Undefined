#!/bin/bash
# Автоматическая настройка сервисного пользователя БД

set -e

# Загружаем .env файл
if [ -f .env ]; then
    source .env
else
    echo "Ошибка: файл .env не найден"
    exit 1
fi



# Проверяем наличие переменных
if [ -z "$POSTGRES_APP_PASSWORD" ]; then
    echo "Ошибка: POSTGRES_APP_PASSWORD не установлен в .env"
    exit 1
fi

if [ -z "$POSTGRES_DB" ]; then
    echo "Ошибка: POSTGRES_DB не установлен в .env"
    exit 1
fi

if [ -z "$POSTGRES_SUPERUSER" ]; then
    echo "Ошибка: POSTGRES_SUPERUSER не установлен в .env"
    exit 1
fi

docker exec -i gramm_db psql -U "$POSTGRES_SUPERUSER" -d "$POSTGRES_DB" -v pwd="$POSTGRES_APP_PASSWORD" < db/create_service_user.sql

echo "✓ Пользователь $POSTGRES_APP_USER успешно создан"
