package session

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/pkg/uuid"
)

// InMemoryStore provides an in-memory implementation of the session store.
// Suitable for development and testing; not recommended for production use
// as sessions are lost on restart and not shared across instances.
type InMemoryStore struct {
	mu   sync.RWMutex
	data map[string]*Session
	ttl  time.Duration
}

// NewInMemoryStore creates a new in-memory session store with the specified TTL.
// ttlMinutes specifies the session lifetime in minutes.
func NewInMemoryStore(ttlMinutes int) *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string]*Session),
		ttl:  time.Duration(ttlMinutes) * time.Minute,
	}
}

// Create creates a new session for the specified user with roles and permissions.
// The session is assigned a unique ID and expiration time based on the configured TTL.
func (s *InMemoryStore) Create(ctx context.Context, userID domain.UserID, roles, permissions []string, userAgent, ip string) (*Session, error) {
	now := time.Now().UTC()
	sess := &Session{
		ID:          uuid.NewV7().String(),
		UserID:      userID,
		Roles:       roles,
		Permissions: permissions,
		UserAgent:   userAgent,
		IP:          ip,
		CreatedAt:   now,
		ExpiresAt:   now.Add(s.ttl),
	}

	s.mu.Lock()
	s.data[sess.Key()] = sess
	s.mu.Unlock()

	return sess, nil
}

// Get retrieves a session by ID. Returns an error if the session is not found or expired.
func (s *InMemoryStore) Get(ctx context.Context, id string) (*Session, error) {
	s.mu.RLock()
	sess, ok := s.data[sessionPrefix+id]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session not found")
	}

	if sess.IsExpired() {
		_ = s.Delete(ctx, id)
		return nil, fmt.Errorf("session expired")
	}

	return sess, nil
}

// Delete removes a session from the store by ID.
func (s *InMemoryStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, sessionPrefix+id)
	return nil
}

// DeleteAllForUser removes all sessions associated with the specified user ID.
func (s *InMemoryStore) DeleteAllForUser(ctx context.Context, userID domain.UserID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, sess := range s.data {
		if sess.UserID == userID {
			delete(s.data, key)
		}
	}

	return nil
}

// Touch extends the session expiration time by the configured TTL.
func (s *InMemoryStore) Touch(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := sessionPrefix + id
	sess, ok := s.data[key]
	if !ok {
		return fmt.Errorf("session not found")
	}

	sess.ExpiresAt = time.Now().Add(s.ttl)
	return nil
}

// CleanupExpired removes all expired sessions from the store.
// Returns the number of sessions removed.
func (s *InMemoryStore) CleanupExpired() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	var count int
	for key, sess := range s.data {
		if sess.IsExpired() {
			delete(s.data, key)
			count++
		}
	}
	return count
}

// MarshalJSON serializes all sessions to JSON for debugging or backup purposes.
func (s *InMemoryStore) MarshalJSON() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return json.Marshal(s.data)
}
