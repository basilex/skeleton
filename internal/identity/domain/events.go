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
	return e.UserID.String()
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
	return e.UserID.String()
}

// GetRoleID returns the role ID as a string.
func (e RoleAssigned) GetRoleID() string {
	return e.RoleID.String()
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
	return e.UserID.String()
}

// GetRoleID returns the role ID as a string.
func (e RoleRevoked) GetRoleID() string {
	return e.RoleID.String()
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
	return e.UserID.String()
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
	return e.UserID.String()
}

// SessionCreated is emitted when a new session is created.
type SessionCreated struct {
	SessionID  SessionID
	UserID     UserID
	DeviceType string
	UserAgent  string
	IPAddress  string
	ExpiresAt  time.Time
	occurredAt time.Time
}

// EventName returns the event name for SessionCreated.
func (e SessionCreated) EventName() string {
	return "identity.session_created"
}

// OccurredAt returns when the SessionCreated event occurred.
func (e SessionCreated) OccurredAt() time.Time {
	return e.occurredAt
}

// GetSessionID returns the session ID as a string.
func (e SessionCreated) GetSessionID() string {
	return e.SessionID.String()
}

// GetUserID returns the user ID as a string.
func (e SessionCreated) GetUserID() string {
	return e.UserID.String()
}

// SessionRefreshed is emitted when a session is refreshed.
type SessionRefreshed struct {
	SessionID  SessionID
	UserID     UserID
	ExpiresAt  time.Time
	occurredAt time.Time
}

// EventName returns the event name for SessionRefreshed.
func (e SessionRefreshed) EventName() string {
	return "identity.session_refreshed"
}

// OccurredAt returns when the SessionRefreshed event occurred.
func (e SessionRefreshed) OccurredAt() time.Time {
	return e.occurredAt
}

// GetSessionID returns the session ID as a string.
func (e SessionRefreshed) GetSessionID() string {
	return e.SessionID.String()
}

// GetUserID returns the user ID as a string.
func (e SessionRefreshed) GetUserID() string {
	return e.UserID.String()
}

// SessionRevoked is emitted when a session is revoked.
type SessionRevoked struct {
	SessionID  SessionID
	UserID     UserID
	Reason     string
	occurredAt time.Time
}

// EventName returns the event name for SessionRevoked.
func (e SessionRevoked) EventName() string {
	return "identity.session_revoked"
}

// OccurredAt returns when the SessionRevoked event occurred.
func (e SessionRevoked) OccurredAt() time.Time {
	return e.occurredAt
}

// GetSessionID returns the session ID as a string.
func (e SessionRevoked) GetSessionID() string {
	return e.SessionID.String()
}

// GetUserID returns the user ID as a string.
func (e SessionRevoked) GetUserID() string {
	return e.UserID.String()
}

// SessionExpired is emitted when a session expires.
type SessionExpired struct {
	SessionID  SessionID
	UserID     UserID
	occurredAt time.Time
}

// EventName returns the event name for SessionExpired.
func (e SessionExpired) EventName() string {
	return "identity.session_expired"
}

// OccurredAt returns when the SessionExpired event occurred.
func (e SessionExpired) OccurredAt() time.Time {
	return e.occurredAt
}

// GetSessionID returns the session ID as a string.
func (e SessionExpired) GetSessionID() string {
	return e.SessionID.String()
}

// GetUserID returns the user ID as a string.
func (e SessionExpired) GetUserID() string {
	return e.UserID.String()
}

// SessionLoggedOut is emitted when a user logs out.
type SessionLoggedOut struct {
	SessionID  SessionID
	UserID     UserID
	occurredAt time.Time
}

// EventName returns the event name for SessionLoggedOut.
func (e SessionLoggedOut) EventName() string {
	return "identity.session_logged_out"
}

// OccurredAt returns when the SessionLoggedOut event occurred.
func (e SessionLoggedOut) OccurredAt() time.Time {
	return e.occurredAt
}

// GetSessionID returns the session ID as a string.
func (e SessionLoggedOut) GetSessionID() string {
	return e.SessionID.String()
}

// GetUserID returns the user ID as a string.
func (e SessionLoggedOut) GetUserID() string {
	return e.UserID.String()
}
