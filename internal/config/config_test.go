package config

import (
	"os"
	"testing"
	"time"
)

// TestGetEnvOrDefault tests the getEnvOrDefault helper function
func TestGetEnvOrDefault(t *testing.T) {
	// Save the original environment
	originalTestVar := os.Getenv("TEST_VAR")
	defer os.Setenv("TEST_VAR", originalTestVar)

	tests := []struct {
		name         string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			name:         "Environment variable set",
			envValue:     "env_value",
			defaultValue: "default_value",
			expected:     "env_value",
		},
		{
			name:         "Environment variable not set",
			envValue:     "",
			defaultValue: "default_value",
			expected:     "default_value",
		},
		{
			name:         "Environment variable is empty string",
			envValue:     "",
			defaultValue: "default_value",
			expected:     "default_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("TEST_VAR", tt.envValue)
			result := getEnvOrDefault("TEST_VAR", tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvOrDefault() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetTaskRetentionFromEnv tests the getTaskRetentionFromEnv helper function
func TestGetTaskRetentionFromEnv(t *testing.T) {
	// Save the original environment
	originalTaskRetention := os.Getenv("TASK_RETENTION")
	defer os.Setenv("TASK_RETENTION", originalTaskRetention)

	tests := []struct {
		name     string
		envValue string
		expected time.Duration
	}{
		{
			name:     "Valid hours",
			envValue: "48h",
			expected: 48 * time.Hour,
		},
		{
			name:     "Valid minutes",
			envValue: "30m",
			expected: 30 * time.Minute,
		},
		{
			name:     "Invalid format",
			envValue: "invalid",
			expected: 24 * time.Hour, // Default value
		},
		{
			name:     "Empty value",
			envValue: "",
			expected: 24 * time.Hour, // Default value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("TASK_RETENTION", tt.envValue)
			result := getTaskRetentionFromEnv()
			if result != tt.expected {
				t.Errorf("getTaskRetentionFromEnv() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestConfigFields tests that the Config struct contains the expected fields
func TestConfigFields(t *testing.T) {
	// Create a config with known values
	config := &Config{
		Port:          "3000",
		OutputDir:     "/tmp/videos",
		Env:           "production",
		RedisAddr:     "redis:6379",
		TaskRetention: 48 * time.Hour,
		BaseURL:       "https://example.com",
	}

	// Verify the values
	if config.Port != "3000" {
		t.Errorf("Port = %v, want %v", config.Port, "3000")
	}
	if config.OutputDir != "/tmp/videos" {
		t.Errorf("OutputDir = %v, want %v", config.OutputDir, "/tmp/videos")
	}
	if config.Env != "production" {
		t.Errorf("Env = %v, want %v", config.Env, "production")
	}
	if config.RedisAddr != "redis:6379" {
		t.Errorf("RedisAddr = %v, want %v", config.RedisAddr, "redis:6379")
	}
	if config.TaskRetention != 48*time.Hour {
		t.Errorf("TaskRetention = %v, want %v", config.TaskRetention, 48*time.Hour)
	}
	if config.BaseURL != "https://example.com" {
		t.Errorf("BaseURL = %v, want %v", config.BaseURL, "https://example.com")
	}
}

// TestEnvironmentVariableOverrides tests that environment variables properly override defaults
// Note: We can't directly test Load() because it defines flags which can't be redefined in tests
func TestEnvironmentVariableOverrides(t *testing.T) {
	// Test that getEnvOrDefault returns values from environment variables correctly
	t.Run("Port from environment", func(t *testing.T) {
		originalPort := os.Getenv("PORT")
		defer os.Setenv("PORT", originalPort)

		os.Setenv("PORT", "3000")
		value := getEnvOrDefault("PORT", "8080")
		if value != "3000" {
			t.Errorf("Port = %v, want %v", value, "3000")
		}
	})

	t.Run("OutputDir from environment", func(t *testing.T) {
		originalOutputDir := os.Getenv("OUTPUT_DIR")
		defer os.Setenv("OUTPUT_DIR", originalOutputDir)

		os.Setenv("OUTPUT_DIR", "/tmp/videos")
		value := getEnvOrDefault("OUTPUT_DIR", "downloads")
		if value != "/tmp/videos" {
			t.Errorf("OutputDir = %v, want %v", value, "/tmp/videos")
		}
	})

	t.Run("Task retention from environment", func(t *testing.T) {
		originalTaskRetention := os.Getenv("TASK_RETENTION")
		defer os.Setenv("TASK_RETENTION", originalTaskRetention)

		os.Setenv("TASK_RETENTION", "48h")
		duration := getTaskRetentionFromEnv()
		if duration != 48*time.Hour {
			t.Errorf("TaskRetention = %v, want %v", duration, 48*time.Hour)
		}
	})
}
