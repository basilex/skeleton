// Package query provides query handlers for reading session and preferences data.
package query

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/identity/domain"
)

// GetSessionHandler handles queries to retrieve a single session by ID.
type GetSessionHandler struct {
	sessions domain.SessionRepository
}

// NewGetSessionHandler creates a new GetSessionHandler.
func NewGetSessionHandler(sessions domain.SessionRepository) *GetSessionHandler {
	return &GetSessionHandler{
		sessions: sessions,
	}
}

// GetSessionQuery represents a query to retrieve a session.
type GetSessionQuery struct {
	SessionID string
}

// SessionDTO represents session data for API responses.
type SessionDTO struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Status       string    `json:"status"`
	DeviceType   string    `json:"device_type"`
	OS           string    `json:"os"`
	Browser      string    `json:"browser"`
	DeviceName   string    `json:"device_name"`
	UserAgent    string    `json:"user_agent"`
	IPAddress    string    `json:"ip_address"`
	ExpiresAt    time.Time `json:"expires_at"`
	LastActivity time.Time `json:"last_activity"`
	CreatedAt    time.Time `json:"created_at"`
	IsActive     bool      `json:"is_active"`
}

// Handle executes the GetSessionQuery.
func (h *GetSessionHandler) Handle(ctx context.Context, q GetSessionQuery) (SessionDTO, error) {
	sessionID, err := domain.ParseSessionID(q.SessionID)
	if err != nil {
		return SessionDTO{}, fmt.Errorf("parse session id: %w", err)
	}

	session, err := h.sessions.FindByID(ctx, sessionID)
	if err != nil {
		return SessionDTO{}, fmt.Errorf("find session: %w", err)
	}

	var ipAddress string
	if ip := session.IPAddress(); ip != nil {
		ipAddress = ip.String()
	}

	return SessionDTO{
		ID:           session.ID().String(),
		UserID:       session.UserID().String(),
		Status:       session.Status().String(),
		DeviceType:   session.Device().DeviceType(),
		OS:           session.Device().OS(),
		Browser:      session.Device().Browser(),
		DeviceName:   session.Device().DeviceName(),
		UserAgent:    session.Device().UserAgent(),
		IPAddress:    ipAddress,
		ExpiresAt:    session.ExpiresAt(),
		LastActivity: session.LastActivity(),
		CreatedAt:    session.CreatedAt(),
		IsActive:     session.IsActive(),
	}, nil
}

// ListUserSessionsHandler handles queries to list all sessions for a user.
type ListUserSessionsHandler struct {
	sessions domain.SessionRepository
}

// NewListUserSessionsHandler creates a new ListUserSessionsHandler.
func NewListUserSessionsHandler(sessions domain.SessionRepository) *ListUserSessionsHandler {
	return &ListUserSessionsHandler{
		sessions: sessions,
	}
}

// ListUserSessionsQuery represents a query to list sessions for a user.
type ListUserSessionsQuery struct {
	UserID     string
	ActiveOnly bool
}

// Handle executes the ListUserSessionsQuery.
func (h *ListUserSessionsHandler) Handle(ctx context.Context, q ListUserSessionsQuery) ([]SessionDTO, error) {
	userID, err := domain.ParseUserID(q.UserID)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}

	var sessions []*domain.Session
	if q.ActiveOnly {
		sessions, err = h.sessions.FindActiveByUserID(ctx, userID)
	} else {
		sessions, err = h.sessions.FindByUserID(ctx, userID)
	}
	if err != nil {
		return nil, fmt.Errorf("find sessions: %w", err)
	}

	dtos := make([]SessionDTO, 0, len(sessions))
	for _, session := range sessions {
		var ipAddress string
		if ip := session.IPAddress(); ip != nil {
			ipAddress = ip.String()
		}

		dtos = append(dtos, SessionDTO{
			ID:           session.ID().String(),
			UserID:       session.UserID().String(),
			Status:       session.Status().String(),
			DeviceType:   session.Device().DeviceType(),
			OS:           session.Device().OS(),
			Browser:      session.Device().Browser(),
			DeviceName:   session.Device().DeviceName(),
			UserAgent:    session.Device().UserAgent(),
			IPAddress:    ipAddress,
			ExpiresAt:    session.ExpiresAt(),
			LastActivity: session.LastActivity(),
			CreatedAt:    session.CreatedAt(),
			IsActive:     session.IsActive(),
		})
	}

	return dtos, nil
}

// GetPreferencesHandler handles queries to retrieve user preferences.
type GetPreferencesHandler struct {
	preferences domain.PreferencesRepository
}

// NewGetPreferencesHandler creates a new GetPreferencesHandler.
func NewGetPreferencesHandler(preferences domain.PreferencesRepository) *GetPreferencesHandler {
	return &GetPreferencesHandler{
		preferences: preferences,
	}
}

// GetPreferencesQuery represents a query to retrieve user preferences.
type GetPreferencesQuery struct {
	UserID string
}

// PreferencesDTO represents preferences data for API responses.
type PreferencesDTO struct {
	ID                    string    `json:"id"`
	UserID                string    `json:"user_id"`
	Theme                 string    `json:"theme"`
	Language              string    `json:"language"`
	DateFormat            string    `json:"date_format"`
	Timezone              string    `json:"timezone"`
	EmailEnabled          bool      `json:"email_enabled"`
	SMSEnabled            bool      `json:"sms_enabled"`
	PushEnabled           bool      `json:"push_enabled"`
	InAppEnabled          bool      `json:"in_app_enabled"`
	MarketingEmails       bool      `json:"marketing_emails"`
	WeeklyDigest          bool      `json:"weekly_digest"`
	QuietHoursStart       *int      `json:"quiet_hours_start"`
	QuietHoursEnd         *int      `json:"quiet_hours_end"`
	NotificationsTimezone string    `json:"notifications_timezone"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// Handle executes the GetPreferencesQuery.
func (h *GetPreferencesHandler) Handle(ctx context.Context, q GetPreferencesQuery) (PreferencesDTO, error) {
	userID, err := domain.ParseUserID(q.UserID)
	if err != nil {
		return PreferencesDTO{}, fmt.Errorf("parse user id: %w", err)
	}

	prefs, err := h.preferences.FindByUserID(ctx, userID)
	if err != nil {
		return PreferencesDTO{}, fmt.Errorf("find preferences: %w", err)
	}

	notifications := prefs.Notifications()

	return PreferencesDTO{
		ID:                    prefs.ID().String(),
		UserID:                prefs.UserID().String(),
		Theme:                 prefs.Theme().String(),
		Language:              prefs.Language().String(),
		DateFormat:            prefs.DateFormat().String(),
		Timezone:              prefs.Timezone(),
		EmailEnabled:          notifications.EmailEnabled(),
		SMSEnabled:            notifications.SMSEnabled(),
		PushEnabled:           notifications.PushEnabled(),
		InAppEnabled:          notifications.InAppEnabled(),
		MarketingEmails:       notifications.MarketingEmails(),
		WeeklyDigest:          notifications.WeeklyDigest(),
		QuietHoursStart:       notifications.QuietHoursStart(),
		QuietHoursEnd:         notifications.QuietHoursEnd(),
		NotificationsTimezone: notifications.Timezone(),
		CreatedAt:             prefs.CreatedAt(),
		UpdatedAt:             prefs.UpdatedAt(),
	}, nil
}
