package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/tasks/domain"
)

type CreateScheduleCommand struct {
	Name     string
	TaskType domain.TaskType
	Payload  domain.TaskPayload
	Cron     string
	Timezone string
}

type CreateScheduleHandler struct {
	repo domain.ScheduleRepository
}

func NewCreateScheduleHandler(repo domain.ScheduleRepository) *CreateScheduleHandler {
	return &CreateScheduleHandler{repo: repo}
}

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

type DeleteScheduleCommand struct {
	ScheduleID domain.ScheduleID
}

type DeleteScheduleHandler struct {
	repo domain.ScheduleRepository
}

func NewDeleteScheduleHandler(repo domain.ScheduleRepository) *DeleteScheduleHandler {
	return &DeleteScheduleHandler{repo: repo}
}

func (h *DeleteScheduleHandler) Handle(ctx context.Context, cmd DeleteScheduleCommand) error {
	if err := h.repo.Delete(ctx, cmd.ScheduleID); err != nil {
		return fmt.Errorf("delete schedule: %w", err)
	}
	return nil
}

type ActivateScheduleCommand struct {
	ScheduleID domain.ScheduleID
}

type ActivateScheduleHandler struct {
	repo domain.ScheduleRepository
}

func NewActivateScheduleHandler(repo domain.ScheduleRepository) *ActivateScheduleHandler {
	return &ActivateScheduleHandler{repo: repo}
}

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

type DeactivateScheduleCommand struct {
	ScheduleID domain.ScheduleID
}

type DeactivateScheduleHandler struct {
	repo domain.ScheduleRepository
}

func NewDeactivateScheduleHandler(repo domain.ScheduleRepository) *DeactivateScheduleHandler {
	return &DeactivateScheduleHandler{repo: repo}
}

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
