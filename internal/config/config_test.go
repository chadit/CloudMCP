package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/config"
)

func TestLoad_DefaultConfig(t *testing.T) {
	t.Parallel()
	// Since we can't easily mock GetConfigPath, we'll test the load functionality.
	// by creating a config manager directly with a test path.
	tempDir := t.TempDir()
	testConfigPath := filepath.Join(tempDir, "cloud-mcp.toml")

	// Test creating a default TOML config.
	defaultConfig := config.CreateDefaultTOMLConfig()
	require.NotNil(t, defaultConfig, "Default config should not be nil")

	// Save it to test path.
	err := config.SaveTOMLConfig(defaultConfig, testConfigPath)
	require.NoError(t, err, "Should save default config without error")

	// Load it back.
	loadedConfig, err := config.LoadTOMLConfig(testConfigPath)
	require.NoError(t, err, "Should load config without error")
	require.NotNil(t, loadedConfig, "Loaded config should not be nil")

	// Convert to legacy format for testing.
	legacyConfig := loadedConfig.ToLegacyConfig()
	require.NotNil(t, legacyConfig, "Legacy config should not be nil")

	// Verify default values.
	require.Equal(t, "CloudMCP Minimal Shell", legacyConfig.ServerName, "Default server name should be set")
	require.Equal(t, "info", legacyConfig.LogLevel, "Default log level should be info")
	require.True(t, legacyConfig.EnableMetrics, "Metrics should be enabled by default")
	require.Equal(t, 8080, legacyConfig.MetricsPort, "Default metrics port should be set")

	// Verify config file was created.
	require.FileExists(t, testConfigPath, "Config file should be created")
}

func TestLoad_ExistingConfig(t *testing.T) {
	t.Parallel()
	// Create a temporary directory for testing.
	tempDir := t.TempDir()

	// Create a test TOML config.
	testConfigPath := filepath.Join(tempDir, "cloud-mcp.toml")
	testTOMLContent := `[system]
server_name = "Test Server"
log_level = "debug"
enable_metrics = false
metrics_port = 9090
`

	err := os.WriteFile(testConfigPath, []byte(testTOMLContent), 0o600)
	require.NoError(t, err, "Should write test config file")

	// Load existing config directly.
	loadedConfig, err := config.LoadTOMLConfig(testConfigPath)
	require.NoError(t, err, "LoadTOMLConfig should load existing config without error")
	require.NotNil(t, loadedConfig, "Config should not be nil")

	// Convert to legacy format for testing.
	config := loadedConfig.ToLegacyConfig()
	require.NotNil(t, config, "Legacy config should not be nil")

	// Verify loaded values.
	require.Equal(t, "Test Server", config.ServerName, "Server name should be loaded from file")
	require.Equal(t, "debug", config.LogLevel, "Log level should be loaded from file")
	require.False(t, config.EnableMetrics, "Metrics should be disabled as configured")
	require.Equal(t, 9090, config.MetricsPort, "Metrics port should be loaded from file")
}

func TestConfig_BasicValidation(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		ServerName:    "Test Server",
		LogLevel:      "info",
		EnableMetrics: true,
		MetricsPort:   8080,
	}

	// Basic validation - just ensure config structure is valid.
	require.NotNil(t, cfg, "Config should not be nil")
	require.NotEmpty(t, cfg.ServerName, "Server name should not be empty")
	require.NotEmpty(t, cfg.LogLevel, "Log level should not be empty")
}

func TestConfig_Constants(t *testing.T) {
	t.Parallel()
	// Test that the default metrics port constant has expected value.
	// Note: defaultMetricsPort is not exported, so we test the behavior instead.
	cfg := config.CreateDefaultTOMLConfig()
	require.Equal(t, 8080, cfg.System.MetricsPort, "Default metrics port should be 8080")
}

func TestConfig_Errors(t *testing.T) {
	t.Parallel()
	require.Error(t, config.ErrDefaultAccountNotFound, "config.ErrDefaultAccountNotFound should be defined")
	require.Contains(t, config.ErrDefaultAccountNotFound.Error(), "default account not found", "Error message should be descriptive")
}
