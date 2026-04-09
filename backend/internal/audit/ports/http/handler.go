// Package http provides HTTP handlers for the audit service.
// This package implements the HTTP layer (ports) for handling audit log queries
// and providing access to historical system event records.
package http

import (
	"net/http"
	"time"

	"github.com/basilex/skeleton/internal/audit/application/query"
	"github.com/basilex/skeleton/internal/audit/domain"
	"github.com/basilex/skeleton/pkg/apierror"
	"github.com/gin-gonic/gin"
)

// Handler provides HTTP handlers for audit-related operations.
// It orchestrates query handlers to process HTTP requests for audit record retrieval
// with various filtering options including date ranges, actors, and resources.
type Handler struct {
	listRecords *query.ListRecordsHandler
}

// NewHandler creates a new HTTP handler for audit operations.
func NewHandler(listRecords *query.ListRecordsHandler) *Handler {
	return &Handler{
		listRecords: listRecords,
	}
}

// ListRecordsRequest represents query parameters for listing audit records.
// It provides filtering options by actor, resource, action, and date range.
type ListRecordsRequest struct {
	ActorID  string `form:"actor_id"`
	Resource string `form:"resource"`
	Action   string `form:"action"`
	DateFrom string `form:"date_from"`
	DateTo   string `form:"date_to"`
	Cursor   string `form:"cursor"`
	Limit    int    `form:"limit"`
}

// ListRecords godoc
// @Summary List audit records
// @Description Returns paginated list of audit records with optional filters
// @Tags audit
// @Produce json
// @Security BearerAuth
// @Param actor_id query string false "Filter by actor ID"
// @Param resource query string false "Filter by resource (user, role, auth)"
// @Param action query string false "Filter by action (create, read, update, delete, login, logout, etc.)"
// @Param date_from query string false "Filter from date (RFC3339 format)"
// @Param date_to query string false "Filter to date (RFC3339 format)"
// @Param cursor query string false "Pagination cursor (UUID v7)"
// @Param limit query int false "Items per page (default 20, max 100)"
// @Success 200 {object} map[string]interface{} "Paginated audit records"
// @Failure 401 {object} apierror.APIError "Unauthorized"
// @Failure 403 {object} apierror.APIError "Forbidden"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/audit/records [get]
func (h *Handler) ListRecords(c *gin.Context) {
	var req ListRecordsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	var dateFrom, dateTo time.Time
	var err error

	if req.DateFrom != "" {
		dateFrom, err = time.Parse(time.RFC3339, req.DateFrom)
		if err != nil {
			apierror.RespondError(c, apierror.NewValidation("invalid date_from format, use RFC3339", c.Request.URL.Path, getRequestID(c)))
			return
		}
	}

	if req.DateTo != "" {
		dateTo, err = time.Parse(time.RFC3339, req.DateTo)
		if err != nil {
			apierror.RespondError(c, apierror.NewValidation("invalid date_to format, use RFC3339", c.Request.URL.Path, getRequestID(c)))
			return
		}
	}

	actorID, _ := c.Get("user_id")
	actorType := domain.ActorUser
	if actorID == nil {
		actorID = "system"
		actorType = domain.ActorSystem
	}

	result, err := h.listRecords.Handle(c.Request.Context(), query.ListRecordsQuery{
		ActorID:              req.ActorID,
		Resource:             req.Resource,
		Action:               req.Action,
		DateFrom:             dateFrom,
		DateTo:               dateTo,
		Cursor:               req.Cursor,
		Limit:                req.Limit,
		RequestedByActorID:   actorID.(string),
		RequestedByActorType: actorType,
	})
	if err != nil {
		apierror.RespondError(c, apierror.NewInternal(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	c.JSON(http.StatusOK, result)
}

// getRequestID extracts the request ID from the gin context for error tracking.
func getRequestID(c *gin.Context) string {
	if id, exists := c.Get("request_id"); exists {
		return id.(string)
	}
	return ""
}
