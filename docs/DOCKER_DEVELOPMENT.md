# Development Environment Setup

This guide explains how to run the Skeleton Business Engine in a development environment using Docker Compose.

## Prerequisites

- Docker Engine 24.0+
- Docker Compose 2.20+
- Make (optional, for convenience commands)

## Quick Start

```bash
# Start all services
make docker-up

# Or using docker-compose directly
docker-compose up -d

# Check service status
docker-compose ps

# View logs
docker-compose logs -f api
```

## Services

### API Server (`api`)
- **Port**: 8080
- **Health Check**: http://localhost:8080/health
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **Environment**: Development (hot reload enabled)

### PostgreSQL Database (`postgres`)
- **Port**: 5432
- **User**: skeleton
- **Password**: skeleton_dev_password
- **Database**: skeleton
- **Management UI**: http://localhost:5050 (pgAdmin)

### Redis Cache (`redis`)
- **Port**: 6379
- **Used for**: Session storage, event bus, caching

### Swagger UI (`swagger`)
- **Port**: 8081
- **URL**: http://localhost:8081
- **Spec**: mounted from `./docs/swagger.json`

### pgAdmin (`pgadmin`)
- **Port**: 5050
- **URL**: http://localhost:5050
- **Email**: admin@skeleton.local
- **Password**: admin

## Database Migrations

### Run Migrations

```bash
# Using Make
make migrate-up

# Or directly
docker-compose exec api sh -c "make migrate-up"

# Using migrate tool
migrate -path ./migrations -database "postgres://skeleton:skeleton_dev_password@localhost:5432/skeleton?sslmode=disable" up
```

### Create New Migration

```bash
make migrate-create NAME=add_new_table
```

### Rollback Migration

```bash
make migrate-down
```

## Development Workflow

### 1. Start Services

```bash
make docker-up
```

### 2. Check Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f api
docker-compose logs -f postgres
```

### 3. Run Tests

```bash
# Unit tests
make test-unit

# Integration tests (requires running services)
make test-integration

# All tests
make test
```

### 4. database Management

```bash
# Connect to PostgreSQL
docker-compose exec postgres psql -U skeleton -d skeleton

# Or use pgAdmin
open http://localhost:5050
```

### 5. Stop Services

```bash
# Stop all services
make docker-down

# Stop and remove volumes
make docker-clean
```

## Environment Variables

Copy `.env.example` to `.env` and customize:

```bash
cp .env.example .env
```

Key variables:

```env
APP_ENV=dev
APP_PORT=8080

DB_HOST=postgres
DB_PORT=5432
DB_USER=skeleton
DB_PASSWORD=skeleton_dev_password
DB_NAME=skeleton

REDIS_HOST=redis
REDIS_PORT=6379

LOG_LEVEL=debug
```

## Common Tasks

### Reset Database

```bash
make docker-down
docker volume rm skeleton_postgres_data
make docker-up
make migrate-up
```

### View Redis Data

```bash
docker-compose exec redis redis-cli
> KEYS *
> GET some_key
```

### Debug API

```bash
# Attach to running container
docker-compose exec api sh

# Run with delve debugger
go install github.com/go-delve/delve/cmd/dlv@latest
dlv debug ./cmd/api
```

### Generate Swagger Docs

```bash
make swagger-gen
```

## Troubleshooting

### Port Already in Use

```bash
# Check what's using port 8080
lsof -i :8080

# Kill process
kill -9 <PID>
```

### Database Connection Failed

```bash
# Check PostgreSQL logs
docker-compose logs postgres

# Verify PostgreSQL is healthy
docker-compose exec postgres pg_isready -U skeleton

# Connect manually
docker-compose exec postgres psql -U skeleton -d skeleton
```

### Redis Connection Failed

```bash
# Check Redis logs
docker-compose logs redis

# Test connection
docker-compose exec redis redis-cli ping
```

### Migration Errors

```bash
# Check migration version
docker-compose exec postgres psql -U skeleton -d skeleton -c "SELECT * FROM schema_migrations;"

# Force version (dangerous!)
migrate -path ./migrations -database "postgres://..." force <version>
```

## Production Considerations

For production deployment, see:
- `Dockerfile` (multi-stage, optimized)
- `.github/workflows/ci.yml` (CI/CD pipeline)
- `k8s/` (Kubernetes manifests)
- `docs/DEPLOYMENT_GUIDE.md`

## See Also

- [Architecture Documentation](docs/architecture/ARCHITECTURE.md)
- [API Documentation](docs/swagger)
- [Deployment Guide](docs/DEPLOYMENT_GUIDE.md)
- [Testing Strategy](docs/testing-strategy.md)