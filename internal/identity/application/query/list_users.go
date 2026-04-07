package query

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/pkg/pagination"
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
	Cursor   string
	Limit    int
	Search   string
	IsActive *bool
}

type UserDTO struct {
	ID        string   `json:"id"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
	IsActive  bool     `json:"is_active"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

func (h *ListUsersHandler) Handle(ctx context.Context, q ListUsersQuery) (pagination.PageResult[UserDTO], error) {
	pq := pagination.PageQuery{
		Cursor: q.Cursor,
		Limit:  q.Limit,
	}
	pq.Normalize()

	filter := domain.UserFilter{
		Search:   q.Search,
		IsActive: q.IsActive,
		Cursor:   pq.Cursor,
		Limit:    pq.Limit,
	}

	page, err := h.users.FindAll(ctx, filter)
	if err != nil {
		return pagination.PageResult[UserDTO]{}, fmt.Errorf("find all users: %w", err)
	}

	allRoles, err := h.roles.FindAll(ctx)
	if err != nil {
		return pagination.PageResult[UserDTO]{}, fmt.Errorf("find all roles: %w", err)
	}

	roleMap := make(map[domain.RoleID]*domain.Role)
	for _, r := range allRoles {
		roleMap[r.ID()] = r
	}

	items := make([]UserDTO, len(page.Items))
	for i, u := range page.Items {
		roleNames := make([]string, 0)
		for _, roleID := range u.Roles() {
			if r, ok := roleMap[roleID]; ok {
				roleNames = append(roleNames, r.Name())
			}
		}
		items[i] = UserDTO{
			ID:        string(u.ID()),
			Email:     u.Email().String(),
			Roles:     roleNames,
			IsActive:  u.IsActive(),
			CreatedAt: u.CreatedAt().Format("2006-01-02T15:04:05Z"),
			UpdatedAt: u.UpdatedAt().Format("2006-01-02T15:04:05Z"),
		}
	}

	return pagination.NewPageResultWithCursor(items, page.NextCursor, page.HasMore, page.Limit), nil
}
