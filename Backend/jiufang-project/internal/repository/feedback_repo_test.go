// Package repository_test implements unit tests for FeedbackRepository.
// Author: AI Assistant
// Date: 2026-06-04
// Description: Unit tests for feedback data access layer

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"jiufang/internal/model/feedback"
	"jiufang/internal/repository"
)

// TestFeedbackRepository_Create_Success tests TC-FR-01
func TestFeedbackRepository_Create_Success(t *testing.T) {
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

	repo := repository.NewFeedbackRepository(gormDB)
	ctx := context.Background()

	fb := &feedback.Feedback{
		SnowflakeID:   123456789,
		UserID:        1001,
		QueryRecordID: 987654321,
		QueryQuestion: "查询本月销售额",
		Rating:        feedback.RatingSatisfied,
		Reason:        "",
		CreatedAt:     time.Now(),
	}

	// Mock expectations
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `feedbacks`").
		WithArgs(
			fb.SnowflakeID,
			fb.UserID,
			fb.QueryRecordID,
			fb.QueryQuestion,
			fb.Rating,
			fb.Reason,
			fb.CreatedAt,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Act
	err = repo.Create(ctx, fb)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestFeedbackRepository_GetByID_Success tests TC-FR-02
func TestFeedbackRepository_GetByID_Success(t *testing.T) {
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

	repo := repository.NewFeedbackRepository(gormDB)
	ctx := context.Background()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "snowflake_id", "user_id", "query_record_id", "query_question", "rating", "reason", "created_at"}).
		AddRow(1, 123456789, 1001, 987654321, "查询本月销售额", "satisfied", "", now)

	mock.ExpectQuery("SELECT \\* FROM `feedbacks` WHERE `feedbacks`.`id` = \\$1 ORDER BY `feedbacks`.`id` LIMIT 1").
		WithArgs(1).
		WillReturnRows(rows)

	// Act
	fb, err := repo.GetByID(ctx, 1)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, fb)
	assert.Equal(t, int64(123456789), fb.SnowflakeID)
	assert.Equal(t, int64(1001), fb.UserID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestFeedbackRepository_GetByID_NotFound tests TC-FR-03
func TestFeedbackRepository_GetByID_NotFound(t *testing.T) {
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

	repo := repository.NewFeedbackRepository(gormDB)
	ctx := context.Background()

	mock.ExpectQuery("SELECT \\* FROM `feedbacks` WHERE `feedbacks`.`id` = \\$1 ORDER BY `feedbacks`.`id` LIMIT 1").
		WithArgs(999).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	// Act
	fb, err := repo.GetByID(ctx, 999)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, fb)
	assert.Contains(t, err.Error(), "feedback not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestFeedbackRepository_GetBySnowflakeID_Success tests TC-FR-04
func TestFeedbackRepository_GetBySnowflakeID_Success(t *testing.T) {
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

	repo := repository.NewFeedbackRepository(gormDB)
	ctx := context.Background()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "snowflake_id", "user_id", "query_record_id", "query_question", "rating", "reason", "created_at"}).
		AddRow(1, 123456789, 1001, 987654321, "查询本月销售额", "satisfied", "", now)

	mock.ExpectQuery("SELECT \\* FROM `feedbacks` WHERE snowflake_id = \\$1 ORDER BY `feedbacks`.`id` LIMIT 1").
		WithArgs(123456789).
		WillReturnRows(rows)

	// Act
	fb, err := repo.GetBySnowflakeID(ctx, 123456789)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, fb)
	assert.Equal(t, int64(123456789), fb.SnowflakeID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestFeedbackRepository_GetByQueryRecordID_Success tests TC-FR-06
func TestFeedbackRepository_GetByQueryRecordID_Success(t *testing.T) {
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

	repo := repository.NewFeedbackRepository(gormDB)
	ctx := context.Background()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "snowflake_id", "user_id", "query_record_id", "query_question", "rating", "reason", "created_at"}).
		AddRow(1, 123456789, 1001, 987654321, "查询本月销售额", "satisfied", "", now)

	mock.ExpectQuery("SELECT \\* FROM `feedbacks` WHERE query_record_id = \\$1 ORDER BY `feedbacks`.`id` LIMIT 1").
		WithArgs(987654321).
		WillReturnRows(rows)

	// Act
	fb, err := repo.GetByQueryRecordID(ctx, 987654321)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, fb)
	assert.Equal(t, int64(987654321), fb.QueryRecordID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestFeedbackRepository_IsFeedbackExists_True tests TC-FR-10
func TestFeedbackRepository_IsFeedbackExists_True(t *testing.T) {
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

	repo := repository.NewFeedbackRepository(gormDB)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `feedbacks` WHERE query_record_id = \\$1").
		WithArgs(987654321).
		WillReturnRows(rows)

	// Act
	exists, err := repo.IsFeedbackExists(ctx, 987654321)

	// Assert
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestFeedbackRepository_IsFeedbackExists_False tests TC-FR-11
func TestFeedbackRepository_IsFeedbackExists_False(t *testing.T) {
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

	repo := repository.NewFeedbackRepository(gormDB)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `feedbacks` WHERE query_record_id = \\$1").
		WithArgs(999).
		WillReturnRows(rows)

	// Act
	exists, err := repo.IsFeedbackExists(ctx, 999)

	// Assert
	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}