// Package erp provides ERP database access for the AI Agent module.
// This package handles read-only queries to the ERP database.
package erp

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ERPReaderInterface defines the interface for ERP database read operations.
// All operations are read-only to ensure data safety.
type ERPReaderInterface interface {
	// Query executes a read-only SQL query and returns results.
	Query(ctx context.Context, sql string, args ...interface{}) ([]map[string]interface{}, error)

	// QueryWithLimit executes a read-only SQL query with a result limit.
	QueryWithLimit(ctx context.Context, sql string, limit int) ([]map[string]interface{}, error)

	// QueryWithTimeout executes a read-only SQL query with a timeout.
	QueryWithTimeout(ctx context.Context, sql string, timeout time.Duration) ([]map[string]interface{}, error)

	// GetTableSchema returns the schema information for a table.
	GetTableSchema(ctx context.Context, tableName string) (*TableSchema, error)

	// GetTableList returns the list of available tables.
	GetTableList(ctx context.Context) ([]string, error)

	// ValidateSQL validates if the SQL is safe to execute.
	ValidateSQL(sql string) error

	// IsReadOnly checks if the SQL is a read-only query.
	IsReadOnly(sql string) bool

	// GetDB returns the underlying database connection for advanced operations.
	// Use with caution - this bypasses the read-only safety checks.
	GetDB() *gorm.DB
}

// TableSchema represents the schema of a database table.
type TableSchema struct {
	Name        string       `json:"name"`
	Columns     []ColumnInfo `json:"columns"`
	PrimaryKeys []string     `json:"primary_keys"`
	ForeignKeys []ForeignKey `json:"foreign_keys"`
	Indexes     []IndexInfo  `json:"indexes"`
}

// ColumnInfo represents information about a table column.
type ColumnInfo struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Nullable     bool   `json:"nullable"`
	DefaultValue string `json:"default_value"`
	Comment      string `json:"comment"`
}

// ForeignKey represents a foreign key constraint.
type ForeignKey struct {
	Name              string `json:"name"`
	ColumnName        string `json:"column_name"`
	ReferencedTable   string `json:"referenced_table"`
	ReferencedColumn  string `json:"referenced_column"`
}

// IndexInfo represents information about a table index.
type IndexInfo struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
}

// ERPConfig contains the configuration for ERP database connection.
type ERPConfig struct {
	// Driver is the database driver (mysql, postgres, sqlserver)
	Driver string `yaml:"driver"`

	// Host is the database host
	Host string `yaml:"host"`

	// Port is the database port
	Port int `yaml:"port"`

	// Database is the database name
	Database string `yaml:"database"`

	// Username is the database username
	Username string `yaml:"username"`

	// Password is the database password
	Password string `yaml:"password"`

	// MaxOpenConns is the maximum number of open connections
	MaxOpenConns int `yaml:"max_open_conns"`

	// MaxIdleConns is the maximum number of idle connections
	MaxIdleConns int `yaml:"max_idle_conns"`

	// ConnMaxLifetime is the maximum lifetime of a connection
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`

	// QueryTimeout is the default query timeout
	QueryTimeout time.Duration `yaml:"query_timeout"`

	// MaxResultRows is the maximum number of rows returned by a query
	MaxResultRows int `yaml:"max_result_rows"`
}

// DefaultERPConfig returns the default ERP configuration.
func DefaultERPConfig() *ERPConfig {
	return &ERPConfig{
		Driver:          "mysql",
		Host:            "localhost",
		Port:            3306,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
		QueryTimeout:    30 * time.Second,
		MaxResultRows:   10000,
	}
}

// DSN returns the data source name for the database connection.
func (c *ERPConfig) DSN() string {
	switch c.Driver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			c.Username, c.Password, c.Host, c.Port, c.Database)
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			c.Host, c.Port, c.Username, c.Password, c.Database)
	case "sqlserver":
		return fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
			c.Username, c.Password, c.Host, c.Port, c.Database)
	default:
		return ""
	}
}

// ERPError represents an error from ERP operations.
type ERPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Query   string `json:"query,omitempty"`
	Err     error  `json:"-"`
}

// Error implements the error interface.
func (e *ERPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("ERP error [%d]: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("ERP error [%d]: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *ERPError) Unwrap() error {
	return e.Err
}

// Error codes
const (
	ErrCodeInvalidSQL       = 1001
	ErrCodeUnauthorized      = 1002
	ErrCodeQueryTimeout     = 1003
	ErrCodeTooManyRows      = 1004
	ErrCodeConnectionFailed = 1005
	ErrCodeTableNotFound    = 1006
	ErrCodePermissionDenied = 1007
)

// NewERPError creates a new ERP error.
func NewERPError(code int, message string, query string, err error) *ERPError {
	return &ERPError{
		Code:    code,
		Message: message,
		Query:   query,
		Err:     err,
	}
}

// NullString represents a nullable string.
type NullString struct {
	sql.NullString
}

// NullInt64 represents a nullable int64.
type NullInt64 struct {
	sql.NullInt64
}

// NullFloat64 represents a nullable float64.
type NullFloat64 struct {
	sql.NullFloat64
}

// NullTime represents a nullable time.
type NullTime struct {
	sql.NullTime
}