# Testing

## Running Tests

```bash
# All tests
make test

# With coverage
make test-cover

# Race detector
make test-race

# P0 critical tests only
make test-p0
```

## Coverage Targets

| Layer | Target |
|-------|--------|
| Domain | 90%+ |
| Application | 80%+ |
| HTTP Handlers | 70%+ |

## Test Conventions

- **Unit tests**: domain + application, no infrastructure
- **Integration tests**: persistence + HTTP, real PostgreSQL via testcontainers
- **Pattern**: Arrange → Act → Assert
- **Table-driven tests** where possible
- **`testify/require`** (not `assert`) for fatal failures
- **No global state** between tests

## Test Helpers

```go
// PostgreSQL + migrations via testcontainers
db := testutil.NewTestDB(t)

// Create test role
role := testutil.CreateTestRole(t, roleRepo, "users:read", "users:write")

// Create test user
user := testutil.CreateTestUser(t, userRepo)
```