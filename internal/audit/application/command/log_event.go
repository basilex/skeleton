package command

import (
	"context"

	"github.com/basilex/skeleton/internal/audit/domain"
)

type LogEventHandler struct {
	repo domain.AuditRepository
}

func NewLogEventHandler(repo domain.AuditRepository) *LogEventHandler {
	return &LogEventHandler{
		repo: repo,
	}
}

type LogEventCommand struct {
	ActorID    string
	ActorType  domain.ActorType
	Action     domain.Action
	Resource   string
	ResourceID string
	Metadata   string
	IP         string
	UserAgent  string
	Status     int
}

func (h *LogEventHandler) Handle(ctx context.Context, cmd LogEventCommand) error {
	record := domain.NewRecord(
		cmd.ActorID,
		cmd.ActorType,
		cmd.Action,
		cmd.Resource,
		cmd.ResourceID,
		cmd.Metadata,
		cmd.IP,
		cmd.UserAgent,
		cmd.Status,
	)

	return h.repo.Save(ctx, record)
}
