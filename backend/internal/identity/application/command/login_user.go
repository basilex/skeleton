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

// LoginUserHandler handles commands to authenticate a user and issue tokens.
// It validates credentials, checks user status, loads roles, and generates access/refresh tokens.
type LoginUserHandler struct {
	users        domain.UserRepository
	roles        domain.RoleRepository
	tokenService domain.TokenService
	bus          eventbus.Bus
}

// NewLoginUserHandler creates a new LoginUserHandler with the required dependencies.
func NewLoginUserHandler(
	users domain.UserRepository,
	roles domain.RoleRepository,
	tokenService domain.TokenService,
	bus eventbus.Bus,
) *LoginUserHandler {
	return &LoginUserHandler{
		users:        users,
		roles:        roles,
		tokenService: tokenService,
		bus:          bus,
	}
}

// LoginUserCommand represents a command to authenticate a user with email and password.
type LoginUserCommand struct {
	Email    string
	Password string
}

// TokenPair contains the access and refresh tokens returned after successful authentication.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Handle executes the LoginUserCommand to authenticate a user.
// It validates credentials, checks the user is active, loads roles,
// generates tokens, and publishes a login event.
func (h *LoginUserHandler) Handle(ctx context.Context, cmd LoginUserCommand) (TokenPair, error) {
	email, err := domain.NewEmail(cmd.Email)
	if err != nil {
		return TokenPair{}, fmt.Errorf("validate email: %w", err)
	}

	user, err := h.users.FindByEmail(ctx, email)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return TokenPair{}, domain.ErrInvalidPassword
		}
		return TokenPair{}, fmt.Errorf("find user: %w", err)
	}

	if !user.IsActive() {
		return TokenPair{}, domain.ErrUserInactive
	}

	if !user.CheckPassword(cmd.Password) {
		return TokenPair{}, domain.ErrInvalidPassword
	}

	roles, err := h.loadRoles(ctx, user.Roles())
	if err != nil {
		return TokenPair{}, fmt.Errorf("load roles: %w", err)
	}

	accessToken, err := h.tokenService.GenerateAccessToken(user.ID(), roles)
	if err != nil {
		return TokenPair{}, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := h.tokenService.GenerateRefreshToken()
	if err != nil {
		return TokenPair{}, fmt.Errorf("generate refresh token: %w", err)
	}

	if err := h.bus.Publish(ctx, domain.UserLoggedIn{
		UserID:    user.ID(),
		OcurredAt: time.Now().UTC(),
	}); err != nil {
		return TokenPair{}, fmt.Errorf("publish login event: %w", err)
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// loadRoles fetches role entities by their IDs and returns them as a slice.
// It converts the pointer slice returned by the repository to a value slice.
func (h *LoginUserHandler) loadRoles(ctx context.Context, ids []domain.RoleID) ([]domain.Role, error) {
	if len(ids) == 0 {
		return []domain.Role{}, nil
	}
	rolePtrs, err := h.roles.FindByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("find roles by IDs: %w", err)
	}
	roles := make([]domain.Role, len(rolePtrs))
	for i, r := range rolePtrs {
		roles[i] = *r
	}
	return roles, nil
}
