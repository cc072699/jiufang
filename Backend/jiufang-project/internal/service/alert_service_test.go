// Package service implements the alert service unit tests.
package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"jiufang/internal/agent"
	"jiufang/internal/mocks"
	"jiufang/internal/model/report"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestAlertService_CreateAlert_Success tests successful creation of an alert.
func TestAlertService_CreateAlert_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAlertRepo := mocks.NewMockAlertRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()
	sqlValidator := agent.NewSQLValidator()

	svc := NewAlertService(mockAlertRepo, mockIDGen, sqlValidator, logger)

	ctx := context.Background()
	expectedSnowflakeID := int64(123456789)

	req := &report.CreateAlertRequest{
		Name:             "库存低于安全库存预警",
		Description:      "当库存低于100时触发预警",
		SQL:              "SELECT SUM(quantity) as inventory FROM products WHERE status='active'",
		Condition:        "< 100",
		Recipients:       "[1001, 1002]",
		PushChannel:      "wechat",
		TriggerFrequency: "every_time",
		CreatedBy:        1001,
	}

	// Mock expectations
	mockIDGen.EXPECT().Generate().Return(expectedSnowflakeID)
	mockAlertRepo.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, a *report.Alert) error {
			assert.Equal(t, expectedSnowflakeID, a.SnowflakeID)
			assert.Equal(t, req.Name, a.Name)
			assert.Equal(t, report.AlertStatusActive, a.Status)
			return nil
		},
	)

	// Act
	result, err := svc.CreateAlert(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedSnowflakeID, result.SnowflakeID)
	assert.Equal(t, req.Name, result.Name)
	assert.Equal(t, report.AlertStatusActive, result.Status)
}

// TestAlertService_CreateAlert_SQLNotReadOnly tests creation with non-read-only SQL.
func TestAlertService_CreateAlert_SQLNotReadOnly(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAlertRepo := mocks.NewMockAlertRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()
	sqlValidator := agent.NewSQLValidator()

	svc := NewAlertService(mockAlertRepo, mockIDGen, sqlValidator, logger)

	ctx := context.Background()

	req := &report.CreateAlertRequest{
		Name:        "测试预警",
		SQL:         "DELETE FROM products WHERE id = 1",
		Condition:   "< 100",
		Recipients:  "[1001]",
		PushChannel: "wechat",
		CreatedBy:   1001,
	}

	// Act
	result, err := svc.CreateAlert(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "SQL must be read-only (SELECT statement only)", err.Error())
	assert.Nil(t, result)
}

// TestAlertService_CreateAlert_InvalidSQLSyntax tests creation with invalid SQL syntax.
func TestAlertService_CreateAlert_InvalidSQLSyntax(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAlertRepo := mocks.NewMockAlertRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()
	sqlValidator := agent.NewSQLValidator()

	svc := NewAlertService(mockAlertRepo, mockIDGen, sqlValidator, logger)

	ctx := context.Background()

	req := &report.CreateAlertRequest{
		Name:        "测试预警",
		SQL:         "SELECT * FROM products; DROP TABLE users", // SQL injection attempt
		Condition:   "< 100",
		Recipients:  "[1001]",
		PushChannel: "wechat",
		CreatedBy:   1001,
	}

	// Act
	result, err := svc.CreateAlert(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SQL must be read-only")
	assert.Nil(t, result)
}

// TestAlertService_CreateAlert_InvalidSilenceTimeFormat tests creation with invalid silence time format.
func TestAlertService_CreateAlert_InvalidSilenceTimeFormat(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAlertRepo := mocks.NewMockAlertRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()
	sqlValidator := agent.NewSQLValidator()

	svc := NewAlertService(mockAlertRepo, mockIDGen, sqlValidator, logger)

	ctx := context.Background()

	req := &report.CreateAlertRequest{
		Name:             "测试预警",
		SQL:              "SELECT SUM(quantity) FROM products",
		Condition:        "< 100",
		Recipients:       "[1001]",
		PushChannel:      "wechat",
		TriggerFrequency: "every_time",
		SilenceStart:     "25:00", // Invalid time format
		CreatedBy:        1001,
	}

	// Mock expectations - Generate will be called before parsing silence time
	mockIDGen.EXPECT().Generate().Return(int64(123456789))

	// Act
	result, err := svc.CreateAlert(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid silence_start format")
	assert.Nil(t, result)
}

// TestAlertService_GetAlertByID_Success tests successful retrieval by ID.
func TestAlertService_GetAlertByID_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAlertRepo := mocks.NewMockAlertRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()
	sqlValidator := agent.NewSQLValidator()

	svc := NewAlertService(mockAlertRepo, mockIDGen, sqlValidator, logger)

	ctx := context.Background()
	snowflakeID := int64(123456789)

	expectedAlert := &report.Alert{
		SnowflakeID: snowflakeID,
		Name:        "库存预警",
		Status:      report.AlertStatusActive,
		CreatedBy:   1001,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Mock expectations
	mockAlertRepo.EXPECT().GetBySnowflakeID(ctx, snowflakeID).Return(expectedAlert, nil)

	// Act
	result, err := svc.GetAlertByID(ctx, snowflakeID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, snowflakeID, result.SnowflakeID)
	assert.Equal(t, expectedAlert.Name, result.Name)
}

// TestAlertService_GetAlertByID_NotFound tests retrieval when alert not found.
func TestAlertService_GetAlertByID_NotFound(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAlertRepo := mocks.NewMockAlertRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()
	sqlValidator := agent.NewSQLValidator()

	svc := NewAlertService(mockAlertRepo, mockIDGen, sqlValidator, logger)

	ctx := context.Background()
	snowflakeID := int64(123456789)

	// Mock expectations
	mockAlertRepo.EXPECT().GetBySnowflakeID(ctx, snowflakeID).Return(nil, nil)

	// Act
	result, err := svc.GetAlertByID(ctx, snowflakeID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "alert not found", err.Error())
	assert.Nil(t, result)
}

// TestAlertService_ListAlerts_Success tests successful list retrieval.
func TestAlertService_ListAlerts_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAlertRepo := mocks.NewMockAlertRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()
	sqlValidator := agent.NewSQLValidator()

	svc := NewAlertService(mockAlertRepo, mockIDGen, sqlValidator, logger)

	ctx := context.Background()
	page := 1
	pageSize := 20
	name := "预警"
	status := "active"

	now := time.Now()
	alerts := []report.Alert{
		{
			SnowflakeID: 123456789,
			Name:        "库存预警",
			Status:      report.AlertStatusActive,
			CreatedBy:   1001,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
	total := int64(1)

	// Mock expectations
	mockAlertRepo.EXPECT().List(ctx, 0, pageSize, name, status).Return(alerts, total, nil)

	// Act
	result, resultTotal, err := svc.ListAlerts(ctx, page, pageSize, name, status)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, total, resultTotal)
	assert.Len(t, result, 1)
	assert.Equal(t, alerts[0].SnowflakeID, result[0].SnowflakeID)
}

// TestAlertService_UpdateAlert_Success tests successful update.
func TestAlertService_UpdateAlert_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAlertRepo := mocks.NewMockAlertRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()
	sqlValidator := agent.NewSQLValidator()

	svc := NewAlertService(mockAlertRepo, mockIDGen, sqlValidator, logger)

	ctx := context.Background()
	snowflakeID := int64(123456789)

	req := &report.UpdateAlertRequest{
		Name:             "更新后的预警",
		SQL:              "SELECT COUNT(*) FROM products",
		Condition:        "> 50",
		Recipients:       "[1001, 1002]",
		PushChannel:      "wechat",
		TriggerFrequency: "daily",
		Status:           "active",
	}

	now := time.Now()
	expectedAlert := &report.Alert{
		SnowflakeID: snowflakeID,
		Name:        req.Name,
		Status:      report.AlertStatus(req.Status),
		CreatedBy:   1001,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Mock expectations
	mockAlertRepo.EXPECT().Update(ctx, snowflakeID, gomock.Any()).Return(nil)
	mockAlertRepo.EXPECT().GetBySnowflakeID(ctx, snowflakeID).Return(expectedAlert, nil)

	// Act
	result, err := svc.UpdateAlert(ctx, snowflakeID, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, snowflakeID, result.SnowflakeID)
	assert.Equal(t, req.Name, result.Name)
}

// TestAlertService_UpdateAlert_SQLNotReadOnly tests update with non-read-only SQL.
func TestAlertService_UpdateAlert_SQLNotReadOnly(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAlertRepo := mocks.NewMockAlertRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()
	sqlValidator := agent.NewSQLValidator()

	svc := NewAlertService(mockAlertRepo, mockIDGen, sqlValidator, logger)

	ctx := context.Background()
	snowflakeID := int64(123456789)

	req := &report.UpdateAlertRequest{
		Name:        "更新后的预警",
		SQL:         "UPDATE products SET quantity = 0", // Non-read-only SQL
		Condition:   "> 50",
		Recipients:  "[1001]",
		PushChannel: "wechat",
		Status:      "active",
	}

	// Act
	result, err := svc.UpdateAlert(ctx, snowflakeID, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "SQL must be read-only (SELECT statement only)", err.Error())
	assert.Nil(t, result)
}

// TestAlertService_DeleteAlert_Success tests successful deletion.
func TestAlertService_DeleteAlert_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAlertRepo := mocks.NewMockAlertRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()
	sqlValidator := agent.NewSQLValidator()

	svc := NewAlertService(mockAlertRepo, mockIDGen, sqlValidator, logger)

	ctx := context.Background()
	snowflakeID := int64(123456789)

	// Mock expectations
	mockAlertRepo.EXPECT().Delete(ctx, snowflakeID).Return(nil)

	// Act
	err := svc.DeleteAlert(ctx, snowflakeID)

	// Assert
	assert.NoError(t, err)
}

// TestAlertService_DeleteAlert_RepositoryError tests deletion with repository error.
func TestAlertService_DeleteAlert_RepositoryError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAlertRepo := mocks.NewMockAlertRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()
	sqlValidator := agent.NewSQLValidator()

	svc := NewAlertService(mockAlertRepo, mockIDGen, sqlValidator, logger)

	ctx := context.Background()
	snowflakeID := int64(123456789)

	// Mock expectations
	mockAlertRepo.EXPECT().Delete(ctx, snowflakeID).Return(errors.New("database error"))

	// Act
	err := svc.DeleteAlert(ctx, snowflakeID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete alert")
}

// TestAlertService_GetActiveAlerts_Success tests retrieval of active alerts.
func TestAlertService_GetActiveAlerts_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAlertRepo := mocks.NewMockAlertRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()
	sqlValidator := agent.NewSQLValidator()

	svc := NewAlertService(mockAlertRepo, mockIDGen, sqlValidator, logger)

	ctx := context.Background()

	now := time.Now()
	alerts := []report.Alert{
		{
			SnowflakeID: 123456789,
			Name:        "库存预警",
			Status:      report.AlertStatusActive,
			CreatedBy:   1001,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	// Mock expectations
	mockAlertRepo.EXPECT().GetActiveAlerts(ctx).Return(alerts, nil)

	// Act
	result, err := svc.GetActiveAlerts(ctx)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, alerts[0].SnowflakeID, result[0].SnowflakeID)
}
