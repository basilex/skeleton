package domain

import (
	"context"
)

type UserFilter struct {
	Search   string
	IsActive *bool
	Page     int
	PageSize int
}

type UserRepository interface {
	Save(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id UserID) (*User, error)
	FindByEmail(ctx context.Context, email Email) (*User, error)
	FindAll(ctx context.Context, filter UserFilter) ([]*User, int, error)
	Delete(ctx context.Context, id UserID) error
}

type RoleRepository interface {
	Save(ctx context.Context, role *Role) error
	FindByID(ctx context.Context, id RoleID) (*Role, error)
	FindByName(ctx context.Context, name string) (*Role, error)
	FindAll(ctx context.Context) ([]*Role, error)
	FindByIDs(ctx context.Context, ids []RoleID) ([]*Role, error)
}

type TokenService interface {
	GenerateAccessToken(userID UserID, roles []Role) (string, error)
	GenerateRefreshToken() (string, error)
	ValidateAccessToken(token string) (*TokenClaims, error)
	ValidateRefreshToken(token string) (UserID, error)
}

type TokenClaims struct {
	UserID      UserID
	Roles       []string
	Permissions []string
}

type PasswordHasher interface {
	Hash(plainPassword string) (PasswordHash, error)
	Compare(hash PasswordHash, plainPassword string) bool
}

type BcryptHasher struct{}

func (h *BcryptHasher) Hash(plainPassword string) (PasswordHash, error) {
	return NewPasswordHash(plainPassword)
}

func (h *BcryptHasher) Compare(hash PasswordHash, plainPassword string) bool {
	return hash.Matches(plainPassword)
}
