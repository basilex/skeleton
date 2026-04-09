package command

import (
	"context"
	"testing"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
	"github.com/basilex/skeleton/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type mockEventBus struct {
	published []eventbus.Event
}

func (m *mockEventBus) Publish(ctx context.Context, event eventbus.Event) error {
	m.published = append(m.published, event)
	return nil
}

func (m *mockEventBus) Subscribe(eventName string, handler eventbus.Handler) {
}

func TestLoginUserHandler_HappyPath(t *testing.T) {
	email, _ := domain.NewEmail("test@example.com")
	hash, _ := domain.NewPasswordHash("Password1234!")
	user, _ := domain.NewUser(email, hash)

	users := &mockUserRepoLogin{user: user}
	tokenSvc := &mockTokenService{}
	bus := &mockEventBus{}

	handler := NewLoginUserHandler(users, &mockRoleRepo{}, tokenSvc, bus)

	result, err := handler.Handle(context.Background(), LoginUserCommand{
		Email:    "test@example.com",
		Password: "Password1234!",
	})
	require.NoError(t, err)
	require.NotEmpty(t, result.AccessToken)
	require.NotEmpty(t, result.RefreshToken)
}

func TestLoginUserHandler_WrongPassword(t *testing.T) {
	email, _ := domain.NewEmail("test@example.com")
	hash, _ := domain.NewPasswordHash("Password1234!")
	user, _ := domain.NewUser(email, hash)

	users := &mockUserRepoLogin{user: user}
	tokenSvc := &mockTokenService{}
	bus := &mockEventBus{}

	handler := NewLoginUserHandler(users, &mockRoleRepo{}, tokenSvc, bus)

	_, err := handler.Handle(context.Background(), LoginUserCommand{
		Email:    "test@example.com",
		Password: "WrongPassword123!",
	})
	require.ErrorIs(t, err, domain.ErrInvalidPassword)
}

func TestLoginUserHandler_UserNotFound(t *testing.T) {
	users := &mockUserRepoLogin{err: domain.ErrUserNotFound}
	tokenSvc := &mockTokenService{}
	bus := &mockEventBus{}

	handler := NewLoginUserHandler(users, &mockRoleRepo{}, tokenSvc, bus)

	_, err := handler.Handle(context.Background(), LoginUserCommand{
		Email:    "test@example.com",
		Password: "Password1234!",
	})
	require.ErrorIs(t, err, domain.ErrInvalidPassword)
}

func TestLoginUserHandler_InactiveUser(t *testing.T) {
	email, _ := domain.NewEmail("test@example.com")
	hash, _ := domain.NewPasswordHash("Password1234!")
	user, _ := domain.NewUser(email, hash)
	user.Deactivate()

	users := &mockUserRepoLogin{user: user}
	tokenSvc := &mockTokenService{}
	bus := &mockEventBus{}

	handler := NewLoginUserHandler(users, &mockRoleRepo{}, tokenSvc, bus)

	_, err := handler.Handle(context.Background(), LoginUserCommand{
		Email:    "test@example.com",
		Password: "Password1234!",
	})
	require.ErrorIs(t, err, domain.ErrUserInactive)
}

type mockUserRepoLogin struct {
	user *domain.User
	err  error
}

func (m *mockUserRepoLogin) FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.user, nil
}

func (m *mockUserRepoLogin) Save(ctx context.Context, user *domain.User) error {
	return nil
}

func (m *mockUserRepoLogin) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	return nil, nil
}

func (m *mockUserRepoLogin) FindAll(ctx context.Context, filter domain.UserFilter) (pagination.PageResult[*domain.User], error) {
	return pagination.PageResult[*domain.User]{}, nil
}

func (m *mockUserRepoLogin) Delete(ctx context.Context, id domain.UserID) error {
	return nil
}

type mockTokenService struct{}

func (m *mockTokenService) GenerateAccessToken(userID domain.UserID, roles []domain.Role) (string, error) {
	return "access-test-token", nil
}

func (m *mockTokenService) GenerateRefreshToken() (string, error) {
	return "refresh-test-token", nil
}

func (m *mockTokenService) ValidateAccessToken(token string) (*domain.TokenClaims, error) {
	return nil, nil
}

func (m *mockTokenService) ValidateRefreshToken(token string) (domain.UserID, error) {
	return domain.UserID{}, nil
}
