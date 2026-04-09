// Package domain provides domain entities and repository interfaces for the identity module.
// This package contains the core business logic types, value objects, and repository contracts
// for user management, authentication, and authorization.
package domain

import (
	"context"

	"github.com/basilex/skeleton/pkg/pagination"
)

// UserFilter provides filtering options for querying users.
type UserFilter struct {
	Search   string
	IsActive *bool
	Cursor   string
	Limit    int
}

// UserRepository defines the contract for user persistence operations.
type UserRepository interface {
	Save(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id UserID) (*User, error)
	FindByEmail(ctx context.Context, email Email) (*User, error)
	FindAll(ctx context.Context, filter UserFilter) (pagination.PageResult[*User], error)
	Delete(ctx context.Context, id UserID) error
}

// RoleRepository defines the contract for role persistence operations.
type RoleRepository interface {
	Save(ctx context.Context, role *Role) error
	FindByID(ctx context.Context, id RoleID) (*Role, error)
	FindByName(ctx context.Context, name string) (*Role, error)
	FindAll(ctx context.Context) ([]*Role, error)
	FindByIDs(ctx context.Context, ids []RoleID) ([]*Role, error)
}

// TokenService defines the contract for JWT token generation and validation.
type TokenService interface {
	GenerateAccessToken(userID UserID, roles []Role) (string, error)
	GenerateRefreshToken() (string, error)
	ValidateAccessToken(token string) (*TokenClaims, error)
	ValidateRefreshToken(token string) (UserID, error)
}

// TokenClaims represents the claims extracted from a validated access token.
type TokenClaims struct {
	UserID      UserID
	Roles       []string
	Permissions []string
}

// PasswordHasher defines the contract for password hashing and comparison.
type PasswordHasher interface {
	Hash(plainPassword string) (PasswordHash, error)
	Compare(hash PasswordHash, plainPassword string) bool
}

// BcryptHasher implements PasswordHasher using the bcrypt algorithm.
type BcryptHasher struct{}

// Hash creates a bcrypt hash from the plain text password.
func (h *BcryptHasher) Hash(plainPassword string) (PasswordHash, error) {
	return NewPasswordHash(plainPassword)
}

// Compare checks if the plain text password matches the bcrypt hash.
func (h *BcryptHasher) Compare(hash PasswordHash, plainPassword string) bool {
	return hash.Matches(plainPassword)
}

// SessionRepository defines the contract for session persistence operations.
type SessionRepository interface {
	Save(ctx context.Context, session *Session) error
	FindByID(ctx context.Context, id SessionID) (*Session, error)
	FindByUserID(ctx context.Context, userID UserID) ([]*Session, error)
	FindActiveByUserID(ctx context.Context, userID UserID) ([]*Session, error)
	Delete(ctx context.Context, id SessionID) error
	DeleteByUserID(ctx context.Context, userID UserID) error
	DeleteExpired(ctx context.Context) (int64, error)
}

// PreferencesRepository defines the contract for user preferences persistence operations.
type PreferencesRepository interface {
	Save(ctx context.Context, prefs *UserPreferences) error
	FindByID(ctx context.Context, id PreferencesID) (*UserPreferences, error)
	FindByUserID(ctx context.Context, userID UserID) (*UserPreferences, error)
	Delete(ctx context.Context, id PreferencesID) error
	DeleteByUserID(ctx context.Context, userID UserID) error
}
