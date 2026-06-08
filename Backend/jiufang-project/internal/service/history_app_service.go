// Package service implements the application layer for query history management.
package service

import (
	"context"
	"fmt"
	"strconv"

	"jiufang/internal/model/query"
	"jiufang/internal/pkg/id"
	"jiufang/internal/repository"
)

// HistoryAppService manages query history and favorites.
type HistoryAppService struct {
	queryRepo   repository.QueryRepositoryInterface
	idGenerator id.SnowflakeGeneratorInterface
}

// NewHistoryAppService creates a new HistoryAppService instance.
func NewHistoryAppService(queryRepo repository.QueryRepositoryInterface, idGenerator id.SnowflakeGeneratorInterface) *HistoryAppService {
	return &HistoryAppService{
		queryRepo:   queryRepo,
		idGenerator: idGenerator,
	}
}

// GetHistoryList retrieves query history list for a user with pagination and filters.
func (s *HistoryAppService) GetHistoryList(ctx context.Context, userID int64, page, pageSize int, startTime, endTime, status string) ([]query.QueryRecord, int64, error) {
	// Validate and set default pagination
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	records, total, err := s.queryRepo.GetQueryRecordsByUserID(ctx, userID, offset, pageSize, startTime, endTime, status)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get history list: %w", err)
	}

	return records, total, nil
}

// GetHistoryDetail retrieves a single query record detail.
func (s *HistoryAppService) GetHistoryDetail(ctx context.Context, userID int64, recordID string) (*query.QueryRecord, error) {
	// Parse snowflake ID
	snowflakeID, err := strconv.ParseInt(recordID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid record id format: %w", err)
	}

	// Get query record
	record, err := s.queryRepo.GetQueryRecordBySnowflakeID(ctx, snowflakeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get history detail: %w", err)
	}

	// Verify ownership
	if record.UserID != userID {
		return nil, fmt.Errorf("query record not owned by user")
	}

	return record, nil
}

// GetHistoryBySessionID retrieves all query records for a specific session.
func (s *HistoryAppService) GetHistoryBySessionID(ctx context.Context, userID int64, sessionID string) ([]query.QueryRecord, error) {
	// Parse session ID
	sessionIDInt, err := strconv.ParseInt(sessionID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid session id format: %w", err)
	}

	// Get all records for this session
	records, err := s.queryRepo.GetQueryRecordsBySessionID(ctx, sessionIDInt)
	if err != nil {
		return nil, fmt.Errorf("failed to get history by session id: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("session not found")
	}

	// Verify ownership — all records in a session belong to the same user
	if records[0].UserID != userID {
		return nil, fmt.Errorf("session not owned by user")
	}

	return records, nil
}

// DeleteHistory deletes a query record.
func (s *HistoryAppService) DeleteHistory(ctx context.Context, userID int64, recordID string) error {
	// Parse snowflake ID
	snowflakeID, err := strconv.ParseInt(recordID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid record id format: %w", err)
	}

	// Delete query record (repository will verify ownership)
	if err := s.queryRepo.DeleteQueryRecord(ctx, snowflakeID, userID); err != nil {
		return fmt.Errorf("failed to delete history: %w", err)
	}

	return nil
}

// CreateFavorite creates a new favorite.
func (s *HistoryAppService) CreateFavorite(ctx context.Context, userID int64, name, input, sql, description string) (*query.Favorite, error) {
	// Generate snowflake ID for favorite
	favoriteSnowflakeID := s.idGenerator.Generate()

	// Create favorite
	favorite := &query.Favorite{
		SnowflakeID: favoriteSnowflakeID,
		UserID:      userID,
		Name:        name,
		Input:       input,
		Sql:         sql,
		Description: description,
	}

	if err := s.queryRepo.CreateFavorite(ctx, favorite); err != nil {
		return nil, fmt.Errorf("failed to create favorite: %w", err)
	}

	return favorite, nil
}

// GetFavoriteList retrieves favorite list for a user with pagination.
func (s *HistoryAppService) GetFavoriteList(ctx context.Context, userID int64, page, pageSize int, name string) ([]query.Favorite, int64, error) {
	// Validate and set default pagination
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	favorites, total, err := s.queryRepo.GetFavoritesByUserID(ctx, userID, offset, pageSize, name)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get favorite list: %w", err)
	}

	return favorites, total, nil
}

// DeleteFavorite deletes a favorite.
func (s *HistoryAppService) DeleteFavorite(ctx context.Context, userID int64, favoriteID string) error {
	// Parse snowflake ID
	snowflakeID, err := strconv.ParseInt(favoriteID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid favorite id format: %w", err)
	}

	// Delete favorite (repository will verify ownership)
	if err := s.queryRepo.DeleteFavorite(ctx, snowflakeID, userID); err != nil {
		return fmt.Errorf("failed to delete favorite: %w", err)
	}

	return nil
}

// CreateQuerySession creates a new query session for a user.
func (s *HistoryAppService) CreateQuerySession(ctx context.Context, userID int64, dialogID int64) (*query.QuerySession, error) {
	// Generate snowflake ID
	sessionSnowflakeID := s.idGenerator.Generate()

	// Create session
	session := &query.QuerySession{
		SnowflakeID: sessionSnowflakeID,
		UserID:      userID,
		DialogID:    dialogID,
		Status:      query.SessionStatusActive,
	}

	if err := s.queryRepo.CreateQuerySession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create query session: %w", err)
	}

	return session, nil
}

// CloseQuerySession closes a query session.
func (s *HistoryAppService) CloseQuerySession(ctx context.Context, userID int64, sessionID string) error {
	// Parse snowflake ID
	snowflakeID, err := strconv.ParseInt(sessionID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid session id format: %w", err)
	}

	// Verify session belongs to user
	session, err := s.queryRepo.GetQuerySessionBySnowflakeID(ctx, snowflakeID)
	if err != nil {
		return fmt.Errorf("failed to get query session: %w", err)
	}

	if session.UserID != userID {
		return fmt.Errorf("query session not owned by user")
	}

	// Close session
	if err := s.queryRepo.CloseQuerySession(ctx, snowflakeID); err != nil {
		return fmt.Errorf("failed to close query session: %w", err)
	}

	return nil
}
