package handlers

import (
	"context"
	"testing"

	"github.com/sinhaparth5/whatstyle-mcp/configs"
	"github.com/sinhaparth5/whatstyle-mcp/internal/database"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestNewMCPHandler(t *testing.T) {
	// Create test database
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create test config
	config := &configs.Config{
		GrokAPIKey:  "",
		GrokModel:   "grok-beta",
		GrokBaseURL: "https://api.x.ai/v1",
	}

	// Create test implementation
	impl := &mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}

	// Test handler creation
	handler := NewMCPHandler(db, config, impl, nil)
	if handler == nil {
		t.Fatal("Expected handler to be created, got nil")
	}

	if handler.db != db {
		t.Error("Expected handler.db to match provided database")
	}

	if handler.config != config {
		t.Error("Expected handler.config to match provided config")
	}
}

func TestHandleChatTool(t *testing.T) {
	// Create test database
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create test config
	config := &configs.Config{
		GrokAPIKey:  "",
		GrokModel:   "grok-beta",
		GrokBaseURL: "https://api.x.ai/v1",
	}

	// Create test implementation
	impl := &mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}

	// Create handler
	handler := NewMCPHandler(db, config, impl, nil)

	// Test valid chat request
	t.Run("ValidChatRequest", func(t *testing.T) {
		arguments := map[string]interface{}{
			"user_id": "test-user",
			"message": "Hello test",
		}

		result, err := handler.handleChatTool(context.Background(), arguments)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		// Verify result structure
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("Expected result to be map[string]interface{}")
		}

		if _, exists := resultMap["response"]; !exists {
			t.Error("Expected 'response' field in result")
		}

		if _, exists := resultMap["user_id"]; !exists {
			t.Error("Expected 'user_id' field in result")
		}
	})

	// Test invalid request - missing user_id
	t.Run("MissingUserID", func(t *testing.T) {
		arguments := map[string]interface{}{
			"message": "Hello test",
		}

		_, err := handler.handleChatTool(context.Background(), arguments)
		if err == nil {
			t.Error("Expected error for missing user_id")
		}
	})

	// Test invalid request - missing message
	t.Run("MissingMessage", func(t *testing.T) {
		arguments := map[string]interface{}{
			"user_id": "test-user",
		}

		_, err := handler.handleChatTool(context.Background(), arguments)
		if err == nil {
			t.Error("Expected error for missing message")
		}
	})
}

func TestHandleHistoryTool(t *testing.T) {
	// Create test database
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create test config
	config := &configs.Config{
		GrokAPIKey:  "",
		GrokModel:   "grok-beta",
		GrokBaseURL: "https://api.x.ai/v1",
	}

	// Create test implementation
	impl := &mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}

	// Create handler
	handler := NewMCPHandler(db, config, impl, nil)

	// Add some test messages
	userID := "test-user"
	_ = db.SaveMessage(userID, "Hello", "user")
	_ = db.SaveMessage(userID, "Hi there!", "assistant")

	// Test valid history request
	t.Run("ValidHistoryRequest", func(t *testing.T) {
		arguments := map[string]interface{}{
			"user_id": userID,
			"limit":   float64(10),
		}

		result, err := handler.handleHistoryTool(context.Background(), arguments)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		// Verify result structure
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("Expected result to be map[string]interface{}")
		}

		messages, exists := resultMap["messages"]
		if !exists {
			t.Error("Expected 'messages' field in result")
		}

		messagesSlice, ok := messages.([]map[string]interface{})
		if !ok {
			t.Error("Expected messages to be slice")
		}

		if len(messagesSlice) != 2 {
			t.Errorf("Expected 2 messages, got %d", len(messagesSlice))
		}
	})

	// Test invalid request - missing user_id
	t.Run("MissingUserID", func(t *testing.T) {
		arguments := map[string]interface{}{
			"limit": float64(10),
		}

		_, err := handler.handleHistoryTool(context.Background(), arguments)
		if err == nil {
			t.Error("Expected error for missing user_id")
		}
	})
}