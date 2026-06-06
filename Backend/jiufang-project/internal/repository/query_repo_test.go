// Package repository_test implements unit tests for query repository.
package repository_test

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"jiufang/internal/model/query"
	"jiufang/internal/repository"
)

// TestQueryRepository_CreateQuerySession_Success tests successful session creation.
func TestQueryRepository_CreateQuerySession_Success(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDB(t)
	require.NoError(t, err)

	repo := repository.NewQueryRepository(db)
	ctx := context.Background()

	session := &query.QuerySession{
		SnowflakeID: 123456789,
		UserID:      1,
		Status:      query.SessionStatusActive,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "query_sessions" ("snowflake_id","user_id","status") VALUES ($1,$2,$3) RETURNING "id","dialog_id","created_at","updated_at","closed_at"`)).
		WithArgs(session.SnowflakeID, session.UserID, session.Status).
		WillReturnRows(sqlmock.NewRows([]string{"id", "dialog_id", "created_at", "updated_at", "closed_at"}).AddRow(1, nil, time.Now(), time.Now(), nil))
	mock.ExpectCommit()

	// Act
	err = repo.CreateQuerySession(ctx, session)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestQueryRepository_CreateQuerySession_Failure tests session creation failure.
func TestQueryRepository_CreateQuerySession_Failure(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDB(t)
	require.NoError(t, err)

	repo := repository.NewQueryRepository(db)
	ctx := context.Background()

	session := &query.QuerySession{
		SnowflakeID: 123456789,
		UserID:      1,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "query_sessions" ("snowflake_id","user_id","status") VALUES ($1,$2,$3) RETURNING "id","dialog_id","created_at","updated_at","closed_at"`)).
		WithArgs(session.SnowflakeID, session.UserID, query.SessionStatusActive).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	// Act
	err = repo.CreateQuerySession(ctx, session)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create query session")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestQueryRepository_GetQuerySessionByID_Success tests successful session retrieval by ID.
func TestQueryRepository_GetQuerySessionByID_Success(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDB(t)
	require.NoError(t, err)

	repo := repository.NewQueryRepository(db)
	ctx := context.Background()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "snowflake_id", "user_id", "dialog_id", "status", "created_at", "updated_at", "closed_at"}).
		AddRow(1, 123456789, 1, nil, query.SessionStatusActive, now, now, nil)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "query_sessions" WHERE "query_sessions"."id" = $1 ORDER BY "query_sessions"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(rows)

	// Act
	session, err := repo.GetQuerySessionByID(ctx, 1)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, uint(1), session.ID)
	assert.Equal(t, int64(123456789), session.SnowflakeID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestQueryRepository_GetQuerySessionByID_NotFound tests session not found by ID.
func TestQueryRepository_GetQuerySessionByID_NotFound(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDB(t)
	require.NoError(t, err)

	repo := repository.NewQueryRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "query_sessions" WHERE "query_sessions"."id" = $1 ORDER BY "query_sessions"."id" LIMIT $2`)).
		WithArgs(999, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// Act
	session, err := repo.GetQuerySessionByID(ctx, 999)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "query session not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestQueryRepository_GetQuerySessionBySnowflakeID_Success tests successful session retrieval by snowflake ID.
func TestQueryRepository_GetQuerySessionBySnowflakeID_Success(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDB(t)
	require.NoError(t, err)

	repo := repository.NewQueryRepository(db)
	ctx := context.Background()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "snowflake_id", "user_id", "dialog_id", "status", "created_at", "updated_at", "closed_at"}).
		AddRow(1, 123456789, 1, nil, query.SessionStatusActive, now, now, nil)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "query_sessions" WHERE snowflake_id = $1 ORDER BY "query_sessions"."id" LIMIT $2`)).
		WithArgs(123456789, 1).
		WillReturnRows(rows)

	// Act
	session, err := repo.GetQuerySessionBySnowflakeID(ctx, 123456789)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, int64(123456789), session.SnowflakeID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestQueryRepository_GetQuerySessionBySnowflakeID_NotFound tests session not found by snowflake ID.
func TestQueryRepository_GetQuerySessionBySnowflakeID_NotFound(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDB(t)
	require.NoError(t, err)

	repo := repository.NewQueryRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "query_sessions" WHERE snowflake_id = $1 ORDER BY "query_sessions"."id" LIMIT $2`)).
		WithArgs(999999, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// Act
	session, err := repo.GetQuerySessionBySnowflakeID(ctx, 999999)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "query session not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestQueryRepository_GetQuerySessionsByUserID_Success tests successful session list retrieval.
func TestQueryRepository_GetQuerySessionsByUserID_Success(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDB(t)
	require.NoError(t, err)

	repo := repository.NewQueryRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Mock count query
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "query_sessions" WHERE user_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Mock select query
	rows := sqlmock.NewRows([]string{"id", "snowflake_id", "user_id", "dialog_id", "status", "created_at", "updated_at", "closed_at"}).
		AddRow(1, 123456789, 1, nil, query.SessionStatusActive, now, now, nil).
		AddRow(2, 123456790, 1, nil, query.SessionStatusClosed, now, now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "query_sessions" WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`)).
		WithArgs(1, 10).
		WillReturnRows(rows)

	// Act
	sessions, total, err := repo.GetQuerySessionsByUserID(ctx, 1, 0, 10)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, sessions, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestQueryRepository_GetQuerySessionsByUserID_EmptyList tests empty session list.
func TestQueryRepository_GetQuerySessionsByUserID_EmptyList(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDB(t)
	require.NoError(t, err)

	repo := repository.NewQueryRepository(db)
	ctx := context.Background()

	// Mock count query
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "query_sessions" WHERE user_id = $1`)).
		WithArgs(999).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock select query
	rows := sqlmock.NewRows([]string{"id", "snowflake_id", "user_id", "dialog_id", "status", "created_at", "updated_at", "closed_at"})

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "query_sessions" WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`)).
		WithArgs(999, 10).
		WillReturnRows(rows)

	// Act
	sessions, total, err := repo.GetQuerySessionsByUserID(ctx, 999, 0, 10)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, sessions, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestQueryRepository_CloseQuerySession_Success tests successful session close.
func TestQueryRepository_CloseQuerySession_Success(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDB(t)
	require.NoError(t, err)

	repo := repository.NewQueryRepository(db)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "query_sessions" SET "closed_at"=$1,"status"=$2,"updated_at"=$3 WHERE snowflake_id = $4`)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Act
	err = repo.CloseQuerySession(ctx, 123456789)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestQueryRepository_CloseQuerySession_NotFound tests session close not found.
func TestQueryRepository_CloseQuerySession_NotFound(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDB(t)
	require.NoError(t, err)

	repo := repository.NewQueryRepository(db)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "query_sessions" SET "closed_at"=$1,"status"=$2,"updated_at"=$3 WHERE snowflake_id = $4`)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	// Act
	err = repo.CloseQuerySession(ctx, 999999)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query session not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestQueryRepository_CreateQueryRecord_Success tests successful record creation.
func TestQueryRepository_CreateQueryRecord_Success(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDB(t)
	require.NoError(t, err)

	repo := repository.NewQueryRepository(db)
	ctx := context.Background()

	record := &query.QueryRecord{
		SnowflakeID:   123456789,
		SessionID:     1,
		UserID:        1,
		Input:         "test query",
		SQL:           "SELECT * FROM test",
		Status:        query.QueryStatusSuccess,
		ResultCount:   10,
		ExecutionTime: 100,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "query_records"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	// Act
	err = repo.CreateQueryRecord(ctx, record)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestQueryRepository_GetQueryRecordBySnowflakeID_Success tests successful record retrieval.
func TestQueryRepository_GetQueryRecordBySnowflakeID_Success(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDB(t)
	require.NoError(t, err)

	repo := repository.NewQueryRepository(db)
	ctx := context.Background()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "snowflake_id", "session_id", "user_id", "input", "sql", "status", "error_message", "result_count", "execution_time", "result_data", "created_at"}).
		AddRow(1, 123456789, 1, 1, "test query", "SELECT * FROM test", query.QueryStatusSuccess, nil, 10, 100, nil, now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "query_records" WHERE snowflake_id = $1 ORDER BY "query_records"."id" LIMIT $2`)).
		WithArgs(123456789, 1).
		WillReturnRows(rows)

	// Act
	record, err := repo.GetQueryRecordBySnowflakeID(ctx, 123456789)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, record)
	assert.Equal(t, int64(123456789), record.SnowflakeID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestQueryRepository_GetQueryRecordsByUserID_Success tests successful record list retrieval.
func TestQueryRepository_GetQueryRecordsByUserID_Success(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDB(t)
	require.NoError(t, err)

	repo := repository.NewQueryRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Mock count query
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "query_records" WHERE user_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Mock select query
	rows := sqlmock.NewRows([]string{"id", "snowflake_id", "session_id", "user_id", "input", "sql", "status", "error_message", "result_count", "execution_time", "result_data", "created_at"}).
		AddRow(1, 123456789, 1, 1, "query1", "SELECT 1", query.QueryStatusSuccess, nil, 5, 50, nil, now).
		AddRow(2, 123456790, 1, 1, "query2", "SELECT 2", query.QueryStatusFailed, "error", 0, 0, nil, now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "query_records" WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`)).
		WithArgs(1, 10).
		WillReturnRows(rows)

	// Act
	records, total, err := repo.GetQueryRecordsByUserID(ctx, 1, 0, 10, "", "", "")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, records, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestQueryRepository_GetQueryRecordsByUserID_WithTimeFilter tests record list with time filter.
func TestQueryRepository_GetQueryRecordsByUserID_WithTimeFilter(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDB(t)
	require.NoError(t, err)

	repo := repository.NewQueryRepository(db)
	ctx := context.Background()

	startTime := "2024-01-01T00:00:00Z"
	endTime := "2024-12-31T23:59:59Z"

	// Mock count query with time filter
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "query_records" WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Mock select query with time filter
	rows := sqlmock.NewRows([]string{"id", "snowflake_id", "session_id", "user_id", "input", "sql", "status", "error_message", "result_count", "execution_time", "result_data", "created_at"}).
		AddRow(1, 123456789, 1, 1, "query", "SELECT 1", query.QueryStatusSuccess, nil, 5, 50, nil, time.Now())

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "query_records" WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3 ORDER BY created_at DESC LIMIT $4`)).
		WillReturnRows(rows)

	// Act
	records, total, err := repo.GetQueryRecordsByUserID(ctx, 1, 0, 10, startTime, endTime, "")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, records, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestQueryRepository_DeleteQueryRecord_Success tests successful record deletion.
func TestQueryRepository_DeleteQueryRecord_Success(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDB(t)
	require.NoError(t, err)

	repo := repository.NewQueryRepository(db)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "query_records" WHERE snowflake_id = $1 AND user_id = $2`)).
		WithArgs(123456789, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Act
	err = repo.DeleteQueryRecord(ctx, 123456789, 1)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestQueryRepository_DeleteQueryRecord_NotOwned tests record deletion not owned by user.
func TestQueryRepository_DeleteQueryRecord_NotOwned(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDB(t)
	require.NoError(t, err)

	repo := repository.NewQueryRepository(db)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "query_records" WHERE snowflake_id = $1 AND user_id = $2`)).
		WithArgs(123456789, 999).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	// Act
	err = repo.DeleteQueryRecord(ctx, 123456789, 999)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not owned by user")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestQueryRepository_CreateFavorite_Success tests successful favorite creation.
func TestQueryRepository_CreateFavorite_Success(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDB(t)
	require.NoError(t, err)

	repo := repository.NewQueryRepository(db)
	ctx := context.Background()

	favorite := &query.Favorite{
		SnowflakeID: 123456789,
		UserID:      1,
		Name:        "My Favorite",
		Input:       "test query",
		Sql:         "SELECT * FROM test",
		Description: "This is a test favorite",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "favorites" ("snowflake_id","user_id","query_record_id","name","input","sql","description") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "id","created_at"`)).
		WithArgs(favorite.SnowflakeID, favorite.UserID, nil, favorite.Name, favorite.Input, favorite.Sql, favorite.Description).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, time.Now()))
	mock.ExpectCommit()

	// Act
	err = repo.CreateFavorite(ctx, favorite)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestQueryRepository_IsFavoriteNameExists_True tests favorite name exists check returns true.
func TestQueryRepository_IsFavoriteNameExists_True(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDB(t)
	require.NoError(t, err)

	repo := repository.NewQueryRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "favorites" WHERE user_id = $1 AND name = $2`)).
		WithArgs(1, "test favorite").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Act
	exists, err := repo.IsFavoriteNameExists(ctx, 1, "test favorite")

	// Assert
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestQueryRepository_IsFavoriteNameExists_False tests favorite name exists check returns false.
func TestQueryRepository_IsFavoriteNameExists_False(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDB(t)
	require.NoError(t, err)

	repo := repository.NewQueryRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "favorites" WHERE user_id = $1 AND name = $2`)).
		WithArgs(1, "non-existent favorite").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Act
	exists, err := repo.IsFavoriteNameExists(ctx, 1, "non-existent favorite")

	// Assert
	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestQueryRepository_DeleteFavorite_Success tests successful favorite deletion.
func TestQueryRepository_DeleteFavorite_Success(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDB(t)
	require.NoError(t, err)

	repo := repository.NewQueryRepository(db)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "favorites" WHERE snowflake_id = $1 AND user_id = $2`)).
		WithArgs(123456789, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Act
	err = repo.DeleteFavorite(ctx, 123456789, 1)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestQueryRepository_DeleteFavorite_NotOwned tests favorite deletion not owned by user.
func TestQueryRepository_DeleteFavorite_NotOwned(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDB(t)
	require.NoError(t, err)

	repo := repository.NewQueryRepository(db)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "favorites" WHERE snowflake_id = $1 AND user_id = $2`)).
		WithArgs(123456789, 999).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	// Act
	err = repo.DeleteFavorite(ctx, 123456789, 999)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not owned by user")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// setupMockDB creates a mock database connection for testing.
func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, error) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	return gormDB, mock, nil
}
