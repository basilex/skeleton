package http

import (
	"fmt"
	"net/http"

	"github.com/basilex/skeleton/internal/notifications/application/command"
	"github.com/basilex/skeleton/internal/notifications/application/query"
	"github.com/basilex/skeleton/internal/notifications/domain"
	"github.com/basilex/skeleton/pkg/apierror"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	createNotification *command.CreateNotificationHandler
	createFromTemplate *command.CreateFromTemplateHandler
	markDelivered      *command.MarkDeliveredHandler
	markFailed         *command.MarkFailedHandler
	updatePreferences  *command.UpdatePreferencesHandler
	getNotification    *query.GetNotificationHandler
	listNotifications  *query.ListNotificationsHandler
	getPreferences     *query.GetPreferencesHandler
	createTemplate     *command.CreateTemplateHandler
	updateTemplate     *command.UpdateTemplateHandler
	getTemplate        *query.GetTemplateHandler
	listTemplates      *query.ListTemplatesHandler
}

func NewHandler(
	createNotification *command.CreateNotificationHandler,
	createFromTemplate *command.CreateFromTemplateHandler,
	markDelivered *command.MarkDeliveredHandler,
	markFailed *command.MarkFailedHandler,
	updatePreferences *command.UpdatePreferencesHandler,
	getNotification *query.GetNotificationHandler,
	listNotifications *query.ListNotificationsHandler,
	getPreferences *query.GetPreferencesHandler,
	createTemplate *command.CreateTemplateHandler,
	updateTemplate *command.UpdateTemplateHandler,
	getTemplate *query.GetTemplateHandler,
	listTemplates *query.ListTemplatesHandler,
) *Handler {
	return &Handler{
		createNotification: createNotification,
		createFromTemplate: createFromTemplate,
		markDelivered:      markDelivered,
		markFailed:         markFailed,
		updatePreferences:  updatePreferences,
		getNotification:    getNotification,
		listNotifications:  listNotifications,
		getPreferences:     getPreferences,
		createTemplate:     createTemplate,
		updateTemplate:     updateTemplate,
		getTemplate:        getTemplate,
		listTemplates:      listTemplates,
	}
}

// CreateNotification godoc
// @Summary Create a notification
// @Description Create a new notification manually (admin only)
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateNotificationRequest true "Notification data"
// @Success 201 {object} map[string]string "Notification created"
// @Failure 400 {object} apierror.APIError "Validation error"
// @Failure 401 {object} apierror.APIError "Unauthorized"
// @Failure 403 {object} apierror.APIError "Forbidden"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/notifications [post]
func (h *Handler) CreateNotification(c *gin.Context) {
	var req CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	channel, err := domain.ParseChannel(req.Channel)
	if err != nil {
		apierror.RespondError(c, apierror.NewValidation("invalid channel", c.Request.URL.Path, getRequestID(c)))
		return
	}

	priority := domain.PriorityNormal
	if req.Priority != "" {
		priority, err = domain.ParsePriority(req.Priority)
		if err != nil {
			apierror.RespondError(c, apierror.NewValidation("invalid priority", c.Request.URL.Path, getRequestID(c)))
			return
		}
	}

	recipient := domain.Recipient{
		UserID:      parseUserID(req.UserID),
		Email:       req.Email,
		Phone:       req.Phone,
		DeviceToken: req.DeviceToken,
	}

	content := domain.Content{
		Text: req.Content,
	}
	if req.HTMLContent != "" {
		content.HTML = req.HTMLContent
	}

	var scheduledAtStr *string
	if req.ScheduledAt != nil {
		s := req.ScheduledAt.Format("2006-01-02T15:04:05Z07:00")
		scheduledAtStr = &s
	}

	id, err := h.createNotification.Handle(c.Request.Context(), command.CreateNotificationCommand{
		Recipient:   recipient,
		Channel:     channel,
		Subject:     req.Subject,
		Content:     content,
		Priority:    priority,
		ScheduledAt: scheduledAtStr,
		MaxAttempts: req.MaxAttempts,
		Metadata:    req.Metadata,
	})
	if err != nil {
		handleNotificationError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id.String()})
}

// GetNotification godoc
// @Summary Get notification by ID
// @Description Get notification details by ID
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 200 {object} NotificationResponse "Notification details"
// @Failure 400 {object} apierror.APIError "Invalid ID"
// @Failure 401 {object} apierror.APIError "Unauthorized"
// @Failure 404 {object} apierror.APIError "Not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/notifications/{id} [get]
func (h *Handler) GetNotification(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		apierror.RespondError(c, apierror.NewValidation("notification id required", c.Request.URL.Path, getRequestID(c)))
		return
	}

	notificationID, err := domain.ParseNotificationID(id)
	if err != nil {
		apierror.RespondError(c, apierror.NewValidation("invalid notification id", c.Request.URL.Path, getRequestID(c)))
		return
	}

	notification, err := h.getNotification.Handle(c.Request.Context(), query.GetNotificationQuery{ID: notificationID})
	if err != nil {
		handleNotificationError(c, err)
		return
	}

	c.JSON(http.StatusOK, ToNotificationResponse(notification))
}

// ListNotifications godoc
// @Summary List notifications
// @Description List notifications with optional filters
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param user_id query string false "Filter by user ID"
// @Param status query string false "Filter by status (pending, queued, sending, sent, delivered, failed)"
// @Param channel query string false "Filter by channel (email, sms, push, in_app)"
// @Param limit query int false "Items per page (default 20)"
// @Success 200 {array} NotificationResponse "Notifications list"
// @Failure 401 {object} apierror.APIError "Unauthorized"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/notifications [get]
func (h *Handler) ListNotifications(c *gin.Context) {
	userID := c.Query("user_id")
	statusStr := c.Query("status")
	channelStr := c.Query("channel")
	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := parseInt(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	var status *domain.Status
	if statusStr != "" {
		s, err := domain.ParseStatus(statusStr)
		if err != nil {
			apierror.RespondError(c, apierror.NewValidation("invalid status", c.Request.URL.Path, getRequestID(c)))
			return
		}
		status = &s
	}

	var channel *domain.Channel
	if channelStr != "" {
		ch, err := domain.ParseChannel(channelStr)
		if err != nil {
			apierror.RespondError(c, apierror.NewValidation("invalid channel", c.Request.URL.Path, getRequestID(c)))
			return
		}
		channel = &ch
	}

	var userIDPtr *string
	if userID != "" {
		userIDPtr = &userID
	}

	notifications, err := h.listNotifications.Handle(c.Request.Context(), query.ListNotificationsQuery{
		UserID:  userIDPtr,
		Status:  status,
		Channel: channel,
		Limit:   limit,
	})
	if err != nil {
		handleNotificationError(c, err)
		return
	}

	responses := make([]NotificationResponse, len(notifications))
	for i, n := range notifications {
		responses[i] = ToNotificationResponse(n)
	}

	c.JSON(http.StatusOK, responses)
}

// GetPreferences godoc
// @Summary Get user notification preferences
// @Description Get current user's notification preferences
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {object} NotificationPreferencesResponse "User preferences"
// @Failure 401 {object} apierror.APIError "Unauthorized"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/notifications/preferences [get]
func (h *Handler) GetPreferences(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		apierror.RespondError(c, apierror.NewUnauthorized("authentication required", c.Request.URL.Path, getRequestID(c)))
		return
	}

	prefs, err := h.getPreferences.Handle(c.Request.Context(), query.GetPreferencesQuery{UserID: userID})
	if err != nil {
		handleNotificationError(c, err)
		return
	}

	c.JSON(http.StatusOK, ToPreferencesResponse(prefs))
}

// UpdatePreferences godoc
// @Summary Update user notification preferences
// @Description Update current user's notification preferences
// @Tags notifications
// @Accept json
// @Security BearerAuth
// @Param request body NotificationPreferencesRequest true "Preferences data"
// @Success 204 "Preferences updated"
// @Failure 400 {object} apierror.APIError "Validation error"
// @Failure 401 {object} apierror.APIError "Unauthorized"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/notifications/preferences [patch]
func (h *Handler) UpdatePreferences(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		apierror.RespondError(c, apierror.NewUnauthorized("authentication required", c.Request.URL.Path, getRequestID(c)))
		return
	}

	var req NotificationPreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	for ch, pref := range req.Channels {
		channel, err := domain.ParseChannel(ch)
		if err != nil {
			continue
		}

		frequency := domain.FrequencyImmediate
		if pref.Frequency != "" {
			frequency, err = domain.ParseFrequency(pref.Frequency)
			if err != nil {
				frequency = domain.FrequencyImmediate
			}
		}

		cmd := command.UpdatePreferencesCommand{
			UserID:    userID,
			Channel:   channel,
			Enabled:   pref.Enabled,
			Frequency: frequency,
		}

		if pref.QuietHours != nil {
			cmd.QuietHoursStart = &pref.QuietHours.StartHour
			cmd.QuietHoursEnd = &pref.QuietHours.EndHour
			cmd.QuietHoursTZ = &pref.QuietHours.Timezone
		}

		if err := h.updatePreferences.Handle(c.Request.Context(), cmd); err != nil {
			handleNotificationError(c, err)
			return
		}
	}

	c.Status(http.StatusNoContent)
}

// CreateTemplate godoc
// @Summary Create notification template
// @Description Create a new notification template (admin only)
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateTemplateRequest true "Template data"
// @Success 201 {object} map[string]string "Template created"
// @Failure 400 {object} apierror.APIError "Validation error"
// @Failure 401 {object} apierror.APIError "Unauthorized"
// @Failure 403 {object} apierror.APIError "Forbidden"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/notifications/templates [post]
func (h *Handler) CreateTemplate(c *gin.Context) {
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	channel, err := domain.ParseChannel(req.Channel)
	if err != nil {
		apierror.RespondError(c, apierror.NewValidation("invalid channel", c.Request.URL.Path, getRequestID(c)))
		return
	}

	templateID, err := h.createTemplate.Handle(c.Request.Context(), command.CreateTemplateCommand{
		Name:      req.Name,
		Channel:   channel,
		Subject:   req.Subject,
		Body:      req.Body,
		HTMLBody:  req.HTMLBody,
		Variables: req.Variables,
	})
	if err != nil {
		handleNotificationError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": templateID.String()})
}

// GetTemplate godoc
// @Summary Get template by ID
// @Description Get notification template details by ID (admin only)
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param id path string true "Template ID"
// @Success 200 {object} TemplateResponse "Template details"
// @Failure 400 {object} apierror.APIError "Invalid ID"
// @Failure 401 {object} apierror.APIError "Unauthorized"
// @Failure 403 {object} apierror.APIError "Forbidden"
// @Failure 404 {object} apierror.APIError "Not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/notifications/templates/{id} [get]
func (h *Handler) GetTemplate(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		apierror.RespondError(c, apierror.NewValidation("template id required", c.Request.URL.Path, getRequestID(c)))
		return
	}

	templateID, err := domain.ParseTemplateID(id)
	if err != nil {
		apierror.RespondError(c, apierror.NewValidation("invalid template id", c.Request.URL.Path, getRequestID(c)))
		return
	}

	template, err := h.getTemplate.Handle(c.Request.Context(), query.GetTemplateQuery{ID: &templateID})
	if err != nil {
		handleNotificationError(c, err)
		return
	}

	c.JSON(http.StatusOK, ToTemplateResponse(template))
}

// ListTemplates godoc
// @Summary List notification templates
// @Description List all notification templates optionally filtered by channel (admin only)
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param channel query string false "Filter by channel (email, sms, push, in_app)"
// @Success 200 {array} TemplateResponse "Templates list"
// @Failure 401 {object} apierror.APIError "Unauthorized"
// @Failure 403 {object} apierror.APIError "Forbidden"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/notifications/templates [get]
func (h *Handler) ListTemplates(c *gin.Context) {
	channelStr := c.Query("channel")

	var channel *domain.Channel
	if channelStr != "" {
		ch, err := domain.ParseChannel(channelStr)
		if err != nil {
			apierror.RespondError(c, apierror.NewValidation("invalid channel", c.Request.URL.Path, getRequestID(c)))
			return
		}
		channel = &ch
	}

	templates, err := h.listTemplates.Handle(c.Request.Context(), query.ListTemplatesQuery{Channel: channel})
	if err != nil {
		handleNotificationError(c, err)
		return
	}

	responses := make([]TemplateResponse, len(templates))
	for i, t := range templates {
		responses[i] = ToTemplateResponse(t)
	}

	c.JSON(http.StatusOK, responses)
}

// UpdateTemplate godoc
// @Summary Update notification template
// @Description Update an existing notification template (admin only)
// @Tags notifications
// @Accept json
// @Security BearerAuth
// @Param id path string true "Template ID"
// @Param request body UpdateTemplateRequest true "Template data"
// @Success 204 "Template updated"
// @Failure 400 {object} apierror.APIError "Validation error"
// @Failure 401 {object} apierror.APIError "Unauthorized"
// @Failure 403 {object} apierror.APIError "Forbidden"
// @Failure 404 {object} apierror.APIError "Not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/notifications/templates/{id} [patch]
func (h *Handler) UpdateTemplate(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		apierror.RespondError(c, apierror.NewValidation("template id required", c.Request.URL.Path, getRequestID(c)))
		return
	}

	templateID, err := domain.ParseTemplateID(id)
	if err != nil {
		apierror.RespondError(c, apierror.NewValidation("invalid template id", c.Request.URL.Path, getRequestID(c)))
		return
	}

	var req UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	if err := h.updateTemplate.Handle(c.Request.Context(), command.UpdateTemplateCommand{
		ID:        templateID,
		Subject:   req.Subject,
		Body:      req.Body,
		HTMLBody:  req.HTMLBody,
		Variables: req.Variables,
		IsActive:  true,
	}); err != nil {
		handleNotificationError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func handleNotificationError(c *gin.Context, err error) {
	switch err {
	case domain.ErrNotificationNotFound:
		apierror.RespondError(c, apierror.NewNotFound("notification not found", c.Request.URL.Path, getRequestID(c)))
	case domain.ErrTemplateNotFound:
		apierror.RespondError(c, apierror.NewNotFound("template not found", c.Request.URL.Path, getRequestID(c)))
	case domain.ErrPreferencesNotFound:
		apierror.RespondError(c, apierror.NewNotFound("preferences not found", c.Request.URL.Path, getRequestID(c)))
	default:
		apierror.RespondError(c, apierror.NewInternal("internal error", c.Request.URL.Path, getRequestID(c)))
	}
}

func getUserID(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return ""
	}
	return userID.(string)
}

func getRequestID(c *gin.Context) string {
	requestID, exists := c.Get("request_id")
	if !exists {
		return ""
	}
	return requestID.(string)
}

func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}
