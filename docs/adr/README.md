# Architecture Decision Records (ADR)

This directory contains all architecture decisions for the Skeleton API project.

## Active Standards

### Core Architecture

| ADR | Title | Status | Description |
|-----|-------|--------|-------------|
| [ADR-001](ADR-001-hexagonal-architecture.md) | Hexagonal Architecture | ✅ ACCEPTED | DDD + Clean Architecture, bounded contexts |
| [ADR-003](ADR-003-event-bus.md) | Event Bus | ✅ ACCEPTED | In-memory event bus for domain events |
| [ADR-004](ADR-004-rbac-model.md) | RBAC Model | ✅ ACCEPTED | Role-Based Access Control with permissions |

### Database & Storage

| ADR | Title | Status | Description |
|-----|-------|--------|-------------|
**[ADR-016](ADR-016-database-stack.md)** | **Database Stack Standard** | ✅ **ACCEPTED** | **PostgreSQL 16 + pgx v5 + scany v2 + squirrel** |
| [ADR-006](ADR-006-uuid-v7.md) | UUID v7 Implementation | ✅ ACCEPTED | Time-ordered UUIDs, 56% storage savings |
| [ADR-007](ADR-007-cursor-pagination.md) | Cursor Pagination | ✅ ACCEPTED | Keyset pagination for large datasets |
| [ADR-012](ADR-012-files-storage.md) | Files Storage | ✅ ACCEPTED | Multi-provider file storage abstraction |

### API & Integration

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

```
PostgreSQL 16
    └── pgx/v5 (pure driver, zero reflection)
        └── pgxpool (connection pooling, MANDATORY)
            ├── scany v2 (struct scanning, eliminates boilerplate)
            ├── squirrel (type-safe dynamic queries)
            ├── pgx.NamedArgs (complex static queries)
            ├── pgx.Batch (parallel queries)
            └── pgx.CopyFrom (bulk inserts)
```

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

| ADR | Title | Status | Superseded By |
|-----|-------|--------|---------------|
| [ADR-002](archive/ADR-002-sqlite-wal.md) | SQLite WAL | ❌ OBSOLETE | ADR-016 (PostgreSQL Standard) |
| [ADR-005](archive/ADR-005-no-orm.md) | No ORM | ❌ SUPERSEDED | ADR-016 (Database Stack) |
| [ADR-016 (old)](archive/ADR-016-pgx-with-postgres.md) | pgx with PostgreSQL | ❌ SUPERSEDED | ADR-016 (Database Stack Standard) |
| [ADR-017](archive/ADR-017-scany-squirrel-standard.md) | scany + squirrel | ❌ SUPERSEDED | ADR-016 (Database Stack Standard) |

## Quick Reference

### When to Use Each Tool

| Scenario | Tool | Example |
|----------|------|---------|
| Single result | `pgxscan.Get()` | Find by ID |
| Multiple results | `pgxscan.Select()` | List all |
| Dynamic filters | `squirrel` | Search with optional filters |
| Pagination | `squirrel` | Cursor/keyset pagination |
| 5+ parameters | `pgx.NamedArgs` | Complex updates |
| Parallel queries | `pgx.Batch` | Independent updates |
| Bulk insert (1000+) | `pgx.CopyFrom` | Import CSV |
| Simple static query | Raw SQL + `pgxscan` | Get by ID |

### Code Reduction Benefits

- ✅ **30-47% less code** in repositories
- ✅ **Type-safe queries** (compile-time checks)
- ✅ **No parameter counting bugs** (`$1`, `$2`...)
- ✅ **Consistent error handling** with `pgxscan.NotFound()`

### Performance Characteristics

| Tool | Reflection | Overhead | Performance |
|------|------------|----------|-------------|
| pgx/v5 | ❌ None | Zero | ⚡⚡⚡ Fastest |
| scany v2 | ✅ Minimal (struct tags) | Negligible | ⚡⚡ Fast |
| squirrel | ❌ None | Build time | ⚡⚡ Fast |

## Contributing

When creating a new ADR:

1. Follow the numbering sequence
2. Use the template: `docs/adr/templates/ADR-template.md`
3. Include: Status, Context, Decision, Consequences
4. Update this README.md with the new ADR

## References

- ADR Template: `docs/adr/templates/ADR-template.md`
- PostgreSQL 16 Documentation: https://www.postgresql.org/docs/16/
- pgx v5: https://github.com/jackc/pgx
- scany v2: https://github.com/georgysavva/scany
- squirrel: https://github.com/Masterminds/squirrel