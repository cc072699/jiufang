// Package repository implements the data access layer for export records.
package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"jiufang/internal/model/export"
)

// ExportRecordRepository implements ExportRecordRepositoryInterface using GORM.
type ExportRecordRepository struct {
	db *gorm.DB
}

// NewExportRecordRepository creates a new ExportRecordRepository instance.
func NewExportRecordRepository(db *gorm.DB) *ExportRecordRepository {
	return &ExportRecordRepository{db: db}
}

// CreateExportRecord creates a new export record.
func (r *ExportRecordRepository) CreateExportRecord(ctx context.Context, record *export.ExportRecord) error {
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return fmt.Errorf("failed to create export record: %w", err)
	}
	return nil
}

// GetExportRecordByID retrieves an export record by its physical ID.
func (r *ExportRecordRepository) GetExportRecordByID(ctx context.Context, id uint) (*export.ExportRecord, error) {
	var record export.ExportRecord
	if err := r.db.WithContext(ctx).First(&record, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("export record not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get export record by id: %w", err)
	}
	return &record, nil
}

// GetExportRecordBySnowflakeID retrieves an export record by its snowflake ID.
func (r *ExportRecordRepository) GetExportRecordBySnowflakeID(ctx context.Context, snowflakeID int64) (*export.ExportRecord, error) {
	var record export.ExportRecord
	if err := r.db.WithContext(ctx).Where("snowflake_id = ?", snowflakeID).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("export record not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get export record by snowflake id: %w", err)
	}
	return &record, nil
}

// GetExportRecordsByUserID retrieves export records for a specific user with pagination.
func (r *ExportRecordRepository) GetExportRecordsByUserID(ctx context.Context, userID int64, offset, limit int) ([]export.ExportRecord, int64, error) {
	var records []export.ExportRecord
	var total int64

	if err := r.db.WithContext(ctx).Model(&export.ExportRecord{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count export records: %w", err)
	}

	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&records).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get export records: %w", err)
	}

	return records, total, nil
}