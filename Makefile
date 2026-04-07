# ==============================================================================
# Makefile - Skeleton API
# ==============================================================================

.PHONY: help

# ==============================================================================
# VARIABLES
# ==============================================================================

# Version
VERSION_MAJOR ?= 0
VERSION_MINOR ?= 1
VERSION_PATCH ?= 0
VERSION_STAGE ?= dev
VERSION       = $(VERSION_MAJOR).$(VERSION_MINOR).$(VERSION_PATCH)-$(VERSION_STAGE)
COMMIT        ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_TIME     = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Colors for help
CYAN    := \033[36m
GREEN   := \033[32m
YELLOW  := \033[33m
RESET   := \033[0m

# ==============================================================================
# HELP
# ==============================================================================

help: ## Show this help message
	@echo "$(CYAN)Skeleton API - Available Commands$(RESET)\n"
	@awk 'BEGIN {FS = ":.*##"; section = ""} \
		/^#[=]+/ {next} \
		/^# [A-Z]+/ {section = $$0; gsub(/^# /, "", section); printf "\n$(YELLOW)%s$(RESET)\n", section} \
		/^[a-zA-Z_-]+:/ { \
			if ($$0 !~ /##/) next; \
			printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2 \
		} \
		{next}' $(MAKEFILE_LIST)

# ==============================================================================
# BUILD
# ==============================================================================

.PHONY: build run clean tidy

build: ## Build binary with version info
	go build \
	  -ldflags="-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)" \
	  -o bin/api ./cmd/api

run: build ## Build and run the application
	./bin/api

clean: ## Clean build artifacts
	rm -rf bin/ coverage.out coverage.html docs/api/

tidy: ## Tidy go modules
	go mod tidy

# ==============================================================================
# TESTING
# ==============================================================================

.PHONY: test test-cover test-race test-p0

test: ## Run all tests
	go test ./... -timeout 30s

test-cover: ## Run tests with coverage report
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

test-race: ## Run tests with race detector
	go test -race ./...

test-p0: ## Run critical domain tests only
	go test ./internal/identity/domain/... ./pkg/eventbus/memory/... -v

# ==============================================================================
# CODE QUALITY
# ==============================================================================

.PHONY: lint

lint: ## Run golangci-lint
	golangci-lint run ./...

# ==============================================================================
# DOCKER
# ==============================================================================

.PHONY: docker-build docker-dev docker-prod docker-up docker-down docker-logs docker-ps docker-clean

docker-build: ## Build production Docker image
	docker build \
	  --build-arg VERSION_MAJOR=$(VERSION_MAJOR) \
	  --build-arg VERSION_MINOR=$(VERSION_MINOR) \
	  --build-arg VERSION_PATCH=$(VERSION_PATCH) \
	  --build-arg VERSION_STAGE=$(VERSION_STAGE) \
	  --build-arg COMMIT=$(COMMIT) \
	  --build-arg BUILD_TIME=$(BUILD_TIME) \
	  -t skeleton-api:$(VERSION) \
	  .

docker-dev: ## Start development with hot reload
	docker-compose up --build

docker-prod: ## Start production containers
	docker-compose -f docker-compose.prod.yml up --build -d

docker-up: ## Start containers in background
	docker-compose up -d

docker-down: ## Stop and remove containers
	docker-compose down

docker-logs: ## View container logs
	docker-compose logs -f

docker-ps: ## List running containers
	docker-compose ps

docker-clean: ## Remove containers, volumes, and images
	docker-compose down -v --rmi local

docker-dev-redis: ## Start development with Redis
	docker-compose --profile redis up --build

# ==============================================================================
# DATABASE
# ==============================================================================

.PHONY: migrate-up migrate-down seed

migrate-up: ## Apply database migrations
	go run ./scripts/migrate/ up

migrate-down: ## Rollback last migration
	go run ./scripts/migrate/ down

seed: ## Seed development data
	go run ./scripts/seed/

# ==============================================================================
# DOCUMENTATION
# ==============================================================================

.PHONY: swagger swagger-serve

swagger: ## Generate OpenAPI/Swagger documentation
	swag init -g cmd/api/main.go -o docs/api --parseDependency --parseInternal

swagger-serve: swagger ## Generate and serve Swagger UI
	swag serve -d docs/api

# ==============================================================================
# SETUP
# ==============================================================================

.PHONY: keys

keys: ## Generate RSA key pair for JWT
	mkdir -p keys
	openssl genrsa -out keys/private.pem 2048
	openssl rsa -in keys/private.pem -pubout -out keys/public.pem

# ==============================================================================
# DEVELOPMENT
# ==============================================================================

.PHONY: dev setup

dev: migrate-up seed run ## Quick start: migrate, seed, and run

setup: keys migrate-up seed ## Initial setup: keys, migrate, seed
	@echo "\n$(GREEN)Setup complete!$(RESET)"
	@echo "Run with: $(CYAN)make run$(RESET) or $(CYAN)make docker-dev$(RESET)"

# ==============================================================================
# CI/CD
# ==============================================================================

.PHONY: ci

ci: lint test ## Run CI checks (lint + test)
	@echo "$(GREEN)All CI checks passed$(RESET)"

# ==============================================================================
# UTILITIES
# ==============================================================================

.PHONY: fmt deps

fmt: ## Format code
	go fmt ./...

deps: ## Show dependencies
	go mod graph

.DEFAULT_GOAL := help