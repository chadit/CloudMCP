package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	dirPermissions = 0o755 // Read/write/execute for owner, read/execute for group and others
)

// Windows: %APPDATA%\CloudMCP\.
func getConfigDir() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "CloudMCP")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "CloudMCP")
	default:
		if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
			return filepath.Join(xdgConfig, "cloudmcp")
		}

		return filepath.Join(os.Getenv("HOME"), ".config", "cloudmcp")
	}
}

// Windows: %APPDATA%\CloudMCP\logs\.
func getLogDir() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "CloudMCP", "logs")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "CloudMCP", "logs")
	default:
		if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
			return filepath.Join(xdgData, "CloudMCP")
		}

		return filepath.Join(os.Getenv("HOME"), ".local", "share", "CloudMCP")
	}
}

// GetConfigPath returns the full path to the configuration file.
func GetConfigPath() string {
	return filepath.Join(getConfigDir(), "config.toml")
}

// GetLogPath returns the full path to the log file.
func GetLogPath() string {
	return filepath.Join(getLogDir(), "cloudmcp.log")
}

// EnsureConfigDir creates the configuration directory if it doesn't exist.
func EnsureConfigDir() error {
	if err := os.MkdirAll(getConfigDir(), dirPermissions); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return nil
}

// EnsureLogDir creates the log directory if it doesn't exist.
func EnsureLogDir() error {
	if err := os.MkdirAll(getLogDir(), dirPermissions); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	return nil
}
