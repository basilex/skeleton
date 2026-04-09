# Docker Development Quick Start

## Problem Fixed

✅ **Go version updated** - From 1.24 to 1.25to support air@latest
✅ **Container checks added** - Commands will check if containers are running
✅ **Better error messages** - Clear guidance when containers are not ready

## Quick Start

### Start Development Environment

```bash
# Start Docker containers (PostgreSQL + Redis + App)
make docker-dev

# Wait for containers to be ready (healthcheck)
# PostgreSQL: localhost:5432
# Redis: localhost:6379
# API: http://localhost:8080
```

### Check Container Status

```bash
# Check if containers are running
make docker-check

# Output:
# skeleton-postgres-dev: Up 2 minutes (healthy)
# skeleton-redis-dev: Up 2 minutes (healthy)
# skeleton-api-dev: Up 2 minutes (healthy)
```

### Connect to Database

```bash
# Once containers are running, connect to PostgreSQL
make psql

# Inside psql:
skeleton=# \dt              # Show tables
skeleton=# SELECT * FROM users LIMIT 5;
skeleton=# \q              # Exit
```

### Database Commands

```bash
# Check container status first
make docker-check

# If containers not running:
make docker-dev

# Then use database commands:
make db-tables        # Show tables
make db-migrations    # Show migration history
make db-stats         # Database statistics
make db-connections   # Active connections
```

## Common Issues

### Issue: "Container not running"

```bash
# Error:
Container skeleton-postgres-dev is not running. Run: make docker-dev

# Solution:
make docker-dev
```

### Issue: "Go version mismatch"

```bash
# Error:
go: github.com/air-verse/air@latest requires go >= 1.25

# Solution: Already fixed! Dockerfile.dev now uses Go 1.25
# Just rebuild:
docker-compose down
docker-compose up --build
```

### Issue: "Connection refused"

```bash
# Check if healthcheck passed:
docker ps

# Should show "(healthy)" status for postgres container
# If not, wait or check logs:
docker logs skeleton-postgres-dev
```

## Docker Commands

```bash
# Start environment
make docker-dev          # Development (hot reload)
make docker-staging      # Staging
make docker-prod         # Production

# Check status
make docker-check        # Check running containers
make docker-ps           # List containers

# View logs
make docker-logs         # All containers
docker logs skeleton-postgres-dev  # Only PostgreSQL

# Stop environment
make docker-down         # Stop containers
make docker-clean        # Remove containers + volumes
```

## Database Commands

```bash
# Connection
make psql              # Development DB
make psql-staging      # Staging DB
make psql-prod         # Production DB

# Information
make db-tables         # List tables
make db-migrations     # Migration history
make db-stats          # Database stats
make db-connections    # Active connections

# Monitoring
make db-slow-queries   # Slow queries (>100ms)
make db-index-usage    # Index usage
make db-cache-ratio    # Cache hit ratio

# Actions
make db-enable-stats   # Enable pg_stat_statements
make db-sql SQL='...'  # Execute SQL
make db-backup         # Create backup
make db-restore        # Restore backup
```

## Development Workflow

### Typical Session

```bash
# 1. Start environment
make docker-dev

# 2. Wait for containers (watch for "healthy" status)
# Typically takes 10-30 seconds

# 3. Run migrations
make migrate-up

# 4. Seed data
make seed

# 5. Connect to database
make psql

# 6. Check data
skeleton=# SELECT COUNT(*) FROM users;
skeleton=# \q

# 7. Run tests
make test-unit          # Unit tests (no DB)
make test-integration   # Integration tests (needs DB)

# 8. Stop when done
make docker-down
```

### Reset Database

```bash
# Complete reset (migrations down + up + seed)
make db-reset

# Or manual:
make migrate-down
make migrate-up
make seed
```

### Check Database Health

```bash
# Quick health check
make db-stats

# Output:
# Table Sizes:
#  users    | 15 MB
#  files    | 128 MB
#
# Index Sizes:
#  idx_files_metadata | 12 MB
#
# Row Counts:
#  users  | 1,234
#  files  | 10,567
```

## Environment Files

```bash
configs/
├── .env.example     # Template (updated for PostgreSQL)
├── .env.dev         # Development
├── .env.test        # Testing
└── .env.prod        # Production

# Database URLs are configured per environment:
# DEV_DB_CONTAINER=skeleton-postgres-dev
# STAGING_DB_CONTAINER=skeleton-postgres-staging
# PROD_DB_CONTAINER=skeleton-postgres-prod
```

## Troubleshooting

### Logs

```bash
# All logs
make docker-logs

# Specific container
docker logs skeleton-api-dev
docker logs skeleton-postgres-dev
docker logs skeleton-redis-dev

# Follow logs in real-time
docker logs -f skeleton-api-dev
```

### Clean Start

```bash
# Stop and remove everything
make docker-clean

# Rebuild from scratch
make docker-dev
```

### Port Conflicts

```bash
# If ports are occupied:
# PostgreSQL: 5432
# Redis: 6379
# API: 8080

# Find what's using the port:
lsof -i :5432

# Kill process or change port in docker-compose.yml
```

## Architecture

```
┌─────────────────────────────────────────┐
│         Docker Development               │
├─────────────────────────────────────────┤
│                                          │
│  ┌────────────────────────────────────┐ │
│  │ skeleton-api-dev (Go 1.25)         │ │
│  │ - Hot reload with Air             │ │
│  │ - Port: 8080                       │ │
│  └────────────────────────────────────┘ │
│                                          │
│  ┌────────────────────────────────────┐ │
│  │ skeleton-postgres-dev (PostgreSQL 16)│
│  │ - UUID v7 support                  │ │
│  │ - JSONB columns                    │ │
│  │ - Port: 5432                       │ │
│  └────────────────────────────────────┘ │
│                                          │
│  ┌────────────────────────────────────┐ │
│  │ skeleton-redis-dev (Redis 7)     │ │
│  │ - Cache & Event Bus              │ │
│  │ - Port: 6379                      │ │
│  └────────────────────────────────────┘ │
│                                          │
└─────────────────────────────────────────┘
```

## Next Steps

1. ✅ Start environment: `make docker-dev`
2. ✅ Wait for containers: `make docker-check`
3. ✅ Connect to DB: `make psql`
4. ✅ Run migrations: `make migrate-up`
5. ✅ Seed data: `make seed`
6. ✅ Run app: `make run` (or use hot reload)