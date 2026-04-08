// Package eventhandler provides event handlers for processing domain events.
// This package contains handlers that react to events from other bounded contexts
// and trigger appropriate notification-related actions.
package eventhandler

import (
	"context"
	"fmt"

	identityDomain "github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/internal/notifications/application/command"
	notificationDomain "github.com/basilex/skeleton/internal/notifications/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

// IdentityEventHandler handles identity-related domain events and creates notifications.
// It listens for events like user registration and password reset requests to trigger
// the appropriate notification workflows.
type IdentityEventHandler struct {
	createNotificationHandler *command.CreateFromTemplateHandler
}

// NewIdentityEventHandler creates a new IdentityEventHandler with the required command handler.
func NewIdentityEventHandler(
	createNotificationHandler *command.CreateFromTemplateHandler,
) *IdentityEventHandler {
	return &IdentityEventHandler{
		createNotificationHandler: createNotificationHandler,
	}
}

// OnUserRegistered handles the user registered event and sends a welcome notification.
// It creates a welcome email notification from a template using the user's registered email.
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

// OnPasswordResetRequested handles the password reset requested event and sends a reset notification.
// It creates a password reset email notification from a template with the reset token.
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

// parseUserID converts a string to a UserID pointer, returning nil for empty strings.
func parseUserID(s string) *identityDomain.UserID {
	if s == "" {
		return nil
	}
	id, err := identityDomain.ParseUserID(s)
	if err != nil {
		return nil
	}
	return &id
}

// Register subscribes this handler to the relevant events on the event bus.
// It sets up subscriptions for user registration and password reset events.
func (h *IdentityEventHandler) Register(bus eventbus.Bus) {
	bus.Subscribe("identity.user_registered", h.OnUserRegistered)
	bus.Subscribe("identity.password_reset_requested", h.OnPasswordResetRequested)
}
