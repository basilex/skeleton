# Skeleton

Go DDD Hexagonal architecture skeleton project — production-ready foundation for building Go services.

## Features

- **Hexagonal Architecture** — clean separation of domain, application, infrastructure
- **Bounded Contexts** — isolated modules communicating via events
- **RBAC** — role-based access control with wildcard permissions
- **JWT Auth** — RS256 tokens with access/refresh flow
- **Session Management** — cookie-based sessions with Redis/in-memory store
- **Notifications** — multi-channel (Email, SMS, Push, In-App) with templates & preferences
- **Tasks/Jobs** — background job processing with retry & dead letter queue
- **Cursor Pagination** — UUID v7-based pagination for stable, performant lists
- **UUID v7** — time-ordered identifiers for optimal database indexing
- **SQLite WAL** — zero-config database with pure Go driver
- **Event Bus** — pluggable in-memory (dev) / Redis (prod)
- **RFC 7807 Errors** — standardized error responses
- **Swagger/OpenAPI** — auto-generated API documentation
- **Docker Ready** — multi-stage builds, docker-compose, hot reload
- **Production Ready** — health checks, graceful shutdown, security best practices

## Quick Start

```bash
# Local development
make setup && make dev

# Or Docker
make docker-dev
```

## Docker Support

The project supports Docker for development and production:

```bash
# Development with hot reload
make docker-dev

# Production
make docker-prod

# View logs
make docker-logs

# Stop containers
make docker-down
```

See [GETTING_STARTED.md](docs/development/GETTING_STARTED.md) for Docker details.

## Version Management

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.24+ |
| HTTP | Gin |
| Database | SQLite WAL (modernc.org/sqlite) |
| Auth | JWT RS256 (golang-jwt) |
| Sessions | Cookie-based (Redis / in-memory) |
| Events | In-memory / Redis Pub/Sub |
| Testing | stdlib + testify |
| Docs | Swagger/OpenAPI (swag) |

## Project Structure

```
├── cmd/api/              # Application entry point
│   ├── main.go           # Lifecycle orchestration
│   ├── wire.go           # Dependency injection
│   ├── routes.go         # Router and middleware
│   └── logger.go         # Logger setup
├── internal/
│   ├── status/           # Build info, health checks
│   ├── identity/         # Users, roles, auth, sessions
│   │   ├── domain/       # Aggregates, value objects, events
│   │   ├── application/  # Command/query handlers
│   │   ├── infrastructure/ # DB, token, session implementations
│   │   └── ports/        # HTTP handlers, middleware, DTOs
│   ├── audit/            # Audit log, system events
│   │   ├── domain/       # Record aggregate, events
│   │   ├── application/  # Log/Query handlers
│   │   ├── infrastructure/ # Persistence, event handlers
│   │   └── ports/        # HTTP handler
│   ├── notifications/    # Email, SMS, Push, In-App notifications
│   │   ├── domain/       # Notification, Template, Preferences aggregates
│   │   ├── application/  # Command/Query handlers, Event handlers
│   │   ├── infrastructure/ # Repositories, Senders, Worker
│   │   └── ports/        # HTTP handlers
│   └── tasks/            # Background jobs, scheduled tasks
│       ├── domain/       # Task, Schedule, DeadLetter aggregates
│       ├── application/  # Command/Query handlers
│       ├── infrastructure/ # Repositories, Worker
│       └── ports/        # HTTP handlers
├── pkg/                  # Shared packages
│   ├── eventbus/         # Event bus interface + implementations
│   ├── database/         # SQLite setup
│   ├── httpserver/       # Server wrapper
│   ├── middleware/       # Global middleware
│   ├── apierror/         # RFC 7807 errors
│   ├── config/           # Configuration loader
│   ├── pagination/       # Cursor-based pagination
│   └── uuid/             # UUID v7 implementation
├── migrations/           # SQL migrations
├── docs/                 # Architecture, ADR, guides, API
├── configs/              # Environment-specific .env files
└── scripts/              # Migrate, seed utilities
```

## Documentation

- [Architecture](docs/architecture/ARCHITECTURE.md)
- [Bounded Contexts](docs/architecture/BOUNDED_CONTEXTS.md)
- [Event Bus](docs/architecture/EVENT_BUS.md)
- [RBAC Model](docs/architecture/RBAC.md)
- [Getting Started](docs/development/GETTING_STARTED.md)
- [Notifications Guide](docs/development/NOTIFICATIONS.md)
- [Testing](docs/development/TESTING.md)
- [Contributing](docs/development/CONTRIBUTING.md)

## ADRs

- [ADR-001: Hexagonal Architecture](docs/adr/ADR-001-hexagonal-architecture.md)
- [ADR-002: SQLite WAL](docs/adr/ADR-002-sqlite-wal.md)
- [ADR-003: Event Bus Strategy](docs/adr/ADR-003-event-bus.md)
- [ADR-004: RBAC Model](docs/adr/ADR-004-rbac-model.md)
- [ADR-005: No ORM](docs/adr/ADR-005-no-orm.md)
- [ADR-006: UUID v7](docs/adr/ADR-006-uuid-v7.md)
- [ADR-007: Cursor Pagination](docs/adr/ADR-007-cursor-pagination.md)
- [ADR-008: Semantic Versioning Strategy](docs/adr/ADR-008-versioning.md)
- [ADR-009: Mandatory Swagger Annotations](docs/adr/ADR-009-swagger-annotations.md)
- [ADR-010: Notifications Context](docs/adr/ADR-010-notifications.md)
- [ADR-011: Tasks/Jobs Context](docs/adr/ADR-011-tasks-jobs.md)
- [ADR-012: Files/Storage Context](docs/adr/ADR-012-files-storage.md)

## Version Management

The project uses **Semantic Versioning** with environment support:

```bash
# Development build (default: 0.1.0-dev)
make build

# Staging build
VERSION_STAGE=staging make build  # 0.1.0-staging

# Production build
VERSION_STAGE=prod make build      # 0.1.0-prod

# Custom version
VERSION_MAJOR=0 VERSION_MINOR=2 VERSION_PATCH=0 make build  # 0.2.0-dev
```

The `/build` endpoint shows version, `commit` - git hash for reference:

```bash
curl http://localhost:8080/build
# {
#   "version": "0.1.0-dev",
#   "commit": "c4410c8",
#   "build_time": "2026-04-07T10:00:37Z",
#   "go_version": "go1.26.1",
#   "env": "dev"
# }
```

## Commands

```bash
make build          # Build binary
make run            # Build and run
make test           # Run all tests
make test-cover     # Run with coverage report
make test-race      # Run with race detector
make test-p0        # Run critical domain tests only
make lint           # Run golangci-lint
make swagger        # Generate OpenAPI docs
make swagger-serve  # Generate and serve Swagger UI
make keys           # Generate RSA key pair
make migrate-up     # Apply migrations
make migrate-down   # Rollback migrations
make seed           # Seed dev data
make clean          # Clean build artifacts

# Docker
make docker-build   # Build production image
make docker-dev     # Start development (hot reload)
make docker-prod    # Start production containers
make docker-up      # Start in background
make docker-down    # Stop containers
make docker-logs    # View logs
make docker-ps      # List containers
```
