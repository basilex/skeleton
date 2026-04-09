# Setup Guide

> **Installation and Configuration Guide for Skeleton CRM**

## 📋 Prerequisites

### Required

- **Go** 1.25+ - [Download Go](https://golang.org/dl/)
- **Node.js** 20+ (24 recommended) - [Download Node.js](https://nodejs.org/)
- **PostgreSQL** 16+ - [Download PostgreSQL](https://www.postgresql.org/download/)
- **Redis** 7+ - [Download Redis](https://redis.io/download)

### Optional (Recommended)

- **Docker** & **Docker Compose** - [Install Docker](https://docs.docker.com/get-docker/)
- **Make** - Build automation (pre-installed on macOS/Linux)
- **Air** - Go hot reload - `go install github.com/air-verse/air@latest`

---

## 🚀 Quick Start

### Option 1: Docker (Recommended)

```bash
# Clone repository
git clone https://github.com/basilex/skeleton.git
cd skeleton

# Start all services (PostgreSQL + Redis + Backend + Frontend)
make dev

# Wait for services to start...
# PostgreSQL: localhost:5432
# Redis: localhost:6379
# Backend API: localhost:8080
# Frontend: localhost:3000

# In a new terminal, run migrations
make db-migrate

# (Optional) Seed database with sample data
make db-seed

# Default credentials after seed:
# Email: admin@skeleton.local
# Password: Admin1234!
```

### Option 2: Manual Setup

```bash
# 1. Clone repository
git clone https://github.com/basilex/skeleton.git
cd skeleton

# 2. Install dependencies
make install

# 3. Start database services (PostgreSQL + Redis)
make db-up

# 4. Run database migrations
make db-migrate

# 5. (Optional) Seed sample data
make db-seed

# 6. Start backend (Terminal 1)
make backend

# 7. Start frontend (Terminal 2)
make frontend
```

---

## ⚙️ Configuration

### Environment Variables

#### Backend Configuration

Create `backend/.env` (or use system environment variables):

```bash
# Application
APP_ENV=development              # development | staging | production
APP_PORT=8080                    # HTTP server port

# Database
DATABASE_URL=postgres://skeleton:skeleton@localhost:5432/skeleton?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# Session
SESSION_SECRET=your-secret-key-change-in-production
SESSION_TTL=168h                 # 7 days

# JWT
JWT_SECRET=your-jwt-secret-key-change-in-production
JWT_ACCESS_TTL=15m               # 15 minutes
JWT_REFRESH_TTL=168h             # 7 days

# CORS (optional)
CORS_ORIGINS=http://localhost:3000,http://localhost:8080

# Logging
LOG_LEVEL=debug                  # debug | info | warn | error

# Email (optional - for notifications)
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=your-email@example.com
SMTP_PASSWORD=your-password
SMTP_FROM=noreply@example.com

# Storage (optional - for file uploads)
STORAGE_TYPE=local               # local | s3 | minio
STORAGE_PATH=./uploads           # for local storage
# AWS_S3_BUCKET=your-bucket      # for S3
# AWS_S3_REGION=us-east-1
# AWS_ACCESS_KEY_ID=your-key
# AWS_SECRET_ACCESS_KEY=your-secret
```

#### Frontend Configuration

Create `frontend/.env.local`:

```bash
# API Endpoint
NEXT_PUBLIC_API_URL=http://localhost:8080

# Application
NEXT_PUBLIC_APP_NAME=Skeleton CRM

# Environment
NODE_ENV=development              # development | production
```

### Production Environment

```bash
# Backend
APP_ENV=production
DATABASE_URL=postgres://user:password@prod-db:5432/skeleton?sslmode=require
REDIS_URL=redis://prod-redis:6379
SESSION_SECRET=complex-random-secret-min-32-chars
JWT_SECRET=complex-random-secret-min-32-chars
LOG_LEVEL=info

# Frontend
NEXT_PUBLIC_API_URL=https://api.skeleton.com
NODE_ENV=production
```

---

## 🗄️ Database Setup

### Initial Setup

```bash
# Create database (if not using Docker)
createdb skeleton

# Or with psql
psql -U postgres -c "CREATE DATABASE skeleton;"

# Run migrations
make db-migrate

# Seed sample data
make db-seed
```

### Manual Migration

```bash
# Using Makefile
make db-migrate

# Or directly
cd backend
go run ./cmd/api migrate

# Migration files are in backend/migrations/
# Format: 001_description.up.sql / 001_description.down.sql
```

### Reset Database

```bash
# Drop and recreate
make db-reset

# Or manually
psql -U postgres -c "DROP DATABASE IF EXISTS skeleton;"
psql -U postgres -c "CREATE DATABASE skeleton;"
make db-migrate
make db-seed
```

---

## 🐳 Docker Configuration

### docker-compose.yml

Located at project root:

```yaml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: skeleton
      POSTGRES_USER: skeleton
      POSTGRES_PASSWORD: skeleton
    ports:
      - "5432:5432"
    volumes:
      - postgres-data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data

  backend:
    build: ./backend
    depends_on:
      - postgres
      - redis
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: skeleton
      DB_PASSWORD: skeleton
      DB_NAME: skeleton
      REDIS_URL: redis://redis:6379
      PORT: 8080
    ports:
      - "8080:8080"

  frontend:
    build: ./frontend
    depends_on:
      - backend
    environment:
      NEXT_PUBLIC_API_URL: http://backend:8080
    ports:
      - "3000:3000"

volumes:
  postgres-data:
  redis-data:
```

### Running Specific Services

```bash
# Database only
docker-compose up -d postgres redis

# Backend only
docker-compose up -d backend

# All services
docker-compose up -d
```

---

## 🔧 Development Tools

### Air (Hot Reload for Go)

```bash
# Install
go install github.com/air-verse/air@latest

# Run backend with hot reload
cd backend
air

# Or using Makefile
make backend-watch
```

### Database Tools

```bash
# PostgreSQL CLI
make db-shell

# Or directly
docker-compose exec postgres psql -U skeleton -d skeleton

# Redis CLI
docker-compose exec redis redis-cli
```

### Useful Commands

```bash
# View logs
make dev-logs                  # All services
docker-compose logs -f backend # Backend only
docker-compose logs -f frontend # Frontend only

# Check service health
docker-compose ps

# Restart services
docker-compose restart backend

# Clean up
docker-compose down -v         # Stop and remove volumes
```

---

## 🧪 Testing

### Backend Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package
cd backend && go test ./internal/invoicing/...

# Run integration tests
make test-integration
```

### Frontend Tests

```bash
# Run tests
cd frontend && npm test

# Run e2e tests
cd frontend && npm run test:e2e

# Type check
cd frontend && npx tsc --noEmit
```

---

## 📦 IDE Setup

### VSCode (Recommended)

Install extensions:

```json
{
  "recommendations": [
    "golang.go",                 // Go language support
    "ms-vscode.makefile-tools",   // Makefile support
    "dbaeumer.vscode-eslint",     // ESLint
    "esbenp.prettier-vscode",     // Prettier
    "bradlc.vscode-tailwindcss",  // Tailwind CSS
    "ms-azuretools.vscode-docker", // Docker
    "cweijan.vscode-database-client", // Database client
    "redhat.vscode-yaml"          // YAML support
  ]
}
```

### GoLand (Alternative)

1. Open project root
2. Enable Go module support
3. Configure Go SDK (1.25+)
4. Enable database tools for PostgreSQL

---

## 🚨 Troubleshooting

### Port Already in Use

```bash
# Find process using port
lsof -i :8080                  # Backend
lsof -i :3000                  # Frontend
lsof -i :5432                  # PostgreSQL
lsof -i :6379                  # Redis

# Kill process
kill -9 <PID>

# Or change port in config
APP_PORT=8081 make backend
```

### Database Connection Failed

```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Check connection
psql -h localhost -U skeleton -d skeleton -c "SELECT 1;"

# Reset database
make db-reset
make db-migrate
make db-seed
```

### Frontend Build Errors

```bash
# Clear Next.js cache
cd frontend
rm -rf .next node_modules
npm install
npm run build

# Check Node version
node --version  # Should be 20+
```

### Go Module Errors

```bash
# Clean Go cache
cd backend
go clean -modcache
go mod tidy
go mod download
```

---

## 📚 Next Steps

After setup:

1. **Read Architecture**: [docs/ARCHITECTURE.md](ARCHITECTURE.md)
2. **Development Workflow**: [docs/DEVELOPMENT.md](DEVELOPMENT.md)
3. **API Reference**: [docs/API.md](API.md)
4. **Database Schema**: [docs/DATABASE.md](DATABASE.md)

---

## 🆘 Getting Help

- **Documentation**: [docs/](../docs/)
- **Issues**: [GitHub Issues](https://github.com/basilex/skeleton/issues)
- **Discussions**: [GitHub Discussions](https://github.com/basilex/skeleton/discussions)

---

**Setup complete! Run `make dev` to start developing.**