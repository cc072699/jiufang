// Package repository_test implements unit tests for dialog favorite repository.
// Author: AI Agent
// Date: 2026-06-03
// Description: Tests all methods of DialogFavoriteRepository using go-sqlmock

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

	"jiufang/internal/model/dialog"
	"jiufang/internal/repository"
)

// TestDialogFavoriteRepository_CreateDialogFavorite_Success tests successful favorite creation.
func TestDialogFavoriteRepository_CreateDialogFavorite_Success(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDBForFavorite(t)
	require.NoError(t, err)

	repo := repository.NewDialogFavoriteRepository(db)
	ctx := context.Background()

	favorite := &dialog.DialogFavorite{
		SnowflakeID:     123456789,
		UserID:          1,
		DialogSessionID: 100,
		Title:           "Test Favorite",
		CreatedAt:       time.Now(),
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "dialog_favorites" ("snowflake_id","user_id","dialog_session_id","title","created_at") VALUES ($1,$2,$3,$4,$5) RETURNING "id","title","created_at"`)).
		WithArgs(favorite.SnowflakeID, favorite.UserID, favorite.DialogSessionID, favorite.Title, favorite.CreatedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "created_at"}).AddRow(1, favorite.Title, favorite.CreatedAt))
	mock.ExpectCommit()

	// Act
	err = repo.CreateDialogFavorite(ctx, favorite)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestDialogFavoriteRepository_CreateDialogFavorite_Failure tests favorite creation failure.
func TestDialogFavoriteRepository_CreateDialogFavorite_Failure(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDBForFavorite(t)
	require.NoError(t, err)

	repo := repository.NewDialogFavoriteRepository(db)
	ctx := context.Background()

	favorite := &dialog.DialogFavorite{
		SnowflakeID:     123456789,
		UserID:          1,
		DialogSessionID: 100,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "dialog_favorites" ("snowflake_id","user_id","dialog_session_id") VALUES ($1,$2,$3) RETURNING "id","title","created_at"`)).
		WithArgs(favorite.SnowflakeID, favorite.UserID, favorite.DialogSessionID).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	// Act
	err = repo.CreateDialogFavorite(ctx, favorite)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create dialog favorite")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestDialogFavoriteRepository_GetDialogFavoriteByID_Success tests successful retrieval by ID.
func TestDialogFavoriteRepository_GetDialogFavoriteByID_Success(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDBForFavorite(t)
	require.NoError(t, err)

	repo := repository.NewDialogFavoriteRepository(db)
	ctx := context.Background()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "snowflake_id", "user_id", "dialog_session_id", "title", "created_at"}).
		AddRow(1, 123456789, 1, 100, "Test Favorite", now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dialog_favorites" WHERE "dialog_favorites"."id" = $1 ORDER BY "dialog_favorites"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(rows)

	// Act
	favorite, err := repo.GetDialogFavoriteByID(ctx, 1)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, favorite)
	assert.Equal(t, int64(123456789), favorite.SnowflakeID)
	assert.Equal(t, int64(1), favorite.UserID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestDialogFavoriteRepository_GetDialogFavoriteByID_NotFound tests retrieval when record not found.
func TestDialogFavoriteRepository_GetDialogFavoriteByID_NotFound(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDBForFavorite(t)
	require.NoError(t, err)

	repo := repository.NewDialogFavoriteRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dialog_favorites" WHERE "dialog_favorites"."id" = $1 ORDER BY "dialog_favorites"."id" LIMIT $2`)).
		WithArgs(999, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "snowflake_id", "user_id", "dialog_session_id", "title", "created_at"}))

	// Act
	favorite, err := repo.GetDialogFavoriteByID(ctx, 999)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, favorite)
	assert.Contains(t, err.Error(), "dialog favorite not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestDialogFavoriteRepository_GetDialogFavoriteBySnowflakeID_Success tests successful retrieval by snowflake ID.
func TestDialogFavoriteRepository_GetDialogFavoriteBySnowflakeID_Success(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDBForFavorite(t)
	require.NoError(t, err)

	repo := repository.NewDialogFavoriteRepository(db)
	ctx := context.Background()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "snowflake_id", "user_id", "dialog_session_id", "title", "created_at"}).
		AddRow(1, 123456789, 1, 100, "Test Favorite", now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dialog_favorites" WHERE snowflake_id = $1 ORDER BY "dialog_favorites"."id" LIMIT $2`)).
		WithArgs(123456789, 1).
		WillReturnRows(rows)

	// Act
	favorite, err := repo.GetDialogFavoriteBySnowflakeID(ctx, 123456789)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, favorite)
	assert.Equal(t, int64(123456789), favorite.SnowflakeID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestDialogFavoriteRepository_GetDialogFavoriteBySnowflakeID_NotFound tests retrieval when record not found.
func TestDialogFavoriteRepository_GetDialogFavoriteBySnowflakeID_NotFound(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDBForFavorite(t)
	require.NoError(t, err)

	repo := repository.NewDialogFavoriteRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dialog_favorites" WHERE snowflake_id = $1 ORDER BY "dialog_favorites"."id" LIMIT $2`)).
		WithArgs(999999, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "snowflake_id", "user_id", "dialog_session_id", "title", "created_at"}))

	// Act
	favorite, err := repo.GetDialogFavoriteBySnowflakeID(ctx, 999999)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, favorite)
	assert.Contains(t, err.Error(), "dialog favorite not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestDialogFavoriteRepository_GetDialogFavoritesByUserID_Success tests successful list retrieval.
func TestDialogFavoriteRepository_GetDialogFavoritesByUserID_Success(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDBForFavorite(t)
	require.NoError(t, err)

	repo := repository.NewDialogFavoriteRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Mock count query
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "dialog_favorites" WHERE user_id = $1`)).
		WithArgs(1).
		WillReturnRows(countRows)

	// Mock list query
	listRows := sqlmock.NewRows([]string{"id", "snowflake_id", "user_id", "dialog_session_id", "title", "created_at"}).
		AddRow(1, 123456789, 1, 100, "Favorite 1", now).
		AddRow(2, 123456790, 1, 101, "Favorite 2", now.Add(-1*time.Hour))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dialog_favorites" WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`)).
		WithArgs(1, 10).
		WillReturnRows(listRows)

	// Act
	favorites, total, err := repo.GetDialogFavoritesByUserID(ctx, 1, 0, 10)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, favorites, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestDialogFavoriteRepository_GetDialogFavoritesByUserID_EmptyList tests retrieval when no favorites exist.
func TestDialogFavoriteRepository_GetDialogFavoritesByUserID_EmptyList(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDBForFavorite(t)
	require.NoError(t, err)

	repo := repository.NewDialogFavoriteRepository(db)
	ctx := context.Background()

	// Mock count query
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "dialog_favorites" WHERE user_id = $1`)).
		WithArgs(999).
		WillReturnRows(countRows)

	// Mock list query (empty)
	listRows := sqlmock.NewRows([]string{"id", "snowflake_id", "user_id", "dialog_session_id", "title", "created_at"})
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dialog_favorites" WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`)).
		WithArgs(999, 10).
		WillReturnRows(listRows)

	// Act
	favorites, total, err := repo.GetDialogFavoritesByUserID(ctx, 999, 0, 10)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, favorites, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestDialogFavoriteRepository_GetDialogFavoritesByUserID_CountError tests count query failure.
func TestDialogFavoriteRepository_GetDialogFavoritesByUserID_CountError(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDBForFavorite(t)
	require.NoError(t, err)

	repo := repository.NewDialogFavoriteRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "dialog_favorites" WHERE user_id = $1`)).
		WithArgs(1).
		WillReturnError(errors.New("database error"))

	// Act
	favorites, total, err := repo.GetDialogFavoritesByUserID(ctx, 1, 0, 10)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, favorites)
	assert.Equal(t, int64(0), total)
	assert.Contains(t, err.Error(), "failed to count dialog favorites")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestDialogFavoriteRepository_DeleteDialogFavorite_Success tests successful deletion.
func TestDialogFavoriteRepository_DeleteDialogFavorite_Success(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDBForFavorite(t)
	require.NoError(t, err)

	repo := repository.NewDialogFavoriteRepository(db)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "dialog_favorites" WHERE snowflake_id = $1 AND user_id = $2`)).
		WithArgs(123456789, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Act
	err = repo.DeleteDialogFavorite(ctx, 123456789, 1)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestDialogFavoriteRepository_DeleteDialogFavorite_NotFound tests deletion when record not found.
func TestDialogFavoriteRepository_DeleteDialogFavorite_NotFound(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDBForFavorite(t)
	require.NoError(t, err)

	repo := repository.NewDialogFavoriteRepository(db)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "dialog_favorites" WHERE snowflake_id = $1 AND user_id = $2`)).
		WithArgs(999999, 1).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	// Act
	err = repo.DeleteDialogFavorite(ctx, 999999, 1)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dialog favorite not found or not owned by user")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestDialogFavoriteRepository_IsDialogFavoriteExists_True tests when favorite exists.
func TestDialogFavoriteRepository_IsDialogFavoriteExists_True(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDBForFavorite(t)
	require.NoError(t, err)

	repo := repository.NewDialogFavoriteRepository(db)
	ctx := context.Background()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "dialog_favorites" WHERE user_id = $1 AND dialog_session_id = $2`)).
		WithArgs(1, 100).
		WillReturnRows(countRows)

	// Act
	exists, err := repo.IsDialogFavoriteExists(ctx, 1, 100)

	// Assert
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestDialogFavoriteRepository_IsDialogFavoriteExists_False tests when favorite does not exist.
func TestDialogFavoriteRepository_IsDialogFavoriteExists_False(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDBForFavorite(t)
	require.NoError(t, err)

	repo := repository.NewDialogFavoriteRepository(db)
	ctx := context.Background()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "dialog_favorites" WHERE user_id = $1 AND dialog_session_id = $2`)).
		WithArgs(1, 999).
		WillReturnRows(countRows)

	// Act
	exists, err := repo.IsDialogFavoriteExists(ctx, 1, 999)

	// Assert
	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestDialogFavoriteRepository_GetDialogFavoriteByUserAndSession_Success tests successful retrieval.
func TestDialogFavoriteRepository_GetDialogFavoriteByUserAndSession_Success(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDBForFavorite(t)
	require.NoError(t, err)

	repo := repository.NewDialogFavoriteRepository(db)
	ctx := context.Background()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "snowflake_id", "user_id", "dialog_session_id", "title", "created_at"}).
		AddRow(1, 123456789, 1, 100, "Test Favorite", now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dialog_favorites" WHERE user_id = $1 AND dialog_session_id = $2 ORDER BY "dialog_favorites"."id" LIMIT $3`)).
		WithArgs(1, 100, 1).
		WillReturnRows(rows)

	// Act
	favorite, err := repo.GetDialogFavoriteByUserAndSession(ctx, 1, 100)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, favorite)
	assert.Equal(t, int64(123456789), favorite.SnowflakeID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestDialogFavoriteRepository_GetDialogFavoriteByUserAndSession_NotFound tests when not found (returns nil, not error).
func TestDialogFavoriteRepository_GetDialogFavoriteByUserAndSession_NotFound(t *testing.T) {
	// Arrange
	db, mock, err := setupMockDBForFavorite(t)
	require.NoError(t, err)

	repo := repository.NewDialogFavoriteRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dialog_favorites" WHERE user_id = $1 AND dialog_session_id = $2 ORDER BY "dialog_favorites"."id" LIMIT $3`)).
		WithArgs(1, 999, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "snowflake_id", "user_id", "dialog_session_id", "title", "created_at"}))

	// Act
	favorite, err := repo.GetDialogFavoriteByUserAndSession(ctx, 1, 999)

	// Assert
	assert.NoError(t, err) // Not found is not an error in this case
	assert.Nil(t, favorite)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// setupMockDBForFavorite creates a mock database connection for testing.
func setupMockDBForFavorite(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, error) {
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
