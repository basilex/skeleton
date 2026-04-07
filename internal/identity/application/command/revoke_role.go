package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

type RevokeRoleHandler struct {
	users domain.UserRepository
	roles domain.RoleRepository
	bus   eventbus.Bus
}

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

type RevokeRoleCommand struct {
	UserID string
	RoleID string
}

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
