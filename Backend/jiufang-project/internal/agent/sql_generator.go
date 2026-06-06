// Package agent implements the AI Agent for semantic understanding and SQL generation.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"jiufang/internal/infrastructure/llm"
	"jiufang/internal/model/agent"
)

// SQLGenerator generates SQL from intent and entities.
type SQLGenerator struct {
	llmClient llm.LLMClientInterface
}

// NewSQLGenerator creates a new SQL generator.
func NewSQLGenerator(llmClient llm.LLMClientInterface) *SQLGenerator {
	return &SQLGenerator{
		llmClient: llmClient,
	}
}

// Generate generates SQL from intent and entities.
func (g *SQLGenerator) Generate(ctx context.Context, intent *agent.Intent, entities []agent.Entity) (string, error) {
	// Build prompt for SQL generation
	prompt := buildSQLPrompt(intent, entities)

	// Call LLM
	response, err := g.llmClient.Generate(ctx, prompt)
	if err != nil {
		// Fallback to template-based SQL generation
		return g.generateWithTemplate(intent, entities), nil
	}

	// Parse response
	sql, err := parseSQLResponse(response)
	if err != nil {
		return g.generateWithTemplate(intent, entities), nil
	}

	return sql, nil
}

// GenerateWithSchema generates SQL with schema information.
func (g *SQLGenerator) GenerateWithSchema(ctx context.Context, intent *agent.Intent, entities []agent.Entity, tableSchema string) (string, error) {
	// Build prompt with schema
	prompt := buildSQLPromptWithSchema(intent, entities, tableSchema)

	// Call LLM
	response, err := g.llmClient.Generate(ctx, prompt)
	if err != nil {
		return g.generateWithTemplate(intent, entities), nil
	}

	// Parse response
	sql, err := parseSQLResponse(response)
	if err != nil {
		return g.generateWithTemplate(intent, entities), nil
	}

	return sql, nil
}

// GenerateWithPermissionFilter generates SQL with permission filter.
func (g *SQLGenerator) GenerateWithPermissionFilter(ctx context.Context, intent *agent.Intent, entities []agent.Entity, permissionFilter string) (string, error) {
	// Generate base SQL
	sql, err := g.Generate(ctx, intent, entities)
	if err != nil {
		return "", err
	}

	// Add permission filter
	if permissionFilter != "" {
		sql = addPermissionFilter(sql, permissionFilter)
	}

	return sql, nil
}

// generateWithTemplate generates SQL using templates (fallback mode).
func (g *SQLGenerator) generateWithTemplate(intent *agent.Intent, entities []agent.Entity) string {
	// Get table name from document type entity
	tableName := "purchase_orders" // Default table
	for _, e := range entities {
		if e.Type == agent.EntityTypeDocumentType {
			tableName = mapDocumentTypeToTable(e.Value)
			break
		}
	}

	// Build SQL based on intent type
	switch intent.Type {
	case agent.IntentTypeStatistics:
		return buildStatisticsSQL(tableName, entities)
	case agent.IntentTypeDetail:
		return buildDetailSQL(tableName, entities)
	case agent.IntentTypeTrend:
		return buildTrendSQL(tableName, entities)
	case agent.IntentTypeComparison:
		return buildComparisonSQL(tableName, entities)
	default:
		return buildDetailSQL(tableName, entities)
	}
}

// buildStatisticsSQL builds SQL for statistics query.
func buildStatisticsSQL(tableName string, entities []agent.Entity) string {
	sql := fmt.Sprintf("SELECT COUNT(*) as total_count, SUM(amount) as total_amount FROM %s", tableName)

	// Add WHERE clause from entities
	whereClause := buildWhereClause(entities)
	if whereClause != "" {
		sql += " WHERE " + whereClause
	}

	return sql
}

// buildDetailSQL builds SQL for detail query.
func buildDetailSQL(tableName string, entities []agent.Entity) string {
	sql := fmt.Sprintf("SELECT * FROM %s", tableName)

	// Add WHERE clause from entities
	whereClause := buildWhereClause(entities)
	if whereClause != "" {
		sql += " WHERE " + whereClause
	}

	// Add ORDER BY
	sql += " ORDER BY created_at DESC"

	// Add LIMIT
	sql += " LIMIT 100"

	return sql
}

// buildTrendSQL builds SQL for trend query.
func buildTrendSQL(tableName string, entities []agent.Entity) string {
	sql := fmt.Sprintf("SELECT DATE(created_at) as date, SUM(amount) as daily_amount FROM %s", tableName)

	// Add WHERE clause from entities
	whereClause := buildWhereClause(entities)
	if whereClause != "" {
		sql += " WHERE " + whereClause
	}

	// Add GROUP BY
	sql += " GROUP BY DATE(created_at)"

	// Add ORDER BY
	sql += " ORDER BY date ASC"

	return sql
}

// buildComparisonSQL builds SQL for comparison query.
func buildComparisonSQL(tableName string, entities []agent.Entity) string {
	// For comparison, we need two time ranges
	// This is a simplified implementation
	sql := fmt.Sprintf("SELECT period, SUM(amount) as amount FROM (SELECT 'period1' as period, amount FROM %s WHERE created_at >= '2024-01-01' AND created_at < '2024-04-01' UNION ALL SELECT 'period2' as period, amount FROM %s WHERE created_at >= '2024-04-01' AND created_at < '2024-07-01') as comparison_data GROUP BY period", tableName, tableName)

	return sql
}

// buildWhereClause builds WHERE clause from entities.
func buildWhereClause(entities []agent.Entity) string {
	conditions := []string{}

	for _, e := range entities {
		switch e.Type {
		case agent.EntityTypeTimeRange:
			// Parse time range value
			if strings.Contains(e.Value, " to ") {
				parts := strings.Split(e.Value, " to ")
				if len(parts) == 2 {
					conditions = append(conditions, fmt.Sprintf("created_at >= '%s' AND created_at <= '%s'", parts[0], parts[1]))
				}
			} else {
				conditions = append(conditions, fmt.Sprintf("created_at = '%s'", e.Value))
			}

		case agent.EntityTypeAmount:
			// Parse amount condition
			if strings.HasPrefix(e.Value, ">") {
				conditions = append(conditions, fmt.Sprintf("amount %s", e.Value))
			} else if strings.HasPrefix(e.Value, "<") {
				conditions = append(conditions, fmt.Sprintf("amount %s", e.Value))
			} else if strings.HasPrefix(e.Value, "=") {
				conditions = append(conditions, fmt.Sprintf("amount %s", e.Value))
			} else if strings.HasPrefix(e.Value, "between") {
				conditions = append(conditions, fmt.Sprintf("amount %s", e.Value))
			}

		case agent.EntityTypeSupplier:
			conditions = append(conditions, fmt.Sprintf("supplier_name LIKE '%s'", e.Value))

		case agent.EntityTypeStatus:
			conditions = append(conditions, fmt.Sprintf("status = '%s'", e.Value))

		case agent.EntityTypeDepartment:
			conditions = append(conditions, fmt.Sprintf("department = '%s'", e.Value))

		case agent.EntityTypeCustomer:
			conditions = append(conditions, fmt.Sprintf("customer_name LIKE '%s'", e.Value))
		}
	}

	if len(conditions) == 0 {
		return ""
	}

	return strings.Join(conditions, " AND ")
}

// mapDocumentTypeToTable maps document type to table name.
func mapDocumentTypeToTable(docType string) string {
	tableMap := map[string]string{
		"purchase_order": "purchase_orders",
		"purchase":       "purchase_orders",
		"sales_order":    "sales_orders",
		"sales":          "sales_orders",
		"payment":        "payments",
		"receipt":        "receipts",
		"inbound":        "inbound_orders",
		"outbound":       "outbound_orders",
		"invoice":        "invoices",
		"expense":        "expenses",
	}

	if table, ok := tableMap[docType]; ok {
		return table
	}
	return "purchase_orders" // Default
}

// addPermissionFilter adds permission filter to SQL.
func addPermissionFilter(sql string, permissionFilter string) string {
	// Check if SQL already has WHERE clause
	if strings.Contains(strings.ToUpper(sql), "WHERE") {
		// Add to existing WHERE clause
		return strings.Replace(sql, "WHERE", "WHERE "+permissionFilter+" AND", 1)
	}
	// Add new WHERE clause
	// Find position before ORDER BY, GROUP BY, LIMIT
	insertPos := len(sql)
	for _, keyword := range []string{"ORDER BY", "GROUP BY", "LIMIT"} {
		if idx := strings.Index(strings.ToUpper(sql), keyword); idx > 0 && idx < insertPos {
			insertPos = idx
		}
	}
	return sql[:insertPos] + " WHERE " + permissionFilter + " " + sql[insertPos:]
}

// buildSQLPrompt builds the prompt for SQL generation.
func buildSQLPrompt(intent *agent.Intent, entities []agent.Entity) string {
	entityStr := ""
	for _, e := range entities {
		entityStr += fmt.Sprintf("- %s: %s (原始: %s)\n", e.Type, e.Value, e.RawText)
	}

	return fmt.Sprintf(`请根据以下意图和实体信息生成SQL查询语句。

意图类型：%s
意图描述：%s

实体信息：
%s

数据库表结构：
- purchase_orders: 采购单表（id, supplier_name, amount, status, created_at, department）
- sales_orders: 销售单表（id, customer_name, amount, status, created_at, department）
- payments: 付款单表（id, supplier_name, amount, status, created_at, purchase_order_id）
- receipts: 收款单表（id, customer_name, amount, status, created_at, sales_order_id）

请生成符合以下要求的SQL：
1. 只生成SELECT语句，不要包含DELETE、UPDATE、INSERT等操作
2. 根据意图类型选择合适的聚合函数（统计查询用COUNT/SUM，明细查询不用聚合）
3. 根据实体信息添加WHERE条件
4. 趋势查询需要GROUP BY日期
5. 对比查询需要对比不同时间段的数据

请返回以下JSON格式：
{
  "sql": "生成的SQL语句",
  "understanding": "对用户意图的理解摘要"
}

只返回JSON，不要其他内容。`, intent.Type, intent.Description, entityStr)
}

// buildSQLPromptWithSchema builds the prompt with table schema.
func buildSQLPromptWithSchema(intent *agent.Intent, entities []agent.Entity, tableSchema string) string {
	entityStr := ""
	for _, e := range entities {
		entityStr += fmt.Sprintf("- %s: %s (原始: %s)\n", e.Type, e.Value, e.RawText)
	}

	return fmt.Sprintf(`请根据以下意图、实体信息和表结构生成SQL查询语句。

意图类型：%s
意图描述：%s

实体信息：
%s

表结构：
%s

请生成符合以下要求的SQL：
1. 只生成SELECT语句
2. 根据意图类型选择合适的查询方式
3. 根据实体信息添加WHERE条件
4. 确保字段名与表结构一致

请返回以下JSON格式：
{
  "sql": "生成的SQL语句",
  "understanding": "对用户意图的理解摘要"
}

只返回JSON，不要其他内容。`, intent.Type, intent.Description, entityStr, tableSchema)
}

// parseSQLResponse parses the LLM response for SQL.
func parseSQLResponse(response string) (string, error) {
	response = cleanJSONResponse(response)

	var result struct {
		SQL         string `json:"sql"`
		Understanding string `json:"understanding"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return "", err
	}

	return result.SQL, nil
}

// ParseSQLResponseWithUnderstanding parses the LLM response for SQL and understanding.
func ParseSQLResponseWithUnderstanding(response string) (string, string, error) {
	response = cleanJSONResponse(response)

	var result struct {
		SQL         string `json:"sql"`
		Understanding string `json:"understanding"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return "", "", err
	}

	return result.SQL, result.Understanding, nil
}

// SQLTemplate provides template-based SQL generation.
type SQLTemplate struct {
	templates map[agent.IntentType]string
}

// NewSQLTemplate creates a new SQL template.
func NewSQLTemplate() *SQLTemplate {
	return &SQLTemplate{
		templates: map[agent.IntentType]string{
			agent.IntentTypeStatistics: "SELECT COUNT(*) as total_count, SUM(amount) as total_amount FROM {table} WHERE {conditions}",
			agent.IntentTypeDetail:     "SELECT * FROM {table} WHERE {conditions} ORDER BY created_at DESC LIMIT 100",
			agent.IntentTypeTrend:      "SELECT DATE(created_at) as date, SUM(amount) as daily_amount FROM {table} WHERE {conditions} GROUP BY DATE(created_at) ORDER BY date",
			agent.IntentTypeComparison: "SELECT period, SUM(amount) as amount FROM {comparison_query} GROUP BY period",
		},
	}
}

// GetTemplate returns the template for an intent type.
func (t *SQLTemplate) GetTemplate(intentType agent.IntentType) string {
	if template, ok := t.templates[intentType]; ok {
		return template
	}
	return t.templates[agent.IntentTypeDetail]
}