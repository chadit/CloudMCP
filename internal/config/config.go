// Package config provides comprehensive configuration management for the CloudMCP server,
// supporting both TOML file-based configuration and environment variable overrides.
//
// The package implements a modern configuration system with the following key features:
//
// Configuration Sources and Priority:
//   - TOML configuration files with validation and security checks
//   - Environment variable overrides for deployment flexibility
//   - Automatic default configuration creation for first-time setup
//   - Cross-platform directory support (Windows, macOS, Linux/Unix)
//
// Key Types:
//   - Config: Legacy configuration structure for backward compatibility
//   - TOMLConfig: Modern TOML-based configuration with nested sections
//   - SystemConfig: System-wide settings including logging and metrics
//
// Configuration Loading Process:
//  1. Attempts to load TOML configuration from platform-specific config directory
//  2. Creates default TOML configuration if none exists
//  3. Applies environment variable overrides to loaded configuration
//  4. Validates configuration before returning
//
// Environment Variable Overrides:
//   - ENABLE_METRICS: boolean flag to enable/disable metrics collection
//   - METRICS_PORT: port number for metrics HTTP server
//   - LOG_LEVEL: logging level (debug, info, warn, error)
//   - SERVER_NAME: custom server name for identification
//
// Cross-Platform Directory Structure:
//   - Windows: %APPDATA%\CloudMCP\config.toml
//   - macOS: ~/Library/Application Support/CloudMCP/config.toml
//   - Linux/Unix: ~/.config/cloudmcp/config.toml (XDG Base Directory compliant)
//
// Security Features:
//   - Path traversal protection with input sanitization
//   - Config directory access restrictions
//   - Secure file permissions (0o750 for directories)
//   - Validation of configuration values before use
//
// Usage Example:
//
//	cfg, err := config.Load()
//	if err != nil {
//	    log.Fatalf("Failed to load configuration: %v", err)
//	}
//
//	// Configuration is ready for use
//	fmt.Printf("Server: %s, Port: %d\n", cfg.ServerName, cfg.MetricsPort)
//
// The package automatically handles first-time setup by creating default configuration
// files and directories with appropriate permissions. It provides a seamless migration
// path from legacy configurations while maintaining backward compatibility.
package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
)

const (
	defaultMetricsPort = 8080
)

var ErrDefaultAccountNotFound = errors.New("default account not found in configured accounts")

type Config struct {
	ServerName    string
	LogLevel      string
	EnableMetrics bool
	MetricsPort   int
}

// Load loads configuration from TOML file, creating a default config if none exists.
func Load() (*Config, error) {
	configPath := GetConfigPath()

	// Try to load TOML config.
	if tomlConfig, err := LoadTOMLConfig(configPath); err == nil {
		cfg := tomlConfig.ToLegacyConfig()
		// Override with environment variables if set.
		applyEnvironmentOverrides(cfg)

		return cfg, nil
	}

	// Create default config if none exists.
	defaultConfig := CreateDefaultTOMLConfig()

	if err := SaveTOMLConfig(defaultConfig, configPath); err != nil {
		return nil, fmt.Errorf("failed to create default config: %w", err)
	}

	log.Printf("Created default TOML configuration at: %s", configPath)
	log.Printf("Default configuration created - ready for customization")

	cfg := defaultConfig.ToLegacyConfig()
	// Override with environment variables if set.
	applyEnvironmentOverrides(cfg)

	return cfg, nil
}

// applyEnvironmentOverrides applies environment variable overrides to the configuration.
// This allows environment variables to override TOML configuration values.
func applyEnvironmentOverrides(cfg *Config) {
	// Override settings that might be useful for testing/deployment.
	if envVal := os.Getenv("ENABLE_METRICS"); envVal != "" {
		if parsed, err := strconv.ParseBool(envVal); err == nil {
			cfg.EnableMetrics = parsed
		}
	}

	if envVal := os.Getenv("METRICS_PORT"); envVal != "" {
		if parsed, err := strconv.Atoi(envVal); err == nil {
			cfg.MetricsPort = parsed
		}
	}

	if envVal := os.Getenv("LOG_LEVEL"); envVal != "" {
		cfg.LogLevel = envVal
	}

	if envVal := os.Getenv("SERVER_NAME"); envVal != "" {
		cfg.ServerName = envVal
	}
}
