// Package agent defines the data models for AI Agent module.
// This module handles natural language understanding and SQL generation.
package agent

// IntentType represents the type of query intent.
// Based on PRD BR-001, the system identifies query types: statistics, detail, trend, comparison.
type IntentType string

const (
	// IntentTypeStatistics represents statistical query (e.g., "total purchase amount last month")
	IntentTypeStatistics IntentType = "statistics"

	// IntentTypeDetail represents detail query (e.g., "show all purchase orders from supplier A")
	IntentTypeDetail IntentType = "detail"

	// IntentTypeTrend represents trend query (e.g., "purchase amount trend over last 6 months")
	IntentTypeTrend IntentType = "trend"

	// IntentTypeComparison represents comparison query (e.g., "compare sales between Q1 and Q2")
	IntentTypeComparison IntentType = "comparison"

	// IntentTypeUnknown represents unrecognized intent
	IntentTypeUnknown IntentType = "unknown"
)

// Intent represents the parsed intent from user's natural language input.
// It includes the intent type, confidence score, and description.
type Intent struct {
	// Type is the classified intent type
	Type IntentType `json:"type"`

	// Confidence is the confidence score of intent classification (0.0 ~ 1.0)
	// Based on PRD BR-003, confidence below threshold triggers clarification mechanism
	Confidence float64 `json:"confidence"`

	// Description is the human-readable description of the intent
	Description string `json:"description"`

	// RawInput is the original user input text
	RawInput string `json:"raw_input"`
}

// IsConfident checks if the intent confidence is above the threshold.
// Threshold is 0.7 based on PRD requirements.
func (i *Intent) IsConfident() bool {
	return i.Confidence >= 0.7
}

// NeedsClarification checks if the intent needs clarification.
// Returns true if confidence is low or intent type is unknown.
func (i *Intent) NeedsClarification() bool {
	return !i.IsConfident() || i.Type == IntentTypeUnknown
}

// IntentParserResult represents the result of intent parsing.
// Contains the parsed intent and any detected entities.
type IntentParserResult struct {
	Intent   Intent   `json:"intent"`
	Entities []Entity `json:"entities"`
}
