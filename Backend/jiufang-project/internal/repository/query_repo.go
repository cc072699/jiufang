// Package repository implements the data access layer for query history and favorites.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"jiufang/internal/model/query"
)

// QueryRepository implements QueryRepositoryInterface using GORM.
type QueryRepository struct {
	db *gorm.DB
}

// NewQueryRepository creates a new QueryRepository instance.
func NewQueryRepository(db *gorm.DB) *QueryRepository {
	return &QueryRepository{db: db}
}

// CreateQuerySession creates a new query session.
func (r *QueryRepository) CreateQuerySession(ctx context.Context, session *query.QuerySession) error {
	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		return fmt.Errorf("failed to create query session: %w", err)
	}
	return nil
}

// GetQuerySessionByID retrieves a query session by its physical ID.
func (r *QueryRepository) GetQuerySessionByID(ctx context.Context, id uint) (*query.QuerySession, error) {
	var session query.QuerySession
	if err := r.db.WithContext(ctx).First(&session, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("query session not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get query session by id: %w", err)
	}
	return &session, nil
}

// GetQuerySessionBySnowflakeID retrieves a query session by its snowflake ID.
func (r *QueryRepository) GetQuerySessionBySnowflakeID(ctx context.Context, snowflakeID int64) (*query.QuerySession, error) {
	var session query.QuerySession
	if err := r.db.WithContext(ctx).Where("snowflake_id = ?", snowflakeID).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("query session not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get query session by snowflake id: %w", err)
	}
	return &session, nil
}

// GetQuerySessionsByUserID retrieves query sessions for a specific user with pagination.
func (r *QueryRepository) GetQuerySessionsByUserID(ctx context.Context, userID int64, offset, limit int) ([]query.QuerySession, int64, error) {
	var sessions []query.QuerySession
	var total int64

	if err := r.db.WithContext(ctx).Model(&query.QuerySession{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count query sessions: %w", err)
	}

	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&sessions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get query sessions: %w", err)
	}

	return sessions, total, nil
}

// CloseQuerySession closes a query session by setting its status to closed.
func (r *QueryRepository) CloseQuerySession(ctx context.Context, snowflakeID int64) error {
	result := r.db.WithContext(ctx).Model(&query.QuerySession{}).
		Where("snowflake_id = ?", snowflakeID).
		Updates(map[string]interface{}{
			"status":     query.SessionStatusClosed,
			"closed_at":  time.Now(),
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to close query session: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("query session not found")
	}

	return nil
}

// CreateQueryRecord creates a new query record.
func (r *QueryRepository) CreateQueryRecord(ctx context.Context, record *query.QueryRecord) error {
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return fmt.Errorf("failed to create query record: %w", err)
	}
	return nil
}

// GetQueryRecordByID retrieves a query record by its physical ID.
func (r *QueryRepository) GetQueryRecordByID(ctx context.Context, id uint) (*query.QueryRecord, error) {
	var record query.QueryRecord
	if err := r.db.WithContext(ctx).First(&record, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("query record not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get query record by id: %w", err)
	}
	return &record, nil
}

// GetQueryRecordBySnowflakeID retrieves a query record by its snowflake ID.
func (r *QueryRepository) GetQueryRecordBySnowflakeID(ctx context.Context, snowflakeID int64) (*query.QueryRecord, error) {
	var record query.QueryRecord
	if err := r.db.WithContext(ctx).Where("snowflake_id = ?", snowflakeID).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("query record not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get query record by snowflake id: %w", err)
	}
	return &record, nil
}

// GetQueryRecordsByUserID retrieves query records for a specific user with pagination and filters.
func (r *QueryRepository) GetQueryRecordsByUserID(ctx context.Context, userID int64, offset, limit int, startTime, endTime string, status string) ([]query.QueryRecord, int64, error) {
	var records []query.QueryRecord
	var total int64

	query := r.db.WithContext(ctx).Model(&query.QueryRecord{}).Where("user_id = ?", userID)

	// Apply time range filter
	if startTime != "" {
		startTimeParsed, err := time.Parse(time.RFC3339, startTime)
		if err == nil {
			query = query.Where("created_at >= ?", startTimeParsed)
		}
	}
	if endTime != "" {
		endTimeParsed, err := time.Parse(time.RFC3339, endTime)
		if err == nil {
			query = query.Where("created_at <= ?", endTimeParsed)
		}
	}

	// Apply status filter
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count query records: %w", err)
	}

	// Get records with pagination
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&records).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get query records: %w", err)
	}

	return records, total, nil
}

// GetQueryRecordsBySessionID retrieves all query records for a specific session.
func (r *QueryRepository) GetQueryRecordsBySessionID(ctx context.Context, sessionID int64) ([]query.QueryRecord, error) {
	var records []query.QueryRecord
	if err := r.db.WithContext(ctx).Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get query records by session id: %w", err)
	}
	return records, nil
}

// DeleteQueryRecord deletes a query record by snowflake ID (user can only delete their own records).
func (r *QueryRepository) DeleteQueryRecord(ctx context.Context, snowflakeID int64, userID int64) error {
	result := r.db.WithContext(ctx).
		Where("snowflake_id = ? AND user_id = ?", snowflakeID, userID).
		Delete(&query.QueryRecord{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete query record: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("query record not found or not owned by user")
	}

	return nil
}

// CreateFavorite creates a new favorite.
func (r *QueryRepository) CreateFavorite(ctx context.Context, favorite *query.Favorite) error {
	if err := r.db.WithContext(ctx).Create(favorite).Error; err != nil {
		return fmt.Errorf("failed to create favorite: %w", err)
	}
	return nil
}

// GetFavoriteByID retrieves a favorite by its physical ID.
func (r *QueryRepository) GetFavoriteByID(ctx context.Context, id uint) (*query.Favorite, error) {
	var favorite query.Favorite
	if err := r.db.WithContext(ctx).First(&favorite, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("favorite not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get favorite by id: %w", err)
	}
	return &favorite, nil
}

// GetFavoriteBySnowflakeID retrieves a favorite by its snowflake ID.
func (r *QueryRepository) GetFavoriteBySnowflakeID(ctx context.Context, snowflakeID int64) (*query.Favorite, error) {
	var favorite query.Favorite
	if err := r.db.WithContext(ctx).Where("snowflake_id = ?", snowflakeID).First(&favorite).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("favorite not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get favorite by snowflake id: %w", err)
	}
	return &favorite, nil
}

// GetFavoritesByUserID retrieves favorites for a specific user with pagination.
func (r *QueryRepository) GetFavoritesByUserID(ctx context.Context, userID int64, offset, limit int, name string) ([]query.Favorite, int64, error) {
	var favorites []query.Favorite
	var total int64

	query := r.db.WithContext(ctx).Model(&query.Favorite{}).Where("user_id = ?", userID)

	// Apply name filter if provided
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count favorites: %w", err)
	}

	// Get records with pagination
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&favorites).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get favorites: %w", err)
	}

	return favorites, total, nil
}

// DeleteFavorite deletes a favorite by snowflake ID (user can only delete their own favorites).
func (r *QueryRepository) DeleteFavorite(ctx context.Context, snowflakeID int64, userID int64) error {
	result := r.db.WithContext(ctx).
		Where("snowflake_id = ? AND user_id = ?", snowflakeID, userID).
		Delete(&query.Favorite{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete favorite: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("favorite not found or not owned by user")
	}

	return nil
}

// IsFavoriteNameExists checks if a favorite with the same name already exists for a user.
func (r *QueryRepository) IsFavoriteNameExists(ctx context.Context, userID int64, name string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&query.Favorite{}).
		Where("user_id = ? AND name = ?", userID, name).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check favorite name existence: %w", err)
	}
	return count > 0, nil
}
