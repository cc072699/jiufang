// Package service implements the visualization service for data display.
// Author: AI Agent
// Date: 2026-06-03
// Description: VisualizationService handles chart type recommendation, chart config generation, and data formatting.

package service

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"jiufang/internal/model/agent"
)

// VisualizationServiceInterface defines the interface for visualization operations.
type VisualizationServiceInterface interface {
	// DetermineVisualizationType determines the best visualization type based on data characteristics.
	DetermineVisualizationType(data []map[string]interface{}, queryType string) string

	// GenerateChartConfig generates chart configuration based on data and visualization type.
	GenerateChartConfig(data []map[string]interface{}, vizType string, columns []string, title string) (*agent.ChartConfig, error)

	// FormatData formats data for display (empty value handling, field formatting).
	FormatData(data []map[string]interface{}, columns []string) ([]map[string]interface{}, *agent.FormatMetadata)

	// HandleEmptyValues handles empty/null values based on field type.
	HandleEmptyValues(value interface{}, fieldType string, fieldName string, isKeyField bool) (interface{}, string)
}

// VisualizationService implements VisualizationServiceInterface.
type VisualizationService struct {
	logger *zap.Logger
}

// NewVisualizationService creates a new VisualizationService instance.
func NewVisualizationService(logger *zap.Logger) *VisualizationService {
	return &VisualizationService{
		logger: logger,
	}
}

// DetermineVisualizationType determines the best visualization type based on data characteristics.
// Based on PRD BR-024:
// - Trend queries -> line chart
// - Distribution/proportion queries -> pie chart
// - Comparison queries -> bar chart
// - Detail queries -> table
func (s *VisualizationService) DetermineVisualizationType(data []map[string]interface{}, queryType string) string {
	if len(data) == 0 {
		return agent.ChartTypeTable
	}

	// Analyze data characteristics
	metadata := s.analyzeDataCharacteristics(data)

	// If queryType is explicitly provided, use it as primary factor
	if queryType != "" {
		switch queryType {
		case "trend":
			if metadata.HasTimeDimension {
				return agent.ChartTypeLineChart
			}
		case "distribution", "proportion":
			if len(metadata.CategoryFields) > 0 && len(metadata.NumericFields) > 0 {
				return agent.ChartTypePieChart
			}
		case "comparison":
			if len(metadata.CategoryFields) > 0 && len(metadata.NumericFields) > 0 {
				return agent.ChartTypeBarChart
			}
		case "detail", "list":
			return agent.ChartTypeTable
		}
	}

	// Fallback to data-based heuristics
	// Time dimension + numeric values -> line chart (trend)
	if metadata.HasTimeDimension && len(metadata.NumericFields) > 0 {
		return agent.ChartTypeLineChart
	}

	// Category field + numeric values -> bar chart (comparison)
	if len(metadata.CategoryFields) > 0 && len(metadata.NumericFields) > 0 {
		// If only one category and few data points, use pie chart
		if len(data) <= 10 {
			return agent.ChartTypePieChart
		}
		return agent.ChartTypeBarChart
	}

	// Default to table for detail queries
	return agent.ChartTypeTable
}

// GenerateChartConfig generates chart configuration based on data and visualization type.
func (s *VisualizationService) GenerateChartConfig(
	data []map[string]interface{},
	vizType string,
	columns []string,
	title string,
) (*agent.ChartConfig, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("cannot generate chart config for empty data")
	}

	// Analyze data characteristics
	metadata := s.analyzeDataCharacteristics(data)

	// Create base chart config
	config := agent.NewChartConfig(vizType, title)

	// Generate axis and series configurations based on chart type
	switch vizType {
	case agent.ChartTypeLineChart:
		s.generateLineChartConfig(config, data, columns, metadata)
	case agent.ChartTypeBarChart:
		s.generateBarChartConfig(config, data, columns, metadata)
	case agent.ChartTypePieChart:
		s.generatePieChartConfig(config, data, columns, metadata)
	case agent.ChartTypeTable:
		// Table type doesn't need special config
		config.Series = []agent.SeriesConfig{}
	default:
		return nil, fmt.Errorf("unsupported visualization type: %s", vizType)
	}

	// Set data limit
	if len(data) > config.DataLimit {
		config.DataLimit = 500 // Keep default 500 per PRD
	}

	return config, nil
}

// FormatData formats data for display with empty value handling and field formatting.
// Based on PRD BR-026 and BR-027.
func (s *VisualizationService) FormatData(
	data []map[string]interface{},
	columns []string,
) ([]map[string]interface{}, *agent.FormatMetadata) {
	if len(data) == 0 {
		return data, &agent.FormatMetadata{
			TotalRows:   0,
			DisplayRows: 0,
		}
	}

	metadata := s.analyzeDataCharacteristics(data)
	metadata.TotalRows = len(data)
	metadata.DisplayRows = len(data)

	formattedData := make([]map[string]interface{}, len(data))
	emptyValueCount := 0
	emptyValueFieldsMap := make(map[string]bool)

	for i, row := range data {
		formattedRow := make(map[string]interface{})
		for _, col := range columns {
			value := row[col]
			fieldType := s.detectFieldType(col, value, metadata)

			// Check if this is a key field (amount, quantity, etc.)
			isKeyField := s.isKeyField(col)

			// Handle empty values
			formattedValue, hint := s.HandleEmptyValues(value, fieldType, col, isKeyField)
			if hint != "" {
				emptyValueCount++
				emptyValueFieldsMap[col] = true
			}

			// Format currency fields
			if s.isCurrencyField(col) && formattedValue != nil {
				// Add unit to column name in display (handled by frontend)
				formattedRow[col] = formattedValue
			} else {
				formattedRow[col] = formattedValue
			}
		}
		formattedData[i] = formattedRow
	}

	// Update metadata
	metadata.EmptyValueCount = emptyValueCount
	emptyValueFields := []string{}
	for field := range emptyValueFieldsMap {
		emptyValueFields = append(emptyValueFields, field)
	}
	metadata.EmptyValueFields = emptyValueFields

	return formattedData, metadata
}

// HandleEmptyValues handles empty/null values based on field type.
// Based on PRD BR-027:
// - Numeric fields -> display "0"
// - Other fields (text, date, enum) -> display empty string
// - Key fields (amount, quantity) -> add hint message
func (s *VisualizationService) HandleEmptyValues(
	value interface{},
	fieldType string,
	fieldName string,
	isKeyField bool,
) (interface{}, string) {
	// Check if value is empty
	if value == nil {
		return s.handleEmptyByType(fieldType, fieldName, isKeyField)
	}

	// Check for zero values in numeric types
	if fieldType == "numeric" {
		if num, ok := value.(float64); ok && num == 0 {
			if isKeyField {
				return 0, fmt.Sprintf("%s 为空值", fieldName)
			}
			return 0, ""
		}
		if num, ok := value.(int); ok && num == 0 {
			if isKeyField {
				return 0, fmt.Sprintf("%s 为空值", fieldName)
			}
			return 0, ""
		}
	}

	// Check for empty strings
	if str, ok := value.(string); ok && str == "" {
		return s.handleEmptyByType(fieldType, fieldName, isKeyField)
	}

	return value, ""
}

// handleEmptyByType handles empty values based on field type.
func (s *VisualizationService) handleEmptyByType(fieldType string, fieldName string, isKeyField bool) (interface{}, string) {
	switch fieldType {
	case "numeric":
		if isKeyField {
			return 0, fmt.Sprintf("%s 为空值", fieldName)
		}
		return 0, ""
	case "time", "date":
		return "", ""
	case "currency":
		return 0, fmt.Sprintf("%s 为空值", fieldName)
	default:
		return "", ""
	}
}

// analyzeDataCharacteristics analyzes data to determine field types and characteristics.
func (s *VisualizationService) analyzeDataCharacteristics(data []map[string]interface{}) *agent.FormatMetadata {
	metadata := &agent.FormatMetadata{}

	if len(data) == 0 {
		return metadata
	}

	// Analyze first row to determine field types
	firstRow := data[0]
	currencyFields := []string{}
	numericFields := []string{}
	timeFields := []string{}
	categoryFields := []string{}

	for field, value := range firstRow {
		fieldType := s.detectFieldTypeFromValue(value)

		// If value is nil, infer type from field name
		if fieldType == "unknown" {
			fieldType = s.inferFieldTypeFromName(field)
		}

		switch fieldType {
		case "currency":
			currencyFields = append(currencyFields, field)
			numericFields = append(numericFields, field)
		case "numeric":
			// Check if this is a currency field based on field name
			if s.isCurrencyField(field) {
				currencyFields = append(currencyFields, field)
			}
			numericFields = append(numericFields, field)
		case "time", "date":
			timeFields = append(timeFields, field)
			metadata.HasTimeDimension = true
		case "category", "string":
			categoryFields = append(categoryFields, field)
		}
	}

	metadata.CurrencyFields = currencyFields
	metadata.NumericFields = numericFields
	metadata.TimeFields = timeFields
	metadata.CategoryFields = categoryFields

	return metadata
}

// detectFieldType detects the type of a field based on name and value.
func (s *VisualizationService) detectFieldType(field string, value interface{}, metadata *agent.FormatMetadata) string {
	// Check if field is in metadata lists
	for _, f := range metadata.CurrencyFields {
		if f == field {
			return "currency"
		}
	}
	for _, f := range metadata.TimeFields {
		if f == field {
			return "time"
		}
	}
	for _, f := range metadata.NumericFields {
		if f == field {
			return "numeric"
		}
	}

	// Detect from value
	return s.detectFieldTypeFromValue(value)
}

// detectFieldTypeFromValue detects field type from the value.
func (s *VisualizationService) detectFieldTypeFromValue(value interface{}) string {
	if value == nil {
		return "unknown"
	}

	switch v := value.(type) {
	case int, int64, float64, float32:
		return "numeric"
	case string:
		// Check if it's a time/date string
		if s.isTimeString(v) {
			return "time"
		}
		return "string"
	case time.Time:
		return "time"
	default:
		return "unknown"
	}
}

// isTimeString checks if a string represents a time/date.
func (s *VisualizationService) isTimeString(str string) bool {
	timeFormats := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006/01/02",
		time.RFC3339,
	}

	for _, format := range timeFormats {
		if _, err := time.Parse(format, str); err == nil {
			return true
		}
	}
	return false
}

// inferFieldTypeFromName infers field type from field name when value is nil.
func (s *VisualizationService) inferFieldTypeFromName(field string) string {
	// Check if it's a currency field
	if s.isCurrencyField(field) {
		return "currency"
	}

	// Check if it's a time field
	timeKeywords := []string{
		"日期", "时间", "年", "月", "日",
		"date", "time", "year", "month", "day",
	}
	fieldLower := strings.ToLower(field)
	for _, keyword := range timeKeywords {
		if strings.Contains(fieldLower, keyword) {
			return "time"
		}
	}

	// Check if it's a numeric field
	numericKeywords := []string{
		"数量", "计数", "编号", "id",
		"quantity", "count", "number", "id",
	}
	for _, keyword := range numericKeywords {
		if strings.Contains(fieldLower, keyword) {
			return "numeric"
		}
	}

	// Default to string
	return "string"
}

// isCurrencyField checks if a field name indicates a currency/amount field.
func (s *VisualizationService) isCurrencyField(field string) bool {
	currencyKeywords := []string{
		"金额", "价格", "成本", "费用", "收入", "利润",
		"amount", "price", "cost", "fee", "revenue", "profit",
		"money", "total", "sum",
	}

	fieldLower := strings.ToLower(field)
	for _, keyword := range currencyKeywords {
		if strings.Contains(fieldLower, keyword) {
			return true
		}
	}
	return false
}

// isKeyField checks if a field is a key field (amount, quantity, etc.).
func (s *VisualizationService) isKeyField(field string) bool {
	keyKeywords := []string{
		"金额", "数量", "价格", "成本", "利润",
		"amount", "quantity", "price", "cost", "profit",
		"count", "total", "sum",
	}

	fieldLower := strings.ToLower(field)
	for _, keyword := range keyKeywords {
		if strings.Contains(fieldLower, keyword) {
			return true
		}
	}
	return false
}

// generateLineChartConfig generates configuration for line charts.
func (s *VisualizationService) generateLineChartConfig(
	config *agent.ChartConfig,
	data []map[string]interface{},
	columns []string,
	metadata *agent.FormatMetadata,
) {
	// X-axis: time dimension
	if len(metadata.TimeFields) > 0 {
		timeField := metadata.TimeFields[0]
		xData := []string{}
		for _, row := range data {
			if val, ok := row[timeField]; ok {
				xData = append(xData, fmt.Sprintf("%v", val))
			}
		}
		config.XAxis = &agent.AxisConfig{
			Type: "time",
			Name: timeField,
			Data: xData,
		}
	}

	// Y-axis: numeric values
	if len(metadata.NumericFields) > 0 {
		yField := metadata.NumericFields[0]
		config.YAxis = &agent.AxisConfig{
			Type: "value",
			Name: yField,
		}
		if s.isCurrencyField(yField) {
			config.YAxis.Unit = "元"
		}
	}

	// Series: numeric data
	for _, numField := range metadata.NumericFields {
		seriesData := []interface{}{}
		for _, row := range data {
			if val, ok := row[numField]; ok {
				seriesData = append(seriesData, val)
			}
		}
		config.Series = append(config.Series, agent.SeriesConfig{
			Name:   numField,
			Type:   "line",
			Data:   seriesData,
			Format: s.getFieldFormat(numField),
		})
	}
}

// generateBarChartConfig generates configuration for bar charts.
func (s *VisualizationService) generateBarChartConfig(
	config *agent.ChartConfig,
	data []map[string]interface{},
	columns []string,
	metadata *agent.FormatMetadata,
) {
	// X-axis: category dimension
	if len(metadata.CategoryFields) > 0 {
		catField := metadata.CategoryFields[0]
		xData := []string{}
		for _, row := range data {
			if val, ok := row[catField]; ok {
				xData = append(xData, fmt.Sprintf("%v", val))
			}
		}
		config.XAxis = &agent.AxisConfig{
			Type: "category",
			Name: catField,
			Data: xData,
		}
	}

	// Y-axis: numeric values
	if len(metadata.NumericFields) > 0 {
		yField := metadata.NumericFields[0]
		config.YAxis = &agent.AxisConfig{
			Type: "value",
			Name: yField,
		}
		if s.isCurrencyField(yField) {
			config.YAxis.Unit = "元"
		}
	}

	// Series: numeric data
	for _, numField := range metadata.NumericFields {
		seriesData := []interface{}{}
		for _, row := range data {
			if val, ok := row[numField]; ok {
				seriesData = append(seriesData, val)
			}
		}
		config.Series = append(config.Series, agent.SeriesConfig{
			Name:   numField,
			Type:   "bar",
			Data:   seriesData,
			Format: s.getFieldFormat(numField),
		})
	}
}

// generatePieChartConfig generates configuration for pie charts.
func (s *VisualizationService) generatePieChartConfig(
	config *agent.ChartConfig,
	data []map[string]interface{},
	columns []string,
	metadata *agent.FormatMetadata,
) {
	// Pie chart needs category and numeric field
	if len(metadata.CategoryFields) == 0 || len(metadata.NumericFields) == 0 {
		return
	}

	catField := metadata.CategoryFields[0]
	numField := metadata.NumericFields[0]

	seriesData := []interface{}{}
	for _, row := range data {
		name := ""
		if val, ok := row[catField]; ok {
			name = fmt.Sprintf("%v", val)
		}
		value := 0.0
		if val, ok := row[numField]; ok {
			if num, ok := val.(float64); ok {
				value = num
			}
		}
		seriesData = append(seriesData, map[string]interface{}{
			"name":  name,
			"value": value,
		})
	}

	config.Series = []agent.SeriesConfig{
		{
			Name:   numField,
			Type:   "pie",
			Data:   seriesData,
			Format: s.getFieldFormat(numField),
		},
	}

	// Set legend data
	legendData := []string{}
	for _, row := range data {
		if val, ok := row[catField]; ok {
			legendData = append(legendData, fmt.Sprintf("%v", val))
		}
	}
	config.Legend.Data = legendData
}

// getFieldFormat returns the format type for a field.
func (s *VisualizationService) getFieldFormat(field string) string {
	if s.isCurrencyField(field) {
		return "currency"
	}
	if strings.Contains(strings.ToLower(field), "率") ||
		strings.Contains(strings.ToLower(field), "percent") ||
		strings.Contains(strings.ToLower(field), "ratio") {
		return "percent"
	}
	return "number"
}
