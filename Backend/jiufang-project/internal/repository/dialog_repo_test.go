// Package repository implements data access layer for dialog sessions.
// This file implements unit tests for DialogRepository.
// Author: AI Assistant
// Date: 2026-06-03
// Tested Object: DialogRepository
// Function: Dialog session data access operations

package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"jiufang/internal/model/dialog"
)

// TestDialogRepository_Create tests Create method
func TestDialogRepository_Create(t *testing.T) {
	tests := []struct {
		name        string
		session     *dialog.DialogSession
		mockDB      func(mock sqlmock.Sqlmock)
		wantErr     bool
		errContains string
	}{
		{
			name: "TC-DR-01: Create session successfully",
			session: &dialog.DialogSession{
				SnowflakeID: "123456",
				UserID:      123,
				Status:      string(dialog.StatusActive),
			},
			mockDB: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `dialog_sessions`").
					WithArgs("123456", uint(123), sqlmock.AnyArg(), string(dialog.StatusActive), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "TC-DR-01-2: Create session - database error",
			session: &dialog.DialogSession{
				SnowflakeID: "123456",
				UserID:      123,
				Status:      string(dialog.StatusActive),
			},
			mockDB: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `dialog_sessions`").
					WithArgs("123456", uint(123), sqlmock.AnyArg(), string(dialog.StatusActive), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("database error"))
				mock.ExpectRollback()
			},
			wantErr:     true,
			errContains: "failed to create dialog session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			sqlDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer sqlDB.Close()

			tt.mockDB(mock)

			gormDB, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      sqlDB,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{})
			if err != nil {
				t.Fatalf("failed to open gorm DB: %v", err)
			}

			repo := NewDialogRepository(gormDB)
			ctx := context.Background()

			// Act
			err = repo.Create(ctx, tt.session)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}

// TestDialogRepository_GetByID tests GetByID method
func TestDialogRepository_GetByID(t *testing.T) {
	tests := []struct {
		name        string
		id          uint
		mockDB      func(mock sqlmock.Sqlmock)
		wantSession *dialog.DialogSession
		wantErr     bool
		errContains string
	}{
		{
			name: "TC-DR-02: Get session by ID successfully",
			id:   1,
			mockDB: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "snowflake_id", "user_id", "query_session_id", "status", "created_at", "updated_at", "closed_at"}).
					AddRow(1, "123456", 123, nil, string(dialog.StatusActive), time.Now(), time.Now(), nil)
				mock.ExpectQuery("SELECT \\* FROM `dialog_sessions` WHERE `dialog_sessions`.`id` = \\? ORDER BY `dialog_sessions`.`id` LIMIT \\?").
					WithArgs(uint(1), 1).
					WillReturnRows(rows)
			},
			wantSession: &dialog.DialogSession{
				ID:          1,
				SnowflakeID: "123456",
				UserID:      123,
				Status:      string(dialog.StatusActive),
			},
			wantErr: false,
		},
		{
			name: "TC-DR-03: Get session by ID - not found",
			id:   999,
			mockDB: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `dialog_sessions` WHERE `dialog_sessions`.`id` = \\? ORDER BY `dialog_sessions`.`id` LIMIT \\?").
					WithArgs(uint(999), 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "snowflake_id", "user_id", "query_session_id", "status", "created_at", "updated_at", "closed_at"}))
			},
			wantSession: nil,
			wantErr:     false, // Should return nil, not error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			sqlDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer sqlDB.Close()

			tt.mockDB(mock)

			gormDB, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      sqlDB,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{})
			if err != nil {
				t.Fatalf("failed to open gorm DB: %v", err)
			}

			repo := NewDialogRepository(gormDB)
			ctx := context.Background()

			// Act
			session, err := repo.GetByID(ctx, tt.id)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, session)
			} else {
				assert.NoError(t, err)
				if tt.wantSession != nil {
					assert.NotNil(t, session)
					assert.Equal(t, tt.wantSession.ID, session.ID)
					assert.Equal(t, tt.wantSession.SnowflakeID, session.SnowflakeID)
				} else {
					assert.Nil(t, session)
				}
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}

// TestDialogRepository_GetBySnowflakeID tests GetBySnowflakeID method
func TestDialogRepository_GetBySnowflakeID(t *testing.T) {
	tests := []struct {
		name        string
		snowflakeID string
		mockDB      func(mock sqlmock.Sqlmock)
		wantSession *dialog.DialogSession
		wantErr     bool
	}{
		{
			name:        "TC-DR-04: Get session by snowflake ID successfully",
			snowflakeID: "123456",
			mockDB: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "snowflake_id", "user_id", "query_session_id", "status", "created_at", "updated_at", "closed_at"}).
					AddRow(1, "123456", 123, nil, string(dialog.StatusActive), time.Now(), time.Now(), nil)
				mock.ExpectQuery("SELECT \\* FROM `dialog_sessions` WHERE snowflake_id = \\? ORDER BY `dialog_sessions`.`id` LIMIT \\?").
					WithArgs("123456", 1).
					WillReturnRows(rows)
			},
			wantSession: &dialog.DialogSession{
				ID:          1,
				SnowflakeID: "123456",
				UserID:      123,
				Status:      string(dialog.StatusActive),
			},
			wantErr: false,
		},
		{
			name:        "TC-DR-04-2: Get session by snowflake ID - not found",
			snowflakeID: "invalid",
			mockDB: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `dialog_sessions` WHERE snowflake_id = \\? ORDER BY `dialog_sessions`.`id` LIMIT \\?").
					WithArgs("invalid", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "snowflake_id", "user_id", "query_session_id", "status", "created_at", "updated_at", "closed_at"}))
			},
			wantSession: nil,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			sqlDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer sqlDB.Close()

			tt.mockDB(mock)

			gormDB, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      sqlDB,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{})
			if err != nil {
				t.Fatalf("failed to open gorm DB: %v", err)
			}

			repo := NewDialogRepository(gormDB)
			ctx := context.Background()

			// Act
			session, err := repo.GetBySnowflakeID(ctx, tt.snowflakeID)

			// Assert
			assert.NoError(t, err)
			if tt.wantSession != nil {
				assert.NotNil(t, session)
				assert.Equal(t, tt.wantSession.SnowflakeID, session.SnowflakeID)
			} else {
				assert.Nil(t, session)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestDialogRepository_GetByUserID tests GetByUserID method
func TestDialogRepository_GetByUserID(t *testing.T) {
	tests := []struct {
		name         string
		userID       uint
		offset       int
		limit        int
		mockDB       func(mock sqlmock.Sqlmock)
		wantSessions []dialog.DialogSession
		wantTotal    int64
		wantErr      bool
	}{
		{
			name:   "TC-DR-05: Get sessions by user ID successfully",
			userID: 123,
			offset: 0,
			limit:  10,
			mockDB: func(mock sqlmock.Sqlmock) {
				// Count query
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `dialog_sessions` WHERE user_id = \\?").
					WithArgs(uint(123)).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

				// Select query
				rows := sqlmock.NewRows([]string{"id", "snowflake_id", "user_id", "query_session_id", "status", "created_at", "updated_at", "closed_at"}).
					AddRow(1, "123456", 123, nil, string(dialog.StatusActive), time.Now(), time.Now(), nil).
					AddRow(2, "789012", 123, nil, string(dialog.StatusClosed), time.Now(), time.Now(), time.Now())
				mock.ExpectQuery("SELECT \\* FROM `dialog_sessions` WHERE user_id = \\? ORDER BY created_at DESC LIMIT \\?").
					WithArgs(uint(123), 10).
					WillReturnRows(rows)
			},
			wantSessions: []dialog.DialogSession{
				{ID: 1, SnowflakeID: "123456", UserID: 123, Status: string(dialog.StatusActive)},
				{ID: 2, SnowflakeID: "789012", UserID: 123, Status: string(dialog.StatusClosed)},
			},
			wantTotal: 2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			sqlDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer sqlDB.Close()

			tt.mockDB(mock)

			gormDB, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      sqlDB,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{})
			if err != nil {
				t.Fatalf("failed to open gorm DB: %v", err)
			}

			repo := NewDialogRepository(gormDB)
			ctx := context.Background()

			// Act
			sessions, total, err := repo.GetByUserID(ctx, tt.userID, tt.offset, tt.limit)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTotal, total)
				if tt.wantSessions != nil {
					assert.Len(t, sessions, len(tt.wantSessions))
				}
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}

// TestDialogRepository_CloseSession tests CloseSession method
func TestDialogRepository_CloseSession(t *testing.T) {
	tests := []struct {
		name        string
		snowflakeID string
		mockDB      func(mock sqlmock.Sqlmock)
		wantErr     bool
	}{
		{
			name:        "TC-DR-08: Close session successfully",
			snowflakeID: "123456",
			mockDB: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `dialog_sessions`").
					WithArgs(sqlmock.AnyArg(), string(dialog.StatusClosed), sqlmock.AnyArg(), "123456").
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			sqlDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer sqlDB.Close()

			tt.mockDB(mock)

			gormDB, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      sqlDB,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{})
			if err != nil {
				t.Fatalf("failed to open gorm DB: %v", err)
			}

			repo := NewDialogRepository(gormDB)
			ctx := context.Background()

			// Act
			err = repo.CloseSession(ctx, tt.snowflakeID)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}
