# Testing Strategy with PostgreSQL & Testcontainers

## Overview

This project uses **testcontainers** for integration tests with PostgreSQL, ensuring tests run against real database.

## Requirements

- **Docker** must be running
- **PostgreSQL 16** container will be spawned automatically

## Test Categories

### 1. Unit Tests (Fast, No Database)

Tests for domain logic, application layer, no dependencies:

```bash
make test-unit
# or
go test ./internal/*/domain/... ./internal/*/application/...
```

### 2. Integration Tests (PostgreSQL Required)

Tests for repositories, database integration:

```bash
make test-integration
# or
go test ./internal/*/infrastructure/persistence/... -p 1
```

## Writing Tests

### Unit Tests (Mock Repositories)

```go
// internal/identity/application/command/register_user_test.go
func TestRegisterUserHandler(t *testing.T) {
	// Use testify/mock for repository interfaces
	mockRepo := &MockUserRepository{}
	handler := NewRegisterUserHandler(mockRepo, ...)
	// test business logic
}
```

### Integration Tests (Testcontainers)

```go
// internal/identity/infrastructure/persistence/user_repository_test.go
func TestUserRepository_Save(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)
	
	repo := persistence.NewUserRepository(pool)
	ctx := context.Background()
	
	user := domain.NewUser("test@example.com", "password")
	err := repo.Save(ctx, user)
	require.NoError(t, err)
}
```

## CI/CD Configuration

### GitHub Actions

```yaml
- name: Run tests
  run: |
	go test ./internal/*/domain/... ./internal/*/application/... -v
	go test ./internal/*/infrastructure/persistence/... -v
```

### Docker-in-Docker (DinD)

For CI environments, use Docker-in-Docker or external PostgreSQL service.

## Test Utilities

### SetupPostgres

Creates PostgreSQL container with:

- PostgreSQL 16 Alpine
- Auto-configured connection pool
- Automatic cleanup after test

### DefaultSchema

Contains all tables required for testing:

- users, roles, permissions
- files, file_uploads, processings
- notifications, templates, preferences
- audit_records, tasks, schedules, dead_letters

## Performance

- Unit tests: < 5 seconds
- Integration tests: ~30 seconds (includes container startup)
- Parallel tests can be run with `-p 1` flag for isolation

## Best Practices

1. **Use unit tests for business logic** - faster, no database
2. **Use integration tests for SQL queries** - test real behavior
3. **One database per test package** - use `t.Parallel()` carefully
4. **Reset schema between tests** - testutil handles cleanup

