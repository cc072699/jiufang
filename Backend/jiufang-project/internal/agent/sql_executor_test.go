// Package agent implements the AI Agent for semantic understanding and SQL generation.
// This file implements unit tests for SQLExecutor.
// Author: AI Assistant
// Date: 2026-06-03
// Tested Object: SQLExecutor
// Function: SQL execution on ERP database

package agent

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"jiufang/internal/infrastructure/erp"
	"jiufang/internal/mocks"
	"jiufang/internal/model/agent"
)

// TestSQLExecutor_Execute tests Execute method
func TestSQLExecutor_Execute(t *testing.T) {
	tests := []struct {
		name          string
		sql           string
		mockERP       func(ctrl *gomock.Controller) *mocks.MockERPReader
		wantResult    *agent.QueryResult
		wantErr       bool
		errContains   string
	}{
		{
			name: "TC-SE-01: Execute successfully",
			sql:  "SELECT * FROM purchase_orders LIMIT 100",
			mockERP: func(ctrl *gomock.Controller) *mocks.MockERPReader {
				mock := mocks.NewMockERPReader(ctrl)
				data := []map[string]interface{}{
					{"id": 1, "name": "采购单1", "amount": 1000.0},
					{"id": 2, "name": "采购单2", "amount": 2000.0},
				}
				mock.EXPECT().Query(gomock.Any(), gomock.Any()).Return(data, nil)
				return mock
			},
			wantResult: &agent.QueryResult{
				Data: []map[string]interface{}{
					{"id": 1, "name": "采购单1", "amount": 1000.0},
					{"id": 2, "name": "采购单2", "amount": 2000.0},
				},
				TotalRows: 2,
				IsEmpty:   false,
			},
			wantErr: false,
		},
		{
			name:        "TC-SE-02: SQL validation failure",
			sql:         "DELETE FROM purchase_orders",
			mockERP:     func(ctrl *gomock.Controller) *mocks.MockERPReader {
				return mocks.NewMockERPReader(ctrl)
			},
			wantResult:  nil,
			wantErr:     true,
			errContains: "危险关键字",
		},
		{
			name: "TC-SE-03: ERP query failure",
			sql:  "SELECT * FROM purchase_orders LIMIT 100",
			mockERP: func(ctrl *gomock.Controller) *mocks.MockERPReader {
				mock := mocks.NewMockERPReader(ctrl)
				mock.EXPECT().Query(gomock.Any(), gomock.Any()).Return(nil, errors.New("database error"))
				return mock
			},
			wantResult:  nil,
			wantErr:     true,
			errContains: "查询执行失败",
		},
		{
			name: "TC-SE-04: Execute with empty result",
			sql:  "SELECT * FROM purchase_orders WHERE status = 'unknown' LIMIT 100",
			mockERP: func(ctrl *gomock.Controller) *mocks.MockERPReader {
				mock := mocks.NewMockERPReader(ctrl)
				data := []map[string]interface{}{}
				mock.EXPECT().Query(gomock.Any(), gomock.Any()).Return(data, nil)
				return mock
			},
			wantResult: &agent.QueryResult{
				Data:      []map[string]interface{}{},
				TotalRows: 0,
				IsEmpty:   true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockERP := tt.mockERP(ctrl)
			validator := NewSQLValidator()
			executor := NewSQLExecutor(mockERP, validator)

			// Act
			result, err := executor.Execute(context.Background(), tt.sql)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantResult.TotalRows, result.TotalRows)
				assert.Equal(t, tt.wantResult.IsEmpty, result.IsEmpty)
			}
		})
	}
}

// TestSQLExecutor_ExecuteWithPermission tests ExecuteWithPermission method
func TestSQLExecutor_ExecuteWithPermission(t *testing.T) {
	tests := []struct {
		name          string
		sql           string
		queryContext  *agent.QueryContext
		mockERP       func(ctrl *gomock.Controller) *mocks.MockERPReader
		wantResult    *agent.QueryResult
		wantErr       bool
		errContains   string
	}{
		{
			name: "TC-SE-05: Execute with permission - allowed",
			sql:  "SELECT * FROM purchase_orders LIMIT 100",
			queryContext: &agent.QueryContext{
				UserID:          123,
				SessionID:       "session-123",
				AllowedTables:   []string{"purchase_orders"},
				PermissionFilter: "",
			},
			mockERP: func(ctrl *gomock.Controller) *mocks.MockERPReader {
				mock := mocks.NewMockERPReader(ctrl)
				data := []map[string]interface{}{
					{"id": 1, "name": "采购单1"},
				}
				mock.EXPECT().Query(gomock.Any(), gomock.Any()).Return(data, nil)
				return mock
			},
			wantResult: &agent.QueryResult{
				Data:           []map[string]interface{}{{"id": 1, "name": "采购单1"}},
				TotalRows:      1,
				PermissionFiltered: false,
			},
			wantErr: false,
		},
		{
			name: "TC-SE-06: Execute with permission - not allowed table",
			sql:  "SELECT * FROM users LIMIT 100",
			queryContext: &agent.QueryContext{
				UserID:        123,
				SessionID:     "session-123",
				AllowedTables: []string{"purchase_orders"},
			},
			mockERP: func(ctrl *gomock.Controller) *mocks.MockERPReader {
				return mocks.NewMockERPReader(ctrl)
			},
			wantResult:  nil,
			wantErr:     true,
			errContains: "无权限查询表",
		},
		{
			name: "TC-SE-07: Execute with permission filter",
			sql:  "SELECT * FROM purchase_orders LIMIT 100",
			queryContext: &agent.QueryContext{
				UserID:           123,
				SessionID:        "session-123",
				AllowedTables:    []string{"purchase_orders"},
				PermissionFilter: "department = '销售部'",
			},
			mockERP: func(ctrl *gomock.Controller) *mocks.MockERPReader {
				mock := mocks.NewMockERPReader(ctrl)
				data := []map[string]interface{}{
					{"id": 1, "name": "采购单1", "department": "销售部"},
				}
				mock.EXPECT().Query(gomock.Any(), gomock.Any()).Return(data, nil)
				return mock
			},
			wantResult: &agent.QueryResult{
				Data:           []map[string]interface{}{{"id": 1, "name": "采购单1", "department": "销售部"}},
				TotalRows:      1,
				PermissionFiltered: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockERP := tt.mockERP(ctrl)
			validator := NewSQLValidator()
			executor := NewSQLExecutor(mockERP, validator)

			// Act
			result, err := executor.ExecuteWithPermission(context.Background(), tt.sql, tt.queryContext)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantResult.TotalRows, result.TotalRows)
				assert.Equal(t, tt.wantResult.PermissionFiltered, result.PermissionFiltered)
			}
		})
	}
}

// TestSQLExecutor_ExecuteWithTimeout tests ExecuteWithTimeout method
func TestSQLExecutor_ExecuteWithTimeout(t *testing.T) {
	tests := []struct {
		name          string
		sql           string
		timeout       time.Duration
		mockERP       func(ctrl *gomock.Controller) *mocks.MockERPReader
		wantResult    *agent.QueryResult
		wantErr       bool
		errContains   string
	}{
		{
			name:    "TC-SE-08: Execute with timeout successfully",
			sql:     "SELECT * FROM purchase_orders LIMIT 100",
			timeout: 30 * time.Second,
			mockERP: func(ctrl *gomock.Controller) *mocks.MockERPReader {
				mock := mocks.NewMockERPReader(ctrl)
				data := []map[string]interface{}{
					{"id": 1, "name": "采购单1"},
				}
				mock.EXPECT().QueryWithTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Return(data, nil)
				return mock
			},
			wantResult: &agent.QueryResult{
				Data:      []map[string]interface{}{{"id": 1, "name": "采购单1"}},
				TotalRows: 1,
			},
			wantErr: false,
		},
		{
			name:    "TC-SE-09: Execute with timeout failure",
			sql:     "SELECT * FROM purchase_orders LIMIT 100",
			timeout: 30 * time.Second,
			mockERP: func(ctrl *gomock.Controller) *mocks.MockERPReader {
				mock := mocks.NewMockERPReader(ctrl)
				mock.EXPECT().QueryWithTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("timeout"))
				return mock
			},
			wantResult:  nil,
			wantErr:     true,
			errContains: "查询执行失败",
		},
		{
			name:        "TC-SE-10: Execute with timeout - validation failure",
			sql:         "DELETE FROM purchase_orders",
			timeout:     30 * time.Second,
			mockERP:     func(ctrl *gomock.Controller) *mocks.MockERPReader {
				return mocks.NewMockERPReader(ctrl)
			},
			wantResult:  nil,
			wantErr:     true,
			errContains: "危险关键字",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockERP := tt.mockERP(ctrl)
			validator := NewSQLValidator()
			executor := NewSQLExecutor(mockERP, validator)

			// Act
			result, err := executor.ExecuteWithTimeout(context.Background(), tt.sql, tt.timeout)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantResult.TotalRows, result.TotalRows)
			}
		})
	}
}

// TestSQLExecutor_ExecuteWithLimit tests ExecuteWithLimit method
func TestSQLExecutor_ExecuteWithLimit(t *testing.T) {
	tests := []struct {
		name          string
		sql           string
		limit         int
		mockERP       func(ctrl *gomock.Controller) *mocks.MockERPReader
		wantResult    *agent.QueryResult
		wantErr       bool
		errContains   string
	}{
		{
			name:  "TC-SE-11: Execute with limit successfully",
			sql:   "SELECT * FROM purchase_orders",
			limit: 50,
			mockERP: func(ctrl *gomock.Controller) *mocks.MockERPReader {
				mock := mocks.NewMockERPReader(ctrl)
				data := []map[string]interface{}{
					{"id": 1, "name": "采购单1"},
				}
				mock.EXPECT().QueryWithLimit(gomock.Any(), gomock.Any(), gomock.Any()).Return(data, nil)
				return mock
			},
			wantResult: &agent.QueryResult{
				Data:      []map[string]interface{}{{"id": 1, "name": "采购单1"}},
				TotalRows: 1,
				HasMore:   false,
			},
			wantErr: false,
		},
		{
			name:  "TC-SE-12: Execute with limit - has more results",
			sql:   "SELECT * FROM purchase_orders",
			limit: 50,
			mockERP: func(ctrl *gomock.Controller) *mocks.MockERPReader {
				mock := mocks.NewMockERPReader(ctrl)
				data := make([]map[string]interface{}, 50)
				for i := 0; i < 50; i++ {
					data[i] = map[string]interface{}{"id": i + 1}
				}
				mock.EXPECT().QueryWithLimit(gomock.Any(), gomock.Any(), gomock.Any()).Return(data, nil)
				return mock
			},
			wantResult: &agent.QueryResult{
				Data:      make([]map[string]interface{}, 50),
				TotalRows: 50,
				HasMore:   true,
			},
			wantErr: false,
		},
		{
			name:        "TC-SE-13: Execute with limit - validation failure",
			sql:         "DELETE FROM purchase_orders",
			limit:       50,
			mockERP:     func(ctrl *gomock.Controller) *mocks.MockERPReader {
				return mocks.NewMockERPReader(ctrl)
			},
			wantResult:  nil,
			wantErr:     true,
			errContains: "危险关键字",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockERP := tt.mockERP(ctrl)
			validator := NewSQLValidator()
			executor := NewSQLExecutor(mockERP, validator)

			// Act
			result, err := executor.ExecuteWithLimit(context.Background(), tt.sql, tt.limit)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantResult.TotalRows, result.TotalRows)
				assert.Equal(t, tt.wantResult.HasMore, result.HasMore)
			}
		})
	}
}

// TestSQLExecutor_ExecuteSafe tests ExecuteSafe method
func TestSQLExecutor_ExecuteSafe(t *testing.T) {
	tests := []struct {
		name          string
		sql           string
		queryContext  *agent.QueryContext
		mockERP       func(ctrl *gomock.Controller) *mocks.MockERPReader
		wantResult    *agent.QueryResult
		wantErr       bool
		errContains   string
	}{
		{
			name: "TC-SE-14: ExecuteSafe successfully",
			sql:  "SELECT * FROM purchase_orders LIMIT 100",
			queryContext: &agent.QueryContext{
				UserID:           123,
				SessionID:        "session-123",
				AllowedTables:    []string{"purchase_orders"},
				PermissionFilter: "department = '销售部'",
			},
			mockERP: func(ctrl *gomock.Controller) *mocks.MockERPReader {
				mock := mocks.NewMockERPReader(ctrl)
				data := []map[string]interface{}{
					{"id": 1, "name": "采购单1"},
				}
				mock.EXPECT().QueryWithTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Return(data, nil)
				return mock
			},
			wantResult: &agent.QueryResult{
				Data:           []map[string]interface{}{{"id": 1, "name": "采购单1"}},
				TotalRows:      1,
				PermissionFiltered: true,
			},
			wantErr: false,
		},
		{
			name: "TC-SE-15: ExecuteSafe without query context",
			sql:  "SELECT * FROM purchase_orders LIMIT 100",
			queryContext: nil,
			mockERP: func(ctrl *gomock.Controller) *mocks.MockERPReader {
				mock := mocks.NewMockERPReader(ctrl)
				data := []map[string]interface{}{
					{"id": 1, "name": "采购单1"},
				}
				mock.EXPECT().QueryWithTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Return(data, nil)
				return mock
			},
			wantResult: &agent.QueryResult{
				Data:           []map[string]interface{}{{"id": 1, "name": "采购单1"}},
				TotalRows:      1,
				PermissionFiltered: false,
			},
			wantErr: false,
		},
		{
			name:        "TC-SE-16: ExecuteSafe - validation failure",
			sql:         "DELETE FROM purchase_orders",
			queryContext: nil,
			mockERP:     func(ctrl *gomock.Controller) *mocks.MockERPReader {
				return mocks.NewMockERPReader(ctrl)
			},
			wantResult:  nil,
			wantErr:     true,
			errContains: "危险关键字",
		},
		{
			name: "TC-SE-17: ExecuteSafe - permission failure",
			sql:  "SELECT * FROM users LIMIT 100",
			queryContext: &agent.QueryContext{
				UserID:        123,
				SessionID:     "session-123",
				AllowedTables: []string{"purchase_orders"},
			},
			mockERP: func(ctrl *gomock.Controller) *mocks.MockERPReader {
				return mocks.NewMockERPReader(ctrl)
			},
			wantResult:  nil,
			wantErr:     true,
			errContains: "无权限查询表",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockERP := tt.mockERP(ctrl)
			validator := NewSQLValidator()
			executor := NewSQLExecutor(mockERP, validator)

			// Act
			result, err := executor.ExecuteSafe(context.Background(), tt.sql, tt.queryContext)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantResult.TotalRows, result.TotalRows)
				assert.Equal(t, tt.wantResult.PermissionFiltered, result.PermissionFiltered)
			}
		})
	}
}

// TestSQLExecutor_GetTableSchema tests GetTableSchema method
func TestSQLExecutor_GetTableSchema(t *testing.T) {
	tests := []struct {
		name          string
		tableName     string
		mockERP       func(ctrl *gomock.Controller) *mocks.MockERPReader
		wantSchema    *erp.TableSchema
		wantErr       bool
	}{
		{
			name:      "TC-SE-18: GetTableSchema successfully",
			tableName: "purchase_orders",
			mockERP: func(ctrl *gomock.Controller) *mocks.MockERPReader {
				mock := mocks.NewMockERPReader(ctrl)
				schema := &erp.TableSchema{
					Name: "purchase_orders",
					Columns: []erp.ColumnInfo{
						{Name: "id", Type: "int"},
						{Name: "name", Type: "varchar"},
					},
				}
				mock.EXPECT().GetTableSchema(gomock.Any(), gomock.Any()).Return(schema, nil)
				return mock
			},
			wantSchema: &erp.TableSchema{
				Name: "purchase_orders",
				Columns: []erp.ColumnInfo{
					{Name: "id", Type: "int"},
					{Name: "name", Type: "varchar"},
				},
			},
			wantErr: false,
		},
		{
			name:      "TC-SE-19: GetTableSchema failure",
			tableName: "unknown_table",
			mockERP: func(ctrl *gomock.Controller) *mocks.MockERPReader {
				mock := mocks.NewMockERPReader(ctrl)
				mock.EXPECT().GetTableSchema(gomock.Any(), gomock.Any()).Return(nil, errors.New("table not found"))
				return mock
			},
			wantSchema: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockERP := tt.mockERP(ctrl)
			validator := NewSQLValidator()
			executor := NewSQLExecutor(mockERP, validator)

			// Act
			result, err := executor.GetTableSchema(context.Background(), tt.tableName)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantSchema.Name, result.Name)
				assert.Equal(t, len(tt.wantSchema.Columns), len(result.Columns))
			}
		})
	}
}

// TestSQLExecutor_GetTableList tests GetTableList method
func TestSQLExecutor_GetTableList(t *testing.T) {
	tests := []struct {
		name          string
		mockERP       func(ctrl *gomock.Controller) *mocks.MockERPReader
		wantTables    []string
		wantErr       bool
	}{
		{
			name: "TC-SE-20: GetTableList successfully",
			mockERP: func(ctrl *gomock.Controller) *mocks.MockERPReader {
				mock := mocks.NewMockERPReader(ctrl)
				tables := []string{"purchase_orders", "sales_orders", "payments"}
				mock.EXPECT().GetTableList(gomock.Any()).Return(tables, nil)
				return mock
			},
			wantTables: []string{"purchase_orders", "sales_orders", "payments"},
			wantErr:    false,
		},
		{
			name: "TC-SE-21: GetTableList failure",
			mockERP: func(ctrl *gomock.Controller) *mocks.MockERPReader {
				mock := mocks.NewMockERPReader(ctrl)
				mock.EXPECT().GetTableList(gomock.Any()).Return(nil, errors.New("database error"))
				return mock
			},
			wantTables: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockERP := tt.mockERP(ctrl)
			validator := NewSQLValidator()
			executor := NewSQLExecutor(mockERP, validator)

			// Act
			result, err := executor.GetTableList(context.Background())

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, len(tt.wantTables), len(result))
			}
		})
	}
}

// TestSQLExecutor_ValidateSQL tests ValidateSQL method
func TestSQLExecutor_ValidateSQL(t *testing.T) {
	validator := NewSQLValidator()
	executor := NewSQLExecutor(nil, validator)

	tests := []struct {
		name          string
		sql           string
		wantErr       bool
	}{
		{
			name:    "TC-SE-22: ValidateSQL - valid SQL",
			sql:     "SELECT * FROM purchase_orders LIMIT 100",
			wantErr: false,
		},
		{
			name:    "TC-SE-23: ValidateSQL - invalid SQL",
			sql:     "DELETE FROM purchase_orders",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := executor.ValidateSQL(tt.sql)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSQLExecutor_IsReadOnly tests IsReadOnly method
func TestSQLExecutor_IsReadOnly(t *testing.T) {
	validator := NewSQLValidator()
	executor := NewSQLExecutor(nil, validator)

	tests := []struct {
		name          string
		sql           string
		wantReadOnly  bool
	}{
		{
			name:         "TC-SE-24: IsReadOnly - SELECT query",
			sql:          "SELECT * FROM purchase_orders",
			wantReadOnly: true,
		},
		{
			name:         "TC-SE-25: IsReadOnly - DELETE query",
			sql:          "DELETE FROM purchase_orders",
			wantReadOnly: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := executor.IsReadOnly(tt.sql)

			// Assert
			assert.Equal(t, tt.wantReadOnly, result)
		})
	}
}

// TestNewSQLExecutor tests NewSQLExecutor constructor
func TestNewSQLExecutor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockERP := mocks.NewMockERPReader(ctrl)
	validator := NewSQLValidator()

	// Act
	executor := NewSQLExecutor(mockERP, validator)

	// Assert
	assert.NotNil(t, executor)
	assert.NotNil(t, executor.erpReader)
	assert.NotNil(t, executor.validator)
}

// TestSQLExecutionError tests SQLExecutionError struct
func TestSQLExecutionError(t *testing.T) {
	tests := []struct {
		name          string
		sql           string
		errorType     string
		message       string
		isRetryable   bool
		wantErrString string
	}{
		{
			name:          "TC-SE-26: SQLExecutionError - validation",
			sql:           "DELETE FROM table",
			errorType:     ErrorTypeValidation,
			message:       "SQL validation failed",
			isRetryable:   false,
			wantErrString: "SQL执行错误 [validation]: SQL validation failed",
		},
		{
			name:          "TC-SE-27: SQLExecutionError - timeout",
			sql:           "SELECT * FROM table",
			errorType:     ErrorTypeTimeout,
			message:       "Query timeout",
			isRetryable:   true,
			wantErrString: "SQL执行错误 [timeout]: Query timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := NewSQLExecutionError(tt.sql, tt.errorType, tt.message, tt.isRetryable)

			// Assert
			assert.Equal(t, tt.sql, err.SQL)
			assert.Equal(t, tt.errorType, err.ErrorType)
			assert.Equal(t, tt.message, err.Message)
			assert.Equal(t, tt.isRetryable, err.IsRetryable)
			assert.Contains(t, err.Error(), tt.wantErrString)
		})
	}
}

// TestExecutionStats tests ExecutionStats struct
func TestExecutionStats(t *testing.T) {
	stats := ExecutionStats{
		TotalQueries:      100,
		SuccessQueries:    95,
		FailedQueries:     5,
		TotalTime:         5000,
		AverageTime:       50,
		MaxTime:           200,
		MinTime:           10,
		PermissionFiltered: 20,
	}

	// Assert
	assert.Equal(t, int64(100), stats.TotalQueries)
	assert.Equal(t, int64(95), stats.SuccessQueries)
	assert.Equal(t, int64(5), stats.FailedQueries)
	assert.Equal(t, int64(5000), stats.TotalTime)
	assert.Equal(t, int64(50), stats.AverageTime)
	assert.Equal(t, int64(200), stats.MaxTime)
	assert.Equal(t, int64(10), stats.MinTime)
	assert.Equal(t, int64(20), stats.PermissionFiltered)
}

// TestQueryResult tests QueryResult struct methods
func TestQueryResult_NewEmptyQueryResult(t *testing.T) {
	understanding := "未找到匹配的数据"
	result := agent.NewEmptyQueryResult(understanding)

	// Assert
	assert.NotNil(t, result)
	assert.Empty(t, result.Data)
	assert.Equal(t, understanding, result.Understanding)
	assert.Equal(t, agent.VisualizationTable, result.VisualizationType)
	assert.Equal(t, 0, result.TotalRows)
	assert.True(t, result.IsEmpty)
}

// TestQueryResult_NewQueryResult tests NewQueryResult function
func TestQueryResult_NewQueryResult(t *testing.T) {
	data := []map[string]interface{}{
		{"id": 1, "name": "采购单1", "amount": 1000.0},
		{"id": 2, "name": "采购单2", "amount": 2000.0},
	}
	understanding := "查询采购单数据"
	execTime := int64(150)
	sql := "SELECT * FROM purchase_orders LIMIT 100"

	result := agent.NewQueryResult(data, understanding, execTime, sql)

	// Assert
	assert.NotNil(t, result)
	assert.Equal(t, 2, result.TotalRows)
	assert.Equal(t, understanding, result.Understanding)
	assert.Equal(t, execTime, result.ExecutionTime)
	assert.Equal(t, sql, result.GeneratedSQL)
	assert.False(t, result.IsEmpty)
	assert.NotNil(t, result.Columns)
	assert.NotNil(t, result.Timestamp)
}