# Bounded Contexts

## Overview

| Context | Відповідальність | Публікує події | Підписується на |
|---------|-----------------|----------------|-----------------|
| `status` | Build info, health check | — | — |
| `identity` | Users, roles, auth, RBAC | `identity.user_registered`, `identity.role_assigned`, `identity.role_revoked`, `identity.login`, `identity.logout` | — |
| `audit` | Audit log, all system events | — | `identity.user_registered`, `identity.role_assigned`, `identity.role_revoked`, `identity.login`, `identity.logout` |

## Rules

1. **No cross-context imports** — контекст A ніколи не імпортує пакети контексту B
2. **Event-based communication** — обмін даними між контекстами тільки через `eventbus.Bus`
3. **Event naming** — `{context}.{event_name}` у snake_case
4. **Handler registration** — всі event handler'и реєструються у `cmd/api/main.go`

## Checklist: Додавання нового контексту

- [x] Створити `internal/{context}/` з domain/application/infrastructure/ports
- [x] Визначити domain aggregates та value objects
- [x] Описати repository interfaces у domain
- [x] Реалізувати command/query handlers
- [x] Реалізувати infrastructure adapters
- [x] Створити HTTP handlers у ports
- [x] Зареєструвати routes у `cmd/api/routes.go`
- [x] Додати події якщо потрібна міжконтекстна комунікація
- [x] Написати тести (domain → application → infrastructure)
- [x] Оновити цю документацію
