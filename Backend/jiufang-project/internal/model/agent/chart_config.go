// Package agent defines the chart configuration models for data visualization.
// Author: AI Agent
// Date: 2026-06-03
// Description: ChartConfig and related configuration structures for generating chart visualizations.

package agent

// ChartType constants for visualization types.
const (
	ChartTypeBarChart  = "bar_chart"
	ChartTypeLineChart = "line_chart"
	ChartTypePieChart  = "pie_chart"
	ChartTypeTable     = "table"
)

// ChartConfig represents the configuration for a chart visualization.
// It contains all necessary settings for rendering charts on the frontend.
type ChartConfig struct {
	// Type is the chart type (bar_chart, line_chart, pie_chart, table)
	Type string `json:"type"`

	// Title is the chart title
	Title string `json:"title"`

	// XAxis is the X-axis configuration
	XAxis *AxisConfig `json:"x_axis,omitempty"`

	// YAxis is the Y-axis configuration
	YAxis *AxisConfig `json:"y_axis,omitempty"`

	// Series is the list of data series configurations
	Series []SeriesConfig `json:"series"`

	// Legend is the legend configuration
	Legend *LegendConfig `json:"legend,omitempty"`

	// Tooltip is the tooltip configuration
	Tooltip *TooltipConfig `json:"tooltip,omitempty"`

	// Colors is the color palette for the chart
	Colors []string `json:"colors,omitempty"`

	// DataLimit is the maximum number of rows to display
	// Based on PRD BR-024, default is 500 rows
	DataLimit int `json:"data_limit"`

	// EmptyValueHint is the hint message for empty values
	EmptyValueHint string `json:"empty_value_hint,omitempty"`

	// CanSwitchType indicates if user can manually switch chart type
	CanSwitchType bool `json:"can_switch_type"`
}

// AxisConfig represents the configuration for a chart axis.
type AxisConfig struct {
	// Type is the axis type: category, value, time
	Type string `json:"type"`

	// Name is the axis name/label
	Name string `json:"name"`

	// Unit is the unit for the axis values (e.g., "元", "件", "%")
	Unit string `json:"unit,omitempty"`

	// Data is the axis data for category type axes
	Data []string `json:"data,omitempty"`

	// Min is the minimum value for value type axes
	Min *float64 `json:"min,omitempty"`

	// Max is the maximum value for value type axes
	Max *float64 `json:"max,omitempty"`
}

// SeriesConfig represents the configuration for a data series.
type SeriesConfig struct {
	// Name is the series name
	Name string `json:"name"`

	// Type is the series type (same as chart type for single series)
	Type string `json:"type"`

	// Data is the series data values
	Data []interface{} `json:"data"`

	// Format is the data format type: currency, number, percent, date
	Format string `json:"format,omitempty"`

	// Color is the series color (optional, overrides default palette)
	Color string `json:"color,omitempty"`

	// Label is the series label configuration
	Label *SeriesLabelConfig `json:"label,omitempty"`
}

// SeriesLabelConfig represents the label configuration for a series.
type SeriesLabelConfig struct {
	// Show indicates whether to show labels
	Show bool `json:"show"`

	// Position is the label position: inside, outside, top, bottom
	Position string `json:"position,omitempty"`

	// Formatter is the label format string
	Formatter string `json:"formatter,omitempty"`
}

// LegendConfig represents the configuration for chart legend.
type LegendConfig struct {
	// Show indicates whether to show legend
	Show bool `json:"show"`

	// Position is the legend position: top, bottom, left, right
	Position string `json:"position"`

	// Data is the legend item names
	Data []string `json:"data,omitempty"`
}

// TooltipConfig represents the configuration for chart tooltip.
type TooltipConfig struct {
	// Show indicates whether to show tooltip
	Show bool `json:"show"`

	// Trigger is the tooltip trigger type: item, axis
	Trigger string `json:"trigger"`

	// Formatter is the tooltip format string
	Formatter string `json:"formatter,omitempty"`
}

// FormatMetadata represents metadata about data formatting.
// It records information about how data was formatted for display.
type FormatMetadata struct {
	// TotalRows is the total number of rows before formatting
	TotalRows int `json:"total_rows"`

	// DisplayRows is the number of rows after formatting (may be limited)
	DisplayRows int `json:"display_rows"`

	// EmptyValueCount is the count of empty values found
	EmptyValueCount int `json:"empty_value_count"`

	// EmptyValueFields is the list of fields that contained empty values
	EmptyValueFields []string `json:"empty_value_fields,omitempty"`

	// CurrencyFields is the list of fields formatted as currency
	CurrencyFields []string `json:"currency_fields,omitempty"`

	// NumericFields is the list of numeric fields
	NumericFields []string `json:"numeric_fields,omitempty"`

	// TimeFields is the list of time/date fields
	TimeFields []string `json:"time_fields,omitempty"`

	// CategoryFields is the list of category/categorical fields
	CategoryFields []string `json:"category_fields,omitempty"`

	// HasTimeDimension indicates if data has a time dimension
	HasTimeDimension bool `json:"has_time_dimension"`

	// RecommendedChartType is the recommended chart type based on data analysis
	RecommendedChartType string `json:"recommended_chart_type"`
}

// NewChartConfig creates a new ChartConfig with default settings.
func NewChartConfig(chartType string, title string) *ChartConfig {
	return &ChartConfig{
		Type:          chartType,
		Title:         title,
		DataLimit:     500, // Default limit per PRD
		CanSwitchType: true,
		Colors:        getDefaultColors(),
		Legend:        &LegendConfig{Show: true, Position: "top"},
		Tooltip:       &TooltipConfig{Show: true, Trigger: "item"},
	}
}

// getDefaultColors returns the default color palette for charts.
func getDefaultColors() []string {
	return []string{
		"#5470c6", // Blue
		"#91cc75", // Green
		"#fac858", // Yellow
		"#ee6666", // Red
		"#73c0de", // Light Blue
		"#3ba272", // Dark Green
		"#fc8452", // Orange
		"#9a60b4", // Purple
		"#ea7ccc", // Pink
		"#c23531", // Brown
	}
}
