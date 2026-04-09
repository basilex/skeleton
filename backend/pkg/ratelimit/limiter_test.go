package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTokenBucket_Allow(t *testing.T) {
	config := Config{
		Limit:     5,
		Window:    time.Minute,
		KeyPrefix: "test",
	}
	limiter := NewTokenBucket(config)
	ctx := context.Background()

	t.Run("allow requests within limit", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			allowed, err := limiter.Allow(ctx, "user:test")
			require.NoError(t, err)
			require.True(t, allowed)
		}
	})

	t.Run("deny requests above limit", func(t *testing.T) {
		allowed, err := limiter.Allow(ctx, "user:test")
		require.NoError(t, err)
		require.False(t, allowed)
	})
}

func TestTokenBucket_Remaining(t *testing.T) {
	config := Config{
		Limit:     10,
		Window:    time.Minute,
		KeyPrefix: "test",
	}
	limiter := NewTokenBucket(config)
	ctx := context.Background()

	remaining, err := limiter.Remaining(ctx, "user:new")
	require.NoError(t, err)
	require.Equal(t, 10, remaining)

	limiter.Allow(ctx, "user:new")
	remaining, err = limiter.Remaining(ctx, "user:new")
	require.NoError(t, err)
	require.Equal(t, 9, remaining)
}

func TestTokenBucket_Reset(t *testing.T) {
	config := Config{
		Limit:     5,
		Window:    time.Minute,
		KeyPrefix: "test",
	}
	limiter := NewTokenBucket(config)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		limiter.Allow(ctx, "user:reset")
	}

	err := limiter.Reset(ctx, "user:reset")
	require.NoError(t, err)

	remaining, err := limiter.Remaining(ctx, "user:reset")
	require.NoError(t, err)
	require.Equal(t, 5, remaining)
}
