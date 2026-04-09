# Skeleton API

Production-ready Go API skeleton with Domain-Driven Design (DDD) and Hexagonal Architecture.

## Quick Start

```bash
# One-command setup (⚠️ will delete all data)
make fresh-start

# After completion:
# - API: http://localhost:8080
# - Test credentials: admin@skeleton.local / Admin1234!
```

## Overview

| Component | Technology |
|------------|------------|
| Language | Go 1.25+ |
| Database | PostgreSQL 16 |
| Cache | Redis 7 |
| HTTP | Gin |

### Architecture

**DDD + Hexagonal**: Domain → Application → Infrastructure → Ports

### Bounded Contexts

| Context | Description |
|---------|-------------|
| Identity | Auth, users, sessions, RBAC |
| Audit | Action tracking, compliance |
| Notifications | Email, SMS, Push, In-App |
| Files | Upload, download, processing |
| Parties | Customers, suppliers, partners |
| Contracts | Contract lifecycle management |
| Accounting | Chart of accounts, transactions |
| Ordering | Orders, quotes |
| Catalog | Product catalog |
| Invoicing | Invoices, payments |
| Documents | Document management |
| Inventory | Warehouse, stock management |

## Documentation

### Getting Started

| Document | Description |
|----------|-------------|
| [QUICKSTART.md](docs/QUICKSTART.md) | Quick start guide |
| [DOCKER_QUICKSTART.md](docs/DOCKER_QUICKSTART.md) | Docker setup guide |
| [DEVELOPMENT.md](docs/DEVELOPMENT.md) | Development workflow |
| [ENVIRONMENT_SETUP.md](docs/ENVIRONMENT_SETUP.md) | Environment configuration |

### Architecture & Design

| Document | Description |
|----------|-------------|
| [CROSS_CONTEXT_INTEGRATION.md](docs/CROSS_CONTEXT_INTEGRATION.md) | Context integration patterns |
| [CRM_INTEGRATION_GAPS.md](docs/CRM_INTEGRATION_GAPS.md) | Integration analysis |
| [ADR/](docs/adr/) | Architecture Decision Records |

### Operations

| Document | Description |
|----------|-------------|
| [DATABASE_MIGRATION_GUIDE.md](docs/DATABASE_MIGRATION_GUIDE.md) | Migration guide |
| [DEPLOYMENT_GUIDE.md](docs/DEPLOYMENT_GUIDE.md) | Production deployment |
| [MAKEFILE_REFERENCE.md](docs/MAKEFILE_REFERENCE.md) | Makefile command reference |

### Planning & Progress

| Document | Description |
|----------|-------------|
| [EXECUTION_PLAN.md](docs/EXECUTION_PLAN.md) | Development execution plan |
| [CONTEXTS_HARDENING.md](docs/CONTEXTS_HARDENING.md) | Context hardening progress |
| [CHANGELOG.md](docs/CHANGELOG.md) | Change history |

## Common Commands

```bash
# Docker
make docker-up        # Start containers
make docker-down      # Stop containers
make docker-logs      # View logs

# Database
make psql             # Interactive shell
make migrate-up       # Apply migrations
make db-stats         # Database statistics

# Development
make dev              # Quick start
make test             # Run all tests
make lint             # Run linter
```

## Tech Stack

| Layer | Technology | Purpose |
|-------|------------|---------|
| Domain | Pure Go | Business logic |
| Application | Go + CQRS-lite | Use cases |
| Infrastructure | PostgreSQL 16 + Redis 7 | Persistence, cache |
| Ports | Gin | HTTP API |

## Database

- **UUID v7** - Time-sortable primary keys
- **JSONB** - Queryable metadata
- **40+ indexes** - Optimized for production

## Testing

```bash
make test-unit          # Unit tests
make test-integration   # Integration tests
make bench              # Benchmarks
```

## License

MIT License - see [LICENSE](LICENSE) for details.