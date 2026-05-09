package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache defines the interface for caching operations
type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	DeletePattern(ctx context.Context, pattern string) error
}

// RedisCache implements Cache interface using Redis
type RedisCache struct {
	client *redis.Client
	prefix string
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(client *redis.Client, prefix string) *RedisCache {
	return &RedisCache{
		client: client,
		prefix: prefix,
	}
}

// buildKey builds a full key with prefix
func (c *RedisCache) buildKey(key string) string {
	if c.prefix == "" {
		return key
	}
	return c.prefix + key
}

// Get retrieves a value from cache and unmarshals it into dest
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	fullKey := c.buildKey(key)
	data, err := c.client.Get(ctx, fullKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return fmt.Errorf("cache get: %w", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("cache unmarshal: %w", err)
	}

	return nil
}

// Set stores a value in cache with TTL
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	fullKey := c.buildKey(key)

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal: %w", err)
	}

	if ttl > 0 {
		if err := c.client.Set(ctx, fullKey, data, ttl).Err(); err != nil {
			return fmt.Errorf("cache set: %w", err)
		}
	} else {
		// No TTL, store indefinitely (use 0 as sentinel)
		if err := c.client.Set(ctx, fullKey, data, 0).Err(); err != nil {
			return fmt.Errorf("cache set: %w", err)
		}
	}

	return nil
}

// Delete removes a key from cache
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := c.buildKey(key)
	if err := c.client.Del(ctx, fullKey).Err(); err != nil {
		return fmt.Errorf("cache delete: %w", err)
	}
	return nil
}

// Exists checks if a key exists in cache
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := c.buildKey(key)
	result, err := c.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, fmt.Errorf("cache exists: %w", err)
	}
	return result > 0, nil
}

// DeletePattern removes all keys matching the pattern
func (c *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	fullPattern := c.buildKey(pattern)
	var cursor uint64
	for {
		keys, nextCursor, err := c.client.Scan(ctx, cursor, fullPattern, 100).Result()
		if err != nil {
			return fmt.Errorf("cache scan: %w", err)
		}

		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("cache delete pattern: %w", err)
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}

// Close closes the Redis client connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// Ping checks if Redis connection is alive
func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}
