// Package llm provides LLM client interfaces and implementations.
// Note: This file provides stub implementations for now.
package llm

import (
	"context"

	"github.com/cloudwego/eino/schema"
)

// OpenAIClient implements LLMClientInterface using OpenAI API.
// Note: This is a stub implementation for compilation purposes.
type OpenAIClient struct {
	model  string
	config *LLMConfig
}

// NewOpenAIClient creates a new OpenAI client.
func NewOpenAIClient(config *LLMConfig) (*OpenAIClient, error) {
	if config.APIKey == "" {
		return nil, NewLLMError("openai", 400, "API key is required", nil)
	}

	return &OpenAIClient{
		model:  config.Model,
		config: config,
	}, nil
}

// Generate generates a response for the given prompt.
func (c *OpenAIClient) Generate(ctx context.Context, prompt string) (string, error) {
	return "", NewLLMError("openai", 501, "OpenAI client not fully implemented", nil)
}

// GenerateWithMessages generates a response for the given messages.
func (c *OpenAIClient) GenerateWithMessages(ctx context.Context, messages []*schema.Message) (string, error) {
	return "", NewLLMError("openai", 501, "OpenAI client not fully implemented", nil)
}

// GenerateWithSchema generates a structured response based on the provided schema.
func (c *OpenAIClient) GenerateWithSchema(ctx context.Context, prompt string, schema interface{}) (interface{}, error) {
	return nil, NewLLMError("openai", 501, "OpenAI client not fully implemented", nil)
}

// Stream generates a streaming response for the given prompt.
func (c *OpenAIClient) Stream(ctx context.Context, prompt string) (*schema.StreamReader[*schema.Message], error) {
	return nil, NewLLMError("openai", 501, "OpenAI client not fully implemented", nil)
}

// StreamWithMessages generates a streaming response for the given messages.
func (c *OpenAIClient) StreamWithMessages(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	return nil, NewLLMError("openai", 501, "OpenAI client not fully implemented", nil)
}

// GetName returns the name of the LLM provider.
func (c *OpenAIClient) GetName() string {
	return "openai"
}

// CreateEinoModel returns nil as this is a stub implementation.
func (c *OpenAIClient) CreateEinoModel() (interface{}, error) {
	return nil, NewLLMError("openai", 501, "OpenAI client not fully implemented", nil)
}
