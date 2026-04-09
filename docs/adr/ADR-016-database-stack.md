# ADR: Database Stack Standard

## Status

**ACCEPTED** - Mandatory for all database operations.

## Decision

**PostgreSQL 16 + pgx v5 + scany v2 + squirrel** is the definitive database stack.

**Database Stack Hierarchy:**
- PostgreSQL 16
  - pgx/v5 (pure driver, zero reflection)
    - pgxpool (connection pooling, MANDATORY)
      - scany v2 (struct scanning, eliminates boilerplate)
      - squirrel (type-safe dynamic queries)
      - pgx.NamedArgs (complex static queries, 5+ parameters)
      - pgx.Batch (parallel independent queries)
      - pgx.CopyFrom (bulk inserts, 1000+ rows)

## What We Use

| Component | Purpose | Status |
|-----------|---------|--------|
| **PostgreSQL 16** | Primary database (dev/test/prod) | ✅ MANDATORY |
| **pgx/v5 + pgxpool** | Pure driver + connection pool | ✅ MANDATORY |
| **scany v2** | Struct scanning | ✅ MANDATORY for repositories |
| **squirrel** | Type-safe query builder | ✅ MANDATORY for dynamic queries |
| **pgx.NamedArgs** | Named parameters | ✅ RECOMMENDED for complex queries |
| **pgx.Batch** | Batch operations | ✅ RECOMMENDED for parallel queries |
| **pgx.CopyFrom** | Bulk inserts | ✅ RECOMMENDED for 1000+ rows |

## What We Do NOT Use

| Component | Reason | Status |
|-----------|--------|--------|
| **SQLite** | PostgreSQL only for all environments | ❌ BANNED |
| **sqlx** | Reflection overhead, not pgx-native | ❌ BANNED |
| **GORM/ENT/ORM** | Hides complexity, N+1 problems | ❌ BANNED |
| **sqlc** | Code generation, poor dynamic queries | ❌ BANNED |
| **TEXT IDs** | UUID v7 only (ADR-006) | ❌ BANNED |
| **TEXT metadata** | JSONB only | ❌ BANNED |

## Repository Implementation Standard

### 1. DTO Pattern (MANDATORY)

Always create DTO structs for scanning results:

```go
type userDTO struct {
    ID           string    `db:"id"`
    Email        string    `db:"email"`
    PasswordHash string    `db:"password_hash"`
    IsActive     bool      `db:"is_active"`
    CreatedAt    time.Time `db:"created_at"`
    UpdatedAt    time.Time `db:"updated_at"`
}

// Mapping method to domain entity
func (r *UserRepository) dtoToDomain(dto userDTO) (*domain.User, error) {
    userID, err := domain.ParseUserID(dto.ID)
    if err != nil {
        return nil, fmt.Errorf("parse user id: %w", err)
    }
    email, err := domain.NewEmail(dto.Email)
    if err != nil {
        return nil, fmt.Errorf("parse email: %w", err)
    }
    return domain.ReconstituteUser(userID, email, domain.PasswordHash(dto.PasswordHash),
        []domain.RoleID{}, dto.IsActive, dto.CreatedAt, dto.UpdatedAt)
}

func (r *UserRepository) dtosToDomains(dtos []userDTO) ([]*domain.User, error) {
    users := make([]*domain.User, 0, len(dtos))
    for _, dto := range dtos {
        user, err := r.dtoToDomain(dto)
        if err != nil {
            return nil, err
        }
        users = append(users, user)
    }
    return users, nil
}
```

### 2. Repository Struct (MANDATORY)

```go
import (
    sq "github.com/Masterminds/squirrel"
    "github.com/georgysavva/scany/v2/pgxscan"
    "github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
    pool *pgxpool.Pool           // ✅ ALWAYS pgxpool.Pool
    psql sq.StatementBuilderType // ✅ Squirrel with PostgreSQL placeholders
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
    return &UserRepository{
        pool: pool,
        psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar), // PostgreSQL $1, $2...
    }
}
```

### 3. Single Result Queries (MANDATORY)

Use `pgxscan.Get()` for single results, handle `NotFound`:

```go
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

### 4. Multiple Results Queries (MANDATORY)

Use `pgxscan.Select()` for multiple results:

```go
func (r *TaskRepository) GetPendingTasks(ctx context.Context, limit int) ([]*domain.Task, error) {
    var dtos []taskDTO
    err := pgxscan.Select(ctx, r.pool, &dtos,
        `SELECT id, type, status, priority, payload, result, error_code, error_message, error_details,
                attempts, max_attempts, scheduled_at, started_at, completed_at, created_at, updated_at
         FROM tasks
         WHERE status = $1 AND scheduled_at <= $2
         ORDER BY priority DESC, created_at ASC
         LIMIT $3`,
        domain.TaskStatusPending, time.Now(), limit)
    if err != nil {
        return nil, fmt.Errorf("query tasks: %w", err)
    }
    return r.dtosToDomains(dtos)
}
```

### 5. Dynamic Queries (MANDATORY)

Use `squirrel` for queries with conditional filters, pagination, sorting:

```go
func (r *FileRepository) List(ctx context.Context, filter *domain.FileFilter, limit, offset int) ([]*domain.File, error) {
    q := r.psql.Select("id", "owner_id", "filename", "stored_name", "mime_type", "size", "path",
        "storage_provider", "checksum", "metadata", "access_level",
        "uploaded_at", "expires_at", "processed_at", "created_at", "updated_at").
        From("files")

    // Type-safe conditional filters
    if filter != nil {
        if filter.OwnerID != nil {
            q = q.Where(sq.Eq{"owner_id": *filter.OwnerID})
        }
        if filter.MimeType != nil {
            q = q.Where(sq.ILike{"mime_type": *filter.MimeType + "%"})
        }
        if filter.AccessLevel != nil {
            q = q.Where(sq.Eq{"access_level": string(*filter.AccessLevel)})
        }
    }

    q = q.OrderBy("created_at DESC").Limit(uint64(limit)).Offset(uint64(offset))

    query, args, err := q.ToSql()
    if err != nil {
        return nil, fmt.Errorf("build query: %w", err)
    }

    var dtos []fileDTO
    if err := pgxscan.Select(ctx, r.pool, &dtos, query, args...); err != nil {
        return nil, fmt.Errorf("query files: %w", err)
    }

    return r.dtosToDomains(dtos)
}
```

**DO NOT USE squirrel for simple static queries:**

```go
// ✅ Good - simple static query
err := pgxscan.Get(ctx, pool, &dto, `SELECT * FROM users WHERE id = $1`, id)

// ❌ Bad - unnecessary overhead for static query
q := psql.Select("*").From("users").Where(sq.Eq{"id": id})
query, args, _ := q.ToSql()
pgxscan.Get(ctx, pool, &dto, query, args...)
```

### 6. Complex Static Queries (RECOMMENDED)

Use `pgx.NamedArgs` for queries with 5+ parameters:

```go
result, err := r.pool.Exec(ctx,
    `UPDATE files SET
        filename = @filename,
        stored_name = @stored_name,
        mime_type = @mime_type,
        size = @size,
        checksum = @checksum,
        updated_at = @updated_at
    WHERE id = @id`,
    pgx.NamedArgs{
        "filename":    file.Filename(),
        "stored_name": file.StoredName(),
        "mime_type":   file.MimeType(),
        "size":        file.Size(),
        "checksum":    file.Checksum(),
        "updated_at":  time.Now(),
        "id":          file.ID(),
    })
```

### 7. Batch Operations (RECOMMENDED)

Use `pgx.Batch` for parallel independent queries:

```go
batch := &pgx.Batch{}
batch.Queue("UPDATE users SET is_active = $1 WHERE id = $2", true, userID1)
batch.Queue("UPDATE users SET is_active = $1 WHERE id = $2", false, userID2)
batch.Queue("INSERT INTO audit_records (action, user_id) VALUES ($1, $2)", "bulk_update", adminID)

results := pool.SendBatch(ctx, batch)
defer results.Close()

// Process results
for i := 0; i < batch.Len(); i++ {
    _, err := results.Exec()
    // Handle error
}
```

### 8. Bulk Inserts (RECOMMENDED)

Use `pgx.CopyFrom` for inserting 1000+ rows:

```go
rows := pgx.CopyFromRows([][]any{
    {uuid1, "Alice", tenantID1},
    {uuid2, "Bob", tenantID2},
    // ... 10,000+ rows
})

count, err := pool.CopyFrom(ctx,
    pgx.Identifier{"users"},
    []string{"id", "name", "tenant_id"},
    rows)
```

## When to Use Each Tool

| Scenario | Tool | Example |
|----------|------|---------|
| **Single result** | `pgxscan.Get()` | Find by ID |
| **Multiple results** | `pgxscan.Select()` | List all, filtered queries |
| **Dynamic filters** | `squirrel` | Search with optional filters |
| **Pagination** | `squirrel` | Cursor/keyset pagination |
| **5+ parameters** | `pgx.NamedArgs` | Complex updates |
| **Parallel queries** | `pgx.Batch` | Independent updates |
| **Bulk insert** | `pgx.CopyFrom` | Import CSV, seed data |
| **Simple static query** | Raw SQL + `pgxscan` | Get by ID |

## Performance Characteristics

| Component | Reflection | Overhead | Performance |
|-----------|------------|----------|-------------|
| **pgx/v5** | ❌ None | Zero | ⚡⚡⚡ Fastest |
| **scany v2** | ✅ Minimal (struct tags) | Negligible | ⚡⚡ Fast |
| **squirrel** | ❌ None | Build time | ⚡⚡ Fast |
| **pgx.NamedArgs** | ❌ None | None | ⚡⚡⚡ Fastest |
| **pgx.Batch** | ❌ None | None | ⚡⚡⚡ Fastest |
| **pgx.CopyFrom** | ❌ None | None | ⚡⚡⚡ Fastest |

## Code Reduction

### Before (ManualScanning)

```go
func (r *UserRepository) FindAll(ctx context.Context, filter domain.UserFilter) ([]*domain.User, error) {
    query := `SELECT id, email FROM users WHERE 1=1`
    args := make([]interface{}, 0)
    argNum := 1
    
    if filter.Search != "" {
        query += fmt.Sprintf(" AND email LIKE $%d", argNum)  // Error-prone
        args = append(args, "%"+filter.Search+"%")
        argNum++
    }
    
    rows, err := r.pool.Query(ctx, query, args...)
    if err != nil { return nil, err }
    defer rows.Close()
    
    users := make([]*domain.User, 0)
    for rows.Next() {
        var id, email string  // Boilerplate
        if err := rows.Scan(&id, &email); err != nil {  // Boilerplate
            return nil, err
        }
        // Manual mapping
    }
    return users, nil
}
```

### After (scany + squirrel)

```go
func (r *UserRepository) FindAll(ctx context.Context, filter domain.UserFilter) ([]*domain.User, error) {
    q := r.psql.Select("id", "email", "password_hash", "is_active", "created_at", "updated_at").
        From("users")

    if filter.Search != "" {
        q = q.Where(sq.ILike{"email": "%" + filter.Search + "%"})  // Type-safe
    }

    query, args, err := q.ToSql()
    if err != nil { return nil, fmt.Errorf("build query: %w", err) }

    var dtos []userDTO
    if err := pgxscan.Select(ctx, r.pool, &dtos, query, args...); err != nil {
        return nil, fmt.Errorf("select users: %w", err)
    }

    return r.dtosToDomains(dtos)  // Clean mapping
}
```

**Result:**
- ✅ 30-47% less code
- ✅ Type-safe queries (compile-time checks)
- ✅ No parameter counting bugs
- ✅ Consistent error handling

## Native PostgreSQL Features

Use native PostgreSQL features without abstraction:

- **JSONB** - Store JSON with indexing and querying
- **UUID v7** - Time-ordered UUIDs (ADR-006)
- **Arrays** - PostgreSQL array types
- **LISTEN/NOTIFY** - Real-time notifications
- **Materialized Views** - Pre-computed queries
- **Generated Columns** - Computed fields

```sql
-- Example: JSONB query
SELECT * FROM files WHERE metadata->>'category' = 'image' AND metadata @> '{"size": "large"}'

-- Example: Generated column
ALTER TABLE files ADD COLUMN file_extension TEXT GENERATED ALWAYS AS 
    (LOWER(SUBSTRING(filename FROM '\.([^.]+)$')) STORED;
```

## Connection Pool Configuration

```go
config, err := pgxpool.ParseConfig(databaseURL)
if err != nil {
    return nil, fmt.Errorf("parse database URL: %w", err)
}

// Recommended production settings
config.MaxConns = 25                    // Connection pool size
config.MinConns = 5                     // Minimum connections
config.MaxConnLifetime = time.Hour      // Connection lifetime
config.MaxConnIdleTime = 30 * time.Minute // Idle timeout
config.HealthCheckPeriod = 1 * time.Minute // Health check interval

pool, err := pgxpool.NewWithConfig(ctx, config)
```

## Testing with PostgreSQL

All tests use **PostgreSQL 16** via testcontainers:

```go
func TestUserRepository_Create(t *testing.T) {
    ctx := context.Background()
    
    // ✅ Real PostgreSQL for all tests
    container, err := testcontainers.Run(ctx, "postgres:16-alpine")
    require.NoError(t, err)
    defer container.Terminate(ctx)
    
    pool, err := pgxpool.New(ctx, container.ConnectionString())
    require.NoError(t, err)
    defer pool.Close()
    
    // Run migrations
    // Test repository
}
```

**NEVER use SQLite for testing** - different behavior, features, performance.

## Enforcement

This standard is **mandatory** for all database code:

1. **Code Review** - All repository PRs must follow this ADR
2. **Linting** - No sqlx, GORM, SQL drivers other than pgx
3. **Testing** - Integration tests with PostgreSQL 16
4. **Documentation** - All repositories documented with patterns

## Rationale

### Why This Stack?

| Component | Reason |
|-----------|--------|
| **PostgreSQL 16** | Production database, must test with real DB |
| **pgx/v5** | Fastest driver,zero reflection, native PostgreSQL features |
| **scany v2** | Minimal reflection (struct tags only), eliminates boilerplate |
| **squirrel** | Type-safe dynamic queries, no string concatenation |

### Why NOT Alternatives?

| Alternative | Rejected Because |
|-------------|-----------------|
| **sqlx** | Reflection overhead, not optimized for pgx |
| **GORM/ENT** | ORM anti-patterns, N+1 problems, abstraction leaks |
| **sqlc** | Poor dynamic query support, code generation complexity |
| **SQLite** | Different features, behavior, performance than PostgreSQL |

## References

- ADR-006: UUID v7 Implementation
- PostgreSQL 16 Documentation: https://www.postgresql.org/docs/16/
- pgx v5: https://github.com/jackc/pgx
- scany v2: https://github.com/georgysavva/scany
- squirrel: https://github.com/Masterminds/squirrel

## Date

2026-04-08

## Authors

Repository optimization based on Google engineers' recommendations. Validated with PostgreSQL 16, UUID v7, JSONB. Tested with 200+ integration tests.