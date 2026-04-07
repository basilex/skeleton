# Bounded Contexts

## Overview

| Context | Відповідальність | Публікує події | Підписується на |
|---------|-----------------|----------------|-----------------|
| `status` | Build info, health check | — | — |
| `identity` | Users, roles, auth, RBAC | `identity.user_registered`, `identity.role_assigned`, `identity.role_revoked`, `identity.login`, `identity.logout` | — |
| `audit` | Audit log, all system events | — | `identity.user_registered`, `identity.role_assigned`, `identity.role_revoked`, `identity.login`, `identity.logout` |
| `notifications` | Email, SMS, push, in-app notifications | `notification.created`, `notification.sent`, `notification.delivered`, `notification.failed` | `identity.user_registered`, `identity.password_reset_requested`, `tasks.task_failed` |
| `tasks` | Background jobs, scheduled tasks | `tasks.task_created`, `tasks.task_completed`, `tasks.task_failed` | — |
| `files` | File uploads, storage, processing | `files.file_uploaded`, `files.file_deleted`, `files.processing_completed` | `files.file_uploaded` (for processing) |

## Rules

1. **No cross-context imports** — контекст A ніколи не імпортує пакети контексту B
2. **Event-based communication** — обмін даними між контекстами тільки через `eventbus.Bus`
3. **Event naming** — `{context}.{event_name}` у snake_case
4. **Handler registration** — всі event handler'и реєструються у `cmd/api/main.go`

## Context Details

### Status Context
- **Відповідальність**: Build info, health checks
- **Aggregates**: BuildInfo
- **Events**: Не публікує та не підписується
- **Endpoints**: `/health`, `/build`

### Identity Context
- **Відповідальність**: User registration, authentication, RBAC
- **Aggregates**: User, Role, Session
- **Events**: Публікує `identity.user_registered`, `identity.role_assigned`, `identity.role_revoked`, `identity.login`, `identity.logout`
- **Endpoints**: `/api/v1/auth/*`, `/api/v1/users/*`, `/api/v1/roles/*`

### Audit Context
- **Відповідальність**: Audit log, event history
- **Aggregates**: Record
- **Events**: Підписується на всі identity events
- **Endpoints**: `/api/v1/audit/*`
- **Details**: [ADR-010: Notifications](../adr/ADR-010-notifications.md)

### Notifications Context
- **Відповідальність**: Email, SMS, push, in-app notifications
- **Aggregates**: Notification, NotificationTemplate, NotificationPreferences
- **Events**: Публікує `notification.created`, `notification.sent`, `notification.delivered`, `notification.failed`
- **Subscribes to**: `identity.user_registered` (welcome email), `identity.password_reset_requested` (reset email), `tasks.task_failed` (alert admins)
- **Endpoints**: `/api/v1/notifications/*`, `/api/v1/notifications/preferences/*`
- **Details**: [ADR-010: Notifications](../adr/ADR-010-notifications.md)
- **Uses**: Tasks context for async sending

### Tasks Context
- **Відповідальність**: Background job processing, scheduled tasks
- **Aggregates**: Task, TaskSchedule, DeadLetterTask
- **Events**: Публікує `tasks.task_created`, `tasks.task_completed`, `tasks.task_failed`
- **Subscribes to**: —
- **Endpoints**: `/api/v1/tasks/*`, `/api/v1/tasks/schedules/*`, `/api/v1/tasks/dead-letters/*`
- **Details**: [ADR-011: Tasks/Jobs](../adr/ADR-011-tasks-jobs.md)
- **Used by**: Notifications, Files

### Files Context
- **Відповідальність**: File uploads, storage, image processing
- **Aggregates**: File, FileUpload, FileProcessing
- **Events**: Публікує `files.file_uploaded`, `files.file_deleted`, `files.processing_completed`
- **Subscribes to**: `files.file_uploaded` (trigger processing)
- **Endpoints**: `/api/v1/files/*`
- **Details**: [ADR-012: Files/Storage](../adr/ADR-012-files-storage.md)
- **Uses**: Tasks context for async processing

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
