package query

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/identity/domain"
)

type ListUsersHandler struct {
	users domain.UserRepository
	roles domain.RoleRepository
}

func NewListUsersHandler(users domain.UserRepository, roles domain.RoleRepository) *ListUsersHandler {
	return &ListUsersHandler{
		users: users,
		roles: roles,
	}
}

type ListUsersQuery struct {
	Page     int
	PageSize int
	Search   string
	IsActive *bool
}

type ListUsersResult struct {
	Users []UserDTO `json:"users"`
	Total int       `json:"total"`
}

func (h *ListUsersHandler) Handle(ctx context.Context, q ListUsersQuery) (ListUsersResult, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = 20
	}

	filter := domain.UserFilter{
		Search:   q.Search,
		IsActive: q.IsActive,
		Page:     q.Page,
		PageSize: q.PageSize,
	}

	users, total, err := h.users.FindAll(ctx, filter)
	if err != nil {
		return ListUsersResult{}, fmt.Errorf("find all users: %w", err)
	}

	allRoles, err := h.roles.FindAll(ctx)
	if err != nil {
		return ListUsersResult{}, fmt.Errorf("find all roles: %w", err)
	}

	roleMap := make(map[domain.RoleID]*domain.Role)
	for _, r := range allRoles {
		roleMap[r.ID()] = r
	}

	result := make([]UserDTO, len(users))
	for i, u := range users {
		roleNames := make([]string, 0)
		for _, roleID := range u.Roles() {
			if r, ok := roleMap[roleID]; ok {
				roleNames = append(roleNames, r.Name())
			}
		}
		result[i] = UserDTO{
			ID:        string(u.ID()),
			Email:     u.Email().String(),
			Roles:     roleNames,
			IsActive:  u.IsActive(),
			CreatedAt: u.CreatedAt().Format("2006-01-02T15:04:05Z"),
			UpdatedAt: u.UpdatedAt().Format("2006-01-02T15:04:05Z"),
		}
	}

	return ListUsersResult{
		Users: result,
		Total: total,
	}, nil
}
