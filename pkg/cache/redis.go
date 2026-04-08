package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache implements Cache interface using Redis.
type RedisCache struct {
	client *redis.Client
	prefix string
}

// NewRedisCache creates a new Redis-backed cache.
func NewRedisCache(client *redis.Client, prefix string) *RedisCache {
	return &RedisCache{
		client: client,
		prefix: prefix,
	}
}

func (c *RedisCache) key(k string) string {
	if c.prefix == "" {
		return k
	}
	return c.prefix + ":" + k
}

func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	fullKey := c.key(key)

	data, err := c.client.Get(ctx, fullKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return &CacheError{Op: "get", Err: "key not found", Key: key}
		}
		return &CacheError{Op: "get", Err: err.Error(), Key: key}
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return &CacheError{Op: "get", Err: "decoding failed", Key: key}
	}

	return nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	fullKey := c.key(key)

	data, err := json.Marshal(value)
	if err != nil {
		return &CacheError{Op: "set", Err: "encoding failed", Key: key}
	}

	if err := c.client.Set(ctx, fullKey, data, ttl).Err(); err != nil {
		return &CacheError{Op: "set", Err: err.Error(), Key: key}
	}

	return nil
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := c.key(key)
	return c.client.Del(ctx, fullKey).Err()
}

func (c *RedisCache) GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	fullKeys := make([]string, len(keys))
	for i, k := range keys {
		fullKeys[i] = c.key(k)
	}

	results, err := c.client.MGet(ctx, fullKeys...).Result()
	if err != nil {
		return nil, err
	}

	output := make(map[string]interface{})
	for i, result := range results {
		if result == nil {
			continue
		}

		str, ok := result.(string)
		if !ok {
			continue
		}

		var value interface{}
		if err := json.Unmarshal([]byte(str), &value); err != nil {
			continue
		}

		output[keys[i]] = value
	}

	return output, nil
}

func (c *RedisCache) SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	pipe := c.client.Pipeline()

	for key, value := range items {
		fullKey := c.key(key)
		data, err := json.Marshal(value)
		if err != nil {
			return &CacheError{Op: "setMulti", Err: "encoding failed", Key: key}
		}
		pipe.Set(ctx, fullKey, data, ttl)
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (c *RedisCache) DeleteMulti(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	fullKeys := make([]string, len(keys))
	for i, k := range keys {
		fullKeys[i] = c.key(k)
	}

	return c.client.Del(ctx, fullKeys...).Err()
}

func (c *RedisCache) DeleteByPattern(ctx context.Context, pattern string) error {
	fullPattern := c.key(pattern)

	var cursor uint64
	for {
		keys, nextCursor, err := c.client.Scan(ctx, cursor, fullPattern, 100).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}

func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := c.key(key)
	count, err := c.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (c *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	fullKey := c.key(key)
	ttl, err := c.client.TTL(ctx, fullKey).Result()
	if err != nil {
		return 0, err
	}

	if ttl < 0 {
		return 0, &CacheError{Op: "ttl", Err: "key not found or no expiration", Key: key}
	}

	return ttl, nil
}

func (c *RedisCache) Clear(ctx context.Context) error {
	pattern := c.key("*")
	return c.DeleteByPattern(ctx, pattern)
}

// Close closes the Redis client (optional, usually managed externally).
func (c *RedisCache) Close() error {
	return c.client.Close()
}
