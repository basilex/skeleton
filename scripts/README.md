# Database Scripts

## Migrate

Database migration tool for PostgreSQL.

### Usage

```bash
# Apply pending migrations
make migrate-up
# or
go run ./scripts/migrate -action=up

# Rollback last migration
make migrate-down
# or
go run ./scripts/migrate -action=down

# Check migration status
go run ./scripts/migrate -action=status
```

### Environment Variables

- `DATABASE_URL` - PostgreSQL connection string (default: `postgres://user:password@localhost:5432/skeleton?sslmode=disable`)

### Migration Files

Create migration files in `migrations/` directory:
- `{number}.up.sql` - Apply migration
- `{number}.down.sql` - Rollback migration

Example:
```
migrations/
  001_initial_schema.up.sql
  001_initial_schema.down.sql
  002_add_user_preferences.up.sql
  002_add_user_preferences.down.sql
```

## Seed

Database seeding tool for initial data.

### Usage

```bash
# Seed database with initial data
make seed
# or
go run ./scripts/seed
```

### Environment Variables

- `DATABASE_URL` - PostgreSQL connection string

### What it seeds

- **Roles**: super_admin, admin, viewer
- **Permissions**: users:read, users:write, roles:manage, etc.
- **Role Permissions**: Assign permissions to roles
- **Admin User**: admin@skeleton.local / Admin1234!

## Make Targets

Add to Makefile:

```makefile
migrate-up: ## Apply database migrations
	@echo "Running migrations..."
	go run ./scripts/migrate -action=up

migrate-down: ## Rollback last migration
	@echo "Rolling back migration..."
	go run ./scripts/migrate -action=down

migrate-status: ## Check migration status
	go run ./scripts/migrate -action=status

seed: ## Seed database with initial data
	@echo "Seeding database..."
	go run ./scripts/seed
```

## Docker Compose

For local development with PostgreSQL:

```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: skeleton
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

## Notes

- Both scripts use pure pgxpool (no ORM)
- PostgreSQL 16+ required
- Transactions ensure atomic migrations
- Idempotent seed operations (safe to run multiple times)
