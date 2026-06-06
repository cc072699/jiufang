// Package v1_test tests the export HTTP handlers.
// Author: AI Agent
// Date: 2026-06-03
// Description: Unit tests for ExportHandler covering HTTP request handling and response validation.

package v1_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	v1 "jiufang/internal/api/v1"
	"jiufang/internal/mocks"
	"jiufang/internal/model/export"
)

// setupTestRouter creates a test Gin router with mock services.
func setupTestRouter(ctrl *gomock.Controller) (*gin.Engine, *mocks.MockExportAppService) {
	gin.SetMode(gin.TestMode)

	mockExportSvc := mocks.NewMockExportAppService(ctrl)
	logger := zap.NewNop()

	handler := v1.NewExportHandler(mockExportSvc, logger)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Mock auth middleware - set user_id in context
		c.Set("user_id", int64(100))
		c.Next()
	})

	exportGroup := router.Group("/api/v1/export")
	{
		exportGroup.POST("", handler.ExportQueryResult)
		exportGroup.GET("/records", handler.GetExportRecords)
	}

	return router, mockExportSvc
}

// setupTestRouterWithoutAuth creates a test router without authentication.
func setupTestRouterWithoutAuth(ctrl *gomock.Controller) (*gin.Engine, *mocks.MockExportAppService) {
	gin.SetMode(gin.TestMode)

	mockExportSvc := mocks.NewMockExportAppService(ctrl)
	logger := zap.NewNop()

	handler := v1.NewExportHandler(mockExportSvc, logger)

	router := gin.New()
	exportGroup := router.Group("/api/v1/export")
	{
		exportGroup.POST("", handler.ExportQueryResult)
		exportGroup.GET("/records", handler.GetExportRecords)
	}

	return router, mockExportSvc
}

// TestExportHandler_ExportQueryResult tests the ExportQueryResult HTTP handler.
func TestExportHandler_ExportQueryResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("TC-HDL-01: Normal export request", func(t *testing.T) {
		// Arrange
		router, mockSvc := setupTestRouter(ctrl)

		reqBody := export.ExportRequest{
			QueryRecordID: "200",
			Format:        export.ExportFormatExcel,
			Title:         "采购数据报表",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		expectedResult := &export.ExportResult{
			FileURL:    "/downloads/20260603/export_300.xlsx",
			FileName:   "采购数据报表_300.xlsx",
			FileSize:   1024,
			ExportTime: time.Now(),
		}

		mockSvc.EXPECT().
			ExportQueryResult(gomock.Any(), int64(100), gomock.Any()).
			Return(expectedResult, nil)

		// Act
		req := httptest.NewRequest(http.MethodPost, "/api/v1/export", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])
		assert.Equal(t, "success", response["message"])
	})

	t.Run("TC-HDL-02: Unauthorized request", func(t *testing.T) {
		// Arrange
		router, _ := setupTestRouterWithoutAuth(ctrl)

		reqBody := export.ExportRequest{
			QueryRecordID: "200",
			Format:        export.ExportFormatExcel,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		// Act
		req := httptest.NewRequest(http.MethodPost, "/api/v1/export", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(40101), response["code"])
		assert.Equal(t, "JWT token invalid", response["message"])
	})

	t.Run("TC-HDL-03: Invalid request body", func(t *testing.T) {
		// Arrange
		router, _ := setupTestRouter(ctrl)

		// Invalid JSON body (missing required fields)
		bodyBytes := []byte(`{"format": "excel"}`)

		// Act
		req := httptest.NewRequest(http.MethodPost, "/api/v1/export", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(40001), response["code"])
		assert.Equal(t, "Parameter validation failed", response["message"])
	})

	t.Run("TC-HDL-04: Permission denied - not owned by user", func(t *testing.T) {
		// Arrange
		router, mockSvc := setupTestRouter(ctrl)

		reqBody := export.ExportRequest{
			QueryRecordID: "200",
			Format:        export.ExportFormatExcel,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockSvc.EXPECT().
			ExportQueryResult(gomock.Any(), int64(100), gomock.Any()).
			Return(nil, errors.New("query record not owned by user"))

		// Act
		req := httptest.NewRequest(http.MethodPost, "/api/v1/export", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusForbidden, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(40301), response["code"])
		assert.Equal(t, "Permission denied", response["message"])
	})

	t.Run("TC-HDL-05: Service error", func(t *testing.T) {
		// Arrange
		router, mockSvc := setupTestRouter(ctrl)

		reqBody := export.ExportRequest{
			QueryRecordID: "200",
			Format:        export.ExportFormatExcel,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockSvc.EXPECT().
			ExportQueryResult(gomock.Any(), int64(100), gomock.Any()).
			Return(nil, errors.New("internal service error"))

		// Act
		req := httptest.NewRequest(http.MethodPost, "/api/v1/export", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(50000), response["code"])
		assert.Equal(t, "Export failed", response["message"])
	})
}

// TestExportHandler_GetExportRecords tests the GetExportRecords HTTP handler.
func TestExportHandler_GetExportRecords(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("TC-HDL-06: Normal get export records", func(t *testing.T) {
		// Arrange
		router, mockSvc := setupTestRouter(ctrl)

		testRecords := []export.ExportRecord{
			{
				SnowflakeID:   123456789,
				UserID:        100,
				QueryRecordID: 200,
				Format:        export.ExportFormatExcel,
				FileName:      "export1.xlsx",
				FileSize:      1024,
				CreatedAt:     time.Now(),
			},
		}

		mockSvc.EXPECT().
			GetExportRecordList(gomock.Any(), int64(100), 1, 20).
			Return(testRecords, int64(1), nil)

		// Act
		req := httptest.NewRequest(http.MethodGet, "/api/v1/export/records?page=1&page_size=20", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])
		assert.Equal(t, "success", response["message"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, float64(1), data["total"])
		assert.Equal(t, float64(1), data["page"])
		assert.Equal(t, float64(20), data["page_size"])
	})

	t.Run("TC-HDL-07: Unauthorized get export records", func(t *testing.T) {
		// Arrange
		router, _ := setupTestRouterWithoutAuth(ctrl)

		// Act
		req := httptest.NewRequest(http.MethodGet, "/api/v1/export/records", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(40101), response["code"])
		assert.Equal(t, "JWT token invalid", response["message"])
	})

	t.Run("TC-HDL-08: Service error", func(t *testing.T) {
		// Arrange
		router, mockSvc := setupTestRouter(ctrl)

		mockSvc.EXPECT().
			GetExportRecordList(gomock.Any(), int64(100), 1, 20).
			Return(nil, int64(0), errors.New("database error"))

		// Act
		req := httptest.NewRequest(http.MethodGet, "/api/v1/export/records", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(50000), response["code"])
		assert.Equal(t, "Failed to get export record list", response["message"])
	})
}
