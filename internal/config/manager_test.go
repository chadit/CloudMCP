package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	configpkg "github.com/chadit/CloudMCP/internal/config"
)

func TestTOMLConfigManager_LoadOrCreate(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.toml")

	manager := configpkg.NewTOMLConfigManager(configPath)
	require.NotNil(t, manager, "Manager should not be nil")

	// Test loading when no config exists (should create default)
	err := manager.LoadOrCreate()
	require.NoError(t, err, "LoadOrCreate should not error when creating default config")

	config := manager.GetConfig()
	require.NotNil(t, config, "Config should not be nil after LoadOrCreate")
	require.FileExists(t, configPath, "Config file should be created")

	// Test loading existing config
	manager2 := configpkg.NewTOMLConfigManager(configPath)
	err = manager2.LoadOrCreate()
	require.NoError(t, err, "LoadOrCreate should not error when loading existing config")

	config2 := manager2.GetConfig()
	require.NotNil(t, config2, "Config should not be nil when loading existing")
	require.Equal(t, config.System.ServerName, config2.System.ServerName, "Loaded config should match saved config")
}

func TestTOMLConfigManager_GetConfig(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.toml")

	manager := configpkg.NewTOMLConfigManager(configPath)

	// Test GetConfig before loading (should return nil)
	config := manager.GetConfig()
	require.Nil(t, config, "GetConfig should return nil before loading")

	// Load config and test GetConfig
	err := manager.LoadOrCreate()
	require.NoError(t, err, "LoadOrCreate should not error")

	config = manager.GetConfig()
	require.NotNil(t, config, "GetConfig should return config after loading")

	// Test that returned config is a copy (modifying it doesn't affect internal config)
	originalServerName := config.System.ServerName
	config.System.ServerName = "Modified Name"

	config2 := manager.GetConfig()
	require.Equal(t, originalServerName, config2.System.ServerName, "Internal config should not be modified by external changes")
}

func TestTOMLConfigManager_AddAccount(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.toml")

	manager := configpkg.NewTOMLConfigManager(configPath)
	err := manager.LoadOrCreate()
	require.NoError(t, err, "LoadOrCreate should not error")

	// Test adding valid account
	testAccount := configpkg.AccountConfig{
		Token: "test_token_123",
		Label: "Test Account",
	}

	err = manager.AddAccount("test", testAccount)
	require.NoError(t, err, "AddAccount should not error with valid account")

	config := manager.GetConfig()
	require.Contains(t, config.Accounts, "test", "Account should be added to config")
	require.Equal(t, testAccount.Token, config.Accounts["test"].Token, "Account token should match")
	require.Equal(t, testAccount.Label, config.Accounts["test"].Label, "Account label should match")

	// Test adding account with missing token
	invalidAccount := configpkg.AccountConfig{
		Label: "Invalid Account",
	}

	err = manager.AddAccount("invalid", invalidAccount)
	require.ErrorIs(t, err, configpkg.ErrAccountTokenRequired, "AddAccount should error when token is missing")

	// Test adding account with missing label
	invalidAccount2 := configpkg.AccountConfig{
		Token: "valid_token",
	}

	err = manager.AddAccount("invalid2", invalidAccount2)
	require.ErrorIs(t, err, configpkg.ErrAccountLabelRequired, "AddAccount should error when label is missing")
}

func TestTOMLConfigManager_RemoveAccount(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.toml")

	manager := configpkg.NewTOMLConfigManager(configPath)
	err := manager.LoadOrCreate()
	require.NoError(t, err, "LoadOrCreate should not error")

	// Add test accounts
	testAccount := configpkg.AccountConfig{
		Token: "test_token_123",
		Label: "Test Account",
	}

	err = manager.AddAccount("test", testAccount)
	require.NoError(t, err, "AddAccount should not error")

	err = manager.AddAccount("removable", testAccount)
	require.NoError(t, err, "AddAccount should not error")

	// Try to remove non-existent account
	err = manager.RemoveAccount("nonexistent")
	require.ErrorIs(t, err, configpkg.ErrAccountNotExist, "RemoveAccount should error for non-existent account")

	// Try to remove default account (should fail)
	config := manager.GetConfig()
	defaultAccount := config.System.DefaultAccount

	if _, exists := config.Accounts[defaultAccount]; exists {
		err = manager.RemoveAccount(defaultAccount)
		require.ErrorIs(t, err, configpkg.ErrCannotRemoveDefaultAccount, "RemoveAccount should error when trying to remove default account")
	}

	// Remove non-default account (should succeed)
	err = manager.RemoveAccount("removable")
	require.NoError(t, err, "RemoveAccount should not error for valid non-default account")

	config = manager.GetConfig()
	require.NotContains(t, config.Accounts, "removable", "Account should be removed from config")
}

func TestTOMLConfigManager_UpdateAccount(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.toml")

	manager := configpkg.NewTOMLConfigManager(configPath)
	err := manager.LoadOrCreate()
	require.NoError(t, err, "LoadOrCreate should not error")

	// Add initial account
	initialAccount := configpkg.AccountConfig{
		Token: "initial_token",
		Label: "Initial Label",
	}

	err = manager.AddAccount("test", initialAccount)
	require.NoError(t, err, "AddAccount should not error")

	// Update account with valid data
	updatedAccount := configpkg.AccountConfig{
		Token: "updated_token",
		Label: "Updated Label",
	}

	err = manager.UpdateAccount("test", updatedAccount)
	require.NoError(t, err, "UpdateAccount should not error with valid data")

	config := manager.GetConfig()
	require.Equal(t, updatedAccount.Token, config.Accounts["test"].Token, "Account token should be updated")
	require.Equal(t, updatedAccount.Label, config.Accounts["test"].Label, "Account label should be updated")

	// Try to update non-existent account
	err = manager.UpdateAccount("nonexistent", updatedAccount)
	require.ErrorIs(t, err, configpkg.ErrAccountNotExist, "UpdateAccount should error for non-existent account")

	// Try to update with invalid data
	invalidAccount := configpkg.AccountConfig{
		Label: "No Token",
	}

	err = manager.UpdateAccount("test", invalidAccount)
	require.ErrorIs(t, err, configpkg.ErrAccountTokenRequired, "UpdateAccount should error when token is missing")
}

func TestTOMLConfigManager_SetDefaultAccount(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.toml")

	manager := configpkg.NewTOMLConfigManager(configPath)
	err := manager.LoadOrCreate()
	require.NoError(t, err, "LoadOrCreate should not error")

	// Add test account
	testAccount := configpkg.AccountConfig{
		Token: "test_token_123",
		Label: "Test Account",
	}

	err = manager.AddAccount("test", testAccount)
	require.NoError(t, err, "AddAccount should not error")

	// Set default account to existing account
	err = manager.SetDefaultAccount("test")
	require.NoError(t, err, "SetDefaultAccount should not error for existing account")

	config := manager.GetConfig()
	require.Equal(t, "test", config.System.DefaultAccount, "Default account should be updated")

	// Try to set default to non-existent account
	err = manager.SetDefaultAccount("nonexistent")
	require.ErrorIs(t, err, configpkg.ErrAccountNotExist, "SetDefaultAccount should error for non-existent account")
}

func TestTOMLConfigManager_Save(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.toml")

	manager := configpkg.NewTOMLConfigManager(configPath)
	err := manager.LoadOrCreate()
	require.NoError(t, err, "LoadOrCreate should not error")

	// Test explicit save
	err = manager.Save()
	require.NoError(t, err, "Save should not error")

	// Verify file was updated
	require.FileExists(t, configPath, "Config file should exist after save")
}

func TestTOMLConfigManager_Reload(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.toml")

	manager := configpkg.NewTOMLConfigManager(configPath)
	err := manager.LoadOrCreate()
	require.NoError(t, err, "LoadOrCreate should not error")

	// Modify config externally
	externalTOMLContent := `[system]
server_name = "Externally Modified"
log_level = "debug"
enable_metrics = false
metrics_port = 9090
default_account = "primary"

[account.primary]
token = "external_token"
label = "External Account"
`

	err = os.WriteFile(configPath, []byte(externalTOMLContent), 0o600)
	require.NoError(t, err, "Should write external config file")

	// Reload and verify changes
	err = manager.Reload()
	require.NoError(t, err, "Reload should not error")

	config := manager.GetConfig()
	require.Equal(t, "Externally Modified", config.System.ServerName, "Config should reflect external changes")
	require.Equal(t, "debug", config.System.LogLevel, "Config should reflect external changes")
}

func TestTOMLConfigManager_Errors(t *testing.T) {
	t.Parallel()
	// Test all error constants
	require.Error(t, configpkg.ErrConfigurationNotLoaded, "configpkg.ErrConfigurationNotLoaded should be defined")
	require.Error(t, configpkg.ErrAccountTokenRequired, "configpkg.ErrAccountTokenRequired should be defined")
	require.Error(t, configpkg.ErrAccountLabelRequired, "configpkg.ErrAccountLabelRequired should be defined")
	require.Error(t, configpkg.ErrAccountNotExist, "configpkg.ErrAccountNotExist should be defined")
	require.Error(t, configpkg.ErrCannotRemoveDefaultAccount, "configpkg.ErrCannotRemoveDefaultAccount should be defined")
	require.Error(t, configpkg.ErrNoConfigurationToSave, "configpkg.ErrNoConfigurationToSave should be defined")

	// Test error messages
	require.Contains(t, configpkg.ErrConfigurationNotLoaded.Error(), "configuration not loaded", "Error message should be descriptive")
	require.Contains(t, configpkg.ErrAccountTokenRequired.Error(), "account token is required", "Error message should be descriptive")
	require.Contains(t, configpkg.ErrAccountLabelRequired.Error(), "account label is required", "Error message should be descriptive")
}

func TestTOMLConfigManager_Concurrency(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.toml")

	manager := configpkg.NewTOMLConfigManager(configPath)
	err := manager.LoadOrCreate()
	require.NoError(t, err, "LoadOrCreate should not error")

	// Test concurrent access (basic smoke test)
	done := make(chan bool, 2)

	// Concurrent reads
	go func() {
		for range 10 {
			config := manager.GetConfig()
			if config == nil {
				t.Errorf("GetConfig should not return nil during concurrent access")
			}
		}
		done <- true
	}()

	// Concurrent account operations
	go func() {
		for range 5 {
			testAccount := configpkg.AccountConfig{
				Token: "concurrent_token",
				Label: "Concurrent Account",
			}
			_ = manager.AddAccount("concurrent", testAccount)
			_ = manager.RemoveAccount("concurrent")
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify manager is still functional
	config := manager.GetConfig()
	require.NotNil(t, config, "Manager should remain functional after concurrent access")
}
