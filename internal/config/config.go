// Package config provides simple environment-based configuration for CloudMCP.
package config

import (
	"os"
)

// Config holds the minimal configuration for CloudMCP server.
type Config struct {
	ServerName string
	LogLevel   string
}

// Load loads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	return &Config{
		ServerName: getEnvOrDefault("CLOUD_MCP_SERVER_NAME", "CloudMCP Minimal"),
		LogLevel:   getEnvOrDefault("LOG_LEVEL", "info"),
	}, nil
}

// getEnvOrDefault returns environment variable value or default if not set.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}