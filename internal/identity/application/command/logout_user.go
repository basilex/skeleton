package command

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

type LogoutUserHandler struct {
	users domain.UserRepository
	bus   eventbus.Bus
}

func NewLogoutUserHandler(users domain.UserRepository, bus eventbus.Bus) *LogoutUserHandler {
	return &LogoutUserHandler{
		users: users,
		bus:   bus,
	}
}

type LogoutUserCommand struct {
	UserID string
}

func (h *LogoutUserHandler) Handle(ctx context.Context, cmd LogoutUserCommand) error {
	userID, err := domain.ParseUserID(cmd.UserID)
	if err != nil {
		return fmt.Errorf("parse user id: %w", err)
	}

	if err := h.bus.Publish(ctx, domain.UserLoggedOut{
		UserID:    userID,
		OcurredAt: time.Now().UTC(),
	}); err != nil {
		return fmt.Errorf("publish logout event: %w", err)
	}

	return nil
}
