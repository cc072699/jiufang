// Package agent implements the AI Agent for semantic understanding and SQL generation.
package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"

	agentModel "jiufang/internal/model/agent"
)

// SQLQueryTool is a tool for executing SQL queries.
type SQLQueryTool struct {
	executor *SQLExecutor
	logger   *zap.Logger
}

// NewSQLQueryTool creates a new SQL query tool.
func NewSQLQueryTool(executor *SQLExecutor, logger *zap.Logger) *SQLQueryTool {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &SQLQueryTool{
		executor: executor,
		logger:   logger,
	}
}

// Info returns the tool information.
func (t *SQLQueryTool) Info() *schema.ToolInfo {
	return &schema.ToolInfo{
		Name: "sql_query",
		Desc: "执行SQL查询并返回结果。只支持SELECT查询，自动添加安全限制。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"sql": {
				Type:     schema.String,
				Desc:     "要执行的SQL查询语句",
				Required: true,
			},
			"limit": {
				Type:     schema.Integer,
				Desc:     "返回结果的最大行数",
				Required: false,
			},
		}),
	}
}

// Run executes the tool.
func (t *SQLQueryTool) Run(ctx context.Context, params string) (string, error) {
	// Parse parameters
	var input struct {
		SQL   string `json:"sql"`
		Limit int    `json:"limit"`
	}

	if err := json.Unmarshal([]byte(params), &input); err != nil {
		return "", fmt.Errorf("参数解析失败：%v", err)
	}

	// Validate SQL
	if err := t.executor.ValidateSQL(input.SQL); err != nil {
		return "", fmt.Errorf("SQL校验失败：%v", err)
	}

	// Execute query
	var result *agentModel.QueryResult
	var err error

	if input.Limit > 0 {
		result, err = t.executor.ExecuteWithLimit(ctx, input.SQL, input.Limit)
	} else {
		result, err = t.executor.Execute(ctx, input.SQL)
	}

	if err != nil {
		return "", fmt.Errorf("查询执行失败：%v", err)
	}

	// Format result as JSON
	formatter := NewResultFormatter()
	return formatter.FormatForJSON(result), nil
}

// TableSchemaTool is a tool for getting table schema.
type TableSchemaTool struct {
	executor *SQLExecutor
	logger   *zap.Logger
}

// NewTableSchemaTool creates a new table schema tool.
func NewTableSchemaTool(executor *SQLExecutor, logger *zap.Logger) *TableSchemaTool {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &TableSchemaTool{
		executor: executor,
		logger:   logger,
	}
}

// Info returns the tool information.
func (t *TableSchemaTool) Info() *schema.ToolInfo {
	return &schema.ToolInfo{
		Name: "get_table_schema",
		Desc: "获取数据库表的schema信息，包括字段名、类型、注释等。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"table_name": {
				Type:     schema.String,
				Desc:     "表名",
				Required: true,
			},
		}),
	}
}

// Run executes the tool.
func (t *TableSchemaTool) Run(ctx context.Context, params string) (string, error) {
	// Parse parameters
	var input struct {
		TableName string `json:"table_name"`
	}

	if err := json.Unmarshal([]byte(params), &input); err != nil {
		return "", fmt.Errorf("参数解析失败：%v", err)
	}

	// Get table schema
	schemaInfo, err := t.executor.GetTableSchema(ctx, input.TableName)
	if err != nil {
		return "", fmt.Errorf("获取表schema失败：%v", err)
	}

	// Format as JSON
	return fmt.Sprintf("表名：%s\n字段数：%d\n字段信息：%v", input.TableName, len(schemaInfo.Columns), schemaInfo), nil
}

// TableListTool is a tool for listing all tables.
type TableListTool struct {
	executor *SQLExecutor
	logger   *zap.Logger
}

// NewTableListTool creates a new table list tool.
func NewTableListTool(executor *SQLExecutor, logger *zap.Logger) *TableListTool {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &TableListTool{
		executor: executor,
		logger:   logger,
	}
}

// Info returns the tool information.
func (t *TableListTool) Info() *schema.ToolInfo {
	return &schema.ToolInfo{
		Name: "list_tables",
		Desc: "列出数据库中所有可用的表名。",
	}
}

// Run executes the tool.
func (t *TableListTool) Run(ctx context.Context, params string) (string, error) {
	// Get table list from executor
	tables, err := t.executor.GetTableList(ctx)
	if err != nil {
		return "", fmt.Errorf("获取表列表失败：%v", err)
	}

	result := "可用表列表：\n"
	for i, table := range tables {
		result += fmt.Sprintf("%d. %s\n", i+1, table)
	}

	return result, nil
}

// ValidateSQLTool is a tool for validating SQL.
type ValidateSQLTool struct {
	validator *SQLValidator
	logger    *zap.Logger
}

// NewValidateSQLTool creates a new SQL validation tool.
func NewValidateSQLTool(validator *SQLValidator, logger *zap.Logger) *ValidateSQLTool {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ValidateSQLTool{
		validator: validator,
		logger:    logger,
	}
}

// Info returns the tool information.
func (t *ValidateSQLTool) Info() *schema.ToolInfo {
	return &schema.ToolInfo{
		Name: "validate_sql",
		Desc: "验证SQL语句的安全性和正确性。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"sql": {
				Type:     schema.String,
				Desc:     "要验证的SQL语句",
				Required: true,
			},
		}),
	}
}

// Run executes the tool.
func (t *ValidateSQLTool) Run(ctx context.Context, params string) (string, error) {
	// Parse parameters
	var input struct {
		SQL string `json:"sql"`
	}

	if err := json.Unmarshal([]byte(params), &input); err != nil {
		return "", fmt.Errorf("参数解析失败：%v", err)
	}

	// Validate SQL
	err := t.validator.Validate(input.SQL)
	if err != nil {
		return fmt.Sprintf("SQL验证失败：%v", err), nil
	}

	return "SQL验证通过，语句安全可执行", nil
}

// ToolRegistry manages all available tools.
type ToolRegistry struct {
	tools map[string]interface{}
}

// NewToolRegistry creates a new tool registry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]interface{}),
	}
}

// Register registers a tool.
func (r *ToolRegistry) Register(name string, tool interface{}) {
	r.tools[name] = tool
}

// Get retrieves a tool by name.
func (r *ToolRegistry) Get(name string) (interface{}, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// GetAll returns all registered tools.
func (r *ToolRegistry) GetAll() map[string]interface{} {
	return r.tools
}

// GetToolInfos returns information about all registered tools.
func (r *ToolRegistry) GetToolInfos() []*schema.ToolInfo {
	infos := make([]*schema.ToolInfo, 0, len(r.tools))
	for _, tool := range r.tools {
		if infoer, ok := tool.(interface{ Info() *schema.ToolInfo }); ok {
			infos = append(infos, infoer.Info())
		}
	}
	return infos
}
