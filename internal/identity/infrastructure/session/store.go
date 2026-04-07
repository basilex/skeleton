package session

import (
	"context"
	"time"

	"github.com/basilex/skeleton/internal/identity/domain"
)

const sessionPrefix = "session:"

type Store interface {
	Create(ctx context.Context, userID domain.UserID, roles, permissions []string, userAgent, ip string) (*Session, error)
	Get(ctx context.Context, id string) (*Session, error)
	Delete(ctx context.Context, id string) error
	DeleteAllForUser(ctx context.Context, userID domain.UserID) error
	Touch(ctx context.Context, id string) error
}

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

func (s *Session) Key() string {
	return sessionPrefix + s.ID
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
