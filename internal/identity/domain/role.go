package domain

import (
	"time"

	"github.com/basilex/skeleton/pkg/uuid"
)

type Role struct {
	id          RoleID
	name        string
	description string
	permissions []Permission
	createdAt   time.Time
}

func NewRole(name, description string, permissions []Permission) (*Role, error) {
	if name == "" {
		return nil, ErrRoleNotFound
	}
	r := &Role{
		id:          RoleID(uuid.NewV7().String()),
		name:        name,
		description: description,
		permissions: permissions,
		createdAt:   time.Now().UTC(),
	}
	return r, nil
}

func (r *Role) ID() RoleID                { return r.id }
func (r *Role) Name() string              { return r.name }
func (r *Role) Description() string       { return r.description }
func (r *Role) Permissions() []Permission { return r.permissions }
func (r *Role) CreatedAt() time.Time      { return r.createdAt }

func (r *Role) HasPermission(p Permission) bool {
	for _, rp := range r.permissions {
		if rp.Matches(p) {
			return true
		}
	}
	return false
}

func (r *Role) AddPermission(p Permission) error {
	if _, err := NewPermission(p.String()); err != nil {
		return err
	}
	r.permissions = append(r.permissions, p)
	return nil
}
