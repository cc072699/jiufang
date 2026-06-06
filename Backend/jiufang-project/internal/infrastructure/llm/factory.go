// Package llm provides LLM client interfaces and implementations.
package llm

import (
	"context"
	"fmt"
	"sync"

	"github.com/cloudwego/eino/schema"
)

// Factory creates LLM clients based on configuration.
type Factory struct {
	clients map[string]LLMClientInterface
	mu      sync.RWMutex
}

// NewFactory creates a new LLM client factory.
func NewFactory() *Factory {
	return &Factory{
		clients: make(map[string]LLMClientInterface),
	}
}

// GetClient returns an LLM client for the given provider.
// If the client doesn't exist, it creates a new one based on the config.
func (f *Factory) GetClient(config *LLMConfig) (LLMClientInterface, error) {
	if config == nil {
		config = DefaultLLMConfig()
	}

	f.mu.RLock()
	client, exists := f.clients[config.Provider]
	f.mu.RUnlock()

	if exists {
		return client, nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	// Double check after acquiring write lock
	if client, exists = f.clients[config.Provider]; exists {
		return client, nil
	}

	// Create new client based on provider
	switch config.Provider {
	case "openai":
		client, err := NewOpenAIClient(config)
		if err != nil {
			return nil, err
		}
		f.clients[config.Provider] = client
		return client, nil

	case "deepseek":
		// DeepSeek uses OpenAI-compatible API
		// Ensure endpoint is set to DeepSeek API base
		if config.Endpoint == "" {
			config.Endpoint = "https://api.deepseek.com/v1"
		}
		client, err := NewOpenAIClient(config)
		if err != nil {
			return nil, err
		}
		f.clients[config.Provider] = client
		return client, nil

	case "siliconflow":
		// SiliconFlow uses OpenAI-compatible API for embeddings
		// Can be used for both LLM and embedding models
		if config.Endpoint == "" {
			config.Endpoint = "https://api.siliconflow.cn/v1"
		}
		client, err := NewOpenAIClient(config)
		if err != nil {
			return nil, err
		}
		f.clients[config.Provider] = client
		return client, nil

	case "xunfei":
			// Xunfei Spark uses OpenAI-compatible API
			if config.Endpoint == "" {
				config.Endpoint = "https://maas-coding-api.cn-huabei-1.xf-yun.com/v2"
			}
			client, err := NewOpenAIClient(config)
			if err != nil {
				return nil, err
			}
			f.clients[config.Provider] = client
			return client, nil

		case "ark", "volcengine":
		client, err := NewARKClient(config)
		if err != nil {
			return nil, err
		}
		f.clients[config.Provider] = client
		return client, nil

	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", config.Provider)
	}
}

// GetClientByName returns an existing client by provider name.
func (f *Factory) GetClientByName(provider string) (LLMClientInterface, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	client, exists := f.clients[provider]
	if !exists {
		return nil, fmt.Errorf("client not found for provider: %s", provider)
	}
	return client, nil
}

// RegisterClient registers a client for a provider.
func (f *Factory) RegisterClient(provider string, client LLMClientInterface) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.clients[provider] = client
}

// DefaultFactory is the default factory instance.
var DefaultFactory = NewFactory()

// GetDefaultClient returns the default LLM client.
func GetDefaultClient(config *LLMConfig) (LLMClientInterface, error) {
	return DefaultFactory.GetClient(config)
}

// FallbackClient provides fallback mechanism when primary LLM fails.
type FallbackClient struct {
	primary   LLMClientInterface
	fallback  LLMClientInterface
	config    *LLMConfig
}

// NewFallbackClient creates a new fallback client.
func NewFallbackClient(primary, fallback LLMClientInterface, config *LLMConfig) *FallbackClient {
	return &FallbackClient{
		primary:  primary,
		fallback: fallback,
		config:   config,
	}
}

// Generate generates a response with fallback.
func (c *FallbackClient) Generate(ctx context.Context, prompt string) (string, error) {
	result, err := c.primary.Generate(ctx, prompt)
	if err != nil {
		if c.fallback != nil {
			return c.fallback.Generate(ctx, prompt)
		}
		return "", err
	}
	return result, nil
}

// GenerateWithMessages generates a response with fallback.
func (c *FallbackClient) GenerateWithMessages(ctx context.Context, messages []*schema.Message) (string, error) {
	result, err := c.primary.GenerateWithMessages(ctx, messages)
	if err != nil {
		if c.fallback != nil {
			return c.fallback.GenerateWithMessages(ctx, messages)
		}
		return "", err
	}
	return result, nil
}

// GenerateWithSchema generates a structured response with fallback.
func (c *FallbackClient) GenerateWithSchema(ctx context.Context, prompt string, schema interface{}) (interface{}, error) {
	result, err := c.primary.GenerateWithSchema(ctx, prompt, schema)
	if err != nil {
		if c.fallback != nil {
			return c.fallback.GenerateWithSchema(ctx, prompt, schema)
		}
		return nil, err
	}
	return result, nil
}

// Stream generates a streaming response with fallback.
func (c *FallbackClient) Stream(ctx context.Context, prompt string) (*schema.StreamReader[*schema.Message], error) {
	result, err := c.primary.Stream(ctx, prompt)
	if err != nil {
		if c.fallback != nil {
			return c.fallback.Stream(ctx, prompt)
		}
		return nil, err
	}
	return result, nil
}

// StreamWithMessages generates a streaming response with fallback.
func (c *FallbackClient) StreamWithMessages(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	result, err := c.primary.StreamWithMessages(ctx, messages)
	if err != nil {
		if c.fallback != nil {
			return c.fallback.StreamWithMessages(ctx, messages)
		}
		return nil, err
	}
	return result, nil
}

// GetName returns the name of the LLM provider.
func (c *FallbackClient) GetName() string {
	return "fallback"
}