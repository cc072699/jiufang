// Package agent implements the AI Agent for semantic understanding and SQL generation.
// This module handles natural language queries and converts them to executable SQL.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"

	"jiufang/internal/infrastructure/llm"
	"jiufang/internal/model/agent"
)

// IntentParser parses the intent from natural language input.
type IntentParser struct {
	llmClient llm.LLMClientInterface
}

// NewIntentParser creates a new intent parser.
func NewIntentParser(llmClient llm.LLMClientInterface) *IntentParser {
	return &IntentParser{
		llmClient: llmClient,
	}
}

// Parse parses the intent from user input.
func (p *IntentParser) Parse(ctx context.Context, input string) (*agent.Intent, error) {
	// Build prompt for intent classification
	prompt := buildIntentPrompt(input)

	// Call LLM to classify intent
	response, err := p.llmClient.Generate(ctx, prompt)
	if err != nil {
		// Fallback to template matching if LLM fails
		return p.parseWithTemplate(input), nil
	}

	// Parse LLM response
	intent, err := parseIntentResponse(response)
	if err != nil {
		return p.parseWithTemplate(input), nil
	}

	intent.RawInput = input
	return intent, nil
}

// ParseWithEntities parses intent and extracts entities together.
func (p *IntentParser) ParseWithEntities(ctx context.Context, input string) (*agent.IntentParserResult, error) {
	// Build prompt for intent and entity extraction
	prompt := buildIntentAndEntityPrompt(input)

	// Call LLM
	response, err := p.llmClient.Generate(ctx, prompt)
	if err != nil {
		// Fallback to template matching
		intent := p.parseWithTemplate(input)
		return &agent.IntentParserResult{
			Intent:   *intent,
			Entities: []agent.Entity{},
		}, nil
	}

	// Parse response
	result, err := parseIntentAndEntityResponse(response)
	if err != nil {
		intent := p.parseWithTemplate(input)
		return &agent.IntentParserResult{
			Intent:   *intent,
			Entities: []agent.Entity{},
		}, nil
	}

	result.Intent.RawInput = input
	return result, nil
}

// parseWithTemplate parses intent using template matching (fallback mode).
func (p *IntentParser) parseWithTemplate(input string) *agent.Intent {
	inputLower := strings.ToLower(input)

	// Check for statistics keywords
	if strings.Contains(inputLower, "多少") ||
		strings.Contains(inputLower, "统计") ||
		strings.Contains(inputLower, "总计") ||
		strings.Contains(inputLower, "合计") ||
		strings.Contains(inputLower, "总数") ||
		strings.Contains(inputLower, "金额") {
		return &agent.Intent{
			Type:        agent.IntentTypeStatistics,
			Confidence:  0.6,
			Description: "统计查询",
		}
	}

	// Check for trend keywords
	if strings.Contains(inputLower, "趋势") ||
		strings.Contains(inputLower, "走势") ||
		strings.Contains(inputLower, "变化") ||
		strings.Contains(inputLower, "增长") ||
		strings.Contains(inputLower, "下降") {
		return &agent.Intent{
			Type:        agent.IntentTypeTrend,
			Confidence:  0.6,
			Description: "趋势查询",
		}
	}

	// Check for comparison keywords
	if strings.Contains(inputLower, "对比") ||
		strings.Contains(inputLower, "比较") ||
		strings.Contains(inputLower, "相比") ||
		strings.Contains(inputLower, "差异") {
		return &agent.Intent{
			Type:        agent.IntentTypeComparison,
			Confidence:  0.6,
			Description: "对比查询",
		}
	}

	// Check for detail keywords
	if strings.Contains(inputLower, "查看") ||
		strings.Contains(inputLower, "显示") ||
		strings.Contains(inputLower, "列出") ||
		strings.Contains(inputLower, "所有") ||
		strings.Contains(inputLower, "明细") {
		return &agent.Intent{
			Type:        agent.IntentTypeDetail,
			Confidence:  0.6,
			Description: "明细查询",
		}
	}

	// Default to unknown
	return &agent.Intent{
		Type:        agent.IntentTypeUnknown,
		Confidence:  0.3,
		Description: "无法识别的查询意图",
	}
}

// buildIntentPrompt builds the prompt for intent classification.
func buildIntentPrompt(input string) string {
	return fmt.Sprintf(`请分析以下用户查询的意图，并返回JSON格式的结果。

用户查询：%s

意图类型包括：
- statistics: 统计查询（如"上个月采购总额是多少"）
- detail: 明细查询（如"显示A供应商的所有采购单"）
- trend: 趋势查询（如"最近6个月的销售趋势"）
- comparison: 对比查询（如"对比Q1和Q2的销售额"）

请返回以下JSON格式：
{
  "intent_type": "statistics|detail|trend|comparison",
  "confidence": 0.0-1.0,
  "description": "意图描述"
}

只返回JSON，不要其他内容。`, input)
}

// buildIntentAndEntityPrompt builds the prompt for intent and entity extraction.
func buildIntentAndEntityPrompt(input string) string {
	return fmt.Sprintf(`请分析以下用户查询的意图和实体，并返回JSON格式的结果。

用户查询：%s

意图类型包括：
- statistics: 统计查询
- detail: 明细查询
- trend: 趋势查询
- comparison: 对比查询

实体类型包括：
- time_range: 时间范围（如"上个月"、"2024-01-01到2024-12-31"）
- amount: 金额条件（如"大于10000"、"5000到10000之间"）
- document_type: 单据类型（如"采购单"、"销售单"、"付款单"）
- supplier: 供应商（如"A公司"、"供应商B"）
- department: 部门（如"销售部"、"财务部"）
- customer: 客户（如"客户X"）
- product: 产品（如"产品A"）
- status: 状态（如"已完成"、"待审批"）

请返回以下JSON格式：
{
  "intent": {
    "type": "statistics|detail|trend|comparison",
    "confidence": 0.0-1.0,
    "description": "意图描述"
  },
  "entities": [
    {
      "type": "实体类型",
      "value": "标准化值",
      "raw_text": "原始文本",
      "confidence": 0.0-1.0
    }
  ]
}

只返回JSON，不要其他内容。`, input)
}

// parseIntentResponse parses the LLM response for intent.
func parseIntentResponse(response string) (*agent.Intent, error) {
	// Clean response (remove markdown code blocks if present)
	response = cleanJSONResponse(response)

	var intent agent.Intent
	if err := json.Unmarshal([]byte(response), &intent); err != nil {
		return nil, err
	}

	return &intent, nil
}

// parseIntentAndEntityResponse parses the LLM response for intent and entities.
func parseIntentAndEntityResponse(response string) (*agent.IntentParserResult, error) {
	// Clean response
	response = cleanJSONResponse(response)

	var result struct {
		Intent   agent.Intent   `json:"intent"`
		Entities []agent.Entity `json:"entities"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	return &agent.IntentParserResult{
		Intent:   result.Intent,
		Entities: result.Entities,
	}, nil
}

// cleanJSONResponse cleans the JSON response from LLM.
func cleanJSONResponse(response string) string {
	// Remove markdown code blocks
	response = strings.TrimSpace(response)
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimSuffix(response, "```")
	} else if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```")
		response = strings.TrimSuffix(response, "```")
	}
	return strings.TrimSpace(response)
}

// IntentParserWithSchema uses structured output for intent parsing.
type IntentParserWithSchema struct {
	llmClient llm.LLMClientInterface
}

// NewIntentParserWithSchema creates a new intent parser with schema.
func NewIntentParserWithSchema(llmClient llm.LLMClientInterface) *IntentParserWithSchema {
	return &IntentParserWithSchema{
		llmClient: llmClient,
	}
}

// Parse parses intent using structured output.
func (p *IntentParserWithSchema) Parse(ctx context.Context, input string) (*agent.Intent, error) {
	prompt := buildIntentPrompt(input)

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"intent_type": map[string]interface{}{
				"type": "string",
				"enum": []string{"statistics", "detail", "trend", "comparison"},
			},
			"confidence": map[string]interface{}{
				"type": "number",
			},
			"description": map[string]interface{}{
				"type": "string",
			},
		},
		"required": []string{"intent_type", "confidence", "description"},
	}

	result, err := p.llmClient.GenerateWithSchema(ctx, prompt, schema)
	if err != nil {
		return nil, err
	}

	// Convert result to Intent
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result type")
	}

	intentType, _ := resultMap["intent_type"].(string)
	confidence, _ := resultMap["confidence"].(float64)
	description, _ := resultMap["description"].(string)

	return &agent.Intent{
		Type:        agent.IntentType(intentType),
		Confidence:  confidence,
		Description: description,
		RawInput:    input,
	}, nil
}

// ConvertIntentToMessage converts an intent to a schema message.
func ConvertIntentToMessage(intent *agent.Intent) *schema.Message {
	content := fmt.Sprintf("意图类型：%s\n置信度：%.2f\n描述：%s",
		intent.Type, intent.Confidence, intent.Description)
	return schema.AssistantMessage(content, nil)
}