# Getting Started

## Prerequisites

- Go 1.24+
- Make
- OpenSSL (для генерації ключів)
- Docker (опціонально)

## Quick Start

### Local Development

```bash
# Initial setup (keys + migrate + seed)
make setup

# Run application
make run

# Or one command
make dev
```

### Docker Development

```bash
# Build and run with hot reload
make docker-dev

# Access at http://localhost:8080
```

```bash
# Clone
git clone <repo-url>
cd skeleton

# Copy env
cp configs/.env.example configs/.env.dev

# Generate RSA keys
make keys

# Run migrations
make migrate-up

# Seed dev data
make seed

# Run
make run
```

## Version Management

Проект використовує semantic versioning з environment suffix:

```bash
# Development build (default: 0.1.0-dev)
make build

# Staging build
VERSION_STAGE=staging make build  # 0.1.0-staging

# Production build
VERSION_STAGE=prod make build      # 0.1.0-prod

# Check version
curl http://localhost:8080/build
# {"version":"0.1.0-dev","commit":"c4410c8","build_time":"2026-04-07T10:00:37Z","go_version":"go1.26.1","env":"dev"}
```

See [ADR-008: Semantic Versioning Strategy](../adr/ADR-008-versioning.md).

## API Endpoints

### Status

```bash
# Health check
curl http://localhost:8080/health

# Build info
curl http://localhost:8080/build
```

### Auth (public)

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"Password1234!"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@skeleton.local","password":"Admin1234!"}'
```

### Authenticated (session cookie)

Усі захищені endpoints використовують session cookie — токен передається автоматично через `Set-Cookie` при login.

```bash
# Get my profile (requires session cookie)
curl http://localhost:8080/api/v1/users/me \
  -b "session=<session_id>"

# List users with cursor pagination (requires: users:read)
curl "http://localhost:8080/api/v1/users?limit=20&cursor=019d65d6-de90-7200-b1cf-4f8745597e0a" \
  -b "session=<session_id>"

# Logout (destroys session)
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -b "session=<session_id>"
```

## Token Formats

Система використовує різні формати токенів в залежності від середовища:

### Development/Test (`APP_ENV=dev` або `APP_ENV=test`)

Використовується **MockTokenService** для спрощення розробки без криптографічних операцій:

**Access token:**
```
access-{user_id}-{timestamp}
```
Приклад: `access-019d6746-a5ee-7c00-961f-26d4258d5a32-1775554191`

**Refresh token:**
```
refresh-{UUIDv7}
```
Приклад: `refresh-019d6746-e0b4-7a00-a177-b41b9b2b9c17`

**Особливості:**
- Не потребують RSA ключів
- Швидка генерація та валідація
- Дають повний доступ (`*:*` wildcard permission)
- User ID видно прямо в токені (для дебагу)

### Production (`APP_ENV=prod`)

Використовується **JWTService** з RS256 підписом:

**Access token:**
```
eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMDE5ZDY3NDYtLi4uIiwicm9sZXMiOlsiYWRtaW4iXSwicGVybWlzc2lvbnMiOlsiKjoqIl0sImV4cCI6MTc3NTU1...}
```

**Refresh token:**
```
eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9... (JWT формат)
```

**Особливості:**
- Потрібні RSA ключі (`./keys/private.pem`, `./keys/public.pem`)
- Генерація: `make keys`
- cryptographic signature validation
- Реальні permissions з RBAC моделі
- Access TTL: 15 хвилин (за замовчуванням)
- Refresh TTL: 7 днів (за замовчуванням)

**Увага:** Переконайтеся що `APP_ENV=prod` в продакшені, інакше система буде використовувати mock tokens!

### Audit Logs (requires: audit:read)

```bash
# List audit records (requires authentication)
curl http://localhost:8080/api/v1/audit/records \
  -H "Authorization: Bearer <token>"

# Filter by actor
curl "http://localhost:8080/api/v1/audit/records?actor_id=019d65d6-de90-7200-b1cf-4f8745597e0a" \
  -H "Authorization: Bearer <token>"

# Filter by date range
curl "http://localhost:8080/api/v1/audit/records?date_from=2024-01-01T00:00:00Z&date_to=2024-12-31T23:59:59Z" \
  -H "Authorization: Bearer <token>"
```

### Admin (requires: roles:manage)

```bash
# Assign role
curl -X POST http://localhost:8080/api/v1/users/<user_id>/roles \
  -b "session=<session_id>" \
  -H "Content-Type: application/json" \
  -d '{"role_id":"<role_id>"}'

# Revoke role
curl -X DELETE http://localhost:8080/api/v1/users/<user_id>/roles/<role_id> \
  -b "session=<session_id>"
```

## Pagination

Всі list endpoints використовують cursor-based пагінацію:

```
GET /api/v1/users?limit=20&cursor=019d65d6-de90-7200-b1cf-4f8745597e0a
```

Response:
```json
{
  "items": [...],
  "next_cursor": "019d65d6-de98-7e00-b590-2d70f5506278",
  "has_more": true,
  "limit": 20
}
```

## Dev User

Після `make seed`:
- Email: `admin@skeleton.local`
- Password: `Admin1234!`
- Role: `super_admin`

## Swagger Documentation

Swagger UI доступний за адресою: http://localhost:8080/swagger/index.html

```bash
# Generate swagger docs
make swagger

# Generate and serve Swagger UI
make swagger-serve
```

All HTTP handlers must have swagger annotations (see [ADR-009](../adr/ADR-009-swagger-annotations.md)).

## Docker Support

Проект підтримує Docker для development та production.

### Development (with hot reload)

```bash
# Quick start with Docker
make docker-dev

# Or manually
docker-compose up --build
```

Changes to Go files automatically rebuild and restart.

**Features:**
- Hot reload with Air
- Volume mounts for code changes
- Health checks included
- Optional Redis (`--profile redis`)

### Production

```bash
# Build production image
make docker-build

# Run production containers
make docker-prod

# Or manually
docker-compose -f docker-compose.prod.yml up -d

# View logs
docker-compose -f docker-compose.prod.yml logs -f
```

**Features:**
- Minimal alpine image (~10MB)
- Non-root user for security
- Resource limits (CPU/memory)
- Automatic restarts
- Health checks

### Docker Commands

```bash
make docker-build      # Build production image
make docker-dev        # Development with hot reload
make docker-prod       # Production deployment
make docker-up         # Start in background
make docker-down       # Stop containers
make docker-logs       # View logs
make docker-ps         # List containers
make docker-clean      # Remove containers, volumes, images
make docker-dev-redis  # Development with Redis
```

### With Redis (optional)

```bash
# Development with Redis for session/event bus
make docker-dev-redis

# This starts both app and Redis containers
```

### Environment Variables

**Development:** Copy `configs/.env.example` to `configs/.env.dev`

**Production:** Use `configs/.env.prod.example` as template

```bash
cp configs/.env.prod.example configs/.env.prod
# Edit and customize values
```

### Health Check

Both development and production containers include health checks:

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

### Dockerfile Details

**Multi-stage build:**
- Build stage: `golang:1.24-alpine` (compiles binary)
- Runtime stage: `alpine:3.19` (minimal ~10MB)
- Non-root user for security
- Health check included

**Development Dockerfile:**
- Hot reload with Air
- Fast iteration
- Volume mounts for code changes
