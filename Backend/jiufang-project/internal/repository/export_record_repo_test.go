// Package repository_test tests the export record repository implementation.
// Author: AI Agent
// Date: 2026-06-03
// Description: Unit tests for ExportRecordRepository covering CRUD operations and error scenarios.

package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"jiufang/internal/model/export"
	"jiufang/internal/repository"
)

// TestExportRecordRepository_CreateExportRecord tests the CreateExportRecord method.
func TestExportRecordRepository_CreateExportRecord(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	repo := repository.NewExportRecordRepository(gormDB)
	ctx := context.Background()

	testRecord := &export.ExportRecord{
		SnowflakeID:   123456789,
		UserID:        100,
		QueryRecordID: 200,
		Format:        export.ExportFormatExcel,
		FileName:      "test_export.xlsx",
		FileSize:      1024,
		QuerySummary:  "test query",
		CreatedAt:     time.Now(),
	}

	t.Run("TC-REPO-01: Normal creation", func(t *testing.T) {
		// Arrange
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO `export_records`").
			WithArgs(
				testRecord.SnowflakeID,
				testRecord.UserID,
				testRecord.QueryRecordID,
				testRecord.Format,
				testRecord.FileName,
				testRecord.FileSize,
				testRecord.QuerySummary,
				testRecord.CreatedAt,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		// Act
		err := repo.CreateExportRecord(ctx, testRecord)

		// Assert
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("TC-REPO-02: Database error", func(t *testing.T) {
		// Arrange
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO `export_records`").
			WillReturnError(errors.New("database connection failed"))
		mock.ExpectRollback()

		// Act
		err := repo.CreateExportRecord(ctx, testRecord)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create export record")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestExportRecordRepository_GetExportRecordByID tests the GetExportRecordByID method.
func TestExportRecordRepository_GetExportRecordByID(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	repo := repository.NewExportRecordRepository(gormDB)
	ctx := context.Background()

	testRecord := &export.ExportRecord{
		ID:            1,
		SnowflakeID:   123456789,
		UserID:        100,
		QueryRecordID: 200,
		Format:        export.ExportFormatExcel,
		FileName:      "test_export.xlsx",
		FileSize:      1024,
		QuerySummary:  "test query",
		CreatedAt:     time.Now(),
	}

	t.Run("TC-REPO-03: Normal query by ID", func(t *testing.T) {
		// Arrange
		rows := sqlmock.NewRows([]string{
			"id", "snowflake_id", "user_id", "query_record_id", "format", "file_name", "file_size", "query_summary", "created_at",
		}).AddRow(
			testRecord.ID,
			testRecord.SnowflakeID,
			testRecord.UserID,
			testRecord.QueryRecordID,
			testRecord.Format,
			testRecord.FileName,
			testRecord.FileSize,
			testRecord.QuerySummary,
			testRecord.CreatedAt,
		)

		mock.ExpectQuery("SELECT \\* FROM `export_records` WHERE `export_records`.`id` = \\? ORDER BY `export_records`.`id` LIMIT \\?").
			WithArgs(1, 1).
			WillReturnRows(rows)

		// Act
		result, err := repo.GetExportRecordByID(ctx, 1)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, testRecord.ID, result.ID)
		assert.Equal(t, testRecord.SnowflakeID, result.SnowflakeID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("TC-REPO-04: Record not found", func(t *testing.T) {
		// Arrange
		mock.ExpectQuery("SELECT \\* FROM `export_records` WHERE `export_records`.`id` = \\? ORDER BY `export_records`.`id` LIMIT \\?").
			WithArgs(999, 1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "snowflake_id", "user_id", "query_record_id", "format", "file_name", "file_size", "query_summary", "created_at",
			}))

		// Act
		result, err := repo.GetExportRecordByID(ctx, 999)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "export record not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestExportRecordRepository_GetExportRecordBySnowflakeID tests the GetExportRecordBySnowflakeID method.
func TestExportRecordRepository_GetExportRecordBySnowflakeID(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	repo := repository.NewExportRecordRepository(gormDB)
	ctx := context.Background()

	testRecord := &export.ExportRecord{
		ID:            1,
		SnowflakeID:   123456789,
		UserID:        100,
		QueryRecordID: 200,
		Format:        export.ExportFormatExcel,
		FileName:      "test_export.xlsx",
		FileSize:      1024,
		QuerySummary:  "test query",
		CreatedAt:     time.Now(),
	}

	t.Run("TC-REPO-05: Normal query by snowflake ID", func(t *testing.T) {
		// Arrange
		rows := sqlmock.NewRows([]string{
			"id", "snowflake_id", "user_id", "query_record_id", "format", "file_name", "file_size", "query_summary", "created_at",
		}).AddRow(
			testRecord.ID,
			testRecord.SnowflakeID,
			testRecord.UserID,
			testRecord.QueryRecordID,
			testRecord.Format,
			testRecord.FileName,
			testRecord.FileSize,
			testRecord.QuerySummary,
			testRecord.CreatedAt,
		)

		mock.ExpectQuery("SELECT \\* FROM `export_records` WHERE snowflake_id = \\? ORDER BY `export_records`.`id` LIMIT \\?").
			WithArgs(123456789, 1).
			WillReturnRows(rows)

		// Act
		result, err := repo.GetExportRecordBySnowflakeID(ctx, 123456789)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, testRecord.SnowflakeID, result.SnowflakeID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("TC-REPO-06: Record not found by snowflake ID", func(t *testing.T) {
		// Arrange
		mock.ExpectQuery("SELECT \\* FROM `export_records` WHERE snowflake_id = \\? ORDER BY `export_records`.`id` LIMIT \\?").
			WithArgs(999999999, 1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "snowflake_id", "user_id", "query_record_id", "format", "file_name", "file_size", "query_summary", "created_at",
			}))

		// Act
		result, err := repo.GetExportRecordBySnowflakeID(ctx, 999999999)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "export record not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestExportRecordRepository_GetExportRecordsByUserID tests the GetExportRecordsByUserID method.
func TestExportRecordRepository_GetExportRecordsByUserID(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	repo := repository.NewExportRecordRepository(gormDB)
	ctx := context.Background()

	testRecords := []export.ExportRecord{
		{
			ID:            1,
			SnowflakeID:   123456789,
			UserID:        100,
			QueryRecordID: 200,
			Format:        export.ExportFormatExcel,
			FileName:      "export1.xlsx",
			FileSize:      1024,
			CreatedAt:     time.Now(),
		},
		{
			ID:            2,
			SnowflakeID:   123456790,
			UserID:        100,
			QueryRecordID: 201,
			Format:        export.ExportFormatPDF,
			FileName:      "export2.pdf",
			FileSize:      2048,
			CreatedAt:     time.Now(),
		},
	}

	t.Run("TC-REPO-07: Normal query with pagination", func(t *testing.T) {
		// Arrange - count query
		mock.ExpectQuery("SELECT count\\(\\*\\) FROM `export_records` WHERE user_id = \\?").
			WithArgs(100).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		// Arrange - data query
		rows := sqlmock.NewRows([]string{
			"id", "snowflake_id", "user_id", "query_record_id", "format", "file_name", "file_size", "query_summary", "created_at",
		})
		for _, record := range testRecords {
			rows.AddRow(
				record.ID,
				record.SnowflakeID,
				record.UserID,
				record.QueryRecordID,
				record.Format,
				record.FileName,
				record.FileSize,
				record.QuerySummary,
				record.CreatedAt,
			)
		}

		mock.ExpectQuery("SELECT \\* FROM `export_records` WHERE user_id = \\? ORDER BY created_at DESC LIMIT \\?").
			WithArgs(100, 10).
			WillReturnRows(rows)

		// Act
		records, total, err := repo.GetExportRecordsByUserID(ctx, 100, 0, 10)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, records, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("TC-REPO-08: Empty list for user with no exports", func(t *testing.T) {
		// Arrange - count query
		mock.ExpectQuery("SELECT count\\(\\*\\) FROM `export_records` WHERE user_id = \\?").
			WithArgs(999).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		// Arrange - data query
		mock.ExpectQuery("SELECT \\* FROM `export_records` WHERE user_id = \\? ORDER BY created_at DESC LIMIT \\?").
			WithArgs(999, 10).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "snowflake_id", "user_id", "query_record_id", "format", "file_name", "file_size", "query_summary", "created_at",
			}))

		// Act
		records, total, err := repo.GetExportRecordsByUserID(ctx, 999, 0, 10)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Len(t, records, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
