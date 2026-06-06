// Package dialog implements the dialog management model for multi-turn conversations.
// This file implements the DialogManager which coordinates session and context management.
package dialog

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"jiufang/internal/model/agent"
	"jiufang/internal/pkg/id"
)

// DialogManager coordinates dialog session and context management.
// It integrates with Redis for context storage and PostgreSQL for session metadata.
type DialogManager struct {
	repo           DialogRepositoryInterface
	cache          DialogCacheInterface
	anaphora       *AnaphoraResolver
	merger         *ConditionMerger
	snowflakeGen   id.SnowflakeGeneratorInterface
	logger         *zap.Logger
	maxTurns       int
	enableAnaphora bool
	enableMerge    bool
}

// NewDialogManager creates a new dialog manager.
func NewDialogManager(
	repo DialogRepositoryInterface,
	cache DialogCacheInterface,
	snowflakeGen id.SnowflakeGeneratorInterface,
	logger *zap.Logger,
	maxTurns int,
	enableAnaphora bool,
	enableMerge bool,
) *DialogManager {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &DialogManager{
		repo:           repo,
		cache:          cache,
		anaphora:       NewAnaphoraResolver(logger),
		merger:         NewConditionMerger(logger),
		snowflakeGen:   snowflakeGen,
		logger:         logger,
		maxTurns:       maxTurns,
		enableAnaphora: enableAnaphora,
		enableMerge:    enableMerge,
	}
}

// CreateSession creates a new dialog session.
func (m *DialogManager) CreateSession(ctx context.Context, userID uint) (*DialogSession, error) {
	// Generate snowflake ID
	snowflakeID := m.snowflakeGen.Generate()

	session := &DialogSession{
		SnowflakeID: fmt.Sprintf("%d", snowflakeID),
		UserID:      userID,
		Status:      string(StatusActive),
	}

	// Save to database
	if err := m.repo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create dialog session: %w", err)
	}

	// Create initial context in Redis
	context := NewDialogContext(session.SnowflakeID, userID, m.maxTurns)
	if err := m.cache.SaveContext(ctx, context); err != nil {
		m.logger.Warn("Failed to save initial context to Redis, will fallback to memory",
			zap.String("session_id", session.SnowflakeID),
			zap.Error(err),
		)
	}

	m.logger.Info("Dialog session created",
		zap.String("session_id", session.SnowflakeID),
		zap.Uint("user_id", userID),
	)

	return session, nil
}

// GetSession retrieves a dialog session by snowflake ID.
func (m *DialogManager) GetSession(ctx context.Context, sessionID string) (*DialogSession, error) {
	session, err := m.repo.GetBySnowflakeID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dialog session: %w", err)
	}

	if session == nil {
		m.logger.Warn("Dialog session not found",
			zap.String("session_id", sessionID),
		)
		return nil, nil
	}

	return session, nil
}

// CloseSession closes a dialog session.
func (m *DialogManager) CloseSession(ctx context.Context, sessionID string) error {
	// Close in database
	if err := m.repo.CloseSession(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to close dialog session in database: %w", err)
	}

	// Clear context from Redis
	if err := m.cache.ClearContext(ctx, sessionID); err != nil {
		m.logger.Warn("Failed to clear dialog context from Redis",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
	}

	m.logger.Info("Dialog session closed",
		zap.String("session_id", sessionID),
	)

	return nil
}

// LoadContext loads dialog context from Redis.
func (m *DialogManager) LoadContext(ctx context.Context, sessionID string) (*DialogContext, error) {
	context, err := m.cache.LoadContext(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load dialog context: %w", err)
	}

	// If context not found, create new one
	if context == nil {
		session, err := m.GetSession(ctx, sessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to get session for context creation: %w", err)
		}

		if session == nil {
			return nil, fmt.Errorf("session not found: %s", sessionID)
		}

		context = NewDialogContext(sessionID, session.UserID, m.maxTurns)
		m.logger.Info("Created new dialog context",
			zap.String("session_id", sessionID),
		)
	}

	return context, nil
}

// SaveContext saves dialog context to Redis.
func (m *DialogManager) SaveContext(ctx context.Context, context *DialogContext) error {
	if err := m.cache.SaveContext(ctx, context); err != nil {
		return fmt.Errorf("failed to save dialog context: %w", err)
	}
	return nil
}

// ClearContext clears dialog context from Redis.
func (m *DialogManager) ClearContext(ctx context.Context, sessionID string) error {
	if err := m.cache.ClearContext(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to clear dialog context: %w", err)
	}
	return nil
}

// AddTurn adds a new turn to the conversation.
func (m *DialogManager) AddTurn(ctx context.Context, sessionID string, turn *QueryTurn) error {
	// Load context
	context, err := m.LoadContext(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to load context for adding turn: %w", err)
	}

	// Add turn to context
	context.AddTurn(turn)

	// Save context
	if err := m.SaveContext(ctx, context); err != nil {
		return fmt.Errorf("failed to save context after adding turn: %w", err)
	}

	m.logger.Info("Turn added to dialog",
		zap.String("session_id", sessionID),
		zap.Int("turn_id", turn.TurnID),
		zap.String("input", turn.Input),
	)

	return nil
}

// GetHistory returns the conversation history for a session.
func (m *DialogManager) GetHistory(ctx context.Context, sessionID string, limit int) ([]QueryTurn, error) {
	context, err := m.LoadContext(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load context for history: %w", err)
	}

	history := context.QueryHistory
	if limit > 0 && len(history) > limit {
		history = history[len(history)-limit:]
	}

	return history, nil
}

// ResolveAnaphora resolves anaphora (指代词) in user input.
func (m *DialogManager) ResolveAnaphora(ctx context.Context, sessionID string, input string) (string, error) {
	if !m.enableAnaphora {
		return input, nil
	}

	// Load context
	context, err := m.LoadContext(ctx, sessionID)
	if err != nil {
		m.logger.Warn("Failed to load context for anaphora resolution, using original input",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		return input, nil
	}

	// Resolve anaphora
	resolved, err := m.anaphora.Resolve(input, context)
	if err != nil {
		m.logger.Warn("Anaphora resolution failed, using original input",
			zap.String("session_id", sessionID),
			zap.String("input", input),
			zap.Error(err),
		)
		return input, nil
	}

	m.logger.Info("Anaphora resolved",
		zap.String("session_id", sessionID),
		zap.String("original", input),
		zap.String("resolved", resolved),
	)

	return resolved, nil
}

// MergeConditions merges new conditions with existing context.
func (m *DialogManager) MergeConditions(ctx context.Context, sessionID string, newEntities map[string]interface{}) (map[string]interface{}, error) {
	if !m.enableMerge {
		return newEntities, nil
	}

	// Load context
	context, err := m.LoadContext(ctx, sessionID)
	if err != nil {
		m.logger.Warn("Failed to load context for condition merging, using new entities only",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		return newEntities, nil
	}

	// Merge conditions
	merged := m.merger.Merge(context.CurrentEntities, newEntities)

	m.logger.Info("Conditions merged",
		zap.String("session_id", sessionID),
		zap.Any("existing", context.CurrentEntities),
		zap.Any("new", newEntities),
		zap.Any("merged", merged),
	)

	return merged, nil
}

// ProcessInput processes user input with context management.
// It resolves anaphora and merges conditions before creating a query turn.
func (m *DialogManager) ProcessInput(ctx context.Context, sessionID string, input string, understanding string, entities map[string]interface{}, intent agent.IntentType, result *agent.QueryResult) (*QueryTurn, error) {
	// Resolve anaphora
	resolvedInput, err := m.ResolveAnaphora(ctx, sessionID, input)
	if err != nil {
		m.logger.Warn("Anaphora resolution failed, continuing with original input",
			zap.Error(err),
		)
		resolvedInput = input
	}

	// Merge conditions
	mergedEntities, err := m.MergeConditions(ctx, sessionID, entities)
	if err != nil {
		m.logger.Warn("Condition merging failed, continuing with new entities only",
			zap.Error(err),
		)
		mergedEntities = entities
	}

	// Create query turn
	turn := &QueryTurn{
		Input:         resolvedInput,
		Understanding: understanding,
		Entities:      mergedEntities,
		Intent:        intent,
		Result:        result,
	}

	// Add turn to context
	if err := m.AddTurn(ctx, sessionID, turn); err != nil {
		return nil, fmt.Errorf("failed to add turn: %w", err)
	}

	return turn, nil
}

// GetActiveSessions returns active sessions for a user.
func (m *DialogManager) GetActiveSessions(ctx context.Context, userID uint) ([]DialogSession, error) {
	sessions, err := m.repo.GetActiveSessionsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}
	return sessions, nil
}

// GetUserSessions returns all sessions for a user with pagination.
func (m *DialogManager) GetUserSessions(ctx context.Context, userID uint, offset, limit int) ([]DialogSession, int64, error) {
	sessions, total, err := m.repo.GetByUserID(ctx, userID, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user sessions: %w", err)
	}
	return sessions, total, nil
}
