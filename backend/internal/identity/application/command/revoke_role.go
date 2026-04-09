// Package command provides command handlers for modifying identity state.
// This package implements the command side of CQRS for user-related operations,
// handling write requests that modify user and role assignments.
package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

// RevokeRoleHandler handles commands to revoke a role from a user.
// It validates the user and role exist, removes the role assignment, and publishes domain events.
type RevokeRoleHandler struct {
	users domain.UserRepository
	roles domain.RoleRepository
	bus   eventbus.Bus
}

// NewRevokeRoleHandler creates a new RevokeRoleHandler with the required dependencies.
func NewRevokeRoleHandler(
	users domain.UserRepository,
	roles domain.RoleRepository,
	bus eventbus.Bus,
) *RevokeRoleHandler {
	return &RevokeRoleHandler{
		users: users,
		roles: roles,
		bus:   bus,
	}
}

// RevokeRoleCommand represents a command to revoke a role from a user.
type RevokeRoleCommand struct {
	UserID string
	RoleID string
}

// Handle executes the RevokeRoleCommand to remove a role assignment from a user.
// It validates IDs, finds the user, revokes the role, persists changes, and publishes events.
func (h *RevokeRoleHandler) Handle(ctx context.Context, cmd RevokeRoleCommand) error {
	userID, err := domain.ParseUserID(cmd.UserID)
	if err != nil {
		return fmt.Errorf("parse user id: %w", err)
	}

	roleID, err := domain.ParseRoleID(cmd.RoleID)
	if err != nil {
		return fmt.Errorf("parse role id: %w", err)
	}

	user, err := h.users.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}

	if err := user.RevokeRole(roleID); err != nil {
		return fmt.Errorf("revoke role: %w", err)
	}

	if err := h.users.Save(ctx, user); err != nil {
		return fmt.Errorf("save user: %w", err)
	}

	events := user.PullEvents()
	for _, e := range events {
		if err := h.bus.Publish(ctx, e); err != nil {
			return fmt.Errorf("publish event: %w", err)
		}
	}

	return nil
}
