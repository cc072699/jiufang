// Package llm provides LLM client interfaces and implementations.
// Note: This file provides stub implementations for now.
package llm

import (
	"context"

	"github.com/cloudwego/eino/schema"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
)

// ARKClient implements LLMClientInterface using ByteDance ARK API.
// Note: This is a stub implementation for compilation purposes.
type ARKClient struct {
	client *arkruntime.Client
	model  string
	config *LLMConfig
}

// NewARKClient creates a new ARK client.
func NewARKClient(config *LLMConfig) (*ARKClient, error) {
	if config.APIKey == "" {
		return nil, NewLLMError("ark", 400, "API key is required", nil)
	}

	client := arkruntime.NewClientWithApiKey(config.APIKey)
	if config.Endpoint != "" {
		client = arkruntime.NewClientWithApiKey(config.APIKey, arkruntime.WithBaseUrl(config.Endpoint))
	}

	return &ARKClient{
		client: client,
		model:  config.Model,
		config: config,
	}, nil
}

// Generate generates a response for the given prompt.
func (c *ARKClient) Generate(ctx context.Context, prompt string) (string, error) {
	return "", NewLLMError("ark", 501, "ARK client not fully implemented", nil)
}

// GenerateWithMessages generates a response for the given messages.
func (c *ARKClient) GenerateWithMessages(ctx context.Context, messages []*schema.Message) (string, error) {
	return "", NewLLMError("ark", 501, "ARK client not fully implemented", nil)
}

// GenerateWithSchema generates a structured response based on the provided schema.
func (c *ARKClient) GenerateWithSchema(ctx context.Context, prompt string, schema interface{}) (interface{}, error) {
	return nil, NewLLMError("ark", 501, "ARK client not fully implemented", nil)
}

// Stream generates a streaming response for the given prompt.
func (c *ARKClient) Stream(ctx context.Context, prompt string) (*schema.StreamReader[*schema.Message], error) {
	return nil, NewLLMError("ark", 501, "ARK client not fully implemented", nil)
}

// StreamWithMessages generates a streaming response for the given messages.
func (c *ARKClient) StreamWithMessages(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	return nil, NewLLMError("ark", 501, "ARK client not fully implemented", nil)
}

// GetName returns the name of the LLM provider.
func (c *ARKClient) GetName() string {
	return "ark"
}
