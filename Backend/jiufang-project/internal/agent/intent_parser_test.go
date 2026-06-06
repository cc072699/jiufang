// Package agent implements the AI Agent for semantic understanding and SQL generation.
// This file implements unit tests for IntentParser.
// Author: AI Assistant
// Date: 2026-06-03
// Tested Object: IntentParser
// Function: Intent parsing from natural language input

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

// TestIntentParser_Parse tests Parse method
func TestIntentParser_Parse(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		mockLLM     func(ctrl *gomock.Controller) *mocks.MockLLMClient
		wantIntent  *agent.Intent
		wantErr     bool
		errContains string
	}{
		{
			name:  "TC-IP-01: Parse successfully with LLM",
			input: "上个月采购总额是多少",
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				response := `{"intent_type":"statistics","confidence":0.85,"description":"统计查询"}`
				mock.EXPECT().Generate(gomock.Any(), gomock.Any()).Return(response, nil)
				return mock
			},
			wantIntent: &agent.Intent{
				Type:        agent.IntentTypeStatistics,
				Confidence:  0.85,
				Description: "统计查询",
				RawInput:    "上个月采购总额是多少",
			},
			wantErr: false,
		},
		{
			name:  "TC-IP-02: Parse empty input - fallback to template",
			input: "",
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				mock.EXPECT().Generate(gomock.Any(), gomock.Any()).Return("", errors.New("LLM error"))
				return mock
			},
			wantIntent: &agent.Intent{
				Type:        agent.IntentTypeUnknown,
				Confidence:  0.3,
				Description: "无法识别的查询意图",
				RawInput:    "",
			},
			wantErr: false,
		},
		{
			name:  "TC-IP-03: LLM failure - fallback to template matching",
			input: "统计查询",
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				mock.EXPECT().Generate(gomock.Any(), gomock.Any()).Return("", errors.New("LLM connection error"))
				return mock
			},
			wantIntent: &agent.Intent{
				Type:        agent.IntentTypeStatistics,
				Confidence:  0.6,
				Description: "统计查询",
				RawInput:    "统计查询",
			},
			wantErr: false,
		},
		{
			name:  "TC-IP-04: LLM returns invalid JSON - fallback to template",
			input: "查询",
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				mock.EXPECT().Generate(gomock.Any(), gomock.Any()).Return("invalid json", nil)
				return mock
			},
			wantIntent: &agent.Intent{
				Type:        agent.IntentTypeUnknown,
				Confidence:  0.3,
				Description: "无法识别的查询意图",
				RawInput:    "查询",
			},
			wantErr: false,
		},
		{
			name:  "TC-IP-05: Template matching - trend keywords",
			input: "最近6个月销售趋势",
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				mock.EXPECT().Generate(gomock.Any(), gomock.Any()).Return("", errors.New("LLM error"))
				return mock
			},
			wantIntent: &agent.Intent{
				Type:        agent.IntentTypeTrend,
				Confidence:  0.6,
				Description: "趋势查询",
				RawInput:    "最近6个月销售趋势",
			},
			wantErr: false,
		},
		{
			name:  "TC-IP-06: Template matching - comparison keywords",
			input: "对比Q1和Q2销售额",
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				mock.EXPECT().Generate(gomock.Any(), gomock.Any()).Return("", errors.New("LLM error"))
				return mock
			},
			wantIntent: &agent.Intent{
				Type:        agent.IntentTypeComparison,
				Confidence:  0.6,
				Description: "对比查询",
				RawInput:    "对比Q1和Q2销售额",
			},
			wantErr: false,
		},
		{
			name:  "TC-IP-07: Template matching - detail keywords",
			input: "显示所有采购单",
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				mock.EXPECT().Generate(gomock.Any(), gomock.Any()).Return("", errors.New("LLM error"))
				return mock
			},
			wantIntent: &agent.Intent{
				Type:        agent.IntentTypeDetail,
				Confidence:  0.6,
				Description: "明细查询",
				RawInput:    "显示所有采购单",
			},
			wantErr: false,
		},
		{
			name:  "TC-IP-08: Template matching - statistics keywords with '多少'",
			input: "采购单有多少",
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				mock.EXPECT().Generate(gomock.Any(), gomock.Any()).Return("", errors.New("LLM error"))
				return mock
			},
			wantIntent: &agent.Intent{
				Type:        agent.IntentTypeStatistics,
				Confidence:  0.6,
				Description: "统计查询",
				RawInput:    "采购单有多少",
			},
			wantErr: false,
		},
		{
			name:  "TC-IP-09: LLM returns JSON with markdown code block",
			input: "上个月采购总额",
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				response := "```json\n{\"intent_type\":\"statistics\",\"confidence\":0.9,\"description\":\"统计查询\"}\n```"
				mock.EXPECT().Generate(gomock.Any(), gomock.Any()).Return(response, nil)
				return mock
			},
			wantIntent: &agent.Intent{
				Type:        agent.IntentTypeStatistics,
				Confidence:  0.9,
				Description: "统计查询",
				RawInput:    "上个月采购总额",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLLM := tt.mockLLM(ctrl)
			parser := NewIntentParser(mockLLM)

			// Act
			result, err := parser.Parse(context.Background(), tt.input)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantIntent.Type, result.Type)
				assert.Equal(t, tt.wantIntent.Confidence, result.Confidence)
				assert.Equal(t, tt.wantIntent.Description, result.Description)
				assert.Equal(t, tt.wantIntent.RawInput, result.RawInput)
			}
		})
	}
}

// TestIntentParser_ParseWithEntities tests ParseWithEntities method
func TestIntentParser_ParseWithEntities(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		mockLLM    func(ctrl *gomock.Controller) *mocks.MockLLMClient
		wantResult *agent.IntentParserResult
		wantErr    bool
	}{
		{
			name:  "TC-IP-10: ParseWithEntities successfully",
			input: "上个月采购总额大于10000",
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				response := `{
					"intent": {"type":"statistics","confidence":0.85,"description":"统计查询"},
					"entities": [
						{"type":"time_range","value":"2024-05-01 to 2024-05-31","raw_text":"上个月","confidence":0.8},
						{"type":"amount","value":"> 10000","raw_text":"大于10000","confidence":0.9}
					]
				}`
				mock.EXPECT().Generate(gomock.Any(), gomock.Any()).Return(response, nil)
				return mock
			},
			wantResult: &agent.IntentParserResult{
				Intent: agent.Intent{
					Type:        agent.IntentTypeStatistics,
					Confidence:  0.85,
					Description: "统计查询",
					RawInput:    "上个月采购总额大于10000",
				},
				Entities: []agent.Entity{
					{Type: agent.EntityTypeTimeRange, Value: "2024-05-01 to 2024-05-31", RawText: "上个月", Confidence: 0.8},
					{Type: agent.EntityTypeAmount, Value: "> 10000", RawText: "大于10000", Confidence: 0.9},
				},
			},
			wantErr: false,
		},
		{
			name:  "TC-IP-11: ParseWithEntities LLM failure - fallback",
			input: "查询采购单",
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				mock.EXPECT().Generate(gomock.Any(), gomock.Any()).Return("", errors.New("LLM error"))
				return mock
			},
			wantResult: &agent.IntentParserResult{
				Intent: agent.Intent{
					Type:        agent.IntentTypeUnknown,
					Confidence:  0.3,
					Description: "无法识别的查询意图",
					RawInput:    "查询采购单",
				},
				Entities: []agent.Entity{},
			},
			wantErr: false,
		},
		{
			name:  "TC-IP-12: ParseWithEntities invalid JSON - fallback",
			input: "测试查询",
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				mock.EXPECT().Generate(gomock.Any(), gomock.Any()).Return("invalid", nil)
				return mock
			},
			wantResult: &agent.IntentParserResult{
				Intent: agent.Intent{
					Type:        agent.IntentTypeUnknown,
					Confidence:  0.3,
					Description: "无法识别的查询意图",
					RawInput:    "测试查询",
				},
				Entities: []agent.Entity{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLLM := tt.mockLLM(ctrl)
			parser := NewIntentParser(mockLLM)

			// Act
			result, err := parser.ParseWithEntities(context.Background(), tt.input)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResult.Intent.Type, result.Intent.Type)
				assert.Equal(t, tt.wantResult.Intent.Confidence, result.Intent.Confidence)
				assert.Equal(t, tt.wantResult.Intent.Description, result.Intent.Description)
				assert.Equal(t, tt.wantResult.Intent.RawInput, result.Intent.RawInput)
				assert.Equal(t, len(tt.wantResult.Entities), len(result.Entities))
			}
		})
	}
}

// TestIntentParser_parseWithTemplate tests parseWithTemplate method
func TestIntentParser_parseWithTemplate(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantIntent *agent.Intent
	}{
		{
			name:  "TC-IP-13: Template matching - statistics keyword '多少'",
			input: "采购单有多少",
			wantIntent: &agent.Intent{
				Type:        agent.IntentTypeStatistics,
				Confidence:  0.6,
				Description: "统计查询",
			},
		},
		{
			name:  "TC-IP-14: Template matching - statistics keyword '统计'",
			input: "统计采购金额",
			wantIntent: &agent.Intent{
				Type:        agent.IntentTypeStatistics,
				Confidence:  0.6,
				Description: "统计查询",
			},
		},
		{
			name:  "TC-IP-15: Template matching - trend keyword '趋势'",
			input: "销售趋势分析",
			wantIntent: &agent.Intent{
				Type:        agent.IntentTypeTrend,
				Confidence:  0.6,
				Description: "趋势查询",
			},
		},
		{
			name:  "TC-IP-16: Template matching - comparison keyword '对比'",
			input: "对比两个月的销售额",
			wantIntent: &agent.Intent{
				Type:        agent.IntentTypeComparison,
				Confidence:  0.6,
				Description: "对比查询",
			},
		},
		{
			name:  "TC-IP-17: Template matching - detail keyword '查看'",
			input: "查看采购单明细",
			wantIntent: &agent.Intent{
				Type:        agent.IntentTypeDetail,
				Confidence:  0.6,
				Description: "明细查询",
			},
		},
		{
			name:  "TC-IP-18: Template matching - unknown intent",
			input: "随便说说",
			wantIntent: &agent.Intent{
				Type:        agent.IntentTypeUnknown,
				Confidence:  0.3,
				Description: "无法识别的查询意图",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLLM := mocks.NewMockLLMClient(ctrl)
			parser := NewIntentParser(mockLLM)

			// Act
			result := parser.parseWithTemplate(tt.input)

			// Assert
			assert.Equal(t, tt.wantIntent.Type, result.Type)
			assert.Equal(t, tt.wantIntent.Confidence, result.Confidence)
			assert.Equal(t, tt.wantIntent.Description, result.Description)
		})
	}
}

// TestIntentParserWithSchema_Parse tests IntentParserWithSchema
func TestIntentParserWithSchema_Parse(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		mockLLM    func(ctrl *gomock.Controller) *mocks.MockLLMClient
		wantIntent *agent.Intent
		wantErr    bool
	}{
		{
			name:  "TC-IP-19: ParseWithSchema successfully",
			input: "上个月采购总额",
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				result := map[string]interface{}{
					"intent_type": "statistics",
					"confidence":  0.9,
					"description": "统计查询",
				}
				mock.EXPECT().GenerateWithSchema(gomock.Any(), gomock.Any(), gomock.Any()).Return(result, nil)
				return mock
			},
			wantIntent: &agent.Intent{
				Type:        agent.IntentTypeStatistics,
				Confidence:  0.9,
				Description: "统计查询",
				RawInput:    "上个月采购总额",
			},
			wantErr: false,
		},
		{
			name:  "TC-IP-20: ParseWithSchema LLM failure",
			input: "测试查询",
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				mock.EXPECT().GenerateWithSchema(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("LLM error"))
				return mock
			},
			wantIntent: nil,
			wantErr:    true,
		},
		{
			name:  "TC-IP-21: ParseWithSchema unexpected result type",
			input: "测试",
			mockLLM: func(ctrl *gomock.Controller) *mocks.MockLLMClient {
				mock := mocks.NewMockLLMClient(ctrl)
				mock.EXPECT().GenerateWithSchema(gomock.Any(), gomock.Any(), gomock.Any()).Return("invalid type", nil)
				return mock
			},
			wantIntent: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLLM := tt.mockLLM(ctrl)
			parser := NewIntentParserWithSchema(mockLLM)

			// Act
			result, err := parser.Parse(context.Background(), tt.input)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantIntent.Type, result.Type)
				assert.Equal(t, tt.wantIntent.Confidence, result.Confidence)
				assert.Equal(t, tt.wantIntent.Description, result.Description)
				assert.Equal(t, tt.wantIntent.RawInput, result.RawInput)
			}
		})
	}
}

// TestIntent_IsConfident tests Intent.IsConfident method
func TestIntent_IsConfident(t *testing.T) {
	tests := []struct {
		name          string
		confidence    float64
		wantConfident bool
	}{
		{
			name:          "TC-INT-01: Confidence above threshold",
			confidence:    0.8,
			wantConfident: true,
		},
		{
			name:          "TC-INT-02: Confidence at threshold",
			confidence:    0.7,
			wantConfident: true,
		},
		{
			name:          "TC-INT-03: Confidence below threshold",
			confidence:    0.5,
			wantConfident: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			intent := &agent.Intent{Confidence: tt.confidence}
			result := intent.IsConfident()
			assert.Equal(t, tt.wantConfident, result)
		})
	}
}

// TestIntent_NeedsClarification tests Intent.NeedsClarification method
func TestIntent_NeedsClarification(t *testing.T) {
	tests := []struct {
		name              string
		intentType        agent.IntentType
		confidence        float64
		wantClarification bool
	}{
		{
			name:              "TC-INT-04: High confidence known intent",
			intentType:        agent.IntentTypeStatistics,
			confidence:        0.9,
			wantClarification: false,
		},
		{
			name:              "TC-INT-05: Low confidence known intent",
			intentType:        agent.IntentTypeStatistics,
			confidence:        0.5,
			wantClarification: true,
		},
		{
			name:              "TC-INT-06: Unknown intent type",
			intentType:        agent.IntentTypeUnknown,
			confidence:        0.9,
			wantClarification: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			intent := &agent.Intent{
				Type:       tt.intentType,
				Confidence: tt.confidence,
			}
			result := intent.NeedsClarification()
			assert.Equal(t, tt.wantClarification, result)
		})
	}
}

// TestCleanJSONResponse tests cleanJSONResponse function
func TestCleanJSONResponse(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantOutput string
	}{
		{
			name:       "TC-CLEAN-01: Remove json code block",
			input:      "```json\n{\"test\": \"value\"}\n```",
			wantOutput: "{\"test\": \"value\"}",
		},
		{
			name:       "TC-CLEAN-02: Remove plain code block",
			input:      "```{\"test\": \"value\"}```",
			wantOutput: "{\"test\": \"value\"}",
		},
		{
			name:       "TC-CLEAN-03: No code block",
			input:      "{\"test\": \"value\"}",
			wantOutput: "{\"test\": \"value\"}",
		},
		{
			name:       "TC-CLEAN-04: Trim whitespace",
			input:      "  {\"test\": \"value\"}  ",
			wantOutput: "{\"test\": \"value\"}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanJSONResponse(tt.input)
			assert.Equal(t, tt.wantOutput, result)
		})
	}
}

// TestParseIntentResponse tests parseIntentResponse function
func TestParseIntentResponse(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantIntent *agent.Intent
		wantErr    bool
	}{
		{
			name:  "TC-PARSE-01: Valid JSON",
			input: "{\"intent_type\":\"statistics\",\"confidence\":0.85,\"description\":\"统计查询\"}",
			wantIntent: &agent.Intent{
				Type:        agent.IntentTypeStatistics,
				Confidence:  0.85,
				Description: "统计查询",
			},
			wantErr: false,
		},
		{
			name:       "TC-PARSE-02: Invalid JSON",
			input:      "invalid json",
			wantIntent: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseIntentResponse(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantIntent.Type, result.Type)
				assert.Equal(t, tt.wantIntent.Confidence, result.Confidence)
				assert.Equal(t, tt.wantIntent.Description, result.Description)
			}
		})
	}
}

// TestParseIntentAndEntityResponse tests parseIntentAndEntityResponse function
func TestParseIntentAndEntityResponse(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantResult *agent.IntentParserResult
		wantErr    bool
	}{
		{
			name: "TC-PARSE-03: Valid JSON with entities",
			input: `{
				"intent": {"type":"statistics","confidence":0.85,"description":"统计查询"},
				"entities": [{"type":"time_range","value":"2024-05-01","raw_text":"上个月","confidence":0.8}]
			}`,
			wantResult: &agent.IntentParserResult{
				Intent: agent.Intent{
					Type:        agent.IntentTypeStatistics,
					Confidence:  0.85,
					Description: "统计查询",
				},
				Entities: []agent.Entity{
					{Type: agent.EntityTypeTimeRange, Value: "2024-05-01", RawText: "上个月", Confidence: 0.8},
				},
			},
			wantErr: false,
		},
		{
			name:       "TC-PARSE-04: Invalid JSON",
			input:      "invalid",
			wantResult: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseIntentAndEntityResponse(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResult.Intent.Type, result.Intent.Type)
				assert.Equal(t, len(tt.wantResult.Entities), len(result.Entities))
			}
		})
	}
}

// TestBuildIntentPrompt tests buildIntentPrompt function
func TestBuildIntentPrompt(t *testing.T) {
	input := "测试查询"
	result := buildIntentPrompt(input)
	assert.Contains(t, result, input)
	assert.Contains(t, result, "意图类型")
	assert.Contains(t, result, "JSON")
}

// TestBuildIntentAndEntityPrompt tests buildIntentAndEntityPrompt function
func TestBuildIntentAndEntityPrompt(t *testing.T) {
	input := "测试查询"
	result := buildIntentAndEntityPrompt(input)
	assert.Contains(t, result, input)
	assert.Contains(t, result, "意图类型")
	assert.Contains(t, result, "实体类型")
	assert.Contains(t, result, "JSON")
}

// TestConvertIntentToMessage tests ConvertIntentToMessage function
func TestConvertIntentToMessage(t *testing.T) {
	intent := &agent.Intent{
		Type:        agent.IntentTypeStatistics,
		Confidence:  0.85,
		Description: "统计查询",
	}
	result := ConvertIntentToMessage(intent)
	assert.NotNil(t, result)
	assert.Contains(t, result.Content, "意图类型")
	assert.Contains(t, result.Content, "置信度")
}
