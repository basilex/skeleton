.PHONY: build run test test-cover test-race test-p0 lint swagger migrate-up migrate-down seed keys clean tidy

VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_TIME  = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

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
