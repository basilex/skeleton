# ===========================================
# Skeleton CRM - Monorepo Development
# ===========================================

.PHONY: help install dev backend frontend build test clean db

# Default target
.DEFAULT_GOAL := help

# ===========================================
# Help
# ===========================================

help: ## Show this help message
	@echo ''
	@echo 'Skeleton CRM - Monorepo Development Commands'
	@echo ''
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Development:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## DEV/ {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ''
	@echo 'Backend:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## BE/ {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ''
	@echo 'Frontend:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## FE/ {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ''
	@echo 'Database:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## DB/ {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ''
	@echo 'Build & Deploy:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## BUILD/ {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ''
	@echo 'Testing:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## TEST/ {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ===========================================
# Installation
# ===========================================

install: ## DEV Install all dependencies
	@echo "Installing backend dependencies..."
	cd backend && go mod download
	@echo "Installing frontend dependencies..."
	cd frontend && npm install
	@echo "Done!"

install-backend: ## BE Install backend dependencies
	cd backend && go mod download

install-frontend: ## FE Install frontend dependencies
	cd frontend && npm install

# ===========================================
# Development
# ===========================================

dev: ## DEV Start all services (postgres + redis + backend + frontend)
	docker-compose up --build

dev-detach: ## DEV Start all services in background
	docker-compose up -d --build

dev-logs: ## DEV View logs from all services
	docker-compose logs -f

dev-stop: ## DEV Stop all services
	docker-compose down

dev-reset: ## DEV Stop services and remove volumes
	docker-compose down -v

# ===========================================
# Backend Development
# ===========================================

backend: ## BE Start backend server (requires postgres + redis)
	cd backend && go run ./cmd/api

backend-watch: ## BE Start backend with hot reload (requires air)
	cd backend && air

backend-build: ## BE Build backend binary
	cd backend && go build -o bin/api ./cmd/api

DB_URL ?= postgres://skeleton:skeleton@localhost:5432/skeleton?sslmode=disable
REDIS_URL ?= redis://localhost:6379

# ===========================================
# Frontend Development
# ===========================================

frontend: ## FE Start frontend dev server
	cd frontend && npm run dev

frontend-build: ## FE Build frontend for production
	cd frontend && npm run build

frontend-start: ## FE Start frontend production server
	cd frontend && npm run start

frontend-lint: ## FE Run linter
	cd frontend && npm run lint

frontend-typecheck: ## FE Run TypeScript type checking
	cd frontend && npx tsc --noEmit

# ===========================================
# Database
# ===========================================

db-up: ## DB Start PostgreSQL with Docker
	docker-compose up -d postgres redis

db-migrate: ## DB Run database migrations
	cd backend && go run ./cmd/api migrate

db-reset: ## DB Reset database and run migrations
	cd backend && go run ./cmd/api reset

db-seed: ## DB Seed database with sample data
	cd backend && go run ./scripts/seed/main.go

db-shell: ## DB Open PostgreSQL shell
	docker-compose exec postgres psql -U skeleton -d skeleton

# ===========================================
# Testing
# ===========================================

test: ## TEST Run all backend tests
	cd backend && go test -v ./...

test-coverage: ## TEST Run tests with coverage
	cd backend && go test -v -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

test-integration: ## TEST Run integration tests
	cd backend && go test -v -tags=integration ./tests/integration/...

test-frontend: ## TEST Run frontend tests
	cd frontend && npm test

# ===========================================
# Build & Deploy
# ===========================================

build: ## BUILD Build all services
	docker-compose build

build-backend: ## BUILD Build backend Docker image
	docker build -t skeleton-backend:latest ./backend

build-frontend: ## BUILD Build frontend Docker image
	docker build -t skeleton-frontend:latest ./frontend

clean: ## BUILD Clean up Docker resources
	docker-compose down -v
	docker system prune -f

# ===========================================
# Code Quality
# ===========================================

lint: ## Run linters for all code
	cd backend && golangci-lint run
	cd frontend && npm run lint

fmt: ## Format code
	cd backend && go fmt ./...
	cd frontend && npx prettier --write .

# ===========================================
# Scripts
# ===========================================

scripts-migrate: ## Run migration script manually
	cd backend && go run ./scripts/migrate/main.go -action=up

scripts-benchmark: ## Run benchmarks
	cd scripts && ./run-benchmarks.sh

scripts-deploy-staging: ## Deploy to staging
	cd scripts/deploy && ./deploy-staging.sh