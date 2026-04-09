// Package domain provides domain entities and value objects for the tasks module.
// This package contains the core business logic types for task management,
// including scheduled tasks, dead letter tasks, and domain events.
package domain

import (
	"fmt"
)

// Tasks domain error definitions.
var (
	ErrTaskNotFound          = fmt.Errorf("task not found")
	ErrScheduleNotFound      = fmt.Errorf("schedule not found")
	ErrDeadLetterNotFound    = fmt.Errorf("dead letter not found")
	ErrInvalidTaskType       = fmt.Errorf("invalid task type")
	ErrInvalidTaskStatus     = fmt.Errorf("invalid task status")
	ErrInvalidTaskPriority   = fmt.Errorf("invalid task priority")
	ErrTaskAlreadyCompleted  = fmt.Errorf("task already completed")
	ErrTaskAlreadyFailed     = fmt.Errorf("task already failed")
	ErrTaskAlreadyCancelled  = fmt.Errorf("task already cancelled")
	ErrTaskAlreadyRunning    = fmt.Errorf("task already running")
	ErrMaxAttemptsExceeded   = fmt.Errorf("max attempts exceeded")
	ErrTaskCannotStart       = fmt.Errorf("task cannot start")
	ErrTaskCannotComplete    = fmt.Errorf("task cannot complete")
	ErrTaskCannotFail        = fmt.Errorf("task cannot fail")
	ErrTaskCannotCancel      = fmt.Errorf("task cannot cancel")
	ErrTaskCannotRetry       = fmt.Errorf("task cannot retry")
	ErrHandlerNotRegistered  = fmt.Errorf("handler not registered")
	ErrInvalidCronExpression = fmt.Errorf("invalid cron expression")
	ErrScheduleNotActive     = fmt.Errorf("schedule not active")
)

// NewErrMaxAttemptsExceeded creates an error indicating maximum attempts have been exceeded.
func NewErrMaxAttemptsExceeded(attempts, maxAttempts int) error {
	return fmt.Errorf("%w: attempts=%d, max=%d", ErrMaxAttemptsExceeded, attempts, maxAttempts)
}

// NewErrHandlerNotRegistered creates an error indicating no handler is registered for a task type.
func NewErrHandlerNotRegistered(taskType TaskType) error {
	return fmt.Errorf("%w: type=%s", ErrHandlerNotRegistered, taskType)
}

// NewErrInvalidCronExpression creates an error indicating an invalid cron expression.
func NewErrInvalidCronExpression(cron string) error {
	return fmt.Errorf("%w: expression=%s", ErrInvalidCronExpression, cron)
}
