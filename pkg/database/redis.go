package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis 配置
type Redis struct {
	Addr            string
	Password        string
	DB              int
	PoolSize        int
	MinIdleConns    int
	MaxRetries      int
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	PoolTimeout     time.Duration
	ConnMaxLifetime time.Duration
	Prefix          string
}

// DefaultRedisConfig returns default Redis configuration
func DefaultRedisConfig() Redis {
	return Redis{
		PoolSize:        20,
		MinIdleConns:    5,
		MaxRetries:      3,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		PoolTimeout:     4 * time.Second,
		ConnMaxLifetime: 1 * time.Hour,
	}
}

// NewRedis 初始化 Redis 客户端
func NewRedis(cfg Redis) (*redis.Client, error) {
	// Apply defaults if not set
	if cfg.PoolSize == 0 {
		cfg.PoolSize = 20
	}
	if cfg.MinIdleConns == 0 {
		cfg.MinIdleConns = 5
	}
	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = 5 * time.Second
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 3 * time.Second
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 3 * time.Second
	}
	if cfg.PoolTimeout == 0 {
		cfg.PoolTimeout = 4 * time.Second
	}
	if cfg.ConnMaxLifetime == 0 {
		cfg.ConnMaxLifetime = 1 * time.Hour
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}

	client := redis.NewClient(&redis.Options{
		Addr:            cfg.Addr,
		Password:        cfg.Password,
		DB:              cfg.DB,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		MaxRetries:      cfg.MaxRetries,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		PoolTimeout:     cfg.PoolTimeout,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}

// RedisPoolStats returns connection pool statistics
type RedisPoolStats struct {
	Active   int
	Idle     int
	Waiters  int
	Stale    int
	TotalCnt int
}

// GetRedisPoolStats returns Redis connection pool statistics
func GetRedisPoolStats(client *redis.Client) RedisPoolStats {
	stats := client.PoolStats()
	return RedisPoolStats{
		Active:   int(stats.TotalConns - stats.IdleConns),
		Idle:     int(stats.IdleConns),
		Waiters:  0, // Not available in redis.PoolStats
		Stale:    int(stats.StaleConns),
		TotalCnt: int(stats.TotalConns),
	}
}
