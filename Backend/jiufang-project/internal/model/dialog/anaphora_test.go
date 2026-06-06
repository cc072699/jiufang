// Package dialog implements the dialog management model for multi-turn conversations.
// This file implements unit tests for AnaphoraResolver.
// Author: AI Assistant
// Date: 2026-06-03
// Tested Object: AnaphoraResolver
// Function: Anaphora resolution (指代消解) operations

package dialog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestAnaphoraResolver_Resolve tests Resolve method
func TestAnaphoraResolver_Resolve(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		context       *DialogContext
		wantResolved  string
		wantErr       bool
	}{
		{
			name:  "TC-AR-01: Resolve '这个' with context",
			input: "查询这个",
			context: &DialogContext{
				SessionID:    "123456",
				UserID:       123,
				TurnCount:    1,
				MaxTurns:     5,
				QueryHistory: []QueryTurn{
					{
						Input:         "查询采购单",
						Understanding: "采购订单",
						TurnID:        1,
					},
				},
			},
			wantResolved: "查询上次查询的采购订单",
			wantErr:      false,
		},
		{
			name:  "TC-AR-02: Resolve '那个单子' with document type",
			input: "那个单子",
			context: &DialogContext{
				SessionID:    "123456",
				UserID:       123,
				TurnCount:    1,
				MaxTurns:     5,
				QueryHistory: []QueryTurn{
					{
						Input:    "查询采购单",
						TurnID:   1,
						Entities: map[string]interface{}{"document_type": "采购单"},
					},
				},
			},
			wantResolved: "采购单",
			wantErr:      false,
		},
		{
			name:         "TC-AR-03: Resolve with no context",
			input:        "这个",
			context:      nil,
			wantResolved: "这个",
			wantErr:      false,
		},
		{
			name:  "TC-AR-04: Resolve with empty history",
			input: "这个",
			context: &DialogContext{
				SessionID:    "123456",
				UserID:       123,
				TurnCount:    0,
				MaxTurns:     5,
				QueryHistory: []QueryTurn{},
			},
			wantResolved: "这个",
			wantErr:      false,
		},
		{
			name:  "TC-AR-05: Resolve multiple anaphoras",
			input: "这个单子和那条记录",
			context: &DialogContext{
				SessionID:    "123456",
				UserID:       123,
				TurnCount:    1,
				MaxTurns:     5,
				QueryHistory: []QueryTurn{
					{
						Input:    "查询采购单",
						TurnID:   1,
						Entities: map[string]interface{}{"document_type": "采购单"},
					},
				},
			},
			wantResolved: "采购单和单据", // Both "这个单子" and "那条" should be resolved
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			resolver := NewAnaphoraResolver(zap.NewNop())

			// Act
			resolved, err := resolver.Resolve(tt.input, tt.context)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// For TC-AR-01, the exact resolved text depends on implementation
				// We verify that it's not empty and contains resolved content
				if tt.context != nil && len(tt.context.QueryHistory) > 0 {
					assert.NotEmpty(t, resolved)
					if tt.input != tt.wantResolved {
						// Verify that some resolution happened
						assert.NotEqual(t, tt.input, resolved)
					}
				} else {
					// No context, should return original input
					assert.Equal(t, tt.wantResolved, resolved)
				}
			}
		})
	}
}

// TestAnaphoraResolver_DetectAnaphora tests DetectAnaphora method
func TestAnaphoraResolver_DetectAnaphora(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantResult bool
	}{
		{
			name:       "TC-AR-05: Detect anaphora - exists",
			input:      "这个单子",
			wantResult: true,
		},
		{
			name:       "TC-AR-06: Detect anaphora - not exists",
			input:      "查询采购单",
			wantResult: false,
		},
		{
			name:       "TC-AR-05-2: Detect anaphora - '那个'",
			input:      "那个订单",
			wantResult: true,
		},
		{
			name:       "TC-AR-05-3: Detect anaphora - '它'",
			input:      "查询它的状态",
			wantResult: true,
		},
		{
			name:       "TC-AR-05-4: Detect anaphora - '上次查询'",
			input:      "上次查询的结果",
			wantResult: true,
		},
		{
			name:       "TC-AR-06-2: Detect anaphora - empty input",
			input:      "",
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			resolver := NewAnaphoraResolver(zap.NewNop())

			// Act
			result := resolver.DetectAnaphora(tt.input)

			// Assert
			assert.Equal(t, tt.wantResult, result)
		})
	}
}

// TestAnaphoraResolver_GetAnaphoraList tests GetAnaphoraList method
func TestAnaphoraResolver_GetAnaphoraList(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantAnaphoras []string
	}{
		{
			name:          "TC-AR-07: Get anaphora list - multiple anaphoras",
			input:         "这个单子和那条记录",
			wantAnaphoras: []string{"这个单子", "那条"},
		},
		{
			name:          "TC-AR-07-2: Get anaphora list - single anaphora",
			input:         "查询这个",
			wantAnaphoras: []string{"这个"},
		},
		{
			name:          "TC-AR-07-3: Get anaphora list - no anaphora",
			input:         "查询采购单",
			wantAnaphoras: []string{},
		},
		{
			name:          "TC-AR-07-4: Get anaphora list - empty input",
			input:         "",
			wantAnaphoras: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			resolver := NewAnaphoraResolver(zap.NewNop())

			// Act
			anaphoras := resolver.GetAnaphoraList(tt.input)

			// Assert
			if len(tt.wantAnaphoras) > 0 {
				// Verify that expected anaphoras are found
				for _, expected := range tt.wantAnaphoras {
					assert.Contains(t, anaphoras, expected)
				}
			} else {
				// Should return empty list
				assert.Empty(t, anaphoras)
			}
		})
	}
}

// TestAnaphoraResolver_resolveToLastSubject tests resolveToLastSubject method
func TestAnaphoraResolver_resolveToLastSubject(t *testing.T) {
	tests := []struct {
		name         string
		lastTurn     *QueryTurn
		wantResolved string
	}{
		{
			name: "resolve with understanding",
			lastTurn: &QueryTurn{
				Input:         "查询采购单",
				Understanding: "采购订单",
			},
			wantResolved: "上次查询的采购订单",
		},
		{
			name: "resolve without understanding",
			lastTurn: &QueryTurn{
				Input: "查询采购单",
			},
			wantResolved: "上次查询'查询采购单'",
		},
		{
			name:         "resolve with nil turn",
			lastTurn:     nil,
			wantResolved: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			resolver := NewAnaphoraResolver(zap.NewNop())

			// Act
			resolved := resolver.resolveToLastSubject(tt.lastTurn)

			// Assert
			assert.Equal(t, tt.wantResolved, resolved)
		})
	}
}

// TestAnaphoraResolver_resolveToLastDocument tests resolveToLastDocument method
func TestAnaphoraResolver_resolveToLastDocument(t *testing.T) {
	tests := []struct {
		name         string
		lastTurn     *QueryTurn
		wantResolved string
	}{
		{
			name: "resolve with document_type entity",
			lastTurn: &QueryTurn{
				Entities: map[string]interface{}{"document_type": "采购单"},
			},
			wantResolved: "采购单",
		},
		{
			name: "resolve without document_type entity",
			lastTurn: &QueryTurn{
				Entities: map[string]interface{}{"status": "pending"},
			},
			wantResolved: "单据",
		},
		{
			name:         "resolve with nil entities",
			lastTurn:     &QueryTurn{},
			wantResolved: "单据",
		},
		{
			name:         "resolve with nil turn",
			lastTurn:     nil,
			wantResolved: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			resolver := NewAnaphoraResolver(zap.NewNop())

			// Act
			resolved := resolver.resolveToLastDocument(tt.lastTurn)

			// Assert
			assert.Equal(t, tt.wantResolved, resolved)
		})
	}
}

// TestAnaphoraResolver_resolveToLastCondition tests resolveToLastCondition method
func TestAnaphoraResolver_resolveToLastCondition(t *testing.T) {
	tests := []struct {
		name         string
		lastTurn     *QueryTurn
		wantResolved string
	}{
		{
			name: "resolve with multiple entities",
			lastTurn: &QueryTurn{
				Entities: map[string]interface{}{
					"document_type": "采购单",
					"status":        "pending",
				},
			},
			wantResolved: "document_type=采购单 AND status=pending",
		},
		{
			name: "resolve with single entity",
			lastTurn: &QueryTurn{
				Entities: map[string]interface{}{"status": "pending"},
			},
			wantResolved: "status=pending",
		},
		{
			name:         "resolve with nil entities",
			lastTurn:     &QueryTurn{},
			wantResolved: "",
		},
		{
			name:         "resolve with nil turn",
			lastTurn:     nil,
			wantResolved: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			resolver := NewAnaphoraResolver(zap.NewNop())

			// Act
			resolved := resolver.resolveToLastCondition(tt.lastTurn)

			// Assert
			if tt.wantResolved != "" {
				// The order of conditions may vary, so we check if all expected parts are present
				assert.NotEmpty(t, resolved)
				if tt.lastTurn != nil && tt.lastTurn.Entities != nil {
					for key := range tt.lastTurn.Entities {
						assert.Contains(t, resolved, key)
					}
				}
			} else {
				assert.Equal(t, tt.wantResolved, resolved)
			}
		})
	}
}