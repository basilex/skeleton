package ratelimit

import (
	"context"
	"sync"
	"time"
)

// TokenBucket implements rate limiting using token bucket algorithm.
type TokenBucket struct {
	buckets sync.Map
	config  Config
}

type bucket struct {
	tokens     float64
	lastUpdate time.Time
	mu         sync.Mutex
}

// NewTokenBucket creates a new token bucket rate limiter.
func NewTokenBucket(config Config) *TokenBucket {
	return &TokenBucket{
		config: config,
	}
}

func (tb *TokenBucket) Allow(ctx context.Context, key string) (bool, error) {
	return tb.AllowN(ctx, key, 1)
}

func (tb *TokenBucket) AllowN(ctx context.Context, key string, n int) (bool, error) {
	b := tb.getBucket(key)

	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastUpdate).Seconds()

	// Refill tokens based on time elapsed
	rate := float64(tb.config.Limit) / tb.config.Window.Seconds()
	b.tokens += rate * elapsed
	b.tokens = min(b.tokens, float64(tb.config.Limit))
	b.lastUpdate = now

	// Check if enough tokens available
	if b.tokens >= float64(n) {
		b.tokens -= float64(n)
		return true, nil
	}

	return false, nil
}

func (tb *TokenBucket) Remaining(ctx context.Context, key string) (int, error) {
	b := tb.getBucket(key)

	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastUpdate).Seconds()

	rate := float64(tb.config.Limit) / tb.config.Window.Seconds()
	tokens := b.tokens + rate*elapsed
	tokens = min(tokens, float64(tb.config.Limit))

	return int(tokens), nil
}

func (tb *TokenBucket) Reset(ctx context.Context, key string) error {
	tb.buckets.Delete(tb.prefixedKey(key))
	return nil
}

func (tb *TokenBucket) Configure(config Config) error {
	tb.config = config
	return nil
}

func (tb *TokenBucket) getBucket(key string) *bucket {
	prefixed := tb.prefixedKey(key)

	value, _ := tb.buckets.LoadOrStore(prefixed, &bucket{
		tokens:     float64(tb.config.Limit),
		lastUpdate: time.Now(),
	})

	return value.(*bucket)
}

func (tb *TokenBucket) prefixedKey(key string) string {
	if tb.config.KeyPrefix == "" {
		return key
	}
	return tb.config.KeyPrefix + ":" + key
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
