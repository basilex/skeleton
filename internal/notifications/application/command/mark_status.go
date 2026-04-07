package command

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/notifications/domain"
)

type MarkSentCommand struct {
	NotificationID domain.NotificationID
}

type MarkSentHandler struct {
	notificationRepo domain.NotificationRepository
}

func NewMarkSentHandler(
	notificationRepo domain.NotificationRepository,
) *MarkSentHandler {
	return &MarkSentHandler{
		notificationRepo: notificationRepo,
	}
}

func (h *MarkSentHandler) Handle(ctx context.Context, cmd MarkSentCommand) error {
	notification, err := h.notificationRepo.GetByID(ctx, cmd.NotificationID)
	if err != nil {
		return fmt.Errorf("get notification: %w", err)
	}

	if err := notification.MarkSent(); err != nil {
		return fmt.Errorf("mark sent: %w", err)
	}

	if err := h.notificationRepo.Update(ctx, notification); err != nil {
		return fmt.Errorf("update notification: %w", err)
	}

	return nil
}

type MarkDeliveredCommand struct {
	NotificationID domain.NotificationID
	DeliveredAt    time.Time
}

type MarkDeliveredHandler struct {
	notificationRepo domain.NotificationRepository
}

func NewMarkDeliveredHandler(
	notificationRepo domain.NotificationRepository,
) *MarkDeliveredHandler {
	return &MarkDeliveredHandler{
		notificationRepo: notificationRepo,
	}
}

func (h *MarkDeliveredHandler) Handle(ctx context.Context, cmd MarkDeliveredCommand) error {
	notification, err := h.notificationRepo.GetByID(ctx, cmd.NotificationID)
	if err != nil {
		return fmt.Errorf("get notification: %w", err)
	}

	if err := notification.MarkDelivered(); err != nil {
		return fmt.Errorf("mark delivered: %w", err)
	}

	if err := h.notificationRepo.Update(ctx, notification); err != nil {
		return fmt.Errorf("update notification: %w", err)
	}

	return nil
}

type MarkFailedCommand struct {
	NotificationID domain.NotificationID
	Error          string
}

type MarkFailedHandler struct {
	notificationRepo domain.NotificationRepository
}

func NewMarkFailedHandler(
	notificationRepo domain.NotificationRepository,
) *MarkFailedHandler {
	return &MarkFailedHandler{
		notificationRepo: notificationRepo,
	}
}

func (h *MarkFailedHandler) Handle(ctx context.Context, cmd MarkFailedCommand) error {
	notification, err := h.notificationRepo.GetByID(ctx, cmd.NotificationID)
	if err != nil {
		return fmt.Errorf("get notification: %w", err)
	}

	notification.IncrementAttempts()

	if notification.CanRetry() {
		delay := notification.NextRetryDelay()
		if err := notification.ScheduleRetry(delay); err != nil {
			return fmt.Errorf("schedule retry: %w", err)
		}
	} else {
		if err := notification.MarkFailed(cmd.Error); err != nil {
			return fmt.Errorf("mark failed: %w", err)
		}
	}

	if err := h.notificationRepo.Update(ctx, notification); err != nil {
		return fmt.Errorf("update notification: %w", err)
	}

	return nil
}
