// Package service implements the application layer for dialog favorite management.
package service

import (
	"context"
	"fmt"
	"strconv"

	"jiufang/internal/model/dialog"
	"jiufang/internal/pkg/id"
	"jiufang/internal/repository"
)

// DialogFavoriteAppService manages dialog favorites.
type DialogFavoriteAppService struct {
	dialogFavoriteRepo repository.DialogFavoriteRepositoryInterface
	dialogRepo         repository.DialogRepositoryInterface
	idGenerator        id.SnowflakeGeneratorInterface
}

// NewDialogFavoriteAppService creates a new DialogFavoriteAppService instance.
func NewDialogFavoriteAppService(
	dialogFavoriteRepo repository.DialogFavoriteRepositoryInterface,
	dialogRepo repository.DialogRepositoryInterface,
	idGenerator id.SnowflakeGeneratorInterface,
) *DialogFavoriteAppService {
	return &DialogFavoriteAppService{
		dialogFavoriteRepo: dialogFavoriteRepo,
		dialogRepo:         dialogRepo,
		idGenerator:        idGenerator,
	}
}

// CreateDialogFavorite creates a new favorite for a dialog session.
func (s *DialogFavoriteAppService) CreateDialogFavorite(ctx context.Context, userID int64, dialogSessionID string, title string) (*dialog.DialogFavorite, error) {
	// Parse dialog session snowflake ID
	sessionSnowflakeID, err := strconv.ParseInt(dialogSessionID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid dialog session id format: %w", err)
	}

	// Verify dialog session exists and belongs to user
	session, err := s.dialogRepo.GetBySnowflakeID(ctx, strconv.FormatInt(sessionSnowflakeID, 10))
	if err != nil {
		return nil, fmt.Errorf("failed to get dialog session: %w", err)
	}

	if session.UserID != uint(userID) {
		return nil, fmt.Errorf("dialog session not owned by user")
	}

	// Check if already favorited
	exists, err := s.dialogFavoriteRepo.IsDialogFavoriteExists(ctx, userID, sessionSnowflakeID)
	if err != nil {
		return nil, fmt.Errorf("failed to check dialog favorite existence: %w", err)
	}

	if exists {
		return nil, fmt.Errorf("dialog session already favorited")
	}

	// Generate snowflake ID for favorite
	favoriteSnowflakeID := s.idGenerator.Generate()

	// Create favorite
	favorite := &dialog.DialogFavorite{
		SnowflakeID:     favoriteSnowflakeID,
		UserID:          userID,
		DialogSessionID: sessionSnowflakeID,
		Title:           title,
	}

	if err := s.dialogFavoriteRepo.CreateDialogFavorite(ctx, favorite); err != nil {
		return nil, fmt.Errorf("failed to create dialog favorite: %w", err)
	}

	return favorite, nil
}

// GetDialogFavoriteList retrieves favorite list for a user with pagination.
func (s *DialogFavoriteAppService) GetDialogFavoriteList(ctx context.Context, userID int64, page, pageSize int) ([]dialog.DialogFavorite, int64, error) {
	// Validate and set default pagination
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	favorites, total, err := s.dialogFavoriteRepo.GetDialogFavoritesByUserID(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get dialog favorite list: %w", err)
	}

	return favorites, total, nil
}

// DeleteDialogFavorite deletes a dialog favorite.
func (s *DialogFavoriteAppService) DeleteDialogFavorite(ctx context.Context, userID int64, favoriteID string) error {
	// Parse snowflake ID
	snowflakeID, err := strconv.ParseInt(favoriteID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid favorite id format: %w", err)
	}

	// Delete dialog favorite (repository will verify ownership)
	if err := s.dialogFavoriteRepo.DeleteDialogFavorite(ctx, snowflakeID, userID); err != nil {
		return fmt.Errorf("failed to delete dialog favorite: %w", err)
	}

	return nil
}