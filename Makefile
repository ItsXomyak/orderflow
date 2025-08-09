.PHONY: build run-worker run-api clean deps test help

# Variables
BINARY_DIR=bin
WORKER_BINARY=$(BINARY_DIR)/worker
API_BINARY=$(BINARY_DIR)/api

# Default target
all: build

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Build the application
build: deps
	@echo "Building application..."
	mkdir -p $(BINARY_DIR)
	go build -o $(WORKER_BINARY) ./cmd/worker
	go build -o $(API_BINARY) ./cmd/api
	@echo "Build complete!"

# Run the worker
run-worker: build
	@echo "Starting Temporal worker..."
	$(WORKER_BINARY)

# Run the API server
run-api: build
	@echo "Starting API server..."
	$(API_BINARY)

# Run both worker and API (in separate terminals)
run: build
	@echo "Starting both worker and API..."
	@echo "Worker will run in background. Use 'make stop' to stop it."
	$(WORKER_BINARY) &
	@sleep 2
	$(API_BINARY)

# Stop background processes
stop:
	@echo "Stopping background processes..."
	@pkill -f $(WORKER_BINARY) || true
	@pkill -f $(API_BINARY) || true

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BINARY_DIR)
	go clean

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Generate documentation
docs:
	@echo "Generating documentation..."
	godoc -http=:6060

# Install Temporal CLI (if not installed)
install-temporal-cli:
	@echo "Installing Temporal CLI..."
	@if ! command -v temporal &> /dev/null; then \
		echo "Temporal CLI not found. Installing..."; \
		curl -sSf https://temporal.download/cli.sh | sh; \
		export PATH=$PATH:$$HOME/.temporal/bin; \
	else \
		echo "Temporal CLI already installed"; \
	fi

# Start Temporal server (requires Temporal CLI)
start-temporal:
	@echo "Starting Temporal server..."
	temporal server start-dev

# Docker commands
docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-clean:
	docker-compose down -v --remove-orphans
	docker system prune -f

docker-restart:
	docker-compose restart

# Full stack commands
start-full-stack: docker-up
	@echo "Starting full stack with Docker..."
	@echo "PostgreSQL: localhost:5432"
	@echo "Temporal: localhost:7233"
	@echo "Temporal Web: localhost:8088"
	@echo "API: localhost:8080"

stop-full-stack: docker-down
	@echo "Stopped full stack"

# Database commands
db-migrate:
	docker-compose exec postgres psql -U orderflow -d orderflow -f /docker-entrypoint-initdb.d/001_init_schema.sql

db-connect:
	docker-compose exec postgres psql -U orderflow -d orderflow

db-reset: docker-down
	docker volume rm orderflow_postgres_data
	docker-compose up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 10
	@echo "Running migrations..."
	@make db-migrate

# Help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
