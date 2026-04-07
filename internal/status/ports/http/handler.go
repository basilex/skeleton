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

// GetInfo godoc
// @Summary Get build info
// @Description Returns application build information
// @Tags status
// @Produce json
// @Success 200 {object} domain.BuildInfo "Build info"
// @Router /build [get]
func (h *Handler) GetInfo(c *gin.Context) {
	result := h.getBuildInfo.Handle(c.Request.Context())
	c.JSON(http.StatusOK, result)
}

// Health godoc
// @Summary Health check
// @Description Returns service health status
// @Tags status
// @Produce json
// @Success 200 {object} map[string]string "Service is healthy"
// @Router /health [get]
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Ready godoc
// @Summary Readiness check
// @Description Returns service readiness status
// @Tags status
// @Produce json
// @Success 200 {object} map[string]string "Service is ready"
// @Router /ready [get]
func (h *Handler) Ready(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}
