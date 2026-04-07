# Notifications Context

## Overview

Notifications bounded context provides a multi-channel messaging system with support for templates, user preferences, and retry logic.

## Features

- ✅ **Multi-channel support**: Email, SMS, Push, In-App
- ✅ **Template-based messages**: Go templates with variables
- ✅ **User preferences**: Quiet hours, frequency settings per channel
- ✅ **Async processing**: Background worker with retry logic
- ✅ **Event-driven**: Auto-send on identity events
- ✅ **Priority queues**: Critical notifications processed first
- ✅ **Exponential backoff**: Retry strategy with configurable delays
- ✅ **Dead letter handling**: Failed notifications logging

## Architecture

### Domain Layer

```
notification.go         # Notification aggregate
notification_template.go # Template aggregate
preferences.go          # User preferences aggregate
├── events.go           # Domain events
├── repository.go       # Repository interfaces
└── errors.go           # Domain errors
```

### Application Layer

```
command/
  ├── create_notification.go      # Create notification command
  ├── create_from_template.go     # Create from template command
  ├── mark_status.go              # Mark sent/delivered/failed
  ├── update_preferences.go       # Update user preferences
  └── template.go                 # Template CRUD commands

query/
  ├── queries.go                  # Get/List notifications, templates, preferences

eventhandler/
  └── identity_handler.go         # Subscribe to identity events
```

### Infrastructure Layer

```
persistence/
  ├── notification_repository.go   # SQLite implementation
  ├── template_repository.go       # SQLite implementation
  └── preferences_repository.go    # SQLite implementation

sender/
  ├── interfaces.go               # EmailSender, SMSSender, etc.
  ├── console_sender.go           # Development logger
  └── smtp_sender.go              # Production SMTP

worker/
  └── notification_worker.go      # Background processor

template/
  └── engine.go                   # Go template engine
```

## Database Schema

### notifications

```sql
CREATE TABLE notifications (
    id TEXT PRIMARY KEY,
    user_id TEXT,
    email TEXT,
    phone TEXT,
    device_token TEXT,
    channel TEXT NOT NULL,          -- email, sms, push, in_app
    subject TEXT,
    content TEXT NOT NULL,
    html_content TEXT,
    status TEXT NOT NULL,            -- pending, queued, sending, sent, delivered, failed
    priority TEXT NOT NULL,          -- low, normal, high, critical
    scheduled_at TEXT,
    sent_at TEXT,
    delivered_at TEXT,
    failed_at TEXT,
    failure_reason TEXT,
    attempts INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 5,
    metadata TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
```

### notification_templates

```sql
CREATE TABLE notification_templates (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,      -- welcome_email, password_reset
    channel TEXT NOT NULL,
    subject TEXT,
    body TEXT NOT NULL,              -- Go template: Hello {{.Name}}
    html_body TEXT,
    variables TEXT,                  -- JSON array: ["Name", "Email"]
    is_active INTEGER DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
```

### notification_preferences

```sql
CREATE TABLE notification_preferences (
    id TEXT PRIMARY KEY,
    user_id TEXT UNIQUE NOT NULL,
    preferences TEXT NOT NULL,       -- JSON object
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
```

## API Endpoints

### User Endpoints (Authenticated)

```
GET    /api/v1/notifications                      # List user notifications
GET    /api/v1/notifications/:id                  # Get notification details
GET    /api/v1/notifications/preferences         # Get user preferences
PATCH  /api/v1/notifications/preferences         # Update preferences
```

### Admin Endpoints

```
POST   /api/v1/notifications                     # Create notification (admin)
GET    /api/v1/notifications/templates           # List templates (admin)
GET    /api/v1/notifications/templates/:id       # Get template (admin)
POST   /api/v1/notifications/templates           # Create template (admin)
PATCH  /api/v1/notifications/templates/:id       # Update template (admin)
```

## Usage Examples

### Create Notification (Direct)

```bash
POST /api/v1/notifications
Authorization: Bearer <token>

{
  "channel": "email",
  "email": "user@example.com",
  "subject": "Welcome!",
  "content": "Welcome to our platform!",
  "priority": "normal"
}
```

### Create from Template

Templates automatically fill content from variables:

```go
// Create template
template := domain.NewNotificationTemplate(
    "welcome_email",
    domain.ChannelEmail,
    "Welcome to {{.AppName}}",
    "Hello {{.Name}}, welcome to {{.AppName}}!",
    []string{"Name", "AppName"},
)

// Create notification from template (in code)
handler.Handle(ctx, command.CreateFromTemplateCommand{
    TemplateName: "welcome_email",
    Recipient: domain.Recipient{
        Email: "user@example.com",
    },
    Variables: map[string]string{
        "Name":    "John",
        "AppName": "Skeleton",
    },
    Priority: domain.PriorityNormal,
})
```

Result email:
```
Subject: Welcome to Skeleton
Body: Hello John, welcome to Skeleton!
```

### User Preferences

```bash
GET /api/v1/notifications/preferences

# Response
{
  "user_id": "user-123",
  "channels": {
    "email": {
      "enabled": true,
      "frequency": "immediate",
      "quiet_hours": {
        "start_hour": 22,
        "end_hour": 8,
        "timezone": "UTC"
      }
    },
    "sms": {
      "enabled": false
    },
    "push": {
      "enabled": true,
      "frequency": "daily"
    }
  }
}
```

## Event-Driven Integration

### Auto-send on User Registration

```go
// Identity context publishes event
bus.Publish(ctx, identity.UserRegistered{
    UserID: "user-123",
    Email:  "user@example.com",
})

// Notifications subscribes and sends welcome email
// (configured in eventhandler/identity_handler.go)
```

### Supported Events

- `identity.user_registered` → Welcome email
- `identity.password_reset_requested` → Password reset email
- `tasks.task_failed` → Alert to admins (future)

## Worker Configuration

Worker runs in background goroutine (started in `main.go`):

```go
go func() {
    if err := di.NotificationWorker.Start(context.Background()); err != nil {
        slog.Error("notification worker error", "error", err)
    }
}()
```

### Retry Strategy

Exponential backoff:

```
Attempt 1: immediate
Attempt 2: after 1 second
Attempt 3: after 5 seconds
Attempt 4: after 15 seconds
Attempt 5: after 1 minute
Attempt 6: after 5 minutes
...
Max attempts: 5-10 (configurable)
```

### Priority Processing

Notifications processed by priority order:
1. **Critical** - processed first
2. **High**
3. **Normal**
4. **Low**

## Development Mode

Console sender logs all notifications to stdout:

```
========== EMAIL ==========
To: user@example.com
Subject: Welcome!
----------------------------
Welcome to our platform!
============================
```

Replace with real senders in production:

```go
// In wire.go
emailSender := smtp.NewSMTPSender(cfg.SMTP)
smsSender := twilio.NewSMSSender(cfg.Twilio)
pushSender := firebase.NewPushSender(cfg.Firebase)

compositeSender := sender.NewCompositeSender(
    emailSender,
    smsSender, 
    pushSender,
    inAppSender,
)
```

## Testing

```bash
# Run all notification tests
go test ./internal/notifications/...

# Run specific test
go test ./internal/notifications/domain/...
```

## Monitoring

Worker logs all activity:

```
INFO  notification started
INFO  processing 5 pending notifications
INFO  successfully sent notification abc-123 via email
ERROR failed to send notification xyz-789: SMTP error
INFO  scheduled retry for notification xyz-789 in 5s (attempt 2/5)
```

## Future Enhancements

- [ ] Real-time tracking (WebSocket/SSE for in-app)
- [ ] Email templates with HTML
- [ ] Attachment support
- [ ] Scheduled notifications
- [ ] Batch processing
- [ ] Analytics dashboard

## References

- [ADR-010: Notifications Context](../adr/ADR-010-notifications.md)
- [Bounded Contexts](../architecture/BOUNDED_CONTEXTS.md)
- [Event Bus](../architecture/EVENT_BUS.md)