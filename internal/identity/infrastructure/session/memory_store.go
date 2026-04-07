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

type InMemoryStore struct {
	mu   sync.RWMutex
	data map[string]*Session
	ttl  time.Duration
}

func NewInMemoryStore(ttlMinutes int) *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string]*Session),
		ttl:  time.Duration(ttlMinutes) * time.Minute,
	}
}

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

func (s *InMemoryStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, sessionPrefix+id)
	return nil
}

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

func (s *InMemoryStore) MarshalJSON() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return json.Marshal(s.data)
}
