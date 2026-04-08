package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// SlidingWindow implements rate limiting using Redis sorted sets.
type SlidingWindow struct {
	client *redis.Client
	config Config
}

// NewSlidingWindow creates a new sliding window rate limiter using Redis.
func NewSlidingWindow(client *redis.Client, config Config) *SlidingWindow {
	return &SlidingWindow{
		client: client,
		config: config,
	}
}

func (sw *SlidingWindow) Allow(ctx context.Context, key string) (bool, error) {
	return sw.AllowN(ctx, key, 1)
}

func (sw *SlidingWindow) AllowN(ctx context.Context, key string, n int) (bool, error) {
	prefixed := sw.prefixedKey(key)
	now := time.Now()
	windowStart := now.Add(-sw.config.Window)

	// Use Redis transaction
	tx := sw.client.TxPipeline()

	// Remove old entries outside the window
	tx.ZRemRangeByScore(ctx, prefixed, "-inf", fmt.Sprintf("%d", windowStart.Unix()))

	// Count current entries
	countCmd := tx.ZCard(ctx, prefixed)

	// Execute transaction
	if _, err := tx.Exec(ctx); err != nil {
		return false, err
	}

	// Check if limit exceeded
	count, err := countCmd.Result()
	if err != nil {
		return false, err
	}

	if count+int64(n) > int64(sw.config.Limit) {
		return false, nil
	}

	// Add new entry
	nowNano := now.UnixNano()
	for i := 0; i < n; i++ {
		member := fmt.Sprintf("%s:%d", prefixed, nowNano+int64(i))
		sw.client.ZAdd(ctx, prefixed, redis.Z{
			Score:  float64(now.Unix()),
			Member: member,
		})
	}

	// Set expiration on the key
	sw.client.Expire(ctx, prefixed, sw.config.Window)

	return true, nil
}

func (sw *SlidingWindow) Remaining(ctx context.Context, key string) (int, error) {
	prefixed := sw.prefixedKey(key)
	now := time.Now()
	windowStart := now.Add(-sw.config.Window)

	count, err := sw.client.ZCount(ctx, prefixed, fmt.Sprintf("%d", windowStart.Unix()), fmt.Sprintf("%d", now.Unix())).Result()
	if err != nil && err != redis.Nil {
		return 0, err
	}

	remaining := sw.config.Limit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

func (sw *SlidingWindow) Reset(ctx context.Context, key string) error {
	prefixed := sw.prefixedKey(key)
	return sw.client.Del(ctx, prefixed).Err()
}

func (sw *SlidingWindow) Configure(config Config) error {
	sw.config = config
	return nil
}

func (sw *SlidingWindow) prefixedKey(key string) string {
	if sw.config.KeyPrefix == "" {
		return key
	}
	return sw.config.KeyPrefix + ":" + key
}
