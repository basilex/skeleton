.PHONY: build run test test-cover test-race test-p0 lint swagger migrate-up migrate-down seed keys clean tidy docker-build docker-up docker-down docker-logs docker-ps docker-dev docker-prod

VERSION_MAJOR ?= 0
VERSION_MINOR ?= 1
VERSION_PATCH ?= 0
VERSION_STAGE ?= dev
VERSION       = $(VERSION_MAJOR).$(VERSION_MINOR).$(VERSION_PATCH)-$(VERSION_STAGE)
COMMIT        ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_TIME     = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

build:
	go build \
	  -ldflags="-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)" \
	  -o bin/api ./cmd/api

run: build
	./bin/api

test:
	go test ./... -timeout 30s

test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

test-race:
	go test -race ./...

test-p0:
	go test ./internal/identity/domain/... ./pkg/eventbus/memory/... -v

lint:
	golangci-lint run ./...

swagger:
	swag init -g cmd/api/main.go -o docs/api --parseDependency --parseInternal

swagger-serve: swagger
	swag serve -d docs/api

migrate-up:
	go run ./scripts/migrate/ up

migrate-down:
	go run ./scripts/migrate/ down

seed:
	go run ./scripts/seed/

keys:
	mkdir -p keys
	openssl genrsa -out keys/private.pem 2048
	openssl rsa -in keys/private.pem -pubout -out keys/public.pem

clean:
	rm -rf bin/ coverage.out coverage.html docs/api/

tidy:
	go mod tidy

# Docker commands
docker-build:
	docker build \
	  --build-arg VERSION_MAJOR=$(VERSION_MAJOR) \
	  --build-arg VERSION_MINOR=$(VERSION_MINOR) \
	  --build-arg VERSION_PATCH=$(VERSION_PATCH) \
	  --build-arg VERSION_STAGE=$(VERSION_STAGE) \
	  --build-arg COMMIT=$(COMMIT) \
	  --build-arg BUILD_TIME=$(BUILD_TIME) \
	  -t skeleton-api:$(VERSION) \
	  .

docker-dev:
	docker-compose up --build

docker-prod:
	docker-compose -f docker-compose.prod.yml up --build -d

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-ps:
	docker-compose ps

docker-clean:
	docker-compose down -v --rmi local

# Docker with Redis (optional)
docker-dev-redis:
	docker-compose --profile redis up --build
