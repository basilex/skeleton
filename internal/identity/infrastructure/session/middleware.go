// Package session provides session management infrastructure implementations.
// This package contains in-memory session storage and HTTP middleware for
// session-based authentication.
package session

import (
	"net/http"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/pkg/config"
	"github.com/gin-gonic/gin"
)

// Middleware provides HTTP middleware for session-based authentication.
// It validates session cookies and injects user context into requests.
type Middleware struct {
	store Store
	cfg   config.SessionConfig
}

// NewMiddleware creates a new session middleware with the provided store and configuration.
func NewMiddleware(store Store, cfg config.SessionConfig) *Middleware {
	return &Middleware{
		store: store,
		cfg:   cfg,
	}
}

// SetSession sets a session cookie in the HTTP response.
func (m *Middleware) SetSession(c *gin.Context, sess *Session) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     m.cfg.CookieName,
		Value:    sess.ID,
		Path:     "/",
		Domain:   m.cfg.CookieDomain,
		Secure:   m.cfg.CookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(m.cfg.TTLMinutes * 60),
	})
}

// ClearSession clears the session cookie from the HTTP response.
func (m *Middleware) ClearSession(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     m.cfg.CookieName,
		Value:    "",
		Path:     "/",
		Domain:   m.cfg.CookieDomain,
		Secure:   m.cfg.CookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

// Authenticate returns a Gin middleware that validates session cookies.
// It retrieves the session from the store and injects user context values.
// Returns 401 Unauthorized if the session is missing or invalid.
func (m *Middleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(m.cfg.CookieName)
		if err != nil || cookie == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":  "unauthorized",
				"detail": "session cookie missing",
			})
			return
		}

		sess, err := m.store.Get(c.Request.Context(), cookie)
		if err != nil {
			m.ClearSession(c)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":  "unauthorized",
				"detail": "invalid or expired session",
			})
			return
		}

		c.Set("session_id", sess.ID)
		c.Set("user_id", string(sess.UserID))
		c.Set("user_roles", sess.Roles)
		c.Set("user_permissions", sess.Permissions)

		_ = m.store.Touch(c.Request.Context(), sess.ID)

		c.Next()
	}
}

// GetUserID extracts the user ID from the Gin context.
// Returns an empty string if not found.
func GetUserID(c *gin.Context) domain.UserID {
	v, ok := c.Get("user_id")
	if !ok {
		return ""
	}
	return domain.UserID(v.(string))
}

// GetSessionID extracts the session ID from the Gin context.
// Returns an empty string if not found.
func GetSessionID(c *gin.Context) string {
	v, ok := c.Get("session_id")
	if !ok {
		return ""
	}
	return v.(string)
}

// GetPermissions extracts the user permissions from the Gin context.
// Returns nil if not found.
func GetPermissions(c *gin.Context) []string {
	v, ok := c.Get("user_permissions")
	if !ok {
		return nil
	}
	return v.([]string)
}
