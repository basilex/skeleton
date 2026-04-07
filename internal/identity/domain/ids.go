package domain

import (
	"fmt"

	"github.com/basilex/skeleton/pkg/uuid"
)

type UserID string

func NewUserID() UserID {
	return UserID(uuid.NewV7().String())
}

func ParseUserID(s string) (UserID, error) {
	if _, err := uuid.Parse(s); err != nil {
		return "", fmt.Errorf("invalid user id: %w", err)
	}
	return UserID(s), nil
}

type RoleID string

func NewRoleID() RoleID {
	return RoleID(uuid.NewV7().String())
}

func ParseRoleID(s string) (RoleID, error) {
	if _, err := uuid.Parse(s); err != nil {
		return "", fmt.Errorf("invalid role id: %w", err)
	}
	return RoleID(s), nil
}
