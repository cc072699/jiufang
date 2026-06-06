// Package service_test implements unit tests for history application service.
package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"jiufang/internal/mocks"
	"jiufang/internal/model/query"
	"jiufang/internal/service"
)

// TestHistoryAppService_GetHistoryList_Success tests successful history list retrieval.
func TestHistoryAppService_GetHistoryList_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	service := service.NewHistoryAppService(mockRepo, mockIDGen)

	ctx := context.Background()
	userID := int64(1)
	page := 1
	pageSize := 20

	now := query.QueryRecord{
		SnowflakeID:   123456789,
		SessionID:     1,
		UserID:        userID,
		Input:         "test query",
		SQL:           "SELECT * FROM test",
		Status:        query.QueryStatusSuccess,
		ResultCount:   10,
		ExecutionTime: 100,
	}

	records := []query.QueryRecord{now}
	total := int64(1)

	mockRepo.EXPECT().
		GetQueryRecordsByUserID(ctx, userID, 0, 20, "", "", "").
		Return(records, total, nil)

	// Act
	result, count, err := service.GetHistoryList(ctx, userID, page, pageSize, "", "", "")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, total, count)
	assert.Len(t, result, 1)
}

// TestHistoryAppService_GetHistoryList_DefaultPagination tests default pagination values.
func TestHistoryAppService_GetHistoryList_DefaultPagination(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	service := service.NewHistoryAppService(mockRepo, mockIDGen)

	ctx := context.Background()
	userID := int64(1)

	mockRepo.EXPECT().
		GetQueryRecordsByUserID(ctx, userID, 0, 20, "", "", "").
		Return([]query.QueryRecord{}, int64(0), nil)

	// Act - pass page=0 and pageSize=0 to trigger defaults
	result, count, err := service.GetHistoryList(ctx, userID, 0, 0, "", "", "")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
	assert.Len(t, result, 0)
}

// TestHistoryAppService_GetHistoryList_MaxPageSize tests max page size limit.
func TestHistoryAppService_GetHistoryList_MaxPageSize(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	service := service.NewHistoryAppService(mockRepo, mockIDGen)

	ctx := context.Background()
	userID := int64(1)

	mockRepo.EXPECT().
		GetQueryRecordsByUserID(ctx, userID, 0, 20, "", "", "").
		Return([]query.QueryRecord{}, int64(0), nil)

	// Act - pass pageSize=200 to trigger max limit (will be capped at 20)
	_, count, err := service.GetHistoryList(ctx, userID, 1, 200, "", "", "")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// TestHistoryAppService_GetHistoryList_Failure tests history list retrieval failure.
func TestHistoryAppService_GetHistoryList_Failure(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	service := service.NewHistoryAppService(mockRepo, mockIDGen)

	ctx := context.Background()
	userID := int64(1)

	mockRepo.EXPECT().
		GetQueryRecordsByUserID(ctx, userID, 0, 20, "", "", "").
		Return(nil, int64(0), errors.New("database error"))

	// Act
	result, count, err := service.GetHistoryList(ctx, userID, 1, 20, "", "", "")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get history list")
	assert.Nil(t, result)
	assert.Equal(t, int64(0), count)
}

// TestHistoryAppService_GetHistoryDetail_Success tests successful history detail retrieval.
func TestHistoryAppService_GetHistoryDetail_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	service := service.NewHistoryAppService(mockRepo, mockIDGen)

	ctx := context.Background()
	userID := int64(1)
	recordID := "123456789"

	record := &query.QueryRecord{
		SnowflakeID:   123456789,
		SessionID:     1,
		UserID:        userID,
		Input:         "test query",
		SQL:           "SELECT * FROM test",
		Status:        query.QueryStatusSuccess,
		ResultCount:   10,
		ExecutionTime: 100,
	}

	mockRepo.EXPECT().
		GetQueryRecordBySnowflakeID(ctx, int64(123456789)).
		Return(record, nil)

	// Act
	result, err := service.GetHistoryDetail(ctx, userID, recordID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(123456789), result.SnowflakeID)
}

// TestHistoryAppService_GetHistoryDetail_InvalidIDFormat tests invalid record ID format.
func TestHistoryAppService_GetHistoryDetail_InvalidIDFormat(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	service := service.NewHistoryAppService(mockRepo, mockIDGen)

	ctx := context.Background()
	userID := int64(1)
	recordID := "invalid-id"

	// Act
	result, err := service.GetHistoryDetail(ctx, userID, recordID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid record id format")
	assert.Nil(t, result)
}

// TestHistoryAppService_GetHistoryDetail_NotOwned tests history detail not owned by user.
func TestHistoryAppService_GetHistoryDetail_NotOwned(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	service := service.NewHistoryAppService(mockRepo, mockIDGen)

	ctx := context.Background()
	userID := int64(1)
	recordID := "123456789"

	// Record belongs to different user
	record := &query.QueryRecord{
		SnowflakeID: 123456789,
		SessionID:   1,
		UserID:      999, // Different user
		Input:       "test query",
		SQL:         "SELECT * FROM test",
		Status:      query.QueryStatusSuccess,
	}

	mockRepo.EXPECT().
		GetQueryRecordBySnowflakeID(ctx, int64(123456789)).
		Return(record, nil)

	// Act
	result, err := service.GetHistoryDetail(ctx, userID, recordID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query record not owned by user")
	assert.Nil(t, result)
}

// TestHistoryAppService_DeleteHistory_Success tests successful history deletion.
func TestHistoryAppService_DeleteHistory_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	service := service.NewHistoryAppService(mockRepo, mockIDGen)

	ctx := context.Background()
	userID := int64(1)
	recordID := "123456789"

	mockRepo.EXPECT().
		DeleteQueryRecord(ctx, int64(123456789), userID).
		Return(nil)

	// Act
	err := service.DeleteHistory(ctx, userID, recordID)

	// Assert
	assert.NoError(t, err)
}

// TestHistoryAppService_DeleteHistory_InvalidIDFormat tests invalid record ID format for deletion.
func TestHistoryAppService_DeleteHistory_InvalidIDFormat(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	service := service.NewHistoryAppService(mockRepo, mockIDGen)

	ctx := context.Background()
	userID := int64(1)
	recordID := "invalid-id"

	// Act
	err := service.DeleteHistory(ctx, userID, recordID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid record id format")
}

// TestHistoryAppService_CreateFavorite_Success tests successful favorite creation.
func TestHistoryAppService_CreateFavorite_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	service := service.NewHistoryAppService(mockRepo, mockIDGen)

	ctx := context.Background()
	userID := int64(1)
	name := "My Favorite"
	input := "test query"
	sql := "SELECT * FROM test"
	description := "This is a test favorite"

	favoriteSnowflakeID := int64(123456789)

	mockIDGen.EXPECT().
		Generate().
		Return(favoriteSnowflakeID)

	mockRepo.EXPECT().
		CreateFavorite(ctx, gomock.Any()).
		Return(nil)

	// Act
	favorite, err := service.CreateFavorite(ctx, userID, name, input, sql, description)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, favorite)
	assert.Equal(t, favoriteSnowflakeID, favorite.SnowflakeID)
	assert.Equal(t, userID, favorite.UserID)
	assert.Equal(t, name, favorite.Name)
	assert.Equal(t, input, favorite.Input)
	assert.Equal(t, sql, favorite.Sql)
	assert.Equal(t, description, favorite.Description)
}

// TestHistoryAppService_CreateFavorite_CreateError tests favorite creation failure.
func TestHistoryAppService_CreateFavorite_CreateError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	service := service.NewHistoryAppService(mockRepo, mockIDGen)

	ctx := context.Background()
	userID := int64(1)
	name := "My Favorite"
	input := "test query"
	sql := "SELECT * FROM test"
	description := "This is a test favorite"

	favoriteSnowflakeID := int64(123456789)

	mockIDGen.EXPECT().
		Generate().
		Return(favoriteSnowflakeID)

	mockRepo.EXPECT().
		CreateFavorite(ctx, gomock.Any()).
		Return(errors.New("database error"))

	// Act
	favorite, err := service.CreateFavorite(ctx, userID, name, input, sql, description)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create favorite")
	assert.Nil(t, favorite)
}

// TestHistoryAppService_GetFavoriteList_Success tests successful favorite list retrieval.
func TestHistoryAppService_GetFavoriteList_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	service := service.NewHistoryAppService(mockRepo, mockIDGen)

	ctx := context.Background()
	userID := int64(1)
	page := 1
	pageSize := 20
	name := ""

	favorites := []query.Favorite{
		{
			SnowflakeID: 123456789,
			UserID:      userID,
			Name:        "My Favorite",
			Input:       "test query",
			Sql:         "SELECT * FROM test",
			Description: "This is a test favorite",
		},
	}
	total := int64(1)

	mockRepo.EXPECT().
		GetFavoritesByUserID(ctx, userID, 0, 20, name).
		Return(favorites, total, nil)

	// Act
	result, count, err := service.GetFavoriteList(ctx, userID, page, pageSize, name)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, total, count)
	assert.Len(t, result, 1)
}

// TestHistoryAppService_GetFavoriteList_DefaultPagination tests default pagination for favorite list.
func TestHistoryAppService_GetFavoriteList_DefaultPagination(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	service := service.NewHistoryAppService(mockRepo, mockIDGen)

	ctx := context.Background()
	userID := int64(1)
	name := ""

	mockRepo.EXPECT().
		GetFavoritesByUserID(ctx, userID, 0, 20, name).
		Return([]query.Favorite{}, int64(0), nil)

	// Act - pass page=0 and pageSize=0 to trigger defaults
	result, count, err := service.GetFavoriteList(ctx, userID, 0, 0, name)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
	assert.Len(t, result, 0)
}

// TestHistoryAppService_DeleteFavorite_Success tests successful favorite deletion.
func TestHistoryAppService_DeleteFavorite_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	service := service.NewHistoryAppService(mockRepo, mockIDGen)

	ctx := context.Background()
	userID := int64(1)
	favoriteID := "123456789"

	mockRepo.EXPECT().
		DeleteFavorite(ctx, int64(123456789), userID).
		Return(nil)

	// Act
	err := service.DeleteFavorite(ctx, userID, favoriteID)

	// Assert
	assert.NoError(t, err)
}

// TestHistoryAppService_DeleteFavorite_InvalidIDFormat tests invalid favorite ID format.
func TestHistoryAppService_DeleteFavorite_InvalidIDFormat(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	service := service.NewHistoryAppService(mockRepo, mockIDGen)

	ctx := context.Background()
	userID := int64(1)
	favoriteID := "invalid-id"

	// Act
	err := service.DeleteFavorite(ctx, userID, favoriteID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid favorite id format")
}

// TestHistoryAppService_CreateQuerySession_Success tests successful session creation.
func TestHistoryAppService_CreateQuerySession_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	service := service.NewHistoryAppService(mockRepo, mockIDGen)

	ctx := context.Background()
	userID := int64(1)
	dialogID := int64(100)

	sessionSnowflakeID := int64(123456789)

	mockIDGen.EXPECT().
		Generate().
		Return(sessionSnowflakeID)

	mockRepo.EXPECT().
		CreateQuerySession(ctx, gomock.Any()).
		Return(nil)

	// Act
	session, err := service.CreateQuerySession(ctx, userID, dialogID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, sessionSnowflakeID, session.SnowflakeID)
	assert.Equal(t, userID, session.UserID)
	assert.Equal(t, dialogID, session.DialogID)
	assert.Equal(t, query.SessionStatusActive, session.Status)
}

// TestHistoryAppService_CloseQuerySession_Success tests successful session close.
func TestHistoryAppService_CloseQuerySession_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	service := service.NewHistoryAppService(mockRepo, mockIDGen)

	ctx := context.Background()
	userID := int64(1)
	sessionID := "123456789"

	session := &query.QuerySession{
		SnowflakeID: 123456789,
		UserID:      userID,
		Status:      query.SessionStatusActive,
	}

	mockRepo.EXPECT().
		GetQuerySessionBySnowflakeID(ctx, int64(123456789)).
		Return(session, nil)

	mockRepo.EXPECT().
		CloseQuerySession(ctx, int64(123456789)).
		Return(nil)

	// Act
	err := service.CloseQuerySession(ctx, userID, sessionID)

	// Assert
	assert.NoError(t, err)
}

// TestHistoryAppService_CloseQuerySession_InvalidIDFormat tests invalid session ID format.
func TestHistoryAppService_CloseQuerySession_InvalidIDFormat(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	service := service.NewHistoryAppService(mockRepo, mockIDGen)

	ctx := context.Background()
	userID := int64(1)
	sessionID := "invalid-id"

	// Act
	err := service.CloseQuerySession(ctx, userID, sessionID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid session id format")
}

// TestHistoryAppService_CloseQuerySession_NotOwned tests session close not owned by user.
func TestHistoryAppService_CloseQuerySession_NotOwned(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	service := service.NewHistoryAppService(mockRepo, mockIDGen)

	ctx := context.Background()
	userID := int64(1)
	sessionID := "123456789"

	// Session belongs to different user
	session := &query.QuerySession{
		SnowflakeID: 123456789,
		UserID:      999, // Different user
		Status:      query.SessionStatusActive,
	}

	mockRepo.EXPECT().
		GetQuerySessionBySnowflakeID(ctx, int64(123456789)).
		Return(session, nil)

	// Act
	err := service.CloseQuerySession(ctx, userID, sessionID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query session not owned by user")
}
