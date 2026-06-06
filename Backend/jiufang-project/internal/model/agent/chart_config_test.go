// Package agent_test tests the chart configuration models.
// Author: AI Agent
// Date: 2026-06-03
// Description: Unit tests for ChartConfig and related configuration structures.

package agent_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"jiufang/internal/model/agent"
)

// TestNewChartConfig tests the NewChartConfig function.
func TestNewChartConfig(t *testing.T) {
	t.Run("TC-CFG-01: Create chart config with normal parameters", func(t *testing.T) {
		// Arrange
		chartType := agent.ChartTypeBarChart
		title := "采购金额统计"

		// Act
		config := agent.NewChartConfig(chartType, title)

		// Assert
		assert.NotNil(t, config)
		assert.Equal(t, chartType, config.Type)
		assert.Equal(t, title, config.Title)
		assert.Equal(t, 500, config.DataLimit)
		assert.True(t, config.CanSwitchType)
		assert.NotNil(t, config.Legend)
		assert.True(t, config.Legend.Show)
		assert.Equal(t, "top", config.Legend.Position)
		assert.NotNil(t, config.Tooltip)
		assert.True(t, config.Tooltip.Show)
		assert.Equal(t, "item", config.Tooltip.Trigger)
		assert.Len(t, config.Colors, 10)
	})

	t.Run("TC-CFG-02: Create chart config for different chart types", func(t *testing.T) {
		// Test all chart types
		chartTypes := []string{
			agent.ChartTypeLineChart,
			agent.ChartTypeBarChart,
			agent.ChartTypePieChart,
			agent.ChartTypeTable,
		}

		for _, chartType := range chartTypes {
			// Act
			config := agent.NewChartConfig(chartType, "测试标题")

			// Assert
			assert.Equal(t, chartType, config.Type)
		}
	})

	t.Run("TC-CFG-03: Create chart config with empty title", func(t *testing.T) {
		// Arrange
		chartType := agent.ChartTypeBarChart
		title := ""

		// Act
		config := agent.NewChartConfig(chartType, title)

		// Assert
		assert.NotNil(t, config)
		assert.Equal(t, title, config.Title)
	})
}

// TestChartConfigDefaults tests default values in ChartConfig.
func TestChartConfigDefaults(t *testing.T) {
	t.Run("TC-CFG-04: Default colors are valid", func(t *testing.T) {
		// Arrange
		config := agent.NewChartConfig(agent.ChartTypeBarChart, "测试")

		// Assert
		assert.Len(t, config.Colors, 10)
		for _, color := range config.Colors {
			assert.NotEmpty(t, color)
			assert.Contains(t, color, "#") // Color should start with #
		}
	})

	t.Run("TC-CFG-05: Default legend configuration", func(t *testing.T) {
		// Arrange
		config := agent.NewChartConfig(agent.ChartTypeBarChart, "测试")

		// Assert
		assert.NotNil(t, config.Legend)
		assert.True(t, config.Legend.Show)
		assert.Equal(t, "top", config.Legend.Position)
	})

	t.Run("TC-CFG-06: Default tooltip configuration", func(t *testing.T) {
		// Arrange
		config := agent.NewChartConfig(agent.ChartTypeBarChart, "测试")

		// Assert
		assert.NotNil(t, config.Tooltip)
		assert.True(t, config.Tooltip.Show)
		assert.Equal(t, "item", config.Tooltip.Trigger)
	})

	t.Run("TC-CFG-07: Default data limit", func(t *testing.T) {
		// Arrange
		config := agent.NewChartConfig(agent.ChartTypeBarChart, "测试")

		// Assert
		assert.Equal(t, 500, config.DataLimit)
	})
}

// TestAxisConfig tests AxisConfig structure.
func TestAxisConfig(t *testing.T) {
	t.Run("TC-CFG-08: Create axis config with all fields", func(t *testing.T) {
		// Arrange
		axisConfig := &agent.AxisConfig{
			Type:  "category",
			Name:  "部门",
			Unit:  "",
			Data:  []string{"采购部", "销售部", "财务部"},
			Min:   nil,
			Max:   nil,
		}

		// Assert
		assert.Equal(t, "category", axisConfig.Type)
		assert.Equal(t, "部门", axisConfig.Name)
		assert.Len(t, axisConfig.Data, 3)
	})

	t.Run("TC-CFG-09: Create axis config with unit", func(t *testing.T) {
		// Arrange
		axisConfig := &agent.AxisConfig{
			Type:  "value",
			Name:  "金额",
			Unit:  "元",
		}

		// Assert
		assert.Equal(t, "value", axisConfig.Type)
		assert.Equal(t, "元", axisConfig.Unit)
	})
}

// TestSeriesConfig tests SeriesConfig structure.
func TestSeriesConfig(t *testing.T) {
	t.Run("TC-CFG-10: Create series config with all fields", func(t *testing.T) {
		// Arrange
		seriesConfig := agent.SeriesConfig{
			Name:   "采购金额",
			Type:   "bar",
			Data:   []interface{}{1000.0, 2000.0, 1500.0},
			Format: "currency",
			Color:  "#5470c6",
		}

		// Assert
		assert.Equal(t, "采购金额", seriesConfig.Name)
		assert.Equal(t, "bar", seriesConfig.Type)
		assert.Len(t, seriesConfig.Data, 3)
		assert.Equal(t, "currency", seriesConfig.Format)
	})

	t.Run("TC-CFG-11: Create series config with label", func(t *testing.T) {
		// Arrange
		seriesConfig := agent.SeriesConfig{
			Name: "测试",
			Type: "bar",
			Data: []interface{}{100.0},
			Label: &agent.SeriesLabelConfig{
				Show:      true,
				Position:  "top",
				Formatter: "{c}",
			},
		}

		// Assert
		assert.NotNil(t, seriesConfig.Label)
		assert.True(t, seriesConfig.Label.Show)
		assert.Equal(t, "top", seriesConfig.Label.Position)
	})
}

// TestFormatMetadata tests FormatMetadata structure.
func TestFormatMetadata(t *testing.T) {
	t.Run("TC-CFG-12: Create format metadata with all fields", func(t *testing.T) {
		// Arrange
		metadata := &agent.FormatMetadata{
			TotalRows:            100,
			DisplayRows:          50,
			EmptyValueCount:      10,
			EmptyValueFields:     []string{"amount", "quantity"},
			CurrencyFields:       []string{"amount", "price"},
			NumericFields:        []string{"amount", "quantity", "price"},
			TimeFields:           []string{"date"},
			CategoryFields:       []string{"department"},
			HasTimeDimension:     true,
			RecommendedChartType: agent.ChartTypeLineChart,
		}

		// Assert
		assert.Equal(t, 100, metadata.TotalRows)
		assert.Equal(t, 50, metadata.DisplayRows)
		assert.Equal(t, 10, metadata.EmptyValueCount)
		assert.Len(t, metadata.EmptyValueFields, 2)
		assert.Len(t, metadata.CurrencyFields, 2)
		assert.True(t, metadata.HasTimeDimension)
		assert.Equal(t, agent.ChartTypeLineChart, metadata.RecommendedChartType)
	})

	t.Run("TC-CFG-13: Create empty format metadata", func(t *testing.T) {
		// Arrange
		metadata := &agent.FormatMetadata{}

		// Assert
		assert.Equal(t, 0, metadata.TotalRows)
		assert.Equal(t, 0, metadata.DisplayRows)
		assert.False(t, metadata.HasTimeDimension)
	})
}

// TestChartTypeConstants tests chart type constants.
func TestChartTypeConstants(t *testing.T) {
	t.Run("TC-CFG-14: Chart type constants are valid", func(t *testing.T) {
		// Assert
		assert.Equal(t, "bar_chart", agent.ChartTypeBarChart)
		assert.Equal(t, "line_chart", agent.ChartTypeLineChart)
		assert.Equal(t, "pie_chart", agent.ChartTypePieChart)
		assert.Equal(t, "table", agent.ChartTypeTable)
	})
}