// Package command provides command handlers for modifying notification state.
// This package implements the command side of CQRS for notification-related operations,
// handling write requests that create and modify notification entities.
package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/notifications/domain"
)

// CreateTemplateCommand represents a command to create a new notification template.
type CreateTemplateCommand struct {
	Name      string
	Channel   domain.Channel
	Subject   string
	Body      string
	HTMLBody  string
	Variables []string
}

// CreateTemplateHandler handles commands to create notification templates.
// It validates the template data and persists it to the repository.
type CreateTemplateHandler struct {
	templateRepo domain.TemplateRepository
}

// NewCreateTemplateHandler creates a new CreateTemplateHandler with the required repository.
func NewCreateTemplateHandler(
	templateRepo domain.TemplateRepository,
) *CreateTemplateHandler {
	return &CreateTemplateHandler{
		templateRepo: templateRepo,
	}
}

// Handle executes the CreateTemplateCommand to create a new template.
// It constructs the template entity with optional HTML body and persists it.
func (h *CreateTemplateHandler) Handle(ctx context.Context, cmd CreateTemplateCommand) (domain.TemplateID, error) {
	opts := make([]domain.TemplateOption, 0)
	if cmd.HTMLBody != "" {
		opts = append(opts, domain.WithHTMLBody(cmd.HTMLBody))
	}

	template, err := domain.NewNotificationTemplate(
		cmd.Name,
		cmd.Channel,
		cmd.Subject,
		cmd.Body,
		cmd.Variables,
		opts...,
	)
	if err != nil {
		return "", fmt.Errorf("create template: %w", err)
	}

	if err := h.templateRepo.Create(ctx, template); err != nil {
		return "", fmt.Errorf("save template: %w", err)
	}

	return template.ID(), nil
}

// UpdateTemplateCommand represents a command to update an existing template.
type UpdateTemplateCommand struct {
	ID        domain.TemplateID
	Subject   string
	Body      string
	HTMLBody  string
	Variables []string
	IsActive  bool
}

// UpdateTemplateHandler handles commands to update notification templates.
// It retrieves the existing template, applies updates, and persists changes.
type UpdateTemplateHandler struct {
	templateRepo domain.TemplateRepository
}

// NewUpdateTemplateHandler creates a new UpdateTemplateHandler with the required repository.
func NewUpdateTemplateHandler(
	templateRepo domain.TemplateRepository,
) *UpdateTemplateHandler {
	return &UpdateTemplateHandler{
		templateRepo: templateRepo,
	}
}

// Handle executes the UpdateTemplateCommand to modify an existing template.
// It updates content, variables, and active status while preserving the template identity.
func (h *UpdateTemplateHandler) Handle(ctx context.Context, cmd UpdateTemplateCommand) error {
	template, err := h.templateRepo.GetByID(ctx, cmd.ID)
	if err != nil {
		return fmt.Errorf("get template: %w", err)
	}

	opts := make([]domain.TemplateOption, 0)
	if cmd.HTMLBody != "" {
		opts = append(opts, domain.WithHTMLBody(cmd.HTMLBody))
	}

	if err := template.Update(cmd.Subject, cmd.Body, cmd.Variables, opts...); err != nil {
		return fmt.Errorf("update template: %w", err)
	}

	if cmd.IsActive {
		template.Activate()
	} else {
		template.Deactivate()
	}

	if err := h.templateRepo.Update(ctx, template); err != nil {
		return fmt.Errorf("save template: %w", err)
	}

	return nil
}
