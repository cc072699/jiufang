// Package cache implements dialog context cache using Redis.
// This file implements unit tests for DialogCache.
// Author: AI Assistant
// Date: 2026-06-03
// Tested Object: DialogCache
// Function: Dialog context Redis cache operations

package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"jiufang/internal/mocks"
	"jiufang/internal/model/dialog"
)

// TestDialogCache_LoadContext tests LoadContext method
func TestDialogCache_LoadContext(t *testing.T) {
	tests := []struct {
		name         string
		sessionID    string
		mockRedis    func(ctrl *gomock.Controller) *mocks.MockRedisClient
		wantContext  *dialog.DialogContext
		wantErr      bool
		errContains  string
	}{
		{
			name:      "TC-DC-01: Load context successfully",
			sessionID: "123456",
			mockRedis: func(ctrl *gomock.Controller) *mocks.MockRedisClient {
				mock := mocks.NewMockRedisClient(ctrl)
				contextJSON := `{"session_id":"123456","user_id":123,"turn_count":2,"max_turns":5}`
				mock.EXPECT().Get(gomock.Any(), "dialog:ctx:123456").Return(contextJSON, nil)
				return mock
			},
			wantContext: &dialog.DialogContext{
				SessionID:  "123456",
				UserID:     123,
				TurnCount:  2,
				MaxTurns:   5,
			},
			wantErr: false,
		},
		{
			name:      "TC-DC-02: Load context - not found",
			sessionID: "123456",
			mockRedis: func(ctrl *gomock.Controller) *mocks.MockRedisClient {
				mock := mocks.NewMockRedisClient(ctrl)
				mock.EXPECT().Get(gomock.Any(), "dialog:ctx:123456").Return("", nil)
				return mock
			},
			wantContext: nil,
			wantErr:     false,
		},
		{
			name:      "TC-DC-02-2: Load context - Redis error",
			sessionID: "123456",
			mockRedis: func(ctrl *gomock.Controller) *mocks.MockRedisClient {
				mock := mocks.NewMockRedisClient(ctrl)
				mock.EXPECT().Get(gomock.Any(), "dialog:ctx:123456").Return("", errors.New("redis error"))
				return mock
			},
			wantContext: nil,
			wantErr:     true,
			errContains: "failed to load dialog context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRedis := tt.mockRedis(ctrl)
			cache := NewDialogCache(mockRedis, 30, zap.NewNop())
			ctx := context.Background()

			// Act
			context, err := cache.LoadContext(ctx, tt.sessionID)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, context)
			} else {
				assert.NoError(t, err)
				if tt.wantContext != nil {
					assert.NotNil(t, context)
					assert.Equal(t, tt.wantContext.SessionID, context.SessionID)
					assert.Equal(t, tt.wantContext.UserID, context.UserID)
					assert.Equal(t, tt.wantContext.TurnCount, context.TurnCount)
				} else {
					assert.Nil(t, context)
				}
			}
		})
	}
}

// TestDialogCache_SaveContext tests SaveContext method
func TestDialogCache_SaveContext(t *testing.T) {
	tests := []struct {
		name        string
		context     *dialog.DialogContext
		mockRedis   func(ctrl *gomock.Controller) *mocks.MockRedisClient
		wantErr     bool
		errContains string
	}{
		{
			name: "TC-DC-03: Save context successfully",
			context: &dialog.DialogContext{
				SessionID:  "123456",
				UserID:     123,
				TurnCount:  2,
				MaxTurns:   5,
			},
			mockRedis: func(ctrl *gomock.Controller) *mocks.MockRedisClient {
				mock := mocks.NewMockRedisClient(ctrl)
				mock.EXPECT().Set(gomock.Any(), "dialog:ctx:123456", gomock.Any(), time.Duration(30)*time.Minute).Return(nil)
				return mock
			},
			wantErr: false,
		},
		{
			name: "TC-DC-03-2: Save context - Redis error",
			context: &dialog.DialogContext{
				SessionID:  "123456",
				UserID:     123,
				TurnCount:  2,
				MaxTurns:   5,
			},
			mockRedis: func(ctrl *gomock.Controller) *mocks.MockRedisClient {
				mock := mocks.NewMockRedisClient(ctrl)
				mock.EXPECT().Set(gomock.Any(), "dialog:ctx:123456", gomock.Any(), time.Duration(30)*time.Minute).Return(errors.New("redis error"))
				return mock
			},
			wantErr:     true,
			errContains: "failed to save dialog context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRedis := tt.mockRedis(ctrl)
			cache := NewDialogCache(mockRedis, 30, zap.NewNop())
			ctx := context.Background()

			// Act
			err := cache.SaveContext(ctx, tt.context)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDialogCache_ClearContext tests ClearContext method
func TestDialogCache_ClearContext(t *testing.T) {
	tests := []struct {
		name        string
		sessionID   string
		mockRedis   func(ctrl *gomock.Controller) *mocks.MockRedisClient
		wantErr     bool
		errContains string
	}{
		{
			name:      "TC-DC-04: Clear context successfully",
			sessionID: "123456",
			mockRedis: func(ctrl *gomock.Controller) *mocks.MockRedisClient {
				mock := mocks.NewMockRedisClient(ctrl)
				mock.EXPECT().Delete(gomock.Any(), "dialog:ctx:123456").Return(nil)
				return mock
			},
			wantErr: false,
		},
		{
			name:      "TC-DC-04-2: Clear context - Redis error",
			sessionID: "123456",
			mockRedis: func(ctrl *gomock.Controller) *mocks.MockRedisClient {
				mock := mocks.NewMockRedisClient(ctrl)
				mock.EXPECT().Delete(gomock.Any(), "dialog:ctx:123456").Return(errors.New("redis error"))
				return mock
			},
			wantErr:     true,
			errContains: "failed to clear dialog context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRedis := tt.mockRedis(ctrl)
			cache := NewDialogCache(mockRedis, 30, zap.NewNop())
			ctx := context.Background()

			// Act
			err := cache.ClearContext(ctx, tt.sessionID)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDialogCache_NewDialogCache tests NewDialogCache constructor
func TestDialogCache_NewDialogCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRedis := mocks.NewMockRedisClient(ctrl)

	// Test with valid parameters
	cache := NewDialogCache(mockRedis, 30, zap.NewNop())
	assert.NotNil(t, cache)
	assert.Equal(t, time.Duration(30)*time.Minute, cache.ttl)

	// Test with nil logger (should use NopLogger)
	cache2 := NewDialogCache(mockRedis, 30, nil)
	assert.NotNil(t, cache2)
}