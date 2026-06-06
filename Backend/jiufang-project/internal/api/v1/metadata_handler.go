package v1

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"jiufang/internal/infrastructure/erp"
	"jiufang/internal/pkg/response"
)

// MetadataHandler handles ERP metadata requests (table list, schema).
type MetadataHandler struct {
	erpReader erp.ERPReaderInterface
	logger    *zap.Logger
}

// NewMetadataHandler creates a new MetadataHandler.
func NewMetadataHandler(erpReader erp.ERPReaderInterface, logger *zap.Logger) *MetadataHandler {
	return &MetadataHandler{erpReader: erpReader, logger: logger}
}

// TableInfo represents a table with its columns for the permission UI.
type TableInfo struct {
	Name    string         `json:"name"`
	Label   string         `json:"label"`
	Columns []ColumnDetail `json:"columns"`
}

// ColumnDetail represents a column in a table.
type ColumnDetail struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Comment string `json:"comment"`
}

// GetTables handles GET /api/v1/metadata/tables - returns all ERP tables with their columns.
func (h *MetadataHandler) GetTables(c *gin.Context) {
	ctx := context.Background()

	tables, err := h.erpReader.GetTableList(ctx)
	if err != nil {
		h.logger.Error("failed to get table list", zap.Error(err))
		response.InternalError(c, "failed to get table list")
		return
	}

	result := make([]TableInfo, 0, len(tables))
	for _, tableName := range tables {
		schema, err := h.erpReader.GetTableSchema(ctx, tableName)
		if err != nil {
			h.logger.Warn("failed to get table schema", zap.String("table", tableName), zap.Error(err))
			continue
		}

		columns := make([]ColumnDetail, 0, len(schema.Columns))
		for _, col := range schema.Columns {
			columns = append(columns, ColumnDetail{
				Name:    col.Name,
				Type:    col.Type,
				Comment: col.Comment,
			})
		}

		result = append(result, TableInfo{
			Name:    tableName,
			Label:   tableName,
			Columns: columns,
		})
	}

	response.Success(c, result)
}
