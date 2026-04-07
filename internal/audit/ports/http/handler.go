package http

import (
	"net/http"
	"time"

	"github.com/basilex/skeleton/internal/audit/application/query"
	"github.com/basilex/skeleton/internal/audit/domain"
	"github.com/basilex/skeleton/pkg/apierror"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	listRecords *query.ListRecordsHandler
}

func NewHandler(listRecords *query.ListRecordsHandler) *Handler {
	return &Handler{
		listRecords: listRecords,
	}
}

type ListRecordsRequest struct {
	ActorID  string `form:"actor_id"`
	Resource string `form:"resource"`
	Action   string `form:"action"`
	DateFrom string `form:"date_from"`
	DateTo   string `form:"date_to"`
	Cursor   string `form:"cursor"`
	Limit    int    `form:"limit"`
}

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

func getRequestID(c *gin.Context) string {
	if id, exists := c.Get("request_id"); exists {
		return id.(string)
	}
	return ""
}
