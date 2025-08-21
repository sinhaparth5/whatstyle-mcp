package grok

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sinhaparth5/whatstyle-mcp/internal/models"
)

type Client struct {
	APIKey  string
	BaseURL string
	Model   string
	client  *http.Client
}

type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

func NewClient(apiKey, baseURL, model string) *Client {
	return &Client{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   model,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) GenerateResponse(userMessage string, history []models.Message) (string, error) {
	// Convert history to Grok messages format
	messages := c.convertHistoryToMessages(history)

	// Add current user message
	messages = append(messages, Message{
		Role:    "user",
		Content: userMessage,
	})

	// Create request
	req := ChatCompletionRequest{
		Model:       c.Model,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   1000,
		Stream:      false,
	}

	// Make API call
	response, err := c.makeAPICall(req)
	if err != nil {
		return "", fmt.Errorf("grok API call failed: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned from Grok")
	}

	return response.Choices[0].Message.Content, nil
}

func (c *Client) convertHistoryToMessages(history []models.Message) []Message {
	var messages []Message

	// Add system message
	messages = append(messages, Message{
		Role:    "system",
		Content: "You are a helpful AI assistant integrated with WhatsApp. Provide concise, helpful responses to user messages. Keep responses conversational and appropriate for a messaging context.",
	})

	// Convert last 10 messages for context (to avoid token limits)
	start := 0
	if len(history) > 10 {
		start = len(history) - 10
	}

	for i := start; i < len(history); i++ {
		msg := history[i]
		messages = append(messages, Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	return messages
}

func (c *Client) makeAPICall(req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	// Marshal request
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", c.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	// Make request
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("API error: %s", errorResp.Error.Message)
	}

	// Parse response
	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &chatResp, nil
}
