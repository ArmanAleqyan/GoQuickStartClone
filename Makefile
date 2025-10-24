.PHONY: help build run test clean docker-up docker-down proto

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build all services
	@echo "Building all services..."
	@cd services/api-gateway && go build -o ../../bin/api-gateway ./cmd
	@cd services/auth-service && go build -o ../../bin/auth-service ./cmd
	@cd services/user-service && go build -o ../../bin/user-service ./cmd
	@cd services/blockchain-service && go build -o ../../bin/blockchain-service ./cmd
	@cd services/analytics-service && go build -o ../../bin/analytics-service ./cmd
	@cd services/billing-service && go build -o ../../bin/billing-service ./cmd
	@echo "Build complete!"

run-gateway: ## Run API Gateway
	@cd services/api-gateway && go run ./cmd

run-auth: ## Run Auth Service
	@cd services/auth-service && go run ./cmd

run-user: ## Run User Service
	@cd services/user-service && go run ./cmd

run-blockchain: ## Run Blockchain Service
	@cd services/blockchain-service && go run ./cmd

run-analytics: ## Run Analytics Service
	@cd services/analytics-service && go run ./cmd

run-billing: ## Run Billing Service
	@cd services/billing-service && go run ./cmd

test: ## Run tests for all services
	@echo "Running tests..."
	@go test -v ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf tmp/
	@echo "Clean complete!"

docker-up: ## Start all services with Docker Compose
	@docker-compose up -d

docker-down: ## Stop all services
	@docker-compose down

docker-build: ## Build Docker images
	@docker-compose build

proto: ## Generate protobuf files
	@echo "Generating protobuf files..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		pkg/proto/**/*.proto
	@echo "Protobuf generation complete!"

migrate-up: ## Run database migrations up
	@echo "Running migrations..."
	@go run cmd/migrate/main.go up

migrate-down: ## Run database migrations down
	@go run cmd/migrate/main.go down

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
