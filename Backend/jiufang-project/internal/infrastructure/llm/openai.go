// Package llm provides LLM client interfaces and implementations.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cloudwego/eino/schema"
)

// OpenAIClient implements LLMClientInterface using OpenAI-compatible API.
type OpenAIClient struct {
	model  string
	config *LLMConfig
	httpClient *http.Client
}

// NewOpenAIClient creates a new OpenAI-compatible client.
func NewOpenAIClient(config *LLMConfig) (*OpenAIClient, error) {
	if config.APIKey == "" {
		return nil, NewLLMError("openai", 400, "API key is required", nil)
	}

	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 60
	}

	return &OpenAIClient{
		model:  config.Model,
		config: config,
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}, nil
}

// chatCompletionRequest represents an OpenAI-compatible chat completion request.
type chatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

// chatMessage represents a message in the chat completion request/response.
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatCompletionResponse represents an OpenAI-compatible chat completion response.
type chatCompletionResponse struct {
	ID      string              `json:"id"`
	Object  string              `json:"object"`
	Model   string              `json:"model"`
	Choices []chatCompletionChoice `json:"choices"`
	Usage   *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// chatCompletionChoice represents a choice in the chat completion response.
type chatCompletionChoice struct {
	Index        int         `json:"index"`
	Message      chatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// Generate generates a response for the given prompt using OpenAI-compatible API.
func (c *OpenAIClient) Generate(ctx context.Context, prompt string) (string, error) {
	endpoint := c.config.Endpoint
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1"
	}

	reqBody := chatCompletionRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "user", Content: prompt},
		},
		Temperature: c.config.Temperature,
		MaxTokens:   c.config.MaxTokens,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", NewLLMError("openai", 500, "failed to marshal request", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", NewLLMError("openai", 500, "failed to create request", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", NewLLMError("openai", 0, "request failed", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", NewLLMError("openai", 0, "failed to read response", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if json.Unmarshal(respBytes, &errResp) == nil && errResp.Error != nil {
			return "", NewLLMError("openai", resp.StatusCode, errResp.Error.Message, nil)
		}
		return "", NewLLMError("openai", resp.StatusCode, fmt.Sprintf("unexpected status %d", resp.StatusCode), nil)
	}

	var chatResp chatCompletionResponse
	if err := json.Unmarshal(respBytes, &chatResp); err != nil {
		return "", NewLLMError("openai", 0, "failed to parse response", err)
	}

	if chatResp.Error != nil {
		return "", NewLLMError("openai", resp.StatusCode, chatResp.Error.Message, nil)
	}

	if len(chatResp.Choices) == 0 {
		return "", NewLLMError("openai", 0, "no choices in response", nil)
	}

	return chatResp.Choices[0].Message.Content, nil
}

// GenerateWithMessages generates a response for the given messages.
func (c *OpenAIClient) GenerateWithMessages(ctx context.Context, messages []*schema.Message) (string, error) {
	// Convert messages to API format
	apiMessages := make([]chatMessage, 0, len(messages))
	for _, msg := range messages {
		role := "user"
		if msg.Role == schema.Assistant {
			role = "assistant"
		} else if msg.Role == schema.System {
			role = "system"
		}
			apiMessages = append(apiMessages, chatMessage{Role: role, Content: msg.Content})
	}

	endpoint := c.config.Endpoint
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1"
	}

	reqBody := chatCompletionRequest{
		Model:       c.model,
		Messages:    apiMessages,
		Temperature: c.config.Temperature,
		MaxTokens:   c.config.MaxTokens,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", NewLLMError("openai", 500, "failed to marshal request", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", NewLLMError("openai", 500, "failed to create request", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", NewLLMError("openai", 0, "request failed", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", NewLLMError("openai", 0, "failed to read response", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if json.Unmarshal(respBytes, &errResp) == nil && errResp.Error != nil {
			return "", NewLLMError("openai", resp.StatusCode, errResp.Error.Message, nil)
		}
		return "", NewLLMError("openai", resp.StatusCode, fmt.Sprintf("unexpected status %d", resp.StatusCode), nil)
	}

	var chatResp chatCompletionResponse
	if err := json.Unmarshal(respBytes, &chatResp); err != nil {
		return "", NewLLMError("openai", 0, "failed to parse response", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", NewLLMError("openai", 0, "no choices in response", nil)
	}

	return chatResp.Choices[0].Message.Content, nil
}

// GenerateWithSchema generates a structured response based on the provided schema.
func (c *OpenAIClient) GenerateWithSchema(ctx context.Context, prompt string, schema interface{}) (interface{}, error) {
	return "", NewLLMError("openai", 501, "OpenAI client GenerateWithSchema not fully implemented", nil)
}

// Stream generates a streaming response for the given prompt.
func (c *OpenAIClient) Stream(ctx context.Context, prompt string) (*schema.StreamReader[*schema.Message], error) {
	return nil, NewLLMError("openai", 501, "OpenAI client Stream not fully implemented", nil)
}

// StreamWithMessages generates a streaming response for the given messages.
func (c *OpenAIClient) StreamWithMessages(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	return nil, NewLLMError("openai", 501, "OpenAI client StreamWithMessages not fully implemented", nil)
}

// GetName returns the name of the LLM provider.
func (c *OpenAIClient) GetName() string {
	return "openai"
}

// CreateEinoModel returns nil as this is a minimal implementation.
func (c *OpenAIClient) CreateEinoModel() (interface{}, error) {
	return nil, NewLLMError("openai", 501, "OpenAI client CreateEinoModel not fully implemented", nil)
}
