// Package domain provides domain entities and value objects for the notifications module.
// This package contains the core business logic types for notification management,
// including preferences, templates, and domain events.
package domain

import (
	"context"
	"time"
)

// NotificationRepository defines the contract for notification persistence operations.
type NotificationRepository interface {
	Create(ctx context.Context, notification *Notification) error
	Update(ctx context.Context, notification *Notification) error
	GetByID(ctx context.Context, id NotificationID) (*Notification, error)
	GetByStatus(ctx context.Context, status Status, limit int) ([]*Notification, error)
	GetPendingByUser(ctx context.Context, userID string) ([]*Notification, error)
	GetScheduled(ctx context.Context, before time.Time, limit int) ([]*Notification, error)
	GetStalled(ctx context.Context, olderThan time.Duration, limit int) ([]*Notification, error)
	Delete(ctx context.Context, id NotificationID) error
	DeleteCompleted(ctx context.Context, olderThan time.Duration) (int64, error)
}

// TemplateRepository defines the contract for notification template persistence operations.
type TemplateRepository interface {
	Create(ctx context.Context, template *NotificationTemplate) error
	Update(ctx context.Context, template *NotificationTemplate) error
	GetByID(ctx context.Context, id TemplateID) (*NotificationTemplate, error)
	GetByName(ctx context.Context, name string) (*NotificationTemplate, error)
	List(ctx context.Context, channel *Channel) ([]*NotificationTemplate, error)
	Delete(ctx context.Context, id TemplateID) error
}

// PreferencesRepository defines the contract for user notification preference persistence operations.
type PreferencesRepository interface {
	GetByUserID(ctx context.Context, userID string) (*NotificationPreferences, error)
	Upsert(ctx context.Context, preferences *NotificationPreferences) error
	Delete(ctx context.Context, userID string) error
}
