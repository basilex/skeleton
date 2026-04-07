// Package domain provides domain entities and repository interfaces for the identity module.
// This package contains the core business logic types, value objects, and repository contracts
// for user management, authentication, and authorization.
package domain

import (
	"fmt"

	"github.com/basilex/skeleton/pkg/uuid"
)

// UserID is a unique identifier for a user entity.
type UserID string

// NewUserID generates a new unique UserID using UUID v7.
func NewUserID() UserID {
	return UserID(uuid.NewV7().String())
}

// ParseUserID validates and converts a string to UserID.
func ParseUserID(s string) (UserID, error) {
	if _, err := uuid.Parse(s); err != nil {
		return "", fmt.Errorf("invalid user id: %w", err)
	}
	return UserID(s), nil
}

// RoleID is a unique identifier for a role entity.
type RoleID string

// NewRoleID generates a new unique RoleID using UUID v7.
func NewRoleID() RoleID {
	return RoleID(uuid.NewV7().String())
}

// ParseRoleID validates and converts a string to RoleID.
func ParseRoleID(s string) (RoleID, error) {
	if _, err := uuid.Parse(s); err != nil {
		return "", fmt.Errorf("invalid role id: %w", err)
	}
	return RoleID(s), nil
}
