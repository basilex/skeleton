// Package http provides HTTP handlers for session and preferences operations.
package http

import (
	"net/http"
	"time"

	"github.com/basilex/skeleton/internal/identity/application/command"
	"github.com/basilex/skeleton/internal/identity/application/query"
	"github.com/basilex/skeleton/pkg/apierror"
	"github.com/gin-gonic/gin"
)

// SessionHandler provides HTTP handlers for session operations.
type SessionHandler struct {
	createSession      *command.CreateSessionHandler
	refreshSession     *command.RefreshSessionHandler
	revokeSession      *command.RevokeSessionHandler
	revokeUserSessions *command.RevokeUserSessionsHandler
	getSession         *query.GetSessionHandler
	listUserSessions   *query.ListUserSessionsHandler
}

// NewSessionHandler creates a new SessionHandler.
func NewSessionHandler(
	createSession *command.CreateSessionHandler,
	refreshSession *command.RefreshSessionHandler,
	revokeSession *command.RevokeSessionHandler,
	revokeUserSessions *command.RevokeUserSessionsHandler,
	getSession *query.GetSessionHandler,
	listUserSessions *query.ListUserSessionsHandler,
) *SessionHandler {
	return &SessionHandler{
		createSession:      createSession,
		refreshSession:     refreshSession,
		revokeSession:      revokeSession,
		revokeUserSessions: revokeUserSessions,
		getSession:         getSession,
		listUserSessions:   listUserSessions,
	}
}

// CreateSessionRequest represents the request to create a session.
type CreateSessionRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	UserAgent  string `json:"user_agent"`
	DeviceType string `json:"device_type"`
	OS         string `json:"os"`
	Browser    string `json:"browser"`
	DeviceName string `json:"device_name"`
	IPAddress  string `json:"ip_address"`
	Duration   int    `json:"duration"`
}

// CreateSession creates a new session.
func (h *SessionHandler) CreateSession(c *gin.Context) {
	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	duration := time.Hour * 24
	if req.Duration > 0 {
		duration = time.Duration(req.Duration) * time.Second
	}

	result, err := h.createSession.Handle(c.Request.Context(), command.CreateSessionCommand{
		UserID:     req.UserID,
		UserAgent:  req.UserAgent,
		DeviceType: req.DeviceType,
		OS:         req.OS,
		Browser:    req.Browser,
		DeviceName: req.DeviceName,
		Duration:   duration,
	})
	if err != nil {
		apierror.RespondError(c, apierror.NewInternal(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"session_id": result.SessionID,
		"expires_at": result.ExpiresAt,
	})
}

// RefreshSession refreshes an existing session.
func (h *SessionHandler) RefreshSession(c *gin.Context) {
	sessionID := c.Param("session_id")

	result, err := h.refreshSession.Handle(c.Request.Context(), command.RefreshSessionCommand{
		SessionID: sessionID,
		Duration:  time.Hour * 24,
	})
	if err != nil {
		apierror.RespondError(c, apierror.NewInternal(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"expires_at": result.ExpiresAt,
	})
}

// RevokeSession revokes a session.
func (h *SessionHandler) RevokeSession(c *gin.Context) {
	sessionID := c.Param("session_id")

	err := h.revokeSession.Handle(c.Request.Context(), command.RevokeSessionCommand{
		SessionID: sessionID,
		Reason:    "user_logout",
	})
	if err != nil {
		apierror.RespondError(c, apierror.NewInternal(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	c.Status(http.StatusNoContent)
}

// GetUserSessions lists all sessions for a user.
func (h *SessionHandler) GetUserSessions(c *gin.Context) {
	userID := c.Param("user_id")
	activeOnly := c.Query("active") == "true"

	sessions, err := h.listUserSessions.Handle(c.Request.Context(), query.ListUserSessionsQuery{
		UserID:     userID,
		ActiveOnly: activeOnly,
	})
	if err != nil {
		apierror.RespondError(c, apierror.NewInternal(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	c.JSON(http.StatusOK, sessions)
}

// PreferencesHandler provides HTTP handlers for user preferences operations.
type PreferencesHandler struct {
	updatePreferences *command.UpdatePreferencesHandler
	setTheme          *command.SetThemeHandler
	setLanguage       *command.SetLanguageHandler
	getPreferences    *query.GetPreferencesHandler
}

// NewPreferencesHandler creates a new PreferencesHandler.
func NewPreferencesHandler(
	updatePreferences *command.UpdatePreferencesHandler,
	setTheme *command.SetThemeHandler,
	setLanguage *command.SetLanguageHandler,
	getPreferences *query.GetPreferencesHandler,
) *PreferencesHandler {
	return &PreferencesHandler{
		updatePreferences: updatePreferences,
		setTheme:          setTheme,
		setLanguage:       setLanguage,
		getPreferences:    getPreferences,
	}
}

// UpdatePreferencesRequest represents the request to update preferences.
type UpdatePreferencesRequest struct {
	Theme                 *string `json:"theme"`
	Language              *string `json:"language"`
	DateFormat            *string `json:"date_format"`
	Timezone              *string `json:"timezone"`
	EmailEnabled          *bool   `json:"email_enabled"`
	SMSEnabled            *bool   `json:"sms_enabled"`
	PushEnabled           *bool   `json:"push_enabled"`
	InAppEnabled          *bool   `json:"in_app_enabled"`
	MarketingEmails       *bool   `json:"marketing_emails"`
	WeeklyDigest          *bool   `json:"weekly_digest"`
	QuietHoursStart       *int    `json:"quiet_hours_start"`
	QuietHoursEnd         *int    `json:"quiet_hours_end"`
	NotificationsTimezone *string `json:"notifications_timezone"`
}

// UpdatePreferences updates user preferences.
func (h *PreferencesHandler) UpdatePreferences(c *gin.Context) {
	userID := c.Param("user_id")

	var req UpdatePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	err := h.updatePreferences.Handle(c.Request.Context(), command.UpdatePreferencesCommand{
		UserID:                userID,
		Theme:                 req.Theme,
		Language:              req.Language,
		DateFormat:            req.DateFormat,
		Timezone:              req.Timezone,
		EmailEnabled:          req.EmailEnabled,
		SMSEnabled:            req.SMSEnabled,
		PushEnabled:           req.PushEnabled,
		InAppEnabled:          req.InAppEnabled,
		MarketingEmails:       req.MarketingEmails,
		WeeklyDigest:          req.WeeklyDigest,
		QuietHoursStart:       req.QuietHoursStart,
		QuietHoursEnd:         req.QuietHoursEnd,
		NotificationsTimezone: req.NotificationsTimezone,
	})
	if err != nil {
		apierror.RespondError(c, apierror.NewInternal(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	c.Status(http.StatusNoContent)
}

// GetPreferences gets user preferences.
func (h *PreferencesHandler) GetPreferences(c *gin.Context) {
	userID := c.Param("user_id")

	prefs, err := h.getPreferences.Handle(c.Request.Context(), query.GetPreferencesQuery{
		UserID: userID,
	})
	if err != nil {
		apierror.RespondError(c, apierror.NewInternal(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	c.JSON(http.StatusOK, prefs)
}

// SetThemeRequest represents the request to set theme.
type SetThemeRequest struct {
	Theme string `json:"theme" binding:"required"`
}

// SetTheme sets user theme preference.
func (h *PreferencesHandler) SetTheme(c *gin.Context) {
	userID := c.Param("user_id")

	var req SetThemeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	err := h.setTheme.Handle(c.Request.Context(), command.SetThemeCommand{
		UserID: userID,
		Theme:  req.Theme,
	})
	if err != nil {
		apierror.RespondError(c, apierror.NewInternal(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	c.Status(http.StatusNoContent)
}

// SetLanguageRequest represents the request to set language.
type SetLanguageRequest struct {
	Language string `json:"language" binding:"required"`
}

// SetLanguage sets user language preference.
func (h *PreferencesHandler) SetLanguage(c *gin.Context) {
	userID := c.Param("user_id")

	var req SetLanguageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	err := h.setLanguage.Handle(c.Request.Context(), command.SetLanguageCommand{
		UserID:   userID,
		Language: req.Language,
	})
	if err != nil {
		apierror.RespondError(c, apierror.NewInternal(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	c.Status(http.StatusNoContent)
}
