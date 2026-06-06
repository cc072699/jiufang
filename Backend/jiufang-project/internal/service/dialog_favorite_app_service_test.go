// Package service_test implements unit tests for dialog favorite application service.
// Author: AI Agent
// Date: 2026-06-03
// Description: Tests all methods of DialogFavoriteAppService using gomock

package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"jiufang/internal/mocks"
	"jiufang/internal/model/dialog"
	"jiufang/internal/service"
)

// TestDialogFavoriteAppService_CreateDialogFavorite_Success tests successful favorite creation.
func TestDialogFavoriteAppService_CreateDialogFavorite_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFavoriteRepo := mocks.NewMockDialogFavoriteRepository(ctrl)
	mockDialogRepo := mocks.NewMockDialogRepository(ctrl)
	mockIDGenerator := mocks.NewMockSnowflakeGenerator(ctrl)

	svc := service.NewDialogFavoriteAppService(mockFavoriteRepo, mockDialogRepo, mockIDGenerator)
	ctx := context.Background()

	userID := int64(1)
	dialogSessionID := "100"
	title := "Test Favorite"

	// Mock dialog session retrieval
	session := &dialog.DialogSession{
		ID:          1,
		SnowflakeID: "100",
		UserID:      uint(userID),
		Status:      string(dialog.StatusActive),
		CreatedAt:   time.Now(),
	}
	mockDialogRepo.EXPECT().GetBySnowflakeID(ctx, "100").Return(session, nil)

	// Mock favorite existence check
	mockFavoriteRepo.EXPECT().IsDialogFavoriteExists(ctx, userID, int64(100)).Return(false, nil)

	// Mock ID generation
	favoriteSnowflakeID := int64(123456789)
	mockIDGenerator.EXPECT().Generate().Return(favoriteSnowflakeID)

	// Mock favorite creation
	mockFavoriteRepo.EXPECT().CreateDialogFavorite(ctx, gomock.Any()).Return(nil)

	// Act
	favorite, err := svc.CreateDialogFavorite(ctx, userID, dialogSessionID, title)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, favorite)
	assert.Equal(t, favoriteSnowflakeID, favorite.SnowflakeID)
	assert.Equal(t, userID, favorite.UserID)
	assert.Equal(t, int64(100), favorite.DialogSessionID)
	assert.Equal(t, title, favorite.Title)
}

// TestDialogFavoriteAppService_CreateDialogFavorite_InvalidSessionIDFormat tests invalid session ID format.
func TestDialogFavoriteAppService_CreateDialogFavorite_InvalidSessionIDFormat(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFavoriteRepo := mocks.NewMockDialogFavoriteRepository(ctrl)
	mockDialogRepo := mocks.NewMockDialogRepository(ctrl)
	mockIDGenerator := mocks.NewMockSnowflakeGenerator(ctrl)

	svc := service.NewDialogFavoriteAppService(mockFavoriteRepo, mockDialogRepo, mockIDGenerator)
	ctx := context.Background()

	userID := int64(1)
	dialogSessionID := "abc" // Invalid format
	title := "Test Favorite"

	// Act
	favorite, err := svc.CreateDialogFavorite(ctx, userID, dialogSessionID, title)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, favorite)
	assert.Contains(t, err.Error(), "invalid dialog session id format")
}

// TestDialogFavoriteAppService_CreateDialogFavorite_SessionNotFound tests dialog session not found.
func TestDialogFavoriteAppService_CreateDialogFavorite_SessionNotFound(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFavoriteRepo := mocks.NewMockDialogFavoriteRepository(ctrl)
	mockDialogRepo := mocks.NewMockDialogRepository(ctrl)
	mockIDGenerator := mocks.NewMockSnowflakeGenerator(ctrl)

	svc := service.NewDialogFavoriteAppService(mockFavoriteRepo, mockDialogRepo, mockIDGenerator)
	ctx := context.Background()

	userID := int64(1)
	dialogSessionID := "999"
	title := "Test Favorite"

	// Mock dialog session retrieval (not found)
	mockDialogRepo.EXPECT().GetBySnowflakeID(ctx, "999").Return(nil, errors.New("session not found"))

	// Act
	favorite, err := svc.CreateDialogFavorite(ctx, userID, dialogSessionID, title)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, favorite)
	assert.Contains(t, err.Error(), "failed to get dialog session")
}

// TestDialogFavoriteAppService_CreateDialogFavorite_SessionNotOwnedByUser tests session not owned by user.
func TestDialogFavoriteAppService_CreateDialogFavorite_SessionNotOwnedByUser(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFavoriteRepo := mocks.NewMockDialogFavoriteRepository(ctrl)
	mockDialogRepo := mocks.NewMockDialogRepository(ctrl)
	mockIDGenerator := mocks.NewMockSnowflakeGenerator(ctrl)

	svc := service.NewDialogFavoriteAppService(mockFavoriteRepo, mockDialogRepo, mockIDGenerator)
	ctx := context.Background()

	userID := int64(2) // Different user
	dialogSessionID := "100"
	title := "Test Favorite"

	// Mock dialog session retrieval (owned by user 1)
	session := &dialog.DialogSession{
		ID:          1,
		SnowflakeID: "100",
		UserID:      uint(1), // Owned by user 1, not user 2
		Status:      string(dialog.StatusActive),
		CreatedAt:   time.Now(),
	}
	mockDialogRepo.EXPECT().GetBySnowflakeID(ctx, "100").Return(session, nil)

	// Act
	favorite, err := svc.CreateDialogFavorite(ctx, userID, dialogSessionID, title)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, favorite)
	assert.Contains(t, err.Error(), "dialog session not owned by user")
}

// TestDialogFavoriteAppService_CreateDialogFavorite_AlreadyFavorited tests already favorited session.
func TestDialogFavoriteAppService_CreateDialogFavorite_AlreadyFavorited(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFavoriteRepo := mocks.NewMockDialogFavoriteRepository(ctrl)
	mockDialogRepo := mocks.NewMockDialogRepository(ctrl)
	mockIDGenerator := mocks.NewMockSnowflakeGenerator(ctrl)

	svc := service.NewDialogFavoriteAppService(mockFavoriteRepo, mockDialogRepo, mockIDGenerator)
	ctx := context.Background()

	userID := int64(1)
	dialogSessionID := "100"
	title := "Test Favorite"

	// Mock dialog session retrieval
	session := &dialog.DialogSession{
		ID:          1,
		SnowflakeID: "100",
		UserID:      uint(userID),
		Status:      string(dialog.StatusActive),
		CreatedAt:   time.Now(),
	}
	mockDialogRepo.EXPECT().GetBySnowflakeID(ctx, "100").Return(session, nil)

	// Mock favorite existence check (already exists)
	mockFavoriteRepo.EXPECT().IsDialogFavoriteExists(ctx, userID, int64(100)).Return(true, nil)

	// Act
	favorite, err := svc.CreateDialogFavorite(ctx, userID, dialogSessionID, title)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, favorite)
	assert.Contains(t, err.Error(), "dialog session already favorited")
}

// TestDialogFavoriteAppService_CreateDialogFavorite_CreateFailure tests favorite creation failure.
func TestDialogFavoriteAppService_CreateDialogFavorite_CreateFailure(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFavoriteRepo := mocks.NewMockDialogFavoriteRepository(ctrl)
	mockDialogRepo := mocks.NewMockDialogRepository(ctrl)
	mockIDGenerator := mocks.NewMockSnowflakeGenerator(ctrl)

	svc := service.NewDialogFavoriteAppService(mockFavoriteRepo, mockDialogRepo, mockIDGenerator)
	ctx := context.Background()

	userID := int64(1)
	dialogSessionID := "100"
	title := "Test Favorite"

	// Mock dialog session retrieval
	session := &dialog.DialogSession{
		ID:          1,
		SnowflakeID: "100",
		UserID:      uint(userID),
		Status:      string(dialog.StatusActive),
		CreatedAt:   time.Now(),
	}
	mockDialogRepo.EXPECT().GetBySnowflakeID(ctx, "100").Return(session, nil)

	// Mock favorite existence check
	mockFavoriteRepo.EXPECT().IsDialogFavoriteExists(ctx, userID, int64(100)).Return(false, nil)

	// Mock ID generation
	favoriteSnowflakeID := int64(123456789)
	mockIDGenerator.EXPECT().Generate().Return(favoriteSnowflakeID)

	// Mock favorite creation (failure)
	mockFavoriteRepo.EXPECT().CreateDialogFavorite(ctx, gomock.Any()).Return(errors.New("database error"))

	// Act
	favorite, err := svc.CreateDialogFavorite(ctx, userID, dialogSessionID, title)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, favorite)
	assert.Contains(t, err.Error(), "failed to create dialog favorite")
}

// TestDialogFavoriteAppService_GetDialogFavoriteList_Success tests successful list retrieval.
func TestDialogFavoriteAppService_GetDialogFavoriteList_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFavoriteRepo := mocks.NewMockDialogFavoriteRepository(ctrl)
	mockDialogRepo := mocks.NewMockDialogRepository(ctrl)
	mockIDGenerator := mocks.NewMockSnowflakeGenerator(ctrl)

	svc := service.NewDialogFavoriteAppService(mockFavoriteRepo, mockDialogRepo, mockIDGenerator)
	ctx := context.Background()

	userID := int64(1)
	page := 1
	pageSize := 20

	now := time.Now()
	favorites := []dialog.DialogFavorite{
		{SnowflakeID: 123456789, UserID: userID, DialogSessionID: 100, Title: "Favorite 1", CreatedAt: now},
		{SnowflakeID: 123456790, UserID: userID, DialogSessionID: 101, Title: "Favorite 2", CreatedAt: now.Add(-1 * time.Hour)},
	}
	total := int64(2)

	// Mock repository call
	mockFavoriteRepo.EXPECT().GetDialogFavoritesByUserID(ctx, userID, 0, 20).Return(favorites, total, nil)

	// Act
	result, resultTotal, err := svc.GetDialogFavoriteList(ctx, userID, page, pageSize)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, total, resultTotal)
}

// TestDialogFavoriteAppService_GetDialogFavoriteList_PageAutoCorrection tests page auto correction.
func TestDialogFavoriteAppService_GetDialogFavoriteList_PageAutoCorrection(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFavoriteRepo := mocks.NewMockDialogFavoriteRepository(ctrl)
	mockDialogRepo := mocks.NewMockDialogRepository(ctrl)
	mockIDGenerator := mocks.NewMockSnowflakeGenerator(ctrl)

	svc := service.NewDialogFavoriteAppService(mockFavoriteRepo, mockDialogRepo, mockIDGenerator)
	ctx := context.Background()

	userID := int64(1)
	page := 0 // Invalid, should be corrected to 1
	pageSize := 20

	// Mock repository call (offset should be 0 after correction)
	mockFavoriteRepo.EXPECT().GetDialogFavoritesByUserID(ctx, userID, 0, 20).Return([]dialog.DialogFavorite{}, int64(0), nil)

	// Act
	result, _, err := svc.GetDialogFavoriteList(ctx, userID, page, pageSize)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// TestDialogFavoriteAppService_GetDialogFavoriteList_PageSizeAutoCorrection tests pageSize auto correction.
func TestDialogFavoriteAppService_GetDialogFavoriteList_PageSizeAutoCorrection(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFavoriteRepo := mocks.NewMockDialogFavoriteRepository(ctrl)
	mockDialogRepo := mocks.NewMockDialogRepository(ctrl)
	mockIDGenerator := mocks.NewMockSnowflakeGenerator(ctrl)

	svc := service.NewDialogFavoriteAppService(mockFavoriteRepo, mockDialogRepo, mockIDGenerator)
	ctx := context.Background()

	userID := int64(1)
	page := 1
	pageSize := 0 // Invalid, should be corrected to 20

	// Mock repository call (pageSize should be 20 after correction)
	mockFavoriteRepo.EXPECT().GetDialogFavoritesByUserID(ctx, userID, 0, 20).Return([]dialog.DialogFavorite{}, int64(0), nil)

	// Act
	result, _, err := svc.GetDialogFavoriteList(ctx, userID, page, pageSize)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// TestDialogFavoriteAppService_GetDialogFavoriteList_PageSizeTooLarge tests pageSize too large correction.
func TestDialogFavoriteAppService_GetDialogFavoriteList_PageSizeTooLarge(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFavoriteRepo := mocks.NewMockDialogFavoriteRepository(ctrl)
	mockDialogRepo := mocks.NewMockDialogRepository(ctrl)
	mockIDGenerator := mocks.NewMockSnowflakeGenerator(ctrl)

	svc := service.NewDialogFavoriteAppService(mockFavoriteRepo, mockDialogRepo, mockIDGenerator)
	ctx := context.Background()

	userID := int64(1)
	page := 1
	pageSize := 200 // Too large, should be corrected to 20

	// Mock repository call (pageSize should be 20 after correction)
	mockFavoriteRepo.EXPECT().GetDialogFavoritesByUserID(ctx, userID, 0, 20).Return([]dialog.DialogFavorite{}, int64(0), nil)

	// Act
	result, _, err := svc.GetDialogFavoriteList(ctx, userID, page, pageSize)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// TestDialogFavoriteAppService_GetDialogFavoriteList_RepositoryError tests repository error.
func TestDialogFavoriteAppService_GetDialogFavoriteList_RepositoryError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFavoriteRepo := mocks.NewMockDialogFavoriteRepository(ctrl)
	mockDialogRepo := mocks.NewMockDialogRepository(ctrl)
	mockIDGenerator := mocks.NewMockSnowflakeGenerator(ctrl)

	svc := service.NewDialogFavoriteAppService(mockFavoriteRepo, mockDialogRepo, mockIDGenerator)
	ctx := context.Background()

	userID := int64(1)
	page := 1
	pageSize := 20

	// Mock repository call (error)
	mockFavoriteRepo.EXPECT().GetDialogFavoritesByUserID(ctx, userID, 0, 20).Return(nil, int64(0), errors.New("database error"))

	// Act
	result, resultTotal, err := svc.GetDialogFavoriteList(ctx, userID, page, pageSize)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, int64(0), resultTotal)
	assert.Contains(t, err.Error(), "failed to get dialog favorite list")
}

// TestDialogFavoriteAppService_DeleteDialogFavorite_Success tests successful deletion.
func TestDialogFavoriteAppService_DeleteDialogFavorite_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFavoriteRepo := mocks.NewMockDialogFavoriteRepository(ctrl)
	mockDialogRepo := mocks.NewMockDialogRepository(ctrl)
	mockIDGenerator := mocks.NewMockSnowflakeGenerator(ctrl)

	svc := service.NewDialogFavoriteAppService(mockFavoriteRepo, mockDialogRepo, mockIDGenerator)
	ctx := context.Background()

	userID := int64(1)
	favoriteID := "123456789"

	// Mock repository deletion
	mockFavoriteRepo.EXPECT().DeleteDialogFavorite(ctx, int64(123456789), userID).Return(nil)

	// Act
	err := svc.DeleteDialogFavorite(ctx, userID, favoriteID)

	// Assert
	assert.NoError(t, err)
}

// TestDialogFavoriteAppService_DeleteDialogFavorite_InvalidIDFormat tests invalid favorite ID format.
func TestDialogFavoriteAppService_DeleteDialogFavorite_InvalidIDFormat(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFavoriteRepo := mocks.NewMockDialogFavoriteRepository(ctrl)
	mockDialogRepo := mocks.NewMockDialogRepository(ctrl)
	mockIDGenerator := mocks.NewMockSnowflakeGenerator(ctrl)

	svc := service.NewDialogFavoriteAppService(mockFavoriteRepo, mockDialogRepo, mockIDGenerator)
	ctx := context.Background()

	userID := int64(1)
	favoriteID := "abc" // Invalid format

	// Act
	err := svc.DeleteDialogFavorite(ctx, userID, favoriteID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid favorite id format")
}

// TestDialogFavoriteAppService_DeleteDialogFavorite_DeletionFailure tests deletion failure.
func TestDialogFavoriteAppService_DeleteDialogFavorite_DeletionFailure(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFavoriteRepo := mocks.NewMockDialogFavoriteRepository(ctrl)
	mockDialogRepo := mocks.NewMockDialogRepository(ctrl)
	mockIDGenerator := mocks.NewMockSnowflakeGenerator(ctrl)

	svc := service.NewDialogFavoriteAppService(mockFavoriteRepo, mockDialogRepo, mockIDGenerator)
	ctx := context.Background()

	userID := int64(1)
	favoriteID := "123456789"

	// Mock repository deletion (failure)
	mockFavoriteRepo.EXPECT().DeleteDialogFavorite(ctx, int64(123456789), userID).Return(errors.New("not found"))

	// Act
	err := svc.DeleteDialogFavorite(ctx, userID, favoriteID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete dialog favorite")
}
