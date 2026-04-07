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

// CreateFromTemplateCommand represents a command to create a notification from a template.
// It uses a template to generate notification content with variable substitution.
type CreateFromTemplateCommand struct {
	TemplateName string
	Recipient    domain.Recipient
	Variables    map[string]string
	Priority     domain.Priority
	ScheduledAt  *time.Time
}

// CreateFromTemplateHandler handles commands to create notifications from templates.
// It loads the template, validates variables, renders content, and creates the notification.
type CreateFromTemplateHandler struct {
	notificationRepo domain.NotificationRepository
	templateRepo     domain.TemplateRepository
}

// NewCreateFromTemplateHandler creates a new CreateFromTemplateHandler with the required repositories.
func NewCreateFromTemplateHandler(
	notificationRepo domain.NotificationRepository,
	templateRepo domain.TemplateRepository,
) *CreateFromTemplateHandler {
	return &CreateFromTemplateHandler{
		notificationRepo: notificationRepo,
		templateRepo:     templateRepo,
	}
}

// Handle executes the CreateFromTemplateCommand to create a notification from a template.
// It loads the template, validates variables, renders the content, and creates the notification.
func (h *CreateFromTemplateHandler) Handle(ctx context.Context, cmd CreateFromTemplateCommand) (domain.NotificationID, error) {
	template, err := h.templateRepo.GetByName(ctx, cmd.TemplateName)
	if err != nil {
		return "", fmt.Errorf("get template: %w", err)
	}

	if !template.IsActive() {
		return "", fmt.Errorf("template %s is not active", cmd.TemplateName)
	}

	if err := template.ValidateVariables(cmd.Variables); err != nil {
		return "", fmt.Errorf("validate variables: %w", err)
	}

	content := domain.Content{
		Text: renderTemplate(template.Body(), cmd.Variables),
	}
	if template.HTMLBody() != "" {
		content.HTML = renderTemplate(template.HTMLBody(), cmd.Variables)
	}

	subject := renderTemplate(template.Subject(), cmd.Variables)

	notification, err := domain.NewNotification(
		cmd.Recipient,
		template.Channel(),
		subject,
		content,
		cmd.Priority,
	)
	if err != nil {
		return "", fmt.Errorf("create notification: %w", err)
	}

	if cmd.ScheduledAt != nil {
		notification.ScheduleRetry(cmd.ScheduledAt.Sub(time.Now()))
	}

	if err := h.notificationRepo.Create(ctx, notification); err != nil {
		return "", fmt.Errorf("save notification: %w", err)
	}

	return notification.ID(), nil
}

// renderTemplate performs simple variable substitution on a template string.
// It replaces {{.VariableName}} placeholders with corresponding values.
func renderTemplate(template string, variables map[string]string) string {
	result := template
	for key, value := range variables {
		result = replaceAll(result, "{{."+key+"}}", value)
	}
	return result
}

// replaceAll replaces all occurrences of old with new in the string s.
// This is a custom implementation to avoid regex dependency.
func replaceAll(s, old, new string) string {
	result := ""
	for i := 0; i < len(s); {
		if i <= len(s)-len(old) && s[i:i+len(old)] == old {
			result += new
			i += len(old)
		} else {
			result += string(s[i])
			i++
		}
	}
	return result
}
