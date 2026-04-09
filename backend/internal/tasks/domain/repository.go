// Package domain provides domain entities and value objects for the tasks module.
// This package contains the core business logic types for task management,
// including scheduled tasks, dead letter tasks, and domain events.
package domain

import (
	"context"
	"time"
)

// TaskRepository defines the contract for task persistence operations.
type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	Update(ctx context.Context, task *Task) error
	GetByID(ctx context.Context, id TaskID) (*Task, error)

	GetPendingTasks(ctx context.Context, limit int) ([]*Task, error)
	GetTasksByStatus(ctx context.Context, status TaskStatus, limit int) ([]*Task, error)
	GetTasksByType(ctx context.Context, taskType TaskType, limit int) ([]*Task, error)
	GetScheduledTasks(ctx context.Context, before time.Time, limit int) ([]*Task, error)

	GetActiveTasks(ctx context.Context) ([]*Task, error)
	GetStalledTasks(ctx context.Context, olderThan time.Duration) ([]*Task, error)

	Delete(ctx context.Context, id TaskID) error
	DeleteCompletedTasks(ctx context.Context, olderThan time.Duration) (int64, error)
}

// ScheduleRepository defines the contract for task schedule persistence operations.
type ScheduleRepository interface {
	Create(ctx context.Context, schedule *TaskSchedule) error
	Update(ctx context.Context, schedule *TaskSchedule) error
	GetByID(ctx context.Context, id ScheduleID) (*TaskSchedule, error)
	GetByName(ctx context.Context, name string) (*TaskSchedule, error)
	GetActiveSchedules(ctx context.Context) ([]*TaskSchedule, error)
	List(ctx context.Context) ([]*TaskSchedule, error)
	Delete(ctx context.Context, id ScheduleID) error
}

// DeadLetterRepository defines the contract for dead letter task persistence operations.
type DeadLetterRepository interface {
	Create(ctx context.Context, task *DeadLetterTask) error
	GetByID(ctx context.Context, id DeadLetterID) (*DeadLetterTask, error)
	List(ctx context.Context, limit int, offset int) ([]*DeadLetterTask, error)
	MarkReviewed(ctx context.Context, id DeadLetterID, action DeadLetterAction, reviewedBy *string) error
	Delete(ctx context.Context, id DeadLetterID) error
}

// TaskHandler defines the contract for executing tasks.
type TaskHandler interface {
	Execute(ctx context.Context, payload TaskPayload) (*TaskResult, error)
}

// TaskHandlerRegistry defines the contract for registering and retrieving task handlers.
type TaskHandlerRegistry interface {
	Register(taskType TaskType, handler TaskHandler) error
	Get(taskType TaskType) (TaskHandler, error)
	Exists(taskType TaskType) bool
}
