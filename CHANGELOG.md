# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added ✨

#### Integration Tests & Development Environment (April 2026) 🧪
- **Integration Test Suite**: Cross-context flow verification
  - `tests/integration/testutil.go` - Test database helpers
  - `tests/integration/event_bus_test.go` - Event bus functionality tests  
  - `tests/integration/order_invoice_test.go` - Order → Invoice integration tests
  - All tests passing, verification of event-driven architecture
  
- **Docker Compose Development Environment**: Complete local development setup
  - PostgreSQL 16 with health checks and persistent storage
  - Redis 7 for caching and event bus
  - API Server with hot reload support
  - Swagger UI standalone server (port 8081)
  - pgAdmin database management UI (port 5050)
  - Full documentation in `docs/DOCKER_DEVELOPMENT.md`
  
- **Swagger/OpenAPI Documentation**: Interactive API documentation
  - Complete OpenAPI 2.0 specification (`docs/swagger/`)
  - Interactive Swagger UI at `/swagger/index.html`
  - 12+ bounded contexts documented with authentication methods
  - Session and Bearer authentication support documented
  - Removed obsolete `docs/api/` directory
  
- **Makefile Improvements**: Updated swagger commands
  - `make swagger` - Instructions for manually maintained docs
  - `make swagger-serve` - Local documentation serving guide
  
#### Cross-Context Integration (NEW) 🔗
- **Event-Driven Architecture**: Domain events enable bounded context communication
- **Order → Inventory Integration**: Automatic stock reservation on OrderConfirmed
- **Order → Invoice Integration**: Automatic invoice creation on OrderConfirmed
- **Invoice → Accounting Integration**: Automatic journal entries on InvoiceCreated
- **Event Handlers**: 
  - `inventory/infrastructure/eventhandler/order_handler.go` - Stock reservation lifecycle
  - `invoicing/infrastructure/eventhandler/order_handler.go` - Auto-invoice creation
  - `accounting/infrastructure/eventhandler/invoice_handler.go` - Journal entry automation
- **Wire Integration**: All event handlers registered in `cmd/api/wire.go`
- **Event Bus**: Type-safe `eventbus.Handler` interface for event subscription

#### Integration Tests ✅
- **Test Infrastructure**: `tests/integration/testutil.go` with test database helpers
- **Event Bus Tests**: Cross-context event delivery verification
- **Order → Invoice Tests**: End-to-end invoice creation from order events
- **Integration Coverage**: Order confirmation flow, invoice calculation, event bus pub/sub

#### Development Environment 🐳
- **Docker Compose**: Complete development environment setup
  - PostgreSQL 16 with health checks
  - Redis 7 for caching and event bus
  - API Server with hot reload (Dockerfile.dev)
  - Swagger UI standalone server (port 8081)
  - pgAdmin for database management (port 5050)
- **Development Documentation**: `docs/DOCKER_DEVELOPMENT.md` with complete guide
- **Production Dockerfile**: Multi-stage build for optimized images

#### API Documentation 📖
- **Swagger/OpenAPI 2.0**: Complete API specification
  - `docs/swagger/swagger.json` - JSON format
  - `docs/swagger/swagger.yaml` - YAML format
  - `docs/swagger/index.html` - Swagger UI interface
- **Interactive Documentation**: Access at http://localhost:8080/swagger/index.html
- **API Tags**: 12+ bounded contexts documented (auth, users, parties, contracts, accounting, ordering, catalog, invoicing, inventory, documents, files, status)
- **Authentication Docs**: Session and Bearer token authentication methods
- **Swagger UI**: Standalone container with Docker Compose

#### Ordering Domain Events 📢
- **OrderCreated**: Published when order is created
- **OrderConfirmed**: Published when order status changes to confirmed (triggers inventory + invoicing)
- **OrderCancelled**: Published when order is cancelled (triggers stock release)
- **OrderCompleted**: Published when order is completed (triggers stock fulfillment)
- **OrderStatusChanged**: Published on any status transition (backward compatibility)

#### Inventory Management (NEW) 🏭
- **Warehouse Management**: Create, update, activate/deactivate/maintenance warehouses
- **Stock Management**: Real-time inventory levels with quantity tracking
- **Stock Movements**: Receipts, issues, transfers, adjustments with full history
- **Stock Reservations**: Order-based reservations with expiration tracking
- **HTTP Endpoints**: 18 endpoints for complete inventory management
  - Warehouses: Create, Update, Get, List (4 endpoints)
  - Stock: Create, Adjust, Receipt, Issue, Transfer, Reserve, Get, List (8 endpoints)
  - Reservations: Fulfill, Cancel, Get, List (4 endpoints)
  - Movements: Get, List (2 endpoints)
- **Domain Tests**: 26 tests covering all aggregates (Warehouse, Stock, Movement, Reservation)
- **Database Migration**: 4 tables with ENUMs for status management
  - `warehouses` with `warehouse_status` enum (active/inactive/maintenance)
  - `stock` with available quantity constraints
  - `stock_movements` with `movement_type` enum (receipt/issue/transfer/adjustment/return)
  - `stock_reservations` with `reservation_status` enum (active/fulfilled/cancelled/expired)
- **Wire Integration**: Full dependency injection in `cmd/api/wire.go`
- **Routes Integration**: All endpoints registered with auth + RBAC middleware

#### Database & Infrastructure
- PostgreSQL 16 as primary database with native UUID v7 support
- JSONB columns with GIN indexes for all metadata fields
- Generated columns for computed fields (file extensions, etc.)
- Materialized views for complex aggregations
- Testcontainers integration for all tests
- Connection pooling with pgxpool (max 25 connections, min 5)

#### Repository Optimization
- `scany v2` for struct scanning (eliminates boilerplate)
- `squirrel` for type-safe dynamic queries
- DTO pattern across all repositories
- Consistent error handling with `pgxscan.NotFound()`
- **30-47% code reduction** in repository layer

#### Architecture Decisions (ADR)
- ADR-016: Database Stack Standard (consolidated)
  - Merged ADR-005 (No ORM)
  - Merged ADR-016 (pgx with PostgreSQL)
  - Merged ADR-017 (scany + squirrel)
- ADR README with quick reference table
- 15 active ADRs (down from 17)
- Archived obsolete ADRs to `docs/adr/archive/`

#### Docker & Deployment
- Multi-stage Dockerfile with health checks
- Development Dockerfile with Air hot reload
- Docker Compose for all environments:
  - `docker-compose.yml` - Development
  - `docker-compose.staging.yml` - Staging
  - `docker-compose.prod.yml` - Production
  - `docker-compose.test.yml` - Testing
- Deployment scripts for staging
- PostgreSQL and Redis monitoring queries

#### Documentation
- `QUICKSTART.md` - 5-minute setup guide
- `DATABASE_MIGRATION_GUIDE.md` - PostgreSQL 16 features
- `DEVELOPMENT.md` - Full development workflow
- `DEPLOYMENT_GUIDE.md` - Production deployment
- `MAKEFILE_REFERENCE.md` - All commands
- `ENVIRONMENT_SETUP.md` - Configuration guide
- `docs/adr/README.md` - ADR index
- PostgreSQL monitoring queries (`docs/monitoring/`)

#### Code Quality
- Benchmark suite for performance testing (`internal/benchmark/`)
- Test utilities package (`pkg/testutil/`)
- Redis client wrapper (`pkg/redis/`)
- PostgreSQL connection pool (`pkg/database/postgres.go`)

### Changed 🔄

#### Repository Layer (Breaking)
- `UserRepository` - 36% code reduction with scany v2
- `FileRepository` - 33% code reduction with squirrel
- `TaskRepository` - 47% code reduction with scany v2
- All other repositories updated to PostgreSQL
- Manual `Scan()` replaced with `pgxscan.Get/Select()`
- String concatenation replaced with `squirrel.Where()`

#### Configuration
- All `.env` files moved to `configs/` directory
- SQLite config removed (`.env.dev`, `.env.example` in root)
- Updated all configs to PostgreSQL 16
- `pkg/database/sqlite.go` removed
- `internal/testutil/fixtures.go` removed

#### Migrations
- Consolidated to 17 PostgreSQL 16 migrations (001-017)
- Removed 22 old SQLite migrations
- All tables now use UUID v7 primary keys
- All metadata columns now use JSONB type
- Added generated columns for computed values

#### Docker
- `Dockerfile` - Production-ready multi-stage build
- `Dockerfile.dev` - Development with Air hot reload
- Health checks on `/health` endpoint
- Non-root user for security

### Fixed 🐛

- Parameter counting bugs in dynamic queries (now using squirrel)
- Boilerplate scanning code (now using scany v2)
- Type safety in query builders
- Error handling consistency (`pgxscan.NotFound()`)
- Docker container startup reliability

### Performance 🚀

- UUID v7: 56% storage reduction vs TEXT
- JSONB with GIN indexes: 10-100x faster queries vs TEXT
- Repository code: 30-47% reduction
- Connection pooling: Efficient resource usage
- Query performance optimizations in migrations

### Security 🔒

- Multi-stage Docker builds (smaller attack surface)
- Non-root user in Docker containers
- Health checks for container orchestration
- `sslmode=require` in production configs
- Secrets via environment variables only

---

## Migration Guide

### From SQLite to PostgreSQL 16

**⚠️ IMPORTANT**: This is a complete database reset. All data will be lost.

#### Step 1: Backup Existing Data (Optional)

```bash
# If you need to preserve data, export it first
# SQLite data cannot be directly migrated to PostgreSQL
# You'll need to write custom migration scripts
```

#### Step 2: Update Configuration

```bash
# Remove old config files from root
rm .env.dev .env.example

# Copy new PostgreSQL config
cp configs/.env.example configs/.env.dev

# Update configs/.env.dev with your settings
DATABASE_URL=postgres://skeleton:password@localhost:5432/skeleton?sslmode=disable
REDIS_URL=redis://localhost:6379/0
```

#### Step 3: Start PostgreSQL + Redis

```bash
# Start Docker containers
make docker-up

# Check status
make docker-status
```

#### Step 4: Run Migrations

```bash
# Apply all PostgreSQL 16 migrations
make migrate-up

# Verify migrations
make migrate-status
```

#### Step 5: Seed Data

```bash
# Create initial admin user and roles
make seed
```

#### Step 6: Verify

```bash
# Run all tests
make test

# Start application
make run
```

### From Old Repository Pattern to New

#### Before (Manual Scanning)

```go
func (r *UserRepository) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
    query := `SELECT id, email FROM users WHERE id = $1`
    row := r.pool.QueryRow(ctx, query, id)
    
    var userID, email string
    err := row.Scan(&userID, &email)
    if err != nil {
        return nil, err
    }
    
    return r.mapToDomain(userID, email)
}
```

#### After (scany v2 + squirrel)

```go
type userDTO struct {
    ID    string `db:"id"`
    Email string `db:"email"`
}

func (r *UserRepository) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
    var dto userDTO
    err := pgxscan.Get(ctx, r.pool, &dto,
        `SELECT id, email FROM users WHERE id = $1`, id)
    if err != nil {
        if pgxscan.NotFound(err) {
            return nil, domain.ErrUserNotFound
        }
        return nil, fmt.Errorf("find user: %w", err)
    }
    return r.dtoToDomain(dto)
}
```

### From String Concatenation to Squirrel

#### Before

```go
query := `SELECT * FROM users WHERE 1=1`
args := make([]interface{}, 0)
argNum := 1

if filter.Search != "" {
    query += fmt.Sprintf(" AND email LIKE $%d", argNum)
    args = append(args, "%"+filter.Search+"%")
    argNum++
}
```

#### After

```go
q := r.psql.Select("*").From("users")

if filter.Search != "" {
    q = q.Where(squirrel.ILike{"email": "%" + filter.Search + "%"})
}

query, args, err := q.ToSql()
```

### Configuration Changes

| Old Location | New Location | Status |
|-------------|-------------|--------|
| `.env.dev` (root) | `configs/.env.dev` | ✅ Required |
| `.env.example` (root) | `configs/.env.example` | ✅ Template |
| N/A | `configs/.env.test` | ✅ Testcontainers |
| N/A | `configs/.env.prod` | ✅ Production |
| N/A | `configs/.env.prod.example` | ✅ Template |

**Code loads from:** `configs/.env.{APP_ENV}

### Database Schema Changes

| Table | Old Column Type | New Column Type | Notes |
|-------|----------------|-----------------|-------|
| `users` | `id TEXT(36)` | `id UUID` | UUID v7 |
| `users` | N/A | `created_at` indexed | Generated from UUID |
| `files` | `metadata TEXT` | `metadata JSONB` | Queryable JSON |
| `tasks` | `payload TEXT` | `payload JSONB` | Queryable JSON |
| All | `TEXT PRIMARY KEY` | `UUID PRIMARY KEY` | UUID v7 |

---

## [0.1.0] - 2025-03-15

### Added
- Initial release with SQLite support
- Basic authentication with JWT
- User management
- Role-based access control (RBAC)
- File upload functionality
- Task queue system
- Notification system

---

## Versioning Strategy

This project follows [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes (database schema, API, configuration)
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible

### Breaking Changes Policy

Breaking changes will be documented with:
- ⚠️ Warning emoji in changelog
- Migration guide section
- Breaking changes list
- Code examples (before/after)

### Deprecation Policy

Features will be deprecated for at least one MINOR version before removal:
1. Announce deprecation in CHANGELOG
2. Add runtime warning
3. Update documentation
4. Remove in next MAJOR version

---

## Migration Commands

```bash
# Fresh start (⚠️ DELETES ALL DATA)
make fresh-start

# Check migration status
make migrate-status

# Rollback last migration
make migrate-down

# Apply new migrations
make migrate-up

# Reset and reseed
make reset-db && make seed
```

---

## Documentation

- [ADR-016: Database Stack Standard](docs/adr/ADR-016-database-stack.md)
- [Quick Start Guide](QUICKSTART.md)
- [Database Migration Guide](docs/DATABASE_MIGRATION_GUIDE.md)
- [Development Workflow](docs/DEVELOPMENT.md)
- [Deployment Guide](docs/DEPLOYMENT_GUIDE.md)
- [Makefile Reference](MAKEFILE_REFERENCE.md)
- [Environment Setup](docs/ENVIRONMENT_SETUP.md)

---

## Support

- **Issues**: GitHub Issues
- **Documentation**: [docs/](docs/)
- **Quick Start**: `make fresh-start`
- **Migration Help**: See [Migration Guide](#migration-guide) above

---

## Contributors

See [CONTRIBUTING.md](docs/development/CONTRIBUTING.md) for contribution guidelines.

---

## License

MIT License - see [LICENSE](LICENSE) for details.