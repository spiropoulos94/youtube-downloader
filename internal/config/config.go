package config

import (
	"flag"
	"os"
)

type Config struct {
	Port      string
	OutputDir string
	Env       string
	RedisAddr string
}

func Load() *Config {
	// Define command line flags
	port := flag.String("port", getEnvOrDefault("PORT", "8080"), "Server port")
	outputDir := flag.String("output", getEnvOrDefault("OUTPUT_DIR", "downloads"), "Output directory for downloaded videos")
	env := flag.String("env", getEnvOrDefault("ENV", "development"), "Environment (development/production)")
	redisAddr := flag.String("redis", getEnvOrDefault("REDIS_ADDR", "localhost:6379"), "Redis server address")
	flag.Parse()

	return &Config{
		Port:      *port,
		OutputDir: *outputDir,
		Env:       *env,
		RedisAddr: *redisAddr,
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
