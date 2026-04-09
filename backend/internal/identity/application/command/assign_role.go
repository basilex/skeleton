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

// AssignRoleHandler handles commands to assign a role to a user.
// It validates the user and role exist, assigns the role, and publishes domain events.
type AssignRoleHandler struct {
	users domain.UserRepository
	roles domain.RoleRepository
	bus   eventbus.Bus
}

// NewAssignRoleHandler creates a new AssignRoleHandler with the required dependencies.
func NewAssignRoleHandler(
	users domain.UserRepository,
	roles domain.RoleRepository,
	bus eventbus.Bus,
) *AssignRoleHandler {
	return &AssignRoleHandler{
		users: users,
		roles: roles,
		bus:   bus,
	}
}

// AssignRoleCommand represents a command to assign a role to a user.
type AssignRoleCommand struct {
	UserID string
	RoleID string
}

// Handle executes the AssignRoleCommand to assign a role to a user.
// It validates IDs, verifies the role exists, assigns the role to the user,
// persists changes, and publishes events.
func (h *AssignRoleHandler) Handle(ctx context.Context, cmd AssignRoleCommand) error {
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

	_, err = h.roles.FindByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("find role: %w", err)
	}

	if err := user.AssignRole(roleID); err != nil {
		return fmt.Errorf("assign role: %w", err)
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
