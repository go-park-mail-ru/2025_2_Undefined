# ДЗ 3 - Администрирование СУБД

## Структура

```
db/
├── postgresql.conf          # Конфигурация PostgreSQL
├── create_service_user.sql  # Создание app_user с ограниченными правами
├── setup_user.sh            # Скрипт для автоматического создания пользователя
└── migrations/              # DDL схема БД
```

## Реализовано

1. **postgresql.conf**: max_connections=500, statement_timeout=30s, lock_timeout=10s, логирование для pgbadger
2. **app_user**: DML права (SELECT/INSERT/UPDATE/DELETE), 490 соединений, без DDL
3. **SQL Injection защита**: pgx/v5 с prepared statements ($1, $2)
4. **Connection pool**: MaxConns=100 на микросервис, сбалансировано с max_connections
5. **Мониторинг**: pg_stat_statements, auto_explain, Prometheus+Grafana

## Запуск

```bash
make start  # Запуск БД -> Миграции -> Создание app_user -> Запуск сервисов
```

Makefile автоматически:
1. Поднимает PostgreSQL
2. Выполняет миграции от суперпользователя
3. Создает app_user через db/setup_user.sh
4. Запускает микросервисы с подключением через app_user

## Ключевые решения

**max_connections = 500**
- 3 микросервиса × 100 соединений = 300
- 1 для app.go = 100
- +90 резерв, +10 для суперпользователя (с запасом, чтобы можно было гарантировано подключиться)

**statement_timeout = 30s, lock_timeout = 10s**
- Защита от DOS атак

**listen_addresses = 'localhost,gramm_db'**
- Только локальные подключения и из Docker сети

**app_user без DDL прав**
- Миграции выполняет суперпользователь
- app_user только DML (SELECT/INSERT/UPDATE/DELETE)
- CONNECTION LIMIT 490
