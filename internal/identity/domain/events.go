package domain

import (
	"time"
)

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

type UserRegistered struct {
	UserID    UserID
	Email     Email
	OcurredAt time.Time
}

func (e UserRegistered) EventName() string {
	return "identity.user_registered"
}

func (e UserRegistered) OccurredAt() time.Time {
	return e.OcurredAt
}

func (e UserRegistered) GetUserID() string {
	return string(e.UserID)
}

func (e UserRegistered) GetEmail() string {
	return e.Email.String()
}

type RoleAssigned struct {
	UserID    UserID
	RoleID    RoleID
	OcurredAt time.Time
}

func (e RoleAssigned) EventName() string {
	return "identity.role_assigned"
}

func (e RoleAssigned) OccurredAt() time.Time {
	return e.OcurredAt
}

func (e RoleAssigned) GetUserID() string {
	return string(e.UserID)
}

func (e RoleAssigned) GetRoleID() string {
	return string(e.RoleID)
}

type RoleRevoked struct {
	UserID    UserID
	RoleID    RoleID
	OcurredAt time.Time
}

func (e RoleRevoked) EventName() string {
	return "identity.role_revoked"
}

func (e RoleRevoked) OccurredAt() time.Time {
	return e.OcurredAt
}

func (e RoleRevoked) GetUserID() string {
	return string(e.UserID)
}

func (e RoleRevoked) GetRoleID() string {
	return string(e.RoleID)
}

type UserLoggedIn struct {
	UserID    UserID
	OcurredAt time.Time
}

func (e UserLoggedIn) EventName() string {
	return "identity.login"
}

func (e UserLoggedIn) OccurredAt() time.Time {
	return e.OcurredAt
}

func (e UserLoggedIn) GetUserID() string {
	return string(e.UserID)
}

type UserLoggedOut struct {
	UserID    UserID
	OcurredAt time.Time
}

func (e UserLoggedOut) EventName() string {
	return "identity.logout"
}

func (e UserLoggedOut) OccurredAt() time.Time {
	return e.OcurredAt
}

func (e UserLoggedOut) GetUserID() string {
	return string(e.UserID)
}
