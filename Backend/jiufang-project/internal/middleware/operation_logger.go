package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"strings"
	"time"

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
		// Capture request body before c.Next() (body is consumed by handlers)
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		reqTime := time.Now()
		c.Next()

		method := c.Request.Method
		if method == "GET" || method == "OPTIONS" || method == "HEAD" {
			return
		}

		path := c.Request.URL.Path
		status := c.Writer.Status()

		opType := mapOperationType(method, path)
		if opType == "" {
			return
		}

		// Get user ID from context
		userID, _ := c.Get(UserIDKey)
		var uid int64
		if id, ok := userID.(int64); ok {
			uid = id
		} else if id, ok := userID.(uint); ok {
			uid = int64(id)
		}

		// For login, extract user ID from the response or request body
		if uid == 0 && opType == audit.OperationTypeLogin {
			uid = extractLoginUserID(c, bodyBytes)
		}

		result := audit.OperationResultSuccess
		if status >= 400 {
			result = audit.OperationResultFailed
		}

		ip := getRealIP(c)

		detail := buildOperationDetail(opType, path, bodyBytes, status)

		go m.logService.RecordOperation(
			context.Background(),
			uid,
			opType,
			path,
			detail,
			result,
			ip,
			reqTime,
		)
	}
}

// getRealIP extracts the real client IP address.
func getRealIP(c *gin.Context) string {
	// Try X-Real-IP header first
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	// Try X-Forwarded-For header (first IP is the original client)
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		if ip := strings.Split(xff, ",")[0]; ip != "" {
			return strings.TrimSpace(ip)
		}
	}
	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return ip
}

// extractLoginUserID tries to get the user ID for login operations.
func extractLoginUserID(c *gin.Context, bodyBytes []byte) int64 {
	// After login handler runs, the auth middleware sets user_id in context
	// But for login endpoint, auth middleware doesn't run before the handler
	// So we try to get it from the response writer's captured data
	// Best effort: store username for later resolution
	var body struct {
		Username string `json:"username"`
	}
	if err := json.Unmarshal(bodyBytes, &body); err == nil && body.Username != "" {
		c.Set("_log_username", body.Username)
	}

	// Try to get user_id that may have been set after successful login
	if uid, exists := c.Get(UserIDKey); exists {
		if id, ok := uid.(int64); ok && id > 0 {
			return id
		}
		if id, ok := uid.(uint); ok && id > 0 {
			return int64(id)
		}
	}
	return 0
}

// buildOperationDetail creates a human-readable detail string for the operation.
func buildOperationDetail(opType audit.OperationType, path string, bodyBytes []byte, status int) string {
	var body map[string]interface{}
	if len(bodyBytes) > 0 {
		json.Unmarshal(bodyBytes, &body)
	}

	switch opType {
	case audit.OperationTypeQuery:
		if input, ok := body["input"].(string); ok {
			return "查询内容: " + input
		}
	case audit.OperationTypeCreateUser:
		if name, ok := body["username"].(string); ok {
			return "创建用户: " + name
		}
	case audit.OperationTypeUpdateUser:
		return "更新用户: " + lastPathSegment(path)
	case audit.OperationTypeDeleteUser:
		return "删除用户: " + lastPathSegment(path)
	case audit.OperationTypeConfigPermission:
		return "配置权限: " + path
	case audit.OperationTypeCreateReport:
		if name, ok := body["name"].(string); ok {
			return "创建报告: " + name
		}
	case audit.OperationTypeUpdateReport:
		return "更新报告: " + lastPathSegment(path)
	case audit.OperationTypeDeleteReport:
		return "删除报告: " + lastPathSegment(path)
	case audit.OperationTypeCreateAlert:
		if name, ok := body["name"].(string); ok {
			return "创建预警: " + name
		}
	case audit.OperationTypeUpdateAlert:
		return "更新预警: " + lastPathSegment(path)
	case audit.OperationTypeDeleteAlert:
		return "删除预警: " + lastPathSegment(path)
	case audit.OperationTypeCreateFavorite:
		if name, ok := body["name"].(string); ok {
			return "收藏查询: " + name
		}
	case audit.OperationTypeDeleteFavorite:
		return "删除收藏: " + lastPathSegment(path)
	case audit.OperationTypeDeleteHistory:
		return "删除历史: " + lastPathSegment(path)
	case audit.OperationTypeCreateGroup:
		if name, ok := body["name"].(string); ok {
			return "创建用户组: " + name
		}
	case audit.OperationTypeUpdateGroup:
		return "更新用户组: " + lastPathSegment(path)
	case audit.OperationTypeDeleteGroup:
		return "删除用户组: " + lastPathSegment(path)
	case audit.OperationTypeAddMember:
		return "添加成员: " + path
	case audit.OperationTypeRemoveMember:
		return "移除成员: " + path
	case audit.OperationTypeLogin:
		if name, ok := body["username"].(string); ok {
			return "用户登录: " + name
		}
		return "用户登录"
	case audit.OperationTypeLogout:
		return "用户登出"
	case audit.OperationTypeExport:
		return "数据导出"
	case audit.OperationTypeChangePassword:
		return "修改密码"
	case audit.OperationTypeCreateFeedback:
		return "提交反馈"
	}

	return ""
}

// lastPathSegment returns the last segment of a URL path (usually an ID).
func lastPathSegment(path string) string {
	parts := strings.Split(strings.TrimRight(path, "/"), "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return path
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

	// Groups
	case p == "/groups" && method == "POST":
		return audit.OperationTypeCreateGroup
	case strings.HasPrefix(p, "/groups/") && strings.HasSuffix(p, "/members") && method == "POST":
		return audit.OperationTypeAddMember
	case strings.Contains(p, "/members/") && method == "DELETE":
		return audit.OperationTypeRemoveMember
	case strings.HasPrefix(p, "/groups/") && !strings.Contains(p, "/permissions") && !strings.Contains(p, "/members") && method == "PUT":
		return audit.OperationTypeUpdateGroup
	case strings.HasPrefix(p, "/groups/") && !strings.Contains(p, "/permissions") && !strings.Contains(p, "/members") && method == "DELETE":
		return audit.OperationTypeDeleteGroup

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

	// Favorites
	case p == "/favorites" && method == "POST":
		return audit.OperationTypeCreateFavorite
	case strings.HasPrefix(p, "/favorites/") && method == "DELETE":
		return audit.OperationTypeDeleteFavorite

	// History
	case strings.HasPrefix(p, "/history/") && method == "DELETE":
		return audit.OperationTypeDeleteHistory

	// Profile
	case strings.HasSuffix(p, "/password") && method == "PUT":
		return audit.OperationTypeChangePassword

	// Feedback
	case p == "/feedbacks" && method == "POST":
		return audit.OperationTypeCreateFeedback

	// Export
	case p == "/export" && method == "POST":
		return audit.OperationTypeExport
	}

	return ""
}
