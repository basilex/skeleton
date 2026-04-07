# ADR-010: Notifications Bounded Context

## Status

Accepted

## Context

Every modern application requires a notification mechanism for:
- Welcome emails on registration
- Password reset emails
- Alert notifications (security, system)
- Marketing campaigns (optional)
- In-app notifications
- SMS for critical events (2FA, security alerts)
- Push notifications (mobile apps)

Currently, the Identity context publishes events (`identity.user_registered`, `identity.login`, etc.), but there is no mechanism to convert these events into user notifications.

## Decision

Create a separate **Notifications** bounded context with the following architecture:

### 1. Domain Layer

#### Aggregates

**Notification** - main aggregate:
```go
type Notification struct {
    id          NotificationID
    recipient   Recipient
    channel     Channel      // email, sms, push, in_app
    subject     string
    content     Content      // text, html, template
    status      Status       // pending, queued, sent, delivered, failed
    priority    Priority     // low, normal, high, critical
    scheduledAt *time.Time    // optional scheduling
    sentAt      *time.Time
    deliveredAt *time.Time
    failedAt    *time.Time
    failureReason string
    attempts    int
    maxAttempts int
    metadata    map[string]string
    createdAt   time.Time
    updatedAt   time.Time
}

type Recipient struct {
    userID  *domain.UserID  // authenticated user
    email   string          // external recipient
    phone   string          // SMS recipient
    deviceToken string      // push recipient
}

type Channel string
const (
    ChannelEmail  Channel = "email"
    ChannelSMS    Channel = "sms"
    ChannelPush   Channel = "push"
    ChannelInApp  Channel = "in_app"
)

type Status string
const (
    StatusPending    Status = "pending"
    StatusQueued     Status = "queued"
    StatusSending    Status = "sending"
    StatusSent       Status = "sent"
    StatusDelivered  Status = "delivered"
    StatusFailed     Status = "failed"
)

type Priority string
const (
    PriorityLow      Priority = "low"
    PriorityNormal   Priority = "normal"
    PriorityHigh     Priority = "high"
    PriorityCritical Priority = "critical"
)
```

**NotificationTemplate** - message templates:
```go
type NotificationTemplate struct {
    id          TemplateID
    name        string              // "welcome_email", "password_reset"
    channel     Channel
    subject     string              // template with {{.Subject}}
    body        string              // template with {{.Body}}
    htmlBody    string              // optional HTML version
    variables   []string            // required variables
    isActive    bool
    createdAt   time.Time
    updatedAt   time.Time
}
```

**NotificationPreferences** - user preferences:
```go
type NotificationPreferences struct {
    id        PreferencesID
    userID    domain.UserID
    channels  map[Channel]ChannelPreference
    createdAt time.Time
    updatedAt time.Time
}

type ChannelPreference struct {
    enabled   bool
    frequency Frequency  // immediate, daily, weekly
    quietHours *QuietHours
}

type QuietHours struct {
    start int  // 22 (10 PM)
    end   int  // 8  (8 AM)
    timezone string
}
```

#### Domain Events

```go
// Notification lifecycle events
type NotificationCreated struct { ... }
type NotificationQueued struct { ... }
type NotificationSent struct { ... }
type NotificationDelivered struct { ... }
type NotificationFailed struct { ... }

// Template events
type TemplateCreated struct { ... }
type TemplateUpdated struct { ... }
```

#### Repository Interfaces

```go
type NotificationRepository interface {
    Create(ctx context.Context, notification *Notification) error
    Update(ctx context.Context, notification *Notification) error
    GetByID(ctx context.Context, id NotificationID) (*Notification, error)
    GetByStatus(ctx context.Context, status Status, limit int) ([]*Notification, error)
    GetPendingByUser(ctx context.Context, userID domain.UserID) ([]*Notification, error)
}

type TemplateRepository interface {
    Create(ctx context.Context, template *NotificationTemplate) error
    Update(ctx context.Context, template *NotificationTemplate) error
    GetByID(ctx context.Context, id TemplateID) (*NotificationTemplate, error)
    GetByName(ctx context.Context, name string) (*NotificationTemplate, error)
    List(ctx context.Context, channel Channel) ([]*NotificationTemplate, error)
}

type PreferencesRepository interface {
    GetByUserID(ctx context.Context, userID domain.UserID) (*NotificationPreferences, error)
    Upsert(ctx context.Context, preferences *NotificationPreferences) error
}
```

### 2. Application Layer

#### Commands

```go
// Create notification from template
type CreateNotificationCommand struct {
    TemplateName string
    Recipient    Recipient
    Variables    map[string]string
    Priority     Priority
    ScheduledAt  *time.Time
}

type CreateNotificationHandler func(ctx context.Context, cmd CreateNotificationCommand) (NotificationID, error)

// Create notification with raw content
type SendNotificationCommand struct {
    Recipient    Recipient
    Channel      Channel
    Subject      string
    Content      Content
    Priority     Priority
    ScheduledAt  *time.Time
}

type SendNotificationHandler func(ctx context.Context, cmd SendNotificationCommand) (NotificationID, error)

// Process queued notifications (background worker)
type ProcessPendingNotificationsCommand struct {
    Limit int
}

type ProcessPendingNotificationsHandler func(ctx context.Context, cmd ProcessPendingNotificationsCommand) error

// Mark as delivered/failed
type MarkDeliveredCommand struct {
    NotificationID NotificationID
    DeliveredAt   time.Time
}

type MarkFailedCommand struct {
    NotificationID NotificationID
    Reason        string
    RetryAt       *time.Time
}
```

#### Queries

```go
type GetNotificationQuery struct {
    ID NotificationID
}

type GetNotificationHandler func(ctx context.Context, query GetNotificationQuery) (*Notification, error)

type ListNotificationsQuery struct {
    UserID    domain.UserID
    Status    *Status
    Channel   *Channel
    FromDate  *time.Time
    ToDate    *time.Time
    Limit     int
    Cursor    *string
}

type ListNotificationsHandler func(ctx context.Context, query ListNotificationsQuery) (*NotificationList, error)

type GetPreferencesQuery struct {
    UserID domain.UserID
}

type GetPreferencesHandler func(ctx context.Context, query GetPreferencesQuery) (*NotificationPreferences, error)

type UpdatePreferencesCommand struct {
    UserID   domain.UserID
    Channel  Channel
    Enabled  bool
}

type UpdatePreferencesHandler func(ctx context.Context, cmd UpdatePreferencesCommand) error
```

#### Event Handlers (Subscriptions)

```go
// Subscribe to identity events for automatic notifications
type IdentityEventHandler struct {
    createNotification CreateNotificationHandler
}

func (h *IdentityEventHandler) OnUserRegistered(ctx context.Context, event identity.UserRegistered) {
    // Create welcome email notification
    h.createNotification(ctx, CreateNotificationCommand{
        TemplateName: "welcome_email",
        Recipient:    Recipient{userID: &event.UserID},
        Variables: map[string]string{
            "Email": event.Email,
        },
        Priority: PriorityNormal,
    })
}

func (h *IdentityEventHandler) OnPasswordResetRequested(ctx context.Context, event identity.PasswordResetRequested) {
    // Create password reset email
    h.createNotification(ctx, CreateNotificationCommand{
        TemplateName: "password_reset",
        Recipient:    Recipient{email: event.Email},
        Variables: map[string]string{
            "ResetToken": event.Token,
        },
        Priority: PriorityHigh,
    })
}
```

### 3. Infrastructure Layer

#### Senders (Channel implementations)

```go
type EmailSender interface {
    Send(ctx context.Context, to, subject, textBody, htmlBody string) error
}

type SMSSender interface {
    Send(ctx context.Context, to, message string) error
}

type PushSender interface {
    Send(ctx context.Context, deviceToken, title, body string) error
}

type InAppSender interface {
    Send(ctx context.Context, userID domain.UserID, title, body string) error
}
```

**Implementations:**
- SMTP email sender (built-in Go)
- AWS SES email sender
- Twilio SMS sender
- Firebase/FCM push sender
- In-memory in-app sender (WebSocket/SSE)

#### Notification Worker

```go
type NotificationWorker struct {
    repo            NotificationRepository
    emailSender     EmailSender
    smsSender       SMSSender
    pushSender      PushSender
    inAppSender     InAppSender
    eventBus        eventbus.Bus
    pollInterval    time.Duration
    maxAttempts     int
    retryDelay      time.Duration
}

func (w *NotificationWorker) Start(ctx context.Context) error {
    ticker := time.NewTicker(w.pollInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            w.processPending(ctx)
        }
    }
}

func (w *NotificationWorker) processPending(ctx context.Context) {
    notifications, _ := w.repo.GetByStatus(ctx, StatusPending, 100)
    
    for _, n := range notifications {
        if err := w.send(ctx, n); err != nil {
            w.handleFailure(ctx, n, err)
        } else {
            w.handleSuccess(ctx, n)
        }
    }
}
```

#### Template Engine

```go
type TemplateEngine interface {
    Render(template string, variables map[string]string) (string, error)
}

type GoTemplateEngine struct{}

func (e *GoTemplateEngine) Render(template string, variables map[string]string) (string, error) {
    tmpl, err := template.New("notification").Parse(template)
    if err != nil {
        return "", err
    }
    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, variables); err != nil {
        return "", err
    }
    return buf.String(), nil
}
```

#### Persistence

SQLite tables:
```sql
CREATE TABLE notification_templates (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    channel TEXT NOT NULL,
    subject TEXT,
    body TEXT NOT NULL,
    html_body TEXT,
    variables TEXT, -- JSON array
    is_active BOOLEAN DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE notification_preferences (
    id TEXT PRIMARY KEY,
    user_id TEXT UNIQUE NOT NULL,
    preferences TEXT NOT NULL, -- JSON object
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE notifications (
    id TEXT PRIMARY KEY,
    user_id TEXT,
    email TEXT,
    phone TEXT,
    device_token TEXT,
    channel TEXT NOT NULL,
    subject TEXT,
    content TEXT NOT NULL,
    html_content TEXT,
    status TEXT NOT NULL,
    priority TEXT NOT NULL,
    scheduled_at TEXT,
    sent_at TEXT,
    delivered_at TEXT,
    failed_at TEXT,
    failure_reason TEXT,
    attempts INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 3,
    metadata TEXT, -- JSON object
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_user ON notifications(user_id);
CREATE INDEX idx_notifications_scheduled ON notifications(scheduled_at);
CREATE INDEX idx_notifications_created ON notifications(created_at);
```

### 4. Ports Layer

#### HTTP Handlers

```go
// GET /api/v1/notifications - list user notifications
// GET /api/v1/notifications/:id - get notification details
// POST /api/v1/notifications - create notification (admin only)
// GET /api/v1/notifications/preferences - get user preferences
// PUT /api/v1/notifications/preferences - update user preferences

// Admin endpoints (requires admin role)
// GET /api/v1/admin/notifications/templates - list templates
// POST /api/v1/admin/notifications/templates - create template
// PUT /api/v1/admin/notifications/templates/:id - update template
// GET /api/v1/admin/notifications/stats - notification statistics
```

## Integration with Event Bus

Notifications context subscribes to events from other contexts:

```go
// From Identity context
identity.user_registered         -> send welcome email
identity.password_reset_requested -> send password reset email
identity.login_from_new_device    -> send security alert

// From Audit context (optional)
audit.security_event              -> send security alert (if high severity)

// From Tasks context (future)
task.failed                       -> send alert to admins
task.completed                    -> send notification to user
```

## Retry Strategy

Exponential backoff with max attempts:
```
Attempt 1: immediate
Attempt 2: after 1 minute
Attempt 3: after 5 minutes
Attempt 4: after 15 minutes
Attempt 5: after 1 hour
Attempt 6: after 6 hours
Max attempts: 5-10 (configurable)
```

Failed notifications after max attempts:
- Move to dead letter queue
- Alert administrators
- Log failure reason

## Delivery Tracking

For channels that support delivery confirmation:
- **Email**: SMTP DSN, AWS SES events
- **SMS**: Twilio delivery status webhook
- **Push**: Firebase delivery receipts
- **In-App**: WebSocket acknowledgment

## Testing Strategy

### Unit Tests
- Domain model validation (90%+)
- Template rendering
- Business logic

### Integration Tests
- Repository operations
- Sender implementations (mock)
- Event handler subscriptions

### End-to-End Tests
- Full notification flow
- Worker processing
- Retry logic

## Deployment Considerations

### Development
- In-memory sender (logs to console)
- SQLite for persistence
- No background worker (synchronous sending)

### Production
- Real senders (SMTP/AWS SES, Twilio, Firebase)
- Background worker process
- Monitoring and alerting for failed notifications
- Dead letter queue processing

## Security Considerations

1. **Email Sanitization**: HTML emails sanitized to prevent XSS
2. **Rate Limiting**: Per-user and per-IP rate limits
3. **Unsubscribe Links**: Required for marketing emails (CAN-SPAM compliance)
4. **PII Protection**: Email addresses logged only when necessary
5. **Template Injection**: Variables sanitized before template rendering

## Performance Considerations

1. **Async Processing**: All notifications processed asynchronously by worker
2. **Batching**: Send notifications in batches for efficiency
3. **Queue Priority**: Critical notifications processed first
4. **Connection Pooling**: Reuse SMTP/API connections
5. **Caching**: Cache templates in memory
6. **Database Indexing**: Optimized queries with indexes

## Migration Plan

```sql
-- migrations/017_create_notification_templates.up.sql
CREATE TABLE notification_templates (...);

-- migrations/018_create_notification_preferences.up.sql
CREATE TABLE notification_preferences (...);

-- migrations/019_create_notifications.up.sql
CREATE TABLE notifications (...);
```

## Consequences

### Positive
- ✅ Clean separation of concerns
- ✅ Supports multiple notification channels
- ✅ Template-based notifications
- ✅ User preference management
- ✅ Retry and failure handling
- ✅ Delivery tracking
- ✅ Event-driven architecture (subscribes to other contexts)
- ✅ Production-ready from day one

### Negative
- ❌ Additional complexity (3 aggregates, repositories)
- ❌ Requires background worker process
- ❌ External dependencies for senders (SMTP, Twilio, etc.)

### Neutral
- Worker can be separate process or goroutine
- Can use Tasks context (future) for job scheduling
- Templates can be managed by admins via API

## Alternatives Considered

1. **Direct Sending**: Send notifications synchronously in handlers
   - ❌ Blocked on external APIs
   - ❌ No retry mechanism
   - ❌ Poor user experience on slow APIs

2. **SaaS Integration (SendGrid, Mailgun)**: Use external service exclusively
   - ✅ Less code to maintain
   - ❌ Vendor lock-in
   - ❌ Limited flexibility
   - ❌ Higher cost at scale

3. **Message Queue (Redis Pub/Sub, RabbitMQ)**: Use external queue
   - ✅ Better scalability
   - ✅ Persistent queue
   - ❌ Additional infrastructure dependency
   - ❌ Overkill for MVP (SQLite queue is sufficient)

## References

- [ADR-001: Hexagonal Architecture](ADR-001-hexagonal-architecture.md)
- [ADR-003: Event Bus](ADR-003-event-bus.md)
- [CAN-SPAM Act](https://www.ftc.gov/business-guidance/resources/can-spam-act-compliance-guide-business)