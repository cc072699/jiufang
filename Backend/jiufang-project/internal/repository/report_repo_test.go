// Package repository_test implements unit tests for ReportRepository.
// Author: AI Assistant
// Date: 2026-06-04
// Description: Unit tests for scheduled report data access layer

package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"jiufang/internal/model/report"
	"jiufang/internal/repository"
)

// TestReportRepository_Create_Success tests TC-RR-01
func TestReportRepository_Create_Success(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	repo := repository.NewReportRepository(gormDB)
	ctx := context.Background()

	scheduledReport := &report.ScheduledReport{
		SnowflakeID:  123456789,
		Name:         "Test Report",
		Description:  "Test Description",
		SQL:          "SELECT * FROM sales WHERE date >= CURDATE()",
		ScheduleType: report.ScheduleTypeDaily,
		ScheduleTime: "09:00:00",
		Recipients:   `[1001]`,
		Status:       report.ReportStatusActive,
		CreatedBy:    1001,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Mock expectations
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `scheduled_reports`").
		WithArgs(
			scheduledReport.SnowflakeID,
			scheduledReport.Name,
			scheduledReport.Description,
			scheduledReport.SQL,
			scheduledReport.ScheduleType,
			scheduledReport.ScheduleTime,
			scheduledReport.Recipients,
			scheduledReport.Status,
			scheduledReport.CreatedBy,
			scheduledReport.CreatedAt,
			scheduledReport.UpdatedAt,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Act
	err = repo.Create(ctx, scheduledReport)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestReportRepository_GetBySnowflakeID_Success tests TC-RR-02
func TestReportRepository_GetBySnowflakeID_Success(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	repo := repository.NewReportRepository(gormDB)
	ctx := context.Background()
	snowflakeID := int64(123456789)

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "snowflake_id", "name", "description", "sql",
		"schedule_type", "schedule_time", "recipients", "status",
		"created_by", "created_at", "updated_at",
	}).AddRow(
		1, snowflakeID, "Test Report", "Test Description", "SELECT * FROM sales WHERE date >= CURDATE()",
		"daily", "09:00:00", `[1001]`, "active",
		1001, now, now,
	)

	// Mock expectations
	mock.ExpectQuery("SELECT \\* FROM `scheduled_reports` WHERE snowflake_id = \\? ORDER BY `scheduled_reports`\\.`id` LIMIT \\?").
		WithArgs(snowflakeID, 1).
		WillReturnRows(rows)

	// Act
	result, err := repo.GetBySnowflakeID(ctx, snowflakeID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, snowflakeID, result.SnowflakeID)
	assert.Equal(t, "Test Report", result.Name)
	assert.Equal(t, "SELECT * FROM sales WHERE date >= CURDATE()", result.SQL)
	assert.Equal(t, report.ScheduleTypeDaily, result.ScheduleType)
	assert.Equal(t, "09:00:00", result.ScheduleTime)
	assert.Equal(t, `[1001]`, result.Recipients)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestReportRepository_GetBySnowflakeID_NotFound tests TC-RR-03
func TestReportRepository_GetBySnowflakeID_NotFound(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	repo := repository.NewReportRepository(gormDB)
	ctx := context.Background()
	snowflakeID := int64(999999)

	// Mock expectations
	mock.ExpectQuery("SELECT \\* FROM `scheduled_reports` WHERE snowflake_id = \\? ORDER BY `scheduled_reports`\\.`id` LIMIT \\?").
		WithArgs(snowflakeID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	// Act
	result, err := repo.GetBySnowflakeID(ctx, snowflakeID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "scheduled report not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestReportRepository_List_Success tests TC-RR-04
func TestReportRepository_List_Success(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	repo := repository.NewReportRepository(gormDB)
	ctx := context.Background()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "snowflake_id", "name", "description", "sql",
		"schedule_type", "schedule_time", "recipients", "status",
		"created_by", "created_at", "updated_at",
	}).AddRow(
		1, 123456789, "Report 1", "Description 1", "SELECT * FROM sales WHERE date >= CURDATE()",
		"daily", "09:00:00", `[1001]`, "active",
		1001, now, now,
	).AddRow(
		2, 123456790, "Report 2", "Description 2", "SELECT * FROM orders WHERE date >= CURDATE()",
		"daily", "10:00:00", `[1002]`, "active",
		1002, now, now,
	)

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)

	// Mock expectations
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `scheduled_reports` WHERE status = \\?").
		WithArgs("active").
		WillReturnRows(countRows)
	mock.ExpectQuery("SELECT \\* FROM `scheduled_reports` WHERE status = \\? ORDER BY created_at DESC LIMIT \\?").
		WithArgs("active", 10).
		WillReturnRows(rows)

	// Act
	results, total, err := repo.List(ctx, 0, 10, "active")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 2, len(results))
	assert.Equal(t, int64(2), total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestReportRepository_Update_Success tests TC-RR-05
func TestReportRepository_Update_Success(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	repo := repository.NewReportRepository(gormDB)
	ctx := context.Background()
	snowflakeID := int64(123456789)
	updates := map[string]interface{}{
		"name":        "Updated Report",
		"description": "Updated Description",
		"status":      "inactive",
	}

	// Mock expectations
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `scheduled_reports`").
		WithArgs(
			"Updated Description",
			"Updated Report",
			"inactive",
			sqlmock.AnyArg(),
			snowflakeID,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Act
	err = repo.Update(ctx, snowflakeID, updates)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestReportRepository_Delete_Success tests TC-RR-06
func TestReportRepository_Delete_Success(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	repo := repository.NewReportRepository(gormDB)
	ctx := context.Background()
	snowflakeID := int64(123456789)

	// Mock expectations
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `scheduled_reports`").
		WithArgs(snowflakeID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Act
	err = repo.Delete(ctx, snowflakeID)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestReportRepository_GetActiveReports_Success tests TC-RR-07
func TestReportRepository_GetActiveReports_Success(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	repo := repository.NewReportRepository(gormDB)
	ctx := context.Background()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "snowflake_id", "name", "description", "query_condition",
		"schedule_cron", "push_targets", "push_channel", "status",
		"created_by", "created_at", "updated_at",
	}).AddRow(
		1, 123456789, "Active Report 1", "Description 1", `{"table": "sales"}`,
		"0 0 9 * * *", `[1001]`, "wechat", "active",
		1001, now, now,
	).AddRow(
		2, 123456790, "Active Report 2", "Description 2", `{"table": "orders"}`,
		"0 0 10 * * *", `[1002]`, "wechat", "active",
		1002, now, now,
	)

	// Mock expectations
	mock.ExpectQuery("SELECT \\* FROM `scheduled_reports`").
		WithArgs("active").
		WillReturnRows(rows)

	// Act
	results, err := repo.GetActiveReports(ctx)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 2, len(results))
	assert.Equal(t, report.ReportStatusActive, results[0].Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestReportRepository_Create_DatabaseError tests database error handling
func TestReportRepository_Create_DatabaseError(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	repo := repository.NewReportRepository(gormDB)
	ctx := context.Background()

	scheduledReport := &report.ScheduledReport{
		SnowflakeID:  123456789,
		Name:         "Test Report",
		Description:  "Test Description",
		SQL:          "SELECT * FROM sales WHERE date >= CURDATE()",
		ScheduleType: report.ScheduleTypeDaily,
		ScheduleTime: "09:00:00",
		Recipients:   `[1001]`,
		Status:       report.ReportStatusActive,
		CreatedBy:    1001,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Mock expectations
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `scheduled_reports`").
		WillReturnError(errors.New("database connection error"))
	mock.ExpectRollback()

	// Act
	err = repo.Create(ctx, scheduledReport)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create scheduled report")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestReportRepository_CreatePushRecord_Success tests push record creation
func TestReportRepository_CreatePushRecord_Success(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	repo := repository.NewReportRepository(gormDB)
	ctx := context.Background()

	pushRecord := &report.PushRecord{
		SnowflakeID: 123456789,
		ReportID:    123456788,
		PushType:    report.PushTypeReport,
		PushContent: "Test content",
		PushTargets: `[1001]`,
		PushChannel: report.PushChannelWeChat,
		PushStatus:  report.PushStatusSuccess,
		PushTime:    time.Now(),
		RetryCount:  0,
	}

	// Mock expectations
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `push_records`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Act
	err = repo.CreatePushRecord(ctx, pushRecord)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
