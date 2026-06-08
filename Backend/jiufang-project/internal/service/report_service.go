// Package service implements the application layer for scheduled report management.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"jiufang/internal/model/report"
	"jiufang/internal/pkg/id"
	"jiufang/internal/repository"
)

// ReportServiceInterface defines the interface for report business service.
type ReportServiceInterface interface {
	CreateReport(ctx context.Context, req *report.CreateReportRequest) (*report.ScheduledReport, error)
	GetReportByID(ctx context.Context, snowflakeID int64) (*report.ScheduledReport, error)
	ListReports(ctx context.Context, req *report.ListReportsRequest) ([]report.ScheduledReport, int64, error)
	UpdateReport(ctx context.Context, snowflakeID int64, req *report.UpdateReportRequest) error
	DeleteReport(ctx context.Context, snowflakeID int64) error
	GetPushRecords(ctx context.Context, req *report.GetPushRecordsRequest) ([]report.PushRecord, int64, error)
}

// ReportService manages scheduled report business logic.
type ReportService struct {
	reportRepo  repository.ReportRepositoryInterface
	idGenerator id.SnowflakeGeneratorInterface
	logger      *zap.Logger
}

// NewReportService creates a new ReportService instance.
func NewReportService(
	reportRepo repository.ReportRepositoryInterface,
	idGenerator id.SnowflakeGeneratorInterface,
	logger *zap.Logger,
) *ReportService {
	return &ReportService{
		reportRepo:  reportRepo,
		idGenerator: idGenerator,
		logger:      logger,
	}
}

// CreateReport creates a new scheduled report.
func (s *ReportService) CreateReport(ctx context.Context, req *report.CreateReportRequest) (*report.ScheduledReport, error) {
	// Validate request
	if err := validateCreateReportRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Validate schedule type
	if err := validateScheduleType(req.ScheduleType); err != nil {
		return nil, fmt.Errorf("invalid schedule type: %w", err)
	}

	// Validate schedule time format
	if err := validateScheduleTime(req.ScheduleTime); err != nil {
		return nil, fmt.Errorf("invalid schedule time: %w", err)
	}

	// Validate recipients JSON format
	if err := validateRecipients(req.Recipients); err != nil {
		return nil, fmt.Errorf("invalid recipients: %w", err)
	}

	// Generate snowflake ID
	snowflakeID := s.idGenerator.Generate()

	// Convert recipients to JSON string
	recipientsJSON, err := json.Marshal(req.Recipients)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal recipients: %w", err)
	}

	// Create scheduled report entity
	pushChannel := req.PushChannel
	if pushChannel == "" {
		pushChannel = report.PushChannelWeChat
	}
	scheduledReport := &report.ScheduledReport{
		SnowflakeID:  snowflakeID,
		Name:         req.Name,
		Description:  req.Description,
		SQL:          req.SQL,
		ScheduleType: req.ScheduleType,
		ScheduleTime: req.ScheduleTime,
		Recipients:   string(recipientsJSON),
		PushChannel:  pushChannel,
		Status:       report.ReportStatusActive,
		CreatedBy:    req.CreatedBy,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save to database
	if err := s.reportRepo.Create(ctx, scheduledReport); err != nil {
		s.logger.Error("failed to create scheduled report",
			zap.Int64("snowflake_id", snowflakeID),
			zap.String("name", req.Name),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create scheduled report: %w", err)
	}

	s.logger.Info("scheduled report created successfully",
		zap.Int64("snowflake_id", snowflakeID),
		zap.String("name", req.Name),
		zap.Int64("created_by", req.CreatedBy),
	)

	return scheduledReport, nil
}

// GetReportByID retrieves a scheduled report by its snowflake ID.
func (s *ReportService) GetReportByID(ctx context.Context, snowflakeID int64) (*report.ScheduledReport, error) {
	scheduledReport, err := s.reportRepo.GetBySnowflakeID(ctx, snowflakeID)
	if err != nil {
		s.logger.Error("failed to get scheduled report",
			zap.Int64("snowflake_id", snowflakeID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get scheduled report: %w", err)
	}

	return scheduledReport, nil
}

// ListReports retrieves scheduled reports with pagination and filters.
func (s *ReportService) ListReports(ctx context.Context, req *report.ListReportsRequest) ([]report.ScheduledReport, int64, error) {
	// Validate request
	if err := validateListReportsRequest(req); err != nil {
		return nil, 0, fmt.Errorf("invalid request: %w", err)
	}

	// Calculate offset
	offset := (req.Page - 1) * req.PageSize

	// Get reports from repository
	scheduledReports, total, err := s.reportRepo.List(ctx, offset, req.PageSize, req.Name, req.Status)
	if err != nil {
		s.logger.Error("failed to list scheduled reports",
			zap.Int("page", req.Page),
			zap.Int("page_size", req.PageSize),
			zap.Error(err),
		)
		return nil, 0, fmt.Errorf("failed to list scheduled reports: %w", err)
	}

	return scheduledReports, total, nil
}

// UpdateReport updates a scheduled report by its snowflake ID.
func (s *ReportService) UpdateReport(ctx context.Context, snowflakeID int64, req *report.UpdateReportRequest) error {
	// Validate request
	if err := validateUpdateReportRequest(req); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	// Validate schedule type if provided
	if req.ScheduleType != "" {
		if err := validateScheduleType(req.ScheduleType); err != nil {
			return fmt.Errorf("invalid schedule type: %w", err)
		}
	}

	// Validate schedule time format if provided
	if req.ScheduleTime != "" {
		if err := validateScheduleTime(req.ScheduleTime); err != nil {
			return fmt.Errorf("invalid schedule time: %w", err)
		}
	}

	// Validate recipients JSON format if provided
	if len(req.Recipients) > 0 {
		if err := validateRecipients(req.Recipients); err != nil {
			return fmt.Errorf("invalid recipients: %w", err)
		}
	}

	// Convert recipients to JSON string if provided
	var recipientsJSON string
	if len(req.Recipients) > 0 {
		recipientsBytes, err := json.Marshal(req.Recipients)
		if err != nil {
			return fmt.Errorf("failed to marshal recipients: %w", err)
		}
		recipientsJSON = string(recipientsBytes)
	}

	// Prepare updates
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.SQL != "" {
		updates["sql"] = req.SQL
	}
	if req.ScheduleType != "" {
		updates["schedule_type"] = req.ScheduleType
	}
	if req.ScheduleTime != "" {
		updates["schedule_time"] = req.ScheduleTime
	}
	if recipientsJSON != "" {
		updates["recipients"] = recipientsJSON
	}
	if req.PushChannel != "" {
		updates["push_channel"] = req.PushChannel
	}

	// Update in database
	if err := s.reportRepo.Update(ctx, snowflakeID, updates); err != nil {
		s.logger.Error("failed to update scheduled report",
			zap.Int64("snowflake_id", snowflakeID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update scheduled report: %w", err)
	}

	s.logger.Info("scheduled report updated successfully",
		zap.Int64("snowflake_id", snowflakeID),
		zap.String("name", req.Name),
	)

	return nil
}

// DeleteReport deletes a scheduled report by its snowflake ID.
func (s *ReportService) DeleteReport(ctx context.Context, snowflakeID int64) error {
	// Delete from database
	if err := s.reportRepo.Delete(ctx, snowflakeID); err != nil {
		s.logger.Error("failed to delete scheduled report",
			zap.Int64("snowflake_id", snowflakeID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete scheduled report: %w", err)
	}

	s.logger.Info("scheduled report deleted successfully",
		zap.Int64("snowflake_id", snowflakeID),
	)

	return nil
}

// GetPushRecords retrieves push records with pagination and filters.
func (s *ReportService) GetPushRecords(ctx context.Context, req *report.GetPushRecordsRequest) ([]report.PushRecord, int64, error) {
	// Validate request
	if err := validateGetPushRecordsRequest(req); err != nil {
		return nil, 0, fmt.Errorf("invalid request: %w", err)
	}

	// Calculate offset
	offset := (req.Page - 1) * req.PageSize

	// Get all push records with filters
	pushRecords, total, err := s.reportRepo.ListPushRecords(ctx, offset, req.PageSize, string(req.PushType), string(req.Status), req.StartTime.Format("2006-01-02 15:04:05"), req.EndTime.Format("2006-01-02 15:04:05"))

	if err != nil {
		s.logger.Error("failed to get push records",
			zap.Int("page", req.Page),
			zap.Int("page_size", req.PageSize),
			zap.Error(err),
		)
		return nil, 0, fmt.Errorf("failed to get push records: %w", err)
	}

	return pushRecords, total, nil
}

// Validation functions

func validateCreateReportRequest(req *report.CreateReportRequest) error {
	if req.Name == "" || len(req.Name) < 1 || len(req.Name) > 100 {
		return fmt.Errorf("name must be between 1 and 100 characters")
	}
	if len(req.Description) > 200 {
		return fmt.Errorf("description must be less than 200 characters")
	}
	if req.SQL == "" {
		return fmt.Errorf("sql is required")
	}
	if req.ScheduleType == "" {
		return fmt.Errorf("schedule_type is required")
	}
	if req.ScheduleTime == "" {
		return fmt.Errorf("schedule_time is required")
	}
	if len(req.Recipients) == 0 {
		return fmt.Errorf("recipients is required")
	}
	if req.CreatedBy <= 0 {
		return fmt.Errorf("created_by must be positive")
	}
	return nil
}

func validateListReportsRequest(req *report.ListReportsRequest) error {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}
	if req.Status != "" && req.Status != "active" && req.Status != "inactive" {
		return fmt.Errorf("status must be active or inactive")
	}
	return nil
}

func validateUpdateReportRequest(req *report.UpdateReportRequest) error {
	if req.Name != "" && (len(req.Name) < 1 || len(req.Name) > 100) {
		return fmt.Errorf("name must be between 1 and 100 characters")
	}
	if len(req.Description) > 200 {
		return fmt.Errorf("description must be less than 200 characters")
	}
	if req.SQL != "" && len(req.SQL) < 1 {
		return fmt.Errorf("sql must not be empty")
	}
	if len(req.Recipients) > 0 {
		for _, recipient := range req.Recipients {
			if recipient == "" {
				return fmt.Errorf("recipient must not be empty")
			}
		}
	}
	return nil
}

func validateGetPushRecordsRequest(req *report.GetPushRecordsRequest) error {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}
	if req.PushType != "" && req.PushType != "report" && req.PushType != "alert" {
		return fmt.Errorf("push_type must be report or alert")
	}
	if req.Status != "" && req.Status != "success" && req.Status != "failed" && req.Status != "retrying" {
		return fmt.Errorf("status must be success, failed, or retrying")
	}
	return nil
}

func validateScheduleType(scheduleType report.ScheduleType) error {
	if scheduleType != report.ScheduleTypeDaily && scheduleType != report.ScheduleTypeWeekly && scheduleType != report.ScheduleTypeMonthly {
		return fmt.Errorf("schedule_type must be daily, weekly, or monthly")
	}
	return nil
}

func validateScheduleTime(scheduleTime string) error {
	// Basic validation: schedule time should not be empty
	if scheduleTime == "" {
		return fmt.Errorf("schedule_time is empty")
	}
	// Add more sophisticated validation if needed (e.g., ISO8601 format)
	return nil
}

func validateRecipients(recipients []string) error {
	// Validate recipients array
	if len(recipients) == 0 {
		return fmt.Errorf("recipients must contain at least one recipient")
	}
	for _, recipient := range recipients {
		if recipient == "" {
			return fmt.Errorf("recipient must not be empty")
		}
	}
	return nil
}
