// Package agent defines the agent domain models.
package agent

// ClarificationRequest represents a request for clarification from the user.
type ClarificationRequest struct {
	Type       string   `json:"type"`
	Message    string   `json:"message"`
	Suggestion string   `json:"suggestion"`
	Context    string   `json:"context"`
	Options    []string `json:"options"`
	Question   string   `json:"question"`
}

// NewClarificationRequest creates a new clarification request.
func NewClarificationRequest(typeStr, message, suggestion, context string, options []string) *ClarificationRequest {
	return &ClarificationRequest{
		Type:       typeStr,
		Message:    message,
		Suggestion: suggestion,
		Context:    context,
		Options:    options,
	}
}

// Error implements the error interface.
func (e *ClarificationRequest) Error() string {
	return e.Message
}

// ClarificationResponse represents the user's response to a clarification request.
type ClarificationResponse struct {
	// SelectedOption is the option selected by the user from the provided options
	SelectedOption string `json:"selected_option,omitempty"`

	// AdditionalInput is any additional input provided by the user
	AdditionalInput string `json:"additional_input,omitempty"`

	// ClarificationID is the ID of the clarification request being responded to
	ClarificationID string `json:"clarification_id,omitempty"`
}

// ClarificationOption represents a single option for clarification.
type ClarificationOption struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	Value       string `json:"value,omitempty"`
}

// NewIntentClarification creates a clarification request for intent ambiguity.
func NewIntentClarification(question string, options []ClarificationOption) *ClarificationRequest {
	optionStrings := make([]string, len(options))
	for i, opt := range options {
		optionStrings[i] = opt.Label
	}

	return &ClarificationRequest{
		Type:     "intent_ambiguity",
		Message:  "需要澄清查询意图",
		Question: question,
		Options:  optionStrings,
	}
}
