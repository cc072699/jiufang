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
	ExecuteAndPushAlert(ctx context.Context, alert *report.Alert, queryResult map[string]interface{}) error
	RetryFailedPushes(ctx context.Context) error
}

// PushAppService manages push operations for scheduled reports.
type PushAppService struct {
	reportRepo       repository.ReportRepositoryInterface
	idGenerator      id.SnowflakeGeneratorInterface
	erpReader        erp.ERPReaderInterface
	emailService     *EmailService
	recipientResolver *RecipientResolver
	logger           *zap.Logger
}

// NewPushAppService creates a new PushAppService instance.
func NewPushAppService(
	reportRepo repository.ReportRepositoryInterface,
	idGenerator id.SnowflakeGeneratorInterface,
	erpReader erp.ERPReaderInterface,
	emailService *EmailService,
	recipientResolver *RecipientResolver,
	logger *zap.Logger,
) *PushAppService {
	return &PushAppService{
		reportRepo:       reportRepo,
		idGenerator:      idGenerator,
		erpReader:        erpReader,
		emailService:     emailService,
		recipientResolver: recipientResolver,
		logger:           logger,
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

	// Dispatch based on push channel
	var sendErr error
	if scheduledReport.PushChannel == report.PushChannelEmail {
		emails, err := s.recipientResolver.ResolveEmails(ctx, scheduledReport.Recipients)
		if err != nil {
			s.logger.Error("failed to resolve recipient emails",
				zap.Int64("report_id", scheduledReport.SnowflakeID),
				zap.Error(err),
			)
			return s.recordPushFailure(ctx, scheduledReport, fmt.Errorf("failed to resolve recipient emails: %w", err))
		}
		if len(emails) > 0 {
			subject := fmt.Sprintf("定时报告: %s", scheduledReport.Name)
			sendErr = s.emailService.SendEmail(emails, subject, pushContent)
		} else {
			s.logger.Warn("no valid email addresses found for recipients",
				zap.Int64("report_id", scheduledReport.SnowflakeID),
			)
		}
	} else {
		// WeChat placeholder
		s.logger.Info("WeChat push not yet integrated, recording success",
			zap.Int64("report_id", scheduledReport.SnowflakeID),
		)
	}

	if sendErr != nil {
		return s.recordPushFailure(ctx, scheduledReport, sendErr)
	}
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
	content := fmt.Sprintf("# %s\n\n", scheduledReport.Name)

	if scheduledReport.Description != "" {
		content += fmt.Sprintf("**描述**: %s\n\n", scheduledReport.Description)
	}

	content += fmt.Sprintf("**执行时间**: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	content += "## 查询结果\n\n"
	for key, value := range queryResult {
		content += fmt.Sprintf("- **%s**: %v\n", key, value)
	}

	content += "\n---\n"
	content += fmt.Sprintf("\n*本报告由ERP对话式查询助手自动生成*\n")

	return content, nil
}

func (s *PushAppService) recordPushSuccess(ctx context.Context, scheduledReport *report.ScheduledReport, pushContent string) error {
	pushRecord := &report.PushRecord{
		SnowflakeID: s.idGenerator.Generate(),
		ReportID:    scheduledReport.SnowflakeID,
		PushType:    report.PushTypeReport,
		PushContent: pushContent,
		PushTargets: scheduledReport.Recipients,
		PushChannel: scheduledReport.PushChannel,
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
	pushRecord := &report.PushRecord{
		SnowflakeID:  s.idGenerator.Generate(),
		ReportID:     scheduledReport.SnowflakeID,
		PushType:     report.PushTypeReport,
		PushContent:  "",
		PushTargets:  scheduledReport.Recipients,
		PushChannel:  scheduledReport.PushChannel,
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

// ExecuteAndPushAlert pushes an alert notification when the alert condition is triggered.
func (s *PushAppService) ExecuteAndPushAlert(ctx context.Context, alert *report.Alert, queryResult map[string]interface{}) error {
	s.logger.Info("starting to push alert",
		zap.Int64("alert_id", alert.SnowflakeID),
		zap.String("alert_name", alert.Name),
	)

	// Generate alert content
	pushContent := fmt.Sprintf("# 预警通知: %s\n\n", alert.Name)
	if alert.Description != "" {
		pushContent += fmt.Sprintf("**描述**: %s\n\n", alert.Description)
	}
	pushContent += fmt.Sprintf("**触发条件**: %s\n\n", alert.Condition)
	pushContent += fmt.Sprintf("**触发时间**: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	pushContent += "## 监测数据\n\n"
	for key, value := range queryResult {
		pushContent += fmt.Sprintf("- **%s**: %v\n", key, value)
	}
	pushContent += "\n---\n"
	pushContent += "\n*本预警由ERP对话式查询助手自动发送*\n"

	// Dispatch based on push channel
	var sendErr error
	if alert.PushChannel == report.PushChannelEmail {
		emails, err := s.recipientResolver.ResolveEmails(ctx, alert.Recipients)
		if err != nil {
			s.logger.Error("failed to resolve alert recipient emails", zap.Error(err))
			return s.recordAlertPushFailure(ctx, alert, fmt.Errorf("failed to resolve recipient emails: %w", err))
		}
		if len(emails) > 0 {
			subject := fmt.Sprintf("预警通知: %s", alert.Name)
			sendErr = s.emailService.SendEmail(emails, subject, pushContent)
		} else {
			s.logger.Warn("no valid email addresses found for alert recipients",
				zap.Int64("alert_id", alert.SnowflakeID),
			)
		}
	} else {
		s.logger.Info("WeChat alert push not yet integrated",
			zap.Int64("alert_id", alert.SnowflakeID),
		)
	}

	if sendErr != nil {
		return s.recordAlertPushFailure(ctx, alert, sendErr)
	}
	return s.recordAlertPushSuccess(ctx, alert, pushContent)
}

func (s *PushAppService) recordAlertPushSuccess(ctx context.Context, alert *report.Alert, pushContent string) error {
	pushRecord := &report.PushRecord{
		SnowflakeID: s.idGenerator.Generate(),
		AlertRuleID: alert.SnowflakeID,
		PushType:    report.PushTypeAlert,
		PushContent: pushContent,
		PushTargets: alert.Recipients,
		PushChannel: alert.PushChannel,
		PushStatus:  report.PushStatusSuccess,
		PushTime:    time.Now(),
		RetryCount:  0,
	}

	if err := s.reportRepo.CreatePushRecord(ctx, pushRecord); err != nil {
		return fmt.Errorf("failed to record alert push success: %w", err)
	}

	s.logger.Info("alert push succeeded and recorded",
		zap.Int64("alert_id", alert.SnowflakeID),
		zap.Int64("push_record_id", pushRecord.SnowflakeID),
	)
	return nil
}

func (s *PushAppService) recordAlertPushFailure(ctx context.Context, alert *report.Alert, pushErr error) error {
	pushRecord := &report.PushRecord{
		SnowflakeID:  s.idGenerator.Generate(),
		AlertRuleID:  alert.SnowflakeID,
		PushType:     report.PushTypeAlert,
		PushContent:  "",
		PushTargets:  alert.Recipients,
		PushChannel:  alert.PushChannel,
		PushStatus:   report.PushStatusFailed,
		PushTime:     time.Now(),
		ErrorMessage: pushErr.Error(),
		RetryCount:   0,
	}

	if err := s.reportRepo.CreatePushRecord(ctx, pushRecord); err != nil {
		return fmt.Errorf("failed to record alert push failure: %w", err)
	}

	s.logger.Info("alert push failed and recorded",
		zap.Int64("alert_id", alert.SnowflakeID),
		zap.String("error", pushErr.Error()),
	)
	return pushErr
}
