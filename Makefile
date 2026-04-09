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
RED     := \033[31m
RESET   := \033[0m

# Docker container names
DEV_DB_CONTAINER    ?= skeleton-postgres-dev
STAGING_DB_CONTAINER ?= skeleton-postgres-staging
PROD_DB_CONTAINER   ?= skeleton-postgres-prod
TEST_DB_CONTAINER   ?= skeleton-postgres-test

# Database credentials (from docker-compose)
DEV_DB_USER ?= skeleton
DEV_DB_NAME ?= skeleton
DEV_DB_PASS ?= skeleton_password
DEV_DB_HOST ?= localhost
DEV_DB_PORT ?= 5432

# Construct DATABASE_URL for migrations
DATABASE_URL ?= postgres://$(DEV_DB_USER):$(DEV_DB_PASS)@$(DEV_DB_HOST):$(DEV_DB_PORT)/$(DEV_DB_NAME)?sslmode=disable

# Export DATABASE_URL for migration scripts
export DATABASE_URL

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

.PHONY: test test-unit test-integration test-cover test-race test-p0

test: ## Run all tests (requires Docker for integration tests)
	go test ./... -timeout 30s

test-unit: ## Run unit tests only (no database required, fast)
	@echo "Running unit tests (domain + application layers)..."
	@go test ./internal/audit/domain/... ./internal/audit/application/... -v -timeout 30s
	@go test ./internal/files/domain/... ./internal/files/application/... -v -timeout 30s
	@go test ./internal/identity/domain/... ./internal/identity/application/... -v -timeout 30s
	@go test ./internal/notifications/domain/... ./internal/notifications/application/... -v -timeout 30s
	@go test ./internal/tasks/domain/... ./internal/tasks/application/... -v -timeout 30s
	@go test ./internal/status/domain/... -v -timeout 30s
	@echo "✅ Unit tests passed"

test-integration: ## Run integration tests (requires Docker)
	@echo "Checking Docker availability..."
	@docker ps > /dev/null 2>&1 || (echo "ERROR: Docker must be running for integration tests" && exit 1)
	@echo "Running integration tests (persistence layer)..."
	go test ./internal/audit/infrastructure/persistence/... -v -timeout 60s
	go test ./internal/files/infrastructure/persistence/... -v -timeout 60s
	go test ./internal/identity/infrastructure/persistence/... -v -timeout 60s
	go test ./internal/notifications/infrastructure/persistence/... -v -timeout 60s
	go test ./internal/tasks/infrastructure/persistence/... -v -timeout 60s
	@echo "✅ Integration tests passed"

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
# DOCKER - CONTAINER MANAGEMENT
# ==============================================================================

.PHONY: docker-up docker-down docker-stop docker-start docker-restart
.PHONY: docker-drop docker-reset docker-logs docker-ps docker-status
.PHONY: docker-build docker-dev docker-prod docker-clean

docker-up: ## Start containers in background (detached mode)
	@echo "$(CYAN)Starting containers...$(RESET)"
	docker-compose up -d
	@echo "$(GREEN)✓ Containers started$(RESET)"
	@echo "\n$(YELLOW)Check status:$(RESET) make docker-status"

docker-down: ## Stop and remove containers (keeps volumes)
	@echo "$(CYAN)Stopping containers...$(RESET)"
	docker-compose down
	@echo "$(GREEN)✓ Containers stopped and removed$(RESET)"

docker-stop: ## Stop containers (keeps containers, doesn't remove)
	@echo "$(CYAN)Stopping containers...$(RESET)"
	docker-compose stop
	@echo "$(GREEN)✓ Containers stopped$(RESET)"

docker-start: ## Start stopped containers
	@echo "$(CYAN)Starting containers...$(RESET)"
	docker-compose start
	@echo "$(GREEN)✓ Containers started$(RESET)"

docker-restart: ## Restart containers
	@echo "$(CYAN)Restarting containers...$(RESET)"
	docker-compose restart
	@echo "$(GREEN)✓ Containers restarted$(RESET)"

docker-drop: ## Stop, remove containers AND delete volumes (⚠️ DESTRUCTIVE)
	@echo "$(RED)⚠️  This will DELETE ALL DATA!$(RESET)"
	@if [ -z "$(CONFIRM)" ]; then \
		echo "$(YELLOW)Press Ctrl+C to cancel, Enter to continue...$(RESET)"; \
		read confirm; \
	fi
	docker-compose down -v
	@echo "$(GREEN)✓ Containers and volumes removed$(RESET)"

docker-reset: ## Complete reset: drop volumes, rebuild, start fresh
	@echo "$(RED)⚠️  Complete reset - ALL DATA WILL BE LOST!$(RESET)"
	@echo "$(YELLOW)This will:$(RESET)"
	@echo "  1. Stop and remove all containers"
	@echo "  2. Delete all volumes (database data)"
	@echo "  3. Rebuild containers from scratch"
	@echo "  4. Apply migrations"
	@echo "\n$(YELLOW)Press Ctrl+C to cancel, Enter to continue...$(RESET)"
	@read confirm
	@docker-compose down -v
	@docker-compose up -d --build
	@echo "$(YELLOW)Waiting for containers to be healthy...$(RESET)"
	@sleep 10
	@$(MAKE) migrate-up
	@echo "$(GREEN)✓ Environment reset complete!$(RESET)"

docker-logs: ## View all container logs (follow mode)
	docker-compose logs -f

docker-logs-app: ## View API container logs only
	docker-compose logs -f skeleton-api-dev

docker-logs-db: ## View PostgreSQL container logs only
	docker-compose logs -f postgres

docker-logs-redis: ## View Redis container logs only
	docker-compose logs -f redis

docker-ps: ## List running containers (basic info)
	docker-compose ps

docker-status: ## Detailed container status with health checks
	@echo "$(CYAN)Container Status:$(RESET)\n"
	@docker ps --filter "name=skeleton" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" || echo "$(YELLOW)No containers running$(RESET)"
	@echo "\n$(CYAN)Health Checks:$(RESET)"
	@docker ps --filter "name=skeleton" --format "{{.Names}}: {{.Status}}" | grep -E "(healthy|unhealthy)" || echo "$(YELLOW)No health check info$(RESET)"
	@echo "\n$(YELLOW)Quick commands:$(RESET)"
	@echo "  make docker-logs-app   - View API logs"
	@echo "  make docker-logs-db    - View PostgreSQL logs"
	@echo "  make psql              - Connect to database"
	@echo "  make docker-status     - Show this status"

docker-build: ## Build production Docker image
	@echo "$(CYAN)Building production image...$(RESET)"
	docker build \
	  --build-arg VERSION_MAJOR=$(VERSION_MAJOR) \
	  --build-arg VERSION_MINOR=$(VERSION_MINOR) \
	  --build-arg VERSION_PATCH=$(VERSION_PATCH) \
	  --build-arg VERSION_STAGE=$(VERSION_STAGE) \
	  --build-arg COMMIT=$(COMMIT) \
	  --build-arg BUILD_TIME=$(BUILD_TIME) \
	  -t skeleton-api:$(VERSION) \
	  .

docker-dev: ## Start development environment (PostgreSQL + Redis + Hot Reload)
	@echo "$(CYAN)Starting development environment...$(RESET)"
	@echo "$(YELLOW)This may take a few minutes on first run...$(RESET)"
	docker-compose up --build

docker-prod: ## Start production containers in background
	@echo "Starting production environment..."
	docker-compose -f docker-compose.prod.yml up --build -d

docker-clean: ## Remove stopped containers, networks, and images
	docker-compose down -v --rmi local
	@echo "$(GREEN)✓ Cleanup complete$(RESET)"

# ==============================================================================
# DOCKER - UTILITY COMMANDS
# ==============================================================================

.PHONY: docker-shell docker-shell-root

docker-shell: ## Get shell in API container
	@echo "$(CYAN)Opening shell in api container...$(RESET)"
	docker exec -it skeleton-api-dev sh

docker-shell-root: ## Get root shell in API container
	@echo "$(CYAN)Opening root shell in api container...$(RESET)"
	docker exec -it -u root skeleton-api-dev sh

docker-exec: ## Execute command in container (usage: make docker-exec CMD="ls -la")
	@echo "$(CYAN)Executing command in api container...$(RESET)"
	@if [ -z "$(CMD)" ]; then \
		echo "$(YELLOW)Usage: make docker-exec CMD='ls -la'$(RESET)"; \
		exit 1; \
	fi
	docker exec skeleton-api-dev $(CMD)

docker-check: ## Check if Docker containers are running
	@echo "$(CYAN)Checking Docker containers...$(RESET)"
	@docker ps --filter "name=skeleton-postgres-dev" --format "{{.Names}}: {{.Status}}" || echo "$(YELLOW)Dev containers not running$(RESET)"
	@docker ps --filter "name=skeleton-postgres-staging" --format "{{.Names}}: {{.Status}}" || true
	@docker ps --filter "name=skeleton-postgres-prod" --format "{{.Names}}: {{.Status}}" || true

# ==============================================================================
# DATABASE - CONNECTION & QUERIES
# ==============================================================================

.PHONY: psql psql-staging psql-prod psql-test db-console
.PHONY: db-tables db-stats db-migrations db-connections

psql: ## Connect to development database (psql in Docker)
	@echo "$(CYAN)Connecting to development database (in Docker)...$(RESET)"
	@docker ps --filter "name=$(DEV_DB_CONTAINER)" --format "{{.Names}}" | grep -q "$(DEV_DB_CONTAINER)" || (echo "$(YELLOW)Container $(DEV_DB_CONTAINER) is not running. Run: make docker-up$(RESET)" && exit 1)
	docker exec -it $(DEV_DB_CONTAINER) psql -U $(DEV_DB_USER) -d $(DEV_DB_NAME)

psql-staging: ## Connect to staging database (psql in Docker)
	@echo "$(CYAN)Connecting to staging database (in Docker)...$(RESET)"
	docker exec -it $(STAGING_DB_CONTAINER) psql -U skeleton -d skeleton_staging

psql-prod: ## Connect to production database (psql in Docker)
	@echo "$(CYAN)Connecting to production database (in Docker)...$(RESET)"
	docker exec -it $(PROD_DB_CONTAINER) psql -U skeleton -d skeleton

psql-test: ## Connect to test database (psql in Docker)
	@echo "$(CYAN)Connecting to test database (in Docker)...$(RESET)"
	docker exec -it $(TEST_DB_CONTAINER) psql -U test -d skeleton_test

db-console: psql ## Alias for psql

db-tables: ## List all database tables
	@echo "$(CYAN)Database Tables:$(RESET)\n"
	@docker ps --filter "name=$(DEV_DB_CONTAINER)" --format "{{.Names}}" | grep -q "$(DEV_DB_CONTAINER)" || (echo "$(YELLOW)Container $(DEV_DB_CONTAINER) is not running. Run: make docker-up$(RESET)" && exit 1)
	docker exec $(DEV_DB_CONTAINER) psql -U $(DEV_DB_USER) -d $(DEV_DB_NAME) -c "\dt"

db-migrations: ## Show migration history
	@echo "$(CYAN)Migration History:$(RESET)\n"
	docker exec $(DEV_DB_CONTAINER) psql -U $(DEV_DB_USER) -d $(DEV_DB_NAME) -c "SELECT version, applied_at FROM schema_migrations ORDER BY applied_at DESC LIMIT 20;"

db-stats: ## Show database statistics (sizes, row counts)
	@echo "$(CYAN)Database Statistics:$(RESET)\n"
	@echo "$(YELLOW)Table Sizes:$(RESET)"
	docker exec $(DEV_DB_CONTAINER) psql -U $(DEV_DB_USER) -d $(DEV_DB_NAME) -c "SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size FROM pg_tables WHERE schemaname = 'public' ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC LIMIT 10;"
	@echo "\n$(YELLOW)Index Sizes:$(RESET)"
	docker exec $(DEV_DB_CONTAINER) psql -U $(DEV_DB_USER) -d $(DEV_DB_NAME) -c "SELECT indexname, pg_size_pretty(pg_relation_size(schemaname||'.'||indexname)) as size FROM pg_indexes WHERE schemaname = 'public' ORDER BY pg_relation_size(schemaname||'.'||indexname) DESC LIMIT 10;"
	@echo "\n$(YELLOW)Table Row Counts:$(RESET)"
	docker exec $(DEV_DB_CONTAINER) psql -U $(DEV_DB_USER) -d $(DEV_DB_NAME) -c "SELECT schemaname, relname as table, n_live_tup as rows FROM pg_stat_user_tables ORDER BY n_live_tup DESC LIMIT 10;"

db-connections: ## Show active database connections
	@echo "$(CYAN)Active Connections:$(RESET)\n"
	docker exec $(DEV_DB_CONTAINER) psql -U $(DEV_DB_USER) -d $(DEV_DB_NAME) -c "SELECT pid, usename, datname, state, query_start, EXTRACT('epoch' FROM now() - query_start) as duration_seconds FROM pg_stat_activity WHERE datname = current_database() ORDER BY query_start;"

# ==============================================================================
# DATABASE - PERFORMANCE & MONITORING
# ==============================================================================

.PHONY: db-slow-queries db-enable-stats db-index-usage db-cache-ratio

db-slow-queries: ## Show slow queries (>100ms)
	@echo "$(CYAN)Slow Queries (>100ms):$(RESET)\n"
	@echo "$(YELLOW)Requires pg_stat_statements extension. Run: make db-enable-stats$(RESET)"
	docker exec $(DEV_DB_CONTAINER) psql -U $(DEV_DB_USER) -d $(DEV_DB_NAME) -c "SELECT query, calls, round(mean_time::numeric, 2) as avg_ms, round(total_time::numeric, 2) as total_ms FROM pg_stat_statements WHERE mean_time > 100 ORDER BY mean_time DESC LIMIT 20;" 2>/dev/null || echo "pg_stat_statements not enabled. Run: make db-enable-stats"

db-enable-stats: ## Enable pg_stat_statements extension
	@echo "$(CYAN)Enabling pg_stat_statements extension...$(RESET)"
	docker exec $(DEV_DB_CONTAINER) psql -U $(DEV_DB_USER) -d $(DEV_DB_NAME) -c "CREATE EXTENSION IF NOT EXISTS pg_stat_statements;"
	@echo "$(GREEN)✓ pg_stat_statements enabled$(RESET)"
	@echo "$(YELLOW)Note: You may need to restart PostgreSQL for full functionality$(RESET)"

db-index-usage: ## Show index usage statistics
	@echo "$(CYAN)Index Usage Statistics:$(RESET)\n"
	docker exec $(DEV_DB_CONTAINER) psql -U $(DEV_DB_USER) -d $(DEV_DB_NAME) -c "SELECT schemaname, tablename, indexname, idx_scan as scans, pg_size_pretty(pg_relation_size(indexrelid)) as size, CASE WHEN idx_scan = 0 THEN '❌ UNUSED' WHEN idx_scan < 100 THEN '⚠️  LOW' ELSE '✓ ACTIVE' END as status FROM pg_stat_user_indexes ORDER BY idx_scan ASC;"

db-cache-ratio: ## Show cache hit ratio (should be >99%)
	@echo "$(CYAN)Cache Hit Ratio:$(RESET)\n"
	@echo "$(YELLOW)Table Cache:$(RESET)"
	docker exec $(DEV_DB_CONTAINER) psql -U $(DEV_DB_USER) -d $(DEV_DB_NAME) -c "SELECT round((sum(heap_blks_hit)::float / (sum(heap_blks_hit) + sum(heap_blks_read))) * 100, 2) as table_cache_ratio_pct FROM pg_statio_user_tables;"
	@echo "\n$(YELLOW)Index Cache:$(RESET)"
	docker exec $(DEV_DB_CONTAINER) psql -U $(DEV_DB_USER) -d $(DEV_DB_NAME) -c "SELECT round((sum(idx_blks_hit)::float / (sum(idx_blks_hit) + sum(idx_blks_read))) * 100, 2) as index_cache_ratio_pct FROM pg_statio_user_indexes;"

# ==============================================================================
# DATABASE - BACKUP & RESTORE
# ==============================================================================

.PHONY: db-backup db-restore db-export db-import

db-backup: ## Create database backup (in Docker volume)
	@echo "$(CYAN)Creating database backup...$(RESET)"
	@mkdir -p backups
	docker exec $(DEV_DB_CONTAINER) pg_dump -U $(DEV_DB_USER) $(DEV_DB_NAME) -Fc > backups/backup_$$(date +%Y%m%d_%H%M%S).dump
	@echo "$(GREEN)✓ Backup created in backups/$(RESET)"
	@ls -lh backups/*.dump | tail -1

db-restore: ## Restore database from backup (requires BACKUP_FILE env var)
	@echo "$(CYAN)Restoring database from backup...$(RESET)"
	@if [ -z "$(BACKUP_FILE)" ]; then \
		echo "$(YELLOW)Available backups:$(RESET)"; \
		ls -lh backups/*.dump 2>/dev/null || echo "No backups found"; \
		echo "\n$(YELLOW)Usage:$(RESET) make db-restore BACKUP_FILE=backups/backup_YYYYMMDD_HHMMSS.dump"; \
		exit 1; \
	fi
	cat $(BACKUP_FILE) | docker exec -i $(DEV_DB_CONTAINER) pg_restore -U $(DEV_DB_USER) -d $(DEV_DB_NAME) --clean --if-exists
	@echo "$(GREEN)✓ Database restored$(RESET)"

db-export: ## Export database to SQL file
	@echo "$(CYAN)Exporting database to SQL...$(RESET)"
	@mkdir -p exports
	docker exec $(DEV_DB_CONTAINER) pg_dump -U $(DEV_DB_USER) $(DEV_DB_NAME) > exports/export_$$(date +%Y%m%d_%H%M%S).sql
	@echo "$(GREEN)✓ Database exported to exports/$(RESET)"

db-import: ## Import database from SQL file (requires SQL_FILE env var)
	@echo "$(CYAN)Importing database from SQL...$(RESET)"
	@if [ -z "$(SQL_FILE)" ]; then \
		echo "$(YELLOW)Usage:$(RESET) make db-import SQL_FILE=exports/export_YYYYMMDD_HHMMSS.sql"; \
		exit 1; \
	fi
	cat $(SQL_FILE) | docker exec -i $(DEV_DB_CONTAINER) psql -U $(DEV_DB_USER) -d $(DEV_DB_NAME)
	@echo "$(GREEN)✓ Database imported$(RESET)"

# ==============================================================================
# DATABASE - ADVANCED OPERATIONS
# ==============================================================================

.PHONY: db-sql db-shell db-truncate db-drop-tables

db-sql: ## Execute SQL query (requires SQL env var)
	@echo "$(CYAN)Executing SQL...$(RESET)"
	@if [ -z "$(SQL)" ]; then \
		echo "$(YELLOW)Usage:$(RESET) make db-sql SQL='SELECT * FROM users LIMIT 5;'"; \
		exit 1; \
	fi
	docker exec $(DEV_DB_CONTAINER) psql -U $(DEV_DB_USER) -d $(DEV_DB_NAME) -c "$(SQL)"

db-shell: ## Get PostgreSQL shell in container
	@echo "$(CYAN)PostgreSQL shell in container...$(RESET)"
	docker exec -it $(DEV_DB_CONTAINER) bash

db-truncate: ## Truncate all tables (keep schema, remove data ⚠️ DESTRUCTIVE)
	@echo "$(RED)⚠️  This will DELETE ALL DATA from all tables!$(RESET)"
	@echo "$(YELLOW)Press Ctrl+C to cancel, Enter to continue...$(RESET)"
	@read confirm
	@echo "$(CYAN)Truncating all tables...$(RESET)"
	@docker exec $(DEV_DB_CONTAINER) psql -U $(DEV_DB_USER) -d $(DEV_DB_NAME) -c "DO \$\$ DECLARE r RECORD; BEGIN FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename != 'schema_migrations') LOOP EXECUTE 'TRUNCATE TABLE ' || quote_ident(r.tablename) || ' CASCADE'; END LOOP; END \$\$;"
	@echo "$(GREEN)✓ All tables truncated$(RESET)"

db-drop-tables: ## Drop all tables (completely remove ⚠️ DESTRUCTIVE)
	@echo "$(RED)⚠️  This will DROP ALL TABLES!$(RESET)"
	@echo "$(YELLOW)Press Ctrl+C to cancel, Enter to continue...$(RESET)"
	@read confirm
	@echo "$(CYAN)Dropping all tables...$(RESET)"
	@docker exec $(DEV_DB_CONTAINER) psql -U $(DEV_DB_USER) -d $(DEV_DB_NAME) -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@echo "$(GREEN)✓ All tables dropped$(RESET)"

# ==============================================================================
# MIGRATIONS
# ==============================================================================

.PHONY: migrate-up migrate-down migrate-status migrate-reset

migrate-up: ## Apply database migrations
	@echo "Running PostgreSQL migrations..."
	go run ./scripts/migrate -action=up

migrate-down: ## Rollback last migration
	@echo "Rolling back migration..."
	go run ./scripts/migrate -action=down

migrate-status: ## Check migration status
	go run ./scripts/migrate -action=status

migrate-reset: ## Reset database (down all + up all)
	@echo "$(RED)⚠️  This will reset ALL DATA!$(RESET)"
	@echo "$(YELLOW)Press Ctrl+C to cancel, Enter to continue...$(RESET)"
	@read confirm
	@docker exec $(DEV_DB_CONTAINER) psql -U $(DEV_DB_USER) -d $(DEV_DB_NAME) -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@$(MAKE) migrate-up
	@echo "$(GREEN)✓ Database reset complete$(RESET)"

# ==============================================================================
# SEED DATA
# ==============================================================================

.PHONY: seed seed-dev seed-test

seed: ## Seed database with initial data
	@echo "Seeding database..."
	go run ./scripts/seed

seed-dev: ## Seed development data
	@echo "Seeding development data..."
	go run ./scripts/seed -env=dev

seed-test: ## Seed test data
	@echo "Seeding test data..."
	go run ./scripts/seed -env=test

# ==============================================================================
# BENCHMARKS & PERFORMANCE
# ==============================================================================

.PHONY: bench bench-save analyze-performance

bench: ## Run performance benchmarks
	@echo "$(CYAN)Running PostgreSQL benchmarks...$(RESET)"
	./scripts/run-benchmarks.sh

bench-save: ## Save benchmark as new baseline
	@echo "$(CYAN)Saving benchmark as baseline...$(RESET)"
	./scripts/run-benchmarks.sh
	cp benchmark_results/*/benchmark_results.txt benchmark_results/baseline.txt
	@echo "$(GREEN)✓ Baseline saved$(RESET)"

analyze-performance: ## Analyze and optimize database performance
	@echo "$(CYAN)Analyzing database performance...$(RESET)"
	./scripts/optimize-indexes.sh --dry-run
	@echo "\n$(YELLOW)Review results in:$(RESET) index_analysis/*/recommendations.md"

# ==============================================================================
# DOCUMENTATION
# ==============================================================================

.PHONY: swagger swagger-serve

swagger: ## Generate OpenAPI/Swagger documentation
	swag init -g cmd/api/main.go -o docs/swagger

swagger-serve: swagger ## Generate and serve Swagger UI
	@echo "Swagger UI available at: http://localhost:8080/swagger/index.html"
	@echo "Or start services with: docker-compose up -d"
	@echo "Access Swagger UI at: http://localhost:8081"

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

.PHONY: dev setup fresh-start

dev: migrate-up seed run ## Quick start: migrate, seed, and run

setup: keys migrate-up seed ## Initial setup: keys, migrate, seed
	@echo "\n$(GREEN)Setup complete!$(RESET)"
	@echo "Run with: $(CYAN)make run$(RESET) or $(CYAN)make docker-dev$(RESET)"

fresh-start: ## Complete fresh start: drop volumes, rebuild, migrate, seed
	@echo "$(RED)⚠️  Complete fresh start - ALL DATA WILL BE LOST!$(RESET)"
	@echo "$(YELLOW)This will:$(RESET)"
	@echo "  1. Stop all containers and delete volumes"
	@echo "  2. Start fresh containers"
	@echo "  3. Apply all migrations"
	@echo "  4. Seed initial data"
	@echo "\n$(YELLOW)Press Ctrl+C to cancel, Enter to continue...$(RESET)"
	@read confirm
	@$(MAKE) docker-drop CONFIRM=yes
	@$(MAKE) docker-up
	@sleep 10
	@$(MAKE) migrate-up
	@$(MAKE) seed
	@echo "$(GREEN)✓ Fresh start complete!$(RESET)"

# ==============================================================================
# HEALTH & STATUS
# ==============================================================================

.PHONY: health status

health: ## Check application health endpoints
	@echo "$(CYAN)Checking health endpoints...$(RESET)"
	@echo "\n$(YELLOW)API Health:$(RESET)"
	@curl -s http://localhost:8080/health || echo "$(RED)API not responding$(RESET)"
	@echo "\n\n$(YELLOW)API System Info:$(RESET)"
	@curl -s http://localhost:8080/system/info || echo "$(RED)API not responding$(RESET)"

status: ## Show complete system status
	@echo "$(CYAN)=== System Status ===$(RESET)\n"
	@echo "$(YELLOW)Containers:$(RESET)"
	@docker ps --filter "name=skeleton" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || echo "$(RED)Docker not running$(RESET)"
	@echo "\n$(YELLOW)Migrations:$(RESET)"
	@$(MAKE) migrate-status 2>/dev/null || echo "Migrations not available"
	@echo "\n$(YELLOW)Recent Logs (last 10 lines):$(RESET)"
	@docker-compose logs --tail=10 2>/dev/null || echo "No logs available"
	@echo "\n$(YELLOW)Quick Commands:$(RESET)"
	@echo "  make health          - Check health endpoints"
	@echo "  make docker-status   - Container status"
	@echo "  make db-stats        - Database statistics"
	@echo "  make psql            - Connect to database"

# ==============================================================================
# CI/CD
# ==============================================================================

.PHONY: ci

ci: lint test ## Run CI checks (lint + test)
	@echo "$(GREEN)All CI checks passed$(RESET)"

# ==============================================================================
# UTILITIES
# ==============================================================================

.PHONY: fmt deps watch-logs

fmt: ## Format code
	go fmt ./...

deps: ## Show dependencies
	go mod graph

watch-logs: ## Watch all logs in real-time (alternative to docker-logs)
	@echo "$(CYAN)Watching logs (press Ctrl+C to stop)...$(RESET)"
	@docker-compose logs -f --tail=100

.DEFAULT_GOAL := help