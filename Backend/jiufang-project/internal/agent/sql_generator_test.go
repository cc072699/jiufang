// Package agent implements the AI Agent for semantic understanding and SQL generation.
// This file implements unit tests for SQLGenerator.
// Author: AI Assistant
// Date: 2026-06-03
// Tested Object: SQLGenerator
// Function: SQL generation from intent and entities

package agent

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"jiufang/internal/mocks"
	"jiufang/internal/model/agent"
)

// TestSQLGenerator_Generate tests Generate method
func TestSQLGenerator_Generate(t *testing.T) {
	tests := []struct {
		name          string
		intent        *agent.Intent
		entities      []agent.Entity
		mockLLM       func(ctrl *gomock.Controller) *mocks.MockLLMClient
		wantSQL       string
		wantErr       bool
	}{
		{
			name: "TC-SG-01: Generate SQL successfully with LLM",
			intent: &agent.Intent{
				Type:        agent.IntentTypeStatistics,
				Confidence:  0.85,
				Description: "统计查询",
			},
			entities: []agent.Entity{
				{Type: agent.EntityTypeTimeRange, Value: "2024-05-01 to 2024-05-31", RawText: "上个月"},
			},
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				response := `{"sql":"SELECT COUNT(*) as total_count, SUM(amount) as total_amount FROM purchase_orders WHERE created_at >= '2024-05-01' AND created_at <= '2024-05-31'","understanding":"查询上个月采购总额"}`
				mock.EXPECT().Generate(gomock.Any(), gomock.Any()).Return(response, nil)
				return mock
			},
			wantSQL: "SELECT COUNT(*) as total_count, SUM(amount) as total_amount FROM purchase_orders WHERE created_at >= '2024-05-01' AND created_at <= '2024-05-31'",
			wantErr: false,
		},
		{
			name: "TC-SG-02: LLM failure - fallback to template",
			intent: &agent.Intent{
				Type:        agent.IntentTypeStatistics,
				Confidence:  0.6,
				Description: "统计查询",
			},
			entities: []agent.Entity{
				{Type: agent.EntityTypeDocumentType, Value: "purchase_order", RawText: "采购单"},
			},
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				mock.EXPECT().Generate(gomock.Any(), gomock.Any()).Return("", errors.New("LLM error"))
				return mock
			},
			wantSQL: "SELECT COUNT(*) as total_count, SUM(amount) as total_amount FROM purchase_orders",
			wantErr: false,
		},
		{
			name: "TC-SG-03: LLM returns invalid JSON - fallback",
			intent: &agent.Intent{
				Type:        agent.IntentTypeDetail,
				Confidence:  0.6,
				Description: "明细查询",
			},
			entities: []agent.Entity{},
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				mock.EXPECT().Generate(gomock.Any(), gomock.Any()).Return("invalid json", nil)
				return mock
			},
			wantSQL: "SELECT * FROM purchase_orders ORDER BY created_at DESC LIMIT 100",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLLM := tt.mockLLM(ctrl)
			generator := NewSQLGenerator(mockLLM, nil)

			// Act
			result, err := generator.Generate(context.Background(), tt.intent, tt.entities)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, result, tt.wantSQL[:20]) // Check first 20 chars
			}
		})
	}
}

// TestSQLGenerator_GenerateWithSchema tests GenerateWithSchema method
func TestSQLGenerator_GenerateWithSchema(t *testing.T) {
	tests := []struct {
		name          string
		intent        *agent.Intent
		entities      []agent.Entity
		tableSchema   string
		mockLLM       func(ctrl *gomock.Controller) *mocks.MockLLMClient
		wantSQL       string
		wantErr       bool
	}{
		{
			name: "TC-SG-04: GenerateWithSchema successfully",
			intent: &agent.Intent{
				Type:        agent.IntentTypeDetail,
				Confidence:  0.85,
				Description: "明细查询",
			},
			entities: []agent.Entity{
				{Type: agent.EntityTypeSupplier, Value: "A公司", RawText: "A公司"},
			},
			tableSchema: "purchase_orders (id, supplier_name, amount, status, created_at)",
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				response := `{"sql":"SELECT id, supplier_name, amount, status FROM purchase_orders WHERE supplier_name LIKE 'A公司'","understanding":"查询A公司的采购单"}`
				mock.EXPECT().Generate(gomock.Any(), gomock.Any()).Return(response, nil)
				return mock
			},
			wantSQL: "SELECT",
			wantErr: false,
		},
		{
			name: "TC-SG-05: GenerateWithSchema LLM failure - fallback",
			intent: &agent.Intent{
				Type:        agent.IntentTypeDetail,
				Confidence:  0.6,
				Description: "明细查询",
			},
			entities: []agent.Entity{},
			tableSchema: "",
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				mock.EXPECT().Generate(gomock.Any(), gomock.Any()).Return("", errors.New("LLM error"))
				return mock
			},
			wantSQL: "SELECT * FROM purchase_orders",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLLM := tt.mockLLM(ctrl)
			generator := NewSQLGenerator(mockLLM, nil)

			// Act
			result, err := generator.GenerateWithSchema(context.Background(), tt.intent, tt.entities, tt.tableSchema)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, result, tt.wantSQL)
			}
		})
	}
}

// TestSQLGenerator_GenerateWithPermissionFilter tests GenerateWithPermissionFilter method
func TestSQLGenerator_GenerateWithPermissionFilter(t *testing.T) {
	tests := []struct {
		name              string
		intent            *agent.Intent
		entities          []agent.Entity
		permissionFilter  string
		mockLLM           func(ctrl *gomock.Controller) *mocks.MockLLMClient
		wantSQLContains   string
		wantErr           bool
	}{
		{
			name: "TC-SG-06: GenerateWithPermissionFilter - add filter",
			intent: &agent.Intent{
				Type:        agent.IntentTypeStatistics,
				Confidence:  0.85,
				Description: "统计查询",
			},
			entities: []agent.Entity{},
			permissionFilter: "department = '销售部'",
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				mock.EXPECT().Generate(gomock.Any(), gomock.Any()).Return("", errors.New("LLM error"))
				return mock
			},
			wantSQLContains: "department = '销售部'",
			wantErr: false,
		},
		{
			name: "TC-SG-07: GenerateWithPermissionFilter - empty filter",
			intent: &agent.Intent{
				Type:        agent.IntentTypeDetail,
				Confidence:  0.85,
				Description: "明细查询",
			},
			entities: []agent.Entity{},
			permissionFilter: "",
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				mock.EXPECT().Generate(gomock.Any(), gomock.Any()).Return("", errors.New("LLM error"))
				return mock
			},
			wantSQLContains: "SELECT",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLLM := tt.mockLLM(ctrl)
			generator := NewSQLGenerator(mockLLM, nil)

			// Act
			result, err := generator.GenerateWithPermissionFilter(context.Background(), tt.intent, tt.entities, tt.permissionFilter)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, result, tt.wantSQLContains)
			}
		})
	}
}

// TestBuildStatisticsSQL tests buildStatisticsSQL function
func TestBuildStatisticsSQL(t *testing.T) {
	tests := []struct {
		name         string
		tableName    string
		entities     []agent.Entity
		wantSQL      string
	}{
		{
			name:      "TC-SG-08: Build statistics SQL with entities",
			tableName: "purchase_orders",
			entities: []agent.Entity{
				{Type: agent.EntityTypeTimeRange, Value: "2024-05-01 to 2024-05-31"},
				{Type: agent.EntityTypeStatus, Value: "completed"},
			},
			wantSQL: "SELECT COUNT(*) as total_count, SUM(amount) as total_amount FROM purchase_orders WHERE",
		},
		{
			name:      "TC-SG-09: Build statistics SQL without entities",
			tableName: "purchase_orders",
			entities:  []agent.Entity{},
			wantSQL:   "SELECT COUNT(*) as total_count, SUM(amount) as total_amount FROM purchase_orders",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := buildStatisticsSQL(tt.tableName, tt.entities)

			// Assert
			assert.Contains(t, result, tt.wantSQL)
		})
	}
}

// TestBuildDetailSQL tests buildDetailSQL function
func TestBuildDetailSQL(t *testing.T) {
	tests := []struct {
		name         string
		tableName    string
		entities     []agent.Entity
		wantSQL      string
	}{
		{
			name:      "TC-SG-10: Build detail SQL with entities",
			tableName: "purchase_orders",
			entities: []agent.Entity{
				{Type: agent.EntityTypeSupplier, Value: "A公司"},
			},
			wantSQL: "SELECT * FROM purchase_orders WHERE",
		},
		{
			name:      "TC-SG-11: Build detail SQL without entities",
			tableName: "purchase_orders",
			entities:  []agent.Entity{},
			wantSQL:   "SELECT * FROM purchase_orders ORDER BY created_at DESC LIMIT 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := buildDetailSQL(tt.tableName, tt.entities)

			// Assert
			assert.Contains(t, result, tt.wantSQL)
		})
	}
}

// TestBuildTrendSQL tests buildTrendSQL function
func TestBuildTrendSQL(t *testing.T) {
	tests := []struct {
		name         string
		tableName    string
		entities     []agent.Entity
		wantSQL      string
	}{
		{
			name:      "TC-SG-12: Build trend SQL with entities",
			tableName: "purchase_orders",
			entities: []agent.Entity{
				{Type: agent.EntityTypeTimeRange, Value: "2024-01-01 to 2024-12-31"},
			},
			wantSQL: "SELECT DATE(created_at) as date, SUM(amount) as daily_amount FROM purchase_orders WHERE",
		},
		{
			name:      "TC-SG-13: Build trend SQL without entities",
			tableName: "purchase_orders",
			entities:  []agent.Entity{},
			wantSQL:   "SELECT DATE(created_at) as date, SUM(amount) as daily_amount FROM purchase_orders GROUP BY DATE(created_at)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := buildTrendSQL(tt.tableName, tt.entities)

			// Assert
			assert.Contains(t, result, tt.wantSQL)
		})
	}
}

// TestBuildComparisonSQL tests buildComparisonSQL function
func TestBuildComparisonSQL(t *testing.T) {
	tests := []struct {
		name         string
		tableName    string
		entities     []agent.Entity
		wantSQL      string
	}{
		{
			name:      "TC-SG-14: Build comparison SQL",
			tableName: "purchase_orders",
			entities:  []agent.Entity{},
			wantSQL:   "SELECT period, SUM(amount) as amount FROM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := buildComparisonSQL(tt.tableName, tt.entities)

			// Assert
			assert.Contains(t, result, tt.wantSQL)
		})
	}
}

// TestBuildWhereClause tests buildWhereClause function
func TestBuildWhereClause(t *testing.T) {
	tests := []struct {
		name            string
		entities        []agent.Entity
		wantClause      string
		wantEmpty       bool
	}{
		{
			name: "TC-SG-15: Build WHERE clause with time range",
			entities: []agent.Entity{
				{Type: agent.EntityTypeTimeRange, Value: "2024-05-01 to 2024-05-31"},
			},
			wantClause: "created_at >= '2024-05-01' AND created_at <= '2024-05-31'",
			wantEmpty: false,
		},
		{
			name: "TC-SG-16: Build WHERE clause with amount",
			entities: []agent.Entity{
				{Type: agent.EntityTypeAmount, Value: "> 5000.00"},
			},
			wantClause: "amount > 5000.00",
			wantEmpty: false,
		},
		{
			name: "TC-SG-17: Build WHERE clause with supplier",
			entities: []agent.Entity{
				{Type: agent.EntityTypeSupplier, Value: "A公司"},
			},
			wantClause: "supplier_name LIKE 'A公司'",
			wantEmpty: false,
		},
		{
			name: "TC-SG-18: Build WHERE clause with status",
			entities: []agent.Entity{
				{Type: agent.EntityTypeStatus, Value: "completed"},
			},
			wantClause: "status = 'completed'",
			wantEmpty: false,
		},
		{
			name: "TC-SG-19: Build WHERE clause with multiple entities",
			entities: []agent.Entity{
				{Type: agent.EntityTypeTimeRange, Value: "2024-05-01 to 2024-05-31"},
				{Type: agent.EntityTypeStatus, Value: "completed"},
			},
			wantClause: "created_at >= '2024-05-01' AND created_at <= '2024-05-31' AND status = 'completed'",
			wantEmpty: false,
		},
		{
			name:      "TC-SG-20: Build WHERE clause with no entities",
			entities:  []agent.Entity{},
			wantClause: "",
			wantEmpty: true,
		},
		{
			name: "TC-SG-21: Build WHERE clause with amount range",
			entities: []agent.Entity{
				{Type: agent.EntityTypeAmount, Value: "between 5000.00 and 10000.00"},
			},
			wantClause: "amount between 5000.00 and 10000.00",
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := buildWhereClause(tt.entities)

			// Assert
			if tt.wantEmpty {
				assert.Empty(t, result)
			} else {
				assert.Contains(t, result, tt.wantClause)
			}
		})
	}
}

// TestMapDocumentTypeToTable tests mapDocumentTypeToTable function
func TestMapDocumentTypeToTable(t *testing.T) {
	tests := []struct {
		name         string
		docType      string
		wantTable    string
	}{
		{
			name:      "TC-SG-22: Map purchase_order",
			docType:   "purchase_order",
			wantTable: "purchase_orders",
		},
		{
			name:      "TC-SG-23: Map sales_order",
			docType:   "sales_order",
			wantTable: "sales_orders",
		},
		{
			name:      "TC-SG-24: Map payment",
			docType:   "payment",
			wantTable: "payments",
		},
		{
			name:      "TC-SG-25: Map receipt",
			docType:   "receipt",
			wantTable: "receipts",
		},
		{
			name:      "TC-SG-26: Map invoice",
			docType:   "invoice",
			wantTable: "invoices",
		},
		{
			name:      "TC-SG-27: Map unknown type",
			docType:   "unknown",
			wantTable: "purchase_orders", // Default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := mapDocumentTypeToTable(tt.docType)

			// Assert
			assert.Equal(t, tt.wantTable, result)
		})
	}
}

// TestAddPermissionFilter tests addPermissionFilter function
func TestAddPermissionFilter(t *testing.T) {
	tests := []struct {
		name              string
		sql               string
		permissionFilter  string
		wantSQLContains   string
	}{
		{
			name:              "TC-SG-28: Add filter to SQL without WHERE",
			sql:               "SELECT * FROM purchase_orders ORDER BY created_at DESC",
			permissionFilter:  "department = '销售部'",
			wantSQLContains:   "WHERE department = '销售部'",
		},
		{
			name:              "TC-SG-29: Add filter to SQL with WHERE",
			sql:               "SELECT * FROM purchase_orders WHERE status = 'completed' ORDER BY created_at DESC",
			permissionFilter:  "department = '销售部'",
			wantSQLContains:   "WHERE department = '销售部' AND status = 'completed'",
		},
		{
			name:              "TC-SG-30: Add filter to SQL with LIMIT",
			sql:               "SELECT * FROM purchase_orders LIMIT 100",
			permissionFilter:  "department = '销售部'",
			wantSQLContains:   "WHERE department = '销售部'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := addPermissionFilter(tt.sql, tt.permissionFilter)

			// Assert
			assert.Contains(t, result, tt.wantSQLContains)
		})
	}
}

// TestParseSQLResponse tests parseSQLResponse function
func TestParseSQLResponse(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantSQL       string
		wantErr       bool
	}{
		{
			name:    "TC-SG-31: Valid JSON",
			input:   `{"sql":"SELECT * FROM table","understanding":"test"}`,
			wantSQL: "SELECT * FROM table",
			wantErr: false,
		},
		{
			name:    "TC-SG-32: Invalid JSON",
			input:   "invalid json",
			wantSQL: "",
			wantErr: true,
		},
		{
			name:    "TC-SG-33: JSON with markdown",
			input:   "```json\n{\"sql\":\"SELECT * FROM table\",\"understanding\":\"test\"}\n```",
			wantSQL: "SELECT * FROM table",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result, err := parseSQLResponse(tt.input)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantSQL, result)
			}
		})
	}
}

// TestParseSQLResponseWithUnderstanding tests ParseSQLResponseWithUnderstanding function
func TestParseSQLResponseWithUnderstanding(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		wantSQL           string
		wantUnderstanding string
		wantErr           bool
	}{
		{
			name:              "TC-SG-34: Valid JSON with understanding",
			input:             `{"sql":"SELECT * FROM table","understanding":"查询所有数据"}`,
			wantSQL:           "SELECT * FROM table",
			wantUnderstanding: "查询所有数据",
			wantErr:           false,
		},
		{
			name:              "TC-SG-35: Invalid JSON",
			input:             "invalid json",
			wantSQL:           "",
			wantUnderstanding: "",
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			sql, understanding, err := ParseSQLResponseWithUnderstanding(tt.input)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, sql)
				assert.Empty(t, understanding)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantSQL, sql)
				assert.Equal(t, tt.wantUnderstanding, understanding)
			}
		})
	}
}

// TestSQLTemplate tests SQLTemplate
func TestSQLTemplate(t *testing.T) {
	template := NewSQLTemplate()

	tests := []struct {
		name           string
		intentType     agent.IntentType
		wantTemplate   string
	}{
		{
			name:         "TC-SG-36: Get statistics template",
			intentType:   agent.IntentTypeStatistics,
			wantTemplate: "SELECT COUNT(*) as total_count, SUM(amount) as total_amount FROM {table} WHERE {conditions}",
		},
		{
			name:         "TC-SG-37: Get detail template",
			intentType:   agent.IntentTypeDetail,
			wantTemplate: "SELECT * FROM {table} WHERE {conditions} ORDER BY created_at DESC LIMIT 100",
		},
		{
			name:         "TC-SG-38: Get trend template",
			intentType:   agent.IntentTypeTrend,
			wantTemplate: "SELECT DATE(created_at) as date, SUM(amount) as daily_amount FROM {table} WHERE {conditions} GROUP BY DATE(created_at) ORDER BY date",
		},
		{
			name:         "TC-SG-39: Get unknown type - fallback to detail",
			intentType:   agent.IntentTypeUnknown,
			wantTemplate: "SELECT * FROM {table} WHERE {conditions} ORDER BY created_at DESC LIMIT 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := template.GetTemplate(tt.intentType)

			// Assert
			assert.Contains(t, result, tt.wantTemplate[:20])
		})
	}
}

// TestBuildSQLPrompt tests buildSQLPrompt function
func TestBuildSQLPrompt(t *testing.T) {
	intent := &agent.Intent{
		Type:        agent.IntentTypeStatistics,
		Description: "统计查询",
	}
	entities := []agent.Entity{
		{Type: agent.EntityTypeTimeRange, Value: "2024-05-01", RawText: "上个月"},
	}
	result := buildSQLPrompt(intent, entities)
	assert.Contains(t, result, "统计查询")
	assert.Contains(t, result, "time_range")
	assert.Contains(t, result, "JSON")
}

// TestBuildSQLPromptWithSchema tests buildSQLPromptWithSchema function
func TestBuildSQLPromptWithSchema(t *testing.T) {
	intent := &agent.Intent{
		Type:        agent.IntentTypeDetail,
		Description: "明细查询",
	}
	entities := []agent.Entity{}
	tableSchema := "purchase_orders (id, name, amount)"
	result := buildSQLPromptWithSchema(intent, entities, tableSchema)
	assert.Contains(t, result, "明细查询")
	assert.Contains(t, result, "purchase_orders")
	assert.Contains(t, result, "JSON")
}