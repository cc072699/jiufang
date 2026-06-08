// Package v1 implements the HTTP handlers for natural language query.
package v1

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"jiufang/internal/middleware"
	agentmodel "jiufang/internal/model/agent"
	"jiufang/internal/model/query"
	"jiufang/internal/pkg/id"
	"jiufang/internal/pkg/response"
	"jiufang/internal/service"
)

// QueryHandler handles natural language query HTTP requests.
type QueryHandler struct {
	queryService      *service.QueryAppService
	permissionService *service.PermissionAppService
	logger            *zap.Logger
}

// NewQueryHandler creates a new QueryHandler instance.
func NewQueryHandler(queryService *service.QueryAppService, permissionService *service.PermissionAppService, logger *zap.Logger) *QueryHandler {
	return &QueryHandler{
		queryService:      queryService,
		permissionService: permissionService,
		logger:            logger,
	}
}

// ExecuteQuery handles POST /api/v1/query - execute natural language query.
func (h *QueryHandler) ExecuteQuery(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user ID from context
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Parse request
	var req query.NaturalLanguageQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters: "+err.Error())
		return
	}

	// Validate request
	if req.Input == "" {
		response.BadRequest(c, "input is required")
		return
	}

	// Generate or use existing session ID using snowflake ID generator
	sessionID := req.SessionID
	if sessionID == "" {
		// Generate new session ID using snowflake ID (19-digit number)
		sessionIDInt, err := id.Generate()
		if err != nil {
			h.logger.Error("Failed to generate session ID", zap.Error(err))
			response.InternalError(c, "failed to generate session id")
			return
		}
		sessionID = strconv.FormatInt(sessionIDInt, 10)
	}

	// Check if execute_immediately is false (only return understanding summary)
	if !req.ExecuteImmediately {
		// Only return understanding summary, do not execute query
		// Note: Since there's no dedicated GetUnderstanding method, we return a placeholder message
		response.Success(c, query.QueryResultResponse{
			SessionID:          sessionID,
			Understanding:      "系统已理解您的查询意图，但未执行查询。请设置 execute_immediately=true 以执行查询。",
			ResultType:         "empty",
			SuggestedQuestions: []string{},
			CanExport:          false,
		})
		return
	}

	// Load user permissions from groups
	groupIDs := middleware.GetGroups(c)
	queryContext := h.buildQueryContext(c, sessionID, uint(userID), groupIDs)

	// Non-admin users must have permissions configured
	role := middleware.GetRole(c)
	if role != "admin" && (queryContext == nil || len(queryContext.AllowedTables) == 0) {
		response.Error(c, http.StatusForbidden, "用户未配置查询权限，请联系管理员")
		return
	}

	// Execute query with or without permission filtering
	var result *agentmodel.QueryResult
	var err error
	if queryContext != nil && len(queryContext.AllowedTables) > 0 {
		result, err = h.queryService.ExecuteQueryWithPermission(ctx, req.Input, queryContext)
	} else {
		result, err = h.queryService.ExecuteQuery(ctx, req.Input, sessionID, uint(userID))
	}
	if err != nil {
		h.logger.Error("Failed to execute query",
			zap.String("input", req.Input),
			zap.String("session_id", sessionID),
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
		response.InternalError(c, "failed to execute query: "+err.Error())
		return
	}

	// Build response
	queryResult := query.QueryResultResponse{
		SessionID:          sessionID,
		Understanding:      result.Understanding,
		ResultType:         "table",    // Default to table type
		SQL:                result.GeneratedSQL,
		SuggestedQuestions: []string{}, // TODO: Generate suggested questions based on query context
		CanExport:          true,
	}

	// Add columns if result has data
	if result.Data != nil && len(result.Data) > 0 {
		queryResult.Columns = buildColumnDefinitions(result.Data)
		queryResult.Rows = result.Data
	}

	// Add chart config if available
	if result.ChartConfig != nil {
		queryResult.ResultType = "chart"
		queryResult.ChartConfig = result.ChartConfig
	}

	// Return success response
	response.Success(c, queryResult)
}

// buildQueryContext loads permissions for the user's groups and builds a QueryContext.
func (h *QueryHandler) buildQueryContext(c *gin.Context, sessionID string, userID uint, groupIDs []int64) *agentmodel.QueryContext {
	if len(groupIDs) == 0 || h.permissionService == nil {
		return nil
	}

	reqCtx := c.Request.Context()
	var allowedTables []string
	var allowedFields = make(map[string][]string)
	var tableFilters = make(map[string]string)
	var unrestrictedSet = make(map[string]bool) // tables seen as "*" or ""

	for _, groupID := range groupIDs {
		perms, err := h.permissionService.GetPermissionsByGroup(reqCtx, groupID)
		if err != nil {
			h.logger.Warn("failed to load permissions for group", zap.Int64("group_id", groupID), zap.Error(err))
			continue
		}
		for _, p := range perms {
			if p.TableName != "" {
				allowedTables = append(allowedTables, p.TableName)
				tblKey := strings.ToLower(p.TableName)
				if p.AllowedFields == "*" || p.AllowedFields == "" {
					unrestrictedSet[tblKey] = true
				} else {
					fields := parseAllowedFields(p.AllowedFields)
					allowedFields[tblKey] = append(allowedFields[tblKey], fields...)
				}
				if p.FilterCondition != "" {
					if existing, ok := tableFilters[tblKey]; ok {
						tableFilters[tblKey] = existing + " AND (" + p.FilterCondition + ")"
					} else {
						tableFilters[tblKey] = "(" + p.FilterCondition + ")"
					}
				}
			}
		}
	}

	if len(allowedTables) == 0 {
		return nil
	}

	// If any group specified restricted fields for a table, it stays restricted
	// (specific field restrictions override wildcard access from other groups)
	unrestrictedTables := make(map[string]bool)
	for tbl := range unrestrictedSet {
		if _, hasRestrictions := allowedFields[tbl]; !hasRestrictions {
			unrestrictedTables[tbl] = true
		}
	}

	// Deduplicate allowed tables
	seen := make(map[string]bool)
	var uniqueTables []string
	for _, t := range allowedTables {
		key := strings.ToLower(t)
		if !seen[key] {
			seen[key] = true
			uniqueTables = append(uniqueTables, t)
		}
	}

	return &agentmodel.QueryContext{
		UserID:             userID,
		SessionID:          sessionID,
		AllowedTables:      uniqueTables,
		AllowedFields:      allowedFields,
		TableFilters:       tableFilters,
		UnrestrictedTables: unrestrictedTables,
	}
}

// parseAllowedFields parses allowed fields from either JSON array or comma-separated string format.
func parseAllowedFields(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	// Try JSON array format first: ["field1","field2"]
	if strings.HasPrefix(raw, "[") {
		var arr []string
		if err := json.Unmarshal([]byte(raw), &arr); err == nil {
			var result []string
			for _, f := range arr {
				f = strings.TrimSpace(f)
				if f != "" {
					result = append(result, f)
				}
			}
			return result
		}
	}

	// Fallback: comma-separated format: field1,field2
	parts := strings.Split(raw, ",")
	var result []string
	for _, f := range parts {
		f = strings.TrimSpace(f)
		// Strip surrounding quotes if present
		f = strings.Trim(f, "\"'")
		if f != "" {
			result = append(result, f)
		}
	}
	return result
}

func buildColumnDefinitions(data []map[string]interface{}) []query.ColumnDefinition {
	if len(data) == 0 {
		return []query.ColumnDefinition{}
	}

	// Get column names from first row
	firstRow := data[0]
	columns := make([]query.ColumnDefinition, 0, len(firstRow))

	for name := range firstRow {
		// Determine column type based on value
		value := firstRow[name]
		colType := "string"
		switch value.(type) {
		case int, int64, float64:
			colType = "number"
		case string:
			// Check if it's a date format
			// Simplified check for now
			colType = "string"
		}

		columns = append(columns, query.ColumnDefinition{
			Name: name,
			Type: colType,
		})
	}

	return columns
}
