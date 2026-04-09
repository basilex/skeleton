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

// RegisterUserHandler handles commands to register a new user in the system.
// It validates the email, checks for duplicates, hashes the password, and creates the user.
type RegisterUserHandler struct {
	users  domain.UserRepository
	roles  domain.RoleRepository
	bus    eventbus.Bus
	hasher domain.PasswordHasher
}

// NewRegisterUserHandler creates a new RegisterUserHandler with the required dependencies.
func NewRegisterUserHandler(
	users domain.UserRepository,
	roles domain.RoleRepository,
	bus eventbus.Bus,
	hasher domain.PasswordHasher,
) *RegisterUserHandler {
	return &RegisterUserHandler{
		users:  users,
		roles:  roles,
		bus:    bus,
		hasher: hasher,
	}
}

// RegisterUserCommand represents a command to register a new user.
type RegisterUserCommand struct {
	Email    string
	Password string
}

// RegisterUserResult contains the result of a successful user registration.
type RegisterUserResult struct {
	UserID string
}

// Handle executes the RegisterUserCommand to create a new user.
// It validates the email format, ensures uniqueness, hashes the password,
// creates the user entity, persists it, and publishes domain events.
func (h *RegisterUserHandler) Handle(ctx context.Context, cmd RegisterUserCommand) (RegisterUserResult, error) {
	email, err := domain.NewEmail(cmd.Email)
	if err != nil {
		return RegisterUserResult{}, fmt.Errorf("validate email: %w", err)
	}

	existing, err := h.users.FindByEmail(ctx, email)
	if err != nil && err != domain.ErrUserNotFound {
		return RegisterUserResult{}, fmt.Errorf("find user by email: %w", err)
	}
	if existing != nil {
		return RegisterUserResult{}, domain.ErrUserAlreadyExists
	}

	passwordHash, err := domain.NewPasswordHash(cmd.Password)
	if err != nil {
		return RegisterUserResult{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := domain.NewUser(email, passwordHash)
	if err != nil {
		return RegisterUserResult{}, fmt.Errorf("create user: %w", err)
	}

	if err := h.users.Save(ctx, user); err != nil {
		return RegisterUserResult{}, fmt.Errorf("save user: %w", err)
	}

	events := user.PullEvents()
	for _, e := range events {
		if err := h.bus.Publish(ctx, e); err != nil {
			return RegisterUserResult{}, fmt.Errorf("publish event: %w", err)
		}
	}

	return RegisterUserResult{
		UserID: user.ID().String(),
	}, nil
}
