// Package repository implements the data access layer for feedbacks.
package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"jiufang/internal/model/feedback"
)

// FeedbackRepository implements FeedbackRepositoryInterface using GORM.
type FeedbackRepository struct {
	db *gorm.DB
}

// NewFeedbackRepository creates a new FeedbackRepository instance.
func NewFeedbackRepository(db *gorm.DB) *FeedbackRepository {
	return &FeedbackRepository{db: db}
}

// Create creates a new feedback.
func (r *FeedbackRepository) Create(ctx context.Context, feedback *feedback.Feedback) error {
	if err := r.db.WithContext(ctx).Create(feedback).Error; err != nil {
		return fmt.Errorf("failed to create feedback: %w", err)
	}
	return nil
}

// GetByID retrieves a feedback by its physical ID.
func (r *FeedbackRepository) GetByID(ctx context.Context, id uint) (*feedback.Feedback, error) {
	var fb feedback.Feedback
	if err := r.db.WithContext(ctx).First(&fb, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("feedback not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get feedback by id: %w", err)
	}
	return &fb, nil
}

// GetBySnowflakeID retrieves a feedback by its snowflake ID.
func (r *FeedbackRepository) GetBySnowflakeID(ctx context.Context, snowflakeID int64) (*feedback.Feedback, error) {
	var fb feedback.Feedback
	if err := r.db.WithContext(ctx).Where("snowflake_id = ?", snowflakeID).First(&fb).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("feedback not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get feedback by snowflake id: %w", err)
	}
	return &fb, nil
}

// GetByQueryRecordID retrieves a feedback by its query record ID.
func (r *FeedbackRepository) GetByQueryRecordID(ctx context.Context, queryRecordID int64) (*feedback.Feedback, error) {
	var fb feedback.Feedback
	if err := r.db.WithContext(ctx).Where("query_record_id = ?", queryRecordID).First(&fb).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("feedback not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get feedback by query record id: %w", err)
	}
	return &fb, nil
}

// GetByUserID retrieves feedbacks by user ID with pagination.
func (r *FeedbackRepository) GetByUserID(ctx context.Context, userID int64, offset, limit int) ([]feedback.Feedback, int64, error) {
	var feedbacks []feedback.Feedback
	var total int64

	query := r.db.WithContext(ctx).Model(&feedback.Feedback{}).Where("user_id = ?", userID)

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count feedbacks: %w", err)
	}

	// Get records with pagination
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&feedbacks).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get feedbacks by user id: %w", err)
	}

	return feedbacks, total, nil
}

// List retrieves feedbacks with pagination and filters.
func (r *FeedbackRepository) List(ctx context.Context, offset, limit int, userID int64, rating string) ([]feedback.Feedback, int64, error) {
	var feedbacks []feedback.Feedback
	var total int64

	query := r.db.WithContext(ctx).Model(&feedback.Feedback{})

	// Apply user_id filter
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	// Apply rating filter
	if rating != "" {
		query = query.Where("rating = ?", rating)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count feedbacks: %w", err)
	}

	// Get records with pagination
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&feedbacks).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list feedbacks: %w", err)
	}

	return feedbacks, total, nil
}

// IsFeedbackExists checks if a feedback exists for a given query record ID.
func (r *FeedbackRepository) IsFeedbackExists(ctx context.Context, queryRecordID int64) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&feedback.Feedback{}).Where("query_record_id = ?", queryRecordID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check feedback existence: %w", err)
	}
	return count > 0, nil
}