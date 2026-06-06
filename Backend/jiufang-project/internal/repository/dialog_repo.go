// Package repository implements data access layer for dialog sessions.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"jiufang/internal/model/dialog"
)

// DialogRepository implements DialogRepositoryInterface using GORM.
type DialogRepository struct {
	db *gorm.DB
}

// NewDialogRepository creates a new dialog repository.
func NewDialogRepository(db *gorm.DB) *DialogRepository {
	return &DialogRepository{db: db}
}

// Create creates a new dialog session.
func (r *DialogRepository) Create(ctx context.Context, session *dialog.DialogSession) error {
	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		return fmt.Errorf("failed to create dialog session: %w", err)
	}
	return nil
}

// GetByID retrieves a dialog session by its ID.
func (r *DialogRepository) GetByID(ctx context.Context, id uint) (*dialog.DialogSession, error) {
	var session dialog.DialogSession
	err := r.db.WithContext(ctx).First(&session, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get dialog session by id: %w", err)
	}
	return &session, nil
}

// GetBySnowflakeID retrieves a dialog session by its snowflake ID.
func (r *DialogRepository) GetBySnowflakeID(ctx context.Context, snowflakeID string) (*dialog.DialogSession, error) {
	var session dialog.DialogSession
	err := r.db.WithContext(ctx).Where("snowflake_id = ?", snowflakeID).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get dialog session by snowflake_id: %w", err)
	}
	return &session, nil
}

// GetByUserID retrieves dialog sessions by user ID with pagination.
func (r *DialogRepository) GetByUserID(ctx context.Context, userID uint, offset, limit int) ([]dialog.DialogSession, int64, error) {
	var sessions []dialog.DialogSession
	var total int64

	query := r.db.WithContext(ctx).Model(&dialog.DialogSession{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count dialog sessions: %w", err)
	}

	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&sessions).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list dialog sessions: %w", err)
	}

	return sessions, total, nil
}

// GetActiveSessionsByUserID retrieves active dialog sessions by user ID.
func (r *DialogRepository) GetActiveSessionsByUserID(ctx context.Context, userID uint) ([]dialog.DialogSession, error) {
	var sessions []dialog.DialogSession
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, dialog.StatusActive).
		Order("created_at DESC").
		Find(&sessions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get active dialog sessions: %w", err)
	}
	return sessions, nil
}

// Update updates a dialog session.
func (r *DialogRepository) Update(ctx context.Context, session *dialog.DialogSession) error {
	if err := r.db.WithContext(ctx).Save(session).Error; err != nil {
		return fmt.Errorf("failed to update dialog session: %w", err)
	}
	return nil
}

// UpdateStatus updates the status of a dialog session.
func (r *DialogRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	err := r.db.WithContext(ctx).
		Model(&dialog.DialogSession{}).
		Where("id = ?", id).
		Update("status", status).Error
	if err != nil {
		return fmt.Errorf("failed to update dialog session status: %w", err)
	}
	return nil
}

// Delete deletes a dialog session.
func (r *DialogRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&dialog.DialogSession{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete dialog session: %w", err)
	}
	return nil
}

// CloseSession closes a dialog session by snowflake ID.
func (r *DialogRepository) CloseSession(ctx context.Context, snowflakeID string) error {
	now := time.Now()
	err := r.db.WithContext(ctx).
		Model(&dialog.DialogSession{}).
		Where("snowflake_id = ?", snowflakeID).
		Updates(map[string]interface{}{
			"status":    dialog.StatusClosed,
			"closed_at": now,
			"updated_at": now,
		}).Error
	if err != nil {
		return fmt.Errorf("failed to close dialog session: %w", err)
	}
	return nil
}