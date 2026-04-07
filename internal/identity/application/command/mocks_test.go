package command

import (
	"context"
	"testing"

	"github.com/basilex/skeleton/internal/identity/domain"
)

type mockUserRepo struct {
	t             *testing.T
	existingEmail string
	saved         *domain.User
}

func (m *mockUserRepo) Save(ctx context.Context, user *domain.User) error {
	m.saved = user
	return nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	return nil, nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	if m.existingEmail == email.String() {
		email, _ := domain.NewEmail(m.existingEmail)
		hash, _ := domain.NewPasswordHash("Password1234!")
		user, _ := domain.NewUser(email, hash)
		return user, nil
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepo) FindAll(ctx context.Context, filter domain.UserFilter) ([]*domain.User, int, error) {
	return nil, 0, nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id domain.UserID) error {
	return nil
}

type mockRoleRepo struct{}

func (m *mockRoleRepo) Save(ctx context.Context, role *domain.Role) error {
	return nil
}

func (m *mockRoleRepo) FindByID(ctx context.Context, id domain.RoleID) (*domain.Role, error) {
	return nil, nil
}

func (m *mockRoleRepo) FindByName(ctx context.Context, name string) (*domain.Role, error) {
	return nil, nil
}

func (m *mockRoleRepo) FindAll(ctx context.Context) ([]*domain.Role, error) {
	return nil, nil
}

func (m *mockRoleRepo) FindByIDs(ctx context.Context, ids []domain.RoleID) ([]*domain.Role, error) {
	return nil, nil
}
