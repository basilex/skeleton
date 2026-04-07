# Getting Started

## Prerequisites

- Go 1.24+
- Make
- OpenSSL (для генерації ключів)

## Setup

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

详见 [ADR-008: Semantic Versioning Strategy](../adr/ADR-008-versioning.md).

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

## Docker

Проект підтримує Docker для development та production.

### Development (with hot reload)

```bash
# Start with hot reload
make docker-dev

# Or manually
docker-compose up --build
```

Changes to Go files will automatically rebuild and restart.

### Production

```bash
# Build production image
make docker-build

# Run production container
make docker-prod

# Or manually
docker-compose -f docker-compose.prod.yml up -d

# View logs
docker-compose -f docker-compose.prod.yml logs -f
```

### Docker Commands

```bash
make docker-build      # Build production image
make docker-dev        # Start development with hot reload
make docker-prod       # Start production containers
make docker-up         # Start containers in background
make docker-down       # Stop and remove containers
make docker-logs       # View container logs
make docker-ps         # List running containers
make docker-clean      # Remove containers, volumes, images
```

### With Redis (optional)

```bash
# Development with Redis
make docker-dev-redis

# This starts both app and Redis containers
```

### Environment Variables

Production configuration in `configs/.env.prod.example`:
- Copy to `.env.prod` and customize
- Docker Compose will use these values

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
