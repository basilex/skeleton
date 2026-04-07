package http

import (
	"time"

	identityDomain "github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/internal/notifications/domain"
)

type CreateNotificationRequest struct {
	UserID      *string           `json:"user_id" example:"019d65d6-de90-7200-b1cf-4f8745597e0a"`
	Email       string            `json:"email" example:"user@example.com"`
	Phone       string            `json:"phone" example:"+1234567890"`
	DeviceToken string            `json:"device_token" example:"abc123def456"`
	Channel     string            `json:"channel" binding:"required" example:"email"`
	Subject     string            `json:"subject" binding:"required" example:"Welcome"`
	Content     string            `json:"content" binding:"required" example:"Welcome to our platform!"`
	HTMLContent string            `json:"html_content" example:"<h1>Welcome</h1>"`
	Priority    string            `json:"priority" example:"normal"`
	ScheduledAt *time.Time        `json:"scheduled_at" example:"2024-01-01T00:00:00Z"`
	MaxAttempts *int              `json:"max_attempts" example:"5"`
	Metadata    map[string]string `json:"metadata" example:"{\"key\":\"value\"}"`
}

type CreateFromTemplateRequest struct {
	TemplateName string            `json:"template_name" binding:"required" example:"welcome_email"`
	UserID       *string           `json:"user_id" example:"019d65d6-de90-7200-b1cf-4f8745597e0a"`
	Email        string            `json:"email" example:"user@example.com"`
	Phone        string            `json:"phone" example:"+1234567890"`
	DeviceToken  string            `json:"device_token" example:"abc123def456"`
	Variables    map[string]string `json:"variables" binding:"required" example:"{\"Email\":\"user@example.com\"}"`
	Priority     string            `json:"priority" example:"normal"`
	ScheduledAt  *time.Time        `json:"scheduled_at" example:"2024-01-01T00:00:00Z"`
	MaxAttempts  *int              `json:"max_attempts" example:"5"`
	Metadata     map[string]string `json:"metadata" example:"{\"key\":\"value\"}"`
}

type NotificationResponse struct {
	ID          string            `json:"id"`
	UserID      *string           `json:"user_id,omitempty"`
	Email       string            `json:"email,omitempty"`
	Phone       string            `json:"phone,omitempty"`
	DeviceToken string            `json:"device_token,omitempty"`
	Channel     string            `json:"channel"`
	Subject     string            `json:"subject"`
	Content     string            `json:"content"`
	HTMLContent string            `json:"html_content,omitempty"`
	Status      string            `json:"status"`
	Priority    string            `json:"priority"`
	ScheduledAt *time.Time        `json:"scheduled_at,omitempty"`
	SentAt      *time.Time        `json:"sent_at,omitempty"`
	DeliveredAt *time.Time        `json:"delivered_at,omitempty"`
	FailedAt    *time.Time        `json:"failed_at,omitempty"`
	Attempts    int               `json:"attempts"`
	MaxAttempts int               `json:"max_attempts"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type NotificationListResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	NextCursor    *string                `json:"next_cursor,omitempty"`
}

type NotificationPreferencesRequest struct {
	Channels map[string]ChannelPreferenceRequest `json:"channels" binding:"required"`
}

type ChannelPreferenceRequest struct {
	Enabled    bool               `json:"enabled"`
	Frequency  string             `json:"frequency" example:"immediate"`
	QuietHours *QuietHoursRequest `json:"quiet_hours,omitempty"`
}

type QuietHoursRequest struct {
	StartHour int    `json:"start_hour" example:"22"`
	EndHour   int    `json:"end_hour" example:"8"`
	Timezone  string `json:"timezone" example:"UTC"`
}

type NotificationPreferencesResponse struct {
	UserID    string                               `json:"user_id"`
	Channels  map[string]ChannelPreferenceResponse `json:"channels"`
	CreatedAt time.Time                            `json:"created_at"`
	UpdatedAt time.Time                            `json:"updated_at"`
}

type ChannelPreferenceResponse struct {
	Enabled    bool                `json:"enabled"`
	Frequency  string              `json:"frequency"`
	QuietHours *QuietHoursResponse `json:"quiet_hours,omitempty"`
}

type QuietHoursResponse struct {
	StartHour int    `json:"start_hour"`
	EndHour   int    `json:"end_hour"`
	Timezone  string `json:"timezone"`
}

type CreateTemplateRequest struct {
	Name      string   `json:"name" binding:"required" example:"welcome_email"`
	Channel   string   `json:"channel" binding:"required" example:"email"`
	Subject   string   `json:"subject" binding:"required" example:"Welcome to {{.AppName}}"`
	Body      string   `json:"body" binding:"required" example:"Hello {{.Name}}, welcome!"`
	HTMLBody  string   `json:"html_body" example:"<h1>Hello {{.Name}}</h1>"`
	Variables []string `json:"variables" example:"Name,AppName"`
}

type UpdateTemplateRequest struct {
	Subject   string   `json:"subject" binding:"required" example:"Welcome to {{.AppName}}"`
	Body      string   `json:"body" binding:"required" example:"Hello {{.Name}}, welcome!"`
	HTMLBody  string   `json:"html_body" example:"<h1>Hello {{.Name}}</h1>"`
	Variables []string `json:"variables" example:"Name,AppName"`
}

type TemplateResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Channel   string    `json:"channel"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	HTMLBody  string    `json:"html_body,omitempty"`
	Variables []string  `json:"variables"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToNotificationResponse(n *domain.Notification) NotificationResponse {
	var userID *string
	if n.Recipient().UserID != nil {
		id := string(*n.Recipient().UserID)
		userID = &id
	}

	return NotificationResponse{
		ID:          n.ID().String(),
		UserID:      userID,
		Email:       n.Recipient().Email,
		Phone:       n.Recipient().Phone,
		DeviceToken: n.Recipient().DeviceToken,
		Channel:     n.Channel().String(),
		Subject:     n.Subject(),
		Content:     n.Content().Text,
		HTMLContent: n.Content().HTML,
		Status:      n.Status().String(),
		Priority:    n.Priority().String(),
		ScheduledAt: n.ScheduledAt(),
		SentAt:      n.SentAt(),
		DeliveredAt: n.DeliveredAt(),
		FailedAt:    n.FailedAt(),
		Attempts:    n.Attempts(),
		MaxAttempts: n.MaxAttempts(),
		Metadata:    n.Metadata(),
		CreatedAt:   n.CreatedAt(),
		UpdatedAt:   n.UpdatedAt(),
	}
}

func ToPreferencesResponse(p *domain.NotificationPreferences) NotificationPreferencesResponse {
	channels := make(map[string]ChannelPreferenceResponse)
	for ch, pref := range p.Channels() {
		var qh *QuietHoursResponse
		if pref.QuietHours() != nil {
			qh = &QuietHoursResponse{
				StartHour: pref.QuietHours().StartHour(),
				EndHour:   pref.QuietHours().EndHour(),
				Timezone:  pref.QuietHours().Timezone(),
			}
		}
		channels[ch.String()] = ChannelPreferenceResponse{
			Enabled:    pref.Enabled(),
			Frequency:  pref.Frequency().String(),
			QuietHours: qh,
		}
	}

	return NotificationPreferencesResponse{
		UserID:    string(p.UserID()),
		Channels:  channels,
		CreatedAt: p.CreatedAt(),
		UpdatedAt: p.UpdatedAt(),
	}
}

func ToTemplateResponse(t *domain.NotificationTemplate) TemplateResponse {
	return TemplateResponse{
		ID:        t.ID().String(),
		Name:      t.Name(),
		Channel:   t.Channel().String(),
		Subject:   t.Subject(),
		Body:      t.Body(),
		HTMLBody:  t.HTMLBody(),
		Variables: t.Variables(),
		IsActive:  t.IsActive(),
		CreatedAt: t.CreatedAt(),
		UpdatedAt: t.UpdatedAt(),
	}
}

func parseUserID(s *string) *identityDomain.UserID {
	if s == nil || *s == "" {
		return nil
	}
	id := identityDomain.UserID(*s)
	return &id
}
