package query

import (
	"context"

	"github.com/basilex/skeleton/internal/notifications/domain"
)

type GetNotificationQuery struct {
	ID domain.NotificationID
}

type GetNotificationHandler struct {
	repo domain.NotificationRepository
}

func NewGetNotificationHandler(repo domain.NotificationRepository) *GetNotificationHandler {
	return &GetNotificationHandler{repo: repo}
}

func (h *GetNotificationHandler) Handle(ctx context.Context, query GetNotificationQuery) (*domain.Notification, error) {
	return h.repo.GetByID(ctx, query.ID)
}

type ListNotificationsQuery struct {
	UserID   *string
	Status   *domain.Status
	Channel  *domain.Channel
	FromDate *string
	ToDate   *string
	Limit    int
	Cursor   *string
}

type ListNotificationsHandler struct {
	repo domain.NotificationRepository
}

func NewListNotificationsHandler(repo domain.NotificationRepository) *ListNotificationsHandler {
	return &ListNotificationsHandler{repo: repo}
}

func (h *ListNotificationsHandler) Handle(ctx context.Context, query ListNotificationsQuery) ([]*domain.Notification, error) {
	if query.Limit <= 0 {
		query.Limit = 100
	}

	if query.Status != nil {
		return h.repo.GetByStatus(ctx, *query.Status, query.Limit)
	}

	return h.repo.GetByStatus(ctx, domain.StatusPending, query.Limit)
}

type GetPreferencesQuery struct {
	UserID string
}

type GetPreferencesHandler struct {
	repo domain.PreferencesRepository
}

func NewGetPreferencesHandler(repo domain.PreferencesRepository) *GetPreferencesHandler {
	return &GetPreferencesHandler{repo: repo}
}

func (h *GetPreferencesHandler) Handle(ctx context.Context, query GetPreferencesQuery) (*domain.NotificationPreferences, error) {
	return h.repo.GetByUserID(ctx, query.UserID)
}

type GetTemplateQuery struct {
	ID   *domain.TemplateID
	Name *string
}

type GetTemplateHandler struct {
	repo domain.TemplateRepository
}

func NewGetTemplateHandler(repo domain.TemplateRepository) *GetTemplateHandler {
	return &GetTemplateHandler{repo: repo}
}

func (h *GetTemplateHandler) Handle(ctx context.Context, query GetTemplateQuery) (*domain.NotificationTemplate, error) {
	if query.ID != nil {
		return h.repo.GetByID(ctx, *query.ID)
	}
	if query.Name != nil {
		return h.repo.GetByName(ctx, *query.Name)
	}
	return nil, domain.ErrTemplateNotFound
}

type ListTemplatesQuery struct {
	Channel *domain.Channel
}

type ListTemplatesHandler struct {
	repo domain.TemplateRepository
}

func NewListTemplatesHandler(repo domain.TemplateRepository) *ListTemplatesHandler {
	return &ListTemplatesHandler{repo: repo}
}

func (h *ListTemplatesHandler) Handle(ctx context.Context, query ListTemplatesQuery) ([]*domain.NotificationTemplate, error) {
	return h.repo.List(ctx, query.Channel)
}
