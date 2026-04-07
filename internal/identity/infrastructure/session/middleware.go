package session

import (
	"net/http"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/pkg/config"
	"github.com/gin-gonic/gin"
)

type Middleware struct {
	store Store
	cfg   config.SessionConfig
}

func NewMiddleware(store Store, cfg config.SessionConfig) *Middleware {
	return &Middleware{
		store: store,
		cfg:   cfg,
	}
}

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

func GetUserID(c *gin.Context) domain.UserID {
	v, ok := c.Get("user_id")
	if !ok {
		return ""
	}
	return domain.UserID(v.(string))
}

func GetSessionID(c *gin.Context) string {
	v, ok := c.Get("session_id")
	if !ok {
		return ""
	}
	return v.(string)
}

func GetPermissions(c *gin.Context) []string {
	v, ok := c.Get("user_permissions")
	if !ok {
		return nil
	}
	return v.([]string)
}
