package command

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/tasks/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

type CreateTaskCommand struct {
	TaskType    domain.TaskType
	Payload     domain.TaskPayload
	Priority    domain.TaskPriority
	ScheduledAt *time.Time
	MaxAttempts *int
}

type CreateTaskHandler struct {
	repo     domain.TaskRepository
	eventBus eventbus.Bus
}

func NewCreateTaskHandler(repo domain.TaskRepository, eventBus eventbus.Bus) *CreateTaskHandler {
	return &CreateTaskHandler{
		repo:     repo,
		eventBus: eventBus,
	}
}

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

type CancelTaskCommand struct {
	TaskID domain.TaskID
	Reason string
}

type CancelTaskHandler struct {
	repo domain.TaskRepository
}

func NewCancelTaskHandler(repo domain.TaskRepository) *CancelTaskHandler {
	return &CancelTaskHandler{repo: repo}
}

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

type RetryDeadLetterCommand struct {
	DeadLetterID domain.DeadLetterID
}

type RetryDeadLetterHandler struct {
	deadLetterRepo domain.DeadLetterRepository
	taskRepo       domain.TaskRepository
}

func NewRetryDeadLetterHandler(
	deadLetterRepo domain.DeadLetterRepository,
	taskRepo domain.TaskRepository,
) *RetryDeadLetterHandler {
	return &RetryDeadLetterHandler{
		deadLetterRepo: deadLetterRepo,
		taskRepo:       taskRepo,
	}
}

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
