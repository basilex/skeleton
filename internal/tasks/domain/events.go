package domain

import (
	"time"
)

type TaskCreated struct {
	taskID     TaskID
	taskType   TaskType
	payload    TaskPayload
	occurredAt time.Time
}

func NewTaskCreated(taskID TaskID, taskType TaskType, payload TaskPayload) TaskCreated {
	return TaskCreated{
		taskID:     taskID,
		taskType:   taskType,
		payload:    payload,
		occurredAt: time.Now(),
	}
}

func (e TaskCreated) EventName() string {
	return "tasks.task_created"
}

func (e TaskCreated) OccurredAt() time.Time {
	return e.occurredAt
}

func (e TaskCreated) TaskID() TaskID {
	return e.taskID
}

func (e TaskCreated) TaskType() TaskType {
	return e.taskType
}

func (e TaskCreated) Payload() TaskPayload {
	return e.payload
}

type TaskStarted struct {
	taskID     TaskID
	startedAt  time.Time
	occurredAt time.Time
}

func NewTaskStarted(taskID TaskID) TaskStarted {
	now := time.Now()
	return TaskStarted{
		taskID:     taskID,
		startedAt:  now,
		occurredAt: now,
	}
}

func (e TaskStarted) EventName() string {
	return "tasks.task_started"
}

func (e TaskStarted) OccurredAt() time.Time {
	return e.occurredAt
}

func (e TaskStarted) TaskID() TaskID {
	return e.taskID
}

func (e TaskStarted) StartedAt() time.Time {
	return e.startedAt
}

type TaskCompleted struct {
	taskID      TaskID
	taskType    TaskType
	result      *TaskResult
	durationMs  int64
	completedAt time.Time
	occurredAt  time.Time
}

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

func (e TaskCompleted) EventName() string {
	return "tasks.task_completed"
}

func (e TaskCompleted) OccurredAt() time.Time {
	return e.occurredAt
}

func (e TaskCompleted) TaskID() TaskID {
	return e.taskID
}

func (e TaskCompleted) TaskType() TaskType {
	return e.taskType
}

func (e TaskCompleted) Result() *TaskResult {
	return e.result
}

func (e TaskCompleted) DurationMs() int64 {
	return e.durationMs
}

func (e TaskCompleted) CompletedAt() time.Time {
	return e.completedAt
}

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

func (e TaskFailed) EventName() string {
	return "tasks.task_failed"
}

func (e TaskFailed) OccurredAt() time.Time {
	return e.occurredAt
}

func (e TaskFailed) TaskID() TaskID {
	return e.taskID
}

func (e TaskFailed) TaskType() TaskType {
	return e.taskType
}

func (e TaskFailed) Error() *TaskError {
	return e.error
}

func (e TaskFailed) Attempts() int {
	return e.attempts
}

func (e TaskFailed) WillRetry() bool {
	return e.willRetry
}

func (e TaskFailed) NextRetryAt() *time.Time {
	return e.nextRetryAt
}

func (e TaskFailed) FailedAt() time.Time {
	return e.failedAt
}

type TaskRetrying struct {
	taskID      TaskID
	attempt     int
	nextRetryAt time.Time
	occurredAt  time.Time
}

func NewTaskRetrying(taskID TaskID, attempt int, nextRetryAt time.Time) TaskRetrying {
	return TaskRetrying{
		taskID:      taskID,
		attempt:     attempt,
		nextRetryAt: nextRetryAt,
		occurredAt:  time.Now(),
	}
}

func (e TaskRetrying) EventName() string {
	return "tasks.task_retrying"
}

func (e TaskRetrying) OccurredAt() time.Time {
	return e.occurredAt
}

func (e TaskRetrying) TaskID() TaskID {
	return e.taskID
}

func (e TaskRetrying) Attempt() int {
	return e.attempt
}

func (e TaskRetrying) NextRetryAt() time.Time {
	return e.nextRetryAt
}

type TaskCancelled struct {
	taskID      TaskID
	reason      string
	cancelledAt time.Time
	occurredAt  time.Time
}

func NewTaskCancelled(taskID TaskID, reason string) TaskCancelled {
	now := time.Now()
	return TaskCancelled{
		taskID:      taskID,
		reason:      reason,
		cancelledAt: now,
		occurredAt:  now,
	}
}

func (e TaskCancelled) EventName() string {
	return "tasks.task_cancelled"
}

func (e TaskCancelled) OccurredAt() time.Time {
	return e.occurredAt
}

func (e TaskCancelled) TaskID() TaskID {
	return e.taskID
}

func (e TaskCancelled) Reason() string {
	return e.reason
}

func (e TaskCancelled) CancelledAt() time.Time {
	return e.cancelledAt
}
