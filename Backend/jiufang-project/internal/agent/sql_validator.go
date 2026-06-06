// Package agent implements the AI Agent for semantic understanding and SQL generation.
package agent

import (
	"fmt"
	"regexp"
	"strings"

	"jiufang/internal/model/agent"
)

// SQLValidator validates SQL for safety and correctness.
type SQLValidator struct {
	dangerousKeywords *regexp.Regexp
	dangerousPatterns []string
	allowedTables     map[string]bool
	maxResultRows     int
}

// NewSQLValidator creates a new SQL validator.
func NewSQLValidator() *SQLValidator {
	// Compile dangerous keywords pattern
	dangerousKeywords := regexp.MustCompile(`(?i)(DELETE|UPDATE|INSERT|DROP|ALTER|CREATE|TRUNCATE|GRANT|REVOKE|EXEC|EXECUTE|MERGE|CALL)`)

	// Additional dangerous patterns
	dangerousPatterns := []string{
		"--",        // SQL comment injection
		";",         // Multiple statement injection
		"/*",        // Block comment start
		"*/",        // Block comment end
		"UNION",     // UNION injection
		"INTO OUTFILE", // File operations
		"INTO DUMPFILE", // File operations
		"LOAD_FILE", // File operations
	}

	return &SQLValidator{
		dangerousKeywords: dangerousKeywords,
		dangerousPatterns: dangerousPatterns,
		allowedTables:     make(map[string]bool),
		maxResultRows:     10000,
	}
}

// Validate validates the SQL for safety.
func (v *SQLValidator) Validate(sql string) error {
	// Check if SQL is empty
	if strings.TrimSpace(sql) == "" {
		return agent.NewClarificationRequest("empty_sql", "SQL语句为空", "", "", nil)
	}

	// Check for dangerous keywords
	if v.dangerousKeywords.MatchString(sql) {
		return fmt.Errorf("SQL包含危险关键字，只允许SELECT查询")
	}

	// Check for dangerous patterns
	for _, pattern := range v.dangerousPatterns {
		if strings.Contains(strings.ToUpper(sql), strings.ToUpper(pattern)) {
			return fmt.Errorf("SQL包含危险模式：%s", pattern)
		}
	}

	// Check if SQL starts with SELECT
	trimmedSQL := strings.TrimSpace(strings.ToUpper(sql))
	if !strings.HasPrefix(trimmedSQL, "SELECT") {
		return fmt.Errorf("只允许SELECT查询语句")
	}

	// Check for allowed tables (if configured)
	if len(v.allowedTables) > 0 {
		tables := extractTablesFromSQL(sql)
		for _, table := range tables {
			if !v.allowedTables[table] {
				return fmt.Errorf("不允许查询表：%s", table)
			}
		}
	}

	// Check for LIMIT clause (recommended)
	if !strings.Contains(strings.ToUpper(sql), "LIMIT") {
		// Add warning but don't fail
		// In production, we might auto-add LIMIT
	}

	return nil
}

// IsReadOnly checks if the SQL is a read-only query.
func (v *SQLValidator) IsReadOnly(sql string) bool {
	trimmedSQL := strings.TrimSpace(strings.ToUpper(sql))
	if !strings.HasPrefix(trimmedSQL, "SELECT") {
		return false
	}

	// Check for dangerous keywords
	if v.dangerousKeywords.MatchString(sql) {
		return false
	}

	// Check for dangerous patterns
	for _, pattern := range v.dangerousPatterns {
		if strings.Contains(strings.ToUpper(sql), strings.ToUpper(pattern)) {
			return false
		}
	}

	return true
}

// ValidateWithPermission validates SQL with permission check.
func (v *SQLValidator) ValidateWithPermission(sql string, queryContext *agent.QueryContext) error {
	// First validate SQL safety
	if err := v.Validate(sql); err != nil {
		return err
	}

	// Check table permissions
	if len(queryContext.AllowedTables) > 0 {
		tables := extractTablesFromSQL(sql)
		for _, table := range tables {
			isAllowed := false
			for _, allowedTable := range queryContext.AllowedTables {
				if strings.ToLower(table) == strings.ToLower(allowedTable) {
					isAllowed = true
					break
				}
			}
			if !isAllowed {
				return fmt.Errorf("用户无权限查询表：%s", table)
			}
		}
	}

	// Check field permissions
	if len(queryContext.AllowedFields) > 0 {
		fields := extractFieldsFromSQL(sql)
		for table, allowedFields := range queryContext.AllowedFields {
			for _, field := range fields {
				isAllowed := false
				for _, allowedField := range allowedFields {
					if strings.ToLower(field) == strings.ToLower(allowedField) || field == "*" {
						isAllowed = true
						break
					}
				}
				if !isAllowed {
					return fmt.Errorf("用户无权限查询字段：%s.%s", table, field)
				}
			}
		}
	}

	return nil
}

// SetAllowedTables sets the list of allowed tables.
func (v *SQLValidator) SetAllowedTables(tables []string) {
	v.allowedTables = make(map[string]bool)
	for _, table := range tables {
		v.allowedTables[strings.ToLower(table)] = true
	}
}

// SetMaxResultRows sets the maximum result rows.
func (v *SQLValidator) SetMaxResultRows(maxRows int) {
	v.maxResultRows = maxRows
}

// GetMaxResultRows returns the maximum result rows.
func (v *SQLValidator) GetMaxResultRows() int {
	return v.maxResultRows
}

// AddSafetyLimit adds LIMIT clause to SQL if not present.
func (v *SQLValidator) AddSafetyLimit(sql string) string {
	if strings.Contains(strings.ToUpper(sql), "LIMIT") {
		return sql
	}

	// Find position to insert LIMIT
	insertPos := len(sql)
	for _, keyword := range []string{"ORDER BY", "GROUP BY", "HAVING"} {
		if idx := strings.Index(strings.ToUpper(sql), keyword); idx > 0 && idx < insertPos {
			insertPos = idx
		}
	}

	// Add LIMIT before ORDER BY/GROUP BY
	return sql[:insertPos] + fmt.Sprintf(" LIMIT %d ", v.maxResultRows) + sql[insertPos:]
}

// extractTablesFromSQL extracts table names from SQL.
func extractTablesFromSQL(sql string) []string {
	tables := []string{}

	// Simple extraction: look for FROM and JOIN keywords
	fromPattern := regexp.MustCompile(`FROM\s+([a-zA-Z_][a-zA-Z0-9_]*)`)
	joinPattern := regexp.MustCompile(`JOIN\s+([a-zA-Z_][a-zA-Z0-9_]*)`)

	fromMatches := fromPattern.FindAllStringSubmatch(sql, -1)
	for _, match := range fromMatches {
		if len(match) > 1 {
			tables = append(tables, strings.ToLower(match[1]))
		}
	}

	joinMatches := joinPattern.FindAllStringSubmatch(sql, -1)
	for _, match := range joinMatches {
		if len(match) > 1 {
			tables = append(tables, strings.ToLower(match[1]))
		}
	}

	return tables
}

// extractFieldsFromSQL extracts field names from SQL.
func extractFieldsFromSQL(sql string) []string {
	fields := []string{}

	// Extract fields between SELECT and FROM
	selectPattern := regexp.MustCompile(`SELECT\s+(.+?)\s+FROM`)
	matches := selectPattern.FindStringSubmatch(sql)
	if len(matches) > 1 {
		fieldList := matches[1]
		// Split by comma
		for _, field := range strings.Split(fieldList, ",") {
			field = strings.TrimSpace(field)
			// Remove aliases (AS keyword)
			if strings.Contains(strings.ToUpper(field), "AS") {
				field = strings.Split(field, " ")[0]
			}
			// Remove table prefix
			if strings.Contains(field, ".") {
				field = strings.Split(field, ".")[1]
			}
			fields = append(fields, strings.ToLower(field))
		}
	}

	return fields
}

// SQLValidationResult represents the result of SQL validation.
type SQLValidationResult struct {
	IsValid      bool     `json:"is_valid"`
	Errors       []string `json:"errors"`
	Warnings     []string `json:"warnings"`
	Tables       []string `json:"tables"`
	Fields       []string `json:"fields"`
	IsReadOnly   bool     `json:"is_read_only"`
	HasLimit     bool     `json:"has_limit"`
	SafeSQL      string   `json:"safe_sql"`
}

// ValidateDetailed performs detailed validation and returns result.
func (v *SQLValidator) ValidateDetailed(sql string) *SQLValidationResult {
	result := &SQLValidationResult{
		IsValid:    true,
		Errors:     []string{},
		Warnings:   []string{},
		Tables:     extractTablesFromSQL(sql),
		Fields:     extractFieldsFromSQL(sql),
		IsReadOnly: v.IsReadOnly(sql),
		HasLimit:   strings.Contains(strings.ToUpper(sql), "LIMIT"),
		SafeSQL:    sql,
	}

	// Check dangerous keywords
	if v.dangerousKeywords.MatchString(sql) {
		result.IsValid = false
		result.Errors = append(result.Errors, "SQL包含危险关键字")
	}

	// Check dangerous patterns
	for _, pattern := range v.dangerousPatterns {
		if strings.Contains(strings.ToUpper(sql), strings.ToUpper(pattern)) {
			result.IsValid = false
			result.Errors = append(result.Errors, fmt.Sprintf("SQL包含危险模式：%s", pattern))
		}
	}

	// Check SELECT prefix
	trimmedSQL := strings.TrimSpace(strings.ToUpper(sql))
	if !strings.HasPrefix(trimmedSQL, "SELECT") {
		result.IsValid = false
		result.Errors = append(result.Errors, "只允许SELECT查询语句")
	}

	// Check LIMIT
	if !result.HasLimit {
		result.Warnings = append(result.Warnings, "建议添加LIMIT限制结果数量")
		result.SafeSQL = v.AddSafetyLimit(sql)
	}

	// Check table permissions
	if len(v.allowedTables) > 0 {
		for _, table := range result.Tables {
			if !v.allowedTables[table] {
				result.IsValid = false
				result.Errors = append(result.Errors, fmt.Sprintf("不允许查询表：%s", table))
			}
		}
	}

	return result
}

// ValidateIntentMatch validates if SQL matches the intent.
func (v *SQLValidator) ValidateIntentMatch(sql string, intent *agent.Intent) error {
	sqlUpper := strings.ToUpper(sql)

	switch intent.Type {
	case agent.IntentTypeStatistics:
		// Statistics query should have aggregation functions
		if !strings.Contains(sqlUpper, "COUNT") &&
			!strings.Contains(sqlUpper, "SUM") &&
			!strings.Contains(sqlUpper, "AVG") &&
			!strings.Contains(sqlUpper, "MAX") &&
			!strings.Contains(sqlUpper, "MIN") {
			return fmt.Errorf("统计查询应包含聚合函数（COUNT/SUM/AVG/MAX/MIN）")
		}

	case agent.IntentTypeTrend:
		// Trend query should have GROUP BY date
		if !strings.Contains(sqlUpper, "GROUP BY") {
			return fmt.Errorf("趋势查询应包含GROUP BY日期")
		}

	case agent.IntentTypeComparison:
		// Comparison query should compare different periods
		// This is a simplified check
		if !strings.Contains(sqlUpper, "UNION") &&
			!strings.Contains(sqlUpper, "CASE") {
			return fmt.Errorf("对比查询应包含对比逻辑")
		}
	}

	return nil
}