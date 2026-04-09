// Package query provides query handlers for reading identity data.
// This package implements the query side of CQRS for user-related operations,
// handling read-only requests that return data transfer objects without modifying state.
package query

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/identity/domain"
)

// GetUserHandler handles queries to retrieve a single user by ID.
// It fetches user details and resolves associated role names.
type GetUserHandler struct {
	users domain.UserRepository
	roles domain.RoleRepository
}

// NewGetUserHandler creates a new GetUserHandler with the required repositories.
func NewGetUserHandler(users domain.UserRepository, roles domain.RoleRepository) *GetUserHandler {
	return &GetUserHandler{
		users: users,
		roles: roles,
	}
}

// GetUserQuery represents a query to retrieve a specific user by their ID.
type GetUserQuery struct {
	UserID string
}

// Handle executes the GetUserQuery and returns the user data transfer object.
// It validates the user ID, retrieves the user from the repository, and resolves
// role names for a complete user representation.
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
		ID:        user.ID().String(),
		Email:     user.Email().String(),
		Roles:     roleNames,
		IsActive:  user.IsActive(),
		CreatedAt: user.CreatedAt().Format("2006-01-02T15:04:05Z"),
		UpdatedAt: user.UpdatedAt().Format("2006-01-02T15:04:05Z"),
	}, nil
}

// loadRoleNames resolves role IDs to their string names.
// It fetches the role entities from the repository and extracts their names.
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
