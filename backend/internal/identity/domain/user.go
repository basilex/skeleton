// Package domain provides domain entities and repository interfaces for the identity module.
// This package contains the core business logic types, value objects, and repository contracts
// for user management, authentication, and authorization.
package domain

import (
	"time"
)

// User represents a user entity in the identity domain.
// Users have an email, password hash, and can be assigned multiple roles.
type User struct {
	id           UserID
	email        Email
	passwordHash PasswordHash
	roles        []RoleID
	isActive     bool
	createdAt    time.Time
	updatedAt    time.Time
	events       []DomainEvent
}

// NewUser creates a new user with the provided email and password hash.
// The user is created with an active status and emits a UserRegistered event.
func NewUser(email Email, passwordHash PasswordHash) (*User, error) {
	now := time.Now().UTC()
	u := &User{
		id:           NewUserID(),
		email:        email,
		passwordHash: passwordHash,
		roles:        make([]RoleID, 0),
		isActive:     true,
		createdAt:    now,
		updatedAt:    now,
		events:       make([]DomainEvent, 0),
	}
	u.events = append(u.events, UserRegistered{
		UserID:    u.id,
		Email:     u.email,
		OcurredAt: now,
	})
	return u, nil
}

// ID returns the user's unique identifier.
func (u *User) ID() UserID { return u.id }

// Email returns the user's email address.
func (u *User) Email() Email { return u.email }

// PasswordHash returns the user's password hash as a string.
func (u *User) PasswordHash() string { return string(u.passwordHash) }

// Roles returns the list of role IDs assigned to the user.
func (u *User) Roles() []RoleID { return u.roles }

// IsActive returns whether the user account is active.
func (u *User) IsActive() bool { return u.isActive }

// CreatedAt returns the timestamp when the user was created.
func (u *User) CreatedAt() time.Time { return u.createdAt }

// UpdatedAt returns the timestamp when the user was last updated.
func (u *User) UpdatedAt() time.Time { return u.updatedAt }

// CheckPassword verifies if the provided plain text password matches the stored hash.
func (u *User) CheckPassword(plainPassword string) bool {
	return u.passwordHash.Matches(plainPassword)
}

// AssignRole adds a role to the user's role list.
// Returns ErrUserInactive if the user is deactivated, or ErrRoleAlreadyAssigned if the role is already assigned.
func (u *User) AssignRole(roleID RoleID) error {
	if !u.isActive {
		return ErrUserInactive
	}
	for _, r := range u.roles {
		if r == roleID {
			return ErrRoleAlreadyAssigned
		}
	}
	u.roles = append(u.roles, roleID)
	u.updatedAt = time.Now().UTC()
	u.events = append(u.events, RoleAssigned{
		UserID:    u.id,
		RoleID:    roleID,
		OcurredAt: u.updatedAt,
	})
	return nil
}

// RevokeRole removes a role from the user's role list.
// Returns ErrUserInactive if the user is deactivated, or ErrRoleNotAssigned if the role is not assigned.
func (u *User) RevokeRole(roleID RoleID) error {
	if !u.isActive {
		return ErrUserInactive
	}
	found := false
	for i, r := range u.roles {
		if r == roleID {
			u.roles = append(u.roles[:i], u.roles[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return ErrRoleNotAssigned
	}
	u.updatedAt = time.Now().UTC()
	u.events = append(u.events, RoleRevoked{
		UserID:    u.id,
		RoleID:    roleID,
		OcurredAt: u.updatedAt,
	})
	return nil
}

// Deactivate marks the user account as inactive.
func (u *User) Deactivate() {
	u.isActive = false
	u.updatedAt = time.Now().UTC()
}

// HasPermission checks if the user has a specific permission through any of their assigned roles.
func (u *User) HasPermission(permission Permission, roles []Role) bool {
	for _, roleID := range u.roles {
		for _, role := range roles {
			if role.ID() == roleID {
				if role.HasPermission(permission) {
					return true
				}
			}
		}
	}
	return false
}

// PullEvents returns all pending domain events and clears the internal event buffer.
func (u *User) PullEvents() []DomainEvent {
	events := u.events
	u.events = make([]DomainEvent, 0)
	return events
}

// SetPasswordHash updates the user's password hash.
func (u *User) SetPasswordHash(hash PasswordHash) {
	u.passwordHash = hash
	u.updatedAt = time.Now().UTC()
}

// SetRoles replaces the user's role assignments.
func (u *User) SetRoles(roles []RoleID) {
	u.roles = roles
	u.updatedAt = time.Now().UTC()
}

// ReconstituteUser reconstructs a user entity from persisted state.
// This is used by repositories to hydrate user entities from storage.
func ReconstituteUser(id UserID, email Email, passwordHash PasswordHash, roles []RoleID, isActive bool, createdAt, updatedAt time.Time) (*User, error) {
	return &User{
		id:           id,
		email:        email,
		passwordHash: passwordHash,
		roles:        roles,
		isActive:     isActive,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		events:       make([]DomainEvent, 0),
	}, nil
}
