package query

import (
	"context"

	"github.com/basilex/skeleton/internal/status/domain"
)

type GetBuildInfoHandler struct {
	buildInfo domain.BuildInfo
}

func NewGetBuildInfoHandler(buildInfo domain.BuildInfo) *GetBuildInfoHandler {
	return &GetBuildInfoHandler{
		buildInfo: buildInfo,
	}
}

func (h *GetBuildInfoHandler) Handle(_ context.Context) domain.BuildInfo {
	return h.buildInfo
}
