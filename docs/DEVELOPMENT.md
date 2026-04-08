# Development Guide

Complete guide for developers working with the Skeleton API project.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Development Workflow](#development-workflow)
3. [Database Management](#database-management)
4. [Docker Operations](#docker-operations)
5. [Testing](#testing)
6. [Code Style & Patterns](#code-style--patterns)
7. [Common Tasks](#common-tasks)
8. [Troubleshooting](#troubleshooting)

## Getting Started

### Prerequisites

- **Go 1.25+** - Required for air hot reload
- **Docker & Docker Compose** - For PostgreSQL 16 + Redis
- **Make** - For convenient commands
- **golangci-lint** - For code quality (optional, `brew install golangci-lint`)

### Initial Setup

```bash
# 1. Clone repository
git clone <repo-url>
cd skeleton

# 2. Complete setup from scratch (⚠️ deletes all data)
make fresh-start

# This automatically:
# - Generates RSA keys for JWT
# - Starts PostgreSQL 16 + Redis containers
# - Applies all migrations (17 total)
# - Seeds initial data (admin user, roles, permissions)
# - Starts API server with hot reload

# 3. Check status
make status

# 4. Verify health
make health
```

### What Gets Created

**Database Tables (17 tables):**
- Users & auth: `users`, `roles`, `user_roles`, `permissions`, `role_permissions`, `refresh_tokens`
- Business entities: `files`, `file_uploads`, `file_processings`, `notifications`, `notification_templates`, `notification_preferences`
- System: `tasks`, `task_schedules`, `dead_letters`, `audit_records`
- Migration tracking: `schema_migrations`

**Functions:**
- `uuid_generate_v7()` - Time-sortable UUID generation
- `uuid_v7_to_timestamp(uuid)` - Extract timestamp from UUID

**Seed Data:**
- Admin user: `admin@skeleton.local` / `Admin1234!`
- Roles: `super_admin`, `admin`, `viewer`
- Permissions: `users:read`, `users:write`, `users:delete`, etc.

## Development Workflow

### Daily Development

```bash
# Start your workday
make docker-up          # Start containers (if not running)
make migrate-status    # Check migrations are current
make docker-logs-app    # Check API logs (optional)

# Make code changes...
# Hot reload automatically applies changes

# Run tests
make test-unit          # Fast unit tests
make test-integration  # Slower integration tests

# Check code quality
make lint               # Run golangci-lint

# End of day
make docker-down        # Stop containers (optional, keeps data)
```

### Database Changes

```bash
# 1. Create new migration
touch migrations/018_new_table.up.sql
touch migrations/018_new_table.down.sql

# 2. Write migration (use UUID v7!)
-- migrations/018_new_table.up.sql:
CREATE TABLE new_table (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    name TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_new_table_name ON new_table(name);
CREATE INDEX idx_new_table_metadata_gin ON new_table USING GIN (metadata);

-- 3. Apply migration
make migrate-up

# 4. Verify
make db-tables
make psql  # Interactive session
```

### Testing Workflow

```bash
# Quick unit tests (no Docker needed)
make test-unit

# Integration tests (requires Docker)
make test-integration

# All tests
make test

# With coverage report
make test-cover
open coverage.html

# Specific test
go test ./internal/identity/domain/... -v

# Benchmarks
make bench
```

### Git Workflow

```bash
# Create feature branch
git checkout -b feature/user-profiles

# Make changes...

# Run checks before commit
make ci  # lint + test

# Commit
git add .
git commit -m "feat: add user profile management"

# Push
git push origin feature/user-profiles
```

## Database Management

### Connection Methods

```bash
# Interactive psql (in Docker)
make psql

# Execute single query
make db-sql SQL='SELECT * FROM users LIMIT 5;'

# From host machine
psql postgres://skeleton:skeleton_password@localhost:5432/skeleton
```

### Querying & Analysis

```bash
# List all tables
make db-tables

# Table sizes and row counts
make db-stats

# Active connections
make db-connections

# Slow queries (>100ms)
make db-slow-queries

# Index usage (find unused indexes)
make db-index-usage

# Cache hit ratio (should be >99%)
make db-cache-ratio

# Migration history
make db-migrations
```

### Common SQL Queries

```sql
-- Check UUID v7 generation
SELECT uuid_generate_v7(), uuid_v7_to_timestamp(uuid_generate_v7());

-- Query JSONB metadata
SELECT * FROM files 
WHERE metadata @> '{"width": 1920}';

-- Check index usage
SELECT schemaname, tablename, indexname, idx_scan 
FROM pg_stat_user_indexes 
ORDER BY idx_scan ASC;

-- Find missing indexes
SELECT schemaname, tablename, attname, n_distinct, correlation
FROM pg_stats
WHERE n_distinct > 100 AND correlation < 0.1;
```

### Backup & Restore

```bash
# Create backup
make db-backup
# Creates: backups/backup_20260408_150000.dump

# List backups
ls -lh backups/

# Restore from backup
make db-restore BACKUP_FILE=backups/backup_20260408_150000.dump

# Export to SQL (text format)
make db-export
# Creates: exports/export_20260408_150000.sql

# Import from SQL
make db-import SQL_FILE=exports/export_20260408_150000.sql
```

### Data Operations

```bash
# Empty all tables (keep schema) ⚠️
make db-truncate

# Drop all tables (remove schema) ⚠️
make db-drop-tables

# Reset migrations (drop + re-apply) ⚠️
make migrate-reset

# Full reset with seed data ⚠️
make db-drop-tables
make migrate-up
make seed
```

## Docker Operations

### Container Management

```bash
# Start in background (keeps existing data)
make docker-up

# Stop (keeps containers and data)
make docker-stop

# Start stopped containers
make docker-start

# Restart containers
make docker-restart

# Stop and remove containers (keeps volumes)
make docker-down

# Delete everything (containers + volumes) ⚠️
make docker-drop

# Complete rebuild from scratch ⚠️
make docker-reset
# or
make fresh-start
```

### Viewing Logs

```bash
# All services logs
make docker-logs

# Specific service logs
make docker-logs-app     # API only
make docker-logs-db      # PostgreSQL only
make docker-logs-redis   # Redis only

# Real-time monitoring
make watch-logs

# Last N lines
docker-compose logs --tail=100 skeleton-api-dev
```

### Container Shell

```bash
# Shell in API container
make docker-shell

# Root shell (for system packages)
make docker-shell-root

# Execute single command
make docker-exec CMD='ls -la /app'

# PostgreSQL shell
make db-shell
```

### Container Status

```bash
# Basic status
make docker-ps

# Detailed status with health checks
make docker-status

# System-wide status
make status
```

### Common Scenarios

**Rebuild after Dockerfile changes:**
```bash
make docker-drop
make docker-up
# or
make fresh-start
```

**Port conflicts:**
```bash
# Find what's using port 8080
lsof -i :8080

# Stop conflicting service
make docker-down
```

**Container won't start:**
```bash
# Check logs
make docker-logs-app

# Complete reset
make docker-drop
make fresh-start
```

## Testing

### Test Types

**Unit Tests** (fast, no external dependencies):
```bash
# Run unit tests only
make test-unit

# Specific package
go test ./internal/identity/domain/... -v

# With race detection
go test -race ./internal/audit/domain/...
```

**Integration Tests** (requires Docker):
```bash
# Run integration tests
make test-integration

# Uses testcontainers to spin up PostgreSQL
# Automatically creates test database
# Cleans up after tests
```

**All Tests:**
```bash
# Run everything
make test

# With coverage
make test-cover
open coverage.html
```

### Writing Tests

**Unit Test Pattern:**
```go
func TestUser_Email(t *testing.T) {
    // Arrange
    email, err := domain.NewEmail("test@example.com")
    require.NoError(t, err)
    
    // Act & Assert
    assert.Equal(t, "test@example.com", email.String())
    assert.True(t, email.IsValid())
}
```

**Integration Test Pattern:**
```go
func TestUserRepository_Create(t *testing.T) {
    if testing.Short() {
        t.Skip("integration test")
    }
    
    // Setup (uses testcontainers)
    pool, cleanup := setupTestDB(t)
    defer cleanup()
    
    repo := postgres.NewUserRepository(pool)
    
    // Test
    user, err := repo.Create(context.Background(), params)
    
    // Assert
    require.NoError(t, err)
    assert.NotEmpty(t, user.ID())
}
```

### Benchmarks

```bash
# Run benchmarks
make bench

# Compare with baseline
make bench-save
# ... make changes ...
make bench
diff benchmark_results/baseline.txt benchmark_results/latest.txt

# CPU profile
go test -bench=. -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

## Code Style & Patterns

### Project Structure

```
internal/
├── identity/                 # Bounded context
│   ├── domain/              # Domain layer
│   │   ├── user.go          # Aggregate root
│   │   ├── email.go         # Value object
│   │   ├── ids.go           # Typed IDs (UUID v7)
│   │   └── events.go        # Domain events
│   ├── application/          # Application layer
│   │   ├── commands.go      # Write operations
│   │   ├── queries.go       # Read operations
│   │   └── handlers.go      # Event handlers
│   └── infrastructure/       # Infrastructure layer
│       └── persistence/
│           └── user_repository.go
└── ...
```

### Domain-Driven Design Rules

1. **Entities have identity**: `type User struct { id UserID }`
2. **Value objects are immutable**: `type Email string`
3. **Aggregates enforce invariants**: `func (u *User) SetEmail(e Email) error`
4. **Domain events**: `type UserRegistered struct { UserID UserID }`
5. **Repositories interface**: `type UserRepository interface { Save(ctx, *User) error }`

### UUID v7 Usage

```go
// In domain layer
type UserID uuid.UUID

func NewUserID() UserID {
    return UserID(uuid.NewV7())  // Time-sortable!
}

// In repository
func (r *UserRepository) Create(ctx context.Context, user *User) error {
    _, err := r.pool.Exec(ctx,
        `INSERT INTO users (id, email, ...) VALUES ($1, $2, ...)`,
        user.ID(),   // UUID type, not string!
        user.Email(),
    )
    return err
}

// Extract timestamp from UUID
createdAt := uuid_v7_to_timestamp(user.ID())
```

### JSONB Patterns

```go
// In domain
type Metadata map[string]interface{}

// In repository - insert
metadataJSON, _ := json.Marshal(metadata)
_, err := r.pool.Exec(ctx,
    `INSERT INTO files (..., metadata) VALUES (..., $1)`,
    metadataJSON,  // Go marshals to JSONB
)

// Query JSONB
rows, err := r.pool.Query(ctx,
    `SELECT * FROM files WHERE metadata @> $1`,
    `{"width": 1920}`,  // JSONB containment query
)
```

### Error Handling

```go
// Domain errors
var (
    ErrUserNotFound     = errors.New("user not found")
    ErrUserInactive     = errors.New("user is inactive")
    ErrInvalidEmail     = errors.New("invalid email format")
)

// Wrap with context
if user == nil {
    return nil, fmt.Errorf("%w: email=%s", ErrUserNotFound, email)
}

// Application layer
func (s *UserService) CreateUser(ctx context.Context, cmd CreateUserCommand) error {
    user, err := domain.NewUser(cmd.Email, cmd.PasswordHash)
    if err != nil {
        return fmt.Errorf("create user: %w", err)  // Wrap
    }
    return s.repo.Save(ctx, user)
}
```

## Common Tasks

### Add New Entity

```bash
# 1. Create bounded context directory
mkdir -p internal/newfeature/domain
mkdir -p internal/newfeature/application
mkdir -p internal/newfeature/infrastructure/persistence

# 2. Create domain entity
# internal/newfeature/domain/entity.go
package domain

import (
    "github.com/basilex/skeleton/pkg/uuid"
)

type EntityID uuid.UUID

func NewEntityID() EntityID {
    return EntityID(uuid.NewV7())
}

type Entity struct {
    id          EntityID
    name        string
    metadata    Metadata
    createdAt   time.Time
}

func NewEntity(name string) (*Entity, error) {
    return &Entity{
        id:        NewEntityID(),
        name:      name,
        metadata:  make(Metadata),
        createdAt: time.Now(),
    }, nil
}

# 3. Create migration
touch migrations/018_new_entities.up.sql

# 4. Create repository interface
# internal/newfeature/domain/repository.go
type EntityRepository interface {
    Save(ctx context.Context, entity *Entity) error
    FindByID(ctx context.Context, id EntityID) (*Entity, error)
}

# 5. Implement repository
# internal/newfeature/infrastructure/persistence/entity_repository.go

# 6. Create application service
# internal/newfeature/application/service.go

# 7. Create HTTP handler
# internal/newfeature/ports/http/handler.go
```

### Add New Permission

```sql
-- In migration
INSERT INTO permissions (id, name, description, created_at)
VALUES (uuid_generate_v7(), 'entities:read', 'Read entities', NOW());
```

```go
// In code
const (
    PermissionEntitiesRead = "entities:read"
)

// Check permission
if !user.HasPermission(PermissionEntitiesRead) {
    return ErrForbidden
}
```

### Add Background Task

```go
// 1. Create task handler
type MyTaskHandler struct {
    repo Repository
}

func (h *MyTaskHandler) Handle(ctx context.Context, task *Task) error {
    // Process task
    return nil
}

// 2. Register in worker
worker.RegisterHandler("my_task", &MyTaskHandler{repo: repo})

// 3. Schedule task
task := tasks.NewTask("my_task", payload)
taskRepo.Schedule(ctx, task)
```

## Troubleshooting

### Container Issues

**Containers won't start:**
```bash
# Check logs
make docker-logs

# Check Docker status
docker info

# Reset everything
make docker-drop
make fresh-start
```

**Port already in use:**
```bash
# Find process
lsof -i :8080

# Kill process
kill -9 <PID>

# Or change port in docker-compose.yml
```

**API not responding:**
```bash
# Check health
curl http://localhost:8080/health

# Check logs
make docker-logs-app

# Check database connection
make db-connections
```

### Database Issues

**Migration failed:**
```bash
# Check migration status
make migrate-status

# Manual rollback
make migrate-down

# Reset everything
make migrate-reset
```

**Connection refused:**
```bash
# Check if PostgreSQL is running
make docker-ps | grep postgres

# Check connection params
docker exec skeleton-postgres-dev psql -U skeleton -d skeleton -c "SELECT 1;"

# Check environment
make db-sql SQL='SELECT current_database(), current_user;'
```

**UUID not generating:**
```sql
-- Check function exists
make db-sql SQL='\df uuid_generate_v7'

-- Test generation
SELECT uuid_generate_v7(), uuid_v7_to_timestamp(uuid_generate_v7());

-- Re-run init migration
make migrate-reset
```

**JSONB queries slow:**
```sql
-- Check if GIN index exists
SELECT indexname FROM pg_indexes WHERE schemaname = 'public' AND indexname LIKE '%gin%';

-- Create missing index
CREATE INDEX idx_table_metadata_gin ON table USING GIN (metadata);

-- Check query plan
EXPLAIN ANALYZE SELECT * FROM files WHERE metadata @> '{"key":"value"}';
```

### Test Issues

**Integration tests failing:**
```bash
# Check Docker is running
docker ps

# Clean up test containers
docker ps -a | grep test | awk '{print $1}' | xargs docker rm -f

# Rebuild test environment
make test-integration
```

**Race conditions detected:**
```bash
# Run with more iterations
go test -race -count=10 ./...

# Find specific test
go test -race -run TestName ./...
```

### Performance Issues

**Slow queries:**
```bash
# Enable pg_stat_statements
make db-enable-stats

# Find slow queries
make db-slow-queries

# Check index usage
make db-index-usage

# Analyze query plan
make db-sql SQL='EXPLAIN ANALYZE SELECT ...'
```

**High memory usage:**
```bash
# Check connection pool
make db-connections

# Check table bloat
make db-sql SQL="SELECT schemaname, tablename, 
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
  FROM pg_tables WHERE schemaname = 'public';"

# Vacuum analyze
make db-sql SQL='VACUUM ANALYZE;'
```

### Common Error Messages

**"uuid_generate_v7() does not exist":**
```bash
# Function not created, run init migration
make migrate-reset
```

**"relation \"users\" does not exist":**
```bash
# Migrations not applied
make migrate-up
```

**"password authentication failed":**
```bash
# Wrong DATABASE_URL
# Check Makefile has correct credentials
grep DATABASE_URL Makefile
```

## Environment Variables

Essential environment variables (set in Makefile):

```bash
DATABASE_URL=postgres://skeleton:skeleton_password@localhost:5432/skeleton?sslmode=disable
REDIS_URL=redis://localhost:6379
APP_ENV=dev
APP_PORT=8080
```

For production, set via environment:

```bash
export DATABASE_URL="postgres://user:pass@prod-host:5432/db?sslmode=require"
export REDIS_URL="redis://prod-host:6379"
export APP_ENV=prod
```

## Best Practices

1. **Always use UUID v7** for new tables: `id UUID PRIMARY KEY DEFAULT uuid_generate_v7()`
2. **Always use TIMESTAMPTZ** for timestamps: `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
3. **Always use JSONB** for flexible metadata: `metadata JSONB DEFAULT '{}'`
4. **Always add GIN index** for JSONB: `CREATE INDEX idx_table_metadata_gin ON table USING GIN (metadata)`
5. **Never use TEXT** for IDs or timestamps
6. **Run tests before commit**: `make test`
7. **Check migrations in**: `make migrate-status`
8. **Use connnection pooling**: pgxpool in production

## Resources

- **PostgreSQL 16 Docs**: https://www.postgresql.org/docs/16/index.html
- **UUID v7 RFC**: https://www.rfc-editor.org/rfc/rfc9562.html
- **pgx Documentation**: https://github.com/jackc/pgx
- **Domain-Driven Design**: https://domainlanguage.com/ddd/
- **Hexagonal Architecture**: https://alistair.cockburn.us/hexagonal-architecture/

---

**Need help?**
- Check [Troubleshooting](#troubleshooting)
- Review [Makefile Guide](MAKEFILE_REFERENCE.md)
- Read [Architecture Decision Records](docs/adr/)