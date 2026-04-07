# ADR-011: Tasks/Jobs Bounded Context

## Status

Accepted

## Context

Many operations in the application require asynchronous processing:
- Sending email/SMS notifications
- File processing (resize, thumbnail generation)
- Cleanup of old data
- Report generation
- Integrations with external APIs
- Batch operations (batch processing)

Currently there is no centralized system for background tasks. The Notifications context needs a mechanism to send messages asynchronously with retry logic.

## Decision

Create a separate **Tasks** bounded context as a universal system for background tasks.

### 1. Domain Layer

#### Aggregates

**Task** - main aggregate:
```go
type Task struct {
    id          TaskID
    type        TaskType
    status      TaskStatus
    priority    TaskPriority
    payload     TaskPayload      // JSON serializable data
    result      *TaskResult
    error       *TaskError
    attempts    int
    maxAttempts int
    scheduledAt time.Time
    startedAt   *time.Time
    completedAt *time.Time
    createdAt   time.Time
    updatedAt   time.Time
}

type TaskID string

type TaskType string
const (
    TaskTypeSendEmail          TaskType = "send_email"
    TaskTypeSendSMS            TaskType = "send_sms"
    TaskTypeSendPush           TaskType = "send_push"
    TaskTypeProcessFile        TaskType = "process_file"
    TaskTypeCleanupOldData     TaskType = "cleanup_old_data"
    TaskTypeGenerateReport     TaskType = "generate_report"
    TaskTypeSyncExternalAPI    TaskType = "sync_external_api"
    TaskTypeBatchOperation     TaskType = "batch_operation"
)

type TaskStatus string
const (
    TaskStatusPending    TaskStatus = "pending"     // waiting to be picked up
    TaskStatusQueued     TaskStatus = "queued"      // in queue, waiting for worker
    TaskStatusRunning    TaskStatus = "running"     // currently executing
    TaskStatusCompleted  TaskStatus = "completed"   // successfully finished
    TaskStatusFailed     TaskStatus = "failed"      // failed after max attempts
    TaskStatusCancelled  TaskStatus = "cancelled"   // manually cancelled
)

type TaskPriority string
const (
    TaskPriorityLow      TaskPriority = "low"
    TaskPriorityNormal   TaskPriority = "normal"
    TaskPriorityHigh     TaskPriority = "high"
    TaskPriorityCritical TaskPriority = "critical"
)

type TaskPayload map[string]interface{}  // JSON serializable

type TaskResult struct {
    Data       map[string]interface{}
    OutputPath string      // optional file output
    DurationMs int64
}

type TaskError struct {
    Code    string
    Message string
    Details map[string]string
}

// Business methods
func NewTask(taskType TaskType, payload TaskPayload, opts ...TaskOption) (*Task, error)
func (t *Task) Start() error
func (t *Task) Complete(result *TaskResult) error
func (t *Task) Fail(err error) error
func (t *Task) Retry(delay time.Duration) error
func (t *Task) Cancel(reason string) error
func (t *Task) CanRetry() bool
func (t *Task) NextRetryDelay() time.Duration  // exponential backoff
```

**TaskSchedule** - for periodic tasks:
```go
type TaskSchedule struct {
    id          ScheduleID
    name        string
    taskType    TaskType
    payload     TaskPayload
    cron        string           // cron expression
    timezone    string
    lastRunAt   *time.Time
    nextRunAt   *time.Time
    isActive    bool
    createdAt   time.Time
    updatedAt   time.Time
}

type ScheduleID string

// Business methods
func NewTaskSchedule(name string, taskType TaskType, cron string, payload TaskPayload) (*TaskSchedule, error)
func (s *TaskSchedule) UpdateNextRun() error
func (s *TaskSchedule) ShouldRun(now time.Time) bool
```

**DeadLetterQueue** - for tasks that failed to execute:
```go
type DeadLetterTask struct {
    id          DeadLetterID
    originalTask Task
    failedAt    time.Time
    reason      string
    reviewed    bool
    reviewedAt  *time.Time
    reviewedBy  *string
    action      DeadLetterAction  // retry, ignore, investigate
    createdAt   time.Time
}

type DeadLetterAction string
const (
    DeadLetterActionRetry      DeadLetterAction = "retry"
    DeadLetterActionIgnore     DeadLetterAction = "ignore"
    DeadLetterActionInvestigate DeadLetterAction = "investigate"
)

type DeadLetterID string
```

#### Domain Events

```go
// Task lifecycle events
type TaskCreated struct {
    TaskID  TaskID
    TaskType TaskType
    Payload TaskPayload
}

type TaskStarted struct {
    TaskID    TaskID
    StartedAt time.Time
}

type TaskCompleted struct {
    TaskID      TaskID
    TaskType    TaskType
    Result      TaskResult
    DurationMs  int64
    CompletedAt time.Time
}

type TaskFailed struct {
    TaskID      TaskID
    TaskType    TaskType
    Error       TaskError
    Attempts    int
    WillRetry   bool
    NextRetryAt *time.Time
}

type TaskRetrying struct {
    TaskID      TaskID
    Attempt     int
    NextRetryAt time.Time
}

type TaskCancelled struct {
    TaskID     TaskID
    Reason     string
    CancelledAt time.Time
}

// Schedule events
type ScheduleCreated struct { ... }
type ScheduleUpdated struct { ... }
type ScheduleTriggered struct {
    ScheduleID ScheduleID
    TaskID     TaskID
}
```

#### Repository Interfaces

```go
type TaskRepository interface {
    Create(ctx context.Context, task *Task) error
    Update(ctx context.Context, task *Task) error
    GetByID(ctx context.Context, id TaskID) (*Task, error)
    
    // Query methods
    GetPendingTasks(ctx context.Context, limit int) ([]*Task, error)
    GetTasksByStatus(ctx context.Context, status TaskStatus, limit int) ([]*Task, error)
    GetTasksByType(ctx context.Context, taskType TaskType, limit int) ([]*Task, error)
    
    // For scheduling
    GetScheduledTasks(ctx context.Context, before time.Time, limit int) ([]*Task, error)
    
    // For monitoring
    GetActiveTasks(ctx context.Context) ([]*Task, error)
    GetStalledTasks(ctx context.Context, olderThan time.Duration) ([]*Task, error)
    
    // Cleanup
    DeleteCompletedTasks(ctx context.Context, olderThan time.Duration) (int64, error)
}

type ScheduleRepository interface {
    Create(ctx context.Context, schedule *TaskSchedule) error
    Update(ctx context.Context, schedule *TaskSchedule) error
    GetByID(ctx context.Context, id ScheduleID) (*TaskSchedule, error)
    GetByName(ctx context.Context, name string) (*TaskSchedule, error)
    GetActiveSchedules(ctx context.Context) ([]*TaskSchedule, error)
    Delete(ctx context.Context, id ScheduleID) error
}

type DeadLetterRepository interface {
    Create(ctx context.Context, task *DeadLetterTask) error
    GetByID(ctx context.Context, id DeadLetterID) (*DeadLetterTask, error)
    List(ctx context.Context, limit int, offset int) ([]*DeadLetterTask, error)
    MarkReviewed(ctx context.Context, id DeadLetterID, action DeadLetterAction) error
    Delete(ctx context.Context, id DeadLetterID) error
}
```

#### Task Handler Registry

```go
type TaskHandler interface {
    Execute(ctx context.Context, payload TaskPayload) (*TaskResult, error)
}

type TaskHandlerRegistry interface {
    Register(taskType TaskType, handler TaskHandler) error
    Get(taskType TaskType) (TaskHandler, error)
    Exists(taskType TaskType) bool
}

// In-memory registry implementation
type InMemoryHandlerRegistry struct {
    handlers map[TaskType]TaskHandler
}

func (r *InMemoryHandlerRegistry) Register(taskType TaskType, handler TaskHandler) error {
    r.handlers[taskType] = handler
    return nil
}

func (r *InMemoryHandlerRegistry) Get(taskType TaskType) (TaskHandler, error) {
    handler, ok := r.handlers[taskType]
    if !ok {
        return nil, fmt.Errorf("no handler registered for task type: %s", taskType)
    }
    return handler, nil
}
```

### 2. Application Layer

#### Commands

```go
// Create task
type CreateTaskCommand struct {
    Type        TaskType
    Payload     TaskPayload
    Priority    TaskPriority
    ScheduledAt *time.Time
    MaxAttempts *int
}

type CreateTaskHandler func(ctx context.Context, cmd CreateTaskCommand) (TaskID, error)

// Cancel task
type CancelTaskCommand struct {
    TaskID TaskID
    Reason string
}

type CancelTaskHandler func(ctx context.Context, cmd CancelTaskCommand) error

// Retry dead letter
type RetryDeadLetterCommand struct {
    DeadLetterID DeadLetterID
}

type RetryDeadLetterHandler func(ctx context.Context, cmd RetryDeadLetterCommand) error

// Create schedule
type CreateScheduleCommand struct {
    Name     string
    TaskType TaskType
    Payload  TaskPayload
    Cron     string
    Timezone string
}

type CreateScheduleHandler func(ctx context.Context, cmd CreateScheduleCommand) (ScheduleID, error)

// Delete schedule
type DeleteScheduleCommand struct {
    ScheduleID ScheduleID
}

type DeleteScheduleHandler func(ctx context.Context, cmd DeleteScheduleCommand) error
```

#### Queries

```go
// Get task details
type GetTaskQuery struct {
    ID TaskID
}

type GetTaskHandler func(ctx context.Context, query GetTaskQuery) (*Task, error)

// List tasks
type ListTasksQuery struct {
    Type     *TaskType
    Status   *TaskStatus
    Priority *TaskPriority
    FromDate *time.Time
    ToDate   *time.Time
    Limit    int
    Cursor   *string
}

type ListTasksHandler func(ctx context.Context, query ListTasksQuery) (*TaskList, error)

// Get active tasks (running now)
type GetActiveTasksQuery struct{}

type GetActiveTasksHandler func(ctx context.Context, query GetActiveTasksQuery) ([]*Task, error)

// Get task statistics
type GetTaskStatsQuery struct {
    FromDate time.Time
    ToDate   time.Time
}

type TaskStats struct {
    TotalTasks     int64
    CompletedTasks int64
    FailedTasks    int64
    AvgDurationMs  int64
    ByType         map[TaskType]TypeStats
}

type GetTaskStatsHandler func(ctx context.Context, query GetTaskStatsQuery) (*TaskStats, error)

// List schedules
type ListSchedulesQuery struct{}

type ListSchedulesHandler func(ctx context.Context, query ListSchedulesQuery) ([]*TaskSchedule, error)

// List dead letters
type ListDeadLettersQuery struct {
    Limit  int
    Offset int
}

type ListDeadLettersHandler func(ctx context.Context, query ListDeadLettersQuery) ([]*DeadLetterTask, error)
```

### 3. Infrastructure Layer

#### Worker Implementation

```go
type TaskWorker struct {
    repo            TaskRepository
    handlerRegistry TaskHandlerRegistry
    eventBus        eventbus.Bus
    pollInterval    time.Duration
    maxConcurrent   int
    timeout         time.Duration
    
    // Concurrency control
    semaphore       chan struct{}
    stopChan        chan struct{}
}

func NewTaskWorker(
    repo TaskRepository,
    handlerRegistry TaskHandlerRegistry,
    eventBus eventbus.Bus,
    pollInterval time.Duration,
    maxConcurrent int,
    timeout time.Duration,
) *TaskWorker {
    return &TaskWorker{
        repo:            repo,
        handlerRegistry: handlerRegistry,
        eventBus:        eventBus,
        pollInterval:    pollInterval,
        maxConcurrent:   maxConcurrent,
        timeout:         timeout,
        semaphore:       make(chan struct{}, maxConcurrent),
        stopChan:        make(chan struct{}),
    }
}

func (w *TaskWorker) Start(ctx context.Context) error {
    ticker := time.NewTicker(w.pollInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-w.stopChan:
            return nil
        case <-ticker.C:
            w.processPendingTasks(ctx)
        }
    }
}

func (w *TaskWorker) Stop() {
    close(w.stopChan)
}

func (w *TaskWorker) processPendingTasks(ctx context.Context) {
    tasks, err := w.repo.GetPendingTasks(ctx, w.maxConcurrent)
    if err != nil {
        // log error
        return
    }
    
    for _, task := range tasks {
        // Acquire semaphore (blocks if maxConcurrent reached)
        w.semaphore <- struct{}{}
        
        go func(t *Task) {
            defer func() { <-w.semaphore }()
            
            w.executeTask(ctx, t)
        }(task)
    }
}

func (w *Worker) executeTask(ctx context.Context, task *Task) {
    // Start task
    task.Start()
    w.repo.Update(ctx, task)
    w.eventBus.Publish(ctx, TaskStarted{TaskID: task.ID, StartedAt: *task.StartedAt})
    
    // Get handler
    handler, err := w.handlerRegistry.Get(task.Type)
    if err != nil {
        w.handleTaskFailure(ctx, task, err)
        return
    }
    
    // Execute with timeout
    timeoutCtx, cancel := context.WithTimeout(ctx, w.timeout)
    defer cancel()
    
    result, err := handler.Execute(timeoutCtx, task.Payload)
    if err != nil {
        w.handleTaskFailure(ctx, task, err)
        return
    }
    
    // Complete task
    task.Complete(result)
    w.repo.Update(ctx, task)
    w.eventBus.Publish(ctx, TaskCompleted{
        TaskID: task.ID,
        TaskType: task.Type,
        Result: *result,
    })
}

func (w *Worker) handleTaskFailure(ctx context.Context, task *Task, err error) {
    task.Fail(err)
    
    if task.CanRetry() {
        delay := task.NextRetryDelay()
        task.Retry(delay)
        w.repo.Update(ctx, task)
        w.eventBus.Publish(ctx, TaskRetrying{
            TaskID: task.ID,
            Attempt: task.Attempts,
            NextRetryAt: *task.ScheduledAt,
        })
    } else {
        // Move to dead letter queue
        w.repo.Update(ctx, task)
        w.eventBus.Publish(ctx, TaskFailed{
            TaskID: task.ID,
            TaskType: task.Type,
            Error: TaskError{
                Message: err.Error(),
            },
        })
    }
}
```

#### Exponential Backoff

```go
func (t *Task) NextRetryDelay() time.Duration {
    // Exponential backoff: 1s, 5s, 15s, 1m, 5m, 15m, 1h, 6h
    delays := []time.Duration{
        1 * time.Second,
        5 * time.Second,
        15 * time.Second,
        1 * time.Minute,
        5 * time.Minute,
        15 * time.Minute,
        1 * time.Hour,
        6 * time.Hour,
    }
    
    if t.attempts >= len(delays) {
        return delays[len(delays)-1]
    }
    
    return delays[t.attempts]
}
```

#### Schedule Runner

```go
type ScheduleRunner struct {
    scheduleRepo ScheduleRepository
    taskRepo     TaskRepository
    pollInterval time.Duration
}

func (r *ScheduleRunner) Start(ctx context.Context) error {
    ticker := time.NewTicker(r.pollInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            r.processSchedules(ctx)
        }
    }
}

func (r *ScheduleRunner) processSchedules(ctx context.Context) {
    schedules, _ := r.scheduleRepo.GetActiveSchedules(ctx)
    
    for _, schedule := range schedules {
        if schedule.ShouldRun(time.Now()) {
            // Create new task from schedule
            task := NewTask(schedule.TaskType, schedule.Payload)
            r.taskRepo.Create(ctx, task)
            
            // Update schedule
            schedule.UpdateNextRun()
            r.scheduleRepo.Update(ctx, schedule)
        }
    }
}
```

#### Persistence

SQLite tables:
```sql
CREATE TABLE tasks (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    status TEXT NOT NULL,
    priority TEXT NOT NULL,
    payload TEXT NOT NULL, -- JSON
    result TEXT,            -- JSON
    error_code TEXT,
    error_message TEXT,
    error_details TEXT,     -- JSON
    attempts INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 5,
    scheduled_at TEXT NOT NULL,
    started_at TEXT,
    completed_at TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_type ON tasks(type);
CREATE INDEX idx_tasks_scheduled ON tasks(scheduled_at);
CREATE INDEX idx_tasks_priority ON tasks(priority);

CREATE TABLE task_schedules (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    task_type TEXT NOT NULL,
    payload TEXT NOT NULL,  -- JSON
    cron TEXT NOT NULL,
    timezone TEXT DEFAULT 'UTC',
    last_run_at TEXT,
    next_run_at TEXT,
    is_active BOOLEAN DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX idx_schedules_next_run ON task_schedules(next_run_at);

CREATE TABLE dead_letters (
    id TEXT PRIMARY KEY,
    original_task_id TEXT NOT NULL,
    original_task TEXT NOT NULL, -- JSON serialized Task
    failed_at TEXT NOT NULL,
    reason TEXT NOT NULL,
    reviewed BOOLEAN DEFAULT 0,
    reviewed_at TEXT,
    reviewed_by TEXT,
    action TEXT,
    created_at TEXT NOT NULL
);

CREATE INDEX idx_dead_letters_reviewed ON dead_letters(reviewed);
CREATE INDEX idx_dead_letters_failed_at ON dead_letters(failed_at);
```

### 4. Ports Layer

#### HTTP Handlers

```go
// Task management endpoints
// GET /api/v1/tasks - list tasks (admin)
// GET /api/v1/tasks/:id - get task details
// POST /api/v1/tasks - create task
// DELETE /api/v1/tasks/:id - cancel task

// Schedule management endpoints
// GET /api/v1/tasks/schedules - list schedules
// GET /api/v1/tasks/schedules/:id - get schedule details
// POST /api/v1/tasks/schedules - create schedule
// PUT /api/v1/tasks/schedules/:id - update schedule
// DELETE /api/v1/tasks/schedules/:id - delete schedule

// Dead letter management endpoints
// GET /api/v1/tasks/dead-letters - list dead letters (admin)
// GET /api/v1/tasks/dead-letters/:id - get dead letter details
// POST /api/v1/tasks/dead-letters/:id/retry - retry dead letter
// DELETE /api/v1/tasks/dead-letters/:id - ignore dead letter

// Monitoring endpoints
// GET /api/v1/tasks/stats - get task statistics (admin)
// GET /api/v1/tasks/active - get active tasks (admin)
```

### 5. Integration with Other Contexts

#### Task Handlers Registration

```go
// In wire.go for API initialization
func registerTaskHandlers(registry TaskHandlerRegistry, 
    notificationRepo NotificationRepository,
    emailSender EmailSender,
    // ... other dependencies
) {
    // Notification handlers
    registry.Register(TaskTypeSendEmail, &SendEmailHandler{
        repo: notificationRepo,
        sender: emailSender,
    })
    
    registry.Register(TaskTypeSendSMS, &SendSMSHandler{
        repo: notificationRepo,
        sender: smsSender,
    })
    
    // File processing handlers
    registry.Register(TaskTypeProcessFile, &ProcessFileHandler{
        fileRepo: fileRepository,
        storage: fileStorage,
    })
    
    // Cleanup handlers
    registry.Register(TaskTypeCleanupOldData, &CleanupHandler{
        auditRepo: auditRepository,
        taskRepo: taskRepository,
    })
}
```

#### Event-Driven Integration

```go
// Tasks context publishes events
tasks.task_created        -> other contexts can react
tasks.task_completed      -> notifications context sends alert
tasks.task_failed         -> notifications context sends alert to admins

// Tasks context subscribes to events
audit.security_event     -> tasks creates high-priority notification task
files.file_uploaded      -> tasks creates process_file task
```

## Task Handler Examples

### Send Email Handler

```go
type SendEmailHandler struct {
    notificationRepo NotificationRepository
    emailSender      EmailSender
}

func (h *SendEmailHandler) Execute(ctx context.Context, payload TaskPayload) (*TaskResult, error) {
    notificationID, ok := payload["notification_id"].(string)
    if !ok {
        return nil, fmt.Errorf("missing notification_id in payload")
    }
    
    notification, err := h.notificationRepo.GetByID(ctx, NotificationID(notificationID))
    if err != nil {
        return nil, fmt.Errorf("get notification: %w", err)
    }
    
    if err := h.emailSender.Send(ctx, notification.Recipient.Email, notification.Subject, notification.Content); err != nil {
        return nil, fmt.Errorf("send email: %w", err)
    }
    
    return &TaskResult{
        Data: map[string]interface{}{
            "notification_id": notificationID,
            "sent_at": time.Now(),
        },
    }, nil
}
```

### Cleanup Handler

```go
type CleanupHandler struct {
    auditRepo AuditRepository
    taskRepo  TaskRepository
}

func (h *CleanupHandler) Execute(ctx context.Context, payload TaskPayload) (*TaskResult, error) {
    olderThanDays, ok := payload["older_than_days"].(float64)
    if !ok {
        olderThanDays = 30
    }
    
    olderThan := time.Now().AddDate(0, 0, -int(olderThanDays))
    
    count, err := h.auditRepo.DeleteOldRecords(ctx, olderThan)
    if err != nil {
        return nil, fmt.Errorf("delete old records: %w", err)
    }
    
    return &TaskResult{
        Data: map[string]interface{}{
            "deleted_count": count,
        },
    }, nil
}
```

## Retry Strategy

Exponential backoff with increasing delays:
```
Attempt  Delay
1        1 second
2        5 seconds
3        15 seconds
4        1 minute
5        5 minutes
6        15 minutes
7        1 hour
8        6 hours
```

Max attempts configurable per task type:
- Critical tasks (notifications, payments): 10 attempts
- Normal tasks (file processing): 5 attempts
- Low priority (cleanup): 3 attempts

## Testing Strategy

### Unit Tests
- Task lifecycle methods
- Retry logic with backoff
- Schedule parsing and execution
- Handler registry

### Integration Tests
- Repository operations
- Worker execution
- Schedule runner
- Dead letter queue

### End-to-End Tests
- Full task flow (create → process → complete)
- Retry scenario
- Schedule execution
- Concurrent task execution

## Deployment Considerations

### Development
- Single worker goroutine
- In-memory task queue (SQLite)
- Synchronous execution (optional)
- Short poll interval (1 second)

### Production
- Multiple worker processes (horizontal scaling)
- Background process separate from API
- Longer poll interval (5-10 seconds)
- Monitoring and alerting
- Dead letter queue review process

### Worker Configuration

```go
type WorkerConfig struct {
    PollInterval    time.Duration  // How often to check for tasks
    MaxConcurrent   int            // Max concurrent tasks per worker
    TaskTimeout     time.Duration  // Max execution time per task
    RetryDelays     []time.Duration
    MaxAttempts     int
    HealthCheckPort int            // For monitoring
}
```

## Monitoring

### Metrics to track:
- Tasks created (by type)
- Tasks completed (by type)
- Tasks failed (by type)
- Average execution time
- Active workers
- Queue size (pending tasks)
- Dead letter queue size
- Retry rate

### Health checks:
- `/health` - worker is running
- `/ready` - worker can accept tasks
- `/metrics` - Prometheus metrics

## Security Considerations

1. **Payload Sanitization**: Validate JSON payload size and structure
2. **Access Control**: Admin-only endpoints for task management
3. **Rate Limiting**: Limit task creation per user
4. **Timeout Enforcement**: Prevent runaway tasks
5. **Sensitive Data**: Payload should not contain passwords/tokens (use references instead)

## Performance Considerations

1. **Connection Pooling**: Reuse database connections
2. **Concurrent Execution**: Multiple tasks in parallel
3. **Index Optimization**: Fast queries for pending tasks
4. **Batch Processing**: Process similar tasks together
5. **Memory Management**: Clean up completed tasks
6. **Worker Scaling**: Horizontal scaling for high load

## Migration Plan

```sql
-- migrations/014_create_tasks.up.sql
CREATE TABLE tasks (...);

-- migrations/015_create_schedules.up.sql
CREATE TABLE task_schedules (...);

-- migrations/016_create_dead_letters.up.sql
CREATE TABLE dead_letters (...);
```

## Consequences

### Positive
- ✅ Generic task system for all async operations
- ✅ Retry with exponential backoff
- ✅ Priority queues (critical tasks processed first)
- ✅ Dead letter queue for failed tasks
- ✅ Schedule support (cron-like)
- ✅ Monitoring and observability
- ✅ Horizontal scaling (multiple workers)
- ✅ Clean separation from business logic

### Negative
- ❌ Additional complexity (worker process, persistence)
- ❌ Need to register handlers for each task type
- ❌ Dead letter queue requires manual review

### Neutral
- Worker can be in-process (goroutine) or separate process
- SQLite queue sufficient for moderate load
- Can migrate to Redis/RabbitMQ for high load

## Alternatives Considered

1. **External Queue (Redis, RabbitMQ, SQS)**: Use external message queue
   - ✅ Better scalability
   - ✅ Persistent queue
   - ✅ Built-in retry logic
   - ❌ Additional infrastructure dependency
   - ❌ Higher operational complexity
   - ❌ SQLite queue is simpler for MVP

2. **Library-based (Asynq, Machinery)**: Use existing Go libraries
   - ✅ Less code to write
   - ✅ Battle-tested implementations
   - ❌ Less control over implementation
   - ❌ May not fit exactly with our architecture
   - ❌ Additional dependencies

3. **In-process Goroutines**: Use goroutines without persistence
   - ✅ Simplest approach
   - ❌ No persistence (tasks lost on restart)
   - ❌ No retry mechanism
   - ❌ No monitoring
   - ❌ Not production-ready

## References

- [ADR-001: Hexagonal Architecture](ADR-001-hexagonal-architecture.md)
- [ADR-003: Event Bus](ADR-003-event-bus.md)
- [ADR-010: Notifications Context](ADR-010-notifications.md)
- [Exponential Backoff Algorithm](https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/)
- [Cron Expression Syntax](https://crontab.guru/)