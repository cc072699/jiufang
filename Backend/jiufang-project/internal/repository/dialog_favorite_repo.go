// Package repository implements the data access layer for dialog favorites.
package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"jiufang/internal/model/dialog"
)

// DialogFavoriteRepository implements DialogFavoriteRepositoryInterface using GORM.
type DialogFavoriteRepository struct {
	db *gorm.DB
}

// NewDialogFavoriteRepository creates a new DialogFavoriteRepository instance.
func NewDialogFavoriteRepository(db *gorm.DB) *DialogFavoriteRepository {
	return &DialogFavoriteRepository{db: db}
}

// CreateDialogFavorite creates a new dialog favorite.
func (r *DialogFavoriteRepository) CreateDialogFavorite(ctx context.Context, favorite *dialog.DialogFavorite) error {
	if err := r.db.WithContext(ctx).Create(favorite).Error; err != nil {
		return fmt.Errorf("failed to create dialog favorite: %w", err)
	}
	return nil
}

// GetDialogFavoriteByID retrieves a dialog favorite by its physical ID.
func (r *DialogFavoriteRepository) GetDialogFavoriteByID(ctx context.Context, id uint) (*dialog.DialogFavorite, error) {
	var favorite dialog.DialogFavorite
	if err := r.db.WithContext(ctx).First(&favorite, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("dialog favorite not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get dialog favorite by id: %w", err)
	}
	return &favorite, nil
}

// GetDialogFavoriteBySnowflakeID retrieves a dialog favorite by its snowflake ID.
func (r *DialogFavoriteRepository) GetDialogFavoriteBySnowflakeID(ctx context.Context, snowflakeID int64) (*dialog.DialogFavorite, error) {
	var favorite dialog.DialogFavorite
	if err := r.db.WithContext(ctx).Where("snowflake_id = ?", snowflakeID).First(&favorite).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("dialog favorite not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get dialog favorite by snowflake id: %w", err)
	}
	return &favorite, nil
}

// GetDialogFavoritesByUserID retrieves dialog favorites for a specific user with pagination.
func (r *DialogFavoriteRepository) GetDialogFavoritesByUserID(ctx context.Context, userID int64, offset, limit int) ([]dialog.DialogFavorite, int64, error) {
	var favorites []dialog.DialogFavorite
	var total int64

	if err := r.db.WithContext(ctx).Model(&dialog.DialogFavorite{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count dialog favorites: %w", err)
	}

	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&favorites).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get dialog favorites: %w", err)
	}

	return favorites, total, nil
}

// DeleteDialogFavorite deletes a dialog favorite by snowflake ID (user can only delete their own favorites).
func (r *DialogFavoriteRepository) DeleteDialogFavorite(ctx context.Context, snowflakeID int64, userID int64) error {
	result := r.db.WithContext(ctx).
		Where("snowflake_id = ? AND user_id = ?", snowflakeID, userID).
		Delete(&dialog.DialogFavorite{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete dialog favorite: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("dialog favorite not found or not owned by user")
	}

	return nil
}

// IsDialogFavoriteExists checks if a dialog favorite already exists for a user and dialog session.
func (r *DialogFavoriteRepository) IsDialogFavoriteExists(ctx context.Context, userID int64, dialogSessionID int64) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&dialog.DialogFavorite{}).
		Where("user_id = ? AND dialog_session_id = ?", userID, dialogSessionID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check dialog favorite existence: %w", err)
	}
	return count > 0, nil
}

// GetDialogFavoriteByUserAndSession retrieves a dialog favorite by user ID and dialog session ID.
func (r *DialogFavoriteRepository) GetDialogFavoriteByUserAndSession(ctx context.Context, userID int64, dialogSessionID int64) (*dialog.DialogFavorite, error) {
	var favorite dialog.DialogFavorite
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND dialog_session_id = ?", userID, dialogSessionID).
		First(&favorite).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Not found is not an error in this case
		}
		return nil, fmt.Errorf("failed to get dialog favorite by user and session: %w", err)
	}
	return &favorite, nil
}