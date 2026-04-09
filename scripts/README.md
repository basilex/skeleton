# Utility Scripts

This directory contains operational scripts for the Skeleton CRM monorepo.

## Overview

All scripts in this directory are operational/maintenance tools, not part of the core application code.

## Go Scripts

Go scripts in this directory have their own `go.mod` with a replace directive pointing to `../backend`. This allows them to import packages from the backend module:

```go
// scripts/go.mod
module github.com/basilex/scripts
go 1.25.0
require github.com/basilex/skeleton v0.0.0
replace github.com/basilex/skeleton => ../backend
```

### Seed (`seed/`)

Database seeding script for initial/test data.

```bash
# Run from project root
make db-seed

# Or directly
cd scripts/seed && go run main.go
```

**Environment Variables:**
- `DATABASE_URL` - PostgreSQL connection string
  - Default: `postgres://skeleton:skeleton@localhost:5432/skeleton?sslmode=disable`

### Migrate (`migrate/`)

Database migration runner (alternative to main backend migration tool).

```bash
# Run migrations
cd scripts/migrate && go run main.go -action=up

# Rollback last migration
cd scripts/migrate && go run main.go -action=down
```

## Shell Scripts

### deploy-staging.sh

Deployment script for staging environment. Configures environment and deploys backend/frontend.

### run-benchmarks.sh

Performance benchmarking script for API endpoints.

### optimize-indexes.sh

Database index optimization script.

### docker-clean-cache.sh

Docker cleanup utility script.

## Usage from Makefile

All scripts have corresponding Makefile targets:

```bash
make db-seed           # Seed database
make scripts-migrate   # Run migration script
make scripts-benchmark # Run benchmarks
```

See root `Makefile` for all available targets.

## Directory Structure

```
scripts/
├── go.mod              # Go module for scripts
├── go.sum              # Dependencies
├── README.md           # This file
├── seed/               # Database seeding
│   ├── main.go
│   └── seed            # Binary (gitignored)
├── migrate/            # Migration runner
│   └── main.go
├── deploy-staging.sh   # Staging deployment
├── run-benchmarks.sh   # Performance tests
├── optimize-indexes.sh # DB optimization
└── docker-clean-cache.sh
```