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

- **Unit tests**: domain + application, без інфраструктури
- **Integration tests**: persistence + HTTP, реальний SQLite `:memory:`
- **Pattern**: Arrange → Act → Assert
- **Table-driven tests** де можливо
- **`testify/require`** (не `assert`) для fatal failures
- **No global state** між тестами

## Test Helpers

```go
// SQLite :memory: + міграції
db := testutil.NewTestDB(t)

// Створити тестову роль
role := testutil.CreateTestRole(t, roleRepo, "users:read", "users:write")

// Створити тестового юзера
user := testutil.CreateTestUser(t, userRepo)
```
