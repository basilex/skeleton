# Setup Guide

Complete guide to setting up and running the Skeleton Business Engine locally.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start](#quick-start)
3. [Local Development Setup](#local-development-setup)
4. [Docker Setup](#docker-setup)
5. [Database Setup](#database-setup)
6. [Configuration](#configuration)
7. [IDE Setup](#ide-setup)
8. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Required Software

| Software | Version | Purpose |
|----------|---------|---------|
| Go | 1.25+ | Programming language |
| PostgreSQL | 16+ | Database |
| Docker | 24+ | Containerization (optional) |
| Docker Compose | 2+ | Multi-container orchestration |
| Make | 3.x | Build automation |
| Git | 2.x | Version control |

### Operating Systems

- macOS 12+
- Ubuntu 20.04+
- Windows 10+ (with WSL2)

---

## Quick Start

### Minimal Setup (5 minutes)

```bash
# 1. Clone repository
git clone <repository-url>
cd skeleton

# 2. Install Go dependencies
make deps

# 3. Setup PostgreSQL database
createdb skeleton
psql -d skeleton -f migrations/*.up.sql

# 4. Run the application
make run

# Server running at http://localhost:8080
```

### Docker Setup (2 minutes)

```bash
# 1. Start all services
docker-compose up -d

# 2. Run migrations
docker-compose exec api make migrate

# 3. View logs
docker-compose logs -f api

# Server running at http://localhost:8080
```

---

## Local Development Setup

### Step 1: Install Go

**macOS:**
```bash
brew install go
```

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install -y golang-go
```

**Verify installation:**
```bash
go version  # Should show go1.25+
```

### Step 2: Install PostgreSQL

**macOS:**
```bash
brew install postgresql@16
brew services start postgresql@16
```

**Ubuntu:**
```bash
sudo apt install -y postgresql-16
sudo systemctl start postgresql
```

**Create database:**
```bash
psql -U postgres -c "CREATE DATABASE skeleton;"
psql -U postgres -c "CREATE USER skeleton WITH PASSWORD 'secret';"
psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE skeleton TO skeleton;"
```

### Step 3: Clone and Configure

```bash
# Clone
git clone <repository-url>
cd skeleton

# Install dependencies
make deps

# Configuration
cp .env.example .env
```

Edit `.env`:
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=skeleton
DB_PASSWORD=secret
DB_NAME=skeleton
SERVER_PORT=8080
JWT_SECRET=your-secret-key-change-in-production
```

### Step 4: Database Setup

```bash
# Run migrations
make migrate

# Or manually
psql -d skeleton -f migrations/001_uuid.up.sql
psql -d skeleton -f migrations/002_audit.up.sql
# ... continue for all migrations
```

### Step 5: Run Application

```bash
# Development mode (with hot reload)
make dev

# Production mode
make run

# Specific package
go run ./cmd/api
```

---

## Docker Setup

### Docker Compose Configuration

`docker-compose.yml` includes:
- API server
- PostgreSQL database
- Redis (optional)
- Adminer (database UI)

### Commands

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f api

# Stop services
docker-compose down

# Rebuild
docker-compose build --no-cache

# Access container
docker-compose exec api sh
```

### Environment Variables

Create `.env` file:
```env
# Database
POSTGRES_HOST=db
POSTGRES_PORT=5432
POSTGRES_USER=skeleton
POSTGRES_PASSWORD=skeleton_secret
POSTGRES_DB=skeleton

# Application
APP_ENV=development
APP_PORT=8080
APP_SECRET=change-me-in-production

# JWT
JWT_SECRET=your-jwt-secret
JWT_EXPIRY=24h
```

---

## Database Setup

### Migrations

Migrations are located in `migrations/` directory:

```
migrations/
├── 001_uuid.up.sql           # UUID extension
├── 002_audit.up.sql          # Audit tables
├── 003_identity.up.sql       # Users, roles, sessions
├── ...
└── 023_invoicing.up.sql     # Invoice tables
```

### Run Migrations

```bash
# All migrations
make migrate

# Or use golang-migrate
migrate -path migrations -database "postgres://user:pass@localhost:5432/skeleton?sslmode=disable" up
```

### Create New Migration

```bash
migrate create -ext sql -dir migrations -seq migration_name
```

### Migration Best Practices

1. **Always use BIGINT for money**:
   ```sql
   total BIGINT NOT NULL DEFAULT 0,  -- Stores cents
   ```

2. **Use UUIDv7 for primary keys**:
   ```sql
   id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
   ```

3. **Add timestamps**:
   ```sql
   created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   ```

4. **Add audit trigger**:
   ```sql
   CREATE TRIGGER table_updated_at
       BEFORE UPDATE ON table
       FOR EACH ROW
       EXECUTE FUNCTION update_updated_at();
   ```

---

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_ENV` | Environment (development/production) | development |
| `APP_PORT` | Server port | 8080 |
| `DB_HOST` | Database host | localhost |
| `DB_PORT` | Database port | 5432 |
| `DB_USER` | Database user | skeleton |
| `DB_PASSWORD` | Database password | - |
| `DB_NAME` | Database name | skeleton |
| `JWT_SECRET` | JWT signing secret | - |
| `JWT_EXPIRY` | Token expiry duration | 24h |

### Configuration Files

- `.env` - Environment variables (not in git)
- `.env.example` - Template file
- `config/config.go` - Configuration struct
- `cmd/api/main.go` - Application entry point

---

## IDE Setup

### VS Code

Install extensions:
```
ms-vscode.go
ms-azuretools.vscode-docker
ckolkman.vscode-postgres
```

Recommended `settings.json`:
```json
{
    "go.useLanguageServer": true,
    "go.lintTool": "golangci-lint",
    "go.lintOnSave": "package",
    "go.testFlags": ["-v"],
    "go.coverOnSave": true
}
```

Launch configuration `.vscode/launch.json`:
```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch API",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/cmd/api",
            "envFile": "${workspaceFolder}/.env"
        }
    ]
}
```

### GoLand

1. Open project
2. Enable Go modules support
3. Set GOROOT to Go installation
4. Configure run/debug configuration for `cmd/api`

---

## Troubleshooting

### Common Issues

#### 1. Database Connection Failed

**Error**: `connection refused`

**Solution**:
```bash
# Check PostgreSQL is running
pg_isready

# Start PostgreSQL
brew services start postgresql@16  # macOS
sudo systemctl start postgresql      # Ubuntu

# Verify connection
psql -h localhost -U skeleton -d skeleton
```

#### 2. Port Already in Use

**Error**: `bind: address already in use`

**Solution**:
```bash
# Find process using port 8080
lsof -i :8080

# Kill process
kill -9 <PID>

# Or change port in .env
APP_PORT=8081
```

#### 3. Migration Failed

**Error**: `dirty database version`

**Solution**:
```bash
# Force version
migrate -path migrations -database "postgres://..." force <version>

# Or reset database
dropdb skeleton
createdb skeleton
make migrate
```

#### 4. Go Dependency Issues

**Error**: `package not found`

**Solution**:
```bash
# Clean go module cache
go clean -modcache

# Re-download dependencies
go mod download
go mod tidy
```

#### 5. Docker Issues

**Error**: `no space left on device`

**Solution**:
```bash
# Clean Docker
docker system prune -a
docker volume prune

# Rebuild
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

---

## Make Commands

```bash
deps              # Install dependencies
build             # Build binary
run               # Run application
dev               # Run with hot reload
test              # Run tests
test-coverage     # Run tests with coverage
migrate           # Run database migrations
db-reset          # Reset database (drop and recreate)
db-setup          # Setup database from scratch
clean             # Clean generated files
help              # Show all commands
```

---

## Next Steps

After setup, proceed to:

1. **[Architecture Overview](ARCHITECTURE.md)** - Understand the system architecture
2. **[Development Guide](DEVELOPMENT.md)** - Learn development workflow
3. **[Testing Strategy](TESTING.md)** - Understand testing approach
4. **[Main README](../README.md)** - Review project features

---

## Getting Help

- **Documentation**: Check other docs in this directory
- **Issues**: Open an issue in the repository
- **Code**: Review inline comments and ADRs
- **Examples**: See `internal/` for bounded context implementations