package config

import (
	"fmt"
	"sync"
)

// Manager provides thread-safe management of TOML configuration.
type Manager interface {
	GetConfig() *TOMLConfig
	AddAccount(name string, account AccountConfig) error
	RemoveAccount(name string) error
	UpdateAccount(name string, account AccountConfig) error
	SetDefaultAccount(name string) error
	Save() error
	Reload() error
}

// TOMLConfigManager implements Manager for TOML-based configuration.
type TOMLConfigManager struct {
	configPath string
	config     *TOMLConfig
	mutex      sync.RWMutex
}

// NewTOMLConfigManager creates a new TOML configuration manager.
func NewTOMLConfigManager(configPath string) *TOMLConfigManager {
	return &TOMLConfigManager{
		configPath: configPath,
	}
}

// LoadOrCreate loads existing configuration or creates a default one.
func (cm *TOMLConfigManager) LoadOrCreate() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Try to load existing configuration
	config, err := LoadTOMLConfig(cm.configPath)
	if err != nil {
		// Create default configuration
		config = CreateDefaultTOMLConfig()
		
		// Save default configuration
		if saveErr := SaveTOMLConfig(config, cm.configPath); saveErr != nil {
			return fmt.Errorf("failed to create default config: %w", saveErr)
		}
	}

	cm.config = config
	return nil
}

// GetConfig returns a copy of the current configuration.
func (cm *TOMLConfigManager) GetConfig() *TOMLConfig {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	if cm.config == nil {
		return nil
	}

	// Return a deep copy to prevent external modifications
	configCopy := &TOMLConfig{
		System:   cm.config.System,
		Accounts: make(map[string]AccountConfig),
	}

	for name, account := range cm.config.Accounts {
		configCopy.Accounts[name] = account
	}

	return configCopy
}

// AddAccount adds a new account to the configuration.
func (cm *TOMLConfigManager) AddAccount(name string, account AccountConfig) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if cm.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// Validate account
	if account.Token == "" {
		return fmt.Errorf("account token is required")
	}
	if account.Label == "" {
		return fmt.Errorf("account label is required")
	}

	// Add account
	cm.config.Accounts[name] = account

	// Save configuration
	return cm.save()
}

// RemoveAccount removes an account from the configuration.
func (cm *TOMLConfigManager) RemoveAccount(name string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if cm.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// Check if account exists
	if _, exists := cm.config.Accounts[name]; !exists {
		return fmt.Errorf("account %q does not exist", name)
	}

	// Prevent removing the default account
	if name == cm.config.System.DefaultAccount {
		return fmt.Errorf("cannot remove the default account %q", name)
	}

	// Remove account
	delete(cm.config.Accounts, name)

	// Save configuration
	return cm.save()
}

// UpdateAccount updates an existing account.
func (cm *TOMLConfigManager) UpdateAccount(name string, account AccountConfig) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if cm.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// Check if account exists
	if _, exists := cm.config.Accounts[name]; !exists {
		return fmt.Errorf("account %q does not exist", name)
	}

	// Validate account
	if account.Token == "" {
		return fmt.Errorf("account token is required")
	}
	if account.Label == "" {
		return fmt.Errorf("account label is required")
	}

	// Update account
	cm.config.Accounts[name] = account

	// Save configuration
	return cm.save()
}

// SetDefaultAccount sets the default account.
func (cm *TOMLConfigManager) SetDefaultAccount(name string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if cm.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// Check if account exists
	if _, exists := cm.config.Accounts[name]; !exists {
		return fmt.Errorf("account %q does not exist", name)
	}

	// Update default account
	cm.config.System.DefaultAccount = name

	// Save configuration
	return cm.save()
}

// Save saves the current configuration to disk.
func (cm *TOMLConfigManager) Save() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	return cm.save()
}

// save is the internal save method (assumes lock is held).
func (cm *TOMLConfigManager) save() error {
	if cm.config == nil {
		return fmt.Errorf("no configuration to save")
	}

	return SaveTOMLConfig(cm.config, cm.configPath)
}

// Reload reloads the configuration from disk.
func (cm *TOMLConfigManager) Reload() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	config, err := LoadTOMLConfig(cm.configPath)
	if err != nil {
		return fmt.Errorf("failed to reload configuration: %w", err)
	}

	cm.config = config
	return nil
}