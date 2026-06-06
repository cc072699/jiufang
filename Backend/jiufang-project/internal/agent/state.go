// Package agent implements the AI Agent for semantic understanding and SQL generation.
package agent

import (
	"encoding/json"
	"time"

	"jiufang/internal/model/agent"
)

// StateManager manages the state of query execution.
type StateManager struct {
	states map[string]*QueryState
}

// NewStateManager creates a new state manager.
func NewStateManager() *StateManager {
	return &StateManager{
		states: make(map[string]*QueryState),
	}
}

// GetState returns the state for a session.
func (m *StateManager) GetState(sessionID string) *QueryState {
	return m.states[sessionID]
}

// SetState sets the state for a session.
func (m *StateManager) SetState(sessionID string, state *QueryState) {
	m.states[sessionID] = state
}

// ClearState clears the state for a session.
func (m *StateManager) ClearState(sessionID string) {
	delete(m.states, sessionID)
}

// ConversationState represents the state of a conversation.
type ConversationState struct {
	SessionID       string              `json:"session_id"`
	UserID          uint                `json:"user_id"`
	QueryHistory    []QueryHistoryItem  `json:"query_history"`
	CurrentEntities []agent.Entity      `json:"current_entities"`
	LastIntent      *agent.Intent       `json:"last_intent"`
	LastResult      *agent.QueryResult  `json:"last_result"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}

// QueryHistoryItem represents a query in the conversation history.
type QueryHistoryItem struct {
	Timestamp time.Time          `json:"timestamp"`
	Input     string             `json:"input"`
	Intent    agent.IntentType   `json:"intent"`
	Entities  []agent.Entity     `json:"entities"`
	Result    *agent.QueryResult `json:"result,omitempty"`
	Success   bool               `json:"success"`
}

// ConversationManager manages conversation states.
type ConversationManager struct {
	conversations map[string]*ConversationState
	maxHistory    int
}

// NewConversationManager creates a new conversation manager.
func NewConversationManager(maxHistory int) *ConversationManager {
	return &ConversationManager{
		conversations: make(map[string]*ConversationState),
		maxHistory:    maxHistory,
	}
}

// GetConversation returns the conversation state for a session.
func (m *ConversationManager) GetConversation(sessionID string) *ConversationState {
	return m.conversations[sessionID]
}

// CreateConversation creates a new conversation state.
func (m *ConversationManager) CreateConversation(sessionID string, userID uint) *ConversationState {
	state := &ConversationState{
		SessionID:    sessionID,
		UserID:       userID,
		QueryHistory: []QueryHistoryItem{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	m.conversations[sessionID] = state
	return state
}

// UpdateConversation updates the conversation state.
func (m *ConversationManager) UpdateConversation(sessionID string, query *QueryHistoryItem) {
	state := m.conversations[sessionID]
	if state == nil {
		return
	}

	// Add to history
	state.QueryHistory = append(state.QueryHistory, *query)

	// Limit history size
	if len(state.QueryHistory) > m.maxHistory {
		state.QueryHistory = state.QueryHistory[len(state.QueryHistory)-m.maxHistory:]
	}

	// Update current entities and intent
	if len(query.Entities) > 0 {
		state.CurrentEntities = query.Entities
	}
	if query.Intent != "" {
		state.LastIntent = &agent.Intent{Type: query.Intent}
	}
	if query.Result != nil {
		state.LastResult = query.Result
	}

	state.UpdatedAt = time.Now()
}

// ClearConversation clears the conversation state.
func (m *ConversationManager) ClearConversation(sessionID string) {
	delete(m.conversations, sessionID)
}

// GetPreviousEntities returns entities from previous queries.
func (m *ConversationManager) GetPreviousEntities(sessionID string) []agent.Entity {
	state := m.conversations[sessionID]
	if state == nil {
		return nil
	}
	return state.CurrentEntities
}

// GetLastIntent returns the last intent for a session.
func (m *ConversationManager) GetLastIntent(sessionID string) *agent.Intent {
	state := m.conversations[sessionID]
	if state == nil {
		return nil
	}
	return state.LastIntent
}

// SerializeState serializes a query state to JSON.
func SerializeState(state *QueryState) (string, error) {
	data, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DeserializeState deserializes a query state from JSON.
func DeserializeState(data string) (*QueryState, error) {
	var state QueryState
	if err := json.Unmarshal([]byte(data), &state); err != nil {
		return nil, err
	}
	return &state, nil
}

// SerializeConversation serializes a conversation state to JSON.
func SerializeConversation(state *ConversationState) (string, error) {
	data, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DeserializeConversation deserializes a conversation state from JSON.
func DeserializeConversation(data string) (*ConversationState, error) {
	var state ConversationState
	if err := json.Unmarshal([]byte(data), &state); err != nil {
		return nil, err
	}
	return &state, nil
}

// StateSnapshot represents a snapshot of the query state at a point in time.
type StateSnapshot struct {
	SessionID   string    `json:"session_id"`
	Step        string    `json:"step"`
	Timestamp   time.Time `json:"timestamp"`
	StateJSON   string    `json:"state_json"`
	IsComplete  bool      `json:"is_complete"`
	HasError    bool      `json:"has_error"`
	ErrorMsg    string    `json:"error_msg,omitempty"`
}

// CreateSnapshot creates a snapshot of the current state.
func CreateSnapshot(state *QueryState) *StateSnapshot {
	stateJSON, _ := SerializeState(state)
	return &StateSnapshot{
		SessionID:  state.SessionID,
		Step:       state.CurrentStep,
		Timestamp:  time.Now(),
		StateJSON:  stateJSON,
		IsComplete: state.IsComplete,
		HasError:   state.Error != "",
		ErrorMsg:   state.Error,
	}
}

// StateRecorder records state snapshots for debugging and analysis.
type StateRecorder struct {
	snapshots map[string][]StateSnapshot
	maxRecords int
}

// NewStateRecorder creates a new state recorder.
func NewStateRecorder(maxRecords int) *StateRecorder {
	return &StateRecorder{
		snapshots: make(map[string][]StateSnapshot),
		maxRecords: maxRecords,
	}
}

// Record records a state snapshot.
func (r *StateRecorder) Record(state *QueryState) {
	snapshot := CreateSnapshot(state)
	sessionID := state.SessionID

	r.snapshots[sessionID] = append(r.snapshots[sessionID], *snapshot)

	// Limit records per session
	if len(r.snapshots[sessionID]) > r.maxRecords {
		r.snapshots[sessionID] = r.snapshots[sessionID][len(r.snapshots[sessionID])-r.maxRecords:]
	}
}

// GetSnapshots returns snapshots for a session.
func (r *StateRecorder) GetSnapshots(sessionID string) []StateSnapshot {
	return r.snapshots[sessionID]
}

// ClearSnapshots clears snapshots for a session.
func (r *StateRecorder) ClearSnapshots(sessionID string) {
	delete(r.snapshots, sessionID)
}

// GetAllSnapshots returns all snapshots.
func (r *StateRecorder) GetAllSnapshots() map[string][]StateSnapshot {
	return r.snapshots
}