package config

import (
	"flag"
	"os"
	"time"
)

type Config struct {
	Port          string
	OutputDir     string
	Env           string
	RedisAddr     string
	TaskRetention time.Duration
	BaseURL       string
}

func Load() *Config {
	// Define command line flags
	port := flag.String("port", getEnvOrDefault("PORT", "8080"), "Server port")
	outputDir := flag.String("output", getEnvOrDefault("OUTPUT_DIR", "downloads"), "Output directory for downloaded videos")
	env := flag.String("env", getEnvOrDefault("ENV", "development"), "Environment (development/production)")
	redisAddr := flag.String("redis", getEnvOrDefault("REDIS_ADDR", "localhost:6379"), "Redis server address")
	taskRetention := flag.Duration("task-retention", getTaskRetentionFromEnv(), "Task retention period in hours")
	baseURL := flag.String("base-url", getEnvOrDefault("BASE_URL", ""), "Base URL for generating absolute URLs")
	flag.Parse()

	return &Config{
		Port:          *port,
		OutputDir:     *outputDir,
		Env:           *env,
		RedisAddr:     *redisAddr,
		TaskRetention: *taskRetention,
		BaseURL:       *baseURL,
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getTaskRetentionFromEnv() time.Duration {
	retentionStr := os.Getenv("TASK_RETENTION")
	if retentionStr != "" {
		// Try to parse as hours
		if retention, err := time.ParseDuration(retentionStr); err == nil {
			return retention
		}
	}
	// Default: 24 hours
	return 24 * time.Hour
}
