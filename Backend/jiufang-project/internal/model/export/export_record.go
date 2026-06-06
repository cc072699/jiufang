// Package export implements the data export models.
// This file defines the ExportRecord entity for user's export history.
package export

import (
	"time"
)

// ExportFormat represents the format of exported file.
type ExportFormat string

const (
	ExportFormatExcel ExportFormat = "excel"
	ExportFormatPDF   ExportFormat = "pdf"
)

// ExportRecord represents a user's export history record.
// It stores metadata about each export operation.
type ExportRecord struct {
	ID             uint         `gorm:"primaryKey;autoIncrement" json:"-"`
	SnowflakeID    int64        `gorm:"uniqueIndex;not null" json:"id,string"`
	UserID         int64        `gorm:"not null;index" json:"user_id"`
	QueryRecordID  int64        `gorm:"not null" json:"query_record_id"`
	Format         ExportFormat `gorm:"type:varchar(20);not null" json:"format"`
	FileName       string       `gorm:"type:varchar(200);not null" json:"file_name"`
	FileSize       int64        `gorm:"not null" json:"file_size"`
	QuerySummary   string       `gorm:"type:text;default:null" json:"query_summary,omitempty"`
	CreatedAt      time.Time    `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

// TableName returns the table name for ExportRecord.
func (ExportRecord) TableName() string {
	return "export_records"
}

// ExportRequest represents the request to export query results.
type ExportRequest struct {
	QueryRecordID string      `json:"query_record_id" binding:"required"`
	Format        ExportFormat `json:"format" binding:"required,oneof=excel pdf"`
	Title         string      `json:"title"`
}

// ExportResult represents the result of an export operation.
type ExportResult struct {
	FileURL    string    `json:"file_url"`
	FileName   string    `json:"file_name"`
	FileSize   int64     `json:"file_size"`
	ExportTime time.Time `json:"export_time"`
}

// WatermarkConfig represents the watermark configuration for PDF export.
type WatermarkConfig struct {
	Text      string // Watermark text (e.g., "用户名 + 导出时间")
	FontSize  float64 // Font size for watermark
	Opacity   float64 // Opacity of watermark (0.0-1.0)
	Angle     float64 // Rotation angle in degrees
	PositionX float64 // X position offset
	PositionY float64 // Y position offset
}