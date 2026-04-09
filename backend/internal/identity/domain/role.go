// Package domain provides domain entities and repository interfaces for the identity module.
// This package contains the core business logic types, value objects, and repository contracts
// for user management, authentication, and authorization.
package domain

import (
	"time"
)

// Role represents a role entity in the identity domain.
// Roles contain a collection of permissions that can be assigned to users.
type Role struct {
	id          RoleID
	name        string
	description string
	permissions []Permission
	createdAt   time.Time
}

// NewRole creates a new role with the provided name, description, and permissions.
func NewRole(name, description string, permissions []Permission) (*Role, error) {
	if name == "" {
		return nil, ErrRoleNotFound
	}
	r := &Role{
		id:          NewRoleID(),
		name:        name,
		description: description,
		permissions: permissions,
		createdAt:   time.Now().UTC(),
	}
	return r, nil
}

// NewRoleWithID creates a role with a specific ID (for reconstitution from persistence).
func NewRoleWithID(id RoleID, name, description string, permissions []Permission, createdAt time.Time) (*Role, error) {
	if name == "" {
		return nil, ErrRoleNotFound
	}
	return &Role{
		id:          id,
		name:        name,
		description: description,
		permissions: permissions,
		createdAt:   createdAt,
	}, nil
}

// ID returns the role's unique identifier.
func (r *Role) ID() RoleID { return r.id }

// Name returns the role's name.
func (r *Role) Name() string { return r.name }

// Description returns the role's description.
func (r *Role) Description() string { return r.description }

// Permissions returns the list of permissions assigned to the role.
func (r *Role) Permissions() []Permission { return r.permissions }

// CreatedAt returns the timestamp when the role was created.
func (r *Role) CreatedAt() time.Time { return r.createdAt }

// HasPermission checks if the role grants a specific permission.
func (r *Role) HasPermission(p Permission) bool {
	for _, rp := range r.permissions {
		if rp.Matches(p) {
			return true
		}
	}
	return false
}

// AddPermission adds a new permission to the role.
func (r *Role) AddPermission(p Permission) error {
	if _, err := NewPermission(p.String()); err != nil {
		return err
	}
	r.permissions = append(r.permissions, p)
	return nil
}
