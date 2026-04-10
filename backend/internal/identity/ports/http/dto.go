package http

import (
	"github.com/basilex/skeleton/internal/identity/domain"
)

// RegisterRequest represents user registration request.
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"StrongPass123!"`
}

// LoginRequest represents user login request.
// It contains credentials required for authenticating an existing user.
type LoginRequest struct {
	Email    string `json:"email" binding:"required" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"StrongPass123!"`
}

// RefreshRequest represents token refresh request.
// It contains the refresh token needed to obtain a new access token.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJSUzI1NiIs..."`
}

// AuthResponse represents the response after successful authentication.
// It includes user information alongside the authentication tokens.
type AuthResponse struct {
	UserID       string   `json:"user_id"`
	Email        string   `json:"email"`
	Roles        []string `json:"roles"`
	IsActive     bool     `json:"is_active"`
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
}

// AssignRoleRequest represents role assignment request.
// It is used to assign a specific role to a user.
type AssignRoleRequest struct {
	RoleID string `json:"role_id" binding:"required" example:"019d65d6-de90-7200-b1cf-4f8745597e0a"`
}

// UserFilterRequest represents user list query parameters.
// It provides filtering and pagination options for listing users.
type UserFilterRequest struct {
	Cursor   string `form:"cursor" example:"019d65d6-de90-7200-b1cf-4f8745597e0a"`
	Limit    int    `form:"limit" example:"20"`
	Search   string `form:"search" example:"user@example.com"`
	IsActive *bool  `form:"is_active" example:"true"`
}

// ToDomain converts the HTTP filter request to domain model.
// It applies default values for optional fields like Limit.
func (r *UserFilterRequest) ToDomain() domain.UserFilter {
	limit := r.Limit
	if limit <= 0 {
		limit = 20
	}
	return domain.UserFilter{
		Search:   r.Search,
		IsActive: r.IsActive,
		Cursor:   r.Cursor,
		Limit:    limit,
	}
}

// ContextKey represents a type-safe key for storing values in gin.Context.
type ContextKey string

// Context key constants used for storing authenticated user information.
const (
	// ContextKeyUserID is the context key for the authenticated user's ID.
	ContextKeyUserID ContextKey = "user_id"
	// ContextKeyUserRoles is the context key for the authenticated user's roles.
	ContextKeyUserRoles ContextKey = "user_roles"
	// ContextKeyPermissions is the context key for the authenticated user's permissions.
	ContextKeyPermissions ContextKey = "permissions"
)
