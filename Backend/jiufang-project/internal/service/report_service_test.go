// Package service_test implements unit tests for ReportService.
// Author: AI Assistant
// Date: 2026-06-04
// Description: Unit tests for scheduled report business service

package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"jiufang/internal/mocks"
	"jiufang/internal/model/report"
	"jiufang/internal/service"
)

// TestReportService_CreateReport_ValidInput_ShouldSucceed tests TC-RS-01
func TestReportService_CreateReport_ValidInput_ShouldSucceed(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReportRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()

	svc := service.NewReportService(mockRepo, mockIDGen, logger)

	ctx := context.Background()
	expectedSnowflakeID := int64(123456789)
	req := &report.CreateReportRequest{
		Name:         "Test Report",
		Description:  "Test Description",
		SQL:          "SELECT SUM(amount) as total_amount FROM sales WHERE date = CURRENT_DATE",
		ScheduleType: report.ScheduleTypeDaily,
		ScheduleTime: "09:00:00",
		Recipients:   []string{"user1@example.com", "user2@example.com"},
		CreatedBy:    1001,
	}

	// Mock expectations
	mockIDGen.EXPECT().Generate().Return(expectedSnowflakeID)
	mockRepo.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, r *report.ScheduledReport) error {
			assert.Equal(t, expectedSnowflakeID, r.SnowflakeID)
			assert.Equal(t, req.Name, r.Name)
			assert.Equal(t, req.Description, r.Description)
			assert.Equal(t, req.SQL, r.SQL)
			assert.Equal(t, req.ScheduleType, r.ScheduleType)
			assert.Equal(t, req.ScheduleTime, r.ScheduleTime)
			// Recipients is stored as JSON string
			expectedRecipientsJSON, _ := json.Marshal(req.Recipients)
			assert.Equal(t, string(expectedRecipientsJSON), r.Recipients)
			assert.Equal(t, report.ReportStatusActive, r.Status)
			assert.Equal(t, req.CreatedBy, r.CreatedBy)
			return nil
		},
	)

	// Act
	result, err := svc.CreateReport(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedSnowflakeID, result.SnowflakeID)
	assert.Equal(t, req.Name, result.Name)
}

// TestReportService_CreateReport_NameTooShort_ShouldReturnError tests TC-RS-02
func TestReportService_CreateReport_NameTooShort_ShouldReturnError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReportRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()

	svc := service.NewReportService(mockRepo, mockIDGen, logger)

	ctx := context.Background()
	req := &report.CreateReportRequest{
		Name:         "", // Empty name to trigger validation error
		Description:  "Test Description",
		SQL:          "SELECT SUM(amount) as total_amount FROM sales WHERE date = CURRENT_DATE",
		ScheduleType: report.ScheduleTypeDaily,
		ScheduleTime: "09:00:00",
		Recipients:   []string{"user1@example.com"},
		CreatedBy:    1001,
	}

	// Act
	result, err := svc.CreateReport(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name must be between 1 and 100 characters")
	assert.Nil(t, result)
}

// TestReportService_CreateReport_InvalidCron_ShouldReturnError tests TC-RS-03
func TestReportService_CreateReport_InvalidCron_ShouldReturnError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReportRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()

	svc := service.NewReportService(mockRepo, mockIDGen, logger)

	ctx := context.Background()
	req := &report.CreateReportRequest{
		Name:         "Test Report",
		Description:  "Test Description",
		SQL:          "SELECT SUM(amount) as total_amount FROM sales WHERE date = CURRENT_DATE",
		ScheduleType: report.ScheduleTypeDaily,
		ScheduleTime: "", // Empty time
		Recipients:   []string{"user1@example.com"},
		CreatedBy:    1001,
	}

	// Act
	result, err := svc.CreateReport(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "schedule_time is required")
}

// TestReportService_CreateReport_InvalidPushTargets_ShouldReturnError tests TC-RS-04
func TestReportService_CreateReport_InvalidPushTargets_ShouldReturnError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReportRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()

	svc := service.NewReportService(mockRepo, mockIDGen, logger)

	ctx := context.Background()
	req := &report.CreateReportRequest{
		Name:         "Test Report",
		Description:  "Test Description",
		SQL:          "SELECT SUM(amount) as total_amount FROM sales WHERE date = CURRENT_DATE",
		ScheduleType: report.ScheduleTypeDaily,
		ScheduleTime: "09:00:00",
		Recipients:   []string{}, // Empty array
		CreatedBy:    1001,
	}

	// Act
	result, err := svc.CreateReport(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "recipients is required")
}

// TestReportService_GetReportByID_ValidID_ShouldSucceed tests TC-RS-05
func TestReportService_GetReportByID_ValidID_ShouldSucceed(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReportRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()

	svc := service.NewReportService(mockRepo, mockIDGen, logger)

	ctx := context.Background()
	snowflakeID := int64(123456789)
	expectedReport := &report.ScheduledReport{
		SnowflakeID:  snowflakeID,
		Name:         "Test Report",
		Description:  "Test Description",
		SQL:          "SELECT SUM(amount) as total_amount FROM sales WHERE date = CURRENT_DATE",
		ScheduleType: report.ScheduleTypeDaily,
		ScheduleTime: "09:00:00",
		Recipients:   "[\"user1@example.com\"]", // JSON数组格式
		Status:       report.ReportStatusActive,
		CreatedBy:    1001,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Mock expectations
	mockRepo.EXPECT().GetBySnowflakeID(ctx, snowflakeID).Return(expectedReport, nil)

	// Act
	result, err := svc.GetReportByID(ctx, snowflakeID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, snowflakeID, result.SnowflakeID)
	assert.Equal(t, "Test Report", result.Name)
}

// TestReportService_GetReportByID_NotFound_ShouldReturnError tests TC-RS-06
func TestReportService_GetReportByID_NotFound_ShouldReturnError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReportRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()

	svc := service.NewReportService(mockRepo, mockIDGen, logger)

	ctx := context.Background()
	snowflakeID := int64(999999)

	// Mock expectations
	mockRepo.EXPECT().GetBySnowflakeID(ctx, snowflakeID).Return(nil, errors.New("scheduled report not found"))

	// Act
	result, err := svc.GetReportByID(ctx, snowflakeID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get scheduled report")
}

// TestReportService_ListReports_ValidRequest_ShouldSucceed tests TC-RS-07
func TestReportService_ListReports_ValidRequest_ShouldSucceed(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReportRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()

	svc := service.NewReportService(mockRepo, mockIDGen, logger)

	ctx := context.Background()
	req := &report.ListReportsRequest{
		Page:     1,
		PageSize: 10,
		Status:   "active",
	}

	expectedReports := []report.ScheduledReport{
		{
			SnowflakeID:  123456789,
			Name:         "Report 1",
			Description:  "Description 1",
			SQL:          "SELECT SUM(amount) as total_amount FROM sales WHERE date = CURRENT_DATE",
			ScheduleType: report.ScheduleTypeDaily,
			ScheduleTime: "09:00:00",
			Recipients:   "[\"user1@example.com\"]", // JSON数组格式
			Status:       report.ReportStatusActive,
			CreatedBy:    1001,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}
	expectedTotal := int64(1)

	// Mock expectations
	mockRepo.EXPECT().List(ctx, 0, 10, "active").Return(expectedReports, expectedTotal, nil)

	// Act
	results, total, err := svc.ListReports(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, expectedTotal, total)
	assert.Equal(t, "Report 1", results[0].Name)
}

// TestReportService_UpdateReport_ValidRequest_ShouldSucceed tests TC-RS-08
func TestReportService_UpdateReport_ValidRequest_ShouldSucceed(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReportRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()

	svc := service.NewReportService(mockRepo, mockIDGen, logger)

	ctx := context.Background()
	snowflakeID := int64(123456789)
	req := &report.UpdateReportRequest{
		Name:         "Updated Report",
		Description:  "Updated Description",
		SQL:          "SELECT SUM(amount) as total_amount FROM sales WHERE date = CURRENT_DATE",
		ScheduleType: report.ScheduleTypeDaily,
		ScheduleTime: "10:00:00",
		Recipients:   []string{"user1@example.com", "user2@example.com"},
	}

	// Mock expectations
	mockRepo.EXPECT().Update(ctx, snowflakeID, gomock.Any()).DoAndReturn(
		func(ctx context.Context, id int64, updates map[string]interface{}) error {
			assert.Equal(t, "Updated Report", updates["name"])
			assert.Equal(t, "Updated Description", updates["description"])
			assert.NotNil(t, updates["updated_at"])
			return nil
		},
	)

	// Act
	err := svc.UpdateReport(ctx, snowflakeID, req)

	// Assert
	assert.NoError(t, err)
}

// TestReportService_DeleteReport_ValidID_ShouldSucceed tests TC-RS-09
func TestReportService_DeleteReport_ValidID_ShouldSucceed(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReportRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()

	svc := service.NewReportService(mockRepo, mockIDGen, logger)

	ctx := context.Background()
	snowflakeID := int64(123456789)

	// Mock expectations
	mockRepo.EXPECT().Delete(ctx, snowflakeID).Return(nil)

	// Act
	err := svc.DeleteReport(ctx, snowflakeID)

	// Assert
	assert.NoError(t, err)
}

// TestReportService_CreateReport_RepositoryError_ShouldReturnError tests repository error handling
func TestReportService_CreateReport_RepositoryError_ShouldReturnError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReportRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()

	svc := service.NewReportService(mockRepo, mockIDGen, logger)

	ctx := context.Background()
	expectedSnowflakeID := int64(123456789)
	req := &report.CreateReportRequest{
		Name:         "Test Report",
		Description:  "Test Description",
		SQL:          "SELECT SUM(amount) as total_amount FROM sales WHERE date = CURRENT_DATE",
		ScheduleType: report.ScheduleTypeDaily,
		ScheduleTime: "09:00:00",
		Recipients:   []string{"user1@example.com"},
		CreatedBy:    1001,
	}

	// Mock expectations
	mockIDGen.EXPECT().Generate().Return(expectedSnowflakeID)
	mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(errors.New("database error"))

	// Act
	result, err := svc.CreateReport(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create scheduled report")
}

// TestReportService_CreateReport_EmptyPushTargets_ShouldReturnError tests empty push targets
func TestReportService_CreateReport_EmptyPushTargets_ShouldReturnError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReportRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()

	svc := service.NewReportService(mockRepo, mockIDGen, logger)

	ctx := context.Background()
	req := &report.CreateReportRequest{
		Name:         "Test Report",
		Description:  "Test Description",
		SQL:          "SELECT SUM(amount) as total_amount FROM sales WHERE date = CURRENT_DATE",
		ScheduleType: report.ScheduleTypeDaily,
		ScheduleTime: "09:00:00",
		Recipients:   []string{}, // Empty array
		CreatedBy:    1001,
	}

	// Act
	result, err := svc.CreateReport(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "recipients is required")
}

// TestReportService_CreateReport_InvalidQueryCondition_ShouldReturnError tests invalid query condition JSON
func TestReportService_CreateReport_InvalidQueryCondition_ShouldReturnError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReportRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()

	svc := service.NewReportService(mockRepo, mockIDGen, logger)

	ctx := context.Background()
	req := &report.CreateReportRequest{
		Name:         "Test Report",
		Description:  "Test Description",
		SQL:          "", // Empty SQL
		ScheduleType: report.ScheduleTypeDaily,
		ScheduleTime: "09:00:00",
		Recipients:   []string{"user1@example.com"},
		CreatedBy:    1001,
	}

	// Act
	result, err := svc.CreateReport(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "sql is required")
}

// TestReportService_ListReports_DefaultPagination_ShouldUseDefaults tests default pagination values
func TestReportService_ListReports_DefaultPagination_ShouldUseDefaults(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReportRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()

	svc := service.NewReportService(mockRepo, mockIDGen, logger)

	ctx := context.Background()
	req := &report.ListReportsRequest{
		Page:     0, // Invalid, should default to 1
		PageSize: 0, // Invalid, should default to 10
	}

	expectedReports := []report.ScheduledReport{}
	expectedTotal := int64(0)

	// Mock expectations - should be called with default values (offset=0, limit=10)
	mockRepo.EXPECT().List(ctx, 0, 10, "").Return(expectedReports, expectedTotal, nil)

	// Act
	results, total, err := svc.ListReports(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, expectedTotal, total)
	// Verify that defaults were applied
	assert.Equal(t, 1, req.Page)
	assert.Equal(t, 10, req.PageSize)
}

// TestReportService_UpdateReport_InvalidStatus_ShouldReturnError tests invalid status in update
func TestReportService_UpdateReport_InvalidStatus_ShouldReturnError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReportRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()

	svc := service.NewReportService(mockRepo, mockIDGen, logger)

	ctx := context.Background()
	snowflakeID := int64(123456789)
	req := &report.UpdateReportRequest{
		Name:         "Updated Report",
		Description:  "Updated Description",
		SQL:          "SELECT SUM(amount) as total_amount FROM sales WHERE date = CURRENT_DATE",
		ScheduleType: "invalid_type", // Invalid schedule type
		ScheduleTime: "10:00:00",
		Recipients:   []string{"user1@example.com"},
	}

	// Act
	err := svc.UpdateReport(ctx, snowflakeID, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid schedule type")
}

// TestReportService_DeleteReport_RepositoryError_ShouldReturnError tests repository error in delete
func TestReportService_DeleteReport_RepositoryError_ShouldReturnError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReportRepository(ctrl)
	mockIDGen := mocks.NewMockSnowflakeGenerator(ctrl)
	logger := zap.NewNop()

	svc := service.NewReportService(mockRepo, mockIDGen, logger)

	ctx := context.Background()
	snowflakeID := int64(123456789)

	// Mock expectations
	mockRepo.EXPECT().Delete(ctx, snowflakeID).Return(errors.New("database error"))

	// Act
	err := svc.DeleteReport(ctx, snowflakeID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete scheduled report")
}
