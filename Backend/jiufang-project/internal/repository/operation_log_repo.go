// Package repository implements the operation log repository.
package repository

import (
	"context"
	"errors"
	"time"

	"jiufang/internal/model/audit"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// OperationLogRepository implements OperationLogRepositoryInterface using GORM.
type OperationLogRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewOperationLogRepository creates a new OperationLogRepository instance.
func NewOperationLogRepository(db *gorm.DB, logger *zap.Logger) *OperationLogRepository {
	return &OperationLogRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new operation log entry.
func (r *OperationLogRepository) Create(ctx context.Context, log *audit.OperationLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		r.logger.Error("failed to create operation log",
			zap.Int64("snowflake_id", log.SnowflakeID),
			zap.String("operation_type", string(log.OperationType)),
			zap.Error(err),
		)
		return err
	}

	r.logger.Info("operation log created successfully",
		zap.Int64("snowflake_id", log.SnowflakeID),
		zap.String("operation_type", string(log.OperationType)),
	)

	return nil
}

// GetByID retrieves an operation log by its primary key ID.
func (r *OperationLogRepository) GetByID(ctx context.Context, id uint) (*audit.OperationLog, error) {
	var log audit.OperationLog
	if err := r.db.WithContext(ctx).First(&log, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("failed to get operation log by ID",
			zap.Uint("id", id),
			zap.Error(err),
		)
		return nil, err
	}

	return &log, nil
}

// GetBySnowflakeID retrieves an operation log by its snowflake ID.
func (r *OperationLogRepository) GetBySnowflakeID(ctx context.Context, snowflakeID int64) (*audit.OperationLog, error) {
	var log audit.OperationLog
	if err := r.db.WithContext(ctx).Where("snowflake_id = ?", snowflakeID).First(&log).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("failed to get operation log by snowflake ID",
			zap.Int64("snowflake_id", snowflakeID),
			zap.Error(err),
		)
		return nil, err
	}

	return &log, nil
}

// List retrieves a list of operation logs with pagination and filtering.
func (r *OperationLogRepository) List(ctx context.Context, offset, limit int, userID int64, operationType string, startTime, endTime string) ([]audit.OperationLog, int64, error) {
	var logs []audit.OperationLog
	var total int64

	query := r.db.WithContext(ctx).Model(&audit.OperationLog{})

	// Apply filters
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	if operationType != "" {
		query = query.Where("operation_type = ?", operationType)
	}

	if startTime != "" {
		start, err := time.Parse(time.RFC3339, startTime)
		if err == nil {
			query = query.Where("created_at >= ?", start)
		}
	}

	if endTime != "" {
		end, err := time.Parse(time.RFC3339, endTime)
		if err == nil {
			query = query.Where("created_at <= ?", end)
		}
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("failed to count operation logs",
			zap.Int64("user_id", userID),
			zap.String("operation_type", operationType),
			zap.Error(err),
		)
		return nil, 0, err
	}

	// Retrieve paginated records
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		r.logger.Error("failed to list operation logs",
			zap.Int("offset", offset),
			zap.Int("limit", limit),
			zap.Error(err),
		)
		return nil, 0, err
	}

	return logs, total, nil
}

// GetByUsername retrieves a user ID by username.
func (r *OperationLogRepository) GetByUsername(ctx context.Context, username string) (int64, error) {
	var userID int64
	if err := r.db.WithContext(ctx).Table("users").Select("snowflake_id").Where("username = ?", username).Scan(&userID).Error; err != nil {
		r.logger.Error("failed to get user ID by username",
			zap.String("username", username),
			zap.Error(err),
		)
		return 0, err
	}

	return userID, nil
}