# ADR-016: Use PostgreSQL with Pure pgx Driver

## Status

**ACCEPTED** - This is the definitive standard for all database operations in this project.<br>
**EXTENDED by ADR-017** - Repository optimization with scany v2 + squirrel.

## Decision

We use **PostgreSQL 16** exclusively with the **pure pgx/pqxpool driver** for all database operations.**NO SQL DATABASE DRIVERS OR ORMS ALLOWED:**
- ❌ NO sqlx - reflection-based, performance overhead
- ❌ NO GORM - ORM, hides complexity, N+1 problems
- ❌ NO sqlc - code generation, poor dynamic query support
- ❌ NO SQLite - PostgreSQL only for all environments (dev, test, prod)
- ❌ NO other SQL databases - PostgreSQL is the standard

**ONLY ACCEPTED APPROACH:**
- ✅ PostgreSQL 16 - our only database
- ✅ pgx/v5 - pure driver, zero reflection
- ✅ pgxpool - connection pooling for all operations
- ✅ scany v2 + squirrel (ADR-017) - eliminates boilerplate while maintaining pgx control
- ✅ Raw SQL queries with type-safe builders (ADR-017)
- ✅ Native PostgreSQL features (JSONB, arrays, LISTEN/NOTIFY, materialized views)

## Rationale

### Why PostgreSQL Only?

1. **Single Source of Truth**: One database, one syntax, one set of features
2. **Production Reality**: PostgreSQL is our production database - dev/test should match
3. **Advanced Features**: JSONB, arrays, full-text search, materialized views unavailable in SQLite
4. **Performance Characteristics**: PostgreSQL behaves differently than SQLite - must test with real DB

### Why Pure pgx/pqxpool?

1. **Zero Reflection Overhead**: Direct scanning into variables, no runtime type discovery
2. **Maximum Performance**: Direct PostgreSQL wire protocol, no abstraction layers
3. **Full Control**: No hidden magic, predictable query execution
4. **Production Battle-Tested**: Used by major Go projects, stable and performant
5. **Connection Pooling**: Built-in pooling with pgxpool, automatic connection management
6. **Native Features**: Direct access to PostgreSQL-specific features without abstraction

## Performance Comparison

| Approach      | Reflection | ORM Overhead | Dynamic Queries | Control | Performance |
|---------------|------------|--------------|-----------------|---------|-------------|
| **pgx**       | ❌ None    | ❌ None      | ✅ Excellent    | ✅ Full | ⚡ Fastest  |
| sqlx          | ✅ Runtime | ❌ None      | ✅ Good         | ✅ Full | 🐌 Slow    |
| sqlc          | ❌ None    | ❌ None      | ❌ Poor         | ✅ Good | ⚡ Fast     |
| GORM          | ✅ Heavy   | ✅ Heavy     | ✅ Good         | ❌ Limited | 🐌🐛 Slow/Buggy |

## Implementation Standard

### Connection Pool (MANDATORY)

```go
import "github.com/jackc/pgx/v5/pgxpool"

pool, err := pgxpool.New(ctx, databaseURL)
// ALWAYS use pgxpool - never use raw pgx.Conn for application code
```

### Repository Pattern with scany v2 + squirrel (ADR-017)

**IMPORTANT**: ADR-017 extends this ADR with repository optimization using scany v2 + squirrel.

**Before ADR-017 (manual scanning):**
```go
// ❌ OLD APPROACH - See ADR-017 for new standard
func (r *UserRepository) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
    query := `SELECT id, email, password_hash, is_active, created_at, updated_at 
              FROM users WHERE id = $1`
    row := r.pool.QueryRow(ctx, query, id)
    
    // Manual variable declaration and scanning
    var userID, email, passwordHash string
    var isActive bool
    var createdAt, updatedAt time.Time
    
    err := row.Scan(&userID, &email, &passwordHash, &isActive, &createdAt, &updatedAt)
    // ... manual mapping to domain
}
```

**After ADR-017 (DTO + scany):**
```go
// ✅ NEW STANDARD - Use ADR-017 pattern for all repositories
type userDTO struct {
    ID           string    `db:"id"`
    Email        string    `db:"email"`
    PasswordHash string    `db:"password_hash"`
    IsActive     bool      `db:"is_active"`
    CreatedAt    time.Time `db:"created_at"`
    UpdatedAt    time.Time `db:"updated_at"`
}

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
```

**For full implementation details**, see **ADR-017: scany v2 + squirrel Repository Standard**.

## Consequences

### Positive
- Maximum performance (zero reflection overhead from pgx)
- Predictable behavior under load
- Direct PostgreSQL feature access
- Full control over query execution
- No ORM anti-patterns
- **Reduced boilerplate (ADR-017)** - 30-47% code reduction via scany v2

### Negative
- ~~More boilerplate code in repositories~~ **MITIGATED by ADR-017**
- ~~Manual scanning for each query~~ **ELIMINATED by ADR-017**

### Mitigation
- **ADR-017 provides DTO pattern + scany v2** - eliminates scan boilerplate
- **ADR-017 provides squirrel** - eliminates dynamic query boilerplate
- Comprehensive repository tests

## Enforcement

This ADR is **non-negotiable**. All database code MUST use pgx/pqxpool.

## See Also

- **ADR-017: scany v2 + squirrel Repository Standard** - Repository optimization with scany v2 + squirrel

## Date

2026-04-08
