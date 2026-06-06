// Package service implements the alert service.
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"jiufang/internal/agent"
	"jiufang/internal/model/report"
	"jiufang/internal/pkg/id"
	"jiufang/internal/repository"

	"go.uber.org/zap"
)

// AlertService implements the business logic for alert rules.
type AlertService struct {
	alertRepo    repository.AlertRepositoryInterface
	idGenerator  id.SnowflakeGeneratorInterface
	sqlValidator *agent.SQLValidator
	logger       *zap.Logger
}

// NewAlertService creates a new AlertService instance.
func NewAlertService(
	alertRepo repository.AlertRepositoryInterface,
	idGenerator id.SnowflakeGeneratorInterface,
	sqlValidator *agent.SQLValidator,
	logger *zap.Logger,
) *AlertService {
	return &AlertService{
		alertRepo:    alertRepo,
		idGenerator:  idGenerator,
		sqlValidator: sqlValidator,
		logger:       logger,
	}
}

// CreateAlert creates a new alert rule.
func (s *AlertService) CreateAlert(ctx context.Context, req *report.CreateAlertRequest) (*report.Alert, error) {
	// Validate SQL is read-only
	if !s.sqlValidator.IsReadOnly(req.SQL) {
		return nil, errors.New("SQL must be read-only (SELECT statement only)")
	}

	// Validate SQL syntax
	if err := s.sqlValidator.Validate(req.SQL); err != nil {
		return nil, fmt.Errorf("SQL validation failed: %w", err)
	}

	// Generate snowflake ID
	snowflakeID := s.idGenerator.Generate()

	// Parse silence time if provided
	var silenceStart, silenceEnd *time.Time
	if req.SilenceStart != "" {
		t, err := parseTime(req.SilenceStart)
		if err != nil {
			return nil, fmt.Errorf("invalid silence_start format: %w", err)
		}
		silenceStart = &t
	}

	if req.SilenceEnd != "" {
		t, err := parseTime(req.SilenceEnd)
		if err != nil {
			return nil, fmt.Errorf("invalid silence_end format: %w", err)
		}
		silenceEnd = &t
	}

	// Serialize recipients to JSON string for storage
	recipientsJSON, err := json.Marshal(req.Recipients)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize recipients: %w", err)
	}

	// Create alert model
	alert := &report.Alert{
		SnowflakeID:      snowflakeID,
		Name:             req.Name,
		Description:      req.Description,
		SQL:              req.SQL,
		Condition:        req.Condition,
		Recipients:       string(recipientsJSON),
		PushChannel:      report.PushChannel(req.PushChannel),
		TriggerFrequency: report.TriggerFrequency(req.TriggerFrequency),
		SilenceStart:     silenceStart,
		SilenceEnd:       silenceEnd,
		Status:           report.AlertStatusActive,
		CreatedBy:        req.CreatedBy,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Save to repository
	if err := s.alertRepo.Create(ctx, alert); err != nil {
		s.logger.Error("failed to create alert",
			zap.Int64("snowflake_id", snowflakeID),
			zap.String("name", req.Name),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create alert: %w", err)
	}

	s.logger.Info("alert created successfully",
		zap.Int64("snowflake_id", snowflakeID),
		zap.String("name", req.Name),
	)

	return alert, nil
}

// GetAlertByID retrieves an alert rule by its snowflake ID.
func (s *AlertService) GetAlertByID(ctx context.Context, snowflakeID int64) (*report.Alert, error) {
	alert, err := s.alertRepo.GetBySnowflakeID(ctx, snowflakeID)
	if err != nil {
		s.logger.Error("failed to get alert by snowflake ID",
			zap.Int64("snowflake_id", snowflakeID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	if alert == nil {
		return nil, errors.New("alert not found")
	}

	return alert, nil
}

// ListAlerts retrieves a list of alert rules with pagination and filtering.
func (s *AlertService) ListAlerts(ctx context.Context, page, pageSize int, name string, status string) ([]report.Alert, int64, error) {
	// Calculate offset
	offset := (page - 1) * pageSize

	alerts, total, err := s.alertRepo.List(ctx, offset, pageSize, name, status)
	if err != nil {
		s.logger.Error("failed to list alerts",
			zap.Int("page", page),
			zap.Int("page_size", pageSize),
			zap.String("name", name),
			zap.String("status", status),
			zap.Error(err),
		)
		return nil, 0, fmt.Errorf("failed to list alerts: %w", err)
	}

	return alerts, total, nil
}

// UpdateAlert updates an existing alert rule.
func (s *AlertService) UpdateAlert(ctx context.Context, snowflakeID int64, req *report.UpdateAlertRequest) (*report.Alert, error) {
	// Validate SQL is read-only
	if !s.sqlValidator.IsReadOnly(req.SQL) {
		return nil, errors.New("SQL must be read-only (SELECT statement only)")
	}

	// Validate SQL syntax
	if err := s.sqlValidator.Validate(req.SQL); err != nil {
		return nil, fmt.Errorf("SQL validation failed: %w", err)
	}

	// Parse silence time if provided
	var silenceStart, silenceEnd *time.Time
	if req.SilenceStart != "" {
		t, err := parseTime(req.SilenceStart)
		if err != nil {
			return nil, fmt.Errorf("invalid silence_start format: %w", err)
		}
		silenceStart = &t
	}

	if req.SilenceEnd != "" {
		t, err := parseTime(req.SilenceEnd)
		if err != nil {
			return nil, fmt.Errorf("invalid silence_end format: %w", err)
		}
		silenceEnd = &t
	}

	// Serialize recipients to JSON string for storage
	recipientsJSON, err := json.Marshal(req.Recipients)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize recipients: %w", err)
	}

	// Prepare updates
	updates := map[string]interface{}{
		"name":              req.Name,
		"description":       req.Description,
		"sql":               req.SQL,
		"condition":         req.Condition,
		"recipients":        string(recipientsJSON),
		"push_channel":      req.PushChannel,
		"trigger_frequency": req.TriggerFrequency,
		"silence_start":     silenceStart,
		"silence_end":       silenceEnd,
		"status":            req.Status,
		"updated_at":        time.Now(),
	}

	// Update in repository
	if err := s.alertRepo.Update(ctx, snowflakeID, updates); err != nil {
		s.logger.Error("failed to update alert",
			zap.Int64("snowflake_id", snowflakeID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to update alert: %w", err)
	}

	// Retrieve updated alert
	alert, err := s.alertRepo.GetBySnowflakeID(ctx, snowflakeID)
	if err != nil {
		s.logger.Error("failed to get updated alert",
			zap.Int64("snowflake_id", snowflakeID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get updated alert: %w", err)
	}

	s.logger.Info("alert updated successfully",
		zap.Int64("snowflake_id", snowflakeID),
		zap.String("name", req.Name),
	)

	return alert, nil
}

// DeleteAlert deletes an alert rule by its snowflake ID.
func (s *AlertService) DeleteAlert(ctx context.Context, snowflakeID int64) error {
	if err := s.alertRepo.Delete(ctx, snowflakeID); err != nil {
		s.logger.Error("failed to delete alert",
			zap.Int64("snowflake_id", snowflakeID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete alert: %w", err)
	}

	s.logger.Info("alert deleted successfully",
		zap.Int64("snowflake_id", snowflakeID),
	)

	return nil
}

// GetActiveAlerts retrieves all active alert rules.
func (s *AlertService) GetActiveAlerts(ctx context.Context) ([]report.Alert, error) {
	alerts, err := s.alertRepo.GetActiveAlerts(ctx)
	if err != nil {
		s.logger.Error("failed to get active alerts",
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get active alerts: %w", err)
	}

	return alerts, nil
}

// parseTime parses a time string in "HH:MM" format.
func parseTime(timeStr string) (time.Time, error) {
	return time.Parse("15:04", timeStr)
}
