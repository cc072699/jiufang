// Package agent implements the AI Agent for semantic understanding and SQL generation.
package agent

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"jiufang/internal/infrastructure/erp"
	"jiufang/internal/model/agent"
)

// SQLExecutor executes SQL queries on the ERP database.
type SQLExecutor struct {
	erpReader erp.ERPReaderInterface
	validator *SQLValidator
}

// NewSQLExecutor creates a new SQL executor.
func NewSQLExecutor(erpReader erp.ERPReaderInterface, validator *SQLValidator) *SQLExecutor {
	return &SQLExecutor{
		erpReader: erpReader,
		validator: validator,
	}
}

// Execute executes a SQL query and returns results.
func (e *SQLExecutor) Execute(ctx context.Context, sql string) (*agent.QueryResult, error) {
	// Validate SQL
	if err := e.validator.Validate(sql); err != nil {
		return nil, err
	}

	// Add safety limit if not present
	safeSQL := e.validator.AddSafetyLimit(sql)

	// Record start time
	startTime := time.Now()

	// Execute query
	data, err := e.erpReader.Query(ctx, safeSQL)
	if err != nil {
		return nil, fmt.Errorf("查询执行失败：%v", err)
	}

	// Calculate execution time
	execTime := time.Since(startTime).Milliseconds()

	// Build query result
	result := agent.NewQueryResult(data, "", execTime, safeSQL)

	return result, nil
}

// ExecuteWithPermission executes SQL with permission check.
func (e *SQLExecutor) ExecuteWithPermission(ctx context.Context, sql string, queryContext *agent.QueryContext) (*agent.QueryResult, error) {
	// Validate SQL with permission
	if err := e.validator.ValidateWithPermission(sql, queryContext); err != nil {
		return nil, err
	}

	// Rewrite SELECT * to only include allowed fields
	sql = rewriteSelectStar(sql, queryContext.AllowedFields, queryContext.UnrestrictedTables)

	// Apply per-table permission filter
	if len(queryContext.TableFilters) > 0 {
		tables := extractTablesFromSQL(sql)
		for _, table := range tables {
			if filter, ok := queryContext.TableFilters[strings.ToLower(table)]; ok {
				sql = addPermissionFilter(sql, filter)
				break // Only apply the first matching table's filter (main table)
			}
		}
	} else if queryContext.PermissionFilter != "" {
		// Fallback: global permission filter
		sql = addPermissionFilter(sql, queryContext.PermissionFilter)
	}

	// Add safety limit
	safeSQL := e.validator.AddSafetyLimit(sql)

	// Record start time
	startTime := time.Now()

	// Execute query
	data, err := e.erpReader.Query(ctx, safeSQL)
	if err != nil {
		return nil, fmt.Errorf("查询执行失败：%v", err)
	}

	// Calculate execution time
	execTime := time.Since(startTime).Milliseconds()

	// Build query result
	result := agent.NewQueryResult(data, "", execTime, safeSQL)
	result.PermissionFiltered = queryContext.PermissionFilter != ""

	// Strip disallowed fields from result data
	if len(queryContext.AllowedFields) > 0 {
		result.Data = stripDisallowedFields(result.Data, queryContext.AllowedFields)
	}

	return result, nil
}

// rewriteSelectStar rewrites SELECT * to SELECT field1, field2, ... for restricted tables.
// If the SQL contains SELECT * and the main table has field restrictions, it replaces * with allowed fields.
func rewriteSelectStar(sql string, allowedFields map[string][]string, unrestrictedTables map[string]bool) string {
	if len(allowedFields) == 0 {
		return sql
	}

	// Check if SQL has SELECT *
	upper := strings.ToUpper(sql)
	selectStarRe := regexp.MustCompile(`(?i)SELECT\s+\*\s+FROM`)
	if !selectStarRe.MatchString(upper) {
		return sql
	}

	// Find the main table in the FROM clause
	tables := extractTablesFromSQL(sql)
	if len(tables) == 0 {
		return sql
	}

	// Get allowed fields for the main table (first table)
	mainTable := strings.ToLower(tables[0])
	if unrestrictedTables[mainTable] {
		return sql // no restriction needed
	}

	fields, ok := allowedFields[mainTable]
	if !ok || len(fields) == 0 {
		return sql // no specific fields configured
	}

	// Build the replacement field list
	allowedList := strings.Join(fields, ", ")

	// Replace SELECT * with SELECT field1, field2, ...
	return selectStarRe.ReplaceAllString(sql, "SELECT "+allowedList+" FROM")
}

// stripDisallowedFields removes columns from query results that are not in the allowed fields list.
func stripDisallowedFields(data []map[string]interface{}, allowedFields map[string][]string) []map[string]interface{} {
	if len(data) == 0 || len(allowedFields) == 0 {
		return data
	}

	// Build a set of all allowed field names (lowercased)
	allowed := make(map[string]bool)
	for _, fields := range allowedFields {
		for _, f := range fields {
			allowed[strings.ToLower(strings.TrimSpace(f))] = true
		}
	}

	// Keep only allowed columns in each row
	var filtered []map[string]interface{}
	for _, row := range data {
		newRow := make(map[string]interface{})
		for col, val := range row {
			lowerCol := strings.ToLower(col)
			// Always keep non-restricted columns (meta columns like row_num, etc.)
			if allowed[lowerCol] || !isRestrictedColumn(col, allowedFields) {
				newRow[col] = val
			}
		}
		filtered = append(filtered, newRow)
	}
	return filtered
}

// isRestrictedColumn checks if a column name corresponds to a restricted table's field.
func isRestrictedColumn(col string, allowedFields map[string][]string) bool {
	// Check if the column appears in any restricted table's field list
	// If it does, it's restricted and must be explicitly allowed
	lowerCol := strings.ToLower(col)
	for _, fields := range allowedFields {
		for _, f := range fields {
			if strings.ToLower(strings.TrimSpace(f)) == lowerCol {
				return true // this column IS in a restricted list, so it's managed
			}
		}
	}
	return false // this column is not in any restricted table's list, keep it
}

// ExecuteWithTimeout executes SQL with timeout.
func (e *SQLExecutor) ExecuteWithTimeout(ctx context.Context, sql string, timeout time.Duration) (*agent.QueryResult, error) {
	// Validate SQL
	if err := e.validator.Validate(sql); err != nil {
		return nil, err
	}

	// Add safety limit
	safeSQL := e.validator.AddSafetyLimit(sql)

	// Record start time
	startTime := time.Now()

	// Execute query with timeout
	data, err := e.erpReader.QueryWithTimeout(ctx, safeSQL, timeout)
	if err != nil {
		return nil, fmt.Errorf("查询执行失败：%v", err)
	}

	// Calculate execution time
	execTime := time.Since(startTime).Milliseconds()

	// Build query result
	result := agent.NewQueryResult(data, "", execTime, safeSQL)

	return result, nil
}

// ExecuteWithLimit executes SQL with result limit.
func (e *SQLExecutor) ExecuteWithLimit(ctx context.Context, sql string, limit int) (*agent.QueryResult, error) {
	// Validate SQL
	if err := e.validator.Validate(sql); err != nil {
		return nil, err
	}

	// Record start time
	startTime := time.Now()

	// Execute query with limit
	data, err := e.erpReader.QueryWithLimit(ctx, sql, limit)
	if err != nil {
		return nil, fmt.Errorf("查询执行失败：%v", err)
	}

	// Calculate execution time
	execTime := time.Since(startTime).Milliseconds()

	// Build query result
	result := agent.NewQueryResult(data, "", execTime, sql)
	result.HasMore = len(data) == limit

	return result, nil
}

// ExecuteSafe executes SQL with all safety measures.
func (e *SQLExecutor) ExecuteSafe(ctx context.Context, sql string, queryContext *agent.QueryContext) (*agent.QueryResult, error) {
	// Step 1: Validate SQL safety
	if err := e.validator.Validate(sql); err != nil {
		return nil, err
	}

	// Step 2: Validate permissions
	if queryContext != nil {
		if err := e.validator.ValidateWithPermission(sql, queryContext); err != nil {
			return nil, err
		}
	}

	// Step 3: Add permission filter
	if queryContext != nil && queryContext.PermissionFilter != "" {
		sql = addPermissionFilter(sql, queryContext.PermissionFilter)
	}

	// Step 4: Add safety limit
	safeSQL := e.validator.AddSafetyLimit(sql)

	// Step 5: Execute with timeout
	timeout := 30 * time.Second
	startTime := time.Now()

	data, err := e.erpReader.QueryWithTimeout(ctx, safeSQL, timeout)
	if err != nil {
		return nil, fmt.Errorf("查询执行失败：%v", err)
	}

	// Calculate execution time
	execTime := time.Since(startTime).Milliseconds()

	// Build query result
	result := agent.NewQueryResult(data, "", execTime, safeSQL)
	if queryContext != nil {
		result.PermissionFiltered = queryContext.PermissionFilter != ""
	}

	return result, nil
}

// GetTableSchema returns the schema for a table.
func (e *SQLExecutor) GetTableSchema(ctx context.Context, tableName string) (*erp.TableSchema, error) {
	return e.erpReader.GetTableSchema(ctx, tableName)
}

// GetTableList returns the list of available tables.
func (e *SQLExecutor) GetTableList(ctx context.Context) ([]string, error) {
	return e.erpReader.GetTableList(ctx)
}

// ValidateSQL validates SQL without executing.
func (e *SQLExecutor) ValidateSQL(sql string) error {
	return e.validator.Validate(sql)
}

// IsReadOnly checks if SQL is read-only.
func (e *SQLExecutor) IsReadOnly(sql string) bool {
	return e.validator.IsReadOnly(sql)
}

// ExecutionStats represents statistics of SQL execution.
type ExecutionStats struct {
	TotalQueries    int64 `json:"total_queries"`
	SuccessQueries  int64 `json:"success_queries"`
	FailedQueries   int64 `json:"failed_queries"`
	TotalTime       int64 `json:"total_time_ms"`
	AverageTime     int64 `json:"average_time_ms"`
	MaxTime         int64 `json:"max_time_ms"`
	MinTime         int64 `json:"min_time_ms"`
	PermissionFiltered int64 `json:"permission_filtered"`
}

// SQLExecutionError represents an error during SQL execution.
type SQLExecutionError struct {
	SQL         string `json:"sql"`
	ErrorType   string `json:"error_type"`
	Message     string `json:"message"`
	IsRetryable bool   `json:"is_retryable"`
}

// Error implements the error interface.
func (e *SQLExecutionError) Error() string {
	return fmt.Sprintf("SQL执行错误 [%s]: %s (SQL: %s)", e.ErrorType, e.Message, e.SQL)
}

// NewSQLExecutionError creates a new SQL execution error.
func NewSQLExecutionError(sql string, errorType string, message string, isRetryable bool) *SQLExecutionError {
	return &SQLExecutionError{
		SQL:         sql,
		ErrorType:   errorType,
		Message:     message,
		IsRetryable: isRetryable,
	}
}

// Error types
const (
	ErrorTypeValidation      = "validation"
	ErrorTypePermission      = "permission"
	ErrorTypeTimeout         = "timeout"
	ErrorTypeConnection      = "connection"
	ErrorTypeSyntax          = "syntax"
	ErrorTypeExecution       = "execution"
	ErrorTypeTooManyRows     = "too_many_rows"
)