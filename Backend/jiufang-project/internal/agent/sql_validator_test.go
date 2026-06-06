// Package agent implements the AI Agent for semantic understanding and SQL generation.
// This file implements unit tests for SQLValidator.
// Author: AI Assistant
// Date: 2026-06-03
// Tested Object: SQLValidator
// Function: SQL validation for safety and correctness

package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"jiufang/internal/model/agent"
)

// TestSQLValidator_Validate tests Validate method
func TestSQLValidator_Validate(t *testing.T) {
	validator := NewSQLValidator()

	tests := []struct {
		name          string
		sql           string
		wantErr       bool
		errContains   string
	}{
		{
			name:    "TC-SV-01: Valid SELECT query",
			sql:     "SELECT * FROM purchase_orders LIMIT 100",
			wantErr: false,
		},
		{
			name:        "TC-SV-02: Empty SQL",
			sql:         "",
			wantErr:     true,
			errContains: "SQL语句为空",
		},
		{
			name:        "TC-SV-03: DELETE statement",
			sql:         "DELETE FROM purchase_orders WHERE id = 1",
			wantErr:     true,
			errContains: "危险关键字",
		},
		{
			name:        "TC-SV-04: UPDATE statement",
			sql:         "UPDATE purchase_orders SET amount = 100 WHERE id = 1",
			wantErr:     true,
			errContains: "危险关键字",
		},
		{
			name:        "TC-SV-05: INSERT statement",
			sql:         "INSERT INTO purchase_orders (id, name) VALUES (1, 'test')",
			wantErr:     true,
			errContains: "危险关键字",
		},
		{
			name:        "TC-SV-06: DROP statement",
			sql:         "DROP TABLE purchase_orders",
			wantErr:     true,
			errContains: "危险关键字",
		},
		{
			name:        "TC-SV-07: SQL comment injection",
			sql:         "SELECT * FROM purchase_orders -- comment",
			wantErr:     true,
			errContains: "危险模式",
		},
		{
			name:        "TC-SV-08: UNION injection",
			sql:         "SELECT * FROM purchase_orders UNION SELECT * FROM users",
			wantErr:     true,
			errContains: "危险模式",
		},
		{
			name:        "TC-SV-09: Multiple statements with semicolon",
			sql:         "SELECT * FROM purchase_orders; DROP TABLE purchase_orders",
			wantErr:     true,
			errContains: "危险模式",
		},
		{
			name:        "TC-SV-10: Block comment injection",
			sql:         "SELECT * FROM purchase_orders /* comment */",
			wantErr:     true,
			errContains: "危险模式",
		},
		{
			name:        "TC-SV-11: INTO OUTFILE injection",
			sql:         "SELECT * INTO OUTFILE '/tmp/data.txt' FROM purchase_orders",
			wantErr:     true,
			errContains: "危险模式",
		},
		{
			name:        "TC-SV-12: Non-SELECT statement",
			sql:         "SHOW TABLES",
			wantErr:     true,
			errContains: "只允许SELECT",
		},
		{
			name:    "TC-SV-13: Valid SELECT with JOIN",
			sql:     "SELECT * FROM purchase_orders JOIN suppliers ON purchase_orders.supplier_id = suppliers.id LIMIT 100",
			wantErr: false,
		},
		{
			name:    "TC-SV-14: Valid SELECT with WHERE",
			sql:     "SELECT * FROM purchase_orders WHERE status = 'completed' LIMIT 100",
			wantErr: false,
		},
		{
			name:    "TC-SV-15: Valid SELECT without LIMIT",
			sql:     "SELECT * FROM purchase_orders",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := validator.Validate(tt.sql)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSQLValidator_IsReadOnly tests IsReadOnly method
func TestSQLValidator_IsReadOnly(t *testing.T) {
	validator := NewSQLValidator()

	tests := []struct {
		name          string
		sql           string
		wantReadOnly  bool
	}{
		{
			name:         "TC-SV-16: SELECT is read-only",
			sql:          "SELECT * FROM purchase_orders",
			wantReadOnly: true,
		},
		{
			name:         "TC-SV-17: DELETE is not read-only",
			sql:          "DELETE FROM purchase_orders",
			wantReadOnly: false,
		},
		{
			name:         "TC-SV-18: UPDATE is not read-only",
			sql:          "UPDATE purchase_orders SET amount = 100",
			wantReadOnly: false,
		},
		{
			name:         "TC-SV-19: INSERT is not read-only",
			sql:          "INSERT INTO purchase_orders VALUES (1, 'test')",
			wantReadOnly: false,
		},
		{
			name:         "TC-SV-20: SELECT with lowercase",
			sql:          "select * from purchase_orders",
			wantReadOnly: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := validator.IsReadOnly(tt.sql)

			// Assert
			assert.Equal(t, tt.wantReadOnly, result)
		})
	}
}

// TestSQLValidator_ValidateWithPermission tests ValidateWithPermission method
func TestSQLValidator_ValidateWithPermission(t *testing.T) {
	validator := NewSQLValidator()

	tests := []struct {
		name          string
		sql           string
		queryContext  *agent.QueryContext
		wantErr       bool
		errContains   string
	}{
		{
			name: "TC-SV-21: Valid permission - allowed table",
			sql:  "SELECT * FROM purchase_orders LIMIT 100",
			queryContext: &agent.QueryContext{
				AllowedTables: []string{"purchase_orders", "sales_orders"},
			},
			wantErr: false,
		},
		{
			name: "TC-SV-22: Invalid permission - not allowed table",
			sql:  "SELECT * FROM users LIMIT 100",
			queryContext: &agent.QueryContext{
				AllowedTables: []string{"purchase_orders", "sales_orders"},
			},
			wantErr:     true,
			errContains: "无权限查询表",
		},
		{
			name: "TC-SV-23: Empty allowed tables - all allowed",
			sql:  "SELECT * FROM purchase_orders LIMIT 100",
			queryContext: &agent.QueryContext{
				AllowedTables: []string{},
			},
			wantErr: false,
		},
		{
			name: "TC-SV-24: Valid permission - allowed fields",
			sql:  "SELECT id, name FROM purchase_orders LIMIT 100",
			queryContext: &agent.QueryContext{
				AllowedFields: map[string][]string{
					"purchase_orders": []string{"id", "name", "amount"},
				},
			},
			wantErr: false,
		},
		{
			name: "TC-SV-25: Invalid permission - not allowed field",
			sql:  "SELECT password FROM users LIMIT 100",
			queryContext: &agent.QueryContext{
				AllowedFields: map[string][]string{
					"users": []string{"id", "name"},
				},
			},
			wantErr:     true,
			errContains: "无权限查询字段",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := validator.ValidateWithPermission(tt.sql, tt.queryContext)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSQLValidator_SetAllowedTables tests SetAllowedTables method
func TestSQLValidator_SetAllowedTables(t *testing.T) {
	validator := NewSQLValidator()

	// Act
	validator.SetAllowedTables([]string{"purchase_orders", "sales_orders"})

	// Assert - test with allowed table
	err := validator.Validate("SELECT * FROM purchase_orders LIMIT 100")
	assert.NoError(t, err)

	// Assert - test with not allowed table
	err = validator.Validate("SELECT * FROM users LIMIT 100")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不允许查询表")
}

// TestSQLValidator_SetMaxResultRows tests SetMaxResultRows method
func TestSQLValidator_SetMaxResultRows(t *testing.T) {
	validator := NewSQLValidator()

	// Act
	validator.SetMaxResultRows(5000)

	// Assert
	assert.Equal(t, 5000, validator.GetMaxResultRows())
}

// TestSQLValidator_GetMaxResultRows tests GetMaxResultRows method
func TestSQLValidator_GetMaxResultRows(t *testing.T) {
	validator := NewSQLValidator()

	// Assert - default value
	assert.Equal(t, 10000, validator.GetMaxResultRows())

	// Act & Assert - after setting
	validator.SetMaxResultRows(5000)
	assert.Equal(t, 5000, validator.GetMaxResultRows())
}

// TestSQLValidator_AddSafetyLimit tests AddSafetyLimit method
func TestSQLValidator_AddSafetyLimit(t *testing.T) {
	validator := NewSQLValidator()
	validator.SetMaxResultRows(10000)

	tests := []struct {
		name         string
		sql          string
		wantSQL      string
		wantContains string
	}{
		{
			name:         "TC-SV-26: Add LIMIT to SQL without LIMIT",
			sql:          "SELECT * FROM purchase_orders",
			wantContains: "LIMIT 10000",
		},
		{
			name:         "TC-SV-27: Do not add LIMIT if already present",
			sql:          "SELECT * FROM purchase_orders LIMIT 50",
			wantContains: "LIMIT 50",
		},
		{
			name:         "TC-SV-28: Add LIMIT before ORDER BY",
			sql:          "SELECT * FROM purchase_orders ORDER BY created_at DESC",
			wantContains: "LIMIT 10000",
		},
		{
			name:         "TC-SV-29: Add LIMIT before GROUP BY",
			sql:          "SELECT COUNT(*) FROM purchase_orders GROUP BY status",
			wantContains: "LIMIT 10000",
		},
		{
			name:         "TC-SV-30: Add LIMIT to complex query",
			sql:          "SELECT * FROM purchase_orders WHERE status = 'completed' ORDER BY created_at DESC",
			wantContains: "LIMIT 10000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := validator.AddSafetyLimit(tt.sql)

			// Assert
			assert.Contains(t, result, tt.wantContains)
		})
	}
}

// TestExtractTablesFromSQL tests extractTablesFromSQL function
func TestExtractTablesFromSQL(t *testing.T) {
	tests := []struct {
		name         string
		sql          string
		wantTables   []string
	}{
		{
			name:       "TC-SV-31: Extract single table from FROM",
			sql:        "SELECT * FROM purchase_orders",
			wantTables: []string{"purchase_orders"},
		},
		{
			name:       "TC-SV-32: Extract tables from JOIN",
			sql:        "SELECT * FROM purchase_orders JOIN suppliers ON purchase_orders.supplier_id = suppliers.id",
			wantTables: []string{"purchase_orders", "suppliers"},
		},
		{
			name:       "TC-SV-33: Extract multiple JOINs",
			sql:        "SELECT * FROM orders JOIN customers ON orders.customer_id = customers.id JOIN products ON orders.product_id = products.id",
			wantTables: []string{"orders", "customers", "products"},
		},
		{
			name:       "TC-SV-34: No table found",
			sql:        "SELECT 1",
			wantTables: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := extractTablesFromSQL(tt.sql)

			// Assert
			assert.Equal(t, len(tt.wantTables), len(result))
			for i, table := range tt.wantTables {
				if i < len(result) {
					assert.Equal(t, table, result[i])
				}
			}
		})
	}
}

// TestExtractFieldsFromSQL tests extractFieldsFromSQL function
func TestExtractFieldsFromSQL(t *testing.T) {
	tests := []struct {
		name         string
		sql          string
		wantFields   []string
	}{
		{
			name:       "TC-SV-35: Extract fields from SELECT",
			sql:        "SELECT id, name, amount FROM purchase_orders",
			wantFields: []string{"id", "name", "amount"},
		},
		{
			name:       "TC-SV-36: Extract field with table prefix",
			sql:        "SELECT purchase_orders.id, purchase_orders.name FROM purchase_orders",
			wantFields: []string{"id", "name"},
		},
		{
			name:       "TC-SV-37: Extract field with alias",
			sql:        "SELECT id AS order_id, name AS order_name FROM purchase_orders",
			wantFields: []string{"id", "name"},
		},
		{
			name:       "TC-SV-38: Extract all fields with *",
			sql:        "SELECT * FROM purchase_orders",
			wantFields: []string{"*"},
		},
		{
			name:       "TC-SV-39: No fields found",
			sql:        "SELECT COUNT(*) FROM purchase_orders",
			wantFields: []string{"count(*)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := extractFieldsFromSQL(tt.sql)

			// Assert
			assert.Equal(t, len(tt.wantFields), len(result))
		})
	}
}

// TestSQLValidator_ValidateDetailed tests ValidateDetailed method
func TestSQLValidator_ValidateDetailed(t *testing.T) {
	validator := NewSQLValidator()

	tests := []struct {
		name                string
		sql                 string
		wantIsValid         bool
		wantHasErrors       bool
		wantHasWarnings     bool
		wantIsReadOnly      bool
		wantHasLimit        bool
	}{
		{
			name:            "TC-SV-40: Valid SQL with LIMIT",
			sql:             "SELECT * FROM purchase_orders LIMIT 100",
			wantIsValid:     true,
			wantHasErrors:   false,
			wantHasWarnings: false,
			wantIsReadOnly:  true,
			wantHasLimit:    true,
		},
		{
			name:            "TC-SV-41: Valid SQL without LIMIT",
			sql:             "SELECT * FROM purchase_orders",
			wantIsValid:     true,
			wantHasErrors:   false,
			wantHasWarnings: true,
			wantIsReadOnly:  true,
			wantHasLimit:    false,
		},
		{
			name:            "TC-SV-42: Invalid SQL with DELETE",
			sql:             "DELETE FROM purchase_orders",
			wantIsValid:     false,
			wantHasErrors:   true,
			wantHasWarnings: false,
			wantIsReadOnly:  false,
			wantHasLimit:    false,
		},
		{
			name:            "TC-SV-43: Invalid SQL with UNION",
			sql:             "SELECT * FROM purchase_orders UNION SELECT * FROM users",
			wantIsValid:     false,
			wantHasErrors:   true,
			wantHasWarnings: false,
			wantIsReadOnly:  false,
			wantHasLimit:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := validator.ValidateDetailed(tt.sql)

			// Assert
			assert.Equal(t, tt.wantIsValid, result.IsValid)
			assert.Equal(t, tt.wantHasErrors, len(result.Errors) > 0)
			assert.Equal(t, tt.wantHasWarnings, len(result.Warnings) > 0)
			assert.Equal(t, tt.wantIsReadOnly, result.IsReadOnly)
			assert.Equal(t, tt.wantHasLimit, result.HasLimit)
		})
	}
}

// TestSQLValidator_ValidateIntentMatch tests ValidateIntentMatch method
func TestSQLValidator_ValidateIntentMatch(t *testing.T) {
	validator := NewSQLValidator()

	tests := []struct {
		name          string
		sql           string
		intent        *agent.Intent
		wantErr       bool
		errContains   string
	}{
		{
			name: "TC-SV-44: Statistics intent with COUNT",
			sql:  "SELECT COUNT(*) as total FROM purchase_orders",
			intent: &agent.Intent{
				Type: agent.IntentTypeStatistics,
			},
			wantErr: false,
		},
		{
			name: "TC-SV-45: Statistics intent with SUM",
			sql:  "SELECT SUM(amount) as total FROM purchase_orders",
			intent: &agent.Intent{
				Type: agent.IntentTypeStatistics,
			},
			wantErr: false,
		},
		{
			name: "TC-SV-46: Statistics intent without aggregation",
			sql:  "SELECT * FROM purchase_orders",
			intent: &agent.Intent{
				Type: agent.IntentTypeStatistics,
			},
			wantErr:     true,
			errContains: "应包含聚合函数",
		},
		{
			name: "TC-SV-47: Trend intent with GROUP BY",
			sql:  "SELECT DATE(created_at), SUM(amount) FROM purchase_orders GROUP BY DATE(created_at)",
			intent: &agent.Intent{
				Type: agent.IntentTypeTrend,
			},
			wantErr: false,
		},
		{
			name: "TC-SV-48: Trend intent without GROUP BY",
			sql:  "SELECT * FROM purchase_orders",
			intent: &agent.Intent{
				Type: agent.IntentTypeTrend,
			},
			wantErr:     true,
			errContains: "应包含GROUP BY",
		},
		{
			name: "TC-SV-49: Comparison intent with UNION",
			sql:  "SELECT * FROM purchase_orders UNION SELECT * FROM sales_orders",
			intent: &agent.Intent{
				Type: agent.IntentTypeComparison,
			},
			wantErr: false,
		},
		{
			name: "TC-SV-50: Comparison intent without comparison logic",
			sql:  "SELECT * FROM purchase_orders",
			intent: &agent.Intent{
				Type: agent.IntentTypeComparison,
			},
			wantErr:     true,
			errContains: "应包含对比逻辑",
		},
		{
			name: "TC-SV-51: Detail intent - no validation",
			sql:  "SELECT * FROM purchase_orders",
			intent: &agent.Intent{
				Type: agent.IntentTypeDetail,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := validator.ValidateIntentMatch(tt.sql, tt.intent)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSQLValidationResult tests SQLValidationResult struct
func TestSQLValidationResult(t *testing.T) {
	validator := NewSQLValidator()

	// Act
	result := validator.ValidateDetailed("SELECT * FROM purchase_orders")

	// Assert
	assert.NotNil(t, result)
	assert.NotNil(t, result.Tables)
	assert.NotNil(t, result.Fields)
	assert.NotNil(t, result.SafeSQL)
}

// TestNewSQLValidator tests NewSQLValidator constructor
func TestNewSQLValidator(t *testing.T) {
	validator := NewSQLValidator()

	// Assert
	assert.NotNil(t, validator)
	assert.NotNil(t, validator.dangerousKeywords)
	assert.NotNil(t, validator.dangerousPatterns)
	assert.NotNil(t, validator.allowedTables)
	assert.Equal(t, 10000, validator.maxResultRows)
}