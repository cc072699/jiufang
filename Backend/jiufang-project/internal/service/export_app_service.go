// Package service implements the application layer for data export.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"

	"jiufang/internal/model/export"
	querymodel "jiufang/internal/model/query"
	"jiufang/internal/pkg/id"
	"jiufang/internal/repository"
)

// ExportAppServiceInterface defines the interface for export application service.
type ExportAppServiceInterface interface {
	ExportQueryResult(ctx context.Context, userID int64, req *export.ExportRequest) (*export.ExportResult, error)
	GetExportRecordList(ctx context.Context, userID int64, page, pageSize int) ([]export.ExportRecord, int64, error)
}

// ExportAppService manages data export operations.
type ExportAppService struct {
	exportRecordRepo repository.ExportRecordRepositoryInterface
	queryRepo        repository.QueryRepositoryInterface
	idGenerator      id.SnowflakeGeneratorInterface
	logger           *zap.Logger
	exportDir        string // Directory to store exported files
	maxExportRows    int    // Maximum rows allowed for export
}

// NewExportAppService creates a new ExportAppService instance.
func NewExportAppService(
	exportRecordRepo repository.ExportRecordRepositoryInterface,
	queryRepo repository.QueryRepositoryInterface,
	idGenerator id.SnowflakeGeneratorInterface,
	logger *zap.Logger,
	exportDir string,
	maxExportRows int,
) *ExportAppService {
	return &ExportAppService{
		exportRecordRepo: exportRecordRepo,
		queryRepo:        queryRepo,
		idGenerator:      idGenerator,
		logger:           logger,
		exportDir:        exportDir,
		maxExportRows:    maxExportRows,
	}
}

// ExportQueryResult exports query results to Excel or PDF format.
func (s *ExportAppService) ExportQueryResult(ctx context.Context, userID int64, req *export.ExportRequest) (*export.ExportResult, error) {
	// Parse query record ID
	queryRecordID, err := strconv.ParseInt(req.QueryRecordID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid query record id format: %w", err)
	}

	// Get query record
	queryRecord, err := s.queryRepo.GetQueryRecordBySnowflakeID(ctx, queryRecordID)
	if err != nil {
		return nil, fmt.Errorf("failed to get query record: %w", err)
	}

	// Check if query record belongs to user
	if queryRecord.UserID != userID {
		return nil, fmt.Errorf("query record not owned by user")
	}

	// Check if query was successful
	if queryRecord.Status != querymodel.QueryStatusSuccess {
		return nil, fmt.Errorf("cannot export failed query result")
	}

	// Parse query result data
	var queryData []map[string]interface{}
	if err := parseQueryResult(queryRecord.ResultData, &queryData); err != nil {
		return nil, fmt.Errorf("failed to parse query result: %w", err)
	}

	// Check data size
	if len(queryData) > s.maxExportRows {
		return nil, fmt.Errorf("data exceeds maximum export rows (%d), please narrow down query scope", s.maxExportRows)
	}

	// Generate file based on format
	var fileName string
	var fileSize int64
	var fileURL string

	exportTime := time.Now()
	exportSnowflakeID := s.idGenerator.Generate()

	switch req.Format {
	case export.ExportFormatExcel:
		fileName, fileSize, err = s.generateExcel(queryData, req.Title, exportSnowflakeID, exportTime)
	case export.ExportFormatPDF:
		fileName, fileSize, err = s.generatePDF(queryData, req.Title, exportSnowflakeID, exportTime, userID)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", req.Format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate export file: %w", err)
	}

	// Build file URL
	fileURL = fmt.Sprintf("/downloads/%s/%s", exportTime.Format("20060102"), fileName)

	// Create export record
	exportRecord := &export.ExportRecord{
		SnowflakeID:   exportSnowflakeID,
		UserID:        userID,
		QueryRecordID: queryRecordID,
		Format:        req.Format,
		FileName:      fileName,
		FileSize:      fileSize,
		QuerySummary:  queryRecord.Input,
		CreatedAt:     exportTime,
	}

	if err := s.exportRecordRepo.CreateExportRecord(ctx, exportRecord); err != nil {
		s.logger.Error("Failed to save export record",
			zap.Error(err),
			zap.Int64("user_id", userID),
		)
		// Don't fail the export if record saving fails
	}

	// Log export operation
	s.logger.Info("Export completed",
		zap.Int64("user_id", userID),
		zap.String("format", string(req.Format)),
		zap.String("file_name", fileName),
		zap.Int64("file_size", fileSize),
	)

	return &export.ExportResult{
		FileURL:    fileURL,
		FileName:   fileName,
		FileSize:   fileSize,
		ExportTime: exportTime,
	}, nil
}

// GetExportRecordList retrieves export record list for a user with pagination.
func (s *ExportAppService) GetExportRecordList(ctx context.Context, userID int64, page, pageSize int) ([]export.ExportRecord, int64, error) {
	// Validate and set default pagination
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	records, total, err := s.exportRecordRepo.GetExportRecordsByUserID(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get export record list: %w", err)
	}

	return records, total, nil
}

// generateExcel generates an Excel file from query data.
func (s *ExportAppService) generateExcel(data []map[string]interface{}, title string, snowflakeID int64, exportTime time.Time) (string, int64, error) {
	if title == "" {
		title = "查询结果"
	}

	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	// Write title row
	f.SetCellValue(sheetName, "A1", title)
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	f.SetCellStyle(sheetName, "A1", "A1", titleStyle)
	f.MergeCell(sheetName, "A1", "A1")

	// Write export time
	f.SetCellValue(sheetName, "A2", fmt.Sprintf("导出时间: %s", exportTime.Format("2006-01-02 15:04:05")))

	if len(data) == 0 {
		f.SetCellValue(sheetName, "A3", "无数据")
	} else {
		// Extract column names from first row
		var columns []string
		for col := range data[0] {
			columns = append(columns, col)
		}

		// Write header row (row 3)
		headerStyle, _ := f.NewStyle(&excelize.Style{
			Font:      &excelize.Font{Bold: true, Size: 11},
			Fill:      excelize.Fill{Type: "pattern", Color: []string{"#D9E1F2"}, Pattern: 1},
			Alignment: &excelize.Alignment{Horizontal: "center"},
		})
		for ci, col := range columns {
			cell, _ := excelize.CoordinatesToCellName(ci+1, 3)
			f.SetCellValue(sheetName, cell, col)
		}
		startCol, _ := excelize.CoordinatesToCellName(1, 3)
		endCol, _ := excelize.CoordinatesToCellName(len(columns), 3)
		f.SetCellStyle(sheetName, startCol, endCol, headerStyle)

		// Write data rows (starting from row 4)
		for ri, row := range data {
			for ci, col := range columns {
				cell, _ := excelize.CoordinatesToCellName(ci+1, ri+4)
				f.SetCellValue(sheetName, cell, row[col])
			}
		}

		// Auto-fit column widths (approximate)
		for ci, col := range columns {
			width := float64(len(col))*1.5 + 4
			if width < 10 {
				width = 10
			}
			colName, _ := excelize.ColumnNumberToName(ci + 1)
			f.SetColWidth(sheetName, colName, colName, width)
		}
	}

	// Save file
	dir := filepath.Join(s.exportDir, exportTime.Format("20060102"))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", 0, fmt.Errorf("failed to create export directory: %w", err)
	}

	fileName := fmt.Sprintf("%s_%d_%s.xlsx", title, snowflakeID, exportTime.Format("20060102_150405"))
	filePath := filepath.Join(dir, fileName)

	if err := f.SaveAs(filePath); err != nil {
		return "", 0, fmt.Errorf("failed to save Excel file: %w", err)
	}

	// Get file size
	info, err := os.Stat(filePath)
	if err != nil {
		return fileName, 0, nil
	}

	return fileName, info.Size(), nil
}

// generatePDF generates a PDF file from query data with watermark.
func (s *ExportAppService) generatePDF(data []map[string]interface{}, title string, snowflakeID int64, exportTime time.Time, userID int64) (string, int64, error) {
	return "", 0, fmt.Errorf("PDF 导出暂不支持，请使用 Excel 格式")
}

// parseQueryResult parses the query result JSON string into data structure.
// The stored result may be either a flat array or a full result object with a "rows" field.
func parseQueryResult(resultJSON string, data *[]map[string]interface{}) error {
	if resultJSON == "" {
		*data = []map[string]interface{}{}
		return nil
	}

	// Try direct array unmarshal first
	if err := json.Unmarshal([]byte(resultJSON), data); err == nil {
		return nil
	}

	// Try as a full result object with "rows" field
	var resultObj map[string]interface{}
	if err := json.Unmarshal([]byte(resultJSON), &resultObj); err != nil {
		return err
	}

	rowsRaw, ok := resultObj["rows"]
	if !ok {
		*data = []map[string]interface{}{}
		return nil
	}

	rowsJSON, err := json.Marshal(rowsRaw)
	if err != nil {
		return err
	}

	return json.Unmarshal(rowsJSON, data)
}
