// Package repository implements the alert repository unit tests.
package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"jiufang/internal/model/report"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestAlertRepository_Create_Success tests successful creation of an alert.
func TestAlertRepository_Create_Success(t *testing.T) {
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
	repo := NewAlertRepository(gormDB, logger)
	ctx := context.Background()

	alert := &report.Alert{
		SnowflakeID:      123456789,
		Name:             "库存低于安全库存预警",
		Description:      "当库存低于100时触发预警",
		SQL:              "SELECT SUM(quantity) as inventory FROM products WHERE status='active'",
		Condition:        "< 100",
		Recipients:       "[1001, 1002]",
		PushChannel:      report.PushChannelWeChat,
		TriggerFrequency: report.TriggerFrequencyEveryTime,
		Status:           report.AlertStatusActive,
		CreatedBy:        1001,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Mock expectations
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO \"alerts\"")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "last_triggered_at", "created_at", "updated_at"}).AddRow(1, nil, alert.CreatedAt, alert.UpdatedAt))
	mock.ExpectCommit()

	// Act
	err = repo.Create(ctx, alert)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestAlertRepository_GetByID_Success tests successful retrieval by ID.
func TestAlertRepository_GetByID_Success(t *testing.T) {
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
	repo := NewAlertRepository(gormDB, logger)
	ctx := context.Background()

	now := time.Now()
	expectedAlert := &report.Alert{
		ID:               1,
		SnowflakeID:      123456789,
		Name:             "库存低于安全库存预警",
		Description:      "当库存低于100时触发预警",
		SQL:              "SELECT SUM(quantity) as inventory FROM products WHERE status='active'",
		Condition:        "< 100",
		Recipients:       "[1001, 1002]",
		PushChannel:      report.PushChannelWeChat,
		TriggerFrequency: report.TriggerFrequencyEveryTime,
		Status:           report.AlertStatusActive,
		CreatedBy:        1001,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// Mock expectations
	rows := sqlmock.NewRows([]string{
		"id", "snowflake_id", "name", "description", "sql", "condition", "recipients",
		"push_channel", "trigger_frequency", "silence_start", "silence_end", "status",
		"last_triggered_at", "created_by", "created_at", "updated_at",
	}).AddRow(
		expectedAlert.ID, expectedAlert.SnowflakeID, expectedAlert.Name, expectedAlert.Description,
		expectedAlert.SQL, expectedAlert.Condition, expectedAlert.Recipients,
		expectedAlert.PushChannel, expectedAlert.TriggerFrequency, expectedAlert.SilenceStart,
		expectedAlert.SilenceEnd, expectedAlert.Status, expectedAlert.LastTriggeredAt,
		expectedAlert.CreatedBy, expectedAlert.CreatedAt, expectedAlert.UpdatedAt,
	)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM \"alerts\" WHERE \"alerts\".\"id\" = $1 ORDER BY \"alerts\".\"id\" LIMIT $2")).
		WithArgs(1, 1).
		WillReturnRows(rows)

	// Act
	alert, err := repo.GetByID(ctx, 1)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, alert)
	assert.Equal(t, expectedAlert.SnowflakeID, alert.SnowflakeID)
	assert.Equal(t, expectedAlert.Name, alert.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestAlertRepository_GetByID_NotFound tests retrieval when alert not found.
func TestAlertRepository_GetByID_NotFound(t *testing.T) {
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
	repo := NewAlertRepository(gormDB, logger)
	ctx := context.Background()

	// Mock expectations
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM \"alerts\" WHERE \"alerts\".\"id\" = $1 ORDER BY \"alerts\".\"id\" LIMIT $2")).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	// Act
	alert, err := repo.GetByID(ctx, 1)

	// Assert
	assert.NoError(t, err)
	assert.Nil(t, alert)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestAlertRepository_GetBySnowflakeID_Success tests successful retrieval by snowflake ID.
func TestAlertRepository_GetBySnowflakeID_Success(t *testing.T) {
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
	repo := NewAlertRepository(gormDB, logger)
	ctx := context.Background()

	now := time.Now()
	expectedAlert := &report.Alert{
		ID:          1,
		SnowflakeID: 123456789,
		Name:        "库存低于安全库存预警",
		Status:      report.AlertStatusActive,
		CreatedBy:   1001,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Mock expectations
	rows := sqlmock.NewRows([]string{
		"id", "snowflake_id", "name", "description", "sql", "condition", "recipients",
		"push_channel", "trigger_frequency", "silence_start", "silence_end", "status",
		"last_triggered_at", "created_by", "created_at", "updated_at",
	}).AddRow(
		expectedAlert.ID, expectedAlert.SnowflakeID, expectedAlert.Name, expectedAlert.Description,
		expectedAlert.SQL, expectedAlert.Condition, expectedAlert.Recipients,
		expectedAlert.PushChannel, expectedAlert.TriggerFrequency, expectedAlert.SilenceStart,
		expectedAlert.SilenceEnd, expectedAlert.Status, expectedAlert.LastTriggeredAt,
		expectedAlert.CreatedBy, expectedAlert.CreatedAt, expectedAlert.UpdatedAt,
	)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM \"alerts\" WHERE snowflake_id = $1 ORDER BY \"alerts\".\"id\" LIMIT $2")).
		WithArgs(123456789, 1).
		WillReturnRows(rows)

	// Act
	alert, err := repo.GetBySnowflakeID(ctx, 123456789)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, alert)
	assert.Equal(t, expectedAlert.SnowflakeID, alert.SnowflakeID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestAlertRepository_List_Success tests successful list retrieval.
func TestAlertRepository_List_Success(t *testing.T) {
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
	repo := NewAlertRepository(gormDB, logger)
	ctx := context.Background()

	now := time.Now()
	alerts := []report.Alert{
		{
			ID:          1,
			SnowflakeID: 123456789,
			Name:        "库存预警",
			Status:      report.AlertStatusActive,
			CreatedBy:   1001,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          2,
			SnowflakeID: 987654321,
			Name:        "应付款预警",
			Status:      report.AlertStatusActive,
			CreatedBy:   1002,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	// Mock expectations for count
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM \"alerts\" WHERE name LIKE $1 AND status = $2")).
		WithArgs("%预警%", "active").
		WillReturnRows(countRows)

	// Mock expectations for list
	listRows := sqlmock.NewRows([]string{
		"id", "snowflake_id", "name", "description", "sql", "condition", "recipients",
		"push_channel", "trigger_frequency", "silence_start", "silence_end", "status",
		"last_triggered_at", "created_by", "created_at", "updated_at",
	})
	for _, a := range alerts {
		listRows.AddRow(
			a.ID, a.SnowflakeID, a.Name, a.Description, a.SQL, a.Condition, a.Recipients,
			a.PushChannel, a.TriggerFrequency, a.SilenceStart, a.SilenceEnd, a.Status,
			a.LastTriggeredAt, a.CreatedBy, a.CreatedAt, a.UpdatedAt,
		)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM \"alerts\" WHERE name LIKE $1 AND status = $2 ORDER BY created_at DESC LIMIT 2")).
		WithArgs("%预警%", "active").
		WillReturnRows(listRows)

	// Act
	result, total, err := repo.List(ctx, 0, 2, "预警", "active")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestAlertRepository_Update_Success tests successful update.
func TestAlertRepository_Update_Success(t *testing.T) {
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
	repo := NewAlertRepository(gormDB, logger)
	ctx := context.Background()

	updates := map[string]interface{}{
		"name":   "更新后的预警名称",
		"status": report.AlertStatusInactive,
	}

	// Mock expectations
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE \"alerts\" SET \"name\"=$1,\"status\"=$2,\"updated_at\"=$3 WHERE snowflake_id = $4")).
		WithArgs(updates["name"], updates["status"], sqlmock.AnyArg(), 123456789).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Act
	err = repo.Update(ctx, 123456789, updates)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestAlertRepository_Update_NotFound tests update when alert not found.
func TestAlertRepository_Update_NotFound(t *testing.T) {
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
	repo := NewAlertRepository(gormDB, logger)
	ctx := context.Background()

	updates := map[string]interface{}{
		"name": "更新后的预警名称",
	}

	// Mock expectations
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE \"alerts\" SET \"name\"=$1,\"updated_at\"=$2 WHERE snowflake_id = $3")).
		WithArgs(updates["name"], sqlmock.AnyArg(), 123456789).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	// Act
	err = repo.Update(ctx, 123456789, updates)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "alert not found", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestAlertRepository_Delete_Success tests successful deletion.
func TestAlertRepository_Delete_Success(t *testing.T) {
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
	repo := NewAlertRepository(gormDB, logger)
	ctx := context.Background()

	// Mock expectations
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM \"alerts\" WHERE snowflake_id = $1")).
		WithArgs(123456789).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Act
	err = repo.Delete(ctx, 123456789)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestAlertRepository_Delete_NotFound tests deletion when alert not found.
func TestAlertRepository_Delete_NotFound(t *testing.T) {
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
	repo := NewAlertRepository(gormDB, logger)
	ctx := context.Background()

	// Mock expectations
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM \"alerts\" WHERE snowflake_id = $1")).
		WithArgs(123456789).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	// Act
	err = repo.Delete(ctx, 123456789)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "alert not found", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestAlertRepository_GetActiveAlerts_Success tests retrieval of active alerts.
func TestAlertRepository_GetActiveAlerts_Success(t *testing.T) {
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
	repo := NewAlertRepository(gormDB, logger)
	ctx := context.Background()

	now := time.Now()
	alerts := []report.Alert{
		{
			ID:          1,
			SnowflakeID: 123456789,
			Name:        "库存预警",
			Status:      report.AlertStatusActive,
			CreatedBy:   1001,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	// Mock expectations
	rows := sqlmock.NewRows([]string{
		"id", "snowflake_id", "name", "description", "sql", "condition", "recipients",
		"push_channel", "trigger_frequency", "silence_start", "silence_end", "status",
		"last_triggered_at", "created_by", "created_at", "updated_at",
	})
	for _, a := range alerts {
		rows.AddRow(
			a.ID, a.SnowflakeID, a.Name, a.Description, a.SQL, a.Condition, a.Recipients,
			a.PushChannel, a.TriggerFrequency, a.SilenceStart, a.SilenceEnd, a.Status,
			a.LastTriggeredAt, a.CreatedBy, a.CreatedAt, a.UpdatedAt,
		)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM \"alerts\" WHERE status = $1")).
		WithArgs("active").
		WillReturnRows(rows)

	// Act
	result, err := repo.GetActiveAlerts(ctx)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, alerts[0].SnowflakeID, result[0].SnowflakeID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestAlertRepository_UpdateLastTriggeredAt_Success tests successful update of last triggered time.
func TestAlertRepository_UpdateLastTriggeredAt_Success(t *testing.T) {
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
	repo := NewAlertRepository(gormDB, logger)
	ctx := context.Background()

	triggeredAt := time.Now()

	// Mock expectations
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE \"alerts\" SET \"last_triggered_at\"=$1 WHERE snowflake_id = $2")).
		WithArgs(triggeredAt, 123456789).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Act
	err = repo.UpdateLastTriggeredAt(ctx, 123456789, triggeredAt)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
