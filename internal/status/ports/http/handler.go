package http

import (
	"net/http"

	"github.com/basilex/skeleton/internal/status/application/query"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	getBuildInfo *query.GetBuildInfoHandler
}

func NewHandler(getBuildInfo *query.GetBuildInfoHandler) *Handler {
	return &Handler{
		getBuildInfo: getBuildInfo,
	}
}

func (h *Handler) GetInfo(c *gin.Context) {
	result := h.getBuildInfo.Handle(c.Request.Context())
	c.JSON(http.StatusOK, result)
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) Ready(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}
