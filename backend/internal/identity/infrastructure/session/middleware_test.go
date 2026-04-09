package session

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestInMemoryStore_CreateAndGet(t *testing.T) {
	store := NewInMemoryStore(60)
	ctx := context.Background()

	userID := domain.NewUserID()
	sess, err := store.Create(ctx, userID, []string{"admin"}, []string{"*:*"}, "Mozilla", "127.0.0.1")
	require.NoError(t, err)
	require.NotEmpty(t, sess.ID)
	require.Equal(t, userID, sess.UserID)
	require.Equal(t, []string{"admin"}, sess.Roles)

	got, err := store.Get(ctx, sess.ID)
	require.NoError(t, err)
	require.Equal(t, sess.ID, got.ID)
}

func TestInMemoryStore_GetNotFound(t *testing.T) {
	store := NewInMemoryStore(60)
	ctx := context.Background()

	_, err := store.Get(ctx, "nonexistent")
	require.Error(t, err)
}

func TestInMemoryStore_Delete(t *testing.T) {
	store := NewInMemoryStore(60)
	ctx := context.Background()

	sess, _ := store.Create(ctx, domain.NewUserID(), nil, nil, "", "")
	err := store.Delete(ctx, sess.ID)
	require.NoError(t, err)

	_, err = store.Get(ctx, sess.ID)
	require.Error(t, err)
}

func TestInMemoryStore_DeleteAllForUser(t *testing.T) {
	store := NewInMemoryStore(60)
	ctx := context.Background()

	userID1 := domain.NewUserID()
	userID2 := domain.NewUserID()
	s1, _ := store.Create(ctx, userID1, nil, nil, "", "")
	s2, _ := store.Create(ctx, userID1, nil, nil, "", "")
	s3, _ := store.Create(ctx, userID2, nil, nil, "", "")

	err := store.DeleteAllForUser(ctx, userID1)
	require.NoError(t, err)

	_, err = store.Get(ctx, s1.ID)
	require.Error(t, err)
	_, err = store.Get(ctx, s2.ID)
	require.Error(t, err)
	_, err = store.Get(ctx, s3.ID)
	require.NoError(t, err)
}

func TestInMemoryStore_Touch(t *testing.T) {
	store := NewInMemoryStore(60)
	ctx := context.Background()

	sess, _ := store.Create(ctx, domain.NewUserID(), nil, nil, "", "")
	oldExpiry := sess.ExpiresAt

	time.Sleep(10 * time.Millisecond)
	err := store.Touch(ctx, sess.ID)
	require.NoError(t, err)

	got, _ := store.Get(ctx, sess.ID)
	require.True(t, got.ExpiresAt.After(oldExpiry))
}

func TestInMemoryStore_ExpiredSession(t *testing.T) {
	store := &InMemoryStore{
		data: make(map[string]*Session),
		ttl:  -time.Second,
	}
	ctx := context.Background()

	sess, _ := store.Create(ctx, domain.NewUserID(), nil, nil, "", "")
	_, err := store.Get(ctx, sess.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expired")
}

func TestInMemoryStore_CleanupExpired(t *testing.T) {
	store := &InMemoryStore{
		data: make(map[string]*Session),
		ttl:  -time.Second,
	}
	ctx := context.Background()

	store.Create(ctx, domain.NewUserID(), nil, nil, "", "")
	store.Create(ctx, domain.NewUserID(), nil, nil, "", "")

	count := store.CleanupExpired()
	require.Equal(t, 2, count)
}

func TestMiddleware_Authenticate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := NewInMemoryStore(60)
	cfg := config.SessionConfig{CookieName: "session"}
	mw := NewMiddleware(store, cfg)

	ctx := context.Background()
	userID := domain.NewUserID()
	sess, _ := store.Create(ctx, userID, []string{"admin"}, []string{"users:read"}, "Mozilla", "127.0.0.1")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: sess.ID})

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mw.Authenticate()(c)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, userID.String(), c.GetString("user_id"))
	require.Equal(t, sess.ID, c.GetString("session_id"))
	require.Equal(t, []string{"users:read"}, c.GetStringSlice("user_permissions"))
}

func TestMiddleware_Authenticate_MissingCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := NewInMemoryStore(60)
	cfg := config.SessionConfig{CookieName: "session"}
	mw := NewMiddleware(store, cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mw.Authenticate()(c)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMiddleware_Authenticate_InvalidSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := NewInMemoryStore(60)
	cfg := config.SessionConfig{CookieName: "session"}
	mw := NewMiddleware(store, cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "invalid-id"})

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mw.Authenticate()(c)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMiddleware_Authenticate_ExpiredSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &InMemoryStore{
		data: make(map[string]*Session),
		ttl:  -time.Second,
	}
	cfg := config.SessionConfig{CookieName: "session"}
	mw := NewMiddleware(store, cfg)

	ctx := context.Background()
	sess, _ := store.Create(ctx, domain.NewUserID(), nil, nil, "", "")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: sess.ID})

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mw.Authenticate()(c)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMiddleware_SetSession_CookieSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := NewInMemoryStore(60)
	cfg := config.SessionConfig{CookieName: "session", CookieSecure: false, TTLMinutes: 60}
	mw := NewMiddleware(store, cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	sess := &Session{ID: "test-session-id"}
	mw.SetSession(c, sess)

	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, "session", cookies[0].Name)
	require.Equal(t, "test-session-id", cookies[0].Value)
	require.True(t, cookies[0].HttpOnly)
}

func TestMiddleware_ClearSession_CookieCleared(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := NewInMemoryStore(60)
	cfg := config.SessionConfig{CookieName: "session"}
	mw := NewMiddleware(store, cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mw.ClearSession(c)

	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, -1, cookies[0].MaxAge)
}

func TestGetUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	userID := domain.NewUserID()
	c.Set("user_id", userID.String())

	require.Equal(t, userID, GetUserID(c))
}

func TestGetSessionID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("session_id", "sess-123")

	require.Equal(t, "sess-123", GetSessionID(c))
}

func TestGetPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_permissions", []string{"users:read", "users:write"})

	require.Equal(t, []string{"users:read", "users:write"}, GetPermissions(c))
}
