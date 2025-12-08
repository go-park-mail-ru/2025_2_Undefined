# Makefile для проекта 2025_2_Undefined

# Переменные для подключения к БД
DB_URL=postgres://user:password@localhost:5433/gramm?sslmode=disable
MIGRATIONS_PATH=db/migrations
CONFIG_SOURCE=config.yml
ENV_FILE=.env
VERSION=3

# Миграции базы данных (через CLI migrate)
db-up:
	@echo "Применение всех миграций через CLI..."
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up

db-down:
	@echo "Откат всех миграций через CLI..."
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down -all

db-down-1:
	@echo "Откат последней миграции через CLI..."
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down 1

db-version:
	@echo "Текущая версия миграций..."
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" version

db-goto:
	@echo "Миграция к версии $(VERSION)..."
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" goto $(VERSION)

db-drop:
	@echo "Удаление всех объектов из БД..."
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" drop -f

db-force:
	@echo "Принудительная установка версии $(VERSION)..."
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" force $(VERSION)

# Запуск всех тестов
test:
	go test -v ./...

# Запуск тестов с покрытием кода, исключая моки и сгенерированный код
test-coverage:
	@echo "Очистка кэша и старых файлов покрытия..."
	@rm -f coverage.out coverage_filtered.out
	go clean -testcache
	@echo "Запуск тестов с покрытием кода..."
	go test -v -coverprofile=coverage.out -coverpkg=./... ./...
	@echo "Исключаем docs.go, fill.go, mock*.go, main.go, config.go, app.go из покрытия..."
	grep -v -E "(docs|fill\.go|mock.*\.go|generate\.go|cmd/.*main\.go|config/config\.go|internal/app/app\.go|generated/*)" coverage.out > coverage_filtered.out || true
	@echo "Результаты покрытия:"
	go tool cover -func=coverage_filtered.out | grep total

# Создание HTML отчета о покрытии
test-coverage-html: test-coverage
	@echo "Создание HTML отчета..."
	go tool cover -html=coverage_filtered.out -o coverage.html
	@echo "HTML отчет создан: coverage.html"

# Запуск приложения
run:
	@echo "Запуск приложения..."
	go run ./cmd/app/main.go

# Установка зависимостей
deps:
	@echo "Установка зависимостей..."
	go mod download
	go mod tidy

start:
	make swagger
	docker compose up --build

start-background:
	make swagger
	docker compose up --build -d

logs:
	docker-compose logs -f

stop:
	docker compose stop

all-clear:
	docker compose down
	docker compose down -v

clear: 
	@echo "Остановка приложения и очистка БД..."
	docker compose down
	docker compose up -d db auth_redis
	sleep 2
	docker compose run --rm app ./migrate down
	docker compose run --rm app ./migrate up
	docker compose down
	@echo "Очистка завершена"

clear-redis:
	@echo "Очистка Redis..."
	docker compose up -d auth_redis
	sleep 2
	docker compose run --rm app ./migrate clear-redis
	docker compose down
	@echo "Redis очищен"

swagger:
	swag init -g cmd/app/main.go -o docs

generate-mocks:
	@echo "Генерация моков через go generate..."
	go generate ./...

create-env:
	@if [ ! -f $(ENV_FILE) ]; then \
		echo "Generating $(ENV_FILE) from $(CONFIG_SOURCE)..."; \
		python3 -c "import yaml; config = yaml.safe_load(open('$(CONFIG_SOURCE)')); open('$(ENV_FILE)', 'w').write('\n'.join([f'{k}={v}' for k, v in config.items()]))"; \
		echo "$(ENV_FILE) file generated!"; \
	else \
		echo "$(ENV_FILE) already exists. Skipping..."; \
	fi

auth_proto:
	@mkdir -p internal/transport/generated/auth && \
	protoc --proto_path=proto \
		--go_out=internal/transport/generated/auth \
		--go-grpc_out=internal/transport/generated/auth \
		--go-grpc_opt=paths=source_relative \
		--go_opt=paths=source_relative \
		proto/auth.proto

user_proto:
	@mkdir -p internal/transport/generated/user && \
	protoc --proto_path=proto \
		--go_out=internal/transport/generated/user \
		--go-grpc_out=internal/transport/generated/user \
		--go-grpc_opt=paths=source_relative \
		--go_opt=paths=source_relative \
		proto/user.proto

chats_proto:
	@mkdir -p internal/transport/generated/chats && \
	protoc --proto_path=proto \
		--go_out=internal/transport/generated/chats \
		--go-grpc_out=internal/transport/generated/chats \
		--go-grpc_opt=paths=source_relative \
		--go_opt=paths=source_relative \
		proto/chats.proto