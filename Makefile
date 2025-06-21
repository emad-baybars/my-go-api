.PHONY: help build run test clean docker-build docker-run docker-stop swagger deps lint format

# Variables
APP_NAME := backend-template
DOCKER_IMAGE := $(APP_NAME):latest
DOCKER_CONTAINER := $(APP_NAME)-container

# Help command
help: ## Show this help message
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development commands
deps: ## Install dependencies
	go mod download
	go mod tidy

run: ## Run the application locally
	go run main.go

build: ## Build the application
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/$(APP_NAME) .

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run linter
	golangci-lint run

format: ## Format code
	go fmt ./...
	goimports -w .

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

# Swagger documentation
swagger: ## Generate Swagger documentation
	swag init -g main.go -o docs/

# Docker commands
docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Run application in Docker container
	docker run -d --name $(DOCKER_CONTAINER) -p 8080:8080 $(DOCKER_IMAGE)

docker-stop: ## Stop and remove Docker container
	docker stop $(DOCKER_CONTAINER) || true
	docker rm $(DOCKER_CONTAINER) || true

docker-logs: ## Show Docker container logs
	docker logs -f $(DOCKER_CONTAINER)

# Docker Compose commands
compose-up: ## Start all services with Docker Compose
	docker-compose up -d

compose-down: ## Stop all services with Docker Compose
	docker-compose down

compose-logs: ## Show Docker Compose logs
	docker-compose logs -f

compose-build: ## Build all services with Docker Compose
	docker-compose build

# Database commands
db-migrate-up: ## Run database migrations up
	migrate -path migrations -database "postgres://postgres:password123@localhost:5432/backend_template?sslmode=disable" up

db-migrate-down: ## Run database migrations down
	migrate -path migrations -database "postgres://postgres:password123@localhost:5432/backend_template?sslmode=disable" down

db-create-migration: ## Create a new migration file (usage: make db-create-migration NAME=migration_name)
	migrate create -ext sql -dir migrations $(NAME)

# Development setup
setup: deps swagger ## Setup development environment
	cp .env.example .env
	@echo "Development environment setup complete!"
	@echo "Please update .env file with your configuration"

# Production commands
deploy: build ## Deploy to production
	@echo "Deploying to production..."
	# Add your deployment commands here

# Security commands
security-check: ## Run security vulnerability check
	gosec ./...

# Performance commands
benchmark: ## Run benchmarks
	go test -bench=. -benchmem ./...

# Git hooks
install-hooks: ## Install git hooks
	cp scripts/pre-commit .git/hooks/
	chmod +x .git/hooks/pre-commit

# All-in-one commands
dev: deps swagger run ## Setup and run development environment

ci: deps lint test security-check ## Run CI pipeline

# Default command
.DEFAULT_GOAL := help
