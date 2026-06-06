// Package agent implements the AI Agent for semantic understanding and SQL generation.
package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"

	"jiufang/internal/model/agent"
)

// QueryState represents the state of a query execution.
type QueryState struct {
	// Input is the original user input
	Input string `json:"input"`

	// SessionID is the conversation session ID
	SessionID string `json:"session_id"`

	// UserID is the user ID
	UserID uint `json:"user_id"`

	// Intent is the parsed intent
	Intent *agent.Intent `json:"intent"`

	// Entities are the extracted entities
	Entities []agent.Entity `json:"entities"`

	// SQL is the generated SQL
	SQL string `json:"sql"`

	// ValidationResult is the SQL validation result
	ValidationResult *SQLValidationResult `json:"validation_result"`

	// Result is the query result
	Result *agent.QueryResult `json:"result"`

	// Understanding is the understanding summary
	Understanding string `json:"understanding"`

	// Clarification is the clarification request (if needed)
	Clarification *agent.ClarificationRequest `json:"clarification,omitempty"`

	// Error is the error message (if failed)
	Error string `json:"error,omitempty"`

	// CurrentStep is the current execution step
	CurrentStep string `json:"current_step"`

	// IsComplete indicates if the query is complete
	IsComplete bool `json:"is_complete"`

	// NeedsClarification indicates if clarification is needed
	NeedsClarification bool `json:"needs_clarification"`

	// IsFallback indicates if fallback mode was used
	IsFallback bool `json:"is_fallback"`

	// RetryCount is the number of retries
	RetryCount int `json:"retry_count"`
}

// QueryGraph represents the execution graph for query processing.
type QueryGraph struct {
	agent    *QueryAgent
	logger   *zap.Logger
	maxRetry int
}

// NewQueryGraph creates a new query graph.
func NewQueryGraph(agent *QueryAgent, logger *zap.Logger) *QueryGraph {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &QueryGraph{
		agent:    agent,
		logger:   logger,
		maxRetry: 3,
	}
}

// Execute executes the query graph.
func (g *QueryGraph) Execute(ctx context.Context, input string, sessionID string, userID uint) (*QueryState, error) {
	// Initialize state
	state := &QueryState{
		Input:       input,
		SessionID:   sessionID,
		UserID:      userID,
		CurrentStep: "init",
		IsComplete:  false,
	}

	// Execute steps
	steps := []struct {
		name string
		fn   func(context.Context, *QueryState) error
	}{
		{"parse_intent", g.parseIntent},
		{"extract_entities", g.extractEntities},
		{"generate_sql", g.generateSQL},
		{"validate_sql", g.validateSQL},
		{"execute_sql", g.executeSQL},
		{"format_result", g.formatResult},
	}

	for _, step := range steps {
		state.CurrentStep = step.name
		g.logger.Info("Executing step", zap.String("step", step.name))

		if err := step.fn(ctx, state); err != nil {
			g.logger.Error("Step failed", zap.String("step", step.name), zap.Error(err))
			state.Error = err.Error()
			return state, err
		}

		// Check if clarification is needed
		if state.NeedsClarification {
			g.logger.Warn("Clarification needed", zap.String("step", step.name))
			return state, nil
		}
	}

	state.IsComplete = true
	state.CurrentStep = "complete"
	return state, nil
}

// parseIntent parses the intent from input.
func (g *QueryGraph) parseIntent(ctx context.Context, state *QueryState) error {
	intent, err := g.agent.intentParser.Parse(ctx, state.Input)
	if err != nil {
		return err
	}

	state.Intent = intent

	// Check if clarification is needed
	if intent.NeedsClarification() {
		state.NeedsClarification = true
		state.Clarification = agent.NewIntentClarification(state.Input, []agent.ClarificationOption{
			{ID: "statistics", Label: "统计查询", Value: "statistics"},
			{ID: "detail", Label: "明细查询", Value: "detail"},
			{ID: "trend", Label: "趋势查询", Value: "trend"},
			{ID: "comparison", Label: "对比查询", Value: "comparison"},
		})
	}

	return nil
}

// extractEntities extracts entities from input.
func (g *QueryGraph) extractEntities(ctx context.Context, state *QueryState) error {
	if state.NeedsClarification {
		return nil // Skip if clarification is needed
	}

	entities, err := g.agent.entityExtractor.Extract(ctx, state.Input)
	if err != nil {
		return err
	}

	state.Entities = entities
	return nil
}

// generateSQL generates SQL from intent and entities.
func (g *QueryGraph) generateSQL(ctx context.Context, state *QueryState) error {
	if state.NeedsClarification {
		return nil
	}

	sql, err := g.agent.sqlGenerator.Generate(ctx, state.Intent, state.Entities)
	if err != nil {
		return err
	}

	state.SQL = sql
	return nil
}

// validateSQL validates the generated SQL.
func (g *QueryGraph) validateSQL(ctx context.Context, state *QueryState) error {
	if state.NeedsClarification {
		return nil
	}

	validationResult := g.agent.sqlValidator.ValidateDetailed(state.SQL)
	state.ValidationResult = validationResult

	if !validationResult.IsValid {
		return fmt.Errorf("SQL validation failed: %v", validationResult.Errors)
	}

	// Use safe SQL with LIMIT
	if validationResult.SafeSQL != "" {
		state.SQL = validationResult.SafeSQL
	}

	return nil
}

// executeSQL executes the SQL query.
func (g *QueryGraph) executeSQL(ctx context.Context, state *QueryState) error {
	if state.NeedsClarification {
		return nil
	}

	result, err := g.agent.sqlExecutor.Execute(ctx, state.SQL)
	if err != nil {
		return err
	}

	state.Result = result
	return nil
}

// formatResult formats the query result.
func (g *QueryGraph) formatResult(ctx context.Context, state *QueryState) error {
	if state.NeedsClarification {
		return nil
	}

	if state.Result == nil {
		return nil
	}

	// Generate understanding
	state.Understanding = g.agent.generateUnderstanding(state.Intent, state.Entities)
	state.Result.Understanding = state.Understanding

	// Determine visualization type
	state.Result.VisualizationType = g.agent.resultFormatter.DetermineVisualizationType(state.Result, state.Intent)

	return nil
}

// ExecuteWithRetry executes the graph with retry on failure.
func (g *QueryGraph) ExecuteWithRetry(ctx context.Context, input string, sessionID string, userID uint) (*QueryState, error) {
	var lastErr error

	for i := 0; i < g.maxRetry; i++ {
		state, err := g.Execute(ctx, input, sessionID, userID)
		if err == nil {
			state.RetryCount = i
			return state, nil
		}

		lastErr = err
		g.logger.Warn("Retry execution", zap.Int("attempt", i+1), zap.Error(err))

		// Check if error is retryable
		if !isRetryableError(err) {
			break
		}
	}

	return nil, lastErr
}

// isRetryableError checks if an error is retryable.
func isRetryableError(err error) bool {
	// LLM timeout/connection errors are retryable
	// Validation errors are not retryable
	errStr := err.Error()
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "network")
}

// Stream executes the graph with streaming response.
func (g *QueryGraph) Stream(ctx context.Context, input string, sessionID string, userID uint) (*schema.StreamReader[*QueryState], error) {
	msgChan := make(chan *QueryState, 10)

	go func() {
		defer close(msgChan)

		// Initialize state
		state := &QueryState{
			Input:       input,
			SessionID:   sessionID,
			UserID:      userID,
			CurrentStep: "init",
		}

		// Send initial state
		msgChan <- state

		// Execute steps and send updates
		steps := []struct {
			name string
			fn   func(context.Context, *QueryState) error
		}{
			{"parse_intent", g.parseIntent},
			{"extract_entities", g.extractEntities},
			{"generate_sql", g.generateSQL},
			{"validate_sql", g.validateSQL},
			{"execute_sql", g.executeSQL},
			{"format_result", g.formatResult},
		}

		for _, step := range steps {
			state.CurrentStep = step.name
			msgChan <- state

			if err := step.fn(ctx, state); err != nil {
				state.Error = err.Error()
				msgChan <- state
				return
			}

			msgChan <- state
		}

		state.IsComplete = true
		state.CurrentStep = "complete"
		msgChan <- state
	}()

	// Convert channel to StreamReader
	// Note: StreamReader from channel requires custom implementation
	// For now, we collect all results and return as array-based stream
	results := make([]*QueryState, 0)
	for result := range msgChan {
		results = append(results, result)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no results generated")
	}

	reader := schema.StreamReaderFromArray(results)
	return reader, nil
}
