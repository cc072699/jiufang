// Package service implements the application layer for data export.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

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
	// TODO: Implement Excel generation using excelize library
	// This is a placeholder implementation
	fileName := fmt.Sprintf("%s_%d_%s.xlsx", title, snowflakeID, exportTime.Format("20060102_150405"))

	// Placeholder: return dummy file info
	// In real implementation, use github.com/xuri/excelize/v2 to generate Excel
	return fileName, 0, fmt.Errorf("excel generation not implemented yet")
}

// generatePDF generates a PDF file from query data with watermark.
func (s *ExportAppService) generatePDF(data []map[string]interface{}, title string, snowflakeID int64, exportTime time.Time, userID int64) (string, int64, error) {
	// TODO: Implement PDF generation using gofpdf library with watermark
	// This is a placeholder implementation
	fileName := fmt.Sprintf("%s_%d_%s.pdf", title, snowflakeID, exportTime.Format("20060102_150405"))

	// Placeholder: return dummy file info
	// In real implementation, use github.com/jung-kurt/gofpdf to generate PDF with watermark
	return fileName, 0, fmt.Errorf("pdf generation not implemented yet")
}

// parseQueryResult parses the query result JSON string into data structure.
func parseQueryResult(resultJSON string, data *[]map[string]interface{}) error {
	if resultJSON == "" {
		*data = []map[string]interface{}{}
		return nil
	}
	return json.Unmarshal([]byte(resultJSON), data)
}
