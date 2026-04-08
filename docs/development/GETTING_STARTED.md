# Getting Started

## Prerequisites

- Go 1.24+
- Make
- OpenSSL (for key generation)
- Docker (optional)

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

The project uses semantic versioning with environment suffix:

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

All protected endpoints use session cookies — the token is automatically passed via `Set-Cookie` during login.

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

The system uses different token formats depending on the environment:

### Development/Test (`APP_ENV=dev` or `APP_ENV=test`)

Uses **MockTokenService** to simplify development without cryptographicoperations:

**Access token:**
```
access-{user_id}-{timestamp}
```
Example: `access-019d6746-a5ee-7c00-961f-26d4258d5a32-1775554191`

**Refresh token:**
```
refresh-{UUIDv7}
```
Example: `refresh-019d6746-e0b4-7a00-a177-b41b9b2b9c17`

**Features:**
- No RSA keys required
- Fast generation and validation
- Grant full access (`*:*` wildcard permission)
- User ID visible directly in token (for debugging)

### Production (`APP_ENV=prod`)

Uses **JWTService** with RS256 signature:

**Access token:**
```
eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMDE5ZDY3NDYtLi4uIiwicm9sZXMiOlsiYWRtaW4iXSwicGVybWlzc2lvbnMiOlsiKjoqIl0sImV4cCI6MTc3NTU1...} (JWT format)
```

**Refresh token:**
```
eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9... (JWT format)
```

**Features:**
- RSA keys required (`./keys/private.pem`, `./keys/public.pem`)
- Generate with: `make keys`
- cryptographic signature validation
- Real permissions from RBAC model
- Access TTL: 15 minutes (default)
- Refresh TTL: 7 days (default)

**Warning:** Ensure `APP_ENV=prod` in production, otherwise the system will use mock tokens!

### Audit Logs (requires: audit:read)

```bash
# List audit records (requires authentication)
curl http://localhost:8080/api/v1/audit/records \
  -b "session=<session_id>"

# Filter by actor
curl "http://localhost:8080/api/v1/audit/records?actor_id=019d65d6-de90-7200-b1cf-4f8745597e0a" \
  -b "session=<session_id>"

# Filter by date range
curl "http://localhost:8080/api/v1/audit/records?date_from=2024-01-01T00:00:00Z&date_to=2024-12-31T23:59:59Z" \
  -b "session=<session_id>"
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

All list endpoints use cursor-based pagination:

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

After `make seed`:
- Email: `admin@skeleton.local`
- Password: `Admin1234!`
- Role: `super_admin`

## Swagger Documentation

Swagger UI is available at: http://localhost:8080/swagger/index.html

```bash
# Generate swagger docs
make swagger

# Generate and serve Swagger UI
make swagger-serve
```

All HTTP handlers must have swagger annotations (see [ADR-009](../adr/ADR-009-swagger-annotations.md)).

## Docker Support

The project supports Docker for development and production.

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

## Notifications

The project includes a full-featured notification system with support for:
- Multi-channel (Email, SMS, Push, In-App)
- Template-based messages
- User preferences
- Background worker with retry logic

### Quick Start

```bash
# After setup, notifications start automatically with the main app
make dev

# Create notification via API (admin only)
curl -X POST http://localhost:8080/api/v1/notifications \
  -b "session=<session_id>" \
  -H "Content-Type: application/json" \
  -d '{
    "channel": "email",
    "email": "user@example.com",
    "subject": "Welcome!",
    "content": "Welcome to our platform!",
    "priority": "normal"
  }'
```

### Development Mode

In development, notifications log to console:

```
========== EMAIL ==========
To: user@example.com
Subject: Welcome!
----------------------------
Welcome to our platform!
============================
```

### User Preferences

```bash
# Get user notification preferences
curl http://localhost:8080/api/v1/notifications/preferences \
  -b "session=<session_id>"

# Update preferences
curl -X PATCH http://localhost:8080/api/v1/notifications/preferences \
  -b "session=<session_id>" \
  -H "Content-Type: application/json" \
  -d '{
    "channels": {
      "email": {
        "enabled": true,
        "frequency": "immediate"
      },
      "sms": {
        "enabled": false
      }
    }
  }'
```

### See Full Documentation

Detailed usage, templates, worker configuration: [NOTIFICATIONS.md](./NOTIFICATIONS.md)

## Caching

The application supports multi-tier caching for performance optimization.

### Configuration

```bash
# Cache configuration
CACHE_TYPE=memory                    # memory | redis
CACHE_DEFAULT_TTL=300                 # Default TTL in seconds
CACHE_REDIS_PREFIX=skeleton          # Key prefix for Redis
CACHE_CLEANUP_INTERVAL=60            # Cleanup interval for memory cache
```

### Usage

```go
// In development - in-memory cache
cache := cache.NewMemoryCache(time.Minute)

// Set a value
err := cache.Set(ctx, "user:123", userData, 5*time.Minute)

// Get a value
var user User
err := cache.Get(ctx, "user:123", &user)

// Delete by pattern
err := cache.DeleteByPattern(ctx, "user:*")
```

### HTTP Caching

```go
// Cache GET responses for 5 minutes
api.GET("/users", 
    middleware.Cache(cacheClient, 5*time.Minute),
    handler.ListUsers)
```

### See Full Documentation

Detailed architecture, Redis integration: [ADR-014: Caching Strategy](../adr/ADR-014-caching.md)

## Rate Limiting

The application provides built-in rate limiting to protect against abuse.

### Configuration

```bash
# Rate limiting configuration
RATE_LIMIT_ENABLED=true               # Enable rate limiting
RATE_LIMIT_TYPE=token_bucket          # token_bucket | sliding_window
RATE_LIMIT_KEY_PREFIX=ratelimit      # Key prefix

# Global rate limits
RATE_LIMIT_GLOBAL_LIMIT=1000         # Max requests
RATE_LIMIT_GLOBAL_WINDOW=60          # Time window in seconds

# Auth rate limits (login, register)
RATE_LIMIT_AUTH_LIMIT=5
RATE_LIMIT_AUTH_WINDOW=60

# Files rate limits (upload, download)
RATE_LIMIT_FILES_LIMIT=10
RATE_LIMIT_FILES_WINDOW=60
```

### Usage

```go
// Create rate limiter
limiter := ratelimit.NewTokenBucket(ratelimit.Config{
    Limit:  100,
    Window: time.Minute,
    KeyPrefix: "api",
})

// Apply to endpoints
auth.POST("/login",
    middleware.RateLimit(limiter, middleware.ByIP, 5, time.Minute),
    handler.Login)

// Rate limit per user
api.GET("/users",
    middleware.RateLimit(limiter, middleware.ByUser, 100, time.Minute),
    handler.ListUsers)
```

### Key Strategies

```go
middleware.ByIP             // Rate limit by client IP
middleware.ByUser           // Rate limit by authenticated user ID
middleware.ByEndpoint       // Rate limit by endpoint
middleware.ByUserAndEndpoint // Rate limit by user + endpoint combination
```

### Response Headers

```
X-RateLimit-Limit: 100        # Max requests allowed
X-RateLimit-Remaining: 95     # Requests remaining in window
X-RateLimit-Reset: 1640995200 # Unix timestamp when window resets
Retry-After: 60               # Seconds to wait (on 429)
```

### See Full Documentation

Detailed architecture, Sliding Window with Redis: [ADR-015: Rate Limiting Strategy](../adr/ADR-015-rate-limiting.md)
