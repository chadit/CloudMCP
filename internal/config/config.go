package config

import (
	"errors"
	"fmt"
	"log"
)

const (
	defaultMetricsPort = 8080
)

var (
	ErrNoLinodeAccounts       = errors.New("no Linode accounts configured")
	ErrDefaultAccountNotFound = errors.New("default account not found in configured accounts")
)

type Config struct {
	ServerName           string
	LogLevel             string
	EnableMetrics        bool
	MetricsPort          int
	DefaultLinodeAccount string
	LinodeAccounts       map[string]LinodeAccount
}

type LinodeAccount struct {
	Token  string
	Label  string
	APIURL string // Optional custom API URL (defaults to https://api.linode.com/v4)
}

// Load loads configuration from TOML file, creating a default config if none exists.
func Load() (*Config, error) {
	configPath := GetConfigPath()

	// Try to load TOML config
	if tomlConfig, err := LoadTOMLConfig(configPath); err == nil {
		return tomlConfig.ToLegacyConfig(), nil
	}

	// Create default config if none exists
	defaultConfig := CreateDefaultTOMLConfig()
	if err := SaveTOMLConfig(defaultConfig, configPath); err != nil {
		return nil, fmt.Errorf("failed to create default config: %w", err)
	}

	log.Printf("Created default TOML configuration at: %s", configPath)
	log.Printf("Please edit the configuration file to add your Linode API tokens")

	return defaultConfig.ToLegacyConfig(), nil
}



