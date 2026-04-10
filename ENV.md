# Environment Configuration

This document explains how environment variables are managed across the monorepo.

## Backend (Go)

### Architecture

Backend uses **environment-specific config files** located in `backend/configs/`:

```
backend/configs/
├── .env.dev          # Development (default)
├── .env.test         # Testing
├── .env.prod         # Production
├── .env.example      # Template with all options
└── .env.prod.example # Production template
```

### Loading Logic

```go
// pkg/config/config.go
env := os.Getenv("APP_ENV")  // default: "dev"
envFile := fmt.Sprintf("configs/.env.%s", env)
godotenv.Load(envFile)
```

**File selection:**
- `APP_ENV=dev` → `configs/.env.dev`
- `APP_ENV=test` → `configs/.env.test`
- `APP_ENV=prod` → `configs/.env.prod`

### Priority Order

1. **Environment variables** (highest priority) - override everything
2. **Config file** `configs/.env.{APP_ENV}` - loaded second
3. **Defaults in code** - fallback values

### Local Development

```bash
# Option 1: Using config file (default)
cd backend
go run ./cmd/api  # Uses configs/.env.dev

# Option 2: Override with env vars
DATABASE_URL=postgres://... go run ./cmd/api

# Option 3: Specify environment
APP_ENV=test go run ./cmd/api
```

### Docker Compose

Environment variables are passed directly in `docker-compose.yml`:

```yaml
backend:
  environment:
    APP_ENV: production
    DATABASE_URL: postgres://skeleton:skeleton@postgres:5432/skeleton?sslmode=disable
    # ... other vars
```

**No volume mounts** - all config is in docker-compose.yml for transparency.

---

## Frontend (Next.js)

### Architecture

Frontend uses **Next.js standard .env files**:

```
frontend/
├── .env.local        # Local overrides (gitignored)
├── .env.production    # Production build
└── .env.example       # Template with all options
```

### Loading Logic

Next.js automatically loads:
1. `.env.production` for `npm run build` / `npm start`
2. `.env.local` for local overrides (always)
3. `.env.development` for `npm run dev` (if exists)

### Docker

```yaml
frontend:
  environment:
    NEXT_PUBLIC_API_URL: http://localhost:8080
    NEXT_PUBLIC_APP_NAME: Skeleton CRM
```

---

## Quick Reference

| Environment | Backend File | Frontend File | Use Case |
|-------------|--------------|---------------|----------|
| Development | `configs/.env.dev` | `.env.local` | Local dev |
| Testing | `configs/.env.test` | `.env.test` | CI/CD |
| Production | `configs/.env.prod` | `.env.production` | Deploy |

---

## Common Tasks

### Add new environment variable

**Backend:**
1. Add to `backend/configs/.env.dev`
2. Add to `backend/configs/.env.prod`
3. Add to `docker-compose.yml` under `backend.environment`
4. Document in `backend/configs/.env.example`

**Frontend:**
1. Add to `frontend/.env.local`
2. Add to `docker-compose.yml` under `frontend.environment`
3. Document in `frontend/.env.example`

### Run with different environment

```bash
# Backend - development
cd backend
go run ./cmd/api

# Backend - test
APP_ENV=test go run ./cmd/api

# Docker Compose - production
docker-compose up

# Docker Compose - override
APP_ENV=test docker-compose up
```

---

## Security Notes

1. **Never commit** `.env.local`, `.env.production` with secrets
2. **Use `docker-compose.yml`** for production secrets (Docker secrets / Kubernetes secrets recommended)
3. **JWT keys** are committed (`backend/keys/`) - **CHANGE IN PRODUCTION**
4. **Database passwords** in docker-compose are for development only

---

## Files Summary

### Backend

| File | Description | Git |
|------|-------------|-----|
| `configs/.env.dev` | Development config | ✅ Tracked |
| `configs/.env.test` | Test config | ✅ Tracked |
| `configs/.env.prod` | Production config | ✅ Tracked (no secrets!) |
| `configs/.env.example` | Template | ✅ Tracked |
| `.env.example` | Root template | ✅ Tracked |

### Frontend

| File | Description | Git |
|------|-------------|-----|
| `.env.local` | Local overrides | ❌ Gitignored |
| `.env.production` | Production build | ✅ Tracked |
| `.env.example` | Template | ✅ Tracked |

---

## Troubleshooting

### Backend can't find config file

**Error:** `load env file configs/.env.dev: no such file or directory`

**Solution:**
- Ensure you're in `backend/` directory when running
- Or set `APP_ENV` to use different file
- Docker: verify `configs/` is copied (check Dockerfile)

### Database connection fails

**Error:** `failed to connect to postgres`

**Solution:**
- Check `DATABASE_URL` in correct `.env.*` file
- Docker: verify postgres container is running
- Local: verify `configs/.env.dev` has correct URL

### Frontend can't reach backend

**Error:** `Network error` or `CORS error`

**Solution:**
- Check `NEXT_PUBLIC_API_URL` in `.env.local` or docker-compose
- Docker: use `http://backend:8080` inside containers, `http://localhost:8080` from host
- CORS is configured to allow `*` origins in development