// Package dialog implements the dialog management model for multi-turn conversations.
// This file defines the DialogContext value object for storing conversation context.
package dialog

import (
	"encoding/json"
	"fmt"
	"time"

	"jiufang/internal/model/agent"
)

// DialogContext represents the context of a dialog session.
// It is stored in Redis with TTL and contains conversation history and entities.
type DialogContext struct {
	SessionID       string                 `json:"session_id"`
	UserID          uint                   `json:"user_id"`
	TurnCount       int                    `json:"turn_count"`       // Current turn count
	MaxTurns        int                    `json:"max_turns"`        // Maximum turns to retain (default 5)
	QueryHistory    []QueryTurn            `json:"query_history"`    // Query history (max 5 turns)
	CurrentEntities map[string]interface{} `json:"current_entities"` // Current entity conditions
	LastIntent      *agent.Intent          `json:"last_intent"`      // Last intent
	LastResult      *agent.QueryResult     `json:"last_result"`      // Last query result
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// QueryTurn represents a single turn in the conversation.
type QueryTurn struct {
	TurnID        int                    `json:"turn_id"`
	Input         string                 `json:"input"`
	Understanding string                 `json:"understanding"`
	Entities      map[string]interface{} `json:"entities"`
	Intent        agent.IntentType       `json:"intent"`
	Result        *agent.QueryResult     `json:"result,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
}

// NewDialogContext creates a new dialog context.
func NewDialogContext(sessionID string, userID uint, maxTurns int) *DialogContext {
	return &DialogContext{
		SessionID:       sessionID,
		UserID:          userID,
		TurnCount:       0,
		MaxTurns:        maxTurns,
		QueryHistory:    []QueryTurn{},
		CurrentEntities: make(map[string]interface{}),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// AddTurn adds a new turn to the conversation history.
// It maintains the maximum turn limit by removing old turns if necessary.
func (c *DialogContext) AddTurn(turn *QueryTurn) {
	c.TurnCount++
	turn.TurnID = c.TurnCount
	turn.Timestamp = time.Now()

	// Add to history
	c.QueryHistory = append(c.QueryHistory, *turn)

	// Limit history size
	if len(c.QueryHistory) > c.MaxTurns {
		c.QueryHistory = c.QueryHistory[len(c.QueryHistory)-c.MaxTurns:]
	}

	// Update current entities
	if turn.Entities != nil {
		for key, value := range turn.Entities {
			c.CurrentEntities[key] = value
		}
	}

	// Update last intent and result
	if turn.Intent != "" {
		c.LastIntent = &agent.Intent{Type: turn.Intent}
	}
	if turn.Result != nil {
		c.LastResult = turn.Result
	}

	c.UpdatedAt = time.Now()
}

// GetLastTurn returns the last turn in the conversation.
func (c *DialogContext) GetLastTurn() *QueryTurn {
	if len(c.QueryHistory) == 0 {
		return nil
	}
	return &c.QueryHistory[len(c.QueryHistory)-1]
}

// GetPreviousTurns returns the previous turns (excluding the last turn).
func (c *DialogContext) GetPreviousTurns() []QueryTurn {
	if len(c.QueryHistory) <= 1 {
		return []QueryTurn{}
	}
	return c.QueryHistory[:len(c.QueryHistory)-1]
}

// Clear clears the dialog context.
func (c *DialogContext) Clear() {
	c.TurnCount = 0
	c.QueryHistory = []QueryTurn{}
	c.CurrentEntities = make(map[string]interface{})
	c.LastIntent = nil
	c.LastResult = nil
	c.UpdatedAt = time.Now()
}

// MergeEntities merges new entities with existing entities.
// It handles conflicts by preferring new entities.
func (c *DialogContext) MergeEntities(newEntities map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// Copy existing entities
	for key, value := range c.CurrentEntities {
		merged[key] = value
	}

	// Merge new entities (overwrite conflicts)
	for key, value := range newEntities {
		merged[key] = value
	}

	return merged
}

// ToJSON serializes the dialog context to JSON.
func (c *DialogContext) ToJSON() (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("failed to marshal dialog context: %w", err)
	}
	return string(data), nil
}

// FromJSON deserializes the dialog context from JSON.
func FromJSON(data string) (*DialogContext, error) {
	var context DialogContext
	if err := json.Unmarshal([]byte(data), &context); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dialog context: %w", err)
	}
	return &context, nil
}

// ContextKey returns the Redis key for the dialog context.
func ContextKey(sessionID string) string {
	return fmt.Sprintf("dialog:ctx:%s", sessionID)
}
