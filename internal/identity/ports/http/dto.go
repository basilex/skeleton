package http

import (
	"github.com/basilex/skeleton/internal/identity/domain"
)

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AssignRoleRequest struct {
	RoleID string `json:"role_id" binding:"required"`
}

type UserFilterRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Search   string `form:"search"`
	IsActive *bool  `form:"is_active"`
}

func (r *UserFilterRequest) ToDomain() domain.UserFilter {
	page := r.Page
	if page < 1 {
		page = 1
	}
	pageSize := r.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	return domain.UserFilter{
		Search:   r.Search,
		IsActive: r.IsActive,
		Page:     page,
		PageSize: pageSize,
	}
}

type ContextKey string

const (
	ContextKeyUserID      ContextKey = "user_id"
	ContextKeyUserRoles   ContextKey = "user_roles"
	ContextKeyPermissions ContextKey = "permissions"
)
