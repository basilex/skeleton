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
cp configs/.env.example configs/.env

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

### Public

```bash
# Build info
curl http://localhost:8080/api/v1/aux/info

# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"Password1234!"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@skeleton.local","password":"Admin1234!"}'
```

### Authenticated

```bash
# Get my profile
curl http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <access_token>"

# List users (requires: users:read)
curl http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer <access_token>"
```

### Admin (requires: roles:manage)

```bash
# Assign role
curl -X POST http://localhost:8080/api/v1/users/<user_id>/roles \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"role_id":"<role_id>"}'

# Revoke role
curl -X DELETE http://localhost:8080/api/v1/users/<user_id>/roles/<role_id> \
  -H "Authorization: Bearer <access_token>"
```

## Dev User

Після `make seed`:
- Email: `admin@skeleton.local`
- Password: `Admin1234!`
- Role: `super_admin`
