package command

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/notifications/domain"
)

type CreateFromTemplateCommand struct {
	TemplateName string
	Recipient    domain.Recipient
	Variables    map[string]string
	Priority     domain.Priority
	ScheduledAt  *time.Time
}

type CreateFromTemplateHandler struct {
	notificationRepo domain.NotificationRepository
	templateRepo     domain.TemplateRepository
}

func NewCreateFromTemplateHandler(
	notificationRepo domain.NotificationRepository,
	templateRepo domain.TemplateRepository,
) *CreateFromTemplateHandler {
	return &CreateFromTemplateHandler{
		notificationRepo: notificationRepo,
		templateRepo:     templateRepo,
	}
}

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

func renderTemplate(template string, variables map[string]string) string {
	result := template
	for key, value := range variables {
		result = replaceAll(result, "{{."+key+"}}", value)
	}
	return result
}

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
