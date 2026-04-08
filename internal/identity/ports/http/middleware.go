// Package http provides HTTP middleware for authentication and authorization.
// This package implements authentication middleware that validates JWT tokens
// and RBAC middleware that enforces permission-based access control.
package http

import (
	"strings"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/pkg/apierror"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware provides authentication middleware for validating JWT access tokens.
// It extracts bearer tokens from the Authorization header and validates them
// using the token service.
type AuthMiddleware struct {
	tokenService domain.TokenService
}

// NewAuthMiddleware creates a new authentication middleware with the given token service.
func NewAuthMiddleware(tokenService domain.TokenService) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
	}
}

// Authenticate returns a Gin middleware that validates JWT bearer tokens.
// It extracts the token from the Authorization header, validates it,
// and stores the user claims (ID, roles, permissions) in the context.
// If the token is missing or invalid, it returns a 401 Unauthorized response.
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

		c.Set(string(ContextKeyUserID), claims.UserID.String())
		c.Set(string(ContextKeyUserRoles), claims.Roles)
		c.Set(string(ContextKeyPermissions), claims.Permissions)
		c.Next()
	}
}

// RBACMiddleware provides role-based access control middleware.
// It validates that authenticated users have the required permissions
// to access protected resources.
type RBACMiddleware struct{}

// NewRBACMiddleware creates a new RBAC middleware instance.
func NewRBACMiddleware() *RBACMiddleware {
	return &RBACMiddleware{}
}

// Require returns a Gin middleware that checks if the authenticated user
// has all the specified permissions. The middleware expects authenticated
// user permissions to be stored in the context from a previous authentication step.
// Returns 403 Forbidden if the user lacks required permissions.
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

// hasPermission checks if any of the user's permissions matches the required permission.
// It uses domain-level permission matching to support wildcard permissions.
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

// ExtractBearerToken extracts the bearer token from an Authorization header value.
// Returns an empty string if the header is not a valid Bearer token.
func ExtractBearerToken(authHeader string) string {
	if len(authHeader) < 7 || !strings.EqualFold(authHeader[:6], "Bearer") {
		return ""
	}
	return strings.TrimSpace(authHeader[6:])
}
