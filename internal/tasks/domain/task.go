package domain

import (
	"fmt"
	"time"

	"github.com/basilex/skeleton/pkg/uuid"
)

type TaskID string

func NewTaskID() TaskID {
	return TaskID(uuid.NewV7().String())
}

func ParseTaskID(s string) (TaskID, error) {
	if s == "" {
		return "", fmt.Errorf("task ID cannot be empty")
	}
	return TaskID(s), nil
}

func (id TaskID) String() string {
	return string(id)
}

type TaskType string

const (
	TaskTypeSendEmail         TaskType = "send_email"
	TaskTypeSendSMS           TaskType = "send_sms"
	TaskTypeSendPush          TaskType = "send_push"
	TaskTypeSendInApp         TaskType = "send_in_app"
	TaskTypeProcessFile       TaskType = "process_file"
	TaskTypeGenerateThumbnail TaskType = "generate_thumbnail"
	TaskTypeCleanupOldData    TaskType = "cleanup_old_data"
	TaskTypeGenerateReport    TaskType = "generate_report"
	TaskTypeSyncExternalAPI   TaskType = "sync_external_api"
	TaskTypeBatchOperation    TaskType = "batch_operation"
)

func (t TaskType) String() string {
	return string(t)
}

func ParseTaskType(s string) (TaskType, error) {
	switch s {
	case string(TaskTypeSendEmail):
		return TaskTypeSendEmail, nil
	case string(TaskTypeSendSMS):
		return TaskTypeSendSMS, nil
	case string(TaskTypeSendPush):
		return TaskTypeSendPush, nil
	case string(TaskTypeSendInApp):
		return TaskTypeSendInApp, nil
	case string(TaskTypeProcessFile):
		return TaskTypeProcessFile, nil
	case string(TaskTypeGenerateThumbnail):
		return TaskTypeGenerateThumbnail, nil
	case string(TaskTypeCleanupOldData):
		return TaskTypeCleanupOldData, nil
	case string(TaskTypeGenerateReport):
		return TaskTypeGenerateReport, nil
	case string(TaskTypeSyncExternalAPI):
		return TaskTypeSyncExternalAPI, nil
	case string(TaskTypeBatchOperation):
		return TaskTypeBatchOperation, nil
	default:
		return "", fmt.Errorf("invalid task type: %s", s)
	}
}

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusQueued    TaskStatus = "queued"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

func (s TaskStatus) String() string {
	return string(s)
}

func ParseTaskStatus(s string) (TaskStatus, error) {
	switch s {
	case string(TaskStatusPending):
		return TaskStatusPending, nil
	case string(TaskStatusQueued):
		return TaskStatusQueued, nil
	case string(TaskStatusRunning):
		return TaskStatusRunning, nil
	case string(TaskStatusCompleted):
		return TaskStatusCompleted, nil
	case string(TaskStatusFailed):
		return TaskStatusFailed, nil
	case string(TaskStatusCancelled):
		return TaskStatusCancelled, nil
	default:
		return "", fmt.Errorf("invalid task status: %s", s)
	}
}

type TaskPriority string

const (
	TaskPriorityLow      TaskPriority = "low"
	TaskPriorityNormal   TaskPriority = "normal"
	TaskPriorityHigh     TaskPriority = "high"
	TaskPriorityCritical TaskPriority = "critical"
)

func (p TaskPriority) String() string {
	return string(p)
}

func ParseTaskPriority(s string) (TaskPriority, error) {
	switch s {
	case string(TaskPriorityLow):
		return TaskPriorityLow, nil
	case string(TaskPriorityNormal):
		return TaskPriorityNormal, nil
	case string(TaskPriorityHigh):
		return TaskPriorityHigh, nil
	case string(TaskPriorityCritical):
		return TaskPriorityCritical, nil
	default:
		return "", fmt.Errorf("invalid task priority: %s", s)
	}
}

type TaskPayload map[string]interface{}

type TaskResult struct {
	Data       map[string]interface{}
	OutputPath string
	DurationMs int64
}

type TaskError struct {
	Code    string
	Message string
	Details map[string]string
}

type Task struct {
	id          TaskID
	taskType    TaskType
	status      TaskStatus
	priority    TaskPriority
	payload     TaskPayload
	result      *TaskResult
	taskError   *TaskError
	attempts    int
	maxAttempts int
	scheduledAt time.Time
	startedAt   *time.Time
	completedAt *time.Time
	createdAt   time.Time
	updatedAt   time.Time
}

func NewTask(
	taskType TaskType,
	payload TaskPayload,
	opts ...TaskOption,
) (*Task, error) {
	if taskType == "" {
		return nil, fmt.Errorf("task type cannot be empty")
	}

	now := time.Now()
	task := &Task{
		id:          NewTaskID(),
		taskType:    taskType,
		status:      TaskStatusPending,
		priority:    TaskPriorityNormal,
		payload:     payload,
		attempts:    0,
		maxAttempts: 5,
		scheduledAt: now,
		createdAt:   now,
		updatedAt:   now,
	}

	for _, opt := range opts {
		opt(task)
	}

	return task, nil
}

type TaskOption func(*Task)

func WithTaskPriority(priority TaskPriority) TaskOption {
	return func(t *Task) {
		t.priority = priority
	}
}

func WithTaskMaxAttempts(max int) TaskOption {
	return func(t *Task) {
		t.maxAttempts = max
	}
}

func WithTaskScheduledAt(scheduledAt time.Time) TaskOption {
	return func(t *Task) {
		t.scheduledAt = scheduledAt
	}
}

func (t *Task) ID() TaskID {
	return t.id
}

func (t *Task) Type() TaskType {
	return t.taskType
}

func (t *Task) Status() TaskStatus {
	return t.status
}

func (t *Task) Priority() TaskPriority {
	return t.priority
}

func (t *Task) Payload() TaskPayload {
	return t.payload
}

func (t *Task) Result() *TaskResult {
	return t.result
}

func (t *Task) Error() *TaskError {
	return t.taskError
}

func (t *Task) Attempts() int {
	return t.attempts
}

func (t *Task) MaxAttempts() int {
	return t.maxAttempts
}

func (t *Task) ScheduledAt() time.Time {
	return t.scheduledAt
}

func (t *Task) StartedAt() *time.Time {
	return t.startedAt
}

func (t *Task) CompletedAt() *time.Time {
	return t.completedAt
}

func (t *Task) CreatedAt() time.Time {
	return t.createdAt
}

func (t *Task) UpdatedAt() time.Time {
	return t.updatedAt
}

func (t *Task) Start() error {
	if t.status != TaskStatusPending && t.status != TaskStatusQueued {
		return fmt.Errorf("cannot start task with status %s", t.status)
	}

	t.status = TaskStatusRunning
	now := time.Now()
	t.startedAt = &now
	t.updatedAt = now
	return nil
}

func (t *Task) Complete(result *TaskResult) error {
	if t.status != TaskStatusRunning {
		return fmt.Errorf("cannot complete task with status %s", t.status)
	}

	t.status = TaskStatusCompleted
	t.result = result
	now := time.Now()
	t.completedAt = &now
	t.updatedAt = now
	return nil
}

func (t *Task) Fail(errorMsg string) error {
	t.status = TaskStatusFailed
	if t.taskError == nil {
		t.taskError = &TaskError{
			Message: errorMsg,
		}
	} else {
		t.taskError.Message = errorMsg
	}
	now := time.Now()
	t.completedAt = &now
	t.updatedAt = now
	return nil
}

func (t *Task) Cancel(reason string) error {
	if t.status == TaskStatusCompleted || t.status == TaskStatusFailed || t.status == TaskStatusCancelled {
		return fmt.Errorf("cannot cancel task with status %s", t.status)
	}

	t.status = TaskStatusCancelled
	if t.taskError == nil {
		t.taskError = &TaskError{
			Message: reason,
		}
	} else {
		t.taskError.Message = reason
	}
	t.updatedAt = time.Now()
	return nil
}

func (t *Task) Retry(delay time.Duration) error {
	if t.status != TaskStatusFailed {
		return fmt.Errorf("can only retry failed tasks")
	}

	t.status = TaskStatusPending
	t.attempts++
	t.scheduledAt = time.Now().Add(delay)
	t.taskError = nil
	t.updatedAt = time.Now()
	return nil
}

func (t *Task) CanRetry() bool {
	return t.attempts < t.maxAttempts && t.status == TaskStatusFailed
}

func (t *Task) NextRetryDelay() time.Duration {
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

func (t *Task) IncrementAttempts() {
	t.attempts++
	t.updatedAt = time.Now()
}

func (t *Task) IsPending() bool {
	return t.status == TaskStatusPending
}

func (t *Task) IsRunning() bool {
	return t.status == TaskStatusRunning
}

func (t *Task) IsCompleted() bool {
	return t.status == TaskStatusCompleted
}

func (t *Task) IsFailed() bool {
	return t.status == TaskStatusFailed
}

func (t *Task) IsCancelled() bool {
	return t.status == TaskStatusCancelled
}

func (t *Task) IsScheduled() bool {
	return t.scheduledAt.After(time.Now())
}
