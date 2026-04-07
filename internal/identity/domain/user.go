package domain

import (
	"time"
)

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

func (u *User) ID() UserID           { return u.id }
func (u *User) Email() Email         { return u.email }
func (u *User) PasswordHash() string { return string(u.passwordHash) }
func (u *User) Roles() []RoleID      { return u.roles }
func (u *User) IsActive() bool       { return u.isActive }
func (u *User) CreatedAt() time.Time { return u.createdAt }
func (u *User) UpdatedAt() time.Time { return u.updatedAt }

func (u *User) CheckPassword(plainPassword string) bool {
	return u.passwordHash.Matches(plainPassword)
}

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

func (u *User) Deactivate() {
	u.isActive = false
	u.updatedAt = time.Now().UTC()
}

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

func (u *User) PullEvents() []DomainEvent {
	events := u.events
	u.events = make([]DomainEvent, 0)
	return events
}

func (u *User) SetPasswordHash(hash PasswordHash) {
	u.passwordHash = hash
	u.updatedAt = time.Now().UTC()
}

func (u *User) SetRoles(roles []RoleID) {
	u.roles = roles
	u.updatedAt = time.Now().UTC()
}

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
