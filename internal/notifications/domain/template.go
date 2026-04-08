// Package domain provides domain entities and value objects for the notifications module.
// This package contains the core business logic types for notification management,
// including preferences, templates, and domain events.
package domain

import (
	"fmt"
	"time"

	"github.com/basilex/skeleton/pkg/uuid"
)

// TemplateID is a unique identifier for a notification template.
type TemplateID uuid.UUID

// NewTemplateID generates a new unique TemplateID using UUID v7.
func NewTemplateID() TemplateID {
	return TemplateID(uuid.NewV7())
}

// ParseTemplateID validates and converts a string to TemplateID.
func ParseTemplateID(s string) (TemplateID, error) {
	if s == "" {
		return TemplateID{}, fmt.Errorf("template ID cannot be empty")
	}
	u, err := uuid.Parse(s)
	if err != nil {
		return TemplateID{}, fmt.Errorf("invalid template id: %w", err)
	}
	return TemplateID(u), nil
}

// String returns the string representation of TemplateID.
func (id TemplateID) String() string {
	return uuid.UUID(id).String()
}

// NotificationTemplate represents a reusable template for generating notifications.
type NotificationTemplate struct {
	id        TemplateID
	name      string
	channel   Channel
	subject   string
	body      string
	htmlBody  string
	variables []string
	isActive  bool
	createdAt time.Time
	updatedAt time.Time
}

// NewNotificationTemplate creates a new notification template with the provided details.
// Variables define the template placeholders that must be provided when rendering.
func NewNotificationTemplate(
	name string,
	channel Channel,
	subject string,
	body string,
	variables []string,
	opts ...TemplateOption,
) (*NotificationTemplate, error) {
	if name == "" {
		return nil, fmt.Errorf("template name cannot be empty")
	}
	if subject == "" {
		return nil, fmt.Errorf("template subject cannot be empty")
	}
	if body == "" {
		return nil, fmt.Errorf("template body cannot be empty")
	}

	now := time.Now()
	template := &NotificationTemplate{
		id:        NewTemplateID(),
		name:      name,
		channel:   channel,
		subject:   subject,
		body:      body,
		variables: variables,
		isActive:  true,
		createdAt: now,
		updatedAt: now,
	}

	for _, opt := range opts {
		opt(template)
	}

	return template, nil
}

// TemplateOption is a functional option for configuring a NotificationTemplate.
type TemplateOption func(*NotificationTemplate)

// WithHTMLBody sets the HTML body content for the template.
func WithHTMLBody(html string) TemplateOption {
	return func(t *NotificationTemplate) {
		t.htmlBody = html
	}
}

// ID returns the template's unique identifier.
func (t *NotificationTemplate) ID() TemplateID {
	return t.id
}

// Name returns the template's name.
func (t *NotificationTemplate) Name() string {
	return t.name
}

// Channel returns the notification channel this template is for.
func (t *NotificationTemplate) Channel() Channel {
	return t.channel
}

// Subject returns the template's subject line.
func (t *NotificationTemplate) Subject() string {
	return t.subject
}

// Body returns the template's plain text body.
func (t *NotificationTemplate) Body() string {
	return t.body
}

// HTMLBody returns the template's HTML body content.
func (t *NotificationTemplate) HTMLBody() string {
	return t.htmlBody
}

// Variables returns the list of required template variables.
func (t *NotificationTemplate) Variables() []string {
	return t.variables
}

// IsActive returns whether the template is active and available for use.
func (t *NotificationTemplate) IsActive() bool {
	return t.isActive
}

// CreatedAt returns when the template was created.
func (t *NotificationTemplate) CreatedAt() time.Time {
	return t.createdAt
}

// UpdatedAt returns when the template was last updated.
func (t *NotificationTemplate) UpdatedAt() time.Time {
	return t.updatedAt
}

// Activate enables the template for use.
func (t *NotificationTemplate) Activate() error {
	t.isActive = true
	t.updatedAt = time.Now()
	return nil
}

// Deactivate disables the template from use.
func (t *NotificationTemplate) Deactivate() error {
	t.isActive = false
	t.updatedAt = time.Now()
	return nil
}

// Update modifies the template's content and variables.
func (t *NotificationTemplate) Update(
	subject string,
	body string,
	variables []string,
	opts ...TemplateOption,
) error {
	if subject == "" {
		return fmt.Errorf("subject cannot be empty")
	}
	if body == "" {
		return fmt.Errorf("body cannot be empty")
	}

	t.subject = subject
	t.body = body
	t.variables = variables

	for _, opt := range opts {
		opt(t)
	}

	t.updatedAt = time.Now()
	return nil
}

// ValidateVariables checks that all required template variables are provided.
func (t *NotificationTemplate) ValidateVariables(provided map[string]string) error {
	for _, required := range t.variables {
		if _, ok := provided[required]; !ok {
			return fmt.Errorf("missing required variable: %s", required)
		}
	}
	return nil
}
