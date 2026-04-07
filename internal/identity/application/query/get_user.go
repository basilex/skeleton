package query

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/identity/domain"
)

type GetUserHandler struct {
	users domain.UserRepository
	roles domain.RoleRepository
}

func NewGetUserHandler(users domain.UserRepository, roles domain.RoleRepository) *GetUserHandler {
	return &GetUserHandler{
		users: users,
		roles: roles,
	}
}

type GetUserQuery struct {
	UserID string
}

func (h *GetUserHandler) Handle(ctx context.Context, q GetUserQuery) (UserDTO, error) {
	userID, err := domain.ParseUserID(q.UserID)
	if err != nil {
		return UserDTO{}, fmt.Errorf("parse user id: %w", err)
	}

	user, err := h.users.FindByID(ctx, userID)
	if err != nil {
		return UserDTO{}, fmt.Errorf("find user: %w", err)
	}

	roleNames, err := h.loadRoleNames(ctx, user.Roles())
	if err != nil {
		return UserDTO{}, fmt.Errorf("load role names: %w", err)
	}

	return UserDTO{
		ID:        string(user.ID()),
		Email:     user.Email().String(),
		Roles:     roleNames,
		IsActive:  user.IsActive(),
		CreatedAt: user.CreatedAt().Format("2006-01-02T15:04:05Z"),
		UpdatedAt: user.UpdatedAt().Format("2006-01-02T15:04:05Z"),
	}, nil
}

func (h *GetUserHandler) loadRoleNames(ctx context.Context, ids []domain.RoleID) ([]string, error) {
	if len(ids) == 0 {
		return []string{}, nil
	}
	rolePtrs, err := h.roles.FindByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("find roles by IDs: %w", err)
	}
	names := make([]string, len(rolePtrs))
	for i, r := range rolePtrs {
		names[i] = r.Name()
	}
	return names, nil
}
