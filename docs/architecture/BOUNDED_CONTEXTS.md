# Bounded Contexts

## Overview

| Context | Відповідальність | Публікує події | Підписується на |
|---------|-----------------|----------------|-----------------|
| `status` | Build info, health check | — | — |
| `identity` | Users, roles, auth, RBAC | `identity.user_registered`, `identity.role_assigned`, `identity.role_revoked` | — |

## Rules

1. **No cross-context imports** — контекст A ніколи не імпортує пакети контексту B
2. **Event-based communication** — обмін даними між контекстами тільки через `eventbus.Bus`
3. **Event naming** — `{context}.{event_name}` у snake_case
4. **Handler registration** — всі event handler'и реєструються у `cmd/api/main.go`

## Checklist: Додавання нового контексту

- [ ] Створити `internal/{context}/` з domain/application/infrastructure/ports
- [ ] Визначити domain aggregates та value objects
- [ ] Описати repository interfaces у domain
- [ ] Реалізувати command/query handlers
- [ ] Реалізувати infrastructure adapters
- [ ] Створити HTTP handlers у ports
- [ ] Зареєструвати routes у `cmd/api/main.go`
- [ ] Додати події якщо потрібна міжконтекстна комунікація
- [ ] Написати тести (domain → application → infrastructure)
- [ ] Оновити цю документацію
