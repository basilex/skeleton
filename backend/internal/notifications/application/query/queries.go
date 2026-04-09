// Package query provides query handlers for reading notification data.
// This package implements the query side of CQRS for notification-related operations,
// handling read-only requests that return notification data transfer objects.
package query

import (
	"context"

	"github.com/basilex/skeleton/internal/notifications/domain"
)

// GetNotificationQuery represents a query to retrieve a single notification by ID.
type GetNotificationQuery struct {
	ID domain.NotificationID
}

// GetNotificationHandler handles queries to retrieve a single notification.
type GetNotificationHandler struct {
	repo domain.NotificationRepository
}

// NewGetNotificationHandler creates a new GetNotificationHandler with the required repository.
func NewGetNotificationHandler(repo domain.NotificationRepository) *GetNotificationHandler {
	return &GetNotificationHandler{repo: repo}
}

// Handle executes the GetNotificationQuery and returns the notification entity.
func (h *GetNotificationHandler) Handle(ctx context.Context, query GetNotificationQuery) (*domain.Notification, error) {
	return h.repo.GetByID(ctx, query.ID)
}

// ListNotificationsQuery represents a query to list notifications with optional filtering.
type ListNotificationsQuery struct {
	UserID   *string
	Status   *domain.Status
	Channel  *domain.Channel
	FromDate *string
	ToDate   *string
	Limit    int
	Cursor   *string
}

// ListNotificationsHandler handles queries to retrieve a list of notifications.
type ListNotificationsHandler struct {
	repo domain.NotificationRepository
}

// NewListNotificationsHandler creates a new ListNotificationsHandler with the required repository.
func NewListNotificationsHandler(repo domain.NotificationRepository) *ListNotificationsHandler {
	return &ListNotificationsHandler{repo: repo}
}

// Handle executes the ListNotificationsQuery and returns matching notifications.
// If status is specified, it filters by status; otherwise returns pending notifications.
func (h *ListNotificationsHandler) Handle(ctx context.Context, query ListNotificationsQuery) ([]*domain.Notification, error) {
	if query.Limit <= 0 {
		query.Limit = 100
	}

	if query.Status != nil {
		return h.repo.GetByStatus(ctx, *query.Status, query.Limit)
	}

	return h.repo.GetByStatus(ctx, domain.StatusPending, query.Limit)
}

// GetPreferencesQuery represents a query to retrieve notification preferences for a user.
type GetPreferencesQuery struct {
	UserID string
}

// GetPreferencesHandler handles queries to retrieve user notification preferences.
type GetPreferencesHandler struct {
	repo domain.PreferencesRepository
}

// NewGetPreferencesHandler creates a new GetPreferencesHandler with the required repository.
func NewGetPreferencesHandler(repo domain.PreferencesRepository) *GetPreferencesHandler {
	return &GetPreferencesHandler{repo: repo}
}

// Handle executes the GetPreferencesQuery and returns the user's notification preferences.
func (h *GetPreferencesHandler) Handle(ctx context.Context, query GetPreferencesQuery) (*domain.NotificationPreferences, error) {
	return h.repo.GetByUserID(ctx, query.UserID)
}

// GetTemplateQuery represents a query to retrieve a notification template by ID or name.
type GetTemplateQuery struct {
	ID   *domain.TemplateID
	Name *string
}

// GetTemplateHandler handles queries to retrieve a notification template.
type GetTemplateHandler struct {
	repo domain.TemplateRepository
}

// NewGetTemplateHandler creates a new GetTemplateHandler with the required repository.
func NewGetTemplateHandler(repo domain.TemplateRepository) *GetTemplateHandler {
	return &GetTemplateHandler{repo: repo}
}

// Handle executes the GetTemplateQuery and returns the template.
// It looks up by ID if provided, otherwise by name.
func (h *GetTemplateHandler) Handle(ctx context.Context, query GetTemplateQuery) (*domain.NotificationTemplate, error) {
	if query.ID != nil {
		return h.repo.GetByID(ctx, *query.ID)
	}
	if query.Name != nil {
		return h.repo.GetByName(ctx, *query.Name)
	}
	return nil, domain.ErrTemplateNotFound
}

// ListTemplatesQuery represents a query to list notification templates with optional channel filter.
type ListTemplatesQuery struct {
	Channel *domain.Channel
}

// ListTemplatesHandler handles queries to list notification templates.
type ListTemplatesHandler struct {
	repo domain.TemplateRepository
}

// NewListTemplatesHandler creates a new ListTemplatesHandler with the required repository.
func NewListTemplatesHandler(repo domain.TemplateRepository) *ListTemplatesHandler {
	return &ListTemplatesHandler{repo: repo}
}

// Handle executes the ListTemplatesQuery and returns matching templates.
// Results can be filtered by channel if specified.
func (h *ListTemplatesHandler) Handle(ctx context.Context, query ListTemplatesQuery) ([]*domain.NotificationTemplate, error) {
	return h.repo.List(ctx, query.Channel)
}
