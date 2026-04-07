package eventhandler

import (
	"context"
	"fmt"

	identityDomain "github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/internal/notifications/application/command"
	notificationDomain "github.com/basilex/skeleton/internal/notifications/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

type IdentityEventHandler struct {
	createNotificationHandler *command.CreateFromTemplateHandler
}

func NewIdentityEventHandler(
	createNotificationHandler *command.CreateFromTemplateHandler,
) *IdentityEventHandler {
	return &IdentityEventHandler{
		createNotificationHandler: createNotificationHandler,
	}
}

func (h *IdentityEventHandler) OnUserRegistered(ctx context.Context, event eventbus.Event) error {
	data, ok := event.(interface {
		GetUserID() string
		GetEmail() string
	})
	if !ok {
		return fmt.Errorf("invalid event type")
	}

	cmd := command.CreateFromTemplateCommand{
		TemplateName: "welcome_email",
		Recipient: notificationDomain.Recipient{
			UserID: parseUserID(data.GetUserID()),
			Email:  data.GetEmail(),
		},
		Variables: map[string]string{
			"Email": data.GetEmail(),
		},
		Priority: notificationDomain.PriorityNormal,
	}

	_, err := h.createNotificationHandler.Handle(ctx, cmd)
	if err != nil {
		return fmt.Errorf("create welcome notification: %w", err)
	}

	return nil
}

func (h *IdentityEventHandler) OnPasswordResetRequested(ctx context.Context, event eventbus.Event) error {
	data, ok := event.(interface {
		GetUserID() string
		GetEmail() string
		GetToken() string
	})
	if !ok {
		return fmt.Errorf("invalid event type")
	}

	cmd := command.CreateFromTemplateCommand{
		TemplateName: "password_reset",
		Recipient: notificationDomain.Recipient{
			UserID: parseUserID(data.GetUserID()),
			Email:  data.GetEmail(),
		},
		Variables: map[string]string{
			"Email":      data.GetEmail(),
			"ResetToken": data.GetToken(),
		},
		Priority: notificationDomain.PriorityHigh,
	}

	_, err := h.createNotificationHandler.Handle(ctx, cmd)
	if err != nil {
		return fmt.Errorf("create password reset notification: %w", err)
	}

	return nil
}

func parseUserID(s string) *identityDomain.UserID {
	if s == "" {
		return nil
	}
	id := identityDomain.UserID(s)
	return &id
}
