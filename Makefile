.PHONY: help build run test clean docker-up docker-down migrate seed

# Переменные
APP_NAME=orderflow
BINARY_NAME=orderflow
DOCKER_COMPOSE_FILE=docker-compose.yaml

help: ## Показать справку
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Собрать приложение
	go build -o bin/$(BINARY_NAME) cmd/main.go

run: ## Запустить приложение
	go run cmd/main.go

test: ## Запустить тесты
	go test ./...

clean: ## Очистить бинарные файлы
	rm -rf bin/

docker-up: ## Запустить инфраструктуру (PostgreSQL + Temporal)
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d postgres temporal temporal-web

docker-down: ## Остановить инфраструктуру
	docker-compose -f $(DOCKER_COMPOSE_FILE) down

docker-logs: ## Показать логи контейнеров
	docker-compose -f $(DOCKER_COMPOSE_FILE) logs -f

docker-build: ## Собрать Docker образ приложения
	docker-compose -f $(DOCKER_COMPOSE_FILE) build app

docker-run: ## Запустить полную систему (инфраструктура + приложение)
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d

docker-stop: ## Остановить все контейнеры
	docker-compose -f $(DOCKER_COMPOSE_FILE) down

docker-restart: ## Перезапустить все контейнеры
	docker-compose -f $(DOCKER_COMPOSE_FILE) restart

docker-clean: ## Остановить и удалить все контейнеры и volumes
	docker-compose -f $(DOCKER_COMPOSE_FILE) down -v --remove-orphans

migrate: ## Применить миграции
	@echo "Миграции применяются автоматически при запуске PostgreSQL"

seed: ## Заполнить демо-данными
	@echo "Демо-данные заполняются автоматически при запуске PostgreSQL"

dev: docker-up ## Запустить в режиме разработки
	@echo "Ожидание запуска инфраструктуры..."
	@sleep 10
	@echo "Запуск приложения..."
	go run cmd/main.go

stop: docker-down ## Остановить все сервисы
	@echo "Все сервисы остановлены"

restart: stop dev ## Перезапустить все сервисы

# Команды для работы с заказами (примеры)
create-order: ## Создать тестовый заказ
	@curl -X POST http://localhost:8080/api/orders \
		-H "Content-Type: application/json" \
		-d '{"customer_id":"customer-001","items":[{"product_id":"prod-001","name":"iPhone 15 Pro","quantity":1,"price":999.99}]}'

get-status: ## Получить статус заказа (нужно указать workflow_id)
	@echo "Использование: make get-status WORKFLOW_ID=<id>"
	@if [ -z "$(WORKFLOW_ID)" ]; then echo "Ошибка: WORKFLOW_ID не указан"; exit 1; fi
	@curl -X GET "http://localhost:8080/api/orders/status?workflow_id=$(WORKFLOW_ID)"

cancel-order: ## Отменить заказ (нужно указать workflow_id)
	@echo "Использование: make cancel-order WORKFLOW_ID=<id>"
	@if [ -z "$(WORKFLOW_ID)" ]; then echo "Ошибка: WORKFLOW_ID не указан"; exit 1; fi
	@curl -X POST "http://localhost:8080/api/orders/cancel?workflow_id=$(WORKFLOW_ID)"

health: ## Проверить здоровье приложения
	@curl -X GET http://localhost:8080/health

# Команды для мониторинга
temporal-ui: ## Открыть Temporal Web UI
	@echo "Temporal Web UI доступен по адресу: http://localhost:8088"
	@open http://localhost:8088 2>/dev/null || echo "Не удалось открыть браузер автоматически"

logs: ## Показать логи приложения
	@echo "Логи приложения (если запущено через docker):"
	@docker-compose -f $(DOCKER_COMPOSE_FILE) logs -f app 2>/dev/null || echo "Приложение не запущено в docker"

# Команды для разработки
deps: ## Установить зависимости
	go mod download
	go mod tidy

fmt: ## Форматировать код
	go fmt ./...

lint: ## Проверить код линтером
	golangci-lint run

vet: ## Проверить код go vet
	go vet ./...

# Полная проверка кода
check: fmt lint vet test ## Выполнить все проверки кода
