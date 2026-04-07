// Package command provides command handlers for modifying task and schedule state.
// This package implements the command side of CQRS for task-related operations,
// handling write requests that create, modify, and delete tasks and schedules.
package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/tasks/domain"
)

// CreateScheduleCommand represents a command to create a new task schedule.
// Schedules define recurring tasks that execute based on a cron expression.
type CreateScheduleCommand struct {
	Name     string
	TaskType domain.TaskType
	Payload  domain.TaskPayload
	Cron     string
	Timezone string
}

// CreateScheduleHandler handles commands to create new task schedules.
// It validates the cron expression and creates the schedule entity.
type CreateScheduleHandler struct {
	repo domain.ScheduleRepository
}

// NewCreateScheduleHandler creates a new CreateScheduleHandler with the required repository.
func NewCreateScheduleHandler(repo domain.ScheduleRepository) *CreateScheduleHandler {
	return &CreateScheduleHandler{repo: repo}
}

// Handle executes the CreateScheduleCommand to create and persist a new schedule.
func (h *CreateScheduleHandler) Handle(ctx context.Context, cmd CreateScheduleCommand) (domain.ScheduleID, error) {
	opts := make([]domain.ScheduleOption, 0)

	if cmd.Timezone != "" {
		opts = append(opts, domain.WithTimezone(cmd.Timezone))
	}

	schedule, err := domain.NewTaskSchedule(
		cmd.Name,
		cmd.TaskType,
		cmd.Cron,
		cmd.Payload,
		opts...,
	)
	if err != nil {
		return "", fmt.Errorf("create schedule: %w", err)
	}

	if err := h.repo.Create(ctx, schedule); err != nil {
		return "", fmt.Errorf("save schedule: %w", err)
	}

	return schedule.ID(), nil
}

// DeleteScheduleCommand represents a command to delete a schedule.
type DeleteScheduleCommand struct {
	ScheduleID domain.ScheduleID
}

// DeleteScheduleHandler handles commands to delete task schedules.
type DeleteScheduleHandler struct {
	repo domain.ScheduleRepository
}

// NewDeleteScheduleHandler creates a new DeleteScheduleHandler with the required repository.
func NewDeleteScheduleHandler(repo domain.ScheduleRepository) *DeleteScheduleHandler {
	return &DeleteScheduleHandler{repo: repo}
}

// Handle executes the DeleteScheduleCommand to remove a schedule.
func (h *DeleteScheduleHandler) Handle(ctx context.Context, cmd DeleteScheduleCommand) error {
	if err := h.repo.Delete(ctx, cmd.ScheduleID); err != nil {
		return fmt.Errorf("delete schedule: %w", err)
	}
	return nil
}

// ActivateScheduleCommand represents a command to activate a schedule.
// Active schedules generate tasks according to their cron expression.
type ActivateScheduleCommand struct {
	ScheduleID domain.ScheduleID
}

// ActivateScheduleHandler handles commands to activate task schedules.
type ActivateScheduleHandler struct {
	repo domain.ScheduleRepository
}

// NewActivateScheduleHandler creates a new ActivateScheduleHandler with the required repository.
func NewActivateScheduleHandler(repo domain.ScheduleRepository) *ActivateScheduleHandler {
	return &ActivateScheduleHandler{repo: repo}
}

// Handle executes the ActivateScheduleCommand to enable a schedule.
func (h *ActivateScheduleHandler) Handle(ctx context.Context, cmd ActivateScheduleCommand) error {
	schedule, err := h.repo.GetByID(ctx, cmd.ScheduleID)
	if err != nil {
		return fmt.Errorf("get schedule: %w", err)
	}

	schedule.Activate()

	if err := h.repo.Update(ctx, schedule); err != nil {
		return fmt.Errorf("update schedule: %w", err)
	}

	return nil
}

// DeactivateScheduleCommand represents a command to deactivate a schedule.
// Deactivated schedules stop generating new tasks.
type DeactivateScheduleCommand struct {
	ScheduleID domain.ScheduleID
}

// DeactivateScheduleHandler handles commands to deactivate task schedules.
type DeactivateScheduleHandler struct {
	repo domain.ScheduleRepository
}

// NewDeactivateScheduleHandler creates a new DeactivateScheduleHandler with the required repository.
func NewDeactivateScheduleHandler(repo domain.ScheduleRepository) *DeactivateScheduleHandler {
	return &DeactivateScheduleHandler{repo: repo}
}

// Handle executes the DeactivateScheduleCommand to disable a schedule.
func (h *DeactivateScheduleHandler) Handle(ctx context.Context, cmd DeactivateScheduleCommand) error {
	schedule, err := h.repo.GetByID(ctx, cmd.ScheduleID)
	if err != nil {
		return fmt.Errorf("get schedule: %w", err)
	}

	schedule.Deactivate()

	if err := h.repo.Update(ctx, schedule); err != nil {
		return fmt.Errorf("update schedule: %w", err)
	}

	return nil
}
