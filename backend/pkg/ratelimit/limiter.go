package ratelimit

import (
	"context"
	"time"
)

// Limiter defines the interface for rate limiting operations.
type Limiter interface {
	// Allow checks if a request is allowed for the given key.
	Allow(ctx context.Context, key string) (bool, error)

	// AllowN checks if N requests are allowed for the given key.
	AllowN(ctx context.Context, key string, n int) (bool, error)

	// Remaining returns the number of requests remaining in the current window.
	Remaining(ctx context.Context, key string) (int, error)

	// Reset resets the rate limit for the given key.
	Reset(ctx context.Context, key string) error

	// Configure updates the rate limit configuration.
	Configure(config Config) error
}

// Config represents rate limit configuration.
type Config struct {
	Limit     int           // Maximum requests allowed
	Window    time.Duration // Time window
	KeyPrefix string        // Key prefix for namespacing
}

// RateLimitError represents a rate limit error.
type RateLimitError struct {
	Limit     int
	Remaining int
	Reset     time.Time
}

func (e *RateLimitError) Error() string {
	return "rate limit exceeded"
}
