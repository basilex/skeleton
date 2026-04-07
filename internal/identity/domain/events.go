// Package domain provides domain entities and repository interfaces for the identity module.
// This package contains the core business logic types, value objects, and repository contracts
// for user management, authentication, and authorization.
package domain

import (
	"time"
)

// DomainEvent defines the interface for domain events in the identity module.
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

// UserRegistered is emitted when a new user account is created.
type UserRegistered struct {
	UserID    UserID
	Email     Email
	OcurredAt time.Time
}

// EventName returns the event name for UserRegistered.
func (e UserRegistered) EventName() string {
	return "identity.user_registered"
}

// OccurredAt returns when the UserRegistered event occurred.
func (e UserRegistered) OccurredAt() time.Time {
	return e.OcurredAt
}

// GetUserID returns the user ID as a string.
func (e UserRegistered) GetUserID() string {
	return string(e.UserID)
}

// GetEmail returns the user's email address as a string.
func (e UserRegistered) GetEmail() string {
	return e.Email.String()
}

// RoleAssigned is emitted when a role is assigned to a user.
type RoleAssigned struct {
	UserID    UserID
	RoleID    RoleID
	OcurredAt time.Time
}

// EventName returns the event name for RoleAssigned.
func (e RoleAssigned) EventName() string {
	return "identity.role_assigned"
}

// OccurredAt returns when the RoleAssigned event occurred.
func (e RoleAssigned) OccurredAt() time.Time {
	return e.OcurredAt
}

// GetUserID returns the user ID as a string.
func (e RoleAssigned) GetUserID() string {
	return string(e.UserID)
}

// GetRoleID returns the role ID as a string.
func (e RoleAssigned) GetRoleID() string {
	return string(e.RoleID)
}

// RoleRevoked is emitted when a role is removed from a user.
type RoleRevoked struct {
	UserID    UserID
	RoleID    RoleID
	OcurredAt time.Time
}

// EventName returns the event name for RoleRevoked.
func (e RoleRevoked) EventName() string {
	return "identity.role_revoked"
}

// OccurredAt returns when the RoleRevoked event occurred.
func (e RoleRevoked) OccurredAt() time.Time {
	return e.OcurredAt
}

// GetUserID returns the user ID as a string.
func (e RoleRevoked) GetUserID() string {
	return string(e.UserID)
}

// GetRoleID returns the role ID as a string.
func (e RoleRevoked) GetRoleID() string {
	return string(e.RoleID)
}

// UserLoggedIn is emitted when a user successfully authenticates.
type UserLoggedIn struct {
	UserID    UserID
	OcurredAt time.Time
}

// EventName returns the event name for UserLoggedIn.
func (e UserLoggedIn) EventName() string {
	return "identity.login"
}

// OccurredAt returns when the UserLoggedIn event occurred.
func (e UserLoggedIn) OccurredAt() time.Time {
	return e.OcurredAt
}

// GetUserID returns the user ID as a string.
func (e UserLoggedIn) GetUserID() string {
	return string(e.UserID)
}

// UserLoggedOut is emitted when a user ends their session.
type UserLoggedOut struct {
	UserID    UserID
	OcurredAt time.Time
}

// EventName returns the event name for UserLoggedOut.
func (e UserLoggedOut) EventName() string {
	return "identity.logout"
}

// OccurredAt returns when the UserLoggedOut event occurred.
func (e UserLoggedOut) OccurredAt() time.Time {
	return e.OcurredAt
}

// GetUserID returns the user ID as a string.
func (e UserLoggedOut) GetUserID() string {
	return string(e.UserID)
}
