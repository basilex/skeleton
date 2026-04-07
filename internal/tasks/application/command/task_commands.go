// Package command provides command handlers for modifying task and schedule state.
// This package implements the command side of CQRS for task-related operations,
// handling write requests that create, modify, and delete tasks and schedules.
package command

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/tasks/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

// CreateTaskCommand represents a command to create a new task.
// Tasks are units of work that can be scheduled and processed asynchronously.
type CreateTaskCommand struct {
	TaskType    domain.TaskType
	Payload     domain.TaskPayload
	Priority    domain.TaskPriority
	ScheduledAt *time.Time
	MaxAttempts *int
}

// CreateTaskHandler handles commands to create new tasks.
// It creates the task entity, persists it, and publishes a creation event.
type CreateTaskHandler struct {
	repo     domain.TaskRepository
	eventBus eventbus.Bus
}

// NewCreateTaskHandler creates a new CreateTaskHandler with the required dependencies.
func NewCreateTaskHandler(repo domain.TaskRepository, eventBus eventbus.Bus) *CreateTaskHandler {
	return &CreateTaskHandler{
		repo:     repo,
		eventBus: eventBus,
	}
}

// Handle executes the CreateTaskCommand to create and persist a new task.
// It applies optional configuration, creates the entity, saves it, and publishes an event.
func (h *CreateTaskHandler) Handle(ctx context.Context, cmd CreateTaskCommand) (domain.TaskID, error) {
	opts := make([]domain.TaskOption, 0)

	if cmd.Priority != "" {
		opts = append(opts, domain.WithTaskPriority(cmd.Priority))
	}

	if cmd.MaxAttempts != nil {
		opts = append(opts, domain.WithTaskMaxAttempts(*cmd.MaxAttempts))
	}

	if cmd.ScheduledAt != nil {
		opts = append(opts, domain.WithTaskScheduledAt(*cmd.ScheduledAt))
	}

	task, err := domain.NewTask(cmd.TaskType, cmd.Payload, opts...)
	if err != nil {
		return "", fmt.Errorf("create task: %w", err)
	}

	if err := h.repo.Create(ctx, task); err != nil {
		return "", fmt.Errorf("save task: %w", err)
	}

	event := domain.NewTaskCreated(task.ID(), task.Type(), task.Payload())
	h.eventBus.Publish(ctx, event)

	return task.ID(), nil
}

// CancelTaskCommand represents a command to cancel a task.
// Cancelled tasks are stopped and marked with a cancellation reason.
type CancelTaskCommand struct {
	TaskID domain.TaskID
	Reason string
}

// CancelTaskHandler handles commands to cancel tasks.
// It validates the task can be cancelled and updates its status.
type CancelTaskHandler struct {
	repo domain.TaskRepository
}

// NewCancelTaskHandler creates a new CancelTaskHandler with the required repository.
func NewCancelTaskHandler(repo domain.TaskRepository) *CancelTaskHandler {
	return &CancelTaskHandler{repo: repo}
}

// Handle executes the CancelTaskCommand to cancel a task.
// It retrieves the task, cancels it with the provided reason, and persists the change.
func (h *CancelTaskHandler) Handle(ctx context.Context, cmd CancelTaskCommand) error {
	task, err := h.repo.GetByID(ctx, cmd.TaskID)
	if err != nil {
		return fmt.Errorf("get task: %w", err)
	}

	if err := task.Cancel(cmd.Reason); err != nil {
		return fmt.Errorf("cancel task: %w", err)
	}

	if err := h.repo.Update(ctx, task); err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	return nil
}

// RetryDeadLetterCommand represents a command to retry a dead letter task.
// Dead letter tasks can be retried after fixing underlying issues.
type RetryDeadLetterCommand struct {
	DeadLetterID domain.DeadLetterID
}

// RetryDeadLetterHandler handles commands to retry dead letter tasks.
// It retrieves the dead letter, schedules a retry, and removes it from the dead letter queue.
type RetryDeadLetterHandler struct {
	deadLetterRepo domain.DeadLetterRepository
	taskRepo       domain.TaskRepository
}

// NewRetryDeadLetterHandler creates a new RetryDeadLetterHandler with the required repositories.
func NewRetryDeadLetterHandler(
	deadLetterRepo domain.DeadLetterRepository,
	taskRepo domain.TaskRepository,
) *RetryDeadLetterHandler {
	return &RetryDeadLetterHandler{
		deadLetterRepo: deadLetterRepo,
		taskRepo:       taskRepo,
	}
}

// Handle executes the RetryDeadLetterCommand to retry a failed task.
// It retrieves the dead letter, schedules the task for retry, and removes it from the dead letter queue.
func (h *RetryDeadLetterHandler) Handle(ctx context.Context, cmd RetryDeadLetterCommand) error {
	deadLetter, err := h.deadLetterRepo.GetByID(ctx, cmd.DeadLetterID)
	if err != nil {
		return fmt.Errorf("get dead letter: %w", err)
	}

	if !deadLetter.CanRetry() {
		return domain.ErrTaskCannotRetry
	}

	delay := deadLetter.OriginalTask().NextRetryDelay()
	if err := deadLetter.OriginalTask().Retry(delay); err != nil {
		return fmt.Errorf("retry task: %w", err)
	}

	if err := h.taskRepo.Update(ctx, deadLetter.OriginalTask()); err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	if err := h.deadLetterRepo.Delete(ctx, cmd.DeadLetterID); err != nil {
		return fmt.Errorf("delete dead letter: %w", err)
	}

	return nil
}
