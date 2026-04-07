package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/notifications/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

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

type CreateNotificationHandler struct {
	notificationRepo domain.NotificationRepository
	eventBus         eventbus.Bus
}

func NewCreateNotificationHandler(
	notificationRepo domain.NotificationRepository,
	eventBus eventbus.Bus,
) *CreateNotificationHandler {
	return &CreateNotificationHandler{
		notificationRepo: notificationRepo,
		eventBus:         eventBus,
	}
}

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
		return "", fmt.Errorf("create notification: %w", err)
	}

	if err := h.notificationRepo.Create(ctx, notification); err != nil {
		return "", fmt.Errorf("save notification: %w", err)
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
