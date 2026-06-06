// Package erp provides ERP database access for the AI Agent module.
package erp

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Reader implements ERPReaderInterface for reading from ERP database.
type Reader struct {
	db            *gorm.DB
	config        *ERPConfig
	dangerousSQL  *regexp.Regexp
	allowedTables map[string]bool
}

// NewReader creates a new ERP reader.
func NewReader(config *ERPConfig) (*Reader, error) {
	if config == nil {
		config = DefaultERPConfig()
	}

	// Open database connection
	var db *gorm.DB
	var err error

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	switch config.Driver {
	case "mysql":
		db, err = gorm.Open(mysql.Open(config.DSN()), gormConfig)
	case "postgres":
		db, err = gorm.Open(postgres.Open(config.DSN()), gormConfig)
	default:
		return nil, NewERPError(ErrCodeConnectionFailed, "unsupported database driver", "", fmt.Errorf("driver: %s", config.Driver))
	}

	if err != nil {
		return nil, NewERPError(ErrCodeConnectionFailed, "failed to connect to database", "", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, NewERPError(ErrCodeConnectionFailed, "failed to get database connection", "", err)
	}

	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Compile dangerous SQL patterns
	dangerousSQL := regexp.MustCompile(`(?i)(DELETE|UPDATE|INSERT|DROP|ALTER|CREATE|TRUNCATE|GRANT|REVOKE|EXEC|EXECUTE)`)

	return &Reader{
		db:           db,
		config:       config,
		dangerousSQL: dangerousSQL,
		allowedTables: make(map[string]bool),
	}, nil
}

// Query executes a read-only SQL query and returns results.
func (r *Reader) Query(ctx context.Context, sql string, args ...interface{}) ([]map[string]interface{}, error) {
	// Validate SQL
	if err := r.ValidateSQL(sql); err != nil {
		return nil, err
	}

	// Apply timeout
	if r.config.QueryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.config.QueryTimeout)
		defer cancel()
	}

	// Execute query
	var results []map[string]interface{}
	if err := r.db.WithContext(ctx).Raw(sql, args...).Scan(&results).Error; err != nil {
		return nil, NewERPError(ErrCodeQueryTimeout, "query execution failed", sql, err)
	}

	// Check result limit
	if r.config.MaxResultRows > 0 && len(results) > r.config.MaxResultRows {
		return nil, NewERPError(ErrCodeTooManyRows, "query returned too many rows", sql, nil)
	}

	return results, nil
}

// QueryWithLimit executes a read-only SQL query with a result limit.
func (r *Reader) QueryWithLimit(ctx context.Context, sql string, limit int) ([]map[string]interface{}, error) {
	// Validate SQL
	if err := r.ValidateSQL(sql); err != nil {
		return nil, err
	}

	// Add LIMIT clause if not present
	sql = strings.TrimSpace(sql)
	if !strings.Contains(strings.ToUpper(sql), "LIMIT") {
		sql = fmt.Sprintf("%s LIMIT %d", sql, limit)
	}

	return r.Query(ctx, sql)
}

// QueryWithTimeout executes a read-only SQL query with a timeout.
func (r *Reader) QueryWithTimeout(ctx context.Context, sql string, timeout time.Duration) ([]map[string]interface{}, error) {
	// Validate SQL
	if err := r.ValidateSQL(sql); err != nil {
		return nil, err
	}

	// Apply timeout
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	return r.Query(ctx, sql)
}

// GetTableSchema returns the schema information for a table.
func (r *Reader) GetTableSchema(ctx context.Context, tableName string) (*TableSchema, error) {
	// Check if table exists
	var count int64
	if err := r.db.WithContext(ctx).Raw(
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = ?",
		r.config.Database, tableName,
	).Scan(&count).Error; err != nil {
		return nil, NewERPError(ErrCodeTableNotFound, "failed to check table existence", "", err)
	}

	if count == 0 {
		return nil, NewERPError(ErrCodeTableNotFound, "table not found", "", nil)
	}

	schema := &TableSchema{
		Name:    tableName,
		Columns: []ColumnInfo{},
	}

	// Get column information
	var columns []struct {
		ColumnName    string `gorm:"column:COLUMN_NAME"`
		DataType      string `gorm:"column:DATA_TYPE"`
		IsNullable    string `gorm:"column:IS_NULLABLE"`
		ColumnDefault string `gorm:"column:COLUMN_DEFAULT"`
		ColumnComment string `gorm:"column:COLUMN_COMMENT"`
	}

	if err := r.db.WithContext(ctx).Raw(
		`SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT, COLUMN_COMMENT
		 FROM information_schema.columns
		 WHERE table_schema = ? AND table_name = ?
		 ORDER BY ordinal_position`,
		r.config.Database, tableName,
	).Scan(&columns).Error; err != nil {
		return nil, NewERPError(ErrCodeQueryTimeout, "failed to get column information", "", err)
	}

	for _, col := range columns {
		schema.Columns = append(schema.Columns, ColumnInfo{
			Name:         col.ColumnName,
			Type:         col.DataType,
			Nullable:     col.IsNullable == "YES",
			DefaultValue: col.ColumnDefault,
			Comment:      col.ColumnComment,
		})
	}

	return schema, nil
}

// GetTableList returns the list of available tables.
func (r *Reader) GetTableList(ctx context.Context) ([]string, error) {
	var tables []string

	if err := r.db.WithContext(ctx).Raw(
		"SELECT table_name FROM information_schema.tables WHERE table_schema = ?",
		r.config.Database,
	).Scan(&tables).Error; err != nil {
		return nil, NewERPError(ErrCodeQueryTimeout, "failed to get table list", "", err)
	}

	return tables, nil
}

// ValidateSQL validates if the SQL is safe to execute.
func (r *Reader) ValidateSQL(sql string) error {
	// Check for dangerous SQL keywords
	if r.dangerousSQL.MatchString(sql) {
		return NewERPError(ErrCodeInvalidSQL, "SQL contains dangerous keywords", sql, nil)
	}

	// Check if SQL starts with SELECT (case-insensitive)
	trimmedSQL := strings.TrimSpace(strings.ToUpper(sql))
	if !strings.HasPrefix(trimmedSQL, "SELECT") {
		return NewERPError(ErrCodeInvalidSQL, "only SELECT queries are allowed", sql, nil)
	}

	return nil
}

// IsReadOnly checks if the SQL is a read-only query.
func (r *Reader) IsReadOnly(sql string) bool {
	trimmedSQL := strings.TrimSpace(strings.ToUpper(sql))
	return strings.HasPrefix(trimmedSQL, "SELECT")
}

// GetDB returns the underlying database connection.
func (r *Reader) GetDB() *gorm.DB {
	return r.db
}

// SetAllowedTables sets the list of allowed tables for queries.
func (r *Reader) SetAllowedTables(tables []string) {
	r.allowedTables = make(map[string]bool)
	for _, table := range tables {
		r.allowedTables[strings.ToLower(table)] = true
	}
}

// IsTableAllowed checks if a table is allowed for queries.
func (r *Reader) IsTableAllowed(tableName string) bool {
	if len(r.allowedTables) == 0 {
		return true // All tables allowed if not restricted
	}
	return r.allowedTables[strings.ToLower(tableName)]
}

// Close closes the database connection.
func (r *Reader) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Ping checks if the database connection is alive.
func (r *Reader) Ping(ctx context.Context) error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}