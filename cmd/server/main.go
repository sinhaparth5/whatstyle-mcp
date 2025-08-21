package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/sinhaparth5/whatstyle-mcp/configs"
	"github.com/sinhaparth5/whatstyle-mcp/internal/database"
	"github.com/sinhaparth5/whatstyle-mcp/internal/handlers"
	"github.com/sinhaparth5/whatstyle-mcp/internal/whatsapp"

	"github.com/gorilla/mux"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	// Load configuration
	config := configs.Load()

	// Validate required configuration
	if config.GrokAPIKey == "" {
		log.Printf("Warning: GROK_API_KEY not set. Server will use fallback responses.")
	}

	// Initialize database
	db, err := database.InitDB(config.DatabasePath)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Start background cleanup routine
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		
		for range ticker.C {
			if err := db.CleanupExpiredSessions(); err != nil {
				log.Printf("Error cleaning up sessions: %v", err)
			}
		}
	}()

	// Create MCP server implementation
	implementation := &mcp.Implementation{
		Name:    "whatsapp-mcp-server",
		Version: "1.0.0",
	}

	// Initialize MCP handlers
	mcpHandler := handlers.NewMCPHandler(db, config, implementation, nil)
	whatsappHandler := whatsapp.NewHandler(config)

	// Create MCP server
	server := mcp.NewServer(implementation, nil)
	
	// Register tools with the MCP server
	mcpHandler.RegisterTools(server)

	// Setup HTTP routes
	router := mux.NewRouter()
	
	// MCP endpoint - handle MCP protocol over HTTP
	router.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			mcpHandler.HandleHTTPRequest(w, r, server)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}).Methods("POST", "OPTIONS")
	
	// WhatsApp webhook endpoints
	router.HandleFunc("/webhook", whatsappHandler.VerifyWebhook).Methods("GET")
	router.HandleFunc("/webhook", whatsappHandler.HandleWebhook).Methods("POST")
	
	// Health and info endpoints
	router.HandleFunc("/health", mcpHandler.HealthCheck).Methods("GET")
	router.HandleFunc("/tools", func(w http.ResponseWriter, r *http.Request) {
		tools := mcpHandler.GetAvailableTools()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"tools": tools,
		})
	}).Methods("GET")
	
	router.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		stats, err := db.GetStats()
		if err != nil {
			http.Error(w, "Failed to get stats", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}).Methods("GET")

	// Add middleware
	router.Use(loggingMiddleware)
	router.Use(corsMiddleware)

	log.Printf("WhatsApp MCP Server starting...")
	log.Printf("Repository: github.com/sinhaparth5/whatstyle-mcp")
	log.Printf("Port: %s", config.Port)
	log.Printf("Environment: %s", config.Environment)
	log.Printf("Grok Model: %s", config.GrokModel)
	log.Printf("Database: %s", config.DatabasePath)
	log.Printf("")
	log.Printf("Endpoints:")
	log.Printf("  Health:   http://localhost:%s/health", config.Port)
	log.Printf("  Tools:    http://localhost:%s/tools", config.Port)
	log.Printf("  Stats:    http://localhost:%s/stats", config.Port)
	log.Printf("  MCP:      http://localhost:%s/mcp", config.Port)
	log.Printf("  Webhook:  http://localhost:%s/webhook", config.Port)

	// Start server
	port := config.Port
	if port == "" {
		port = os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
	}

	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s] %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}