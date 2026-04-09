// Package command provides command handlers for modifying session state.
package command

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

// CreateSessionHandler handles commands to create a new session.
type CreateSessionHandler struct {
	sessions domain.SessionRepository
	bus      eventbus.Bus
}

// NewCreateSessionHandler creates a new CreateSessionHandler.
func NewCreateSessionHandler(
	sessions domain.SessionRepository,
	bus eventbus.Bus,
) *CreateSessionHandler {
	return &CreateSessionHandler{
		sessions: sessions,
		bus:      bus,
	}
}

// CreateSessionCommand represents a command to create a new session.
type CreateSessionCommand struct {
	UserID     string
	UserAgent  string
	DeviceType string
	OS         string
	Browser    string
	DeviceName string
	IPAddress  net.IP
	Duration   time.Duration
}

// CreateSessionResult contains the result of a successful session creation.
type CreateSessionResult struct {
	SessionID string
	ExpiresAt time.Time
}

// Handle executes the CreateSessionCommand.
func (h *CreateSessionHandler) Handle(ctx context.Context, cmd CreateSessionCommand) (CreateSessionResult, error) {
	userID, err := domain.ParseUserID(cmd.UserID)
	if err != nil {
		return CreateSessionResult{}, fmt.Errorf("parse user id: %w", err)
	}

	device := domain.NewDeviceInfo(cmd.UserAgent, cmd.DeviceType, cmd.OS, cmd.Browser, cmd.DeviceName)
	session, err := domain.NewSession(userID, device, cmd.IPAddress, cmd.Duration)
	if err != nil {
		return CreateSessionResult{}, fmt.Errorf("create session: %w", err)
	}

	if err := h.sessions.Save(ctx, session); err != nil {
		return CreateSessionResult{}, fmt.Errorf("save session: %w", err)
	}

	events := session.PullEvents()
	for _, e := range events {
		if err := h.bus.Publish(ctx, e); err != nil {
			return CreateSessionResult{}, fmt.Errorf("publish event: %w", err)
		}
	}

	return CreateSessionResult{
		SessionID: session.ID().String(),
		ExpiresAt: session.ExpiresAt(),
	}, nil
}

// RefreshSessionHandler handles commands to refresh a session.
type RefreshSessionHandler struct {
	sessions domain.SessionRepository
	bus      eventbus.Bus
}

// NewRefreshSessionHandler creates a new RefreshSessionHandler.
func NewRefreshSessionHandler(
	sessions domain.SessionRepository,
	bus eventbus.Bus,
) *RefreshSessionHandler {
	return &RefreshSessionHandler{
		sessions: sessions,
		bus:      bus,
	}
}

// RefreshSessionCommand represents a command to refresh a session.
type RefreshSessionCommand struct {
	SessionID string
	Duration  time.Duration
}

// RefreshSessionResult contains the result of a successful session refresh.
type RefreshSessionResult struct {
	ExpiresAt time.Time
}

// Handle executes the RefreshSessionCommand.
func (h *RefreshSessionHandler) Handle(ctx context.Context, cmd RefreshSessionCommand) (RefreshSessionResult, error) {
	sessionID, err := domain.ParseSessionID(cmd.SessionID)
	if err != nil {
		return RefreshSessionResult{}, fmt.Errorf("parse session id: %w", err)
	}

	session, err := h.sessions.FindByID(ctx, sessionID)
	if err != nil {
		return RefreshSessionResult{}, fmt.Errorf("find session: %w", err)
	}

	if err := session.Refresh(cmd.Duration); err != nil {
		return RefreshSessionResult{}, fmt.Errorf("refresh session: %w", err)
	}

	if err := h.sessions.Save(ctx, session); err != nil {
		return RefreshSessionResult{}, fmt.Errorf("save session: %w", err)
	}

	events := session.PullEvents()
	for _, e := range events {
		if err := h.bus.Publish(ctx, e); err != nil {
			return RefreshSessionResult{}, fmt.Errorf("publish event: %w", err)
		}
	}

	return RefreshSessionResult{
		ExpiresAt: session.ExpiresAt(),
	}, nil
}

// RevokeSessionHandler handles commands to revoke a session.
type RevokeSessionHandler struct {
	sessions domain.SessionRepository
	bus      eventbus.Bus
}

// NewRevokeSessionHandler creates a new RevokeSessionHandler.
func NewRevokeSessionHandler(
	sessions domain.SessionRepository,
	bus eventbus.Bus,
) *RevokeSessionHandler {
	return &RevokeSessionHandler{
		sessions: sessions,
		bus:      bus,
	}
}

// RevokeSessionCommand represents a command to revoke a session.
type RevokeSessionCommand struct {
	SessionID string
	Reason    string
}

// Handle executes the RevokeSessionCommand.
func (h *RevokeSessionHandler) Handle(ctx context.Context, cmd RevokeSessionCommand) error {
	sessionID, err := domain.ParseSessionID(cmd.SessionID)
	if err != nil {
		return fmt.Errorf("parse session id: %w", err)
	}

	session, err := h.sessions.FindByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("find session: %w", err)
	}

	if err := session.Revoke(cmd.Reason); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}

	if err := h.sessions.Save(ctx, session); err != nil {
		return fmt.Errorf("save session: %w", err)
	}

	events := session.PullEvents()
	for _, e := range events {
		if err := h.bus.Publish(ctx, e); err != nil {
			return fmt.Errorf("publish event: %w", err)
		}
	}

	return nil
}

// RevokeUserSessionsHandler handles commands to revoke all sessions for a user.
type RevokeUserSessionsHandler struct {
	sessions domain.SessionRepository
	bus      eventbus.Bus
}

// NewRevokeUserSessionsHandler creates a new RevokeUserSessionsHandler.
func NewRevokeUserSessionsHandler(
	sessions domain.SessionRepository,
	bus eventbus.Bus,
) *RevokeUserSessionsHandler {
	return &RevokeUserSessionsHandler{
		sessions: sessions,
		bus:      bus,
	}
}

// RevokeUserSessionsCommand represents a command to revoke all sessions for a user.
type RevokeUserSessionsCommand struct {
	UserID string
	Reason string
}

// Handle executes the RevokeUserSessionsCommand.
func (h *RevokeUserSessionsHandler) Handle(ctx context.Context, cmd RevokeUserSessionsCommand) error {
	userID, err := domain.ParseUserID(cmd.UserID)
	if err != nil {
		return fmt.Errorf("parse user id: %w", err)
	}

	sessions, err := h.sessions.FindActiveByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find active sessions: %w", err)
	}

	for _, session := range sessions {
		if err := session.Revoke(cmd.Reason); err != nil {
			continue
		}

		if err := h.sessions.Save(ctx, session); err != nil {
			continue
		}

		events := session.PullEvents()
		for _, e := range events {
			_ = h.bus.Publish(ctx, e)
		}
	}

	return nil
}

// CleanupExpiredSessionsHandler handles commands to cleanup expired sessions.
type CleanupExpiredSessionsHandler struct {
	sessions domain.SessionRepository
}

// NewCleanupExpiredSessionsHandler creates a new CleanupExpiredSessionsHandler.
func NewCleanupExpiredSessionsHandler(sessions domain.SessionRepository) *CleanupExpiredSessionsHandler {
	return &CleanupExpiredSessionsHandler{
		sessions: sessions,
	}
}

// CleanupExpiredSessionsCommand represents a command to cleanup expired sessions.
type CleanupExpiredSessionsCommand struct{}

// CleanupExpiredSessionsResult contains the result of cleanup.
type CleanupExpiredSessionsResult struct {
	DeletedCount int64
}

// Handle executes the CleanupExpiredSessionsCommand.
func (h *CleanupExpiredSessionsHandler) Handle(ctx context.Context, cmd CleanupExpiredSessionsCommand) (CleanupExpiredSessionsResult, error) {
	count, err := h.sessions.DeleteExpired(ctx)
	if err != nil {
		return CleanupExpiredSessionsResult{}, fmt.Errorf("delete expired sessions: %w", err)
	}

	return CleanupExpiredSessionsResult{
		DeletedCount: count,
	}, nil
}
