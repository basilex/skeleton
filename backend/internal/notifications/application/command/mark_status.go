// Package command provides command handlers for modifying notification state.
// This package implements the command side of CQRS for notification-related operations,
// handling write requests that create and modify notification entities.
package command

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/notifications/domain"
)

// MarkSentCommand represents a command to mark a notification as sent.
type MarkSentCommand struct {
	NotificationID domain.NotificationID
}

// MarkSentHandler handles commands to mark notifications as sent.
// It retrieves the notification, updates its status, and persists the change.
type MarkSentHandler struct {
	notificationRepo domain.NotificationRepository
}

// NewMarkSentHandler creates a new MarkSentHandler with the required repository.
func NewMarkSentHandler(
	notificationRepo domain.NotificationRepository,
) *MarkSentHandler {
	return &MarkSentHandler{
		notificationRepo: notificationRepo,
	}
}

// Handle executes the MarkSentCommand to update the notification status to sent.
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

// MarkDeliveredCommand represents a command to mark a notification as delivered.
type MarkDeliveredCommand struct {
	NotificationID domain.NotificationID
	DeliveredAt    time.Time
}

// MarkDeliveredHandler handles commands to mark notifications as delivered.
// It retrieves the notification, updates its status, and persists the change.
type MarkDeliveredHandler struct {
	notificationRepo domain.NotificationRepository
}

// NewMarkDeliveredHandler creates a new MarkDeliveredHandler with the required repository.
func NewMarkDeliveredHandler(
	notificationRepo domain.NotificationRepository,
) *MarkDeliveredHandler {
	return &MarkDeliveredHandler{
		notificationRepo: notificationRepo,
	}
}

// Handle executes the MarkDeliveredCommand to update the notification status to delivered.
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

// MarkFailedCommand represents a command to mark a notification as failed.
type MarkFailedCommand struct {
	NotificationID domain.NotificationID
	Error          string
}

// MarkFailedHandler handles commands to mark notifications as failed.
// It manages retry logic and escalation to dead letter queue when max attempts are exceeded.
type MarkFailedHandler struct {
	notificationRepo domain.NotificationRepository
}

// NewMarkFailedHandler creates a new MarkFailedHandler with the required repository.
func NewMarkFailedHandler(
	notificationRepo domain.NotificationRepository,
) *MarkFailedHandler {
	return &MarkFailedHandler{
		notificationRepo: notificationRepo,
	}
}

// Handle executes the MarkFailedCommand to handle a failed notification attempt.
// If retries are available, it schedules a retry; otherwise, it marks the notification as permanently failed.
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
