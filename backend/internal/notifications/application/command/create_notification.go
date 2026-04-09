// Package command provides command handlers for modifying notification state.
// This package implements the command side of CQRS for notification-related operations,
// handling write requests that create and modify notification entities.
package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/notifications/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

// CreateNotificationCommand represents a command to create a new notification.
// It contains all the necessary information to create a notification directly
// without using a template.
type CreateNotificationCommand struct {
	Recipient   domain.Recipient
	Channel     domain.Channel
	Subject     string
	Content     domain.Content
	Priority    domain.Priority
	ScheduledAt *string
	MaxAttempts *int
	Metadata    map[string]string
}

// CreateNotificationHandler handles commands to create new notifications.
// It creates the notification entity, persists it, and publishes a creation event.
type CreateNotificationHandler struct {
	notificationRepo domain.NotificationRepository
	eventBus         eventbus.Bus
}

// NewCreateNotificationHandler creates a new CreateNotificationHandler with the required dependencies.
func NewCreateNotificationHandler(
	notificationRepo domain.NotificationRepository,
	eventBus eventbus.Bus,
) *CreateNotificationHandler {
	return &CreateNotificationHandler{
		notificationRepo: notificationRepo,
		eventBus:         eventBus,
	}
}

// Handle executes the CreateNotificationCommand to create and persist a new notification.
// It applies optional configuration, creates the entity, saves it, and publishes an event.
func (h *CreateNotificationHandler) Handle(ctx context.Context, cmd CreateNotificationCommand) (domain.NotificationID, error) {
	opts := make([]domain.NotificationOption, 0)

	if cmd.MaxAttempts != nil {
		opts = append(opts, domain.WithMaxAttempts(*cmd.MaxAttempts))
	}

	if cmd.Metadata != nil {
		opts = append(opts, domain.WithMetadata(cmd.Metadata))
	}

	notification, err := domain.NewNotification(
		cmd.Recipient,
		cmd.Channel,
		cmd.Subject,
		cmd.Content,
		cmd.Priority,
		opts...,
	)
	if err != nil {
		return domain.NotificationID{}, fmt.Errorf("create notification: %w", err)
	}

	if err := h.notificationRepo.Create(ctx, notification); err != nil {
		return domain.NotificationID{}, fmt.Errorf("save notification: %w", err)
	}

	event := domain.NewNotificationCreated(
		notification.ID(),
		notification.Recipient(),
		notification.Channel(),
		notification.Subject(),
		notification.Priority(),
	)
	h.eventBus.Publish(ctx, event)

	return notification.ID(), nil
}
