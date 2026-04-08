# Environment Configuration

## Quick Start

```bash
# Copy example config for development
cp configs/.env.example configs/.env.dev

# Start PostgreSQL + Redis via Docker
make docker-up

# Run migrations
make migrate-up

# Run application
make run
```

## Configuration Files

All configuration files are located in `configs/`:

| File | Purpose | Git Tracked |
|------|---------|-------------|
| `.env.dev` | Development environment | ❌ No |
| `.env.test` | Test environment (testcontainers) | ✅ Yes |
| `.env.prod` | Production environment | ❌ No |
| `.env.example` | Development template | ✅ Yes |
| `.env.prod.example` | Production template | ✅ Yes |

## Environment Loading Order

The application loads configuration based on `APP_ENV`:

```bash
# Default: APP_ENV=dev
go run ./cmd/api  # Loads configs/.env.dev

# Test environment
APP_ENV=test go test ./...
# Loads configs/.env.test

# Production environment
APP_ENV=prod ./bin/api
# Loads configs/.env.prod
```

## Development Setup

### PostgreSQL + Redis (Docker)

```bash
# Start all services
make docker-up

# Check status
make docker-status

# Run migrations
make migrate-up

# Seed test data
make seed

# Stop services
make docker-stop
```

### Local Development (without Docker)

```bash
# Start only PostgreSQL + Redis
make docker-up  # Services only, no API container

# Run API locally
make run  # Uses configs/.env.dev
```

### Environment Variables

Copy and customize for development:

```bash
cp configs/.env.example configs/.env.dev
```

Key settings for development:

```bash
# PostgreSQL (from docker-compose.yml)
DATABASE_URL=postgres://skeleton:password@localhost:5432/skeleton?sslmode=disable

# Redis (from docker-compose.yml)
REDIS_URL=redis://localhost:6379/0

# Mock tokens for easier testing
USE_MOCK_TOKENS=true

# Log level
LOG_LEVEL=debug
```

## Production Setup

### 1. Generate Secrets

```bash
# Generate JWT keys
make keys
```

### 2. Set Environment Variables

**CRITICAL**: Never commit secrets to repository!

```bash
# Set via environment variables
export DATABASE_URL=postgres://user:password@prod-host:5432/skeleton?sslmode=require
export REDIS_URL=redis://password@prod-host:6379/0
export ALLOWED_ORIGINS=https://app.yourdomain.com,https://api.yourdomain.com

# Or use Kubernetes secrets / AWS Secrets Manager / Vault
```

### 3. Production Config Template

```bash
# Copy production template
cp configs/.env.prod.example configs/.env.prod

# Edit with your settings
# IMPORTANT: Set these via environment variables:
# - DATABASE_URL
# - REDIS_URL
# - ALLOWED_ORIGINS
# - JWT keys path
```

### 4. Production Checklist

- ✅ Set `DATABASE_URL` from secrets
- ✅ Set `REDIS_URL` from secrets
- ✅ Set `ALLOWED_ORIGINS` (CORS)
- ✅ Generate JWT keys: `make keys`
- ✅ Set `USE_MOCK_TOKENS=false`
- ✅ Use HTTPS/TLS
- ✅ Configure firewall rules
- ✅ Set up monitoring and alerts
- ✅ Configure backup strategy

## Configuration Reference

### Database

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection URL | Required in production |
| `DATABASE_TYPE` | Database type | `postgres` |
| `DB_MAX_OPEN_CONNS` | Max open connections | 25 (dev), 50 (prod) |
| `DB_MAX_IDLE_CONNS` | Max idle connections | 5 (dev), 10 (prod) |
| `DB_CONN_MAX_LIFETIME` | Connection lifetime | 1h (dev), 10m (prod) |

### Authentication

| Variable | Description | Default |
|----------|-------------|---------|
| `JWT_PRIVATE_KEY_PATH` | Path to RSA private key | `./keys/private.pem` |
| `JWT_PUBLIC_KEY_PATH` | Path to RSA public key | `./keys/public.pem` |
| `JWT_ACCESS_TTL` | Access token TTL | 15m |
| `JWT_REFRESH_TTL` | Refresh token TTL | 168h (7 days) |
| `USE_MOCK_TOKENS` | Use mock tokens (dev only) | true (dev), false (prod) |

### Redis

| Variable | Description | Default |
|----------|-------------|---------|
| `REDIS_URL` | Redis connection URL | `redis://localhost:6379/0` |
| `CACHE_TYPE` | Cache type | `memory` (dev), `redis` (prod) |
| `CACHE_DEFAULT_TTL` | Cache TTL (seconds) | 300 |

### RateLimiting

| Variable | Description | Default |
|----------|-------------|---------|
| `RATE_LIMIT_ENABLED` | Enable rate limiting | true |
| `RATE_LIMIT_TYPE` | Rate limiter type | `token_bucket` (dev), `sliding_window` (prod) |
| `RATE_LIMIT_GLOBAL_LIMIT` | Global requests per window | 1000 |
| `RATE_LIMIT_AUTH_LIMIT` | Auth requests per window | 5 |

### Logging

| Variable | Description | Default |
|----------|-------------|---------|
| `LOG_LEVEL` | Log level | `debug` (dev), `info` (prod) |
| `LOG_FORMAT` | Log format | `text` (dev), `json` (prod) |

## Docker Compose Environments

The `docker-compose.yml` is designed for development:

```yaml
services:
  postgres:
    environment:
      POSTGRES_USER: skeleton
      POSTGRES_PASSWORD: password  # ⚠️ Change in production!
      POSTGRES_DB: skeleton
  
  redis:
    # No password for development
```

For production, use:
- Managed PostgreSQL (AWS RDS, Google Cloud SQL, etc.)
- Managed Redis (AWS ElastiCache, etc.)
- Secrets management (Kubernetes secrets, Vault, etc.)

## Testing

Tests use `testcontainers` to spin up PostgreSQL:

```bash
# Run all tests
make test

# Tests automatically:
# 1. Start PostgreSQL 16 container
# 2. Run migrations
# 3. Execute tests
# 4. Cleanup container
```

## Troubleshooting

### Port Already in Use

```bash
# Check what's using port 8080
lsof -ti:8080

# Stop Docker container if running
docker-compose stop app

# Or run locally
make run
```

### Database Connection Failed

```bash
# Check PostgreSQL is running
make docker-status

# Check connection
make psql

# Run migrations
make migrate-up
```

### Redis Connection Failed

```bash
# Check Redis is running
make docker-status

# Test connection
redis-cli -h localhost -p 6379 ping
```

## Security Notes

1. **Never commit `.env.dev` or `.env.prod`** - These are in `.gitignore`
2. **Use environment variables in production** - Set via Kubernetes/Docker/VM
3. **Rotate JWT keys regularly** - At least every 90 days
4. **Use SSL/TLS in production** - `sslmode=require` in DATABASE_URL
5. **Enable rate limiting in production** - Protect against abuse
6. **Review .env files before commits** - Ensure no secrets

## Architecture Decision Records

See [ADR-016: Database Stack](docs/adr/ADR-016-database-stack.md) for database technology decisions.

## See Also

- [DEVELOPMENT.md](docs/DEVELOPMENT.md) - Development workflow
- [DATABASE_MIGRATION_GUIDE.md](docs/DATABASE_MIGRATION_GUIDE.md) - PostgreSQL 16 features
- [Makefile Reference](MAKEFILE_REFERENCE.md) - All commands