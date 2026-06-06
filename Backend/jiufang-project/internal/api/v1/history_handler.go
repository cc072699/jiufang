// Package v1 implements the HTTP handlers for query history management.
package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"jiufang/internal/middleware"
	"jiufang/internal/pkg/response"
	"jiufang/internal/service"
)

// HistoryHandler handles query history HTTP requests.
type HistoryHandler struct {
	historyService *service.HistoryAppService
	logger         *zap.Logger
}

// NewHistoryHandler creates a new HistoryHandler instance.
func NewHistoryHandler(historyService *service.HistoryAppService, logger *zap.Logger) *HistoryHandler {
	return &HistoryHandler{
		historyService: historyService,
		logger:         logger,
	}
}

// HistoryItem represents a single history item in the list response (API-013).
type HistoryItem struct {
	ID            string `json:"id"`
	Input         string `json:"input"`
	Sql           string `json:"sql"`
	Status        string `json:"status"`
	ErrorMessage  string `json:"error_message,omitempty"`
	ResultCount   int    `json:"result_count,omitempty"`
	ExecutionTime int    `json:"execution_time,omitempty"`
	CreatedAt     string `json:"created_at"`
}

// GetHistoryList handles GET /api/v1/history - list query history.
func (h *HistoryHandler) GetHistoryList(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user ID from context
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Parse filter parameters
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")
	status := c.Query("status")

	// Get history list
	records, total, err := h.historyService.GetHistoryList(ctx, userID, page, pageSize, startTime, endTime, status)
	if err != nil {
		h.logger.Error("Failed to get history list", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to get history list")
		return
	}

	// Map to list response (filter out session_id, user_id, result_data per design doc)
	items := make([]HistoryItem, len(records))
	for i, r := range records {
		items[i] = HistoryItem{
			ID:            strconv.FormatInt(r.SnowflakeID, 10),
			Input:         r.Input,
			Sql:           r.SQL,
			Status:        string(r.Status),
			ErrorMessage:  r.ErrorMessage,
			ResultCount:   r.ResultCount,
			ExecutionTime: r.ExecutionTime,
			CreatedAt:     r.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	// Return paginated response
	response.PageWithField(c, "records", items, total, page, pageSize)
}

// HistoryDetailItem represents a history detail response (API-014).
type HistoryDetailItem struct {
	ID            string `json:"id"`
	SessionID     string `json:"session_id"`
	Input         string `json:"input"`
	Sql           string `json:"sql"`
	Status        string `json:"status"`
	ErrorMessage  string `json:"error_message,omitempty"`
	ResultCount   int    `json:"result_count,omitempty"`
	ExecutionTime int    `json:"execution_time,omitempty"`
	ResultData    string `json:"result_data,omitempty"`
	CreatedAt     string `json:"created_at"`
}

// GetHistoryDetail handles GET /api/v1/history/:record_id - get history detail.
func (h *HistoryHandler) GetHistoryDetail(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user ID from context
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Get record ID from path parameter
	recordID := c.Param("record_id")
	if recordID == "" {
		response.Error(c, http.StatusBadRequest, "record_id is required")
		return
	}

	// Get history detail
	record, err := h.historyService.GetHistoryDetail(ctx, userID, recordID)
	if err != nil {
		h.logger.Error("Failed to get history detail",
			zap.Error(err),
			zap.String("record_id", recordID),
		)
		response.Error(c, http.StatusNotFound, "query record not found")
		return
	}

	// Return success response (filter out user_id per design doc)
	response.Success(c, HistoryDetailItem{
		ID:            strconv.FormatInt(record.SnowflakeID, 10),
		SessionID:     strconv.FormatInt(record.SessionID, 10),
		Input:         record.Input,
		Sql:           record.SQL,
		Status:        string(record.Status),
		ErrorMessage:  record.ErrorMessage,
		ResultCount:   record.ResultCount,
		ExecutionTime: record.ExecutionTime,
		ResultData:    record.ResultData,
		CreatedAt:     record.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// DeleteHistory handles DELETE /api/v1/history/:record_id - delete history.
func (h *HistoryHandler) DeleteHistory(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user ID from context
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Get record ID from path parameter
	recordID := c.Param("record_id")
	if recordID == "" {
		response.Error(c, http.StatusBadRequest, "record_id is required")
		return
	}

	// Delete history
	if err := h.historyService.DeleteHistory(ctx, userID, recordID); err != nil {
		h.logger.Error("Failed to delete history",
			zap.Error(err),
			zap.String("record_id", recordID),
		)
		response.Error(c, http.StatusNotFound, "query record not found")
		return
	}

	// Return success response
	response.Success(c, nil)
}