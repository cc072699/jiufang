// Package repository implements the alert repository.
package repository

import (
	"context"
	"errors"
	"time"

	"jiufang/internal/model/report"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AlertRepository implements AlertRepositoryInterface using GORM.
type AlertRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewAlertRepository creates a new AlertRepository instance.
func NewAlertRepository(db *gorm.DB, logger *zap.Logger) *AlertRepository {
	return &AlertRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new alert rule.
func (r *AlertRepository) Create(ctx context.Context, alert *report.Alert) error {
	if err := r.db.WithContext(ctx).Create(alert).Error; err != nil {
		r.logger.Error("failed to create alert",
			zap.Int64("snowflake_id", alert.SnowflakeID),
			zap.String("name", alert.Name),
			zap.Error(err),
		)
		return err
	}

	r.logger.Info("alert created successfully",
		zap.Int64("snowflake_id", alert.SnowflakeID),
		zap.String("name", alert.Name),
	)

	return nil
}

// GetByID retrieves an alert rule by its primary key ID.
func (r *AlertRepository) GetByID(ctx context.Context, id uint) (*report.Alert, error) {
	var alert report.Alert
	if err := r.db.WithContext(ctx).First(&alert, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("failed to get alert by ID",
			zap.Uint("id", id),
			zap.Error(err),
		)
		return nil, err
	}

	return &alert, nil
}

// GetBySnowflakeID retrieves an alert rule by its snowflake ID.
func (r *AlertRepository) GetBySnowflakeID(ctx context.Context, snowflakeID int64) (*report.Alert, error) {
	var alert report.Alert
	if err := r.db.WithContext(ctx).Where("snowflake_id = ?", snowflakeID).First(&alert).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("failed to get alert by snowflake ID",
			zap.Int64("snowflake_id", snowflakeID),
			zap.Error(err),
		)
		return nil, err
	}

	return &alert, nil
}

// List retrieves a list of alert rules with pagination and filtering.
func (r *AlertRepository) List(ctx context.Context, offset, limit int, name string, status string) ([]report.Alert, int64, error) {
	var alerts []report.Alert
	var total int64

	query := r.db.WithContext(ctx).Model(&report.Alert{})

	// Apply filters
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("failed to count alerts",
			zap.String("name", name),
			zap.String("status", status),
			zap.Error(err),
		)
		return nil, 0, err
	}

	// Retrieve paginated records
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&alerts).Error; err != nil {
		r.logger.Error("failed to list alerts",
			zap.String("name", name),
			zap.String("status", status),
			zap.Int("offset", offset),
			zap.Int("limit", limit),
			zap.Error(err),
		)
		return nil, 0, err
	}

	return alerts, total, nil
}

// Update updates an alert rule by its snowflake ID.
func (r *AlertRepository) Update(ctx context.Context, snowflakeID int64, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&report.Alert{}).Where("snowflake_id = ?", snowflakeID).Updates(updates)
	if result.Error != nil {
		r.logger.Error("failed to update alert",
			zap.Int64("snowflake_id", snowflakeID),
			zap.Any("updates", updates),
			zap.Error(result.Error),
		)
		return result.Error
	}

	if result.RowsAffected == 0 {
		r.logger.Warn("no alert found to update",
			zap.Int64("snowflake_id", snowflakeID),
		)
		return errors.New("alert not found")
	}

	r.logger.Info("alert updated successfully",
		zap.Int64("snowflake_id", snowflakeID),
		zap.Any("updates", updates),
	)

	return nil
}

// Delete deletes an alert rule by its snowflake ID.
func (r *AlertRepository) Delete(ctx context.Context, snowflakeID int64) error {
	result := r.db.WithContext(ctx).Where("snowflake_id = ?", snowflakeID).Delete(&report.Alert{})
	if result.Error != nil {
		r.logger.Error("failed to delete alert",
			zap.Int64("snowflake_id", snowflakeID),
			zap.Error(result.Error),
		)
		return result.Error
	}

	if result.RowsAffected == 0 {
		r.logger.Warn("no alert found to delete",
			zap.Int64("snowflake_id", snowflakeID),
		)
		return errors.New("alert not found")
	}

	r.logger.Info("alert deleted successfully",
		zap.Int64("snowflake_id", snowflakeID),
	)

	return nil
}

// GetActiveAlerts retrieves all active alert rules.
func (r *AlertRepository) GetActiveAlerts(ctx context.Context) ([]report.Alert, error) {
	var alerts []report.Alert
	if err := r.db.WithContext(ctx).Where("status = ?", report.AlertStatusActive).Find(&alerts).Error; err != nil {
		r.logger.Error("failed to get active alerts",
			zap.Error(err),
		)
		return nil, err
	}

	return alerts, nil
}

// UpdateLastTriggeredAt updates the last triggered timestamp for an alert rule.
func (r *AlertRepository) UpdateLastTriggeredAt(ctx context.Context, snowflakeID int64, triggeredAt time.Time) error {
	result := r.db.WithContext(ctx).Model(&report.Alert{}).Where("snowflake_id = ?", snowflakeID).Update("last_triggered_at", triggeredAt)
	if result.Error != nil {
		r.logger.Error("failed to update last triggered at",
			zap.Int64("snowflake_id", snowflakeID),
			zap.Time("triggered_at", triggeredAt),
			zap.Error(result.Error),
		)
		return result.Error
	}

	if result.RowsAffected == 0 {
		r.logger.Warn("no alert found to update last triggered at",
			zap.Int64("snowflake_id", snowflakeID),
		)
		return errors.New("alert not found")
	}

	r.logger.Info("alert last triggered at updated successfully",
		zap.Int64("snowflake_id", snowflakeID),
		zap.Time("triggered_at", triggeredAt),
	)

	return nil
}