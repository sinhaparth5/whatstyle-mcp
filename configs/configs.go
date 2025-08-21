package configs

import (
	"os"
)

type Config struct {
	Port         string
	DatabasePath string
	Environment  string
	GrokAPIKey   string
	GrokModel    string
	GrokBaseURL  string

	// WhatsApp Business API
	WhatsAppAccessToken   string
	WhatsAppVerifyToken   string
	WhatsAppPhoneNumberID string
	WhatsAppWebhookURL    string
}

func Load() *Config {
	config := &Config{
		Port:         getEnv("PORT", "8080"),
		DatabasePath: getEnv("DATABASE_PATH", "./mcp_server.db"),
		Environment:  getEnv("ENVIRONMENT", "development"),
		GrokAPIKey:   getEnv("GROK_API_KEY", ""),
		GrokModel:    getEnv("GROK_MODEL", "grok-beta"),
		GrokBaseURL:  getEnv("GROK_BASE_URL", "https://api.x.ai/v1"),

		// WhatsApp Business API
		WhatsAppAccessToken:   getEnv("WHATSAPP_ACCESS_TOKEN", ""),
		WhatsAppVerifyToken:   getEnv("WHATSAPP_VERIFY_TOKEN", ""),
		WhatsAppPhoneNumberID: getEnv("WHATSAPP_PHONE_NUMBER_ID", ""),
		WhatsAppWebhookURL:    getEnv("WHATSAPP_BASE_URL", ""),
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

