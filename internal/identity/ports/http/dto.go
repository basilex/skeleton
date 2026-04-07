package http

import (
	"github.com/basilex/skeleton/internal/identity/domain"
)

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"StrongPass123!"`
}

// LoginRequest represents user login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"StrongPass123!"`
}

// RefreshRequest represents token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJSUzI1NiIs..."`
}

// TokenResponse represents authentication tokens
type TokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJSUzI1NiIs..."`
	RefreshToken string `json:"refresh_token" example:"019d65d6-de90-7200-b1cf-4f8745597e0a"`
}

// AssignRoleRequest represents role assignment request
type AssignRoleRequest struct {
	RoleID string `json:"role_id" binding:"required" example:"019d65d6-de90-7200-b1cf-4f8745597e0a"`
}

// UserFilterRequest represents user list query parameters
type UserFilterRequest struct {
	Cursor   string `form:"cursor" example:"019d65d6-de90-7200-b1cf-4f8745597e0a"`
	Limit    int    `form:"limit" example:"20"`
	Search   string `form:"search" example:"user@example.com"`
	IsActive *bool  `form:"is_active" example:"true"`
}

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

type ContextKey string

const (
	ContextKeyUserID      ContextKey = "user_id"
	ContextKeyUserRoles   ContextKey = "user_roles"
	ContextKeyPermissions ContextKey = "permissions"
)
