// Package service_test implements unit tests for FeedbackService.
// Author: AI Assistant
// Date: 2026-06-04
// Description: Unit tests for feedback business service

package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"jiufang/internal/mocks"
	"jiufang/internal/model/feedback"
	"jiufang/internal/model/query"
	"jiufang/internal/service"
)

// TestFeedbackService_CreateFeedback_Satisfied_ShouldSucceed tests TC-FS-01
func TestFeedbackService_CreateFeedback_Satisfied_ShouldSucceed(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFeedbackRepo := mocks.NewMockFeedbackRepository(ctrl)
	mockQueryRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()

	svc := service.NewFeedbackService(mockFeedbackRepo, mockQueryRepo, mockIDGen, logger)

	ctx := context.Background()
	userID := int64(1001)
	queryRecordID := int64(987654321)
	expectedSnowflakeID := int64(123456789)

	req := &feedback.CreateFeedbackRequest{
		QueryRecordID: queryRecordID,
		Rating:        "satisfied",
		Reason:        "",
	}

	queryRecord := &query.QueryRecord{
		SnowflakeID: queryRecordID,
		UserID:      userID,
		Input:       "查询本月销售额",
	}

	// Mock expectations
	mockFeedbackRepo.EXPECT().IsFeedbackExists(ctx, queryRecordID).Return(false, nil)
	mockQueryRepo.EXPECT().GetQueryRecordBySnowflakeID(ctx, queryRecordID).Return(queryRecord, nil)
	mockIDGen.EXPECT().Generate().Return(expectedSnowflakeID)
	mockFeedbackRepo.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, f *feedback.Feedback) error {
			assert.Equal(t, expectedSnowflakeID, f.SnowflakeID)
			assert.Equal(t, userID, f.UserID)
			assert.Equal(t, queryRecordID, f.QueryRecordID)
			assert.Equal(t, "查询本月销售额", f.QueryQuestion)
			assert.Equal(t, feedback.RatingSatisfied, f.Rating)
			return nil
		},
	)

	// Act
	result, err := svc.CreateFeedback(ctx, userID, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedSnowflakeID, result.SnowflakeID)
	assert.Equal(t, feedback.RatingSatisfied, result.Rating)
}

// TestFeedbackService_CreateFeedback_Unsatisfied_ShouldSucceed tests TC-FS-02
func TestFeedbackService_CreateFeedback_Unsatisfied_ShouldSucceed(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFeedbackRepo := mocks.NewMockFeedbackRepository(ctrl)
	mockQueryRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()

	svc := service.NewFeedbackService(mockFeedbackRepo, mockQueryRepo, mockIDGen, logger)

	ctx := context.Background()
	userID := int64(1001)
	queryRecordID := int64(987654321)
	expectedSnowflakeID := int64(123456789)

	req := &feedback.CreateFeedbackRequest{
		QueryRecordID: queryRecordID,
		Rating:        "unsatisfied",
		Reason:        "查询结果不准确",
	}

	queryRecord := &query.QueryRecord{
		SnowflakeID: queryRecordID,
		UserID:      userID,
		Input:       "查询本月销售额",
	}

	// Mock expectations
	mockFeedbackRepo.EXPECT().IsFeedbackExists(ctx, queryRecordID).Return(false, nil)
	mockQueryRepo.EXPECT().GetQueryRecordBySnowflakeID(ctx, queryRecordID).Return(queryRecord, nil)
	mockIDGen.EXPECT().Generate().Return(expectedSnowflakeID)
	mockFeedbackRepo.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, f *feedback.Feedback) error {
			assert.Equal(t, feedback.RatingUnsatisfied, f.Rating)
			assert.Equal(t, "查询结果不准确", f.Reason)
			return nil
		},
	)

	// Act
	result, err := svc.CreateFeedback(ctx, userID, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, feedback.RatingUnsatisfied, result.Rating)
	assert.Equal(t, "查询结果不准确", result.Reason)
}

// TestFeedbackService_CreateFeedback_InvalidRating_ShouldReturnError tests TC-FS-03
func TestFeedbackService_CreateFeedback_InvalidRating_ShouldReturnError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFeedbackRepo := mocks.NewMockFeedbackRepository(ctrl)
	mockQueryRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()

	svc := service.NewFeedbackService(mockFeedbackRepo, mockQueryRepo, mockIDGen, logger)

	ctx := context.Background()
	userID := int64(1001)

	req := &feedback.CreateFeedbackRequest{
		QueryRecordID: 987654321,
		Rating:        "invalid_rating",
		Reason:        "",
	}

	// Act
	result, err := svc.CreateFeedback(ctx, userID, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid rating value")
}

// TestFeedbackService_CreateFeedback_UnsatisfiedWithoutReason_ShouldReturnError tests TC-FS-04
func TestFeedbackService_CreateFeedback_UnsatisfiedWithoutReason_ShouldReturnError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFeedbackRepo := mocks.NewMockFeedbackRepository(ctrl)
	mockQueryRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()

	svc := service.NewFeedbackService(mockFeedbackRepo, mockQueryRepo, mockIDGen, logger)

	ctx := context.Background()
	userID := int64(1001)

	req := &feedback.CreateFeedbackRequest{
		QueryRecordID: 987654321,
		Rating:        "unsatisfied",
		Reason:        "",
	}

	// Act
	result, err := svc.CreateFeedback(ctx, userID, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "reason is required")
}

// TestFeedbackService_CreateFeedback_ReasonTooLong_ShouldReturnError tests TC-FS-05
func TestFeedbackService_CreateFeedback_ReasonTooLong_ShouldReturnError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFeedbackRepo := mocks.NewMockFeedbackRepository(ctrl)
	mockQueryRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()

	svc := service.NewFeedbackService(mockFeedbackRepo, mockQueryRepo, mockIDGen, logger)

	ctx := context.Background()
	userID := int64(1001)

	longReason := ""
	for i := 0; i < 501; i++ {
		longReason += "a"
	}

	req := &feedback.CreateFeedbackRequest{
		QueryRecordID: 987654321,
		Rating:        "unsatisfied",
		Reason:        longReason,
	}

	// Act
	result, err := svc.CreateFeedback(ctx, userID, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "reason length must not exceed")
}

// TestFeedbackService_CreateFeedback_AlreadyExists_ShouldReturnError tests TC-FS-06
func TestFeedbackService_CreateFeedback_AlreadyExists_ShouldReturnError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFeedbackRepo := mocks.NewMockFeedbackRepository(ctrl)
	mockQueryRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()

	svc := service.NewFeedbackService(mockFeedbackRepo, mockQueryRepo, mockIDGen, logger)

	ctx := context.Background()
	userID := int64(1001)
	queryRecordID := int64(987654321)

	req := &feedback.CreateFeedbackRequest{
		QueryRecordID: queryRecordID,
		Rating:        "satisfied",
		Reason:        "",
	}

	// Mock expectations
	mockFeedbackRepo.EXPECT().IsFeedbackExists(ctx, queryRecordID).Return(true, nil)

	// Act
	result, err := svc.CreateFeedback(ctx, userID, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "feedback already exists")
}

// TestFeedbackService_CreateFeedback_QueryRecordNotBelongToUser_ShouldReturnError tests TC-FS-08
func TestFeedbackService_CreateFeedback_QueryRecordNotBelongToUser_ShouldReturnError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFeedbackRepo := mocks.NewMockFeedbackRepository(ctrl)
	mockQueryRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()

	svc := service.NewFeedbackService(mockFeedbackRepo, mockQueryRepo, mockIDGen, logger)

	ctx := context.Background()
	userID := int64(1001)
	queryRecordID := int64(987654321)

	req := &feedback.CreateFeedbackRequest{
		QueryRecordID: queryRecordID,
		Rating:        "satisfied",
		Reason:        "",
	}

	queryRecord := &query.QueryRecord{
		SnowflakeID: queryRecordID,
		UserID:      1002, // Different user
		Input:       "查询本月销售额",
	}

	// Mock expectations
	mockFeedbackRepo.EXPECT().IsFeedbackExists(ctx, queryRecordID).Return(false, nil)
	mockQueryRepo.EXPECT().GetQueryRecordBySnowflakeID(ctx, queryRecordID).Return(queryRecord, nil)

	// Act
	result, err := svc.CreateFeedback(ctx, userID, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "query record does not belong")
}

// TestFeedbackService_GetFeedbackByID_Success tests TC-FS-09
func TestFeedbackService_GetFeedbackByID_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFeedbackRepo := mocks.NewMockFeedbackRepository(ctrl)
	mockQueryRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()

	svc := service.NewFeedbackService(mockFeedbackRepo, mockQueryRepo, mockIDGen, logger)

	ctx := context.Background()
	snowflakeID := int64(123456789)

	expectedFeedback := &feedback.Feedback{
		SnowflakeID:   snowflakeID,
		UserID:        1001,
		QueryRecordID: 987654321,
		QueryQuestion: "查询本月销售额",
		Rating:        feedback.RatingSatisfied,
		CreatedAt:     time.Now(),
	}

	// Mock expectations
	mockFeedbackRepo.EXPECT().GetBySnowflakeID(ctx, snowflakeID).Return(expectedFeedback, nil)

	// Act
	result, err := svc.GetFeedbackByID(ctx, snowflakeID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, snowflakeID, result.SnowflakeID)
}

// TestFeedbackService_ListFeedbacks_Success tests TC-FS-10
func TestFeedbackService_ListFeedbacks_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFeedbackRepo := mocks.NewMockFeedbackRepository(ctrl)
	mockQueryRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()

	svc := service.NewFeedbackService(mockFeedbackRepo, mockQueryRepo, mockIDGen, logger)

	ctx := context.Background()

	req := &feedback.ListFeedbacksRequest{
		Page:   1,
		Size:   10,
		UserID: 1001,
		Rating: "satisfied",
	}

	feedbacks := []feedback.Feedback{
		{
			SnowflakeID:   123456789,
			UserID:        1001,
			QueryRecordID: 987654321,
			QueryQuestion: "查询本月销售额",
			Rating:        feedback.RatingSatisfied,
			CreatedAt:     time.Now(),
		},
	}

	// Mock expectations
	mockFeedbackRepo.EXPECT().List(ctx, 0, 10, int64(1001), "satisfied").Return(feedbacks, int64(1), nil)

	// Act
	result, total, err := svc.ListFeedbacks(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
}

// TestFeedbackService_ListFeedbacks_DefaultPagination tests TC-FS-11
func TestFeedbackService_ListFeedbacks_DefaultPagination(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFeedbackRepo := mocks.NewMockFeedbackRepository(ctrl)
	mockQueryRepo := mocks.NewMockQueryRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()

	svc := service.NewFeedbackService(mockFeedbackRepo, mockQueryRepo, mockIDGen, logger)

	ctx := context.Background()

	req := &feedback.ListFeedbacksRequest{
		Page: 0, // Invalid, should be corrected to 1
		Size: 0, // Invalid, should be corrected to 10
	}

	feedbacks := []feedback.Feedback{}

	// Mock expectations - should use corrected values (offset=0, limit=10)
	mockFeedbackRepo.EXPECT().List(ctx, 0, 10, int64(0), "").Return(feedbacks, int64(0), nil)

	// Act
	result, total, err := svc.ListFeedbacks(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, result, 0)
}
