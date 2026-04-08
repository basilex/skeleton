# Bounded Contexts

## Overview

| Context | Responsibility | Publishes Events | Subscribes To | Status |
|---------|----------------|------------------|---------------|--------|
| `status` | Build info, health check | — | — | ✅ |
| `identity` | Users, roles, auth, RBAC | `identity.user_registered`, `identity.role_assigned`, `identity.role_revoked`, `identity.login`, `identity.logout` | — | ✅ |
| `audit` | Audit log, all system events | — | `identity.user_registered`, `identity.role_assigned`, `identity.role_revoked`, `identity.login`, `identity.logout` | ✅ |
| `notifications` | Email, SMS, push, in-app notifications | `notification.created`, `notification.sent`, `notification.delivered`, `notification.failed` | `identity.user_registered`, `identity.password_reset_requested`, `tasks.task_failed` | ✅ |
| `tasks` | Background jobs, scheduled tasks | `tasks.task_created`, `tasks.task_completed`, `tasks.task_failed` | — | ✅ |
| `files` | File uploads, storage, processing | `files.file_uploaded`, `files.file_deleted`, `files.processing_completed` | `files.file_uploaded` (for processing) | ✅ |

## Rules

1. **No cross-context imports** — context A never imports packages from context B
2. **Event-based communication** — data exchange between contexts only through `eventbus.Bus`
3. **Event naming** — `{context}.{event_name}` in snake_case
4. **Handler registration** — all event handlers are registered in `cmd/api/main.go`

## Context Details

### Status Context
- **Responsibility**: Build info, health checks
- **Aggregates**: BuildInfo
- **Events**: Does not publish or subscribe
- **Endpoints**: `/health`, `/build`

### Identity Context
- **Responsibility**: User registration, authentication, RBAC
- **Aggregates**: User, Role, Session
- **Events**: Publishes `identity.user_registered`, `identity.role_assigned`, `identity.role_revoked`, `identity.login`, `identity.logout`
- **Endpoints**: `/api/v1/auth/*`, `/api/v1/users/*`, `/api/v1/roles/*`

### Audit Context
- **Responsibility**: Audit log, event history
- **Aggregates**: Record
- **Events**: Subscribes to all identity events
- **Endpoints**: `/api/v1/audit/*`

### Notifications Context
- **Responsibility**: Email, SMS, push, in-app notifications
- **Aggregates**: Notification, NotificationTemplate, NotificationPreferences
- **Events**: Publishes `notification.created`, `notification.sent`, `notification.delivered`, `notification.failed`
- **Subscribes to**: `identity.user_registered` (welcome email), `identity.password_reset_requested` (reset email), `tasks.task_failed` (alert admins)
- **Endpoints**: `/api/v1/notifications/*`, `/api/v1/notifications/preferences/*`
- **Details**: [ADR-010: Notifications](../adr/ADR-010-notifications.md)
- **Uses**: Tasks context for async sending

### Tasks Context
- **Responsibility**: Background job processing, scheduled tasks
- **Aggregates**: Task, TaskSchedule, DeadLetterTask
- **Events**: Publishes `tasks.task_created`, `tasks.task_completed`, `tasks.task_failed`
- **Subscribes to**: —
- **Endpoints**: `/api/v1/tasks/*`, `/api/v1/tasks/schedules/*`, `/api/v1/tasks/dead-letters/*`
- **Details**: [ADR-011: Tasks/Jobs](../adr/ADR-011-tasks-jobs.md)
- **Used by**: Notifications, Files

### Files Context
- **Responsibility**: File uploads, storage, image processing
- **Aggregates**: File, FileUpload, FileProcessing
- **Events**: Publishes `files.file_uploaded`, `files.file_deleted`, `files.processing_completed`
- **Subscribes to**: `files.file_uploaded` (trigger processing)
- **Endpoints**: `/api/v1/files/*`
- **Status**: ✅ Implemented
- **Details**: [ADR-012: Files/Storage](../adr/ADR-012-files-storage.md)
- **Uses**: Tasks context for async processing

## Checklist: Adding a New Context

- [x] Create `internal/{context}/` with domain/application/infrastructure/ports
- [x] Define domain aggregates and value objects
- [x] Describe repository interfaces in domain
- [x] Implement command/query handlers
- [x] Implement infrastructure adapters
- [x] Create HTTP handlers in ports
- [x] Register routes in `cmd/api/routes.go`
- [x] Add events if cross-context communication is needed
- [x] Write tests (domain → application → infrastructure)
- [x] Update this documentation