package dialog

import (
	"context"
)

// DialogRepositoryInterface defines the interface for dialog repository operations.
// This interface is defined in the dialog package to avoid circular imports.
type DialogRepositoryInterface interface {
	// DialogSession operations
	Create(ctx context.Context, session *DialogSession) error
	CreateSession(ctx context.Context, session *DialogSession) error
	GetSessionByID(ctx context.Context, id uint) (*DialogSession, error)
	GetSessionBySnowflakeID(ctx context.Context, snowflakeID string) (*DialogSession, error)
	GetBySnowflakeID(ctx context.Context, snowflakeID string) (*DialogSession, error)
	UpdateSession(ctx context.Context, session *DialogSession) error
	DeleteSession(ctx context.Context, id uint) error
	CloseSession(ctx context.Context, sessionID string) error
	ListSessions(ctx context.Context, userID uint, limit, offset int) ([]*DialogSession, error)
	GetActiveSessionsByUserID(ctx context.Context, userID uint) ([]DialogSession, error)
	GetByUserID(ctx context.Context, userID uint, offset, limit int) ([]DialogSession, int64, error)

	// DialogContext operations
	CreateContext(ctx context.Context, context *DialogContext) error
	GetContextByID(ctx context.Context, id uint) (*DialogContext, error)
	GetContextBySessionID(ctx context.Context, sessionID uint) (*DialogContext, error)
	UpdateContext(ctx context.Context, context *DialogContext) error
	DeleteContext(ctx context.Context, id uint) error

	// DialogFavorite operations
	CreateFavorite(ctx context.Context, favorite *DialogFavorite) error
	GetFavoriteByID(ctx context.Context, id uint) (*DialogFavorite, error)
	GetFavoriteBySnowflakeID(ctx context.Context, snowflakeID int64) (*DialogFavorite, error)
	UpdateFavorite(ctx context.Context, favorite *DialogFavorite) error
	DeleteFavorite(ctx context.Context, id uint) error
	ListFavorites(ctx context.Context, userID uint, limit, offset int) ([]*DialogFavorite, error)
}
