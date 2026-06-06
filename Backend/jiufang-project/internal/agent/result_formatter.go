// Package agent implements the AI Agent for semantic understanding and SQL generation.
package agent

import (
	"fmt"
	"strings"

	"jiufang/internal/model/agent"
)

// ResultFormatter formats query results for display.
type ResultFormatter struct {
}

// NewResultFormatter creates a new result formatter.
func NewResultFormatter() *ResultFormatter {
	return &ResultFormatter{}
}

// Format formats the query result for display.
func (f *ResultFormatter) Format(result *agent.QueryResult) string {
	if result.IsEmpty {
		return "查询结果为空，没有找到符合条件的数据。"
	}

	// Build understanding summary
	understanding := result.Understanding
	if understanding == "" {
		understanding = f.generateUnderstanding(result)
	}

	// Build result summary
	summary := fmt.Sprintf("共查询到 %d 条数据，耗时 %.2f 秒。\n\n", result.TotalRows, float64(result.ExecutionTime)/1000)

	// Build data preview
	preview := f.buildDataPreview(result)

	return understanding + "\n\n" + summary + preview
}

// FormatWithUnderstanding formats the result with custom understanding.
func (f *ResultFormatter) FormatWithUnderstanding(result *agent.QueryResult, understanding string) string {
	if result.IsEmpty {
		return understanding + "\n\n查询结果为空，没有找到符合条件的数据。"
	}

	summary := fmt.Sprintf("共查询到 %d 条数据，耗时 %.2f 秒。", result.TotalRows, float64(result.ExecutionTime)/1000)

	return understanding + "\n\n" + summary
}

// FormatForJSON formats the result as JSON string.
func (f *ResultFormatter) FormatForJSON(result *agent.QueryResult) string {
	if result.IsEmpty {
		return "{\"data\": [], \"message\": \"查询结果为空\"}"
	}

	// Build JSON structure
	jsonStr := "{\"data\": ["
	for i, row := range result.Data {
		if i > 0 {
			jsonStr += ", "
		}
		jsonStr += f.rowToJSON(row)
	}
	jsonStr += fmt.Sprintf("], \"total_rows\": %d, \"execution_time\": %d}", result.TotalRows, result.ExecutionTime)

	return jsonStr
}

// FormatForTable formats the result as a table.
func (f *ResultFormatter) FormatForTable(result *agent.QueryResult) string {
	if result.IsEmpty {
		return "| 结果 |\n|------|\n| 无数据 |"
	}

	// Build table header
	header := "| "
	for _, col := range result.Columns {
		header += col + " | "
	}
	header += "\n"

	// Build separator
	separator := "|"
	for _, col := range result.Columns {
		separator += strings.Repeat("-", len(col)+2) + "|"
	}
	separator += "\n"

	// Build rows
	rows := ""
	for _, row := range result.Data {
		rows += "| "
		for _, col := range result.Columns {
			value := f.formatValue(row[col])
			rows += value + " | "
		}
		rows += "\n"
	}

	return header + separator + rows
}

// FormatForChart formats the result for chart visualization.
func (f *ResultFormatter) FormatForChart(result *agent.QueryResult, chartType string) string {
	if result.IsEmpty {
		return "{\"type\": \"" + chartType + "\", \"data\": [], \"message\": \"无数据\"}"
	}

	// Build chart data structure
	chartData := "{\"type\": \"" + chartType + "\", \"data\": {"

	switch chartType {
	case "bar_chart":
		chartData += f.buildBarChartData(result)
	case "line_chart":
		chartData += f.buildLineChartData(result)
	case "pie_chart":
		chartData += f.buildPieChartData(result)
	default:
		chartData += "\"values\": []"
	}

	chartData += "}}"

	return chartData
}

// generateUnderstanding generates understanding summary from result.
func (f *ResultFormatter) generateUnderstanding(result *agent.QueryResult) string {
	if result.IsEmpty {
		return "查询结果为空"
	}

	// Analyze result structure
	if len(result.Columns) == 0 {
		return "查询完成"
	}

	// Check if it's a statistics result
	for _, col := range result.Columns {
		if strings.Contains(strings.ToLower(col), "count") ||
			strings.Contains(strings.ToLower(col), "sum") ||
			strings.Contains(strings.ToLower(col), "avg") {
			if len(result.Data) > 0 {
				value := result.Data[0][col]
				return fmt.Sprintf("统计结果：%s = %v", col, value)
			}
		}
	}

	// Default understanding
	return fmt.Sprintf("查询到 %d 条记录", result.TotalRows)
}

// buildDataPreview builds a preview of the data.
func (f *ResultFormatter) buildDataPreview(result *agent.QueryResult) string {
	if len(result.Data) == 0 {
		return ""
	}

	// Show first 3 rows
	preview := "数据预览（前3条）：\n"
	for i := 0; i < 3 && i < len(result.Data); i++ {
		preview += fmt.Sprintf("%d. ", i+1)
		for _, col := range result.Columns {
			value := f.formatValue(result.Data[i][col])
			preview += fmt.Sprintf("%s: %s, ", col, value)
		}
		preview = strings.TrimSuffix(preview, ", ") + "\n"
	}

	if result.TotalRows > 3 {
		preview += fmt.Sprintf("... 还有 %d 条数据\n", result.TotalRows-3)
	}

	return preview
}

// formatValue formats a value for display.
func (f *ResultFormatter) formatValue(value interface{}) string {
	if value == nil {
		return "null"
	}

	switch v := value.(type) {
	case string:
		return v
	case int, int64, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// rowToJSON converts a row to JSON string.
func (f *ResultFormatter) rowToJSON(row map[string]interface{}) string {
	jsonStr := "{"
	for key, value := range row {
		jsonStr += fmt.Sprintf("\"%s\": %s, ", key, f.valueToJSON(value))
	}
	jsonStr = strings.TrimSuffix(jsonStr, ", ")
	jsonStr += "}"
	return jsonStr
}

// valueToJSON converts a value to JSON format.
func (f *ResultFormatter) valueToJSON(value interface{}) string {
	if value == nil {
		return "null"
	}

	switch v := value.(type) {
	case string:
		return fmt.Sprintf("\"%s\"", v)
	case int, int64, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("\"%v\"", v)
	}
}

// buildBarChartData builds data for bar chart.
func (f *ResultFormatter) buildBarChartData(result *agent.QueryResult) string {
	if len(result.Columns) < 2 {
		return "\"labels\": [], \"values\": []"
	}

	labels := "\"labels\": ["
	values := "\"values\": ["

	for i, row := range result.Data {
		if i > 0 {
			labels += ", "
			values += ", "
		}
		labels += fmt.Sprintf("\"%v\"", row[result.Columns[0]])
		values += fmt.Sprintf("%v", row[result.Columns[1]])
	}

	labels += "]"
	values += "]"

	return labels + ", " + values
}

// buildLineChartData builds data for line chart.
func (f *ResultFormatter) buildLineChartData(result *agent.QueryResult) string {
	if len(result.Columns) < 2 {
		return "\"x\": [], \"y\": []"
	}

	x := "\"x\": ["
	y := "\"y\": ["

	for i, row := range result.Data {
		if i > 0 {
			x += ", "
			y += ", "
		}
		x += fmt.Sprintf("\"%v\"", row[result.Columns[0]])
		y += fmt.Sprintf("%v", row[result.Columns[1]])
	}

	x += "]"
	y += "]"

	return x + ", " + y
}

// buildPieChartData builds data for pie chart.
func (f *ResultFormatter) buildPieChartData(result *agent.QueryResult) string {
	if len(result.Columns) < 2 {
		return "\"labels\": [], \"values\": []"
	}

	labels := "\"labels\": ["
	values := "\"values\": ["

	for i, row := range result.Data {
		if i > 0 {
			labels += ", "
			values += ", "
		}
		labels += fmt.Sprintf("\"%v\"", row[result.Columns[0]])
		values += fmt.Sprintf("%v", row[result.Columns[1]])
	}

	labels += "]"
	values += "]"

	return labels + ", " + values
}

// DetermineVisualizationType determines the best visualization type.
func (f *ResultFormatter) DetermineVisualizationType(result *agent.QueryResult, intent *agent.Intent) string {
	if result.IsEmpty {
		return agent.VisualizationNone
	}

	// Based on intent type
	switch intent.Type {
	case agent.IntentTypeStatistics:
		// Single value statistics - use table or pie chart
		if len(result.Data) == 1 && len(result.Columns) <= 3 {
			return agent.VisualizationTable
		}
		// Multi-category statistics - use bar chart
		if len(result.Data) > 1 && len(result.Data) <= 10 {
			return agent.VisualizationBarChart
		}
		return agent.VisualizationTable

	case agent.IntentTypeTrend:
		// Trend data - use line chart
		if len(result.Columns) >= 2 {
			return agent.VisualizationLineChart
		}
		return agent.VisualizationTable

	case agent.IntentTypeComparison:
		// Comparison data - use bar chart
		if len(result.Data) <= 10 {
			return agent.VisualizationBarChart
		}
		return agent.VisualizationTable

	case agent.IntentTypeDetail:
		// Detail data - use table
		return agent.VisualizationTable

	default:
		return agent.VisualizationTable
	}
}

// GenerateSuggestedQuestions generates suggested follow-up questions.
func (f *ResultFormatter) GenerateSuggestedQuestions(result *agent.QueryResult, intent *agent.Intent) []string {
	questions := []string{}

	if result.IsEmpty {
		questions = append(questions, "尝试调整查询条件")
		questions = append(questions, "查看所有数据")
		return questions
	}

	// Based on intent type
	switch intent.Type {
	case agent.IntentTypeStatistics:
		questions = append(questions, "查看详细明细")
		questions = append(questions, "按供应商分组统计")
		questions = append(questions, "查看趋势变化")

	case agent.IntentTypeDetail:
		questions = append(questions, "统计总金额")
		questions = append(questions, "按状态筛选")
		questions = append(questions, "导出数据")

	case agent.IntentTypeTrend:
		questions = append(questions, "查看具体明细")
		questions = append(questions, "对比不同时间段")
		questions = append(questions, "分析增长原因")

	case agent.IntentTypeComparison:
		questions = append(questions, "查看详细对比")
		questions = append(questions, "分析差异原因")
	}

	return questions
}