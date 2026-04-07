// Package session provides session management infrastructure implementations.
// This package contains in-memory session storage and HTTP middleware for
// session-based authentication.
package session

import (
	"context"
	"time"

	"github.com/basilex/skeleton/internal/identity/domain"
)

const sessionPrefix = "session:"

// Store defines the interface for session storage operations.
// Implementations must handle session lifecycle including creation, retrieval,
// deletion, and touch operations for session refresh.
type Store interface {
	Create(ctx context.Context, userID domain.UserID, roles, permissions []string, userAgent, ip string) (*Session, error)
	Get(ctx context.Context, id string) (*Session, error)
	Delete(ctx context.Context, id string) error
	DeleteAllForUser(ctx context.Context, userID domain.UserID) error
	Touch(ctx context.Context, id string) error
}

// Session represents a user session with authentication and authorization data.
// It contains the session ID, user identity, roles, permissions, and metadata.
type Session struct {
	ID          string        `json:"id"`
	UserID      domain.UserID `json:"user_id"`
	Roles       []string      `json:"roles"`
	Permissions []string      `json:"permissions"`
	UserAgent   string        `json:"user_agent"`
	IP          string        `json:"ip"`
	CreatedAt   time.Time     `json:"created_at"`
	ExpiresAt   time.Time     `json:"expires_at"`
}

// Key returns the storage key for the session, prefixed with "session:".
func (s *Session) Key() string {
	return sessionPrefix + s.ID
}

// IsExpired returns true if the session has passed its expiration time.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
