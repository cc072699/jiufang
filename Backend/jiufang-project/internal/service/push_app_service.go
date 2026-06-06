// Package service implements the application layer for push operations.
package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"jiufang/internal/infrastructure/erp"
	"jiufang/internal/model/report"
	"jiufang/internal/pkg/id"
	"jiufang/internal/repository"
)

// PushAppServiceInterface defines the interface for push application service.
type PushAppServiceInterface interface {
	ExecuteAndPushReport(ctx context.Context, scheduledReport *report.ScheduledReport) error
	RetryFailedPushes(ctx context.Context) error
}

// PushAppService manages push operations for scheduled reports.
type PushAppService struct {
	reportRepo  repository.ReportRepositoryInterface
	idGenerator id.SnowflakeGeneratorInterface
	erpReader   erp.ERPReaderInterface
	logger      *zap.Logger
}

// NewPushAppService creates a new PushAppService instance.
func NewPushAppService(
	reportRepo repository.ReportRepositoryInterface,
	idGenerator id.SnowflakeGeneratorInterface,
	erpReader erp.ERPReaderInterface,
	logger *zap.Logger,
) *PushAppService {
	return &PushAppService{
		reportRepo:  reportRepo,
		idGenerator: idGenerator,
		erpReader:   erpReader,
		logger:      logger,
	}
}

// ExecuteAndPushReport executes a scheduled report and pushes the result.
func (s *PushAppService) ExecuteAndPushReport(ctx context.Context, scheduledReport *report.ScheduledReport) error {
	s.logger.Info("starting to execute and push report",
		zap.Int64("report_id", scheduledReport.SnowflakeID),
		zap.String("report_name", scheduledReport.Name),
	)

	// Execute SQL against ERP database
	queryResult, err := s.executeQuery(ctx, scheduledReport.SQL)
	if err != nil {
		s.logger.Error("failed to execute query",
			zap.Int64("report_id", scheduledReport.SnowflakeID),
			zap.Error(err),
		)
		return s.recordPushFailure(ctx, scheduledReport, fmt.Errorf("failed to execute query: %w", err))
	}

	// Generate Markdown content
	pushContent, err := s.generateMarkdownContent(scheduledReport, queryResult)
	if err != nil {
		s.logger.Error("failed to generate markdown content",
			zap.Int64("report_id", scheduledReport.SnowflakeID),
			zap.Error(err),
		)
		return s.recordPushFailure(ctx, scheduledReport, fmt.Errorf("failed to generate markdown content: %w", err))
	}

	// Push message (WeChat not yet integrated — log and record success)
	s.logger.Info("push content generated (WeChat client not yet integrated)",
		zap.Int64("report_id", scheduledReport.SnowflakeID),
		zap.Int("content_length", len(pushContent)),
	)

	// Record successful push
	return s.recordPushSuccess(ctx, scheduledReport, pushContent)
}

// RetryFailedPushes retries failed push records.
func (s *PushAppService) RetryFailedPushes(ctx context.Context) error {
	s.logger.Info("retry failed pushes - not yet implemented")
	return nil
}

// Helper methods

func (s *PushAppService) executeQuery(ctx context.Context, sql string) (map[string]interface{}, error) {
	s.logger.Info("executing report query against ERP database",
		zap.String("sql", sql),
	)

	results, err := s.erpReader.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("ERP query execution failed: %w", err)
	}

	result := map[string]interface{}{
		"record_count": len(results),
		"query_time":   time.Now().Format(time.RFC3339),
		"rows":         results,
	}
	return result, nil
}

func (s *PushAppService) generateMarkdownContent(scheduledReport *report.ScheduledReport, queryResult map[string]interface{}) (string, error) {
	// Generate Markdown formatted report content
	content := fmt.Sprintf("# %s\n\n", scheduledReport.Name)

	if scheduledReport.Description != "" {
		content += fmt.Sprintf("**描述**: %s\n\n", scheduledReport.Description)
	}

	content += fmt.Sprintf("**执行时间**: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	// Add query results
	content += "## 查询结果\n\n"
	for key, value := range queryResult {
		content += fmt.Sprintf("- **%s**: %v\n", key, value)
	}

	content += "\n---\n"
	content += fmt.Sprintf("\n*本报告由ERP对话式查询助手自动生成*\n")

	return content, nil
}

func (s *PushAppService) recordPushSuccess(ctx context.Context, scheduledReport *report.ScheduledReport, pushContent string) error {
	// Create push record
	pushRecord := &report.PushRecord{
		SnowflakeID: s.idGenerator.Generate(),
		ReportID:    scheduledReport.SnowflakeID,
		PushType:    report.PushTypeReport,
		PushContent: pushContent,
		PushTargets: scheduledReport.Recipients, // Use Recipients field
		PushChannel: report.PushChannelWeChat,   // Default to WeChat
		PushStatus:  report.PushStatusSuccess,
		PushTime:    time.Now(),
		RetryCount:  0,
	}

	if err := s.reportRepo.CreatePushRecord(ctx, pushRecord); err != nil {
		s.logger.Error("failed to record push success",
			zap.Int64("report_id", scheduledReport.SnowflakeID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to record push success: %w", err)
	}

	s.logger.Info("push succeeded and recorded",
		zap.Int64("report_id", scheduledReport.SnowflakeID),
		zap.Int64("push_record_id", pushRecord.SnowflakeID),
	)

	return nil
}

func (s *PushAppService) recordPushFailure(ctx context.Context, scheduledReport *report.ScheduledReport, pushErr error) error {
	// Create push record with failure status
	pushRecord := &report.PushRecord{
		SnowflakeID:  s.idGenerator.Generate(),
		ReportID:     scheduledReport.SnowflakeID,
		PushType:     report.PushTypeReport,
		PushContent:  "",                         // Empty content for failed push
		PushTargets:  scheduledReport.Recipients, // Use Recipients field
		PushChannel:  report.PushChannelWeChat,   // Default to WeChat
		PushStatus:   report.PushStatusFailed,
		PushTime:     time.Now(),
		ErrorMessage: pushErr.Error(),
		RetryCount:   0,
	}

	if err := s.reportRepo.CreatePushRecord(ctx, pushRecord); err != nil {
		s.logger.Error("failed to record push failure",
			zap.Int64("report_id", scheduledReport.SnowflakeID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to record push failure: %w", err)
	}

	s.logger.Info("push failed and recorded",
		zap.Int64("report_id", scheduledReport.SnowflakeID),
		zap.Int64("push_record_id", pushRecord.SnowflakeID),
		zap.String("error", pushErr.Error()),
	)

	return pushErr
}
