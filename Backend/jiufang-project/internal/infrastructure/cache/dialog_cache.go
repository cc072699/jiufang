package cache

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"jiufang/internal/model/dialog"
)

// DialogCache manages dialog context storage in Redis.
type DialogCache struct {
	redis  RedisClientInterface
	logger *zap.Logger
	ttl    time.Duration
}

// NewDialogCache creates a new dialog cache.
func NewDialogCache(redis RedisClientInterface, ttlMinutes int, logger *zap.Logger) *DialogCache {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &DialogCache{
		redis:  redis,
		logger: logger,
		ttl:    time.Duration(ttlMinutes) * time.Minute,
	}
}

// LoadContext loads dialog context from Redis.
func (c *DialogCache) LoadContext(ctx context.Context, sessionID string) (*dialog.DialogContext, error) {
	key := dialog.ContextKey(sessionID)

	data, err := c.redis.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to load dialog context: %w", err)
	}

	if data == "" {
		// Context not found, return nil (will create new context)
		c.logger.Debug("Dialog context not found in Redis",
			zap.String("session_id", sessionID),
		)
		return nil, nil
	}

	context, err := dialog.FromJSON(data)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("Dialog context loaded from Redis",
		zap.String("session_id", sessionID),
		zap.Int("turn_count", context.TurnCount),
	)

	return context, nil
}

// SaveContext saves dialog context to Redis with TTL.
func (c *DialogCache) SaveContext(ctx context.Context, context *dialog.DialogContext) error {
	key := dialog.ContextKey(context.SessionID)

	data, err := context.ToJSON()
	if err != nil {
		return err
	}

	if err := c.redis.Set(ctx, key, data, c.ttl); err != nil {
		return fmt.Errorf("failed to save dialog context: %w", err)
	}

	c.logger.Debug("Dialog context saved to Redis",
		zap.String("session_id", context.SessionID),
		zap.Int("turn_count", context.TurnCount),
		zap.Duration("ttl", c.ttl),
	)

	return nil
}

// ClearContext clears dialog context from Redis.
func (c *DialogCache) ClearContext(ctx context.Context, sessionID string) error {
	key := dialog.ContextKey(sessionID)

	if err := c.redis.Delete(ctx, key); err != nil {
		return fmt.Errorf("failed to clear dialog context: %w", err)
	}

	c.logger.Info("Dialog context cleared from Redis",
		zap.String("session_id", sessionID),
	)

	return nil
}

// UpdateTTL updates the TTL for dialog context.
func (c *DialogCache) UpdateTTL(ctx context.Context, sessionID string) error {
	key := dialog.ContextKey(sessionID)

	if err := c.redis.Expire(ctx, key, c.ttl); err != nil {
		return err
	}

	c.logger.Debug("Dialog context TTL updated",
		zap.String("session_id", sessionID),
		zap.Duration("ttl", c.ttl),
	)

	return nil
}

// GetTTL returns the remaining TTL for dialog context.
func (c *DialogCache) GetTTL(ctx context.Context, sessionID string) (time.Duration, error) {
	key := dialog.ContextKey(sessionID)

	ttl, err := c.redis.TTL(ctx, key)
	if err != nil {
		return 0, err
	}

	return ttl, nil
}

// Exists checks if dialog context exists in Redis.
func (c *DialogCache) Exists(ctx context.Context, sessionID string) (bool, error) {
	key := dialog.ContextKey(sessionID)

	exists, err := c.redis.Exists(ctx, key)
	if err != nil {
		return false, err
	}

	return exists, nil
}
