// Package service_test tests the visualization service implementation.
// Author: AI Agent
// Date: 2026-06-03
// Description: Unit tests for VisualizationService covering chart type recommendation, chart config generation, and data formatting.

package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"jiufang/internal/model/agent"
	"jiufang/internal/service"
)

// TestVisualizationService_DetermineVisualizationType tests the DetermineVisualizationType method.
func TestVisualizationService_DetermineVisualizationType(t *testing.T) {
	// Arrange
	logger := zap.NewNop()
	svc := service.NewVisualizationService(logger)

	t.Run("TC-VIZ-01: Empty data returns table", func(t *testing.T) {
		// Arrange
		data := []map[string]interface{}{}

		// Act
		vizType := svc.DetermineVisualizationType(data, "")

		// Assert
		assert.Equal(t, agent.ChartTypeTable, vizType)
	})

	t.Run("TC-VIZ-02: Time dimension data returns line chart", func(t *testing.T) {
		// Arrange
		data := []map[string]interface{}{
			{"date": "2026-01-01", "amount": 1000.0},
			{"date": "2026-01-02", "amount": 1500.0},
			{"date": "2026-01-03", "amount": 2000.0},
		}

		// Act
		vizType := svc.DetermineVisualizationType(data, "")

		// Assert
		assert.Equal(t, agent.ChartTypeLineChart, vizType)
	})

	t.Run("TC-VIZ-03: Category + numeric data (small) returns pie chart", func(t *testing.T) {
		// Arrange
		data := []map[string]interface{}{
			{"category": "A", "value": 100.0},
			{"category": "B", "value": 200.0},
			{"category": "C", "value": 150.0},
		}

		// Act
		vizType := svc.DetermineVisualizationType(data, "")

		// Assert
		assert.Equal(t, agent.ChartTypePieChart, vizType)
	})

	t.Run("TC-VIZ-04: Category + numeric data (large) returns bar chart", func(t *testing.T) {
		// Arrange
		data := []map[string]interface{}{}
		for i := 0; i < 15; i++ {
			data = append(data, map[string]interface{}{
				"category": "Cat" + string(rune(i)),
				"value":    float64(i * 100),
			})
		}

		// Act
		vizType := svc.DetermineVisualizationType(data, "")

		// Assert
		assert.Equal(t, agent.ChartTypeBarChart, vizType)
	})

	t.Run("TC-VIZ-05: Detail data returns table", func(t *testing.T) {
		// Arrange
		data := []map[string]interface{}{
			{"name": "Item1", "description": "Desc1"},
			{"name": "Item2", "description": "Desc2"},
		}

		// Act
		vizType := svc.DetermineVisualizationType(data, "")

		// Assert
		assert.Equal(t, agent.ChartTypeTable, vizType)
	})

	t.Run("TC-VIZ-06: Explicit queryType trend returns line chart", func(t *testing.T) {
		// Arrange
		data := []map[string]interface{}{
			{"date": "2026-01-01", "amount": 1000.0},
		}

		// Act
		vizType := svc.DetermineVisualizationType(data, "trend")

		// Assert
		assert.Equal(t, agent.ChartTypeLineChart, vizType)
	})

	t.Run("TC-VIZ-07: Explicit queryType distribution returns pie chart", func(t *testing.T) {
		// Arrange
		data := []map[string]interface{}{
			{"category": "A", "value": 100.0},
			{"category": "B", "value": 200.0},
		}

		// Act
		vizType := svc.DetermineVisualizationType(data, "distribution")

		// Assert
		assert.Equal(t, agent.ChartTypePieChart, vizType)
	})

	t.Run("TC-VIZ-08: Explicit queryType comparison returns bar chart", func(t *testing.T) {
		// Arrange
		data := []map[string]interface{}{
			{"category": "A", "value": 100.0},
			{"category": "B", "value": 200.0},
		}

		// Act
		vizType := svc.DetermineVisualizationType(data, "comparison")

		// Assert
		assert.Equal(t, agent.ChartTypeBarChart, vizType)
	})

	t.Run("TC-VIZ-09: Explicit queryType detail returns table", func(t *testing.T) {
		// Arrange
		data := []map[string]interface{}{
			{"name": "Item1", "value": 100.0},
		}

		// Act
		vizType := svc.DetermineVisualizationType(data, "detail")

		// Assert
		assert.Equal(t, agent.ChartTypeTable, vizType)
	})
}

// TestVisualizationService_GenerateChartConfig tests the GenerateChartConfig method.
func TestVisualizationService_GenerateChartConfig(t *testing.T) {
	// Arrange
	logger := zap.NewNop()
	svc := service.NewVisualizationService(logger)

	t.Run("TC-VIZ-07: Generate line chart config", func(t *testing.T) {
		// Arrange
		data := []map[string]interface{}{
			{"date": "2026-01-01", "amount": 1000.0},
			{"date": "2026-01-02", "amount": 1500.0},
		}
		columns := []string{"date", "amount"}

		// Act
		config, err := svc.GenerateChartConfig(data, agent.ChartTypeLineChart, columns, "采购趋势")

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, agent.ChartTypeLineChart, config.Type)
		assert.Equal(t, "采购趋势", config.Title)
		assert.NotNil(t, config.XAxis)
		assert.NotNil(t, config.YAxis)
		assert.Len(t, config.Series, 1)
		assert.True(t, config.CanSwitchType)
	})

	t.Run("TC-VIZ-08: Generate bar chart config", func(t *testing.T) {
		// Arrange
		data := []map[string]interface{}{
			{"category": "采购部", "amount": 1000.0},
			{"category": "销售部", "amount": 2000.0},
		}
		columns := []string{"category", "amount"}

		// Act
		config, err := svc.GenerateChartConfig(data, agent.ChartTypeBarChart, columns, "部门采购金额")

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, agent.ChartTypeBarChart, config.Type)
		assert.Equal(t, "部门采购金额", config.Title)
		assert.NotNil(t, config.XAxis)
		assert.NotNil(t, config.YAxis)
		assert.Len(t, config.Series, 1)
	})

	t.Run("TC-VIZ-09: Generate pie chart config", func(t *testing.T) {
		// Arrange
		data := []map[string]interface{}{
			{"category": "采购部", "amount": 1000.0},
			{"category": "销售部", "amount": 2000.0},
		}
		columns := []string{"category", "amount"}

		// Act
		config, err := svc.GenerateChartConfig(data, agent.ChartTypePieChart, columns, "部门占比")

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, agent.ChartTypePieChart, config.Type)
		assert.Equal(t, "部门占比", config.Title)
		assert.Len(t, config.Series, 1)
		assert.NotNil(t, config.Legend)
	})

	t.Run("TC-VIZ-10: Empty data returns error", func(t *testing.T) {
		// Arrange
		data := []map[string]interface{}{}
		columns := []string{}

		// Act
		config, err := svc.GenerateChartConfig(data, agent.ChartTypeBarChart, columns, "测试")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "empty data")
	})

	t.Run("TC-VIZ-11: Unsupported chart type returns error", func(t *testing.T) {
		// Arrange
		data := []map[string]interface{}{
			{"value": 100.0},
		}
		columns := []string{"value"}

		// Act
		config, err := svc.GenerateChartConfig(data, "invalid_type", columns, "测试")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "unsupported visualization type")
	})
}

// TestVisualizationService_FormatData tests the FormatData method.
func TestVisualizationService_FormatData(t *testing.T) {
	// Arrange
	logger := zap.NewNop()
	svc := service.NewVisualizationService(logger)

	t.Run("TC-VIZ-11: Format normal data", func(t *testing.T) {
		// Arrange
		data := []map[string]interface{}{
			{"name": "Item1", "amount": 1000.0, "count": 10},
			{"name": "Item2", "amount": 2000.0, "count": 20},
		}
		columns := []string{"name", "amount", "count"}

		// Act
		formattedData, metadata := svc.FormatData(data, columns)

		// Assert
		assert.Len(t, formattedData, 2)
		assert.NotNil(t, metadata)
		assert.Equal(t, 2, metadata.TotalRows)
		assert.Equal(t, 2, metadata.DisplayRows)
	})

	t.Run("TC-VIZ-12: Format data with empty values", func(t *testing.T) {
		// Arrange
		data := []map[string]interface{}{
			{"name": "Item1", "amount": nil, "count": 10},
			{"name": nil, "amount": 2000.0, "count": nil},
		}
		columns := []string{"name", "amount", "count"}

		// Act
		formattedData, metadata := svc.FormatData(data, columns)

		// Assert
		assert.Len(t, formattedData, 2)
		assert.NotNil(t, metadata)
		assert.True(t, metadata.EmptyValueCount > 0)
	})

	t.Run("TC-VIZ-13: Format empty data", func(t *testing.T) {
		// Arrange
		data := []map[string]interface{}{}
		columns := []string{}

		// Act
		formattedData, metadata := svc.FormatData(data, columns)

		// Assert
		assert.Len(t, formattedData, 0)
		assert.NotNil(t, metadata)
		assert.Equal(t, 0, metadata.TotalRows)
	})
}

// TestVisualizationService_HandleEmptyValues tests the HandleEmptyValues method.
func TestVisualizationService_HandleEmptyValues(t *testing.T) {
	// Arrange
	logger := zap.NewNop()
	svc := service.NewVisualizationService(logger)

	t.Run("TC-VIZ-13: Handle nil value for numeric field", func(t *testing.T) {
		// Arrange
		var value interface{} = nil
		fieldType := "numeric"
		fieldName := "amount"
		isKeyField := false

		// Act
		result, hint := svc.HandleEmptyValues(value, fieldType, fieldName, isKeyField)

		// Assert
		assert.Equal(t, 0, result)
		assert.Empty(t, hint)
	})

	t.Run("TC-VIZ-14: Handle nil value for string field", func(t *testing.T) {
		// Arrange
		var value interface{} = nil
		fieldType := "string"
		fieldName := "name"
		isKeyField := false

		// Act
		result, hint := svc.HandleEmptyValues(value, fieldType, fieldName, isKeyField)

		// Assert
		assert.Equal(t, "", result)
		assert.Empty(t, hint)
	})

	t.Run("TC-VIZ-15: Handle nil value for key field", func(t *testing.T) {
		// Arrange
		var value interface{} = nil
		fieldType := "currency"
		fieldName := "采购金额"
		isKeyField := true

		// Act
		result, hint := svc.HandleEmptyValues(value, fieldType, fieldName, isKeyField)

		// Assert
		assert.Equal(t, 0, result)
		assert.NotEmpty(t, hint)
		assert.Contains(t, hint, "空值")
	})

	t.Run("TC-VIZ-16: Handle non-empty value", func(t *testing.T) {
		// Arrange
		value := 100.0
		fieldType := "numeric"
		fieldName := "amount"
		isKeyField := false

		// Act
		result, hint := svc.HandleEmptyValues(value, fieldType, fieldName, isKeyField)

		// Assert
		assert.Equal(t, 100.0, result)
		assert.Empty(t, hint)
	})

	t.Run("TC-VIZ-17: Handle zero value for key field", func(t *testing.T) {
		// Arrange
		value := 0.0
		fieldType := "numeric"
		fieldName := "数量"
		isKeyField := true

		// Act
		result, hint := svc.HandleEmptyValues(value, fieldType, fieldName, isKeyField)

		// Assert
		assert.Equal(t, 0, result)
		assert.NotEmpty(t, hint)
		assert.Contains(t, hint, "空值")
	})

	t.Run("TC-VIZ-18: Handle empty string", func(t *testing.T) {
		// Arrange
		value := ""
		fieldType := "string"
		fieldName := "name"
		isKeyField := false

		// Act
		result, hint := svc.HandleEmptyValues(value, fieldType, fieldName, isKeyField)

		// Assert
		assert.Equal(t, "", result)
		assert.Empty(t, hint)
	})
}

// TestVisualizationService_FieldDetection tests field type detection through public methods.
func TestVisualizationService_FieldDetection(t *testing.T) {
	// Arrange
	logger := zap.NewNop()
	svc := service.NewVisualizationService(logger)

	t.Run("TC-VIZ-19: Currency field formatting", func(t *testing.T) {
		// Arrange
		data := []map[string]interface{}{
			{"采购金额": 1000.0, "name": "Item1"},
		}
		columns := []string{"采购金额", "name"}

		// Act
		_, metadata := svc.FormatData(data, columns)

		// Assert
		assert.NotNil(t, metadata)
		assert.Contains(t, metadata.CurrencyFields, "采购金额")
	})

	t.Run("TC-VIZ-20: Key field empty value handling", func(t *testing.T) {
		// Arrange
		data := []map[string]interface{}{
			{"amount": nil, "name": "Item1"},
		}
		columns := []string{"amount", "name"}

		// Act
		_, metadata := svc.FormatData(data, columns)

		// Assert
		assert.NotNil(t, metadata)
		assert.True(t, metadata.EmptyValueCount > 0)
	})

	t.Run("TC-VIZ-21: Time field detection", func(t *testing.T) {
		// Arrange
		data := []map[string]interface{}{
			{"date": "2026-01-01", "amount": 1000.0},
		}
		columns := []string{"date", "amount"}

		// Act
		_, metadata := svc.FormatData(data, columns)

		// Assert
		assert.NotNil(t, metadata)
		assert.Contains(t, metadata.TimeFields, "date")
		assert.True(t, metadata.HasTimeDimension)
	})
}
