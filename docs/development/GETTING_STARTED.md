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
