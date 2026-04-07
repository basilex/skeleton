package eventhandler

import (
	"context"
	"log/slog"

	"github.com/basilex/skeleton/internal/audit/application/command"
	"github.com/basilex/skeleton/internal/audit/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

type IdentityEventHandler struct {
	logEvent *command.LogEventHandler
}

func NewIdentityEventHandler(logEvent *command.LogEventHandler) *IdentityEventHandler {
	return &IdentityEventHandler{
		logEvent: logEvent,
	}
}

func (h *IdentityEventHandler) OnUserRegistered(ctx context.Context, event eventbus.Event) error {
	userRegistered, ok := event.(interface {
		GetUserID() string
		GetEmail() string
	})
	if !ok {
		slog.ErrorContext(ctx, "invalid user_registered event type")
		return nil
	}

	slog.InfoContext(ctx, "audit: user registered", "user_id", userRegistered.GetUserID())

	return h.logEvent.Handle(ctx, command.LogEventCommand{
		ActorID:    userRegistered.GetUserID(),
		ActorType:  domain.ActorUser,
		Action:     domain.ActionRegister,
		Resource:   "user",
		ResourceID: userRegistered.GetUserID(),
		Metadata:   `{"email":"` + userRegistered.GetEmail() + `"}`,
		Status:     201,
	})
}

func (h *IdentityEventHandler) OnRoleAssigned(ctx context.Context, event eventbus.Event) error {
	roleAssigned, ok := event.(interface {
		GetUserID() string
		GetRoleID() string
	})
	if !ok {
		slog.ErrorContext(ctx, "invalid role_assigned event type")
		return nil
	}

	slog.InfoContext(ctx, "audit: role assigned", "user_id", roleAssigned.GetUserID(), "role_id", roleAssigned.GetRoleID())

	return h.logEvent.Handle(ctx, command.LogEventCommand{
		ActorID:    roleAssigned.GetUserID(),
		ActorType:  domain.ActorUser,
		Action:     domain.ActionAssignRole,
		Resource:   "role",
		ResourceID: roleAssigned.GetRoleID(),
		Status:     200,
	})
}

func (h *IdentityEventHandler) OnRoleRevoked(ctx context.Context, event eventbus.Event) error {
	roleRevoked, ok := event.(interface {
		GetUserID() string
		GetRoleID() string
	})
	if !ok {
		slog.ErrorContext(ctx, "invalid role_revoked event type")
		return nil
	}

	slog.InfoContext(ctx, "audit: role revoked", "user_id", roleRevoked.GetUserID(), "role_id", roleRevoked.GetRoleID())

	return h.logEvent.Handle(ctx, command.LogEventCommand{
		ActorID:    roleRevoked.GetUserID(),
		ActorType:  domain.ActorUser,
		Action:     domain.ActionRevokeRole,
		Resource:   "role",
		ResourceID: roleRevoked.GetRoleID(),
		Status:     200,
	})
}

func (h *IdentityEventHandler) OnLogin(ctx context.Context, event eventbus.Event) error {
	loginEvent, ok := event.(interface {
		GetUserID() string
	})
	if !ok {
		slog.ErrorContext(ctx, "invalid login event type")
		return nil
	}

	slog.InfoContext(ctx, "audit: user logged in", "user_id", loginEvent.GetUserID())

	return h.logEvent.Handle(ctx, command.LogEventCommand{
		ActorID:    loginEvent.GetUserID(),
		ActorType:  domain.ActorUser,
		Action:     domain.ActionLogin,
		Resource:   "auth",
		ResourceID: loginEvent.GetUserID(),
		Status:     200,
	})
}

func (h *IdentityEventHandler) OnLogout(ctx context.Context, event eventbus.Event) error {
	logoutEvent, ok := event.(interface {
		GetUserID() string
	})
	if !ok {
		slog.ErrorContext(ctx, "invalid logout event type")
		return nil
	}

	slog.InfoContext(ctx, "audit: user logged out", "user_id", logoutEvent.GetUserID())

	return h.logEvent.Handle(ctx, command.LogEventCommand{
		ActorID:    logoutEvent.GetUserID(),
		ActorType:  domain.ActorUser,
		Action:     domain.ActionLogout,
		Resource:   "auth",
		ResourceID: logoutEvent.GetUserID(),
		Status:     200,
	})
}

func (h *IdentityEventHandler) Register(bus eventbus.Bus) {
	bus.Subscribe("identity.user_registered", h.OnUserRegistered)
	bus.Subscribe("identity.role_assigned", h.OnRoleAssigned)
	bus.Subscribe("identity.role_revoked", h.OnRoleRevoked)
	bus.Subscribe("identity.login", h.OnLogin)
	bus.Subscribe("identity.logout", h.OnLogout)
}
