# Quick Start Guide

Get the Skeleton API running in under 5 minutes.

## Prerequisites

- **Docker & Docker Compose** - For PostgreSQL + Redis
- **Make** - For convenient commands  
- **Go 1.25+** - Required for hot reload

## One-Command Start

```bash
# Clone and start (⚠️ will delete all existing data)
make fresh-start
```

This command will:
1. Stop and remove all containers + volumes
2. Start PostgreSQL 16 + Redis 7 + API containers
3. Apply all 17 migrations (including UUID v7 functions)
4. Seed initial data (admin user, roles, permissions)
5. Start API server with hot reload

## Verify Installation

```bash
# Check containers are running
make docker-status

# Check health endpoint
make health

# Connect to database
make psql
```

Expected output:
```
NAMES                   STATUS
skeleton-api-dev        Up 2 minutes (healthy)
skeleton-postgres-dev   Up 2 minutes (healthy)
skeleton-redis-dev      Up 2 minutes (healthy)
skeleton-adminer-dev    Up 2 minutes
```

## Test Credentials

After `make fresh-start`:

```bash
# Admin user
Email: admin@skeleton.local
Password: Admin1234!

# Role: super_admin
# Permissions: All including wildcard (*:*)
```

## Database Access

```bash
# Interactive psql
make psql

# List tables
make db-tables

# Query JSONB metadata
make db-sql SQL="SELECT * FROM files WHERE metadata @> '{\"width\":1920}';"

# Check UUID v7 generation
make db-sql SQL="SELECT uuid_generate_v7(), uuid_v7_to_timestamp(uuid_generate_v7());"
```

## Common Commands

```bash
# Development workflow
make docker-up          # Start containers
make migrate-status    # Check migrations
make seed              # Add test data
make docker-logs-app    # View API logs

# Database management
make psql              # Interactive shell
make db-tables         # List tables
make db-stats          # Database statistics
make db-backup         # Create backup

# Troubleshooting
make status            # System status
make docker-status     # Container status
make health            # Health endpoints
```

## Stop Development

```bash
# Stop containers (keeps data)
make docker-down

# Complete reset (deletes all data)
make docker-drop
```

## What's Next?

- **[DEVELOPMENT.md](docs/DEVELOPMENT.md)** - Complete development guide
- **[DATABASE_MIGRATION_GUIDE.md](docs/DATABASE_MIGRATION_GUIDE.md)** - Migration details
- **[MAKEFILE_REFERENCE.md](MAKEFILE_REFERENCE.md)** - All commands
- **[Architecture](README.md#architecture-overview)** - Project structure

## Architecture Highlights

- **Domain-Driven Design** with Hexagonal Architecture
- **PostgreSQL 16** with native UUID v7, JSONB, generated columns
- **UUID v7** for all primary keys (time-sortable, 56% smaller than TEXT)
- **JSONB** for metadata with GIN indexes (10-100x faster queries)
- **Pure pgx/pqxpool** - Zero reflection, maximum performance
- **Redis 7** for caching, rate limiting, event bus

## Database Schema

All tables use:
- `UUID PRIMARY KEY DEFAULT uuid_generate_v7()` - Time-sortable IDs
- `TIMESTAMPTZ NOT NULL DEFAULT NOW()` - Proper timestamps
- `JSONB DEFAULT '{}'` for flexible metadata
- GIN indexes for JSONB columns

## Getting Help

```bash
make help              # List all commands
make status            # Check system status
cat README.md          # Main documentation
cat docs/DEVELOPMENT.md # Development guide
```

---

**Ready to develop?** Continue with [DEVELOPMENT.md](docs/DEVELOPMENT.md) →
