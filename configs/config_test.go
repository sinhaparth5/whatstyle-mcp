package configs

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Test default values
	t.Run("DefaultValues", func(t *testing.T) {
		// Clear environment variables
		os.Unsetenv("PORT")
		os.Unsetenv("GROK_API_KEY")
		os.Unsetenv("DATABASE_PATH")

		config := Load()

		if config.Port != "8080" {
			t.Errorf("Expected default port 8080, got %s", config.Port)
		}

		if config.DatabasePath != "./mcp_server.db" {
			t.Errorf("Expected default database path, got %s", config.DatabasePath)
		}

		if config.Environment != "development" {
			t.Errorf("Expected default environment development, got %s", config.Environment)
		}

		if config.GrokModel != "grok-beta" {
			t.Errorf("Expected default grok model grok-beta, got %s", config.GrokModel)
		}
	})

	// Test environment variable override
	t.Run("EnvironmentOverride", func(t *testing.T) {
		os.Setenv("PORT", "3000")
		os.Setenv("GROK_API_KEY", "test-api-key")
		os.Setenv("ENVIRONMENT", "production")

		config := Load()

		if config.Port != "3000" {
			t.Errorf("Expected port 3000, got %s", config.Port)
		}

		if config.GrokAPIKey != "test-api-key" {
			t.Errorf("Expected grok api key test-api-key, got %s", config.GrokAPIKey)
		}

		if config.Environment != "production" {
			t.Errorf("Expected environment production, got %s", config.Environment)
		}

		// Cleanup
		os.Unsetenv("PORT")
		os.Unsetenv("GROK_API_KEY")
		os.Unsetenv("ENVIRONMENT")
	})
}

func TestGetEnv(t *testing.T) {
	// Test with existing environment variable
	t.Run("ExistingEnvVar", func(t *testing.T) {
		os.Setenv("TEST_VAR", "test_value")
		defer os.Unsetenv("TEST_VAR")

		result := getEnv("TEST_VAR", "default_value")
		if result != "test_value" {
			t.Errorf("Expected test_value, got %s", result)
		}
	})

	// Test with non-existing environment variable
	t.Run("NonExistingEnvVar", func(t *testing.T) {
		result := getEnv("NON_EXISTING_VAR", "default_value")
		if result != "default_value" {
			t.Errorf("Expected default_value, got %s", result)
		}
	})

	// Test with empty environment variable
	t.Run("EmptyEnvVar", func(t *testing.T) {
		os.Setenv("EMPTY_VAR", "")
		defer os.Unsetenv("EMPTY_VAR")

		result := getEnv("EMPTY_VAR", "default_value")
		if result != "default_value" {
			t.Errorf("Expected default_value for empty env var, got %s", result)
		}
	})
}