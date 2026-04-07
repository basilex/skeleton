// Package domain provides domain entities and value objects for the tasks module.
// This package contains the core business logic types for task management,
// including scheduled tasks, dead letter tasks, and domain events.
package domain

import (
	"time"
)

// TaskCreated is emitted when a new task is created.
type TaskCreated struct {
	taskID     TaskID
	taskType   TaskType
	payload    TaskPayload
	occurredAt time.Time
}

// NewTaskCreated creates a new TaskCreated event.
func NewTaskCreated(taskID TaskID, taskType TaskType, payload TaskPayload) TaskCreated {
	return TaskCreated{
		taskID:     taskID,
		taskType:   taskType,
		payload:    payload,
		occurredAt: time.Now(),
	}
}

// EventName returns the event name for TaskCreated.
func (e TaskCreated) EventName() string {
	return "tasks.task_created"
}

// OccurredAt returns when the TaskCreated event occurred.
func (e TaskCreated) OccurredAt() time.Time {
	return e.occurredAt
}

// TaskID returns the task's unique identifier.
func (e TaskCreated) TaskID() TaskID {
	return e.taskID
}

// TaskType returns the type of task.
func (e TaskCreated) TaskType() TaskType {
	return e.taskType
}

// Payload returns the task payload.
func (e TaskCreated) Payload() TaskPayload {
	return e.payload
}

// TaskStarted is emitted when a task begins execution.
type TaskStarted struct {
	taskID     TaskID
	startedAt  time.Time
	occurredAt time.Time
}

// NewTaskStarted creates a new TaskStarted event.
func NewTaskStarted(taskID TaskID) TaskStarted {
	now := time.Now()
	return TaskStarted{
		taskID:     taskID,
		startedAt:  now,
		occurredAt: now,
	}
}

// EventName returns the event name for TaskStarted.
func (e TaskStarted) EventName() string {
	return "tasks.task_started"
}

// OccurredAt returns when the TaskStarted event occurred.
func (e TaskStarted) OccurredAt() time.Time {
	return e.occurredAt
}

// TaskID returns the task's unique identifier.
func (e TaskStarted) TaskID() TaskID {
	return e.taskID
}

// StartedAt returns when the task execution started.
func (e TaskStarted) StartedAt() time.Time {
	return e.startedAt
}

// TaskCompleted is emitted when a task completes successfully.
type TaskCompleted struct {
	taskID      TaskID
	taskType    TaskType
	result      *TaskResult
	durationMs  int64
	completedAt time.Time
	occurredAt  time.Time
}

// NewTaskCompleted creates a new TaskCompleted event.
func NewTaskCompleted(taskID TaskID, taskType TaskType, result *TaskResult, durationMs int64) TaskCompleted {
	now := time.Now()
	return TaskCompleted{
		taskID:      taskID,
		taskType:    taskType,
		result:      result,
		durationMs:  durationMs,
		completedAt: now,
		occurredAt:  now,
	}
}

// EventName returns the event name for TaskCompleted.
func (e TaskCompleted) EventName() string {
	return "tasks.task_completed"
}

// OccurredAt returns when the TaskCompleted event occurred.
func (e TaskCompleted) OccurredAt() time.Time {
	return e.occurredAt
}

// TaskID returns the task's unique identifier.
func (e TaskCompleted) TaskID() TaskID {
	return e.taskID
}

// TaskType returns the type of task.
func (e TaskCompleted) TaskType() TaskType {
	return e.taskType
}

// Result returns the task execution result.
func (e TaskCompleted) Result() *TaskResult {
	return e.result
}

// DurationMs returns the task execution duration in milliseconds.
func (e TaskCompleted) DurationMs() int64 {
	return e.durationMs
}

// CompletedAt returns when the task completed.
func (e TaskCompleted) CompletedAt() time.Time {
	return e.completedAt
}

// TaskFailed is emitted when a task execution fails.
type TaskFailed struct {
	taskID      TaskID
	taskType    TaskType
	error       *TaskError
	attempts    int
	willRetry   bool
	nextRetryAt *time.Time
	failedAt    time.Time
	occurredAt  time.Time
}

// NewTaskFailed creates a new TaskFailed event.
func NewTaskFailed(taskID TaskID, taskType TaskType, taskError *TaskError, attempts int, willRetry bool, nextRetryAt *time.Time) TaskFailed {
	now := time.Now()
	return TaskFailed{
		taskID:      taskID,
		taskType:    taskType,
		error:       taskError,
		attempts:    attempts,
		willRetry:   willRetry,
		nextRetryAt: nextRetryAt,
		failedAt:    now,
		occurredAt:  now,
	}
}

// EventName returns the event name for TaskFailed.
func (e TaskFailed) EventName() string {
	return "tasks.task_failed"
}

// OccurredAt returns when the TaskFailed event occurred.
func (e TaskFailed) OccurredAt() time.Time {
	return e.occurredAt
}

// TaskID returns the task's unique identifier.
func (e TaskFailed) TaskID() TaskID {
	return e.taskID
}

// TaskType returns the type of task.
func (e TaskFailed) TaskType() TaskType {
	return e.taskType
}

// Error returns the task error details.
func (e TaskFailed) Error() *TaskError {
	return e.error
}

// Attempts returns the number of attempts made.
func (e TaskFailed) Attempts() int {
	return e.attempts
}

// WillRetry returns whether the task will be retried.
func (e TaskFailed) WillRetry() bool {
	return e.willRetry
}

// NextRetryAt returns the scheduled time for the next retry attempt.
func (e TaskFailed) NextRetryAt() *time.Time {
	return e.nextRetryAt
}

// FailedAt returns when the task failed.
func (e TaskFailed) FailedAt() time.Time {
	return e.failedAt
}

// TaskRetrying is emitted when a task is scheduled for retry.
type TaskRetrying struct {
	taskID      TaskID
	attempt     int
	nextRetryAt time.Time
	occurredAt  time.Time
}

// NewTaskRetrying creates a new TaskRetrying event.
func NewTaskRetrying(taskID TaskID, attempt int, nextRetryAt time.Time) TaskRetrying {
	return TaskRetrying{
		taskID:      taskID,
		attempt:     attempt,
		nextRetryAt: nextRetryAt,
		occurredAt:  time.Now(),
	}
}

// EventName returns the event name for TaskRetrying.
func (e TaskRetrying) EventName() string {
	return "tasks.task_retrying"
}

// OccurredAt returns when the TaskRetrying event occurred.
func (e TaskRetrying) OccurredAt() time.Time {
	return e.occurredAt
}

// TaskID returns the task's unique identifier.
func (e TaskRetrying) TaskID() TaskID {
	return e.taskID
}

// Attempt returns the attempt number.
func (e TaskRetrying) Attempt() int {
	return e.attempt
}

// NextRetryAt returns when the task will be retried.
func (e TaskRetrying) NextRetryAt() time.Time {
	return e.nextRetryAt
}

// TaskCancelled is emitted when a task is cancelled.
type TaskCancelled struct {
	taskID      TaskID
	reason      string
	cancelledAt time.Time
	occurredAt  time.Time
}

// NewTaskCancelled creates a new TaskCancelled event.
func NewTaskCancelled(taskID TaskID, reason string) TaskCancelled {
	now := time.Now()
	return TaskCancelled{
		taskID:      taskID,
		reason:      reason,
		cancelledAt: now,
		occurredAt:  now,
	}
}

// EventName returns the event name for TaskCancelled.
func (e TaskCancelled) EventName() string {
	return "tasks.task_cancelled"
}

// OccurredAt returns when the TaskCancelled event occurred.
func (e TaskCancelled) OccurredAt() time.Time {
	return e.occurredAt
}

// TaskID returns the task's unique identifier.
func (e TaskCancelled) TaskID() TaskID {
	return e.taskID
}

// Reason returns the cancellation reason.
func (e TaskCancelled) Reason() string {
	return e.reason
}

// CancelledAt returns when the task was cancelled.
func (e TaskCancelled) CancelledAt() time.Time {
	return e.cancelledAt
}
