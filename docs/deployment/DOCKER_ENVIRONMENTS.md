# Docker Compose Environments

## Overview

Four environments available, all with PostgreSQL 16 + Redis:

| Environment | File | Purpose | Resources |
|-------------|------|---------|-----------|
| **Development** | `docker-compose.yml` | Local development with hot reload | 2 CPU, 512MB RAM |
| **Staging** | `docker-compose.staging.yml` | Pre-production testing | 2 CPU, 1GB RAM |
| **Production** | `docker-compose.prod.yml` | Production deployment | 4 CPU, 2GB RAM |
| **Test** | `docker-compose.test.yml` | CI/CD testing | Minimal, tmpfs storage |

## Common Features

All environments include:
- ✅ PostgreSQL 16 with health checks
- ✅ Redis 7 with persistence and memory limits
- ✅ Health checks for all services
- ✅ Dedicated networks
- ✅ Volume management

## Development Environment

**File:** `docker-compose.yml`

```bash
# Start all services
make docker-dev
# or
docker-compose up

# Services:
# - App: http://localhost:8080 (hot reload)
# - PostgreSQL: localhost:5432
# - Redis: localhost:6379
# - Adminer: http://localhost:8081 (DB management UI)
```

**Environment Variables:**
- `DATABASE_URL`: PostgreSQL connection string
- `REDIS_URL`: Redis connection string
- Database: `skeleton`, User: `skeleton`, Password: `skeleton_password`

## Staging Environment

**File:** `docker-compose.staging.yml`

```bash
# Start staging environment
make docker-staging
# or
docker-compose -f docker-compose.staging.yml up -d
```

**Characteristics:**
- Production-like configuration
- Relaxed resource limits
- Environment: `staging`
- Database: `skeleton_staging`

## Production Environment

**File:** `docker-compose.prod.yml`

```bash
# Start production environment
make docker-prod
# or
docker-compose -f docker-compose.prod.yml up -d

# Requires secrets:
mkdir -p secrets
echo "your_secure_db_password" > secrets/db_password.txt
echo "your_secure_redis_password" > secrets/redis_password.txt
```

**Production Features:**
- Docker Secrets for passwords
- Resource limits (CPU, Memory)
- Log rotation configured
- Optimized Redis settings (1GB maxmemory, LRU eviction)
- No exposed database ports (by default)
- Optional Redis Commander for management

**Environment Variables:**
```bash
export VERSION_MAJOR=0
export VERSION_MINOR=1
export VERSION_PATCH=0
export VERSION_STAGE=prod
export DATABASE_URL=postgres://user:password@host:5432/db
export REDIS_URL=redis://host:6379
export RATE_LIMIT_GLOBAL_LIMIT=1000
export CACHE_REDIS_PREFIX=skeleton:prod
```

## Test Environment

**File:** `docker-compose.test.yml`

```bash
# Start test environment (for CI/CD)
make docker-test
# or
docker-compose -f docker-compose.test.yml up -d
```

**Characteristics:**
- Minimal resources
- tmpfs for fast I/O (ephemeral storage)
- Ephemeral data (no persistence)
- Fast startup times
- Suitable for automated testing

## Makefile Commands

```bash
# Development
make docker-dev        # Start development with hot reload
make docker-up         # Start containers in background
make docker-down       # Stop containers
make docker-logs       # View logs
make docker-ps         # List containers

# Staging
make docker-staging    # Start staging environment

# Production
make docker-prod       # Start production environment
make docker-build      # Build production image

# Test
make docker-test        # Start test environment for CI/CD

# Database
make migrate-up        # Apply migrations
make migrate-down      # Rollback migration
make migrate-status    # Check migration status
make seed              # Seed database
make db-reset          # Full reset (down + up + seed)

# Cleanup
make docker-clean      # Remove containers, volumes, images
```

## Database Connections

### Development
```bash
# PostgreSQL
psql -h localhost -U skeleton -d skeleton
# Password: skeleton_password

# Redis
redis-cli -h localhost -p 6379
```

### Production
```bash
# Requires secrets
kubectl exec -it deployment/skeleton-api -- env

# PostgreSQL
psql $DATABASE_URL

# Redis
redis-cli -u $REDIS_URL
```

## Resource Limits

| Environment | App CPU | App RAM | DB CPU | DB RAM | Redis RAM |
|-------------|---------|---------|--------|--------|-----------|
| Development | 2 cores | 512MB  | 4 cores | 2GB   | 256MB     |
| Staging     | 2 cores | 1GB     | 4 cores | 2GB   | 512MB     |
| Production  | 2 cores | 512MB   | 4 cores | 2GB   | 1GB       |
| Test        | Shared  | 256MB   | Shared  | 1GB   | 256MB     |

## Volumes

### Development
- `postgres_dev_data`: PostgreSQL data (persistent)
- `redis_dev_data`: Redis data (persistent)

### Staging
- `postgres_staging_data`: PostgreSQL data
- `redis_staging_data`: Redis data

### Production
- `postgres_prod_data`: PostgreSQL data (persistent)
- `redis_prod_data`: Redis data (persistent)

### Test
- Uses tmpfs (no persistent volumes)

## Health Checks

All services include health checks:

```yaml
# PostgreSQL
test: ["CMD-SHELL", "pg_isready -U username"]
interval: 10s
timeout: 5s
retries: 5

# Redis
test: ["CMD", "redis-cli", "ping"]
interval: 10s
timeout: 3s
retries: 5

# Application
test: ["CMD", "wget", "--spider", "http://localhost:8080/health"]
interval: 30s
timeout: 3s
retries: 3
```

## Networking

All services run on dedicated network:
- Network: `skeleton_network`
- Driver: `bridge`

Services can communicate using service names:
- `postgres:5432`
- `redis:6379`
- `app:8080`

## Security Considerations

### Development
- No secrets management (plaintext passwords)
- All ports exposed to host
- No SSL/TLS

### Staging
- Environment variables for passwords
- All ports exposed to host
- No SSL/TLS

### Production
- Docker Secrets for passwords
- Minimal port exposure
- Resource limits enforced
- Log rotation configured
- SSL/TLS recommended (configure in reverse proxy)

## Troubleshooting

### View logs
```bash
docker-compose logs app      # Application logs
docker-compose logs postgres # Database logs
docker-compose logs redis    # Redis logs
```

### Check health
```bash
docker-compose ps
docker inspect --format='{{.State.Health.Status}}' skeleton-api-dev
```

### Reset everything
```bash
make docker-clean
make db-reset
```

### Connect to PostgreSQL
```bash
docker exec -it skeleton-postgres-dev psql -U skeleton -d skeleton
```

### Connect to Redis
```bash
docker exec -it skeleton-redis-dev redis-cli
```

## Notes

1. **Development** uses local Docker volumes for persistence
2. **Production** requires secrets in `./secrets/` directory
3. **Test** uses tmpfs for speed (data lost on restart)
4. All environments use PostgreSQL 16 and Redis 7
5. Health checks ensure service readiness before app starts
6. See Makefile for all available commands
