// Package llm provides LLM client interfaces and implementations for the AI Agent module.
// This package supports multiple LLM providers including OpenAI and ARK (ByteDance).
package llm

import (
	"context"

	"github.com/cloudwego/eino/schema"
)

// LLMClientInterface defines the interface for LLM clients.
// This interface abstracts the LLM provider implementation details.
type LLMClientInterface interface {
	// Generate generates a response for the given prompt.
	// This is a synchronous call that returns the complete response.
	Generate(ctx context.Context, prompt string) (string, error)

	// GenerateWithMessages generates a response for the given messages.
	// This allows for multi-turn conversations with context.
	GenerateWithMessages(ctx context.Context, messages []*schema.Message) (string, error)

	// GenerateWithSchema generates a structured response based on the provided schema.
	// This is useful for extracting structured data from LLM responses.
	GenerateWithSchema(ctx context.Context, prompt string, schema interface{}) (interface{}, error)

	// Stream generates a streaming response for the given prompt.
	// This returns a stream reader that yields chunks of the response.
	Stream(ctx context.Context, prompt string) (*schema.StreamReader[*schema.Message], error)

	// StreamWithMessages generates a streaming response for the given messages.
	StreamWithMessages(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error)

	// GetName returns the name of the LLM provider.
	GetName() string
}

// LLMConfig contains the configuration for LLM clients.
type LLMConfig struct {
	// Provider is the LLM provider name (openai, ark)
	Provider string `yaml:"provider"`

	// Model is the model name to use
	Model string `yaml:"model"`

	// APIKey is the API key for authentication
	APIKey string `yaml:"api_key"`

	// Endpoint is the API endpoint (optional, for custom endpoints)
	Endpoint string `yaml:"endpoint"`

	// Temperature is the sampling temperature (0.0 ~ 2.0)
	Temperature float64 `yaml:"temperature"`

	// MaxTokens is the maximum number of tokens to generate
	MaxTokens int `yaml:"max_tokens"`

	// Timeout is the request timeout in seconds
	Timeout int `yaml:"timeout"`

	// RetryCount is the number of retries on failure
	RetryCount int `yaml:"retry_count"`

	// FallbackModel is the fallback model if primary model fails
	FallbackModel string `yaml:"fallback_model"`
}

// DefaultLLMConfig returns the default LLM configuration.
func DefaultLLMConfig() *LLMConfig {
	return &LLMConfig{
		Provider:      "openai",
		Model:         "gpt-4o-mini",
		Temperature:   0.7,
		MaxTokens:     4096,
		Timeout:       30,
		RetryCount:    3,
		FallbackModel: "gpt-3.5-turbo",
	}
}

// LLMError represents an error from LLM operations.
type LLMError struct {
	Provider   string
	StatusCode int
	Message    string
	Err        error
}

// Error implements the error interface.
func (e *LLMError) Error() string {
	if e.Err != nil {
		return e.Provider + " LLM error: " + e.Message + ": " + e.Err.Error()
	}
	return e.Provider + " LLM error: " + e.Message
}

// Unwrap returns the underlying error.
func (e *LLMError) Unwrap() error {
	return e.Err
}

// NewLLMError creates a new LLM error.
func NewLLMError(provider string, statusCode int, message string, err error) *LLMError {
	return &LLMError{
		Provider:   provider,
		StatusCode: statusCode,
		Message:    message,
		Err:        err,
	}
}
