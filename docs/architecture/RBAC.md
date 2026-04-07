# RBAC Model

## Model

```
User ──(has many)──► UserRole ──► Role ──(has many)──► RolePermission ──► Permission
```

## Permission Format

`resource:action` — for example `users:read`, `users:write`, `roles:manage`

## Wildcard Rules

| Pattern | Meaning |
|---------|---------|
| `*:*` | Full access to everything |
| `users:*` | All actions on`users` resource |
| `*:read` | Read access to all resources |

## Built-in Roles

| Role | Permissions |
|------|-------------|
| `super_admin` | `*:*` |
| `admin` | `users:read`, `users:write`, `roles:read`, `roles:manage` |
| `viewer` | `users:read` |

## Middleware Usage

```go
// Authentication
r.GET("/users",
    authMiddleware.Authenticate(),
    handler.ListUsers,
)

// Authentication + RBAC
r.GET("/users",
    authMiddleware.Authenticate(),
    rbacMiddleware.Require("users:read"),
    handler.ListUsers,
)

// Multiple permissions (all must be present)
r.POST("/users",
    authMiddleware.Authenticate(),
    rbacMiddleware.Require("users:write"),
    handler.CreateUser,
)
```

## How to Add a New Permission

1. Add to seed: `scripts/seed.go`
2. Assign to roles: `scripts/seed.go` → `seedRolePermissions`
3. Use in middleware: `rbacMiddleware.Require("resource:action")`