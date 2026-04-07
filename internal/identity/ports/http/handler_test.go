package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/basilex/skeleton/internal/identity/application/command"
	"github.com/basilex/skeleton/internal/identity/application/query"
	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/internal/identity/infrastructure/session"
	"github.com/basilex/skeleton/pkg/config"
	"github.com/basilex/skeleton/pkg/eventbus"
	"github.com/basilex/skeleton/pkg/pagination"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func setupTestRouter(t *testing.T, h *Handler, sessStore session.Store, sessCfg config.SessionConfig) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	sessMw := session.NewMiddleware(sessStore, sessCfg)

	v1 := r.Group("/api/v1")
	v1.POST("/auth/register", h.Register)
	v1.POST("/auth/login", h.Login)
	v1.POST("/auth/refresh", h.Refresh)
	v1.POST("/auth/logout", sessMw.Authenticate(), h.Logout)
	v1.GET("/users/me", sessMw.Authenticate(), h.GetMyProfile)
	v1.GET("/users", sessMw.Authenticate(), h.ListUsers)
	v1.GET("/users/:id", sessMw.Authenticate(), h.GetUser)
	v1.PATCH("/users/:id/deactivate", sessMw.Authenticate(), h.DeactivateUser)
	v1.POST("/users/:id/roles", sessMw.Authenticate(), h.AssignRole)
	v1.DELETE("/users/:id/roles/:rid", sessMw.Authenticate(), h.RevokeRole)

	return r
}

type mockCommandHandler struct {
	registerErr error
	loginErr    error
	loginResult command.TokenPair
}

func TestHandler_Register_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := session.NewInMemoryStore(60)
	sessCfg := config.SessionConfig{CookieName: "session"}

	bus := &mockEventBus{}
	hasher := &domain.BcryptHasher{}
	userRepo := &mockUserRepo{t: t}
	roleRepo := &mockRoleRepo{}

	registerH := command.NewRegisterUserHandler(userRepo, roleRepo, bus, hasher)
	loginH := command.NewLoginUserHandler(userRepo, roleRepo, &mockTokenService{}, bus)
	logoutH := command.NewLogoutUserHandler(userRepo, bus)
	assignH := command.NewAssignRoleHandler(userRepo, roleRepo, bus)
	revokeH := command.NewRevokeRoleHandler(userRepo, roleRepo, bus)
	getUserH := query.NewGetUserHandler(userRepo, roleRepo)
	listUsersH := query.NewListUsersHandler(userRepo, roleRepo)

	handler := NewHandler(registerH, loginH, logoutH, assignH, revokeH, getUserH, listUsersH, store)
	r := setupTestRouter(t, handler, store, sessCfg)

	body := bytes.NewBufferString(`{"email":"test@example.com","password":"Password1234!"}`)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/register", body)
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
}

func TestHandler_Register_InvalidEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := session.NewInMemoryStore(60)
	sessCfg := config.SessionConfig{CookieName: "session"}

	bus := &mockEventBus{}
	hasher := &domain.BcryptHasher{}
	userRepo := &mockUserRepo{t: t}
	roleRepo := &mockRoleRepo{}

	registerH := command.NewRegisterUserHandler(userRepo, roleRepo, bus, hasher)
	loginH := command.NewLoginUserHandler(userRepo, roleRepo, &mockTokenService{}, bus)
	logoutH := command.NewLogoutUserHandler(userRepo, bus)
	assignH := command.NewAssignRoleHandler(userRepo, roleRepo, bus)
	revokeH := command.NewRevokeRoleHandler(userRepo, roleRepo, bus)
	getUserH := query.NewGetUserHandler(userRepo, roleRepo)
	listUsersH := query.NewListUsersHandler(userRepo, roleRepo)

	handler := NewHandler(registerH, loginH, logoutH, assignH, revokeH, getUserH, listUsersH, store)
	r := setupTestRouter(t, handler, store, sessCfg)

	body := bytes.NewBufferString(`{"email":"invalid","password":"Password1234!"}`)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/register", body)
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestHandler_Register_ShortPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := session.NewInMemoryStore(60)
	sessCfg := config.SessionConfig{CookieName: "session"}

	bus := &mockEventBus{}
	hasher := &domain.BcryptHasher{}
	userRepo := &mockUserRepo{t: t}
	roleRepo := &mockRoleRepo{}

	registerH := command.NewRegisterUserHandler(userRepo, roleRepo, bus, hasher)
	loginH := command.NewLoginUserHandler(userRepo, roleRepo, &mockTokenService{}, bus)
	logoutH := command.NewLogoutUserHandler(userRepo, bus)
	assignH := command.NewAssignRoleHandler(userRepo, roleRepo, bus)
	revokeH := command.NewRevokeRoleHandler(userRepo, roleRepo, bus)
	getUserH := query.NewGetUserHandler(userRepo, roleRepo)
	listUsersH := query.NewListUsersHandler(userRepo, roleRepo)

	handler := NewHandler(registerH, loginH, logoutH, assignH, revokeH, getUserH, listUsersH, store)
	r := setupTestRouter(t, handler, store, sessCfg)

	body := bytes.NewBufferString(`{"email":"test@example.com","password":"short"}`)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/register", body)
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestHandler_Register_DuplicateEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := session.NewInMemoryStore(60)
	sessCfg := config.SessionConfig{CookieName: "session"}

	bus := &mockEventBus{}
	hasher := &domain.BcryptHasher{}
	userRepo := &mockUserRepo{t: t, existingEmail: "test@example.com"}
	roleRepo := &mockRoleRepo{}

	registerH := command.NewRegisterUserHandler(userRepo, roleRepo, bus, hasher)
	loginH := command.NewLoginUserHandler(userRepo, roleRepo, &mockTokenService{}, bus)
	logoutH := command.NewLogoutUserHandler(userRepo, bus)
	assignH := command.NewAssignRoleHandler(userRepo, roleRepo, bus)
	revokeH := command.NewRevokeRoleHandler(userRepo, roleRepo, bus)
	getUserH := query.NewGetUserHandler(userRepo, roleRepo)
	listUsersH := query.NewListUsersHandler(userRepo, roleRepo)

	handler := NewHandler(registerH, loginH, logoutH, assignH, revokeH, getUserH, listUsersH, store)
	r := setupTestRouter(t, handler, store, sessCfg)

	body := bytes.NewBufferString(`{"email":"test@example.com","password":"Password1234!"}`)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/register", body)
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandler_Login_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := session.NewInMemoryStore(60)
	sessCfg := config.SessionConfig{CookieName: "session"}

	email, _ := domain.NewEmail("test@example.com")
	hash, _ := domain.NewPasswordHash("Password1234!")
	user, _ := domain.NewUser(email, hash)

	bus := &mockEventBus{}
	userRepo := &mockUserRepoLogin{user: user}
	roleRepo := &mockRoleRepo{}
	tokenSvc := &mockTokenService{}

	registerH := command.NewRegisterUserHandler(userRepo, roleRepo, bus, &domain.BcryptHasher{})
	loginH := command.NewLoginUserHandler(userRepo, roleRepo, tokenSvc, bus)
	logoutH := command.NewLogoutUserHandler(userRepo, bus)
	assignH := command.NewAssignRoleHandler(userRepo, roleRepo, bus)
	revokeH := command.NewRevokeRoleHandler(userRepo, roleRepo, bus)
	getUserH := query.NewGetUserHandler(userRepo, roleRepo)
	listUsersH := query.NewListUsersHandler(userRepo, roleRepo)

	handler := NewHandler(registerH, loginH, logoutH, assignH, revokeH, getUserH, listUsersH, store)
	r := setupTestRouter(t, handler, store, sessCfg)

	body := bytes.NewBufferString(`{"email":"test@example.com","password":"Password1234!"}`)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/login", body)
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.NotEmpty(t, resp["access_token"])
	require.NotEmpty(t, resp["refresh_token"])
}

func TestHandler_Login_WrongPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := session.NewInMemoryStore(60)
	sessCfg := config.SessionConfig{CookieName: "session"}

	email, _ := domain.NewEmail("test@example.com")
	hash, _ := domain.NewPasswordHash("Password1234!")
	user, _ := domain.NewUser(email, hash)

	bus := &mockEventBus{}
	userRepo := &mockUserRepoLogin{user: user}
	roleRepo := &mockRoleRepo{}
	tokenSvc := &mockTokenService{}

	registerH := command.NewRegisterUserHandler(userRepo, roleRepo, bus, &domain.BcryptHasher{})
	loginH := command.NewLoginUserHandler(userRepo, roleRepo, tokenSvc, bus)
	logoutH := command.NewLogoutUserHandler(userRepo, bus)
	assignH := command.NewAssignRoleHandler(userRepo, roleRepo, bus)
	revokeH := command.NewRevokeRoleHandler(userRepo, roleRepo, bus)
	getUserH := query.NewGetUserHandler(userRepo, roleRepo)
	listUsersH := query.NewListUsersHandler(userRepo, roleRepo)

	handler := NewHandler(registerH, loginH, logoutH, assignH, revokeH, getUserH, listUsersH, store)
	r := setupTestRouter(t, handler, store, sessCfg)

	body := bytes.NewBufferString(`{"email":"test@example.com","password":"WrongPassword!"}`)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/login", body)
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandler_GetMyProfile_RequiresSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := session.NewInMemoryStore(60)
	sessCfg := config.SessionConfig{CookieName: "session"}

	bus := &mockEventBus{}
	userRepo := &mockUserRepo{t: t}
	roleRepo := &mockRoleRepo{}
	tokenSvc := &mockTokenService{}

	registerH := command.NewRegisterUserHandler(userRepo, roleRepo, bus, &domain.BcryptHasher{})
	loginH := command.NewLoginUserHandler(userRepo, roleRepo, tokenSvc, bus)
	logoutH := command.NewLogoutUserHandler(userRepo, bus)
	assignH := command.NewAssignRoleHandler(userRepo, roleRepo, bus)
	revokeH := command.NewRevokeRoleHandler(userRepo, roleRepo, bus)
	getUserH := query.NewGetUserHandler(userRepo, roleRepo)
	listUsersH := query.NewListUsersHandler(userRepo, roleRepo)

	handler := NewHandler(registerH, loginH, logoutH, assignH, revokeH, getUserH, listUsersH, store)
	r := setupTestRouter(t, handler, store, sessCfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/users/me", nil)

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_GetMyProfile_WithSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := session.NewInMemoryStore(60)
	sessCfg := config.SessionConfig{CookieName: "session"}

	email, _ := domain.NewEmail("test@example.com")
	hash, _ := domain.NewPasswordHash("Password1234!")
	user, _ := domain.NewUser(email, hash)

	bus := &mockEventBus{}
	userRepo := &mockUserRepoLogin{user: user, findByID: user}
	roleRepo := &mockRoleRepo{}
	tokenSvc := &mockTokenService{}

	registerH := command.NewRegisterUserHandler(userRepo, roleRepo, bus, &domain.BcryptHasher{})
	loginH := command.NewLoginUserHandler(userRepo, roleRepo, tokenSvc, bus)
	logoutH := command.NewLogoutUserHandler(userRepo, bus)
	assignH := command.NewAssignRoleHandler(userRepo, roleRepo, bus)
	revokeH := command.NewRevokeRoleHandler(userRepo, roleRepo, bus)
	getUserH := query.NewGetUserHandler(userRepo, roleRepo)
	listUsersH := query.NewListUsersHandler(userRepo, roleRepo)

	handler := NewHandler(registerH, loginH, logoutH, assignH, revokeH, getUserH, listUsersH, store)
	r := setupTestRouter(t, handler, store, sessCfg)

	ctx := context.Background()
	sess, _ := store.Create(ctx, user.ID(), []string{"admin"}, []string{"users:read"}, "Mozilla", "127.0.0.1")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: sess.ID})

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_Logout_RequiresSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := session.NewInMemoryStore(60)
	sessCfg := config.SessionConfig{CookieName: "session"}

	bus := &mockEventBus{}
	userRepo := &mockUserRepo{t: t}
	roleRepo := &mockRoleRepo{}
	tokenSvc := &mockTokenService{}

	registerH := command.NewRegisterUserHandler(userRepo, roleRepo, bus, &domain.BcryptHasher{})
	loginH := command.NewLoginUserHandler(userRepo, roleRepo, tokenSvc, bus)
	logoutH := command.NewLogoutUserHandler(userRepo, bus)
	assignH := command.NewAssignRoleHandler(userRepo, roleRepo, bus)
	revokeH := command.NewRevokeRoleHandler(userRepo, roleRepo, bus)
	getUserH := query.NewGetUserHandler(userRepo, roleRepo)
	listUsersH := query.NewListUsersHandler(userRepo, roleRepo)

	handler := NewHandler(registerH, loginH, logoutH, assignH, revokeH, getUserH, listUsersH, store)
	r := setupTestRouter(t, handler, store, sessCfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_Logout_WithSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := session.NewInMemoryStore(60)
	sessCfg := config.SessionConfig{CookieName: "session"}

	email, _ := domain.NewEmail("test@example.com")
	hash, _ := domain.NewPasswordHash("Password1234!")
	user, _ := domain.NewUser(email, hash)

	bus := &mockEventBus{}
	userRepo := &mockUserRepoLogin{user: user}
	roleRepo := &mockRoleRepo{}
	tokenSvc := &mockTokenService{}

	registerH := command.NewRegisterUserHandler(userRepo, roleRepo, bus, &domain.BcryptHasher{})
	loginH := command.NewLoginUserHandler(userRepo, roleRepo, tokenSvc, bus)
	logoutH := command.NewLogoutUserHandler(userRepo, bus)
	assignH := command.NewAssignRoleHandler(userRepo, roleRepo, bus)
	revokeH := command.NewRevokeRoleHandler(userRepo, roleRepo, bus)
	getUserH := query.NewGetUserHandler(userRepo, roleRepo)
	listUsersH := query.NewListUsersHandler(userRepo, roleRepo)

	handler := NewHandler(registerH, loginH, logoutH, assignH, revokeH, getUserH, listUsersH, store)
	r := setupTestRouter(t, handler, store, sessCfg)

	ctx := context.Background()
	sess, _ := store.Create(ctx, user.ID(), []string{"admin"}, []string{"users:read"}, "Mozilla", "127.0.0.1")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: sess.ID})

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	_, err := store.Get(ctx, sess.ID)
	require.Error(t, err)
}

type mockEventBus struct{}

func (m *mockEventBus) Publish(ctx context.Context, event eventbus.Event) error { return nil }
func (m *mockEventBus) Subscribe(eventName string, handler eventbus.Handler)    {}

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

func (m *mockUserRepo) FindAll(ctx context.Context, filter domain.UserFilter) (pagination.PageResult[*domain.User], error) {
	return pagination.PageResult[*domain.User]{}, nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id domain.UserID) error {
	return nil
}

type mockRoleRepo struct{}

func (m *mockRoleRepo) Save(ctx context.Context, role *domain.Role) error { return nil }
func (m *mockRoleRepo) FindByID(ctx context.Context, id domain.RoleID) (*domain.Role, error) {
	return nil, nil
}
func (m *mockRoleRepo) FindByName(ctx context.Context, name string) (*domain.Role, error) {
	return nil, nil
}
func (m *mockRoleRepo) FindAll(ctx context.Context) ([]*domain.Role, error) { return nil, nil }
func (m *mockRoleRepo) FindByIDs(ctx context.Context, ids []domain.RoleID) ([]*domain.Role, error) {
	return nil, nil
}

type mockUserRepoLogin struct {
	user     *domain.User
	findByID *domain.User
	err      error
}

func (m *mockUserRepoLogin) FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.user, nil
}

func (m *mockUserRepoLogin) Save(ctx context.Context, user *domain.User) error { return nil }
func (m *mockUserRepoLogin) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	return m.findByID, nil
}
func (m *mockUserRepoLogin) FindAll(ctx context.Context, filter domain.UserFilter) (pagination.PageResult[*domain.User], error) {
	return pagination.PageResult[*domain.User]{}, nil
}
func (m *mockUserRepoLogin) Delete(ctx context.Context, id domain.UserID) error { return nil }

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
	return "", nil
}
