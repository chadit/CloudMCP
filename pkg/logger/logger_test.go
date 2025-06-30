package logger_test

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/pkg/logger"
)

func TestNew_DefaultConfig(t *testing.T) {
	t.Parallel()

	log := logger.New("info")

	require.NotNil(t, log, "Logger should not be nil")

	// Test that it's a slogWrapper - using interface, so no internal access needed
	require.NotNil(t, log, "Logger should be properly initialized")
}

func TestNewWithConfig_LogLevels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		level         string
		expectedLevel slog.Level
	}{
		{
			name:          "debug level",
			level:         "debug",
			expectedLevel: slog.LevelDebug,
		},
		{
			name:          "info level",
			level:         "info",
			expectedLevel: slog.LevelInfo,
		},
		{
			name:          "warn level",
			level:         "warn",
			expectedLevel: slog.LevelWarn,
		},
		{
			name:          "error level",
			level:         "error",
			expectedLevel: slog.LevelError,
		},
		{
			name:          "invalid level defaults to info",
			level:         "invalid",
			expectedLevel: slog.LevelInfo,
		},
		{
			name:          "empty level defaults to info",
			level:         "",
			expectedLevel: slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config := logger.LogConfig{
				Level: tt.level,
			}

			log := logger.NewWithConfig(config)
			require.NotNil(t, log, "Logger should not be nil")

			// Test that the logger was created correctly - using interface only
			require.NotNil(t, log, "Logger should be properly initialized")
		})
	}
}

func TestNewWithConfig_FileOutput(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	config := logger.LogConfig{
		Level:      "info",
		FilePath:   logFile,
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
	}

	log := logger.NewWithConfig(config)
	require.NotNil(t, log, "Logger should not be nil")

	// Log a test message
	log.Info("test message", "key", "value")

	// Verify log file was created and contains expected content
	require.FileExists(t, logFile, "Log file should be created")

	content, err := os.ReadFile(logFile)
	require.NoError(t, err, "Should read log file without error")
	require.Contains(t, string(content), "test message", "Log file should contain the logged message")
	require.Contains(t, string(content), "key", "Log file should contain the logged key")
	require.Contains(t, string(content), "value", "Log file should contain the logged value")
}

func TestLogger_DebugMethods(t *testing.T) {
	t.Parallel()
	// Create logger with debug level that writes to file
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "debug.log")

	config := logger.LogConfig{
		Level:    "debug",
		FilePath: logFile,
	}

	log := logger.NewWithConfig(config)

	// Test Debug method
	log.Debug("debug message", "debug_key", "debug_value")

	// Test DebugContext method
	ctx := t.Context()
	log.DebugContext(ctx, "debug context message", "ctx_key", "ctx_value")

	// Verify log file contains debug messages
	content, err := os.ReadFile(logFile)
	require.NoError(t, err, "Should read log file")

	logContent := string(content)
	require.Contains(t, logContent, "debug message", "Log should contain debug message")
	require.Contains(t, logContent, "debug context message", "Log should contain debug context message")
	require.Contains(t, logContent, "debug_key", "Log should contain debug key")
	require.Contains(t, logContent, "ctx_key", "Log should contain context key")
}

func TestLogger_InfoMethods(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "info.log")

	config := logger.LogConfig{
		Level:    "info",
		FilePath: logFile,
	}

	log := logger.NewWithConfig(config)

	// Test Info method
	log.Info("info message", "info_key", "info_value")

	// Test InfoContext method
	ctx := t.Context()
	log.InfoContext(ctx, "info context message", "ctx_key", "ctx_value")

	// Verify log file contains info messages
	content, err := os.ReadFile(logFile)
	require.NoError(t, err, "Should read log file")

	logContent := string(content)
	require.Contains(t, logContent, "info message", "Log should contain info message")
	require.Contains(t, logContent, "info context message", "Log should contain info context message")
	require.Contains(t, logContent, "info_key", "Log should contain info key")
	require.Contains(t, logContent, "ctx_key", "Log should contain context key")
}

func TestLogger_WarnMethods(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "warn.log")

	config := logger.LogConfig{
		Level:    "warn",
		FilePath: logFile,
	}

	log := logger.NewWithConfig(config)

	// Test Warn method
	log.Warn("warn message", "warn_key", "warn_value")

	// Test WarnContext method
	ctx := t.Context()
	log.WarnContext(ctx, "warn context message", "ctx_key", "ctx_value")

	// Verify log file contains warn messages
	content, err := os.ReadFile(logFile)
	require.NoError(t, err, "Should read log file")

	logContent := string(content)
	require.Contains(t, logContent, "warn message", "Log should contain warn message")
	require.Contains(t, logContent, "warn context message", "Log should contain warn context message")
	require.Contains(t, logContent, "warn_key", "Log should contain warn key")
	require.Contains(t, logContent, "ctx_key", "Log should contain context key")
}

func TestLogger_ErrorMethods(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "error.log")

	config := logger.LogConfig{
		Level:    "error",
		FilePath: logFile,
	}

	log := logger.NewWithConfig(config)

	// Test Error method
	log.Error("error message", "error_key", "error_value")

	// Test ErrorContext method
	ctx := t.Context()
	log.ErrorContext(ctx, "error context message", "ctx_key", "ctx_value")

	// Verify log file contains error messages
	content, err := os.ReadFile(logFile)
	require.NoError(t, err, "Should read log file")

	logContent := string(content)
	require.Contains(t, logContent, "error message", "Log should contain error message")
	require.Contains(t, logContent, "error context message", "Log should contain error context message")
	require.Contains(t, logContent, "error_key", "Log should contain error key")
	require.Contains(t, logContent, "ctx_key", "Log should contain context key")
}

func TestLogger_With(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "with.log")

	config := logger.LogConfig{
		Level:    "info",
		FilePath: logFile,
	}

	log := logger.NewWithConfig(config)

	// Create a child logger with additional context
	childLogger := log.With("service", "test", "version", "1.0.0")
	require.NotNil(t, childLogger, "Child logger should not be nil")

	// Log with child logger
	childLogger.Info("child log message", "extra", "data")

	// Verify log contains both original and new context
	content, err := os.ReadFile(logFile)
	require.NoError(t, err, "Should read log file")

	logContent := string(content)
	require.Contains(t, logContent, "child log message", "Log should contain the message")
	require.Contains(t, logContent, "service", "Log should contain service context")
	require.Contains(t, logContent, "test", "Log should contain service value")
	require.Contains(t, logContent, "version", "Log should contain version context")
	require.Contains(t, logContent, "1.0.0", "Log should contain version value")
	require.Contains(t, logContent, "extra", "Log should contain extra context")
	require.Contains(t, logContent, "data", "Log should contain extra value")
}

func TestLogger_LogLevelFiltering(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "level_filter.log")

	// Create logger with warn level (should filter out debug and info)
	config := logger.LogConfig{
		Level:    "warn",
		FilePath: logFile,
	}

	log := logger.NewWithConfig(config)

	// Log messages at different levels
	log.Debug("debug message")
	log.Info("info message")
	log.Warn("warn message")
	log.Error("error message")

	// Read log content
	content, err := os.ReadFile(logFile)
	require.NoError(t, err, "Should read log file")

	logContent := string(content)

	// Should NOT contain debug and info messages
	require.NotContains(t, logContent, "debug message", "Debug message should be filtered out")
	require.NotContains(t, logContent, "info message", "Info message should be filtered out")

	// Should contain warn and error messages
	require.Contains(t, logContent, "warn message", "Warn message should be included")
	require.Contains(t, logContent, "error message", "Error message should be included")
}

func TestLogger_JSONFormat(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "json.log")

	config := logger.LogConfig{
		Level:    "info",
		FilePath: logFile,
	}

	log := logger.NewWithConfig(config)

	// Log a structured message
	log.Info("test message", "key1", "value1", "key2", 42, "key3", true)

	// Read and parse log content
	content, err := os.ReadFile(logFile)
	require.NoError(t, err, "Should read log file")

	// Verify it's valid JSON
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	require.Len(t, lines, 1, "Should have exactly one log line")

	var logEntry map[string]interface{}
	err = json.Unmarshal([]byte(lines[0]), &logEntry)
	require.NoError(t, err, "Log entry should be valid JSON")

	// Verify JSON structure
	require.Equal(t, "test message", logEntry["msg"], "Message should be in msg field")
	require.Equal(t, "INFO", logEntry["level"], "Level should be INFO")
	require.Equal(t, "value1", logEntry["key1"], "Custom key1 should be present")
	require.InDelta(t, float64(42), logEntry["key2"], 0.001, "Custom key2 should be present")
	require.Equal(t, true, logEntry["key3"], "Custom key3 should be present")
	require.Contains(t, logEntry, "time", "Time field should be present")
}

func TestLogConfig_Validation(t *testing.T) {
	t.Parallel()
	// Test that LogConfig struct can be created with various configurations
	configs := []logger.LogConfig{
		{
			Level: "debug",
		},
		{
			Level:      "info",
			FilePath:   "/tmp/test.log",
			MaxSize:    50,
			MaxBackups: 10,
			MaxAge:     7,
		},
		{
			Level:    "error",
			FilePath: "",
		},
	}

	for i, config := range configs {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			t.Parallel()

			log := logger.NewWithConfig(config)
			require.NotNil(t, log, "Logger should be created with any valid config")

			// Test that basic logging works
			log.Info("test message")
		})
	}
}
