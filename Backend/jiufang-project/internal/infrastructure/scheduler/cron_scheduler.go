// Package scheduler implements the cron-based task scheduler.
package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"jiufang/internal/model/report"
	"jiufang/internal/repository"
	"jiufang/internal/service"
)
type CronScheduler struct {
	cron         *cron.Cron
	reportRepo   repository.ReportRepositoryInterface
	pushService  service.PushAppServiceInterface
	logger       *zap.Logger
	entryMap     map[int64]cron.EntryID // Maps report snowflake ID to cron entry ID
	mu           sync.RWMutex
}

// NewCronScheduler creates a new CronScheduler instance.
func NewCronScheduler(
	reportRepo repository.ReportRepositoryInterface,
	pushService service.PushAppServiceInterface,
	logger *zap.Logger,
) *CronScheduler {
	// Create cron with second precision
	c := cron.New(cron.WithSeconds())

	return &CronScheduler{
		cron:        c,
		reportRepo:  reportRepo,
		pushService: pushService,
		logger:      logger,
		entryMap:    make(map[int64]cron.EntryID),
	}
}

// Start starts the cron scheduler.
func (s *CronScheduler) Start() {
	s.logger.Info("starting cron scheduler")
	s.cron.Start()
	s.logger.Info("cron scheduler started successfully")
}

// Stop stops the cron scheduler.
func (s *CronScheduler) Stop() {
	s.logger.Info("stopping cron scheduler")
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.logger.Info("cron scheduler stopped successfully")
}

// LoadActiveReports loads all active reports and schedules them.
func (s *CronScheduler) LoadActiveReports(ctx context.Context) error {
	s.logger.Info("loading active scheduled reports")

	// Get all active reports from database
	activeReports, err := s.reportRepo.GetActiveReports(ctx)
	if err != nil {
		s.logger.Error("failed to get active reports", zap.Error(err))
		return fmt.Errorf("failed to get active reports: %w", err)
	}

	s.logger.Info("found active reports", zap.Int("count", len(activeReports)))

	// Schedule each report
	for _, scheduledReport := range activeReports {
		if err := s.AddReport(&scheduledReport); err != nil {
			s.logger.Error("failed to schedule report",
				zap.Int64("report_id", scheduledReport.SnowflakeID),
				zap.String("report_name", scheduledReport.Name),
				zap.Error(err),
			)
			// Continue with other reports even if one fails
			continue
		}

		s.logger.Info("report scheduled successfully",
			zap.Int64("report_id", scheduledReport.SnowflakeID),
			zap.String("report_name", scheduledReport.Name),
			zap.String("cron", scheduledReport.ScheduleCron),
		)
	}

	return nil
}

// AddReport adds a scheduled report to the cron scheduler.
func (s *CronScheduler) AddReport(scheduledReport *report.ScheduledReport) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if report is already scheduled
	if _, exists := s.entryMap[scheduledReport.SnowflakeID]; exists {
		s.logger.Warn("report already scheduled, removing old schedule",
			zap.Int64("report_id", scheduledReport.SnowflakeID),
		)
		s.removeReportUnsafe(scheduledReport.SnowflakeID)
	}

	// Add cron job
	entryID, err := s.cron.AddFunc(scheduledReport.ScheduleCron, func() {
		s.executeReport(scheduledReport)
	})

	if err != nil {
		s.logger.Error("failed to add cron job",
			zap.Int64("report_id", scheduledReport.SnowflakeID),
			zap.String("cron", scheduledReport.ScheduleCron),
			zap.Error(err),
		)
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	// Store entry ID
	s.entryMap[scheduledReport.SnowflakeID] = entryID

	s.logger.Info("report added to scheduler",
		zap.Int64("report_id", scheduledReport.SnowflakeID),
		zap.Int("entry_id", int(entryID)),
		zap.String("cron", scheduledReport.ScheduleCron),
	)

	return nil
}

// RemoveReport removes a scheduled report from the cron scheduler.
func (s *CronScheduler) RemoveReport(snowflakeID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.removeReportUnsafe(snowflakeID)
}

// removeReportUnsafe removes a report without locking (internal use).
func (s *CronScheduler) removeReportUnsafe(snowflakeID int64) error {
	entryID, exists := s.entryMap[snowflakeID]
	if !exists {
		s.logger.Warn("report not found in scheduler",
			zap.Int64("report_id", snowflakeID),
		)
		return fmt.Errorf("report not found in scheduler")
	}

	// Remove cron job
	s.cron.Remove(entryID)

	// Remove from map
	delete(s.entryMap, snowflakeID)

	s.logger.Info("report removed from scheduler",
		zap.Int64("report_id", snowflakeID),
		zap.Int("entry_id", int(entryID)),
	)

	return nil
}

// UpdateReport updates a scheduled report in the cron scheduler.
func (s *CronScheduler) UpdateReport(scheduledReport *report.ScheduledReport) error {
	// Remove old schedule
	if err := s.RemoveReport(scheduledReport.SnowflakeID); err != nil {
		// Log warning but continue (report might not exist)
		s.logger.Warn("failed to remove old schedule during update",
			zap.Int64("report_id", scheduledReport.SnowflakeID),
			zap.Error(err),
		)
	}

	// Add new schedule
	return s.AddReport(scheduledReport)
}

// GetScheduledReportIDs returns all scheduled report IDs.
func (s *CronScheduler) GetScheduledReportIDs() []int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]int64, 0, len(s.entryMap))
	for id := range s.entryMap {
		ids = append(ids, id)
	}
	return ids
}

// executeReport executes a scheduled report.
func (s *CronScheduler) executeReport(scheduledReport *report.ScheduledReport) {
	s.logger.Info("executing scheduled report",
		zap.Int64("report_id", scheduledReport.SnowflakeID),
		zap.String("report_name", scheduledReport.Name),
	)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Execute and push report
	if err := s.pushService.ExecuteAndPushReport(ctx, scheduledReport); err != nil {
		s.logger.Error("failed to execute and push report",
			zap.Int64("report_id", scheduledReport.SnowflakeID),
			zap.String("report_name", scheduledReport.Name),
			zap.Error(err),
		)
		return
	}

	s.logger.Info("scheduled report executed successfully",
		zap.Int64("report_id", scheduledReport.SnowflakeID),
		zap.String("report_name", scheduledReport.Name),
	)
}