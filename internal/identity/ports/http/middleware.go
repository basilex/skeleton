package http

import (
	"strings"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/pkg/apierror"
	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	tokenService domain.TokenService
}

func NewAuthMiddleware(tokenService domain.TokenService) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
	}
}

func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		tokenString := ExtractBearerToken(authHeader)
		if tokenString == "" {
			apierror.RespondError(c, apierror.NewUnauthorized("missing or invalid authorization header", c.Request.URL.Path, getRequestID(c)))
			c.Abort()
			return
		}

		claims, err := m.tokenService.ValidateAccessToken(tokenString)
		if err != nil {
			apierror.RespondError(c, apierror.NewUnauthorized("invalid or expired token", c.Request.URL.Path, getRequestID(c)))
			c.Abort()
			return
		}

		c.Set(string(ContextKeyUserID), string(claims.UserID))
		c.Set(string(ContextKeyUserRoles), claims.Roles)
		c.Set(string(ContextKeyPermissions), claims.Permissions)
		c.Next()
	}
}

type RBACMiddleware struct{}

func NewRBACMiddleware() *RBACMiddleware {
	return &RBACMiddleware{}
}

func (m *RBACMiddleware) Require(requiredPermissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permsVal, exists := c.Get(string(ContextKeyPermissions))
		if !exists {
			apierror.RespondError(c, apierror.NewForbidden("no permissions in context", c.Request.URL.Path, getRequestID(c)))
			c.Abort()
			return
		}

		perms, ok := permsVal.([]string)
		if !ok {
			apierror.RespondError(c, apierror.NewForbidden("invalid permissions in context", c.Request.URL.Path, getRequestID(c)))
			c.Abort()
			return
		}

		for _, required := range requiredPermissions {
			if !hasPermission(perms, required) {
				apierror.RespondError(c, apierror.NewForbidden("insufficient permissions", c.Request.URL.Path, getRequestID(c)))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

func hasPermission(userPerms []string, required string) bool {
	reqPerm, err := domain.NewPermission(required)
	if err != nil {
		return false
	}
	for _, p := range userPerms {
		userPerm, err := domain.NewPermission(p)
		if err != nil {
			continue
		}
		if userPerm.Matches(reqPerm) {
			return true
		}
	}
	return false
}

func ExtractBearerToken(authHeader string) string {
	if len(authHeader) < 7 || !strings.EqualFold(authHeader[:6], "Bearer") {
		return ""
	}
	return strings.TrimSpace(authHeader[6:])
}
