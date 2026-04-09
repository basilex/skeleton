// Package command provides command handlers for creating audit records.
// This package implements the command side of CQRS for audit-related operations,
// handling write requests that create new audit log entries.
package command

import (
	"context"

	"github.com/basilex/skeleton/internal/audit/domain"
)

// LogEventHandler handles commands to create a new audit log entry.
// It captures system events and actions for compliance and traceability purposes.
type LogEventHandler struct {
	repo domain.AuditRepository
}

// NewLogEventHandler creates a new LogEventHandler with the required repository.
func NewLogEventHandler(repo domain.AuditRepository) *LogEventHandler {
	return &LogEventHandler{
		repo: repo,
	}
}

// LogEventCommand represents a command to log an audit event.
// It captures all relevant information about an action performed in the system.
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

// Handle executes the LogEventCommand to create a new audit record.
// It constructs the record entity from the command data and persists it.
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
