package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sinhaparth5/whatstyle-mcp/configs"
	"github.com/sinhaparth5/whatstyle-mcp/internal/database"
	"github.com/sinhaparth5/whatstyle-mcp/internal/grok"
	"github.com/sinhaparth5/whatstyle-mcp/internal/models"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type MCPHandler struct {
	db         *database.DB
	grokClient *grok.Client
	config     *configs.Config
	server     *mcp.Server
}

func NewMCPHandler(db *database.DB, config *configs.Config, impl *mcp.Implementation, caps interface{}) *MCPHandler {
	h := &MCPHandler{
		db:     db,
		config: config,
	}
	
	// Initialize Grok client
	if config.GrokAPIKey != "" {
		h.grokClient = grok.NewClient(config.GrokAPIKey, config.GrokBaseURL, config.GrokModel)
		log.Printf("Grok client initialized with model: %s", config.GrokModel)
	} else {
		log.Printf("Warning: GROK_API_KEY not set, using fallback responses")
	}
	
	// Create MCP server
	h.server = mcp.NewServer(impl, nil)
	
	return h
}

func (h *MCPHandler) RegisterTools(server *mcp.Server) {
	// Tools will be handled through HTTP interface
	log.Printf("MCP tools registered: chat, history")
}

func (h *MCPHandler) handleChatTool(ctx context.Context, arguments map[string]interface{}) (interface{}, error) {
	// Extract parameters
	userID, ok := arguments["user_id"].(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user_id is required and must be a string")
	}

	message, ok := arguments["message"].(string)
	if !ok || message == "" {
		return nil, fmt.Errorf("message is required and must be a string")
	}

	// Save user message
	if err := h.db.SaveMessage(userID, message, "user"); err != nil {
		log.Printf("Error saving user message: %v", err)
		return nil, fmt.Errorf("failed to save message: %v", err)
	}

	// Get chat history for context
	history, err := h.db.GetChatHistory(userID, 10)
	if err != nil {
		log.Printf("Error getting chat history: %v", err)
	}

	// Generate response using Grok
	response, err := h.generateResponse(message, history)
	if err != nil {
		log.Printf("Error generating response: %v", err)
		response = "I apologize, but I'm having trouble generating a response right now. Please try again."
	}

	// Save assistant response
	if err := h.db.SaveMessage(userID, response, "assistant"); err != nil {
		log.Printf("Error saving assistant message: %v", err)
	}

	// Return successful result
	result := map[string]interface{}{
		"response": response,
		"user_id":  userID,
	}

	return result, nil
}

func (h *MCPHandler) handleHistoryTool(ctx context.Context, arguments map[string]interface{}) (interface{}, error) {
	// Extract parameters
	userID, ok := arguments["user_id"].(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user_id is required and must be a string")
	}

	limit := 20 // default
	if l, ok := arguments["limit"].(float64); ok {
		limit = int(l)
	}

	// Get chat history
	history, err := h.db.GetChatHistory(userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat history: %v", err)
	}

	// Convert to response format
	var messages []map[string]interface{}
	for _, msg := range history {
		messages = append(messages, map[string]interface{}{
			"id":         msg.ID,
			"user_id":    msg.UserID,
			"content":    msg.Content,
			"role":       msg.Role,
			"created_at": msg.CreatedAt.Format(time.RFC3339),
		})
	}

	result := map[string]interface{}{
		"messages": messages,
		"user_id":  userID,
	}

	return result, nil
}

func (h *MCPHandler) HandleHTTPRequest(w http.ResponseWriter, r *http.Request, server *mcp.Server) {
	// Handle MCP requests over HTTP
	var request map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	method, ok := request["method"].(string)
	if !ok {
		http.Error(w, "Method required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	var response interface{}

	switch method {
	case "initialize":
		response = map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      request["id"],
			"result": map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{},
				},
				"serverInfo": map[string]interface{}{
					"name":    "whatsapp-mcp-server",
					"version": "1.0.0",
				},
			},
		}

	case "tools/list":
		tools := h.GetAvailableTools()
		response = map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      request["id"],
			"result": map[string]interface{}{
				"tools": tools,
			},
		}

	case "tools/call":
		params, ok := request["params"].(map[string]interface{})
		if !ok {
			http.Error(w, "Invalid params", http.StatusBadRequest)
			return
		}

		toolName, ok := params["name"].(string)
		if !ok {
			http.Error(w, "Tool name required", http.StatusBadRequest)
			return
		}

		arguments, ok := params["arguments"].(map[string]interface{})
		if !ok {
			arguments = make(map[string]interface{})
		}

		var result interface{}
		var err error

		switch toolName {
		case "chat":
			result, err = h.handleChatTool(ctx, arguments)
		case "history":
			result, err = h.handleHistoryTool(ctx, arguments)
		default:
			response = map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      request["id"],
				"error": map[string]interface{}{
					"code":    -32601,
					"message": "Tool not found",
				},
			}
			break
		}

		if err != nil {
			response = map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      request["id"],
				"error": map[string]interface{}{
					"code":    -32603,
					"message": err.Error(),
				},
			}
		} else {
			// Format as MCP tool result
			toolResult := map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": fmt.Sprintf("%v", result),
					},
				},
			}
			
			// If result is already JSON-serializable, convert to JSON string
			if resultBytes, err := json.Marshal(result); err == nil {
				toolResult["content"] = []map[string]interface{}{
					{
						"type": "text",
						"text": string(resultBytes),
					},
				}
			}

			response = map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      request["id"],
				"result":  toolResult,
			}
		}

	default:
		response = map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      request["id"],
			"error": map[string]interface{}{
				"code":    -32601,
				"message": "Method not found",
			},
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *MCPHandler) GetAvailableTools() []map[string]interface{} {
	return []map[string]interface{}{
		{
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
		},
		{
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
		},
	}
}

// Generate response using Grok API or fallback
func (h *MCPHandler) generateResponse(userMessage string, history []models.Message) (string, error) {
	// Try Grok API first
	if h.grokClient != nil {
		response, err := h.grokClient.GenerateResponse(userMessage, history)
		if err != nil {
			log.Printf("Grok API error: %v", err)
			// Fall through to fallback
		} else {
			return response, nil
		}
	}

	// Fallback responses when Grok is unavailable
	fallbackResponses := []string{
		"I understand what you're saying. Let me help you with that.",
		"That's an interesting point. Here's what I think about it.",
		"I see what you mean. Let me provide some assistance.",
		"Thanks for sharing that with me. I'm here to help.",
		"I appreciate your message. How can I assist you further?",
	}

	// Simple hash-based response selection for consistency
	hash := 0
	for _, char := range userMessage {
		hash += int(char)
	}

	responseIndex := hash % len(fallbackResponses)
	return fallbackResponses[responseIndex], nil
}

func (h *MCPHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	grokStatus := "configured"
	if h.grokClient == nil {
		grokStatus = "not configured"
	}

	response := map[string]interface{}{
		"status":     "healthy",
		"service":    "WhatsApp MCP Server",
		"timestamp":  time.Now().Format(time.RFC3339),
		"version":    "1.0.0",
		"sdk":        "official-go-sdk",
		"grok_api":   grokStatus,
		"model":      h.config.GrokModel,
		"repository": "github.com/sinhaparth5/whatstyle-mcp",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}