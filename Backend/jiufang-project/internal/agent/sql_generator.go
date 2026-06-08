// Package agent implements the AI Agent for semantic understanding and SQL generation.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"jiufang/internal/infrastructure/erp"
	"jiufang/internal/infrastructure/llm"
	"jiufang/internal/model/agent"
)

// SQLGenerator generates SQL from intent and entities.
type SQLGenerator struct {
	llmClient llm.LLMClientInterface
	erpReader erp.ERPReaderInterface
}

// NewSQLGenerator creates a new SQL generator.
func NewSQLGenerator(llmClient llm.LLMClientInterface, erpReader erp.ERPReaderInterface) *SQLGenerator {
	return &SQLGenerator{
		llmClient: llmClient,
		erpReader: erpReader,
	}
}

// Generate generates SQL from intent and entities.
func (g *SQLGenerator) Generate(ctx context.Context, intent *agent.Intent, entities []agent.Entity) (string, error) {
	// Build dynamic schema from ERP database
	schema := g.buildDynamicSchema(ctx)

	// Build prompt for SQL generation
	prompt := buildSQLPrompt(intent, entities, schema)

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

// buildDynamicSchema builds a schema description string from the ERP database.
func (g *SQLGenerator) buildDynamicSchema(ctx context.Context) string {
	if g.erpReader == nil {
		return "- (ERP数据库未连接，使用默认表结构)"
	}

	tables, err := g.erpReader.GetTableList(ctx)
	if err != nil || len(tables) == 0 {
		return "- (无法获取表列表)"
	}

	var sb strings.Builder
	for _, tableName := range tables {
		schema, err := g.erpReader.GetTableSchema(ctx, tableName)
		if err != nil {
			continue
		}
		cols := make([]string, 0, len(schema.Columns))
		for _, col := range schema.Columns {
			comment := col.Comment
			if comment != "" {
				cols = append(cols, col.Name+" /*"+comment+"*/")
			} else {
				cols = append(cols, col.Name)
			}
		}
		sb.WriteString(fmt.Sprintf("- %s: %s\n", tableName, strings.Join(cols, ", ")))
	}

	return sb.String()
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
	// Determine the amount column based on table
	amountCol := "total_amount"
	joinClause := ""

	switch tableName {
	case "purchase_orders":
		joinClause = " JOIN suppliers ON purchase_orders.supplier_id = suppliers.id"
	case "sales_orders":
		joinClause = " JOIN customers ON sales_orders.customer_id = customers.id"
	}

	sql := fmt.Sprintf("SELECT COUNT(*) as total_count, SUM(%s.%s) as total_amount FROM %s%s", tableName, amountCol, tableName, joinClause)

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
	amountCol := "total_amount"
	sql := fmt.Sprintf("SELECT DATE(created_at) as date, SUM(%s) as daily_amount FROM %s", amountCol, tableName)

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
	amountCol := "total_amount"
	now := time.Now()
	year, month, _ := now.Date()
	// Current quarter start
	cqStart := time.Date(year, ((month-1)/3)*3+1, 1, 0, 0, 0, 0, time.Local)
	// Previous quarter start
	pqStart := cqStart.AddDate(0, -3, 0)
	// Current quarter end (next quarter start)
	cqEnd := cqStart.AddDate(0, 3, 0)
	// Previous quarter end
	pqEnd := cqStart

	sql := fmt.Sprintf(
		"SELECT period, SUM(%s) as amount FROM "+
			"(SELECT 'period1' as period, %s FROM %s WHERE created_at >= '%s' AND created_at < '%s' "+
			"UNION ALL "+
			"SELECT 'period2' as period, %s FROM %s WHERE created_at >= '%s' AND created_at < '%s') as comparison_data "+
			"GROUP BY period",
		amountCol, amountCol, tableName, pqStart.Format("2006-01-02"), pqEnd.Format("2006-01-02"),
		amountCol, tableName, cqStart.Format("2006-01-02"), cqEnd.Format("2006-01-02"),
	)

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
			// Parse amount condition (use total_amount for main order tables)
			if strings.HasPrefix(e.Value, ">") {
				conditions = append(conditions, fmt.Sprintf("total_amount %s", e.Value))
			} else if strings.HasPrefix(e.Value, "<") {
				conditions = append(conditions, fmt.Sprintf("total_amount %s", e.Value))
			} else if strings.HasPrefix(e.Value, "=") {
				conditions = append(conditions, fmt.Sprintf("total_amount %s", e.Value))
			} else if strings.HasPrefix(e.Value, "between") {
				conditions = append(conditions, fmt.Sprintf("total_amount %s", e.Value))
			}

		case agent.EntityTypeSupplier:
			conditions = append(conditions, fmt.Sprintf("suppliers.supplier_name LIKE '%%%s%%'", e.Value))

		case agent.EntityTypeStatus:
			conditions = append(conditions, fmt.Sprintf("status = '%s'", e.Value))

		case agent.EntityTypeDepartment:
			conditions = append(conditions, fmt.Sprintf("purchaser LIKE '%%%s%%'", e.Value))

		case agent.EntityTypeCustomer:
			conditions = append(conditions, fmt.Sprintf("customers.customer_name LIKE '%%%s%%'", e.Value))
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
		"inbound":        "inbound_orders",
		"product":        "products",
		"inventory":      "inventory",
		"production":     "production_tasks",
		"quality":        "quality_inspections",
		"supplier":       "suppliers",
		"customer":       "customers",
	}

	if table, ok := tableMap[docType]; ok {
		return table
	}
	return "purchase_orders" // Default
}

// addPermissionFilter adds permission filter to SQL.
func addPermissionFilter(sql string, permissionFilter string) string {
	// Strip trailing semicolons and whitespace
	sql = strings.TrimRight(sql, "; \t\n\r")
	upper := strings.ToUpper(sql)

	// Find top-level WHERE (skip subqueries by tracking parenthesis depth)
	whereIdx := -1
	depth := 0
	for i := 0; i < len(upper)-5; i++ {
		switch upper[i] {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		}
		if depth == 0 && upper[i:i+5] == "WHERE" {
			// Make sure it's a word boundary (not part of another word)
			if (i == 0 || upper[i-1] == ' ' || upper[i-1] == ')') &&
				(i+5 >= len(upper) || upper[i+5] == ' ' || upper[i+5] == '(') {
				whereIdx = i
				break
			}
		}
	}

	if whereIdx >= 0 {
		return sql[:whereIdx] + "WHERE " + permissionFilter + " AND " + sql[whereIdx+6:]
	}

	// No WHERE found — insert before ORDER BY / GROUP BY / LIMIT
	insertPos := len(sql)
	for _, kw := range []string{"ORDER BY", "GROUP BY", "LIMIT"} {
		idx := strings.Index(upper, kw)
		if idx > 0 && idx < insertPos {
			insertPos = idx
		}
	}
	return strings.TrimRight(sql[:insertPos], " ") + " WHERE " + permissionFilter + " " + sql[insertPos:]
}

// buildSQLPrompt builds the prompt for SQL generation.
func buildSQLPrompt(intent *agent.Intent, entities []agent.Entity, schema string) string {
	entityStr := ""
	for _, e := range entities {
		entityStr += fmt.Sprintf("- %s: %s (原始: %s)\n", e.Type, e.Value, e.RawText)
	}

	return fmt.Sprintf(`请根据以下意图和实体信息生成MySQL查询语句。

意图类型：%s
意图描述：%s

实体信息：
%s

数据库表结构（MySQL）：
%s

请生成符合以下要求的SQL：
1. 只生成SELECT语句，不要包含DELETE、UPDATE、INSERT等操作
2. 根据意图类型选择合适的聚合函数（统计查询用COUNT/SUM，明细查询不用聚合）
3. 根据实体信息添加WHERE条件，注意使用正确的字段名
4. 需要关联供应商名称时，JOIN suppliers表：purchase_orders JOIN suppliers ON purchase_orders.supplier_id = suppliers.id
5. 需要关联客户名称时，JOIN customers表：sales_orders JOIN customers ON sales_orders.customer_id = customers.id
6. 需要关联产品名称时，JOIN products表
7. 趋势查询需要GROUP BY日期
8. 对比查询需要对比不同时间段的数据

请返回以下JSON格式：
{
  "sql": "生成的SQL语句",
  "understanding": "对用户意图的理解摘要"
}

只返回JSON，不要其他内容。`, intent.Type, intent.Description, entityStr, schema)
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
			agent.IntentTypeStatistics: "SELECT COUNT(*) as total_count, SUM(total_amount) as total_amount FROM {table} WHERE {conditions}",
			agent.IntentTypeDetail:     "SELECT * FROM {table} WHERE {conditions} ORDER BY created_at DESC LIMIT 100",
			agent.IntentTypeTrend:      "SELECT DATE(created_at) as date, SUM(total_amount) as daily_amount FROM {table} WHERE {conditions} GROUP BY DATE(created_at) ORDER BY date",
			agent.IntentTypeComparison: "SELECT period, SUM(total_amount) as amount FROM {comparison_query} GROUP BY period",
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