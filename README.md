# Skeleton

Go DDD Hexagonal architecture skeleton project — production-ready foundation for building Go services.

## Features

- **Hexagonal Architecture** — clean separation of domain, application, infrastructure
- **Bounded Contexts** — isolated modules communicating via events
- **RBAC** — role-based access control with wildcard permissions
- **JWT Auth** — RS256 tokens with access/refresh flow
- **SQLite WAL** — zero-config database with pure Go driver
- **Event Bus** — pluggable in-memory (dev) / Redis (prod)
- **RFC 7807 Errors** — standardized error responses

## Quick Start

```bash
make keys && make migrate-up && make seed && make run
```

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.24+ |
| HTTP | Gin |
| Database | SQLite WAL (modernc.org/sqlite) |
| Auth | JWT RS256 (golang-jwt) |
| Events | In-memory / Redis Pub/Sub |
| Testing | stdlib + testify |

## Project Structure

```
├── cmd/api/              # Application entry point
├── internal/
│   ├── status/           # Build info context
│   └── identity/         # Users, roles, auth context
│       ├── domain/       # Aggregates, value objects, events
│       ├── application/  # Command/query handlers
│       ├── infrastructure/ # DB, token implementations
│       └── ports/        # HTTP handlers, middleware
├── pkg/                  # Shared packages
│   ├── eventbus/         # Event bus interface + implementations
│   ├── database/         # SQLite setup
│   ├── httpserver/       # Server wrapper
│   ├── middleware/       # Global middleware
│   ├── apierror/         # RFC 7807 errors
│   └── config/           # Configuration loader
├── migrations/           # SQL migrations
├── docs/                 # Architecture, ADR, guides
└── scripts/              # Migrate, seed utilities
```

## Documentation

- [Architecture](docs/architecture/ARCHITECTURE.md)
- [Bounded Contexts](docs/architecture/BOUNDED_CONTEXTS.md)
- [Event Bus](docs/architecture/EVENT_BUS.md)
- [RBAC Model](docs/architecture/RBAC.md)
- [Getting Started](docs/development/GETTING_STARTED.md)
- [Testing](docs/development/TESTING.md)
- [Contributing](docs/development/CONTRIBUTING.md)

## ADRs

- [ADR-001: Hexagonal Architecture](docs/adr/ADR-001-hexagonal-architecture.md)
- [ADR-002: SQLite WAL](docs/adr/ADR-002-sqlite-wal.md)
- [ADR-003: Event Bus Strategy](docs/adr/ADR-003-event-bus.md)
- [ADR-004: RBAC Model](docs/adr/ADR-004-rbac-model.md)
- [ADR-005: No ORM](docs/adr/ADR-005-no-orm.md)

## Commands

```bash
make build          # Build binary
make run            # Build and run
make test           # Run all tests
make test-cover     # Run with coverage report
make test-race      # Run with race detector
make lint           # Run golangci-lint
make keys           # Generate RSA key pair
make migrate-up     # Apply migrations
make migrate-down   # Rollback migrations
make seed           # Seed dev data
make clean          # Clean build artifacts
```
