.PHONY: dev dev-backend dev-frontend build test clean help

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: ## Start all services with Docker Compose
	docker-compose up --build

dev-backend: ## Start backend in development mode
	cd backend && go run ./cmd/api

dev-frontend: ## Start frontend in development mode (requires Node.js)
	cd frontend && npm run dev

build: ## Build all services
	docker-compose build

test: ## Run backend tests
	cd backend && go test -v ./...

test-integration: ## Run integration tests
	cd backend && go test -v -tags=integration ./tests/integration/...

clean: ## Clean up Docker resources
	docker-compose down -v
	docker system prune -f

db-migrate: ## Run database migrations
	cd backend && go run ./cmd/api migrate

db-reset: ## Reset database and run migrations
	cd backend && go run ./cmd/api reset