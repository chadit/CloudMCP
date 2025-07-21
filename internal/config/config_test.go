package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/config"
)

func TestLoad_DefaultValues(t *testing.T) {
	t.Parallel()
	
	// Clear environment variables to test defaults
	originalServerName := os.Getenv("CLOUD_MCP_SERVER_NAME")
	originalLogLevel := os.Getenv("LOG_LEVEL")
	defer func() {
		os.Setenv("CLOUD_MCP_SERVER_NAME", originalServerName)
		os.Setenv("LOG_LEVEL", originalLogLevel)
	}()
	
	os.Unsetenv("CLOUD_MCP_SERVER_NAME")
	os.Unsetenv("LOG_LEVEL")

	// Load config with default values
	cfg, err := config.Load()
	require.NoError(t, err, "Should load config without error")
	require.NotNil(t, cfg, "Config should not be nil")

	// Verify default values
	require.Equal(t, "CloudMCP Minimal", cfg.ServerName, "Default server name should be set")
	require.Equal(t, "info", cfg.LogLevel, "Default log level should be info")
}

func TestLoad_EnvironmentVariables(t *testing.T) {
	t.Parallel()
	
	// Save original environment variables
	originalServerName := os.Getenv("CLOUD_MCP_SERVER_NAME")
	originalLogLevel := os.Getenv("LOG_LEVEL")
	defer func() {
		os.Setenv("CLOUD_MCP_SERVER_NAME", originalServerName)
		os.Setenv("LOG_LEVEL", originalLogLevel)
	}()

	// Set test environment variables
	os.Setenv("CLOUD_MCP_SERVER_NAME", "Test Server")
	os.Setenv("LOG_LEVEL", "debug")

	// Load config
	cfg, err := config.Load()
	require.NoError(t, err, "Should load config without error")
	require.NotNil(t, cfg, "Config should not be nil")

	// Verify environment values are used
	require.Equal(t, "Test Server", cfg.ServerName, "Server name should be loaded from environment")
	require.Equal(t, "debug", cfg.LogLevel, "Log level should be loaded from environment")
}

func TestLoad_PartialEnvironmentOverride(t *testing.T) {
	t.Parallel()
	
	// Save original environment variables
	originalServerName := os.Getenv("CLOUD_MCP_SERVER_NAME")
	originalLogLevel := os.Getenv("LOG_LEVEL")
	defer func() {
		os.Setenv("CLOUD_MCP_SERVER_NAME", originalServerName)
		os.Setenv("LOG_LEVEL", originalLogLevel)
	}()

	// Set only server name, leave log level as default
	os.Setenv("CLOUD_MCP_SERVER_NAME", "Custom Server")
	os.Unsetenv("LOG_LEVEL")

	// Load config
	cfg, err := config.Load()
	require.NoError(t, err, "Should load config without error")
	require.NotNil(t, cfg, "Config should not be nil")

	// Verify mixed values (environment + default)
	require.Equal(t, "Custom Server", cfg.ServerName, "Server name should be from environment")
	require.Equal(t, "info", cfg.LogLevel, "Log level should use default value")
}

func TestConfig_BasicValidation(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		ServerName: "Test Server",
		LogLevel:   "info",
	}

	// Basic validation - ensure config structure is valid
	require.NotNil(t, cfg, "Config should not be nil")
	require.NotEmpty(t, cfg.ServerName, "Server name should not be empty")
	require.NotEmpty(t, cfg.LogLevel, "Log level should not be empty")
}

func TestLoad_EmptyEnvironmentValues(t *testing.T) {
	t.Parallel()
	
	// Save original environment variables
	originalServerName := os.Getenv("CLOUD_MCP_SERVER_NAME")
	originalLogLevel := os.Getenv("LOG_LEVEL")
	defer func() {
		os.Setenv("CLOUD_MCP_SERVER_NAME", originalServerName)
		os.Setenv("LOG_LEVEL", originalLogLevel)
	}()

	// Set empty environment variables (should use defaults)
	os.Setenv("CLOUD_MCP_SERVER_NAME", "")
	os.Setenv("LOG_LEVEL", "")

	// Load config
	cfg, err := config.Load()
	require.NoError(t, err, "Should load config without error")
	require.NotNil(t, cfg, "Config should not be nil")

	// Verify defaults are used when env vars are empty
	require.Equal(t, "CloudMCP Minimal", cfg.ServerName, "Should use default when env var is empty")
	require.Equal(t, "info", cfg.LogLevel, "Should use default when env var is empty")
}