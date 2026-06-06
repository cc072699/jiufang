// Package repository implements the data access layer for scheduled reports and push records.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"jiufang/internal/model/report"
)

// ReportRepository implements ReportRepositoryInterface using GORM.
type ReportRepository struct {
	db *gorm.DB
}

// NewReportRepository creates a new ReportRepository instance.
func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

// Create creates a new scheduled report.
func (r *ReportRepository) Create(ctx context.Context, scheduledReport *report.ScheduledReport) error {
	if err := r.db.WithContext(ctx).Create(scheduledReport).Error; err != nil {
		return fmt.Errorf("failed to create scheduled report: %w", err)
	}
	return nil
}

// GetByID retrieves a scheduled report by its physical ID.
func (r *ReportRepository) GetByID(ctx context.Context, id uint) (*report.ScheduledReport, error) {
	var scheduledReport report.ScheduledReport
	if err := r.db.WithContext(ctx).First(&scheduledReport, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("scheduled report not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get scheduled report by id: %w", err)
	}
	return &scheduledReport, nil
}

// GetBySnowflakeID retrieves a scheduled report by its snowflake ID.
func (r *ReportRepository) GetBySnowflakeID(ctx context.Context, snowflakeID int64) (*report.ScheduledReport, error) {
	var scheduledReport report.ScheduledReport
	if err := r.db.WithContext(ctx).Where("snowflake_id = ?", snowflakeID).First(&scheduledReport).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("scheduled report not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get scheduled report by snowflake id: %w", err)
	}
	return &scheduledReport, nil
}

// List retrieves scheduled reports with pagination and filters.
func (r *ReportRepository) List(ctx context.Context, offset, limit int, name, status string) ([]report.ScheduledReport, int64, error) {
	var scheduledReports []report.ScheduledReport
	var total int64

	query := r.db.WithContext(ctx).Model(&report.ScheduledReport{})

	// Apply name filter
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	// Apply status filter
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count scheduled reports: %w", err)
	}

	// Get records with pagination
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&scheduledReports).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list scheduled reports: %w", err)
	}

	return scheduledReports, total, nil
}

// Update updates a scheduled report by snowflake ID.
func (r *ReportRepository) Update(ctx context.Context, snowflakeID int64, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&report.ScheduledReport{}).
		Where("snowflake_id = ?", snowflakeID).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update scheduled report: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("scheduled report not found")
	}

	return nil
}

// Delete deletes a scheduled report by snowflake ID.
func (r *ReportRepository) Delete(ctx context.Context, snowflakeID int64) error {
	result := r.db.WithContext(ctx).Where("snowflake_id = ?", snowflakeID).Delete(&report.ScheduledReport{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete scheduled report: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("scheduled report not found")
	}

	return nil
}

// GetActiveReports retrieves all active scheduled reports.
func (r *ReportRepository) GetActiveReports(ctx context.Context) ([]report.ScheduledReport, error) {
	var scheduledReports []report.ScheduledReport
	if err := r.db.WithContext(ctx).
		Where("status = ?", report.ReportStatusActive).
		Order("created_at ASC").
		Find(&scheduledReports).Error; err != nil {
		return nil, fmt.Errorf("failed to get active scheduled reports: %w", err)
	}
	return scheduledReports, nil
}

// CreatePushRecord creates a new push record.
func (r *ReportRepository) CreatePushRecord(ctx context.Context, pushRecord *report.PushRecord) error {
	if err := r.db.WithContext(ctx).Create(pushRecord).Error; err != nil {
		return fmt.Errorf("failed to create push record: %w", err)
	}
	return nil
}

// GetPushRecordByID retrieves a push record by its physical ID.
func (r *ReportRepository) GetPushRecordByID(ctx context.Context, id uint) (*report.PushRecord, error) {
	var pushRecord report.PushRecord
	if err := r.db.WithContext(ctx).First(&pushRecord, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("push record not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get push record by id: %w", err)
	}
	return &pushRecord, nil
}

// GetPushRecordBySnowflakeID retrieves a push record by its snowflake ID.
func (r *ReportRepository) GetPushRecordBySnowflakeID(ctx context.Context, snowflakeID int64) (*report.PushRecord, error) {
	var pushRecord report.PushRecord
	if err := r.db.WithContext(ctx).Where("snowflake_id = ?", snowflakeID).First(&pushRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("push record not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get push record by snowflake id: %w", err)
	}
	return &pushRecord, nil
}

// GetPushRecordsByReportID retrieves push records for a specific report with pagination.
func (r *ReportRepository) GetPushRecordsByReportID(ctx context.Context, reportID int64, offset, limit int, pushStatus string) ([]report.PushRecord, int64, error) {
	var pushRecords []report.PushRecord
	var total int64

	query := r.db.WithContext(ctx).Model(&report.PushRecord{}).Where("report_id = ?", reportID)

	// Apply push status filter
	if pushStatus != "" {
		query = query.Where("push_status = ?", pushStatus)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count push records: %w", err)
	}

	// Get records with pagination
	if err := query.Order("push_time DESC").Offset(offset).Limit(limit).Find(&pushRecords).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get push records: %w", err)
	}

	return pushRecords, total, nil
}

// ListPushRecords retrieves all push records with pagination and filters.
func (r *ReportRepository) ListPushRecords(ctx context.Context, offset, limit int, pushType string, pushStatus string, startTime, endTime string) ([]report.PushRecord, int64, error) {
	var pushRecords []report.PushRecord
	var total int64

	query := r.db.WithContext(ctx).Model(&report.PushRecord{})

	// Apply push type filter
	if pushType != "" {
		query = query.Where("push_type = ?", pushType)
	}

	// Apply push status filter
	if pushStatus != "" {
		query = query.Where("push_status = ?", pushStatus)
	}

	// Apply time range filter
	if startTime != "" {
		startTimeParsed, err := time.Parse(time.RFC3339, startTime)
		if err == nil {
			query = query.Where("push_time >= ?", startTimeParsed)
		}
	}
	if endTime != "" {
		endTimeParsed, err := time.Parse(time.RFC3339, endTime)
		if err == nil {
			query = query.Where("push_time <= ?", endTimeParsed)
		}
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count push records: %w", err)
	}

	// Get records with pagination
	if err := query.Order("push_time DESC").Offset(offset).Limit(limit).Find(&pushRecords).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list push records: %w", err)
	}

	return pushRecords, total, nil
}

// UpdatePushRecord updates a push record by snowflake ID.
func (r *ReportRepository) UpdatePushRecord(ctx context.Context, snowflakeID int64, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&report.PushRecord{}).
		Where("snowflake_id = ?", snowflakeID).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update push record: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("push record not found")
	}

	return nil
}
