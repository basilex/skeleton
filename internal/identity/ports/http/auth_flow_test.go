package http

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/basilex/skeleton/internal/identity/application/command"
	"github.com/basilex/skeleton/internal/identity/application/query"
	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/internal/identity/infrastructure/session"
	"github.com/basilex/skeleton/pkg/config"
	"github.com/basilex/skeleton/pkg/eventbus/memory"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func setupAuthTestRouter(t *testing.T) (*gin.Engine, *session.InMemoryStore, *mockUserRepoLogin, *mockRoleRepo) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	store := session.NewInMemoryStore(60)
	sessCfg := config.SessionConfig{CookieName: "session"}
	sessMw := session.NewMiddleware(store, sessCfg)

	userRepo := &mockUserRepoLogin{findByID: nil}
	roleRepo := &mockRoleRepo{}
	bus := memory.New()
	hasher := &domain.BcryptHasher{}
	tokenSvc := &mockTokenService{}

	registerH := command.NewRegisterUserHandler(userRepo, roleRepo, bus, hasher)
	loginH := command.NewLoginUserHandler(userRepo, roleRepo, tokenSvc)
	assignH := command.NewAssignRoleHandler(userRepo, roleRepo, bus)
	revokeH := command.NewRevokeRoleHandler(userRepo, roleRepo, bus)
	getUserH := query.NewGetUserHandler(userRepo, roleRepo)
	listUsersH := query.NewListUsersHandler(userRepo, roleRepo)

	handler := NewHandler(registerH, loginH, assignH, revokeH, getUserH, listUsersH, store)

	r := gin.New()
	v1 := r.Group("/api/v1")
	v1.POST("/auth/register", handler.Register)
	v1.POST("/auth/login", handler.Login)
	v1.POST("/auth/refresh", handler.Refresh)
	v1.POST("/auth/logout", sessMw.Authenticate(), handler.Logout)
	v1.GET("/users/me", sessMw.Authenticate(), handler.GetMyProfile)
	v1.GET("/users", sessMw.Authenticate(), handler.ListUsers)
	v1.GET("/users/:id", sessMw.Authenticate(), handler.GetUser)
	v1.POST("/users/:id/roles", sessMw.Authenticate(), handler.AssignRole)
	v1.DELETE("/users/:id/roles/:rid", sessMw.Authenticate(), handler.RevokeRole)

	return r, store, userRepo, roleRepo
}

func TestAuthFlow_RegisterLoginLogout(t *testing.T) {
	r, store, userRepo, _ := setupAuthTestRouter(t)

	email, _ := domain.NewEmail("test@example.com")
	hash, _ := domain.NewPasswordHash("Password1234!")
	user, _ := domain.NewUser(email, hash)
	userRepo.user = user
	userRepo.findByID = user

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Body = newBody(`{"email":"test@example.com","password":"Password1234!"}`)

	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	ctx := context.Background()
	sess, _ := store.Create(ctx, user.ID(), []string{"admin"}, []string{"users:read"}, "Mozilla", "127.0.0.1")

	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v1/users/me", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: sess.ID})

	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: sess.ID})

	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	_, err := store.Get(ctx, sess.ID)
	require.Error(t, err)
}

func TestAuthFlow_ProtectedEndpointsRequireAuth(t *testing.T) {
	r, _, _, _ := setupAuthTestRouter(t)

	protectedEndpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/users/me"},
		{"GET", "/api/v1/users"},
		{"GET", "/api/v1/users/some-id"},
		{"POST", "/api/v1/auth/logout"},
		{"POST", "/api/v1/users/some-id/roles"},
		{"DELETE", "/api/v1/users/some-id/roles/some-rid"},
	}

	for _, ep := range protectedEndpoints {
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(ep.method, ep.path, nil)
			r.ServeHTTP(w, req)
			require.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestAuthFlow_InvalidSessionRejected(t *testing.T) {
	r, _, _, _ := setupAuthTestRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "invalid-session-id"})

	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthFlow_SessionCookieProperties(t *testing.T) {
	r, store, userRepo, _ := setupAuthTestRouter(t)

	email, _ := domain.NewEmail("test@example.com")
	hash, _ := domain.NewPasswordHash("Password1234!")
	user, _ := domain.NewUser(email, hash)
	userRepo.user = user
	userRepo.findByID = user

	ctx := context.Background()
	sess, _ := store.Create(ctx, user.ID(), []string{"admin"}, []string{"users:read"}, "Mozilla", "127.0.0.1")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: sess.ID})

	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestAuthFlow_ExpiredSessionRejected(t *testing.T) {
	r, store, userRepo, _ := setupAuthTestRouter(t)

	email, _ := domain.NewEmail("test@example.com")
	hash, _ := domain.NewPasswordHash("Password1234!")
	user, _ := domain.NewUser(email, hash)
	userRepo.user = user

	ctx := context.Background()
	sess, _ := store.Create(ctx, user.ID(), []string{"admin"}, []string{"users:read"}, "Mozilla", "127.0.0.1")

	store.Delete(ctx, sess.ID)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: sess.ID})

	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func newBody(s string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(s))
}
