package config

import (
	"fmt"
	"os"
	"strings"
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

func Load() (*Config, error) {
	cfg := &Config{
		ServerName:     getEnv("CLOUD_MCP_SERVER_NAME", "Cloud MCP Server"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		EnableMetrics:  getEnvBool("ENABLE_METRICS", true),
		MetricsPort:    getEnvInt("METRICS_PORT", 8080),
		LinodeAccounts: make(map[string]LinodeAccount),
	}

	cfg.DefaultLinodeAccount = getEnv("DEFAULT_LINODE_ACCOUNT", "primary")

	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "LINODE_ACCOUNTS_") {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) != 2 {
				continue
			}

			keyParts := strings.Split(parts[0], "_")
			if len(keyParts) < 4 {
				continue
			}

			accountName := strings.ToLower(keyParts[2])
			fieldType := strings.ToLower(keyParts[3])

			if _, exists := cfg.LinodeAccounts[accountName]; !exists {
				cfg.LinodeAccounts[accountName] = LinodeAccount{}
			}

			account := cfg.LinodeAccounts[accountName]
			switch fieldType {
			case "token":
				account.Token = parts[1]
			case "label":
				account.Label = parts[1]
			case "apiurl":
				account.APIURL = parts[1]
			}
			cfg.LinodeAccounts[accountName] = account
		}
	}

	if len(cfg.LinodeAccounts) == 0 {
		return nil, fmt.Errorf("no Linode accounts configured")
	}

	if _, exists := cfg.LinodeAccounts[cfg.DefaultLinodeAccount]; !exists {
		return nil, fmt.Errorf("default account %q not found in configured accounts", cfg.DefaultLinodeAccount)
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return strings.ToLower(value) == "true" || value == "1"
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	var result int
	if _, err := fmt.Sscanf(value, "%d", &result); err != nil {
		return defaultValue
	}
	return result
}
