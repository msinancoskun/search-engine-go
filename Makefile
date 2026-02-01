.PHONY: help build run test clean migrate-up migrate-down docker-build docker-run

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	go build -o bin/api ./cmd/api

run: ## Run the application
	go run ./cmd/api

test: ## Run tests
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-unit: ## Run unit tests only
	go test -v -short ./...

test-integration: ## Run integration tests only
	go test -v -run Integration ./...

lint: ## Run linter
	golangci-lint run

fmt: ## Format code
	go fmt ./...
	goimports -w .

migrate-up: ## Run database migrations up
	@echo "Running migrations..."
	@for file in internal/infrastructure/database/migrations/*.up.sql; do \
		echo "Applying $$file"; \
		psql $$DATABASE_URL -f $$file || exit 1; \
	done

migrate-down: ## Run database migrations down
	@echo "Rolling back migrations..."
	@for file in internal/infrastructure/database/migrations/*.down.sql; do \
		echo "Rolling back $$file"; \
		psql $$DATABASE_URL -f $$file || exit 1; \
	done

docker-build: ## Build Docker image
	docker build -t search-engine-go:latest .

docker-run: ## Run Docker container
	docker run -p 8080:8080 --env-file .env search-engine-go:latest

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

deps: ## Download dependencies
	go mod download
	go mod tidy

install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
