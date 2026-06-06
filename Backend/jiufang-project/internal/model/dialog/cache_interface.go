package dialog

import (
	"context"
)

// DialogCacheInterface defines the interface for dialog context cache operations.
// This interface is defined in the dialog package to avoid circular imports.
type DialogCacheInterface interface {
	LoadContext(ctx context.Context, sessionID string) (*DialogContext, error)
	SaveContext(ctx context.Context, context *DialogContext) error
	ClearContext(ctx context.Context, sessionID string) error
}
