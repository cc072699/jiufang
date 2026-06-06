// Package agent implements the AI Agent for semantic understanding and SQL generation.
// This file implements unit tests for EntityExtractor.
// Author: AI Assistant
// Date: 2026-06-03
// Tested Object: EntityExtractor
// Function: Entity extraction from natural language input

package agent

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"jiufang/internal/mocks"
	"jiufang/internal/model/agent"
)

// TestEntityExtractor_Extract tests Extract method
func TestEntityExtractor_Extract(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		mockLLM       func(ctrl *gomock.Controller) *mocks.MockLLMClient
		wantEntities  []agent.Entity
		wantErr       bool
	}{
		{
			name:  "TC-EE-01: Extract successfully with LLM",
			input: "上个月采购总额大于10000",
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				response := `{
					"entities": [
						{"type":"time_range","value":"2024-05-01 to 2024-05-31","raw_text":"上个月","normalized":"last_month","confidence":0.8},
						{"type":"amount","value":"> 10000.00","raw_text":"大于10000","normalized":"amount_condition","confidence":0.9}
					]
				}`
				mock.EXPECT().Generate(gomock.Any(), gomock.Any()).Return(response, nil)
				return mock
			},
			wantEntities: []agent.Entity{
				{Type: agent.EntityTypeTimeRange, Value: "2024-05-01 to 2024-05-31", RawText: "上个月", Normalized: "last_month", Confidence: 0.8},
				{Type: agent.EntityTypeAmount, Value: "> 10000.00", RawText: "大于10000", Normalized: "amount_condition", Confidence: 0.9},
			},
			wantErr: false,
		},
		{
			name:  "TC-EE-02: LLM failure - fallback to template",
			input: "上个月的采购单",
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				mock.EXPECT().Generate(gomock.Any(), gomock.Any()).Return("", errors.New("LLM error"))
				return mock
			},
			wantEntities: []agent.Entity{
				{Type: agent.EntityTypeTimeRange, Confidence: 0.8, Normalized: "last_month"},
			},
			wantErr: false,
		},
		{
			name:  "TC-EE-03: LLM returns invalid JSON - fallback",
			input: "采购单",
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				mock.EXPECT().Generate(gomock.Any(), gomock.Any()).Return("invalid json", nil)
				return mock
			},
			wantEntities: []agent.Entity{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLLM := tt.mockLLM(ctrl)
			extractor := NewEntityExtractor(mockLLM)

			// Act
			result, err := extractor.Extract(context.Background(), tt.input)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.wantEntities), len(result))
				if len(tt.wantEntities) > 0 && len(result) > 0 {
					assert.Equal(t, tt.wantEntities[0].Type, result[0].Type)
				}
			}
		})
	}
}

// TestEntityExtractor_ExtractWithContext tests ExtractWithContext method
func TestEntityExtractor_ExtractWithContext(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		previousEntities []agent.Entity
		mockLLM          func(ctrl *gomock.Controller) *mocks.MockLLMClient
		wantEntities     []agent.Entity
		wantErr          bool
	}{
		{
			name:  "TC-EE-04: ExtractWithContext successfully",
			input: "它的总额",
			previousEntities: []agent.Entity{
				{Type: agent.EntityTypeDocumentType, Value: "purchase_order", RawText: "采购单"},
			},
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				response := `{
					"entities": [
						{"type":"document_type","value":"purchase_order","raw_text":"采购单","confidence":0.9}
					]
				}`
				mock.EXPECT().Generate(gomock.Any(), gomock.Any()).Return(response, nil)
				return mock
			},
			wantEntities: []agent.Entity{
				{Type: agent.EntityTypeDocumentType, Value: "purchase_order", RawText: "采购单", Confidence: 0.9},
			},
			wantErr: false,
		},
		{
			name:  "TC-EE-05: ExtractWithContext LLM failure - fallback",
			input: "查询采购单",
			previousEntities: []agent.Entity{},
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				mock.EXPECT().Generate(gomock.Any(), gomock.Any()).Return("", errors.New("LLM error"))
				return mock
			},
			wantEntities: []agent.Entity{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLLM := tt.mockLLM(ctrl)
			extractor := NewEntityExtractor(mockLLM)

			// Act
			result, err := extractor.ExtractWithContext(context.Background(), tt.input, tt.previousEntities)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.wantEntities), len(result))
			}
		})
	}
}

// TestEntityExtractor_ResolveAnaphora tests ResolveAnaphora method
func TestEntityExtractor_ResolveAnaphora(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		previousEntities []agent.Entity
		mockLLM          func(ctrl *gomock.Controller) *mocks.MockLLMClient
		wantEntities     []agent.Entity
		wantErr          bool
	}{
		{
			name:  "TC-EE-06: Resolve anaphora successfully",
			input: "它的金额是多少",
			previousEntities: []agent.Entity{
				{Type: agent.EntityTypeDocumentType, Value: "purchase_order", RawText: "采购单"},
				{Type: agent.EntityTypeSupplier, Value: "A公司", RawText: "A公司"},
			},
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				response := `{
					"entities": [
						{"type":"document_type","value":"purchase_order","raw_text":"采购单","confidence":0.9},
						{"type":"supplier","value":"A公司","raw_text":"A公司","confidence":0.8}
					]
				}`
				mock.EXPECT().Generate(gomock.Any(), gomock.Any()).Return(response, nil)
				return mock
			},
			wantEntities: []agent.Entity{
				{Type: agent.EntityTypeDocumentType, Value: "purchase_order", Confidence: 0.9},
				{Type: agent.EntityTypeSupplier, Value: "A公司", Confidence: 0.8},
			},
			wantErr: false,
		},
		{
			name:  "TC-EE-07: No pronoun - use normal Extract",
			input: "采购单的金额",
			previousEntities: []agent.Entity{},
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				mock.EXPECT().Generate(gomock.Any(), gomock.Any()).Return("", errors.New("LLM error"))
				return mock
			},
			wantEntities: []agent.Entity{},
			wantErr: false,
		},
		{
			name:  "TC-EE-08: Resolve anaphora LLM failure - fallback",
			input: "这个采购单的金额",
			previousEntities: []agent.Entity{},
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				mock.EXPECT().Generate(gomock.Any(), gomock.Any()).Return("", errors.New("LLM error")).Times(2)
				return mock
			},
			wantEntities: []agent.Entity{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLLM := tt.mockLLM(ctrl)
			extractor := NewEntityExtractor(mockLLM)

			// Act
			result, err := extractor.ResolveAnaphora(context.Background(), tt.input, tt.previousEntities)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.wantEntities), len(result))
			}
		})
	}
}

// TestExtractTimeRange tests extractTimeRange function
func TestExtractTimeRange(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantEntity   *agent.Entity
		wantNil      bool
	}{
		{
			name:    "TC-EE-09: Extract relative time - last month",
			input:   "上个月的采购单",
			wantEntity: &agent.Entity{
				Type:       agent.EntityTypeTimeRange,
				Normalized: "last_month",
				Confidence: 0.8,
			},
			wantNil: false,
		},
		{
			name:    "TC-EE-10: Extract relative time - this month",
			input:   "本月的采购单",
			wantEntity: &agent.Entity{
				Type:       agent.EntityTypeTimeRange,
				Normalized: "this_month",
				Confidence: 0.8,
			},
			wantNil: false,
		},
		{
			name:    "TC-EE-11: Extract relative time - today",
			input:   "今天的采购单",
			wantEntity: &agent.Entity{
				Type:       agent.EntityTypeTimeRange,
				Normalized: "today",
				Confidence: 0.8,
			},
			wantNil: false,
		},
		{
			name:    "TC-EE-12: Extract relative time - this week",
			input:   "本周的采购单",
			wantEntity: &agent.Entity{
				Type:       agent.EntityTypeTimeRange,
				Normalized: "this_week",
				Confidence: 0.8,
			},
			wantNil: false,
		},
		{
			name:    "TC-EE-13: Extract relative time - this year",
			input:   "今年的采购单",
			wantEntity: &agent.Entity{
				Type:       agent.EntityTypeTimeRange,
				Normalized: "this_year",
				Confidence: 0.8,
			},
			wantNil: false,
		},
		{
			name:    "TC-EE-14: Extract absolute date",
			input:   "2024-01-15的采购单",
			wantEntity: &agent.Entity{
				Type:       agent.EntityTypeTimeRange,
				Normalized: "absolute_date",
				Confidence: 0.9,
			},
			wantNil: false,
		},
		{
			name:    "TC-EE-15: Extract date range",
			input:   "从2024-01-01到2024-12-31的采购单",
			wantEntity: &agent.Entity{
				Type:       agent.EntityTypeTimeRange,
				Normalized: "date_range",
				Confidence: 0.9,
			},
			wantNil: false,
		},
		{
			name:    "TC-EE-16: No time range found",
			input:   "采购单明细",
			wantEntity: nil,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := extractTimeRange(tt.input)

			// Assert
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantEntity.Type, result.Type)
				assert.Equal(t, tt.wantEntity.Normalized, result.Normalized)
				assert.Equal(t, tt.wantEntity.Confidence, result.Confidence)
			}
		})
	}
}

// TestExtractAmount tests extractAmount function
func TestExtractAmount(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantEntity   *agent.Entity
		wantNil      bool
	}{
		{
			name:    "TC-EE-17: Extract amount - greater than",
			input:   "金额大于5000的采购单",
			wantEntity: &agent.Entity{
				Type:       agent.EntityTypeAmount,
				Normalized: "amount_condition",
				Confidence: 0.8,
			},
			wantNil: false,
		},
		{
			name:    "TC-EE-18: Extract amount - less than",
			input:   "金额小于10000的采购单",
			wantEntity: &agent.Entity{
				Type:       agent.EntityTypeAmount,
				Normalized: "amount_condition",
				Confidence: 0.8,
			},
			wantNil: false,
		},
		{
			name:    "TC-EE-19: Extract amount - equal",
			input:   "金额等于5000的采购单",
			wantEntity: &agent.Entity{
				Type:       agent.EntityTypeAmount,
				Normalized: "amount_condition",
				Confidence: 0.8,
			},
			wantNil: false,
		},
		{
			name:    "TC-EE-20: Extract amount - above",
			input:   "5000以上的采购单",
			wantEntity: &agent.Entity{
				Type:       agent.EntityTypeAmount,
				Normalized: "amount_condition",
				Confidence: 0.8,
			},
			wantNil: false,
		},
		{
			name:    "TC-EE-21: Extract amount - below",
			input:   "10000以下的采购单",
			wantEntity: &agent.Entity{
				Type:       agent.EntityTypeAmount,
				Normalized: "amount_condition",
				Confidence: 0.8,
			},
			wantNil: false,
		},
		{
			name:    "TC-EE-22: Extract amount range",
			input:   "5000到10000之间的采购单",
			wantEntity: &agent.Entity{
				Type:       agent.EntityTypeAmount,
				Normalized: "amount_range",
				Confidence: 0.8,
			},
			wantNil: false,
		},
		{
			name:    "TC-EE-23: No amount found",
			input:   "采购单明细",
			wantEntity: nil,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := extractAmount(tt.input)

			// Assert
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantEntity.Type, result.Type)
				assert.Equal(t, tt.wantEntity.Normalized, result.Normalized)
				assert.Equal(t, tt.wantEntity.Confidence, result.Confidence)
			}
		})
	}
}

// TestExtractDocumentType tests extractDocumentType function
func TestExtractDocumentType(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantEntity   *agent.Entity
		wantNil      bool
	}{
		{
			name:    "TC-EE-24: Extract document type - purchase order",
			input:   "采购单明细",
			wantEntity: &agent.Entity{
				Type:       agent.EntityTypeDocumentType,
				Value:      "purchase_order",
				Normalized: "purchase_order",
				Confidence: 0.9,
			},
			wantNil: false,
		},
		{
			name:    "TC-EE-25: Extract document type - sales order",
			input:   "销售单明细",
			wantEntity: &agent.Entity{
				Type:       agent.EntityTypeDocumentType,
				Value:      "sales_order",
				Normalized: "sales_order",
				Confidence: 0.9,
			},
			wantNil: false,
		},
		{
			name:    "TC-EE-26: Extract document type - payment",
			input:   "付款单明细",
			wantEntity: &agent.Entity{
				Type:       agent.EntityTypeDocumentType,
				Value:      "payment",
				Normalized: "payment",
				Confidence: 0.9,
			},
			wantNil: false,
		},
		{
			name:    "TC-EE-27: Extract document type - invoice",
			input:   "发票明细",
			wantEntity: &agent.Entity{
				Type:       agent.EntityTypeDocumentType,
				Value:      "invoice",
				Normalized: "invoice",
				Confidence: 0.9,
			},
			wantNil: false,
		},
		{
			name:    "TC-EE-28: No document type found",
			input:   "查询数据",
			wantEntity: nil,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := extractDocumentType(tt.input)

			// Assert
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantEntity.Type, result.Type)
				assert.Equal(t, tt.wantEntity.Value, result.Value)
				assert.Equal(t, tt.wantEntity.Normalized, result.Normalized)
				assert.Equal(t, tt.wantEntity.Confidence, result.Confidence)
			}
		})
	}
}

// TestExtractSupplier tests extractSupplier function
func TestExtractSupplier(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantEntity   *agent.Entity
		wantNil      bool
	}{
		{
			name:    "TC-EE-29: Extract supplier - A公司",
			input:   "A公司的采购单",
			wantEntity: &agent.Entity{
				Type:       agent.EntityTypeSupplier,
				Value:      "A公司",
				Normalized: "supplier",
				Confidence: 0.7,
			},
			wantNil: false,
		},
		{
			name:    "TC-EE-30: Extract supplier - 供应商B",
			input:   "供应商B的采购单",
			wantEntity: &agent.Entity{
				Type:       agent.EntityTypeSupplier,
				Normalized: "supplier",
				Confidence: 0.7,
			},
			wantNil: false,
		},
		{
			name:    "TC-EE-31: No supplier found",
			input:   "采购单明细",
			wantEntity: nil,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := extractSupplier(tt.input)

			// Assert
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantEntity.Type, result.Type)
				assert.Equal(t, tt.wantEntity.Normalized, result.Normalized)
				assert.Equal(t, tt.wantEntity.Confidence, result.Confidence)
			}
		})
	}
}

// TestExtractStatus tests extractStatus function
func TestExtractStatus(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantEntity   *agent.Entity
		wantNil      bool
	}{
		{
			name:    "TC-EE-32: Extract status - completed",
			input:   "已完成的采购单",
			wantEntity: &agent.Entity{
				Type:       agent.EntityTypeStatus,
				Value:      "completed",
				Normalized: "completed",
				Confidence: 0.9,
			},
			wantNil: false,
		},
		{
			name:    "TC-EE-33: Extract status - pending approval",
			input:   "待审批的采购单",
			wantEntity: &agent.Entity{
				Type:       agent.EntityTypeStatus,
				Value:      "pending_approval",
				Normalized: "pending_approval",
				Confidence: 0.9,
			},
			wantNil: false,
		},
		{
			name:    "TC-EE-34: Extract status - cancelled",
			input:   "已取消的采购单",
			wantEntity: &agent.Entity{
				Type:       agent.EntityTypeStatus,
				Value:      "cancelled",
				Normalized: "cancelled",
				Confidence: 0.9,
			},
			wantNil: false,
		},
		{
			name:    "TC-EE-35: No status found",
			input:   "采购单明细",
			wantEntity: nil,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := extractStatus(tt.input)

			// Assert
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantEntity.Type, result.Type)
				assert.Equal(t, tt.wantEntity.Value, result.Value)
				assert.Equal(t, tt.wantEntity.Normalized, result.Normalized)
				assert.Equal(t, tt.wantEntity.Confidence, result.Confidence)
			}
		})
	}
}

// TestNormalizeTimeRange tests normalizeTimeRange function
func TestNormalizeTimeRange(t *testing.T) {
	tests := []struct {
		name          string
		normalized    string
		wantContains  string
	}{
		{
			name:         "TC-NORM-01: Normalize today",
			normalized:   "today",
			wantContains: time.Now().Format("2006-01-02"),
		},
		{
			name:         "TC-NORM-02: Normalize yesterday",
			normalized:   "yesterday",
			wantContains: time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
		},
		{
			name:         "TC-NORM-03: Normalize this_month",
			normalized:   "this_month",
			wantContains: "to",
		},
		{
			name:         "TC-NORM-04: Normalize last_month",
			normalized:   "last_month",
			wantContains: "to",
		},
		{
			name:         "TC-NORM-05: Normalize this_year",
			normalized:   "this_year",
			wantContains: "to",
		},
		{
			name:         "TC-NORM-06: Normalize last_year",
			normalized:   "last_year",
			wantContains: "to",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := normalizeTimeRange(tt.normalized)

			// Assert
			assert.Contains(t, result, tt.wantContains)
		})
	}
}

// TestNormalizeDate tests normalizeDate function
func TestNormalizeDate(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantOutput   string
	}{
		{
			name:       "TC-NORM-07: Normalize date with Chinese characters",
			input:      "2024年01月15日",
			wantOutput: "2024-01-15",
		},
		{
			name:       "TC-NORM-08: Normalize date with hyphens",
			input:      "2024-01-15",
			wantOutput: "2024-01-15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := normalizeDate(tt.input)

			// Assert
			assert.Equal(t, tt.wantOutput, result)
		})
	}
}

// TestContainsPronoun tests containsPronoun function
func TestContainsPronoun(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantContains  bool
	}{
		{
			name:         "TC-PRON-01: Contains '它'",
			input:        "它的金额是多少",
			wantContains: true,
		},
		{
			name:         "TC-PRON-02: Contains '这个'",
			input:        "这个采购单的金额",
			wantContains: true,
		},
		{
			name:         "TC-PRON-03: Contains '那个'",
			input:        "那个采购单的金额",
			wantContains: true,
		},
		{
			name:         "TC-PRON-04: No pronoun",
			input:        "采购单的金额",
			wantContains: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := containsPronoun(tt.input)

			// Assert
			assert.Equal(t, tt.wantContains, result)
		})
	}
}

// TestParseEntityResponse tests parseEntityResponse function
func TestParseEntityResponse(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantEntities  []agent.Entity
		wantErr       bool
	}{
		{
			name:  "TC-PARSE-01: Valid JSON",
			input: `{"entities":[{"type":"time_range","value":"2024-05-01","raw_text":"上个月","confidence":0.8}]}`,
			wantEntities: []agent.Entity{
				{Type: agent.EntityTypeTimeRange, Value: "2024-05-01", RawText: "上个月", Confidence: 0.8},
			},
			wantErr: false,
		},
		{
			name:         "TC-PARSE-02: Invalid JSON",
			input:        "invalid json",
			wantEntities: nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result, err := parseEntityResponse(tt.input)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.wantEntities), len(result))
				if len(result) > 0 {
					assert.Equal(t, tt.wantEntities[0].Type, result[0].Type)
				}
			}
		})
	}
}

// TestBuildEntityPrompt tests buildEntityPrompt function
func TestBuildEntityPrompt(t *testing.T) {
	input := "测试查询"
	result := buildEntityPrompt(input)
	assert.Contains(t, result, input)
	assert.Contains(t, result, "实体类型")
	assert.Contains(t, result, "JSON")
}

// TestBuildEntityPromptWithContext tests buildEntityPromptWithContext function
func TestBuildEntityPromptWithContext(t *testing.T) {
	input := "测试查询"
	previousEntities := []agent.Entity{
		{Type: agent.EntityTypeDocumentType, Value: "purchase_order", RawText: "采购单"},
	}
	result := buildEntityPromptWithContext(input, previousEntities)
	assert.Contains(t, result, input)
	assert.Contains(t, result, "之前提取的实体")
	assert.Contains(t, result, "purchase_order")
}

// TestBuildAnaphoraResolutionPrompt tests buildAnaphoraResolutionPrompt function
func TestBuildAnaphoraResolutionPrompt(t *testing.T) {
	input := "它的金额"
	previousEntities := []agent.Entity{
		{Type: agent.EntityTypeDocumentType, Value: "purchase_order", RawText: "采购单"},
	}
	result := buildAnaphoraResolutionPrompt(input, previousEntities)
	assert.Contains(t, result, input)
	assert.Contains(t, result, "代词")
	assert.Contains(t, result, "purchase_order")
}