package cache

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

// MemoryCache implements Cache interface using sync.Map.
type MemoryCache struct {
	store  sync.Map
	ttlMap sync.Map
	stopCh chan struct{}
}

type cacheItem struct {
	value     []byte
	expiresAt time.Time
}

// NewMemoryCache creates a new in-memory cache with optional TTL cleanup.
func NewMemoryCache(cleanupInterval time.Duration) *MemoryCache {
	c := &MemoryCache{
		stopCh: make(chan struct{}),
	}

	if cleanupInterval > 0 {
		go c.cleanup(cleanupInterval)
	}

	return c
}

func (c *MemoryCache) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			c.store.Range(func(key, value interface{}) bool {
				if item, ok := value.(*cacheItem); ok {
					if !item.expiresAt.IsZero() && now.After(item.expiresAt) {
						c.store.Delete(key)
						c.ttlMap.Delete(key)
					}
				}
				return true
			})
		case <-c.stopCh:
			return
		}
	}
}

func (c *MemoryCache) Get(ctx context.Context, key string, dest interface{}) error {
	value, ok := c.store.Load(key)
	if !ok {
		return &CacheError{Op: "get", Err: "key not found", Key: key}
	}

	item, ok := value.(*cacheItem)
	if !ok {
		return &CacheError{Op: "get", Err: "invalid type", Key: key}
	}

	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		c.store.Delete(key)
		c.ttlMap.Delete(key)
		return &CacheError{Op: "get", Err: "key expired", Key: key}
	}

	if err := json.Unmarshal(item.value, dest); err != nil {
		return &CacheError{Op: "get", Err: "decoding failed", Key: key}
	}

	return nil
}

func (c *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return &CacheError{Op: "set", Err: "encoding failed", Key: key}
	}

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	c.store.Store(key, &cacheItem{
		value:     data,
		expiresAt: expiresAt,
	})

	return nil
}

func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.store.Delete(key)
	c.ttlMap.Delete(key)
	return nil
}

func (c *MemoryCache) GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, key := range keys {
		var value interface{}
		if err := c.Get(ctx, key, &value); err == nil {
			result[key] = value
		}
	}

	return result, nil
}

func (c *MemoryCache) SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	for key, value := range items {
		if err := c.Set(ctx, key, value, ttl); err != nil {
			return err
		}
	}
	return nil
}

func (c *MemoryCache) DeleteMulti(ctx context.Context, keys []string) error {
	for _, key := range keys {
		c.Delete(ctx, key)
	}
	return nil
}

func (c *MemoryCache) DeleteByPattern(ctx context.Context, pattern string) error {
	// Simple glob matching: prefix*
	// For production, use proper glob matching library

	prefix := pattern
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix = pattern[:len(pattern)-1]
	}

	c.store.Range(func(key, value interface{}) bool {
		keyStr, ok := key.(string)
		if !ok {
			return true
		}

		if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
			if len(keyStr) >= len(prefix) && keyStr[:len(prefix)] == prefix {
				c.Delete(ctx, keyStr)
			}
		} else {
			if keyStr == pattern {
				c.Delete(ctx, keyStr)
			}
		}

		return true
	})

	return nil
}

func (c *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	_, ok := c.store.Load(key)
	if !ok {
		return false, nil
	}

	// Check if expired
	value, ok := c.store.Load(key)
	if !ok {
		return false, nil
	}

	item, ok := value.(*cacheItem)
	if !ok {
		return false, nil
	}

	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		c.Delete(ctx, key)
		return false, nil
	}

	return true, nil
}

func (c *MemoryCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	value, ok := c.store.Load(key)
	if !ok {
		return 0, &CacheError{Op: "ttl", Err: "key not found", Key: key}
	}

	item, ok := value.(*cacheItem)
	if !ok {
		return 0, &CacheError{Op: "ttl", Err: "invalid type", Key: key}
	}

	if item.expiresAt.IsZero() {
		return 0, nil // No expiration
	}

	ttl := time.Until(item.expiresAt)
	if ttl < 0 {
		c.Delete(ctx, key)
		return 0, &CacheError{Op: "ttl", Err: "key expired", Key: key}
	}

	return ttl, nil
}

func (c *MemoryCache) Clear(ctx context.Context) error {
	c.store.Range(func(key, value interface{}) bool {
		c.store.Delete(key)
		return true
	})
	return nil
}

// Close stops the cleanup goroutine.
func (c *MemoryCache) Close() error {
	close(c.stopCh)
	return nil
}

// Type returns the cache type identifier.
func (c *MemoryCache) Type() string {
	return "memory"
}
