# RBAC Model

## Модель

```
User ──(has many)──► UserRole ──► Role ──(has many)──► RolePermission ──► Permission
```

## Permission Format

`resource:action` — наприклад `users:read`, `users:write`, `roles:manage`

## Wildcard Rules

| Pattern | Meaning |
|---------|---------|
| `*:*` | Повний доступ до всього |
| `users:*` | Всі дії на ресурсі `users` |
| `*:read` | Читання всіх ресурсів |

## Вбудовані ролі

| Role | Permissions |
|------|-------------|
| `super_admin` | `*:*` |
| `admin` | `users:read`, `users:write`, `roles:read`, `roles:manage` |
| `viewer` | `users:read` |

## Middleware Usage

```go
// Автентифікація
r.GET("/users",
    authMiddleware.Authenticate(),
    handler.ListUsers,
)

// Автентифікація + RBAC
r.GET("/users",
    authMiddleware.Authenticate(),
    rbacMiddleware.Require("users:read"),
    handler.ListUsers,
)

// Кілька permission (всі мають бути)
r.POST("/users",
    authMiddleware.Authenticate(),
    rbacMiddleware.Require("users:write"),
    handler.CreateUser,
)
```

## Як додати нову permission

1. Додати в seed: `scripts/seed.go`
2. Призначити ролі: `scripts/seed.go` → `seedRolePermissions`
3. Використати в middleware: `rbacMiddleware.Require("resource:action")`
