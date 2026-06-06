// Package v1 implements the operation log HTTP handlers.
package v1

import (
	"strconv"

	"jiufang/internal/service"
	"jiufang/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// OperationLogHandler handles HTTP requests for operation logs.
type OperationLogHandler struct {
	logService *service.OperationLogService
	logger     *zap.Logger
}

// NewOperationLogHandler creates a new OperationLogHandler instance.
func NewOperationLogHandler(logService *service.OperationLogService, logger *zap.Logger) *OperationLogHandler {
	return &OperationLogHandler{
		logService: logService,
		logger:     logger,
	}
}

// ListOperationLogs handles GET /api/v1/logs - retrieves a list of operation logs.
func (h *OperationLogHandler) ListOperationLogs(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	userIDStr := c.Query("user_id")
	operationType := c.Query("operation_type")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	// Parse user ID
	var userID int64
	if userIDStr != "" {
		userID, _ = strconv.ParseInt(userIDStr, 10, 64)
	}

	// List operation logs
	logs, total, err := h.logService.ListOperationLogs(ctx, page, size, userID, operationType, startTime, endTime)
	if err != nil {
		h.logger.Error("failed to list operation logs",
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Int64("user_id", userID),
			zap.String("operation_type", operationType),
			zap.Error(err),
		)
		response.InternalError(c, "failed to list operation logs: "+err.Error())
		return
	}

	// Convert logs to response format and get usernames
	logResponses := make([]gin.H, len(logs))
	for i, log := range logs {
		username := ""
		if log.UserID != nil && *log.UserID > 0 {
			username, _ = h.logService.GetUsernameByUserID(ctx, *log.UserID)
		}

		userIDStr := ""
		if log.UserID != nil && *log.UserID > 0 {
			userIDStr = strconv.FormatInt(*log.UserID, 10)
			username, _ = h.logService.GetUsernameByUserID(ctx, *log.UserID)
		}

		logResponses[i] = gin.H{
			"id":               strconv.FormatInt(log.SnowflakeID, 10),
			"user_id":          userIDStr,
			"username":         username,
			"operation_type":   log.OperationType,
			"operation_object": log.OperationObject,
			"operation_detail": log.OperationDetail,
			"operation_result": log.OperationResult,
			"ip_address":       log.IPAddress,
			"created_at":       log.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	response.Success(c, gin.H{
		"logs":      logResponses,
		"total":     total,
		"page":      page,
		"page_size": size,
	})
}

// GetOperationLog handles GET /api/v1/logs/:id - retrieves an operation log by ID.
func (h *OperationLogHandler) GetOperationLog(c *gin.Context) {
	ctx := c.Request.Context()

	// Get log ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid operation log ID")
		return
	}

	// Get operation log
	log, err := h.logService.GetOperationLogByID(ctx, id)
	if err != nil {
		h.logger.Error("failed to get operation log",
			zap.Int64("log_id", id),
			zap.Error(err),
		)
		response.InternalError(c, "failed to get operation log: "+err.Error())
		return
	}

	if log == nil {
		response.NotFound(c, "operation log not found")
		return
	}

	// Get username
	username := ""
	if log.UserID != nil && *log.UserID > 0 {
		username, _ = h.logService.GetUsernameByUserID(ctx, *log.UserID)
	}

	userIDStr := ""
	if log.UserID != nil && *log.UserID > 0 {
		userIDStr = strconv.FormatInt(*log.UserID, 10)
	}

	response.Success(c, gin.H{
		"id":               strconv.FormatInt(log.SnowflakeID, 10),
		"user_id":          userIDStr,
		"username":         username,
		"operation_type":   log.OperationType,
		"operation_object": log.OperationObject,
		"operation_detail": log.OperationDetail,
		"operation_result": log.OperationResult,
		"ip_address":       log.IPAddress,
		"created_at":       log.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}