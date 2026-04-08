// Package domain provides domain entities and repository interfaces for the identity module.
// This package contains the core business logic types, value objects, and repository contracts
// for user management, authentication, and authorization.
package domain

import (
	"fmt"

	"github.com/basilex/skeleton/pkg/uuid"
)

// UserID is a unique identifier for a user entity.
type UserID uuid.UUID

// NewUserID generates a new unique UserID using UUID v7.
func NewUserID() UserID {
	return UserID(uuid.NewV7())
}

// ParseUserID validates and converts a string to UserID.
func ParseUserID(s string) (UserID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return UserID{}, fmt.Errorf("invalid user id: %w", err)
	}
	return UserID(u), nil
}

// String returns the string representation of UserID.
func (id UserID) String() string {
	return uuid.UUID(id).String()
}

// RoleID is a unique identifier for a role entity.
type RoleID uuid.UUID

// NewRoleID generates a new unique RoleID using UUID v7.
func NewRoleID() RoleID {
	return RoleID(uuid.NewV7())
}

// ParseRoleID validates and converts a string to RoleID.
func ParseRoleID(s string) (RoleID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return RoleID{}, fmt.Errorf("invalid role id: %w", err)
	}
	return RoleID(u), nil
}

// String returns the string representation of RoleID.
func (id RoleID) String() string {
	return uuid.UUID(id).String()
}
