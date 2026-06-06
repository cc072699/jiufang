// Package v1 implements the alert HTTP handlers.
package v1

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"jiufang/internal/model/report"
	"jiufang/internal/pkg/response"
	"jiufang/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AlertHandler handles HTTP requests for alert rules.
type AlertHandler struct {
	alertService *service.AlertService
	logger       *zap.Logger
}

// NewAlertHandler creates a new AlertHandler instance.
func NewAlertHandler(alertService *service.AlertService, logger *zap.Logger) *AlertHandler {
	return &AlertHandler{
		alertService: alertService,
		logger:       logger,
	}
}

// parseRecipients parses a JSON string of recipients into a string slice.
// Supports both []int64 (seed data format: [1001, 1002]) and []string (API format: ["1001", "1002"]).
func parseRecipients(raw string) []string {
	if raw == "" {
		return []string{}
	}
	// Try []string first (API-created format)
	var strRecipients []string
	if err := json.Unmarshal([]byte(raw), &strRecipients); err == nil {
		return strRecipients
	}
	// Try []int64 (seed data format)
	var intRecipients []int64
	if err := json.Unmarshal([]byte(raw), &intRecipients); err == nil {
		result := make([]string, len(intRecipients))
		for i, id := range intRecipients {
			result[i] = strconv.FormatInt(id, 10)
		}
		return result
	}
	return []string{}
}

// CreateAlert handles POST /api/v1/alerts - creates a new alert rule.
func (h *AlertHandler) CreateAlert(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req report.CreateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters: "+err.Error())
		return
	}

	// Set created_by from authenticated user
	req.CreatedBy = userID.(int64)

	// Create alert
	alert, err := h.alertService.CreateAlert(ctx, &req)
	if err != nil {
		h.logger.Error("failed to create alert",
			zap.Int64("user_id", userID.(int64)),
			zap.String("name", req.Name),
			zap.Error(err),
		)
		response.InternalError(c, "failed to create alert: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"id":                strconv.FormatInt(alert.SnowflakeID, 10),
		"name":              alert.Name,
		"description":       alert.Description,
		"sql":               alert.SQL,
		"condition":         alert.Condition,
		"recipients":        parseRecipients(alert.Recipients),
		"push_channel":      alert.PushChannel,
		"trigger_frequency": alert.TriggerFrequency,
		"silence_start":     formatTime(alert.SilenceStart),
		"silence_end":       formatTime(alert.SilenceEnd),
		"status":            alert.Status,
		"created_by":        strconv.FormatInt(alert.CreatedBy, 10),
		"created_at":        alert.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// GetAlert handles GET /api/v1/alerts/:id - retrieves an alert rule by ID.
func (h *AlertHandler) GetAlert(c *gin.Context) {
	ctx := c.Request.Context()

	// Get alert ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid alert ID")
		return
	}

	// Get alert
	alert, err := h.alertService.GetAlertByID(ctx, id)
	if err != nil {
		h.logger.Error("failed to get alert",
			zap.Int64("alert_id", id),
			zap.Error(err),
		)
		response.InternalError(c, "failed to get alert: "+err.Error())
		return
	}

	if alert == nil {
		response.NotFound(c, "alert not found")
		return
	}

	response.Success(c, gin.H{
		"id":                strconv.FormatInt(alert.SnowflakeID, 10),
		"name":              alert.Name,
		"description":       alert.Description,
		"sql":               alert.SQL,
		"condition":         alert.Condition,
		"recipients":        parseRecipients(alert.Recipients),
		"push_channel":      alert.PushChannel,
		"trigger_frequency": alert.TriggerFrequency,
		"silence_start":     formatTime(alert.SilenceStart),
		"silence_end":       formatTime(alert.SilenceEnd),
		"status":            alert.Status,
		"last_triggered_at": formatTimestamp(alert.LastTriggeredAt),
		"created_by":        strconv.FormatInt(alert.CreatedBy, 10),
		"created_at":        alert.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// ListAlerts handles GET /api/v1/alerts - retrieves a list of alert rules.
func (h *AlertHandler) ListAlerts(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	name := c.Query("name")
	status := c.Query("status")

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// List alerts
	alerts, total, err := h.alertService.ListAlerts(ctx, page, pageSize, name, status)
	if err != nil {
		h.logger.Error("failed to list alerts",
			zap.Int("page", page),
			zap.Int("page_size", pageSize),
			zap.String("name", name),
			zap.String("status", status),
			zap.Error(err),
		)
		response.InternalError(c, "failed to list alerts: "+err.Error())
		return
	}

	// Convert alerts to response format
	alertResponses := make([]gin.H, len(alerts))
	for i, alert := range alerts {
		recipients := parseRecipients(alert.Recipients)
		alertResponses[i] = gin.H{
			"id":                strconv.FormatInt(alert.SnowflakeID, 10),
			"name":              alert.Name,
			"description":       alert.Description,
			"condition":         alert.Condition,
			"recipients":        recipients,
			"recipients_count":  len(recipients),
			"status":            alert.Status,
			"last_triggered_at": formatTimestamp(alert.LastTriggeredAt),
			"created_at":        alert.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	response.Success(c, gin.H{
		"alerts":    alertResponses,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// UpdateAlert handles PUT /api/v1/alerts/:id - updates an existing alert rule.
func (h *AlertHandler) UpdateAlert(c *gin.Context) {
	ctx := c.Request.Context()

	// Get alert ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid alert ID")
		return
	}

	var req report.UpdateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters: "+err.Error())
		return
	}

	// Update alert
	alert, err := h.alertService.UpdateAlert(ctx, id, &req)
	if err != nil {
		h.logger.Error("failed to update alert",
			zap.Int64("alert_id", id),
			zap.Error(err),
		)
		response.InternalError(c, "failed to update alert: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"id":                strconv.FormatInt(alert.SnowflakeID, 10),
		"name":              alert.Name,
		"description":       alert.Description,
		"sql":               alert.SQL,
		"condition":         alert.Condition,
		"recipients":        parseRecipients(alert.Recipients),
		"push_channel":      alert.PushChannel,
		"trigger_frequency": alert.TriggerFrequency,
		"silence_start":     formatTime(alert.SilenceStart),
		"silence_end":       formatTime(alert.SilenceEnd),
		"status":            alert.Status,
		"last_triggered_at": formatTimestamp(alert.LastTriggeredAt),
		"created_by":        strconv.FormatInt(alert.CreatedBy, 10),
		"created_at":        alert.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// DeleteAlert handles DELETE /api/v1/alerts/:id - deletes an alert rule.
func (h *AlertHandler) DeleteAlert(c *gin.Context) {
	ctx := c.Request.Context()

	// Get alert ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid alert ID")
		return
	}

	// Delete alert
	if err := h.alertService.DeleteAlert(ctx, id); err != nil {
		h.logger.Error("failed to delete alert",
			zap.Int64("alert_id", id),
			zap.Error(err),
		)
		response.InternalError(c, "failed to delete alert: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "alert deleted successfully",
	})
}

// formatTime formats a time.Time pointer to "HH:MM" string format.
func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("15:04")
}

// formatTimestamp formats a time.Time pointer to ISO8601 string format.
func formatTimestamp(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02T15:04:05Z07:00")
}
