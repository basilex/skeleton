// Package command provides command handlers for modifying notification state.
// This package implements the command side of CQRS for notification-related operations,
// handling write requests that create and modify notification entities.
package command

import (
	"context"

	"github.com/basilex/skeleton/internal/notifications/domain"
)

// ProcessPendingCommand represents a command to process pending notifications.
// It specifies the maximum number of notifications to process in a single batch.
type ProcessPendingCommand struct {
	Limit int
}

// ProcessPendingHandler handles commands to process pending notifications.
// It retrieves pending notifications and transitions them to queued status for processing.
type ProcessPendingHandler struct {
	notificationRepo domain.NotificationRepository
}

// NewProcessPendingHandler creates a new ProcessPendingHandler with the required repository.
func NewProcessPendingHandler(
	notificationRepo domain.NotificationRepository,
) *ProcessPendingHandler {
	return &ProcessPendingHandler{
		notificationRepo: notificationRepo,
	}
}

// Handle executes the ProcessPendingCommand to queue pending notifications.
// It retrieves pending notifications, transitions them to queued status,
// and returns the IDs of successfully queued notifications.
func (h *ProcessPendingHandler) Handle(ctx context.Context, cmd ProcessPendingCommand) ([]domain.NotificationID, error) {
	limit := cmd.Limit
	if limit <= 0 {
		limit = 100
	}

	notifications, err := h.notificationRepo.GetByStatus(ctx, domain.StatusPending, limit)
	if err != nil {
		return nil, err
	}

	ids := make([]domain.NotificationID, len(notifications))
	for i, n := range notifications {
		if err := n.Queue(); err != nil {
			continue
		}
		if err := h.notificationRepo.Update(ctx, n); err != nil {
			continue
		}
		ids[i] = n.ID()
	}

	return ids, nil
}
