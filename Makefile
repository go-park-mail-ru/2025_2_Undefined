# Makefile для проекта 2025_2_Undefined

# Запуск всех тестов
test:
	go test -v ./...

# Запуск тестов с покрытием кода, исключая моки и сгенерированный код
test-coverage:
	@echo "Запуск тестов с покрытием кода..."
	go test -v -coverprofile=coverage.out -coverpkg=./... ./...
	@echo "Исключаем docs.go и fill.go из покрытия..."
	grep -v -E "(docs|fill\.go)" coverage.out > coverage_filtered.out || true
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
	docker compose up --build

clear: 
	docker compose down -v