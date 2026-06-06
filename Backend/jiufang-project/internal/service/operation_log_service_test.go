// Package service implements the operation log service unit tests.
package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"jiufang/internal/model/audit"
	"jiufang/internal/model/user"
	"jiufang/internal/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestOperationLogService_ListOperationLogs_Success tests successful list retrieval.
func TestOperationLogService_ListOperationLogs_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogRepo := mocks.NewMockOperationLogRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()

	service := NewOperationLogService(mockLogRepo, mockUserRepo, logger)
	ctx := context.Background()

	now := time.Now()
	userID := int64(1001)
	expectedLogs := []audit.OperationLog{
		{
			SnowflakeID:     123456789,
			UserID:          &userID,
			OperationType:   audit.OperationTypeLogin,
			OperationResult: audit.OperationResultSuccess,
			CreatedAt:       now,
		},
	}
	expectedTotal := int64(1)

	// Mock repository List call
	mockLogRepo.EXPECT().
		List(ctx, 0, 10, int64(0), "", "", "").
		Return(expectedLogs, expectedTotal, nil)

	// Act
	logs, total, err := service.ListOperationLogs(ctx, 1, 10, 0, "", "", "")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedTotal, total)
	assert.Len(t, logs, 1)
	assert.Equal(t, expectedLogs[0].SnowflakeID, logs[0].SnowflakeID)
}

// TestOperationLogService_ListOperationLogs_PaginationCalculation tests pagination offset calculation.
func TestOperationLogService_ListOperationLogs_PaginationCalculation(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogRepo := mocks.NewMockOperationLogRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()

	service := NewOperationLogService(mockLogRepo, mockUserRepo, logger)
	ctx := context.Background()

	expectedLogs := []audit.OperationLog{}
	expectedTotal := int64(0)

	// Mock repository List call with offset=10 (page=2, size=10)
	mockLogRepo.EXPECT().
		List(ctx, 10, 10, int64(0), "", "", "").
		Return(expectedLogs, expectedTotal, nil)

	// Act
	logs, total, err := service.ListOperationLogs(ctx, 2, 10, 0, "", "", "")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedTotal, total)
	assert.Len(t, logs, 0)
}

// TestOperationLogService_ListOperationLogs_RepositoryError tests list retrieval with repository error.
func TestOperationLogService_ListOperationLogs_RepositoryError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogRepo := mocks.NewMockOperationLogRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()

	service := NewOperationLogService(mockLogRepo, mockUserRepo, logger)
	ctx := context.Background()

	expectedError := errors.New("database connection error")

	// Mock repository List call with error
	mockLogRepo.EXPECT().
		List(ctx, 0, 10, int64(0), "", "", "").
		Return(nil, int64(0), expectedError)

	// Act
	logs, total, err := service.ListOperationLogs(ctx, 1, 10, 0, "", "", "")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list operation logs")
	assert.Contains(t, err.Error(), expectedError.Error())
	assert.Equal(t, int64(0), total)
	assert.Nil(t, logs)
}

// TestOperationLogService_GetOperationLogByID_Success tests successful retrieval by snowflake ID.
func TestOperationLogService_GetOperationLogByID_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogRepo := mocks.NewMockOperationLogRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()

	service := NewOperationLogService(mockLogRepo, mockUserRepo, logger)
	ctx := context.Background()

	now := time.Now()
	userID := int64(1001)
	expectedLog := &audit.OperationLog{
		SnowflakeID:     123456789,
		UserID:          &userID,
		OperationType:   audit.OperationTypeLogin,
		OperationResult: audit.OperationResultSuccess,
		CreatedAt:       now,
	}

	// Mock repository GetBySnowflakeID call
	mockLogRepo.EXPECT().
		GetBySnowflakeID(ctx, int64(123456789)).
		Return(expectedLog, nil)

	// Act
	log, err := service.GetOperationLogByID(ctx, 123456789)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, log)
	assert.Equal(t, expectedLog.SnowflakeID, log.SnowflakeID)
	assert.Equal(t, expectedLog.OperationType, log.OperationType)
}

// TestOperationLogService_GetOperationLogByID_NotFound tests retrieval when log not found.
func TestOperationLogService_GetOperationLogByID_NotFound(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogRepo := mocks.NewMockOperationLogRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()

	service := NewOperationLogService(mockLogRepo, mockUserRepo, logger)
	ctx := context.Background()

	// Mock repository GetBySnowflakeID call returning nil
	mockLogRepo.EXPECT().
		GetBySnowflakeID(ctx, int64(999999)).
		Return(nil, nil)

	// Act
	log, err := service.GetOperationLogByID(ctx, 999999)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation log not found")
	assert.Nil(t, log)
}

// TestOperationLogService_GetOperationLogByID_RepositoryError tests retrieval with repository error.
func TestOperationLogService_GetOperationLogByID_RepositoryError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogRepo := mocks.NewMockOperationLogRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()

	service := NewOperationLogService(mockLogRepo, mockUserRepo, logger)
	ctx := context.Background()

	expectedError := errors.New("database connection error")

	// Mock repository GetBySnowflakeID call with error
	mockLogRepo.EXPECT().
		GetBySnowflakeID(ctx, int64(123456789)).
		Return(nil, expectedError)

	// Act
	log, err := service.GetOperationLogByID(ctx, 123456789)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get operation log")
	assert.Contains(t, err.Error(), expectedError.Error())
	assert.Nil(t, log)
}

// TestOperationLogService_GetUsernameByUserID_Success tests successful retrieval of username.
func TestOperationLogService_GetUsernameByUserID_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogRepo := mocks.NewMockOperationLogRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()

	service := NewOperationLogService(mockLogRepo, mockUserRepo, logger)
	ctx := context.Background()

	expectedUser := &user.User{
		SnowflakeID: 1001,
		Username:    "admin",
	}

	// Mock repository GetBySnowflakeID call
	mockUserRepo.EXPECT().
		GetBySnowflakeID(ctx, int64(1001)).
		Return(expectedUser, nil)

	// Act
	username, err := service.GetUsernameByUserID(ctx, 1001)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "admin", username)
}

// TestOperationLogService_GetUsernameByUserID_UserNotFound tests retrieval when user not found.
func TestOperationLogService_GetUsernameByUserID_UserNotFound(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogRepo := mocks.NewMockOperationLogRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()

	service := NewOperationLogService(mockLogRepo, mockUserRepo, logger)
	ctx := context.Background()

	// Mock repository GetBySnowflakeID call returning nil
	mockUserRepo.EXPECT().
		GetBySnowflakeID(ctx, int64(999999)).
		Return(nil, nil)

	// Act
	username, err := service.GetUsernameByUserID(ctx, 999999)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "", username)
}

// TestOperationLogService_GetUsernameByUserID_RepositoryError tests retrieval with repository error.
func TestOperationLogService_GetUsernameByUserID_RepositoryError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogRepo := mocks.NewMockOperationLogRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()

	service := NewOperationLogService(mockLogRepo, mockUserRepo, logger)
	ctx := context.Background()

	expectedError := errors.New("database connection error")

	// Mock repository GetBySnowflakeID call with error
	mockUserRepo.EXPECT().
		GetBySnowflakeID(ctx, int64(1001)).
		Return(nil, expectedError)

	// Act
	username, err := service.GetUsernameByUserID(ctx, 1001)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user")
	assert.Contains(t, err.Error(), expectedError.Error())
	assert.Equal(t, "", username)
}