package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/notifications/domain"
)

type CreateTemplateCommand struct {
	Name      string
	Channel   domain.Channel
	Subject   string
	Body      string
	HTMLBody  string
	Variables []string
}

type CreateTemplateHandler struct {
	templateRepo domain.TemplateRepository
}

func NewCreateTemplateHandler(
	templateRepo domain.TemplateRepository,
) *CreateTemplateHandler {
	return &CreateTemplateHandler{
		templateRepo: templateRepo,
	}
}

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

type UpdateTemplateCommand struct {
	ID        domain.TemplateID
	Subject   string
	Body      string
	HTMLBody  string
	Variables []string
	IsActive  bool
}

type UpdateTemplateHandler struct {
	templateRepo domain.TemplateRepository
}

func NewUpdateTemplateHandler(
	templateRepo domain.TemplateRepository,
) *UpdateTemplateHandler {
	return &UpdateTemplateHandler{
		templateRepo: templateRepo,
	}
}

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
