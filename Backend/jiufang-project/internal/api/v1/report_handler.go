package v1

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"jiufang/internal/model/report"
	"jiufang/internal/pkg/response"
	"jiufang/internal/service"
)

// ReportHandler handles HTTP requests for scheduled reports.
type ReportHandler struct {
	reportService service.ReportServiceInterface
	alertService  *service.AlertService
	logger        *zap.Logger
}

// NewReportHandler creates a new ReportHandler instance.
func NewReportHandler(
	reportService service.ReportServiceInterface,
	alertService *service.AlertService,
	logger *zap.Logger,
) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
		alertService:  alertService,
		logger:        logger,
	}
}

// CreateReport handles POST /api/v1/reports requests.
func (h *ReportHandler) CreateReport(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req report.CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters: "+err.Error())
		return
	}

	// Set created_by from authenticated user
	req.CreatedBy = userID.(int64)

	// Create report
	scheduledReport, err := h.reportService.CreateReport(ctx, &req)
	if err != nil {
		h.logger.Error("failed to create report",
			zap.String("name", req.Name),
			zap.Error(err),
		)
		response.InternalError(c, "failed to create report: "+err.Error())
		return
	}

	// Parse recipients from JSON string
	var recipients []string
	if err := json.Unmarshal([]byte(scheduledReport.Recipients), &recipients); err != nil {
		h.logger.Error("failed to parse recipients",
			zap.String("recipients", scheduledReport.Recipients),
			zap.Error(err),
		)
		recipients = []string{}
	}

	response.Success(c, gin.H{
		"id":            strconv.FormatInt(scheduledReport.SnowflakeID, 10),
		"name":          scheduledReport.Name,
		"description":   scheduledReport.Description,
		"sql":           scheduledReport.SQL,
		"schedule_type": scheduledReport.ScheduleType,
		"schedule_time": scheduledReport.ScheduleTime,
		"recipients":    recipients,
		"push_channel":  scheduledReport.PushChannel,
		"status":        scheduledReport.Status,
		"created_by":    strconv.FormatInt(scheduledReport.CreatedBy, 10),
		"created_at":    scheduledReport.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// ListReports handles GET /api/v1/reports requests.
func (h *ReportHandler) ListReports(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	name := c.Query("name")
	status := c.Query("status")

	// Build request
	req := &report.ListReportsRequest{
		Page:     page,
		PageSize: pageSize,
		Name:     name,
		Status:   status,
	}

	// List reports
	scheduledReports, total, err := h.reportService.ListReports(ctx, req)
	if err != nil {
		h.logger.Error("failed to list reports",
			zap.Int("page", page),
			zap.Int("page_size", pageSize),
			zap.Error(err),
		)
		response.InternalError(c, "failed to list reports: "+err.Error())
		return
	}

	// Build response
	reportList := make([]gin.H, 0)
	for _, scheduledReport := range scheduledReports {
		// Parse recipients from JSON string
		var recipients []string
		if err := json.Unmarshal([]byte(scheduledReport.Recipients), &recipients); err != nil {
			recipients = []string{}
		}

		reportList = append(reportList, gin.H{
			"id":              strconv.FormatInt(scheduledReport.SnowflakeID, 10),
			"name":            scheduledReport.Name,
			"description":     scheduledReport.Description,
			"sql":             scheduledReport.SQL,
			"schedule_type":   scheduledReport.ScheduleType,
			"schedule_time":   scheduledReport.ScheduleTime,
			"recipients":      recipients,
			"recipients_count": len(recipients),
			"push_channel":    scheduledReport.PushChannel,
			"status":          scheduledReport.Status,
			"last_run_at":     nil,
			"created_by":      strconv.FormatInt(scheduledReport.CreatedBy, 10),
			"created_at":      scheduledReport.CreatedAt.Format("2006-01-02T15:04:05Z"),
			"updated_at":      scheduledReport.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	response.Success(c, gin.H{
		"reports":   reportList,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetReport handles GET /api/v1/reports/:id requests.
func (h *ReportHandler) GetReport(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse report ID
	reportIDStr := c.Param("id")
	reportID, err := strconv.ParseInt(reportIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid report id")
		return
	}

	// Get report
	scheduledReport, err := h.reportService.GetReportByID(ctx, reportID)
	if err != nil {
		h.logger.Error("failed to get report",
			zap.Int64("report_id", reportID),
			zap.Error(err),
		)
		response.NotFound(c, "report not found")
		return
	}

	// Parse recipients from JSON string
	var recipients []string
	if err := json.Unmarshal([]byte(scheduledReport.Recipients), &recipients); err != nil {
		recipients = []string{}
	}

	response.Success(c, gin.H{
		"id":            strconv.FormatInt(scheduledReport.SnowflakeID, 10),
		"name":          scheduledReport.Name,
		"description":   scheduledReport.Description,
		"sql":           scheduledReport.SQL,
		"schedule_type": scheduledReport.ScheduleType,
		"schedule_time": scheduledReport.ScheduleTime,
		"recipients":    recipients,
		"push_channel":  scheduledReport.PushChannel,
		"status":        scheduledReport.Status,
		"created_by":    strconv.FormatInt(scheduledReport.CreatedBy, 10),
		"created_at":    scheduledReport.CreatedAt.Format("2006-01-02T15:04:05Z"),
		"updated_at":    scheduledReport.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// UpdateReport handles PUT /api/v1/reports/:id requests.
func (h *ReportHandler) UpdateReport(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse report ID
	reportIDStr := c.Param("id")
	reportID, err := strconv.ParseInt(reportIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid report id")
		return
	}

	var req report.UpdateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters: "+err.Error())
		return
	}

	// Update report
	if err := h.reportService.UpdateReport(ctx, reportID, &req); err != nil {
		h.logger.Error("failed to update report",
			zap.Int64("report_id", reportID),
			zap.Error(err),
		)
		response.InternalError(c, "failed to update report: "+err.Error())
		return
	}

	// Get updated report
	scheduledReport, err := h.reportService.GetReportByID(ctx, reportID)
	if err != nil {
		response.InternalError(c, "failed to get updated report")
		return
	}

	// Parse recipients from JSON string
	var recipients []string
	if err := json.Unmarshal([]byte(scheduledReport.Recipients), &recipients); err != nil {
		recipients = []string{}
	}

	response.Success(c, gin.H{
		"id":            strconv.FormatInt(scheduledReport.SnowflakeID, 10),
		"name":          scheduledReport.Name,
		"description":   scheduledReport.Description,
		"sql":           scheduledReport.SQL,
		"schedule_type": scheduledReport.ScheduleType,
		"schedule_time": scheduledReport.ScheduleTime,
		"recipients":    recipients,
		"push_channel":  scheduledReport.PushChannel,
		"status":        scheduledReport.Status,
		"created_by":    strconv.FormatInt(scheduledReport.CreatedBy, 10),
		"created_at":    scheduledReport.CreatedAt.Format("2006-01-02T15:04:05Z"),
		"updated_at":    scheduledReport.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// DeleteReport handles DELETE /api/v1/reports/:id requests.
func (h *ReportHandler) DeleteReport(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse report ID
	reportIDStr := c.Param("id")
	reportID, err := strconv.ParseInt(reportIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid report id")
		return
	}

	// Delete report
	if err := h.reportService.DeleteReport(ctx, reportID); err != nil {
		h.logger.Error("failed to delete report",
			zap.Int64("report_id", reportID),
			zap.Error(err),
		)
		response.InternalError(c, "failed to delete report: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "report deleted successfully",
	})
}

// ListPushRecords handles GET /api/v1/reports/push-records requests.
func (h *ReportHandler) ListPushRecords(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	pushType := c.Query("push_type")
	status := c.Query("status")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	// Build request
	req := &report.GetPushRecordsRequest{
		Page:     page,
		PageSize: pageSize,
		PushType: report.PushType(pushType),
		Status:   report.PushStatus(status),
	}

	// Parse time parameters
	if startTime != "" {
		t, err := time.Parse("2006-01-02 15:04:05", startTime)
		if err == nil {
			req.StartTime = t
		}
	}
	if endTime != "" {
		t, err := time.Parse("2006-01-02 15:04:05", endTime)
		if err == nil {
			req.EndTime = t
		}
	}

	// Get push records
	pushRecords, total, err := h.reportService.GetPushRecords(ctx, req)
	if err != nil {
		h.logger.Error("failed to get push records",
			zap.Int("page", page),
			zap.Int("page_size", pageSize),
			zap.Error(err),
		)
		response.InternalError(c, "failed to get push records: "+err.Error())
		return
	}

	// Build response - remap fields to match design doc (API-027)
	pushRecordList := make([]gin.H, 0)
	for _, pushRecord := range pushRecords {
		// Resolve source_id and source_name from push_type
		var sourceID string
		var sourceName string
		if pushRecord.PushType == report.PushTypeReport {
			sourceID = strconv.FormatInt(pushRecord.ReportID, 10)
			if r, err := h.reportService.GetReportByID(ctx, pushRecord.ReportID); err == nil {
				sourceName = r.Name
			}
		} else {
			sourceID = strconv.FormatInt(pushRecord.AlertRuleID, 10)
			if a, err := h.alertService.GetAlertByID(ctx, pushRecord.AlertRuleID); err == nil {
				sourceName = a.Name
			}
		}

		// Extract first recipient from JSON array
		var recipients []string
		_ = json.Unmarshal([]byte(pushRecord.PushTargets), &recipients)
		recipient := ""
		if len(recipients) > 0 {
			recipient = recipients[0]
		}

		pushRecordList = append(pushRecordList, gin.H{
			"id":            strconv.FormatInt(pushRecord.SnowflakeID, 10),
			"push_type":     pushRecord.PushType,
			"source_id":     sourceID,
			"source_name":   sourceName,
			"recipient":     recipient,
			"status":        pushRecord.PushStatus,
			"error_message": pushRecord.ErrorMessage,
			"pushed_at":     pushRecord.PushTime.Format("2006-01-02T15:04:05Z"),
		})
	}

	response.Success(c, gin.H{
		"records":   pushRecordList,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
