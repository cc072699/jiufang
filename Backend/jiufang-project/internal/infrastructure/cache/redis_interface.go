package cache

import (
	"context"
	"time"
)

// RedisClientInterface defines the interface for Redis client operations.
type RedisClientInterface interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	Close() error
}

// Ensure RedisClient implements RedisClientInterface
var _ RedisClientInterface = (*RedisClient)(nil)