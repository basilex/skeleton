package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMemoryCache_SetGet(t *testing.T) {
	cache := NewMemoryCache(time.Minute)
	ctx := context.Background()

	t.Run("set and get", func(t *testing.T) {
		type testData struct {
			Name string
			Age  int
		}

		data := testData{Name: "test", Age: 25}
		err := cache.Set(ctx, "key1", data, time.Minute)
		require.NoError(t, err)

		var result testData
		err = cache.Get(ctx, "key1", &result)
		require.NoError(t, err)
		require.Equal(t, data, result)
	})

	t.Run("get non-existent key", func(t *testing.T) {
		var result string
		err := cache.Get(ctx, "nonexistent", &result)
		require.Error(t, err)
		require.True(t, IsNotFound(err))
	})
}

func TestMemoryCache_Delete(t *testing.T) {
	cache := NewMemoryCache(time.Minute)
	ctx := context.Background()

	cache.Set(ctx, "key1", "value1", time.Minute)
	cache.Delete(ctx, "key1")

	var result string
	err := cache.Get(ctx, "key1", &result)
	require.Error(t, err)
	require.True(t, IsNotFound(err))
}

func TestMemoryCache_Expiration(t *testing.T) {
	cache := NewMemoryCache(50 * time.Millisecond)
	ctx := context.Background()

	cache.Set(ctx, "key1", "value1", 100*time.Millisecond)

	time.Sleep(150 * time.Millisecond)

	var result string
	err := cache.Get(ctx, "key1", &result)
	require.Error(t, err)
}

func TestMemoryCache_Exists(t *testing.T) {
	cache := NewMemoryCache(time.Minute)
	ctx := context.Background()

	cache.Set(ctx, "key1", "value1", time.Minute)

	exists, err := cache.Exists(ctx, "key1")
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = cache.Exists(ctx, "nonexistent")
	require.NoError(t, err)
	require.False(t, exists)
}
