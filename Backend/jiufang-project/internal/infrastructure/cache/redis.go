// Package cache implements Redis cache client for dialog context and query results.
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisConfig represents Redis connection configuration.
type RedisConfig struct {
	Address      string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	DialTimeout  int // seconds
	ReadTimeout  int // seconds
	WriteTimeout int // seconds
}

// RedisClient wraps the go-redis client with additional functionality.
type RedisClient struct {
	client *redis.Client
	logger *zap.Logger
}

// NewRedisClient creates a new Redis client.
func NewRedisClient(config *RedisConfig, logger *zap.Logger) (*RedisClient, error) {
	if logger == nil {
		logger = zap.NewNop()
	}

	client := redis.NewClient(&redis.Options{
		Addr:         config.Address,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  time.Duration(config.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(config.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.WriteTimeout) * time.Second,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis client connected successfully",
		zap.String("address", config.Address),
		zap.Int("db", config.DB),
	)

	return &RedisClient{
		client: client,
		logger: logger,
	}, nil
}

// Get retrieves a value from Redis.
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // Key not found, return empty string
		}
		return "", fmt.Errorf("failed to get from Redis: %w", err)
	}
	return val, nil
}

// Set stores a value in Redis with TTL.
func (r *RedisClient) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set to Redis: %w", err)
	}
	r.logger.Debug("Redis set successful",
		zap.String("key", key),
		zap.Duration("ttl", ttl),
	)
	return nil
}

// Delete deletes a key from Redis.
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete from Redis: %w", err)
	}
	r.logger.Debug("Redis delete successful", zap.String("key", key))
	return nil
}

// Exists checks if a key exists in Redis.
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence in Redis: %w", err)
	}
	return count > 0, nil
}

// Expire sets TTL for a key.
func (r *RedisClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if err := r.client.Expire(ctx, key, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set TTL in Redis: %w", err)
	}
	return nil
}

// TTL gets the remaining TTL for a key.
func (r *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL from Redis: %w", err)
	}
	return ttl, nil
}

// Close closes the Redis client connection.
func (r *RedisClient) Close() error {
	if err := r.client.Close(); err != nil {
		return fmt.Errorf("failed to close Redis client: %w", err)
	}
	r.logger.Info("Redis client closed")
	return nil
}

// GetClient returns the underlying Redis client.
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}