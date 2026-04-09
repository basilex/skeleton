# Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) for the Skeleton API project.

## What is an ADR?

An ADR is a document that captures an important architectural decision made along with its context and consequences.

## Quick Navigation

### Core Architecture

| ADR | Title | Status | Description |
|-----|-------|--------|-------------|
| [ADR-001](ADR-001-hexagonal-architecture.md) | Hexagonal Architecture | ✅ ACCEPTED | Clean architecture separation |
| [ADR-003](ADR-003-event-bus.md) | Event Bus | ✅ ACCEPTED | Domain event system |
| [ADR-004](ADR-004-rbac-model.md) | RBAC Model | ✅ ACCEPTED | Role-based access control |
| [ADR-006](ADR-006-uuid-v7.md) | UUID v7 | ✅ ACCEPTED | Time-sortable identifiers |
| [ADR-007](ADR-007-cursor-pagination.md) | Cursor Pagination | ✅ ACCEPTED | Efficient cursor-based pagination |

### API Design

| ADR | Title | Status | Description |
|-----|-------|--------|-------------|
| [ADR-008](ADR-008-versioning.md) | API Versioning | ✅ ACCEPTED | URL path versioning strategy |
| [ADR-009](ADR-009-swagger-annotations.md) | Swagger Annotations | ✅ ACCEPTED | OpenAPI documentation standard |

### Domain Features

| ADR | Title | Status | Description |
|-----|-------|--------|-------------|
| [ADR-010](ADR-010-notifications.md) | Notifications | ✅ ACCEPTED | Multi-channel notification system |
| [ADR-011](ADR-011-tasks-jobs.md) | Tasks & Jobs | ✅ ACCEPTED | Background job processing |
| [ADR-013](ADR-013-language-policy.md) | Language Policy | ✅ ACCEPTED | Ukrainian/Russian in transit, English in docs |
| [ADR-014](ADR-014-caching.md) | Caching Strategy | ✅ ACCEPTED | Redis for distributed caching |
| [ADR-015](ADR-015-rate-limiting.md) | Rate Limiting | ✅ ACCEPTED | Token bucket algorithm |

## Database Stack Standard

**ADR-016** defines the mandatory database stack for all repository implementations:

**Database Stack Hierarchy:**
- PostgreSQL 16
  - pgx/v5 (pure driver, zero reflection)
    - pgxpool (connection pooling, MANDATORY)
      - scany v2 (struct scanning, eliminates boilerplate)
      - squirrel (type-safe dynamic queries)
      - pgx.NamedArgs (complex static queries)
      - pgx.Batch (parallel queries)
      - pgx.CopyFrom (bulk inserts)

### What We Use

| Component | Purpose | Status |
|-----------|---------|--------|
| **PostgreSQL 16** | Primary database (dev/test/prod) | ✅ MANDATORY |
| **pgx/v5 + pgxpool** | Pure driver + connection pool | ✅ MANDATORY |
| **scany v2** | Struct scanning | ✅ MANDATORY for repositories |
| **squirrel** | Type-safe query builder | ✅ MANDATORY for dynamic queries |

### What We Do NOT Use

| Component | Reason | Status |
|-----------|--------|--------|
| **SQLite** | PostgreSQL only for all environments | ❌ BANNED |
| **sqlx** | Reflection overhead, not pgx-native | ❌ BANNED |
| **GORM/ENT/ORM** | Hides complexity, N+1 problems | ❌ BANNED |
| **sqlc** | Code generation, poor dynamic queries | ❌ BANNED |

See [ADR-016](ADR-016-database-stack.md) for full implementation details.

## Bounded Contexts

The following architectural decisions define each bounded context in the system:

| ADR | Context | Description |
|-----|---------|-------------|
| [ADR-017](ADR-017-parties.md) | Parties | Customer, supplier, partner, employee management |
| [ADR-018](ADR-018-contracts.md) | Contracts | Contract lifecycle with DATERANGE |
| [ADR-019](ADR-019-accounting.md) | Accounting | Chart of accounts, double-entry transactions |
| [ADR-020](ADR-020-ordering.md) | Ordering | Order management with state machine |
| [ADR-021](ADR-021-catalog.md) | Catalog | Product catalog with LTREE categories |
| [ADR-022](ADR-022-inventory.md) | Inventory | Warehouse and stock management 🆕 |

Each context follows:
- **Domain-Driven Design** - Aggregates, value objects, domain events
- **Hexagonal Architecture** - Domain → Application → Infrastructure → HTTP
- **CQRS-lite** - Commands for writes, queries for reads
- **Event-Driven** - Cross-context communication via EventBus

## Repository Implementation Example

```go
import (
    sq "github.com/Masterminds/squirrel"
    "github.com/georgysavva/scany/v2/pgxscan"
    "github.com/jackc/pgx/v5/pgxpool"
)

// 1. DTO Pattern (MANDATORY)
type userDTO struct {
    ID           string    `db:"id"`
    Email        string    `db:"email"`
    PasswordHash string    `db:"password_hash"`
    IsActive     bool      `db:"is_active"`
    CreatedAt    time.Time `db:"created_at"`
    UpdatedAt    time.Time `db:"updated_at"`
}

// 2. Repository Struct (MANDATORY)
type UserRepository struct {
    pool *pgxpool.Pool           // ✅ ALWAYS pgxpool.Pool
    psql sq.StatementBuilderType // ✅ Squirrel with PostgreSQL placeholders
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
    return &UserRepository{
        pool: pool,
        psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
    }
}

// 3. Single Result (MANDATORY)
func (r *UserRepository) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
    var dto userDTO
    err := pgxscan.Get(ctx, r.pool, &dto,
        `SELECT id, email, password_hash, is_active, created_at, updated_at 
         FROM users WHERE id = $1`, id)
    if err != nil {
        if pgxscan.NotFound(err) {
            return nil, domain.ErrUserNotFound
        }
        return nil, fmt.Errorf("find user: %w", err)
    }
    return r.dtoToDomain(dto)
}

// 4. Dynamic Query (MANDATORY for filters)
func (r *UserRepository) FindAll(ctx context.Context, filter domain.UserFilter) ([]*domain.User, error) {
    q := r.psql.Select("id", "email", "password_hash", "is_active", "created_at", "updated_at").
        From("users")

    if filter.Search != "" {
        q = q.Where(sq.ILike{"email": "%" + filter.Search + "%"})
    }
    if filter.IsActive != nil {
        q = q.Where(sq.Eq{"is_active": *filter.IsActive})
    }

    q = q.OrderBy("id DESC").Limit(uint64(filter.Limit + 1))

    query, args, err := q.ToSql()
    if err != nil {
        return nil, fmt.Errorf("build query: %w", err)
    }

    var dtos []userDTO
    if err := pgxscan.Select(ctx, r.pool, &dtos, query, args...); err != nil {
        return nil, fmt.Errorf("select users: %w", err)
    }

    return r.dtosToDomains(dtos)
}
```

## Archived Decisions

The following ADRs have been superseded or are no longer relevant:

- ADR-002: SQLite WAL - SQLite support removed (PostgreSQL only)
- ADR-005: No ORM - Merged into ADR-016
- ADR-016 (old): pgx with PostgreSQL - Merged into ADR-016
- ADR-017 (old): Scany/Squirrel Standard - Merged into ADR-016

## Creating a New ADR

1. Copy the template below
2. Name it `ADR-NNN-short-title.md` where NNN is the next sequential number
3. Fill in all sections
4. Add to the appropriate table above

## Template

```markdown
# ADR-NNN: Title

## Status

[Proposed|Accepted|Deprecated|Superseded]

## Context

What is the issue we're addressing?

## Decision

What is the change we're proposing/have made?

## Consequences

What becomes easier or harder because of this change?

## Alternatives Considered

What other options were evaluated?

## Implementation

How will this be implemented?
```

## Related Documentation

- **[README.md](../README.md)** - Project overview
- **[IMPLEMENTATION_STATUS.md](../IMPLEMENTATION_STATUS.md)** - Current implementation status
- **[ARCHITECTURE.md](../architecture/ARCHITECTURE.md)** - Architecture overview