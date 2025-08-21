package mcp

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Re-export types from the official SDK
type (
	// Core MCP types
	Implementation = mcp.Implementation
	Server         = mcp.Server
	
	// Transport types  
	Transport        = mcp.Transport
	CommandTransport = mcp.CommandTransport
	StdioTransport   = mcp.StdioTransport
)

// Constants
const (
	ProtocolVersion = "2024-11-05"
	
	// Standard MCP error codes
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

// Custom types for our application
type ChatParams struct {
	UserID  string `json:"user_id"`
	Message string `json:"message"`
}

type ChatResult struct {
	Response string `json:"response"`
	UserID   string `json:"user_id"`
}

type HistoryParams struct {
	UserID string `json:"user_id"`
	Limit  *int   `json:"limit,omitempty"`
}

type HistoryResult struct {
	Messages []ChatMessage `json:"messages"`
	UserID   string        `json:"user_id"`
}

type ChatMessage struct {
	ID        int    `json:"id"`
	UserID    string `json:"user_id"`
	Content   string `json:"content"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

// Helper functions for creating MCP tool definitions
func NewChatToolDefinition() map[string]interface{} {
	return map[string]interface{}{
		"name":        "chat",
		"description": "Send a chat message and get AI response using Grok",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"user_id": map[string]interface{}{
					"type":        "string",
					"description": "Unique identifier for the user",
				},
				"message": map[string]interface{}{
					"type":        "string",
					"description": "The message content",
				},
			},
			"required": []string{"user_id", "message"},
		},
	}
}

func NewHistoryToolDefinition() map[string]interface{} {
	return map[string]interface{}{
		"name":        "history",
		"description": "Get chat history for a user",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"user_id": map[string]interface{}{
					"type":        "string",
					"description": "Unique identifier for the user",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of messages to return",
					"default":     20,
				},
			},
			"required": []string{"user_id"},
		},
	}
}

// Helper functions for MCP responses
func NewTextContent(text string) map[string]interface{} {
	return map[string]interface{}{
		"type": "text",
		"text": text,
	}
}

func NewSuccessResult(content []map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"content": content,
	}
}

func NewErrorResult(message string) map[string]interface{} {
	return map[string]interface{}{
		"content": []map[string]interface{}{
			NewTextContent(message),
		},
		"isError": true,
	}
}