// Package domain provides domain entities and value objects for the tasks module.
// This package contains the core business logic types for task management,
// including scheduled tasks, dead letter tasks, and domain events.
package domain

import (
	"fmt"
	"time"

	"github.com/basilex/skeleton/pkg/uuid"
)

// TaskID is a unique identifier for a task.
type TaskID string

// NewTaskID generates a new unique TaskID using UUID v7.
func NewTaskID() TaskID {
	return TaskID(uuid.NewV7().String())
}

// ParseTaskID validates and converts a string to TaskID.
func ParseTaskID(s string) (TaskID, error) {
	if s == "" {
		return "", fmt.Errorf("task ID cannot be empty")
	}
	return TaskID(s), nil
}

// String returns the string representation of TaskID.
func (id TaskID) String() string {
	return string(id)
}

// TaskType represents the type of task to be executed.
type TaskType string

// Task type constants.
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

// String returns the string representation of TaskType.
func (t TaskType) String() string {
	return string(t)
}

// ParseTaskType converts a string to a TaskType value.
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

// TaskStatus represents the lifecycle state of a task.
type TaskStatus string

// Task status constants.
const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusQueued    TaskStatus = "queued"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// String returns the string representation of TaskStatus.
func (s TaskStatus) String() string {
	return string(s)
}

// ParseTaskStatus converts a string to a TaskStatus value.
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

// TaskPriority represents the urgency level of a task.
type TaskPriority string

// Task priority constants.
const (
	TaskPriorityLow      TaskPriority = "low"
	TaskPriorityNormal   TaskPriority = "normal"
	TaskPriorityHigh     TaskPriority = "high"
	TaskPriorityCritical TaskPriority = "critical"
)

// String returns the string representation of TaskPriority.
func (p TaskPriority) String() string {
	return string(p)
}

// ParseTaskPriority converts a string to a TaskPriority value.
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

// TaskPayload is the data payload for task execution.
type TaskPayload map[string]interface{}

// TaskResult contains the result of a successful task execution.
type TaskResult struct {
	Data       map[string]interface{}
	OutputPath string
	DurationMs int64
}

// TaskError contains error details for a failed task.
type TaskError struct {
	Code    string
	Message string
	Details map[string]string
}

// Task represents an asynchronous task in the domain.
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

// NewTask creates a new task with the provided type and payload.
// Optional configuration can be applied via TaskOption functions.
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

// TaskOption is a functional option for configuring a Task.
type TaskOption func(*Task)

// WithTaskPriority sets the task priority.
func WithTaskPriority(priority TaskPriority) TaskOption {
	return func(t *Task) {
		t.priority = priority
	}
}

// WithTaskMaxAttempts sets the maximum number of retry attempts.
func WithTaskMaxAttempts(max int) TaskOption {
	return func(t *Task) {
		t.maxAttempts = max
	}
}

// WithTaskScheduledAt sets the scheduled execution time.
func WithTaskScheduledAt(scheduledAt time.Time) TaskOption {
	return func(t *Task) {
		t.scheduledAt = scheduledAt
	}
}

// ID returns the task's unique identifier.
func (t *Task) ID() TaskID {
	return t.id
}

// Type returns the task type.
func (t *Task) Type() TaskType {
	return t.taskType
}

// Status returns the task's current status.
func (t *Task) Status() TaskStatus {
	return t.status
}

// Priority returns the task priority.
func (t *Task) Priority() TaskPriority {
	return t.priority
}

// Payload returns the task payload data.
func (t *Task) Payload() TaskPayload {
	return t.payload
}

// Result returns the task execution result, if completed successfully.
func (t *Task) Result() *TaskResult {
	return t.result
}

// Error returns the task error, if failed.
func (t *Task) Error() *TaskError {
	return t.taskError
}

// Attempts returns the number of execution attempts made.
func (t *Task) Attempts() int {
	return t.attempts
}

// MaxAttempts returns the maximum number of attempts allowed.
func (t *Task) MaxAttempts() int {
	return t.maxAttempts
}

// ScheduledAt returns when the task is scheduled to run.
func (t *Task) ScheduledAt() time.Time {
	return t.scheduledAt
}

// StartedAt returns when the task started execution, if applicable.
func (t *Task) StartedAt() *time.Time {
	return t.startedAt
}

// CompletedAt returns when the task completed, if applicable.
func (t *Task) CompletedAt() *time.Time {
	return t.completedAt
}

// CreatedAt returns when the task was created.
func (t *Task) CreatedAt() time.Time {
	return t.createdAt
}

// UpdatedAt returns when the task was last updated.
func (t *Task) UpdatedAt() time.Time {
	return t.updatedAt
}

// Start transitions the task to running status.
// Returns an error if the task is not in pending or queued status.
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

// Complete transitions the task to completed status with the result.
// Returns an error if the task is not in running status.
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

// Fail transitions the task to failed status with an error message.
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

// Cancel transitions the task to cancelled status with a reason.
// Returns an error if the task is already completed, failed, or cancelled.
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

// Retry resets the task to pending status with a delay before the next attempt.
// Returns an error if the task is not in failed status.
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

// CanRetry returns whether the task can be retried.
func (t *Task) CanRetry() bool {
	return t.attempts < t.maxAttempts && t.status == TaskStatusFailed
}

// NextRetryDelay calculates the delay before the next retry attempt using exponential backoff.
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

// IncrementAttempts increments the execution attempt counter.
func (t *Task) IncrementAttempts() {
	t.attempts++
	t.updatedAt = time.Now()
}

// IsPending returns whether the task is in pending status.
func (t *Task) IsPending() bool {
	return t.status == TaskStatusPending
}

// IsRunning returns whether the task is in running status.
func (t *Task) IsRunning() bool {
	return t.status == TaskStatusRunning
}

// IsCompleted returns whether the task is in completed status.
func (t *Task) IsCompleted() bool {
	return t.status == TaskStatusCompleted
}

// IsFailed returns whether the task is in failed status.
func (t *Task) IsFailed() bool {
	return t.status == TaskStatusFailed
}

// IsCancelled returns whether the task is in cancelled status.
func (t *Task) IsCancelled() bool {
	return t.status == TaskStatusCancelled
}

// IsScheduled returns whether the task is scheduled for future execution.
func (t *Task) IsScheduled() bool {
	return t.scheduledAt.After(time.Now())
}
