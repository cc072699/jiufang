// Package agent implements the AI Agent for semantic understanding and SQL generation.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"jiufang/internal/infrastructure/llm"
	"jiufang/internal/model/agent"
)

// EntityExtractor extracts entities from natural language input.
type EntityExtractor struct {
	llmClient llm.LLMClientInterface
}

// NewEntityExtractor creates a new entity extractor.
func NewEntityExtractor(llmClient llm.LLMClientInterface) *EntityExtractor {
	return &EntityExtractor{
		llmClient: llmClient,
	}
}

// Extract extracts entities from user input.
func (e *EntityExtractor) Extract(ctx context.Context, input string) ([]agent.Entity, error) {
	// Build prompt for entity extraction
	prompt := buildEntityPrompt(input)

	// Call LLM
	response, err := e.llmClient.Generate(ctx, prompt)
	if err != nil {
		// Fallback to template matching
		return e.extractWithTemplate(input), nil
	}

	// Parse response
	entities, err := parseEntityResponse(response)
	if err != nil {
		return e.extractWithTemplate(input), nil
	}

	return entities, nil
}

// ExtractWithContext extracts entities considering previous context.
func (e *EntityExtractor) ExtractWithContext(ctx context.Context, input string, previousEntities []agent.Entity) ([]agent.Entity, error) {
	// Build prompt with context
	prompt := buildEntityPromptWithContext(input, previousEntities)

	// Call LLM
	response, err := e.llmClient.Generate(ctx, prompt)
	if err != nil {
		return e.extractWithTemplate(input), nil
	}

	// Parse response
	entities, err := parseEntityResponse(response)
	if err != nil {
		return e.extractWithTemplate(input), nil
	}

	return entities, nil
}

// ResolveAnaphora resolves anaphora (pronoun references) in the input.
func (e *EntityExtractor) ResolveAnaphora(ctx context.Context, input string, previousEntities []agent.Entity) ([]agent.Entity, error) {
	// Check if input contains pronouns
	if !containsPronoun(input) {
		return e.Extract(ctx, input)
	}

	// Build prompt for anaphora resolution
	prompt := buildAnaphoraResolutionPrompt(input, previousEntities)

	// Call LLM
	response, err := e.llmClient.Generate(ctx, prompt)
	if err != nil {
		return e.Extract(ctx, input)
	}

	// Parse response
	entities, err := parseEntityResponse(response)
	if err != nil {
		return e.Extract(ctx, input)
	}

	return entities, nil
}

// extractWithTemplate extracts entities using template matching (fallback mode).
func (e *EntityExtractor) extractWithTemplate(input string) []agent.Entity {
	entities := []agent.Entity{}

	// Extract time range
	timeEntity := extractTimeRange(input)
	if timeEntity != nil {
		entities = append(entities, *timeEntity)
	}

	// Extract amount
	amountEntity := extractAmount(input)
	if amountEntity != nil {
		entities = append(entities, *amountEntity)
	}

	// Extract document type
	docEntity := extractDocumentType(input)
	if docEntity != nil {
		entities = append(entities, *docEntity)
	}

	// Extract supplier
	supplierEntity := extractSupplier(input)
	if supplierEntity != nil {
		entities = append(entities, *supplierEntity)
	}

	// Extract status
	statusEntity := extractStatus(input)
	if statusEntity != nil {
		entities = append(entities, *statusEntity)
	}

	return entities
}

// extractTimeRange extracts time range from input.
func extractTimeRange(input string) *agent.Entity {
	inputLower := strings.ToLower(input)

	// Relative time patterns
	timePatterns := map[string]string{
		"今天":     "today",
		"昨天":     "yesterday",
		"本周":     "this_week",
		"上周":     "last_week",
		"本月":     "this_month",
		"上个月":    "last_month",
		"上月":     "last_month",
		"本季度":    "this_quarter",
		"上季度":    "last_quarter",
		"今年":     "this_year",
		"去年":     "last_year",
		"最近":     "recent",
		"近":      "recent",
	}

	for pattern, normalized := range timePatterns {
		if strings.Contains(inputLower, pattern) {
			value := normalizeTimeRange(normalized)
			return &agent.Entity{
				Type:       agent.EntityTypeTimeRange,
				Value:      value,
				RawText:    pattern,
				Normalized: normalized,
				Confidence: 0.8,
			}
		}
	}

	// Absolute date patterns
	datePattern := regexp.MustCompile(`(\d{4}[-年]\d{1,2}[-月]\d{1,2}[日]?)`)
	if datePattern.MatchString(input) {
		date := datePattern.FindString(input)
		normalized := normalizeDate(date)
		return &agent.Entity{
			Type:       agent.EntityTypeTimeRange,
			Value:      normalized,
			RawText:    date,
			Normalized: "absolute_date",
			Confidence: 0.9,
		}
	}

	// Date range patterns
	rangePattern := regexp.MustCompile(`从(.+?)到(.+?)`)
	if rangePattern.MatchString(input) {
		matches := rangePattern.FindStringSubmatch(input)
		startDate := normalizeDate(matches[1])
		endDate := normalizeDate(matches[2])
		value := startDate + " to " + endDate
		return &agent.Entity{
			Type:       agent.EntityTypeTimeRange,
			Value:      value,
			RawText:    matches[0],
			Normalized: "date_range",
			Confidence: 0.9,
		}
	}

	return nil
}

// extractAmount extracts amount condition from input.
func extractAmount(input string) *agent.Entity {
	inputLower := strings.ToLower(input)

	// Amount comparison patterns
	amountPatterns := []struct {
		pattern   *regexp.Regexp
		operator  string
		extractFn func(string) float64
	}{
		{
			pattern:  regexp.MustCompile(`大于(\d+)`),
			operator: ">",
		},
		{
			pattern:  regexp.MustCompile(`超过(\d+)`),
			operator: ">",
		},
		{
			pattern:  regexp.MustCompile(`小于(\d+)`),
			operator: "<",
		},
		{
			pattern:  regexp.MustCompile(`低于(\d+)`),
			operator: "<",
		},
		{
			pattern:  regexp.MustCompile(`等于(\d+)`),
			operator: "=",
		},
		{
			pattern:  regexp.MustCompile(`(\d+)以上`),
			operator: ">=",
		},
		{
			pattern:  regexp.MustCompile(`(\d+)以下`),
			operator: "<=",
		},
	}

	for _, ap := range amountPatterns {
		if ap.pattern.MatchString(inputLower) {
			matches := ap.pattern.FindStringSubmatch(inputLower)
			amount := parseFloat(matches[1])
			value := fmt.Sprintf("%s %.2f", ap.operator, amount)
			return &agent.Entity{
				Type:       agent.EntityTypeAmount,
				Value:      value,
				RawText:    matches[0],
				Normalized: "amount_condition",
				Confidence: 0.8,
			}
		}
	}

	// Amount range pattern
	rangePattern := regexp.MustCompile(`(\d+)到(\d+)之间`)
	if rangePattern.MatchString(inputLower) {
		matches := rangePattern.FindStringSubmatch(inputLower)
		start := parseFloat(matches[1])
		end := parseFloat(matches[2])
		value := fmt.Sprintf("between %.2f and %.2f", start, end)
		return &agent.Entity{
			Type:       agent.EntityTypeAmount,
			Value:      value,
			RawText:    matches[0],
			Normalized: "amount_range",
			Confidence: 0.8,
		}
	}

	return nil
}

// extractDocumentType extracts document type from input.
func extractDocumentType(input string) *agent.Entity {
	inputLower := strings.ToLower(input)

	docTypes := map[string]string{
		"采购单":   "purchase_order",
		"采购订单":  "purchase_order",
		"采购":    "purchase",
		"销售单":   "sales_order",
		"销售订单":  "sales_order",
		"销售":    "sales",
		"付款单":   "payment",
		"付款":    "payment",
		"收款单":   "receipt",
		"收款":    "receipt",
		"入库单":   "inbound",
		"入库":    "inbound",
		"出库单":   "outbound",
		"出库":    "outbound",
		"发票":    "invoice",
		"报销单":   "expense",
		"报销":    "expense",
	}

	for pattern, normalized := range docTypes {
		if strings.Contains(inputLower, pattern) {
			return &agent.Entity{
				Type:       agent.EntityTypeDocumentType,
				Value:      normalized,
				RawText:    pattern,
				Normalized: normalized,
				Confidence: 0.9,
			}
		}
	}

	return nil
}

// extractSupplier extracts supplier from input.
func extractSupplier(input string) *agent.Entity {
	// Pattern: "A公司", "供应商B", "公司C"
	supplierPattern := regexp.MustCompile(`([A-Z]?公司|供应商[A-Z]?|公司[A-Z]?)`)
	if supplierPattern.MatchString(input) {
		matches := supplierPattern.FindStringSubmatch(input)
		return &agent.Entity{
			Type:       agent.EntityTypeSupplier,
			Value:      matches[1],
			RawText:    matches[1],
			Normalized: "supplier",
			Confidence: 0.7,
		}
	}

	return nil
}

// extractStatus extracts status from input.
func extractStatus(input string) *agent.Entity {
	inputLower := strings.ToLower(input)

	statusMap := map[string]string{
		"已完成":   "completed",
		"完成":    "completed",
		"待审批":   "pending_approval",
		"审批中":   "pending_approval",
		"待处理":   "pending",
		"进行中":   "in_progress",
		"已取消":   "cancelled",
		"取消":    "cancelled",
		"已拒绝":   "rejected",
		"拒绝":    "rejected",
		"已通过":   "approved",
		"通过":    "approved",
	}

	for pattern, normalized := range statusMap {
		if strings.Contains(inputLower, pattern) {
			return &agent.Entity{
				Type:       agent.EntityTypeStatus,
				Value:      normalized,
				RawText:    pattern,
				Normalized: normalized,
				Confidence: 0.9,
			}
		}
	}

	return nil
}

// normalizeTimeRange normalizes relative time range to actual dates.
func normalizeTimeRange(normalized string) string {
	now := time.Now()

	switch normalized {
	case "today":
		return formatDate(now)
	case "yesterday":
		return formatDate(now.AddDate(0, 0, -1))
	case "this_week":
		start := now.AddDate(0, 0, -int(now.Weekday()))
		return formatDateRange(start, start.AddDate(0, 0, 6))
	case "last_week":
		start := now.AddDate(0, 0, -int(now.Weekday())-7)
		return formatDateRange(start, start.AddDate(0, 0, 6))
	case "this_month":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end := start.AddDate(0, 1, -1)
		return formatDateRange(start, end)
	case "last_month":
		start := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())
		end := start.AddDate(0, 1, -1)
		return formatDateRange(start, end)
	case "this_quarter":
		quarter := (int(now.Month()) - 1) / 3
		start := time.Date(now.Year(), time.Month(quarter*3+1), 1, 0, 0, 0, 0, now.Location())
		end := start.AddDate(0, 3, -1)
		return formatDateRange(start, end)
	case "last_quarter":
		quarter := (int(now.Month()) - 1) / 3
		start := time.Date(now.Year(), time.Month(quarter*3+1-3), 1, 0, 0, 0, 0, now.Location())
		if quarter == 0 {
			start = time.Date(now.Year()-1, time.Month(10), 1, 0, 0, 0, 0, now.Location())
		}
		end := start.AddDate(0, 3, -1)
		return formatDateRange(start, end)
	case "this_year":
		start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		end := time.Date(now.Year(), 12, 31, 0, 0, 0, 0, now.Location())
		return formatDateRange(start, end)
	case "last_year":
		start := time.Date(now.Year()-1, 1, 1, 0, 0, 0, 0, now.Location())
		end := time.Date(now.Year()-1, 12, 31, 0, 0, 0, 0, now.Location())
		return formatDateRange(start, end)
	default:
		return formatDate(now)
	}
}

// normalizeDate normalizes date string to standard format.
func normalizeDate(date string) string {
	// Replace Chinese characters
	date = strings.ReplaceAll(date, "年", "-")
	date = strings.ReplaceAll(date, "月", "-")
	date = strings.ReplaceAll(date, "日", "")

	// Try to parse and format
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return formatDate(t)
}

// formatDate formats a date to YYYY-MM-DD.
func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// formatDateRange formats a date range.
func formatDateRange(start, end time.Time) string {
	return formatDate(start) + " to " + formatDate(end)
}

// parseFloat parses a float from string.
func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

// containsPronoun checks if input contains pronouns.
func containsPronoun(input string) bool {
	pronouns := []string{"它", "他", "她", "这个", "那个", "这些", "那些", "其"}
	inputLower := strings.ToLower(input)
	for _, pronoun := range pronouns {
		if strings.Contains(inputLower, pronoun) {
			return true
		}
	}
	return false
}

// buildEntityPrompt builds the prompt for entity extraction.
func buildEntityPrompt(input string) string {
	return fmt.Sprintf(`请从以下用户查询中提取实体信息，并返回JSON格式的结果。

用户查询：%s

实体类型包括：
- time_range: 时间范围
- amount: 金额条件
- document_type: 单据类型
- supplier: 供应商
- department: 部门
- customer: 客户
- product: 产品
- status: 状态

请返回以下JSON格式：
{
  "entities": [
    {
      "type": "实体类型",
      "value": "标准化值（如时间范围转换为YYYY-MM-DD格式）",
      "raw_text": "原始文本",
      "normalized": "标准化标识",
      "confidence": 0.0-1.0
    }
  ]
}

只返回JSON，不要其他内容。`, input)
}

// buildEntityPromptWithContext builds the prompt with previous context.
func buildEntityPromptWithContext(input string, previousEntities []agent.Entity) string {
	contextStr := ""
	for _, e := range previousEntities {
		contextStr += fmt.Sprintf("- %s: %s (原始: %s)\n", e.Type, e.Value, e.RawText)
	}

	return fmt.Sprintf(`请从以下用户查询中提取实体信息，考虑之前的对话上下文，并返回JSON格式的结果。

用户查询：%s

之前提取的实体：
%s

请返回以下JSON格式：
{
  "entities": [
    {
      "type": "实体类型",
      "value": "标准化值",
      "raw_text": "原始文本",
      "normalized": "标准化标识",
      "confidence": 0.0-1.0
    }
  ]
}

只返回JSON，不要其他内容。`, input, contextStr)
}

// buildAnaphoraResolutionPrompt builds the prompt for anaphora resolution.
func buildAnaphoraResolutionPrompt(input string, previousEntities []agent.Entity) string {
	contextStr := ""
	for _, e := range previousEntities {
		contextStr += fmt.Sprintf("- %s: %s (原始: %s)\n", e.Type, e.Value, e.RawText)
	}

	return fmt.Sprintf(`请解析以下查询中的代词引用，并返回完整的实体信息。

用户查询：%s

之前提到的实体：
%s

请判断代词（如"它"、"这个"、"那个"）指的是哪个实体，并返回完整的实体列表。

请返回以下JSON格式：
{
  "entities": [
    {
      "type": "实体类型",
      "value": "标准化值",
      "raw_text": "原始文本或引用的实体",
      "normalized": "标准化标识",
      "confidence": 0.0-1.0
    }
  ]
}

只返回JSON，不要其他内容。`, input, contextStr)
}

// parseEntityResponse parses the LLM response for entities.
func parseEntityResponse(response string) ([]agent.Entity, error) {
	response = cleanJSONResponse(response)

	var result struct {
		Entities []agent.Entity `json:"entities"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	return result.Entities, nil
}