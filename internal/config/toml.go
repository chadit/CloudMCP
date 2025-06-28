package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// TOMLConfig represents the new TOML-based configuration structure.
type TOMLConfig struct {
	System   SystemConfig             `toml:"system"`
	Accounts map[string]AccountConfig `toml:"account"`
}

// SystemConfig contains system-wide configuration settings.
type SystemConfig struct {
	ServerName     string `toml:"server_name"`
	LogLevel       string `toml:"log_level"`
	EnableMetrics  bool   `toml:"enable_metrics"`
	MetricsPort    int    `toml:"metrics_port"`
	DefaultAccount string `toml:"default_account"`

	// Logging configuration
	LogFile       string `toml:"log_file"`        // Empty = use default path
	LogMaxSize    int    `toml:"log_max_size"`    // MB
	LogMaxBackups int    `toml:"log_max_backups"` // Number of files to keep
	LogMaxAge     int    `toml:"log_max_age"`     // Days to keep logs
}

// AccountConfig represents a Linode account configuration.
type AccountConfig struct {
	Token  string `toml:"token"`
	Label  string `toml:"label"`
	APIURL string `toml:"apiurl,omitempty"` // Optional custom API URL
}

// ToLegacyConfig converts TOMLConfig to the legacy Config structure.
func (tc *TOMLConfig) ToLegacyConfig() *Config {
	cfg := &Config{
		ServerName:           tc.System.ServerName,
		LogLevel:             tc.System.LogLevel,
		EnableMetrics:        tc.System.EnableMetrics,
		MetricsPort:          tc.System.MetricsPort,
		DefaultLinodeAccount: tc.System.DefaultAccount,
		LinodeAccounts:       make(map[string]LinodeAccount),
	}

	// Convert accounts
	for name, account := range tc.Accounts {
		cfg.LinodeAccounts[name] = LinodeAccount{
			Token:  account.Token,
			Label:  account.Label,
			APIURL: account.APIURL,
		}
	}

	return cfg
}

// LoadTOMLConfig loads configuration from a TOML file.
func LoadTOMLConfig(configPath string) (*TOMLConfig, error) {
	var config TOMLConfig

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	// Decode TOML file
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, fmt.Errorf("failed to parse TOML config: %w", err)
	}

	// Set defaults if not specified
	if config.System.ServerName == "" {
		config.System.ServerName = "Cloud MCP Server"
	}
	if config.System.LogLevel == "" {
		config.System.LogLevel = "info"
	}
	if config.System.MetricsPort == 0 {
		config.System.MetricsPort = defaultMetricsPort
	}
	if config.System.LogMaxSize == 0 {
		config.System.LogMaxSize = 10 // 10MB
	}
	if config.System.LogMaxBackups == 0 {
		config.System.LogMaxBackups = 5
	}
	if config.System.LogMaxAge == 0 {
		config.System.LogMaxAge = 30 // 30 days
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// SaveTOMLConfig saves the configuration to a TOML file.
func SaveTOMLConfig(config *TOMLConfig, configPath string) error {
	// Ensure directory exists
	if err := EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create or truncate file
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	// Encode to TOML
	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to encode TOML config: %w", err)
	}

	return nil
}

// Validate checks if the TOML configuration is valid.
func (tc *TOMLConfig) Validate() error {
	if len(tc.Accounts) == 0 {
		return ErrNoLinodeAccounts
	}

	if tc.System.DefaultAccount == "" {
		return fmt.Errorf("default_account is required")
	}

	if _, exists := tc.Accounts[tc.System.DefaultAccount]; !exists {
		return fmt.Errorf("%w: %q", ErrDefaultAccountNotFound, tc.System.DefaultAccount)
	}

	// Validate accounts
	for name, account := range tc.Accounts {
		if account.Token == "" {
			return fmt.Errorf("account %q is missing token", name)
		}
		if account.Label == "" {
			return fmt.Errorf("account %q is missing label", name)
		}
	}

	return nil
}

// CreateDefaultTOMLConfig creates a default TOML configuration.
func CreateDefaultTOMLConfig() *TOMLConfig {
	return &TOMLConfig{
		System: SystemConfig{
			ServerName:     "Cloud MCP Server",
			LogLevel:       "info",
			EnableMetrics:  true,
			MetricsPort:    defaultMetricsPort,
			DefaultAccount: "primary",
			LogFile:        "", // Use default path
			LogMaxSize:     10,
			LogMaxBackups:  5,
			LogMaxAge:      30,
		},
		Accounts: map[string]AccountConfig{
			"primary": {
				Token: "your_linode_token_here",
				Label: "Primary Account",
			},
		},
	}
}