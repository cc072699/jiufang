// Package service implements the operation log service.
package service

import (
	"context"
	"fmt"

	"jiufang/internal/model/audit"
	"jiufang/internal/pkg/id"
	"jiufang/internal/repository"

	"go.uber.org/zap"
)

// OperationLogService implements the business logic for operation logs.
type OperationLogService struct {
	logRepo  repository.OperationLogRepositoryInterface
	userRepo repository.UserRepositoryInterface
	logger   *zap.Logger
}

// NewOperationLogService creates a new OperationLogService instance.
func NewOperationLogService(
	logRepo repository.OperationLogRepositoryInterface,
	userRepo repository.UserRepositoryInterface,
	logger *zap.Logger,
) *OperationLogService {
	return &OperationLogService{
		logRepo:  logRepo,
		userRepo: userRepo,
		logger:   logger,
	}
}

// RecordOperation records an operation log entry.
func (s *OperationLogService) RecordOperation(ctx context.Context, userID int64, opType audit.OperationType, object string, detail string, result audit.OperationResult, ip string) {
	snowflakeID, err := id.Generate()
	if err != nil {
		s.logger.Warn("failed to generate snowflake ID for operation log", zap.Error(err))
		return
	}

	var uid *int64
	if userID > 0 {
		uid = &userID
	}

	log := &audit.OperationLog{
		SnowflakeID:     snowflakeID,
		UserID:          uid,
		OperationType:   opType,
		OperationObject: object,
		OperationDetail: detail,
		OperationResult: result,
		IPAddress:       ip,
	}

	if err := s.logRepo.Create(ctx, log); err != nil {
		s.logger.Warn("failed to record operation log", zap.Error(err))
	}
}

// ListOperationLogs retrieves a list of operation logs with pagination and filtering.
func (s *OperationLogService) ListOperationLogs(ctx context.Context, page, size int, userID int64, operationType string, startTime, endTime string) ([]audit.OperationLog, int64, error) {
	// Calculate offset
	offset := (page - 1) * size

	logs, total, err := s.logRepo.List(ctx, offset, size, userID, operationType, startTime, endTime)
	if err != nil {
		s.logger.Error("failed to list operation logs",
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Int64("user_id", userID),
			zap.String("operation_type", operationType),
			zap.Error(err),
		)
		return nil, 0, fmt.Errorf("failed to list operation logs: %w", err)
	}

	return logs, total, nil
}

// GetOperationLogByID retrieves an operation log by its snowflake ID.
func (s *OperationLogService) GetOperationLogByID(ctx context.Context, snowflakeID int64) (*audit.OperationLog, error) {
	log, err := s.logRepo.GetBySnowflakeID(ctx, snowflakeID)
	if err != nil {
		s.logger.Error("failed to get operation log by snowflake ID",
			zap.Int64("snowflake_id", snowflakeID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get operation log: %w", err)
	}

	if log == nil {
		return nil, fmt.Errorf("operation log not found")
	}

	return log, nil
}

// GetUsernameByUserID retrieves username by user ID.
func (s *OperationLogService) GetUsernameByUserID(ctx context.Context, userID int64) (string, error) {
	user, err := s.userRepo.GetBySnowflakeID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get user by ID",
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return "", nil
	}

	return user.Username, nil
}
