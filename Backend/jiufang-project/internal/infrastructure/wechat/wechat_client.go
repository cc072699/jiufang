// Package wechat implements the WeChat Work API client.
package wechat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// WeChatClientInterface defines the interface for WeChat Work API client.
type WeChatClientInterface interface {
	SendMessage(ctx context.Context, req *SendMessageRequest) error
}

// SendMessageRequest represents the request to send a message.
type SendMessageRequest struct {
	ToUser  string `json:"touser"`   // User ID list, separated by '|'
	MsgType string `json:"msgtype"`  // Message type: markdown
	Content string `json:"content"`  // Message content (Markdown format)
}

// WeChatClient implements WeChat Work API client.
type WeChatClient struct {
	webhookURL string
	httpClient *http.Client
	logger     *zap.Logger
	maxRetries int
	retryDelay time.Duration
}

// WeChatClientConfig represents the configuration for WeChat client.
type WeChatClientConfig struct {
	WebhookURL string
	MaxRetries int
	RetryDelay time.Duration
	Timeout    time.Duration
}

// NewWeChatClient creates a new WeChatClient instance.
func NewWeChatClient(config *WeChatClientConfig, logger *zap.Logger) *WeChatClient {
	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	if config.MaxRetries <= 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay <= 0 {
		config.RetryDelay = 5 * time.Second
	}
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}

	return &WeChatClient{
		webhookURL: config.WebhookURL,
		httpClient: httpClient,
		logger:     logger,
		maxRetries: config.MaxRetries,
		retryDelay: config.RetryDelay,
	}
}

// SendMessage sends a message to WeChat Work.
func (c *WeChatClient) SendMessage(ctx context.Context, req *SendMessageRequest) error {
	// Validate request
	if req.ToUser == "" {
		return fmt.Errorf("touser is required")
	}
	if req.MsgType == "" {
		req.MsgType = "markdown"
	}
	if req.Content == "" {
		return fmt.Errorf("content is required")
	}

	// Prepare request body
	body := map[string]interface{}{
		"msgtype": req.MsgType,
		req.MsgType: map[string]string{
			"content": req.Content,
		},
	}

	// Add touser if specified
	if req.ToUser != "" {
		body["touser"] = req.ToUser
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Send request with retry
	var lastErr error
	for retry := 0; retry <= c.maxRetries; retry++ {
		if retry > 0 {
			c.logger.Info("retrying WeChat message send",
				zap.Int("retry", retry),
				zap.Duration("delay", c.retryDelay),
			)
			time.Sleep(c.retryDelay)
		}

		err := c.sendHTTPRequest(ctx, jsonBody)
		if err == nil {
			c.logger.Info("WeChat message sent successfully",
				zap.String("touser", req.ToUser),
				zap.String("msgtype", req.MsgType),
			)
			return nil
		}

		lastErr = err
		c.logger.Warn("failed to send WeChat message",
			zap.Int("retry", retry),
			zap.Error(err),
		)
	}

	return fmt.Errorf("failed to send WeChat message after %d retries: %w", c.maxRetries, lastErr)
}

// sendHTTPRequest sends HTTP request to WeChat API.
func (c *WeChatClient) sendHTTPRequest(ctx context.Context, body []byte) error {
	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("WeChat API returned non-200 status: %d, body: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var weChatResp WeChatResponse
	if err := json.Unmarshal(respBody, &weChatResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Check response status
	if weChatResp.ErrCode != 0 {
		return fmt.Errorf("WeChat API error: code=%d, message=%s", weChatResp.ErrCode, weChatResp.ErrMsg)
	}

	return nil
}

// WeChatResponse represents the response from WeChat API.
type WeChatResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// SendMarkdownMessage sends a Markdown message to specified users.
func (c *WeChatClient) SendMarkdownMessage(ctx context.Context, toUsers []string, content string) error {
	// Join user IDs with '|'
	toUserStr := ""
	for i, user := range toUsers {
		if i > 0 {
			toUserStr += "|"
		}
		toUserStr += user
	}

	req := &SendMessageRequest{
		ToUser:  toUserStr,
		MsgType: "markdown",
		Content: content,
	}

	return c.SendMessage(ctx, req)
}