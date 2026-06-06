// Package repository implements the operation log repository unit tests.
package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"jiufang/internal/model/audit"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestOperationLogRepository_Create_Success tests successful creation of an operation log.
func TestOperationLogRepository_Create_Success(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	logger := zap.NewNop()
	repo := NewOperationLogRepository(gormDB, logger)
	ctx := context.Background()

	userID := int64(1001)
	log := &audit.OperationLog{
		SnowflakeID:     123456789,
		UserID:          &userID,
		OperationType:   audit.OperationTypeLogin,
		OperationObject: "user_login",
		OperationDetail: "用户登录成功",
		OperationResult: audit.OperationResultSuccess,
		IPAddress:       "192.168.1.100",
		CreatedAt:       time.Now(),
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "operation_logs"`)).
		WithArgs(log.SnowflakeID, log.UserID, log.OperationType, log.OperationObject, log.OperationDetail, log.OperationResult, log.IPAddress, log.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Act
	err = repo.Create(ctx, log)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestOperationLogRepository_Create_DatabaseError tests creation with database error.
func TestOperationLogRepository_Create_DatabaseError(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	logger := zap.NewNop()
	repo := NewOperationLogRepository(gormDB, logger)
	ctx := context.Background()

	userID := int64(1001)
	log := &audit.OperationLog{
		SnowflakeID:     123456789,
		UserID:          &userID,
		OperationType:   audit.OperationTypeLogin,
		OperationResult: audit.OperationResultSuccess,
		CreatedAt:       time.Now(),
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "operation_logs"`)).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	// Act
	err = repo.Create(ctx, log)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrInvalidData, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestOperationLogRepository_GetByID_Success tests successful retrieval by ID.
func TestOperationLogRepository_GetByID_Success(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	logger := zap.NewNop()
	repo := NewOperationLogRepository(gormDB, logger)
	ctx := context.Background()

	now := time.Now()
	userID := int64(1001)
	expectedLog := &audit.OperationLog{
		ID:              1,
		SnowflakeID:     123456789,
		UserID:          &userID,
		OperationType:   audit.OperationTypeLogin,
		OperationResult: audit.OperationResultSuccess,
		CreatedAt:       now,
	}

	rows := sqlmock.NewRows([]string{"id", "snowflake_id", "user_id", "operation_type", "operation_object", "operation_detail", "operation_result", "ip_address", "created_at"}).
		AddRow(1, 123456789, 1001, "login", "", "", "success", "", now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "operation_logs" WHERE "operation_logs"."id" = $1 ORDER BY "operation_logs"."id" LIMIT 1`)).
		WithArgs(1).
		WillReturnRows(rows)

	// Act
	result, err := repo.GetByID(ctx, 1)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedLog.ID, result.ID)
	assert.Equal(t, expectedLog.SnowflakeID, result.SnowflakeID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestOperationLogRepository_GetByID_NotFound tests retrieval when record not found.
func TestOperationLogRepository_GetByID_NotFound(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	logger := zap.NewNop()
	repo := NewOperationLogRepository(gormDB, logger)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "operation_logs" WHERE "operation_logs"."id" = $1 ORDER BY "operation_logs"."id" LIMIT 1`)).
		WithArgs(999).
		WillReturnRows(sqlmock.NewRows([]string{"id", "snowflake_id", "user_id", "operation_type", "operation_object", "operation_detail", "operation_result", "ip_address", "created_at"}))

	// Act
	result, err := repo.GetByID(ctx, 999)

	// Assert
	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestOperationLogRepository_GetBySnowflakeID_Success tests successful retrieval by snowflake ID.
func TestOperationLogRepository_GetBySnowflakeID_Success(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	logger := zap.NewNop()
	repo := NewOperationLogRepository(gormDB, logger)
	ctx := context.Background()

	now := time.Now()
	userID := int64(1001)
	expectedLog := &audit.OperationLog{
		ID:              1,
		SnowflakeID:     123456789,
		UserID:          &userID,
		OperationType:   audit.OperationTypeLogin,
		OperationResult: audit.OperationResultSuccess,
		CreatedAt:       now,
	}

	rows := sqlmock.NewRows([]string{"id", "snowflake_id", "user_id", "operation_type", "operation_object", "operation_detail", "operation_result", "ip_address", "created_at"}).
		AddRow(1, 123456789, 1001, "login", "", "", "success", "", now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "operation_logs" WHERE snowflake_id = $1 ORDER BY "operation_logs"."id" LIMIT 1`)).
		WithArgs(123456789).
		WillReturnRows(rows)

	// Act
	result, err := repo.GetBySnowflakeID(ctx, 123456789)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedLog.SnowflakeID, result.SnowflakeID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestOperationLogRepository_GetBySnowflakeID_NotFound tests retrieval when record not found.
func TestOperationLogRepository_GetBySnowflakeID_NotFound(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	logger := zap.NewNop()
	repo := NewOperationLogRepository(gormDB, logger)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "operation_logs" WHERE snowflake_id = $1 ORDER BY "operation_logs"."id" LIMIT 1`)).
		WithArgs(999999).
		WillReturnRows(sqlmock.NewRows([]string{"id", "snowflake_id", "user_id", "operation_type", "operation_object", "operation_detail", "operation_result", "ip_address", "created_at"}))

	// Act
	result, err := repo.GetBySnowflakeID(ctx, 999999)

	// Assert
	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestOperationLogRepository_List_Success tests successful list retrieval without filters.
func TestOperationLogRepository_List_Success(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	logger := zap.NewNop()
	repo := NewOperationLogRepository(gormDB, logger)
	ctx := context.Background()

	now := time.Now()

	// Mock count query
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "operation_logs"`)).
		WillReturnRows(countRows)

	// Mock list query
	listRows := sqlmock.NewRows([]string{"id", "snowflake_id", "user_id", "operation_type", "operation_object", "operation_detail", "operation_result", "ip_address", "created_at"}).
		AddRow(2, 123456790, 1002, "logout", "", "", "success", "", now).
		AddRow(1, 123456789, 1001, "login", "", "", "success", "", now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "operation_logs" ORDER BY created_at DESC LIMIT 10`)).
		WillReturnRows(listRows)

	// Act
	logs, total, err := repo.List(ctx, 0, 10, 0, "", "", "")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, logs, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestOperationLogRepository_List_WithUserIDFilter tests list retrieval with user ID filter.
func TestOperationLogRepository_List_WithUserIDFilter(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	logger := zap.NewNop()
	repo := NewOperationLogRepository(gormDB, logger)
	ctx := context.Background()

	now := time.Now()
	userID := int64(1001)

	// Mock count query with user_id filter
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "operation_logs" WHERE user_id = $1`)).
		WithArgs(1001).
		WillReturnRows(countRows)

	// Mock list query with user_id filter
	listRows := sqlmock.NewRows([]string{"id", "snowflake_id", "user_id", "operation_type", "operation_object", "operation_detail", "operation_result", "ip_address", "created_at"}).
		AddRow(1, 123456789, 1001, "login", "", "", "success", "", now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "operation_logs" WHERE user_id = $1 ORDER BY created_at DESC LIMIT 10`)).
		WithArgs(1001).
		WillReturnRows(listRows)

	// Act
	logs, total, err := repo.List(ctx, 0, 10, userID, "", "", "")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, logs, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestOperationLogRepository_List_WithTimeFilter tests list retrieval with time filters.
func TestOperationLogRepository_List_WithTimeFilter(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	logger := zap.NewNop()
	repo := NewOperationLogRepository(gormDB, logger)
	ctx := context.Background()

	now := time.Now()
	startTime := now.Add(-24 * time.Hour).Format(time.RFC3339)
	endTime := now.Format(time.RFC3339)

	// Mock count query with time filters
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "operation_logs" WHERE created_at >= $1 AND created_at <= $2`)).
		WillReturnRows(countRows)

	// Mock list query with time filters
	listRows := sqlmock.NewRows([]string{"id", "snowflake_id", "user_id", "operation_type", "operation_object", "operation_detail", "operation_result", "ip_address", "created_at"}).
		AddRow(1, 123456789, 1001, "login", "", "", "success", "", now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "operation_logs" WHERE created_at >= $1 AND created_at <= $2 ORDER BY created_at DESC LIMIT 10`)).
		WillReturnRows(listRows)

	// Act
	logs, total, err := repo.List(ctx, 0, 10, 0, "", startTime, endTime)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, logs, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestOperationLogRepository_List_DatabaseError tests list retrieval with database error.
func TestOperationLogRepository_List_DatabaseError(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	logger := zap.NewNop()
	repo := NewOperationLogRepository(gormDB, logger)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "operation_logs"`)).
		WillReturnError(gorm.ErrInvalidDB)

	// Act
	logs, total, err := repo.List(ctx, 0, 10, 0, "", "", "")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, int64(0), total)
	assert.Nil(t, logs)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestOperationLogRepository_GetByUsername_Success tests successful retrieval of user ID by username.
func TestOperationLogRepository_GetByUsername_Success(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	logger := zap.NewNop()
	repo := NewOperationLogRepository(gormDB, logger)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"snowflake_id"}).AddRow(1001)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT snowflake_id FROM "users" WHERE username = $1`)).
		WithArgs("admin").
		WillReturnRows(rows)

	// Act
	userID, err := repo.GetByUsername(ctx, "admin")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(1001), userID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestOperationLogRepository_GetByUsername_NotFound tests retrieval when user not found.
func TestOperationLogRepository_GetByUsername_NotFound(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	logger := zap.NewNop()
	repo := NewOperationLogRepository(gormDB, logger)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT snowflake_id FROM "users" WHERE username = $1`)).
		WithArgs("nonexistent").
		WillReturnRows(sqlmock.NewRows([]string{"snowflake_id"}))

	// Act
	userID, err := repo.GetByUsername(ctx, "nonexistent")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(0), userID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestOperationLogRepository_GetByUsername_DatabaseError tests retrieval with database error.
func TestOperationLogRepository_GetByUsername_DatabaseError(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	logger := zap.NewNop()
	repo := NewOperationLogRepository(gormDB, logger)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT snowflake_id FROM "users" WHERE username = $1`)).
		WithArgs("admin").
		WillReturnError(gorm.ErrInvalidDB)

	// Act
	userID, err := repo.GetByUsername(ctx, "admin")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, int64(0), userID)
	assert.NoError(t, mock.ExpectationsWereMet())
}
