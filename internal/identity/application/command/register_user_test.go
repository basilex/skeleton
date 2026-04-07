package command

import (
	"context"
	"testing"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/pkg/eventbus/memory"
	"github.com/stretchr/testify/require"
)

func TestRegisterUserHandler_HappyPath(t *testing.T) {
	bus := memory.New()
	hasher := &domain.BcryptHasher{}
	users := &mockUserRepo{t: t}
	roles := &mockRoleRepo{}

	handler := NewRegisterUserHandler(users, roles, bus, hasher)

	result, err := handler.Handle(context.Background(), RegisterUserCommand{
		Email:    "test@example.com",
		Password: "Password1234!",
	})
	require.NoError(t, err)
	require.NotEmpty(t, result.UserID)
}

func TestRegisterUserHandler_DuplicateEmail(t *testing.T) {
	bus := memory.New()
	hasher := &domain.BcryptHasher{}
	users := &mockUserRepo{t: t, existingEmail: "test@example.com"}
	roles := &mockRoleRepo{}

	handler := NewRegisterUserHandler(users, roles, bus, hasher)

	_, err := handler.Handle(context.Background(), RegisterUserCommand{
		Email:    "test@example.com",
		Password: "Password1234!",
	})
	require.ErrorIs(t, err, domain.ErrUserAlreadyExists)
}

func TestRegisterUserHandler_InvalidEmail(t *testing.T) {
	bus := memory.New()
	hasher := &domain.BcryptHasher{}
	users := &mockUserRepo{t: t}
	roles := &mockRoleRepo{}

	handler := NewRegisterUserHandler(users, roles, bus, hasher)

	_, err := handler.Handle(context.Background(), RegisterUserCommand{
		Email:    "not-an-email",
		Password: "Password1234!",
	})
	require.Error(t, err)
}

func TestRegisterUserHandler_ShortPassword(t *testing.T) {
	bus := memory.New()
	hasher := &domain.BcryptHasher{}
	users := &mockUserRepo{t: t}
	roles := &mockRoleRepo{}

	handler := NewRegisterUserHandler(users, roles, bus, hasher)

	_, err := handler.Handle(context.Background(), RegisterUserCommand{
		Email:    "test@example.com",
		Password: "short",
	})
	require.Error(t, err)
}
