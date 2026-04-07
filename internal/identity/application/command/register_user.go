package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

type RegisterUserHandler struct {
	users  domain.UserRepository
	roles  domain.RoleRepository
	bus    eventbus.Bus
	hasher domain.PasswordHasher
}

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

type RegisterUserCommand struct {
	Email    string
	Password string
}

type RegisterUserResult struct {
	UserID string
}

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
		UserID: string(user.ID()),
	}, nil
}
