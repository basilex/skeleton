package command

import (
	"context"

	"github.com/basilex/skeleton/internal/notifications/domain"
)

type ProcessPendingCommand struct {
	Limit int
}

type ProcessPendingHandler struct {
	notificationRepo domain.NotificationRepository
}

func NewProcessPendingHandler(
	notificationRepo domain.NotificationRepository,
) *ProcessPendingHandler {
	return &ProcessPendingHandler{
		notificationRepo: notificationRepo,
	}
}

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
