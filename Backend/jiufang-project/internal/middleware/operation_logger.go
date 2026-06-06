package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	"jiufang/internal/model/audit"
	"jiufang/internal/service"
)

// OperationLogger is a middleware that records write operations to the operation log.
type OperationLogger struct {
	logService *service.OperationLogService
}

// NewOperationLogger creates a new OperationLogger middleware.
func NewOperationLogger(logService *service.OperationLogService) *OperationLogger {
	return &OperationLogger{logService: logService}
}

// Log returns a gin.HandlerFunc that logs POST/PUT/DELETE operations.
func (m *OperationLogger) Log() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		method := c.Request.Method
		if method == "GET" || method == "OPTIONS" || method == "HEAD" {
			return
		}

		path := c.Request.URL.Path
		status := c.Writer.Status()

		// Determine operation type from path and method
		opType := mapOperationType(method, path)
		if opType == "" {
			return
		}

		// Get user ID from context (0 if not authenticated, e.g. login)
		userID, _ := c.Get(UserIDKey)
		var uid int64
		if id, ok := userID.(int64); ok {
			uid = id
		} else if id, ok := userID.(uint); ok {
			uid = int64(id)
		}

		result := audit.OperationResultSuccess
		if status >= 400 {
			result = audit.OperationResultFailed
		}

		ip := c.ClientIP()

		// Record asynchronously to avoid blocking the response
		go m.logService.RecordOperation(
			context.Background(),
			uid,
			opType,
			path,
			"",
			result,
			ip,
		)
	}
}

// mapOperationType maps HTTP method + path to an OperationType.
func mapOperationType(method, path string) audit.OperationType {
	p := strings.TrimPrefix(path, "/api/v1")

	switch {
	// Auth
	case p == "/auth/login" && method == "POST":
		return audit.OperationTypeLogin
	case p == "/auth/logout" && method == "POST":
		return audit.OperationTypeLogout

	// Query
	case p == "/query" && method == "POST":
		return audit.OperationTypeQuery

	// Users
	case p == "/users" && method == "POST":
		return audit.OperationTypeCreateUser
	case strings.HasPrefix(p, "/users/") && method == "PUT":
		return audit.OperationTypeUpdateUser
	case strings.HasPrefix(p, "/users/") && method == "DELETE":
		return audit.OperationTypeDeleteUser

	// Permissions
	case strings.HasSuffix(p, "/permissions") && method == "POST":
		return audit.OperationTypeConfigPermission

	// Reports
	case p == "/reports" && method == "POST":
		return audit.OperationTypeCreateReport
	case strings.HasPrefix(p, "/reports/") && method == "PUT":
		return audit.OperationTypeUpdateReport
	case strings.HasPrefix(p, "/reports/") && method == "DELETE":
		return audit.OperationTypeDeleteReport

	// Alerts
	case p == "/alerts" && method == "POST":
		return audit.OperationTypeCreateAlert
	case strings.HasPrefix(p, "/alerts/") && method == "PUT":
		return audit.OperationTypeUpdateAlert
	case strings.HasPrefix(p, "/alerts/") && method == "DELETE":
		return audit.OperationTypeDeleteAlert

	// Export
	case p == "/export" && method == "POST":
		return audit.OperationTypeExport
	}

	return ""
}
