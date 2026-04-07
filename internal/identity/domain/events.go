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
