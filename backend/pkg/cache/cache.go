package cache

import (
	"context"
	"time"
)

// Cache defines the interface for caching operations.
type Cache interface {
	// Get retrieves a value from the cache and deserializes it into dest.
	// Returns ErrNotFound if the key doesn't exist.
	Get(ctx context.Context, key string, dest interface{}) error

	// Set stores a value in the cache with the given TTL.
	// If TTL is 0, uses default TTL.
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete removes a key from the cache.
	Delete(ctx context.Context, key string) error

	// GetMulti retrieves multiple values from the cache.
	// Returns a map of found keys to values.
	GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error)

	// SetMulti stores multiple values in the cache with the given TTL.
	SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error

	// DeleteMulti removes multiple keys from the cache.
	DeleteMulti(ctx context.Context, keys []string) error

	// DeleteByPattern removes all keys matching the pattern.
	// Pattern uses Redis-style glob matching (* for wildcard).
	DeleteByPattern(ctx context.Context, pattern string) error

	// Exists checks if a key exists in the cache.
	Exists(ctx context.Context, key string) (bool, error)

	// TTL returns the remaining time-to-live for a key.
	// Returns ErrNotFound if the key doesn't exist.
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Clear removes all keys from the cache.
	Clear(ctx context.Context) error
}

// Errors
var (
	ErrNotFound     = &CacheError{Op: "get", Err: "key not found"}
	ErrExpired      = &CacheError{Op: "get", Err: "key expired"}
	ErrInvalidType  = &CacheError{Op: "get", Err: "invalid type"}
	ErrEncodeFailed = &CacheError{Op: "set", Err: "encoding failed"}
	ErrDecodeFailed = &CacheError{Op: "get", Err: "decoding failed"}
)

// CacheError represents a cache error.
type CacheError struct {
	Op  string
	Err string
	Key string
}

func (e *CacheError) Error() string {
	if e.Key != "" {
		return "cache " + e.Op + ": " + e.Err + " (key: " + e.Key + ")"
	}
	return "cache " + e.Op + ": " + e.Err
}

// IsNotFound checks if the error is ErrNotFound.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	cacheErr, ok := err.(*CacheError)
	if !ok {
		return false
	}
	return cacheErr.Err == "key not found"
}
