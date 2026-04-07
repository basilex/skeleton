// Package command provides command handlers for modifying identity state.
// This package implements the command side of CQRS for user-related operations,
// handling write requests that modify user and role assignments.
package command

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

// LogoutUserHandler handles commands to log out a user.
// It publishes a logout event to signal other parts of the system about the logout.
type LogoutUserHandler struct {
	users domain.UserRepository
	bus   eventbus.Bus
}

// NewLogoutUserHandler creates a new LogoutUserHandler with the required dependencies.
func NewLogoutUserHandler(users domain.UserRepository, bus eventbus.Bus) *LogoutUserHandler {
	return &LogoutUserHandler{
		users: users,
		bus:   bus,
	}
}

// LogoutUserCommand represents a command to log out a user.
type LogoutUserCommand struct {
	UserID string
}

// Handle executes the LogoutUserCommand to log out a user.
// It validates the user ID and publishes a UserLoggedOut domain event.
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
