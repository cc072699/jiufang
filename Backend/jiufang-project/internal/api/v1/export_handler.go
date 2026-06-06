// Package v1 implements the HTTP handlers for API version 1.
package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"jiufang/internal/model/export"
	"jiufang/internal/service"
)

// ExportHandler handles export-related HTTP requests.
type ExportHandler struct {
	exportAppService service.ExportAppServiceInterface
	logger           *zap.Logger
}

// NewExportHandler creates a new ExportHandler instance.
func NewExportHandler(
	exportAppService service.ExportAppServiceInterface,
	logger *zap.Logger,
) *ExportHandler {
	return &ExportHandler{
		exportAppService: exportAppService,
		logger:           logger,
	}
}

// ExportQueryResult handles POST /api/v1/export requests.
// It exports query results to Excel or PDF format.
func (h *ExportHandler) ExportQueryResult(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    40101,
			"message": "JWT token invalid",
		})
		return
	}

	// Bind request
	var req export.ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind export request",
			zap.Error(err),
			zap.Any("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "Parameter validation failed",
			"error":   err.Error(),
		})
		return
	}

	// Check export permission (TODO: integrate with permission service)
	// For now, assume all authenticated users have export permission

	// Execute export
	result, err := h.exportAppService.ExportQueryResult(c.Request.Context(), userID.(int64), &req)
	if err != nil {
		h.logger.Error("Failed to export query result",
			zap.Error(err),
			zap.Any("user_id", userID),
			zap.Any("request", req),
		)

		// Determine error response based on error type
		if err.Error() == "query record not owned by user" {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    40301,
				"message": "Permission denied",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "Export failed",
			"error":   err.Error(),
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    result,
	})
}

// GetExportRecords handles GET /api/v1/export/records requests.
// It retrieves the export record list for the current user.
func (h *ExportHandler) GetExportRecords(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    40101,
			"message": "JWT token invalid",
		})
		return
	}

	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Get export record list
	records, total, err := h.exportAppService.GetExportRecordList(c.Request.Context(), userID.(int64), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get export record list",
			zap.Error(err),
			zap.Any("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "Failed to get export record list",
			"error":   err.Error(),
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"records":   records,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}
