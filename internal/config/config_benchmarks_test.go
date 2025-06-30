package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	configpkg "github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/internal/services/linode"
	"github.com/chadit/CloudMCP/pkg/logger"
)

// BenchmarkConfigReload benchmarks the performance of configuration reload operations
// to establish baseline performance metrics for dynamic configuration management.
//
// **Benchmark Coverage:**
// • Configuration file loading and parsing
// • TOML configuration reload operations
// • Service reconfiguration with new accounts
// • Memory allocation patterns during reload
//
// **Performance Targets:**
// • Configuration reload: <10ms for typical config files
// • Service reconfiguration: <50ms with new accounts
// • Memory allocation: Minimal allocations per reload
// • File I/O operations: <5ms for small config files
//
// **Benchmark Environment:**
// • Temporary configuration files for isolation
// • Realistic configuration structures
// • Multiple account configurations for comprehensive testing
func BenchmarkConfigReload(b *testing.B) {
	tempDir := b.TempDir()
	configPath := filepath.Join(tempDir, "benchmark-config.toml")

	// Create initial configuration manager
	manager := configpkg.NewTOMLConfigManager(configPath)
	require.NotNil(b, manager, "Config manager should not be nil")

	// Initialize with base configuration
	err := manager.LoadOrCreate()
	require.NoError(b, err, "Initial config creation should succeed")

	// Add multiple accounts for realistic scenario
	accounts := []struct {
		name    string
		account configpkg.AccountConfig
	}{
		{"primary", configpkg.AccountConfig{Token: "token-primary-12345", Label: "Primary Account"}},
		{"development", configpkg.AccountConfig{Token: "token-dev-67890", Label: "Development Account"}},
		{"staging", configpkg.AccountConfig{Token: "token-staging-54321", Label: "Staging Account"}},
		{"production", configpkg.AccountConfig{Token: "token-prod-98765", Label: "Production Account"}},
	}

	for _, acc := range accounts {
		err := manager.AddAccount(acc.name, acc.account)
		require.NoError(b, err, "Adding account should succeed")
	}

	err = manager.Save()
	require.NoError(b, err, "Saving initial config should succeed")

	b.Run("ConfigReload", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			err := manager.Reload()
			if err != nil {
				b.Fatalf("Config reload failed: %v", err)
			}
		}
	})

	b.Run("ConfigGetAfterReload", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			// Reload configuration
			err := manager.Reload()
			if err != nil {
				b.Fatalf("Config reload failed: %v", err)
			}

			// Get configuration
			config := manager.GetConfig()
			if config == nil {
				b.Fatal("Config should not be nil")
			}
		}
	})

	b.Run("ConfigSave", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			err := manager.Save()
			if err != nil {
				b.Fatalf("Config save failed: %v", err)
			}
		}
	})
}

// BenchmarkConfigMemoryAllocation benchmarks memory allocation patterns
// during configuration operations to identify optimization opportunities.
//
// **Memory Optimization Targets:**
// • Minimize allocations per reload operation
// • Reduce GC pressure for frequent configuration access
// • Optimize TOML parsing and structure creation
// • Monitor memory growth patterns during operations
func BenchmarkConfigMemoryAllocation(b *testing.B) {
	tempDir := b.TempDir()
	configPath := filepath.Join(tempDir, "memory-benchmark-config.toml")

	manager := configpkg.NewTOMLConfigManager(configPath)
	require.NotNil(b, manager, "Config manager should not be nil")

	err := manager.LoadOrCreate()
	require.NoError(b, err, "Initial config creation should succeed")

	// Add accounts for realistic memory usage
	for i := 0; i < 10; i++ {
		err := manager.AddAccount(
			"account"+string(rune('A'+i)),
			configpkg.AccountConfig{
				Token: "token-" + string(rune('A'+i)) + "-12345",
				Label: "Account " + string(rune('A'+i)),
			},
		)
		require.NoError(b, err, "Adding account should succeed")
	}

	err = manager.Save()
	require.NoError(b, err, "Saving config should succeed")

	b.Run("MemoryPerReload", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			err := manager.Reload()
			if err != nil {
				b.Fatalf("Config reload failed: %v", err)
			}
		}
	})

	b.Run("MemoryPerConfigGet", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			config := manager.GetConfig()
			if config == nil {
				b.Fatal("Config should not be nil")
			}
		}
	})

	b.Run("MemoryPerAccountAdd", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := range b.N {
			accountName := "temp-account-" + string(rune(i%26+'A'))
			err := manager.AddAccount(accountName, configpkg.AccountConfig{
				Token: "temp-token-" + string(rune(i%26+'A')),
				Label: "Temporary Account " + string(rune(i%26+'A')),
			})
			if err != nil {
				b.Fatalf("Adding account failed: %v", err)
			}

			// Clean up to avoid memory accumulation affecting results
			if i%10 == 9 {
				err := manager.RemoveAccount(accountName)
				if err != nil {
					// Account might not exist, which is fine for benchmark
					b.Logf("Account removal expected result: %v", err)
				}
			}
		}
	})
}

// BenchmarkServiceReconfiguration benchmarks service reconfiguration performance.
// when configuration changes occur, simulating real-world dynamic updates.
//
// **Service Reconfiguration Performance:**
// • Service creation with new configuration
// • Account manager updates with configuration changes
// • Memory allocation during service reconfiguration
// • Impact of account count on reconfiguration performance
func BenchmarkServiceReconfiguration(b *testing.B) {
	tempDir := b.TempDir()
	configPath := filepath.Join(tempDir, "service-reconfig-benchmark.toml")
	log := logger.New("error") // Minimize logging overhead

	manager := configpkg.NewTOMLConfigManager(configPath)
	require.NotNil(b, manager, "Config manager should not be nil")

	err := manager.LoadOrCreate()
	require.NoError(b, err, "Initial config creation should succeed")

	// Create base configuration
	baseAccounts := []struct {
		name    string
		account configpkg.AccountConfig
	}{
		{"primary", configpkg.AccountConfig{Token: "token-primary-12345", Label: "Primary Account"}},
		{"development", configpkg.AccountConfig{Token: "token-dev-67890", Label: "Development Account"}},
	}

	for _, acc := range baseAccounts {
		err := manager.AddAccount(acc.name, acc.account)
		require.NoError(b, err, "Adding account should succeed")
	}

	err = manager.SetDefaultAccount("primary")
	require.NoError(b, err, "Setting default account should succeed")

	err = manager.Save()
	require.NoError(b, err, "Saving config should succeed")

	b.Run("ServiceCreationWithConfig", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			config := manager.GetConfig()
			if config == nil {
				b.Fatal("Config should not be nil")
			}

			legacyConfig := config.ToLegacyConfig()
			if legacyConfig == nil {
				b.Fatal("Legacy config should not be nil")
			}

			service, err := linode.New(legacyConfig, log)
			if err != nil {
				b.Fatalf("Service creation failed: %v", err)
			}

			if service == nil {
				b.Fatal("Service should not be nil")
			}
		}
	})

	b.Run("ConfigReloadAndServiceUpdate", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			// Reload configuration
			err := manager.Reload()
			if err != nil {
				b.Fatalf("Config reload failed: %v", err)
			}

			// Get updated configuration
			config := manager.GetConfig()
			if config == nil {
				b.Fatal("Config should not be nil")
			}

			// Create service with updated config
			legacyConfig := config.ToLegacyConfig()
			service, err := linode.New(legacyConfig, log)
			if err != nil {
				b.Fatalf("Service creation with updated config failed: %v", err)
			}

			if service == nil {
				b.Fatal("Service should not be nil")
			}
		}
	})

	b.Run("MemoryPerServiceReconfiguration", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			// Reload and recreate service
			err := manager.Reload()
			if err != nil {
				b.Fatalf("Config reload failed: %v", err)
			}

			config := manager.GetConfig()
			legacyConfig := config.ToLegacyConfig()

			service, err := linode.New(legacyConfig, log)
			if err != nil {
				b.Fatalf("Service reconfiguration failed: %v", err)
			}

			if service == nil {
				b.Fatal("Service should not be nil")
			}
		}
	})
}

// BenchmarkConfigurationPersistence benchmarks configuration file I/O operations.
// to establish baseline performance for configuration persistence scenarios.
//
// **File I/O Performance Targets:**
// • Configuration loading: <5ms for small files
// • Configuration saving: <10ms for typical configurations
// • File system operations: Minimal overhead
// • TOML parsing/serialization: Efficient processing
func BenchmarkConfigurationPersistence(b *testing.B) {
	tempDir := b.TempDir()
	configPath := filepath.Join(tempDir, "persistence-benchmark.toml")

	// Create a realistic configuration file
	createRealisticConfig := func() error {
		manager := configpkg.NewTOMLConfigManager(configPath)
		err := manager.LoadOrCreate()
		if err != nil {
			return fmt.Errorf("failed to load or create config: %w", err)
		}

		// Add multiple accounts with realistic data
		accounts := map[string]configpkg.AccountConfig{
			"production":  {Token: "pat_1234567890abcdef1234567890abcdef12345678", Label: "Production Environment - Main Account"},
			"staging":     {Token: "pat_abcdef1234567890abcdef1234567890abcdef12", Label: "Staging Environment - Testing Account"},
			"development": {Token: "pat_567890abcdef1234567890abcdef1234567890ab", Label: "Development Environment - Local Testing"},
			"backup":      {Token: "pat_cdef1234567890abcdef1234567890abcdef1234", Label: "Backup Account - Emergency Access"},
		}

		for name, account := range accounts {
			err := manager.AddAccount(name, account)
			if err != nil {
				return fmt.Errorf("failed to add account %s: %w", name, err)
			}
		}

		err = manager.SetDefaultAccount("production")
		if err != nil {
			return fmt.Errorf("failed to set default account: %w", err)
		}

		return manager.Save()
	}

	err := createRealisticConfig()
	require.NoError(b, err, "Creating realistic config should succeed")

	b.Run("ConfigurationLoading", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			manager := configpkg.NewTOMLConfigManager(configPath)
			err := manager.LoadOrCreate()
			if err != nil {
				b.Fatalf("Configuration loading failed: %v", err)
			}
		}
	})

	b.Run("ConfigurationSaving", func(b *testing.B) {
		manager := configpkg.NewTOMLConfigManager(configPath)
		err := manager.LoadOrCreate()
		require.NoError(b, err, "Initial load should succeed")

		b.ResetTimer()
		for range b.N {
			err := manager.Save()
			if err != nil {
				b.Fatalf("Configuration saving failed: %v", err)
			}
		}
	})

	b.Run("FileIOOperations", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			// Read file
			_, err := os.ReadFile(configPath)
			if err != nil {
				b.Fatalf("File read failed: %v", err)
			}

			// Write file (with different content to avoid caching)
			testContent := "# Test configuration file\n[system]\nserver_name = \"Test\"\n"
			err = os.WriteFile(configPath+".temp", []byte(testContent), 0o644)
			if err != nil {
				b.Fatalf("File write failed: %v", err)
			}

			// Clean up
			os.Remove(configPath + ".temp")
		}
	})

	b.Run("MemoryPerPersistenceOperation", func(b *testing.B) {
		manager := configpkg.NewTOMLConfigManager(configPath)
		err := manager.LoadOrCreate()
		require.NoError(b, err, "Initial load should succeed")

		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			// Complete persistence cycle
			err := manager.Reload()
			if err != nil {
				b.Fatalf("Config reload failed: %v", err)
			}

			err = manager.Save()
			if err != nil {
				b.Fatalf("Config save failed: %v", err)
			}
		}
	})
}

// BenchmarkConcurrentConfigAccess benchmarks concurrent configuration access patterns.
// to ensure thread safety and performance under concurrent load.
//
// **Concurrency Performance:**
// • Multiple goroutines reading configuration simultaneously
// • Configuration reload with concurrent access
// • Thread safety verification under load
// • Performance degradation under concurrent access
func BenchmarkConcurrentConfigAccess(b *testing.B) {
	tempDir := b.TempDir()
	configPath := filepath.Join(tempDir, "concurrent-benchmark.toml")

	manager := configpkg.NewTOMLConfigManager(configPath)
	require.NotNil(b, manager, "Config manager should not be nil")

	err := manager.LoadOrCreate()
	require.NoError(b, err, "Initial config creation should succeed")

	// Add accounts for realistic concurrent testing
	for i := 0; i < 5; i++ {
		err := manager.AddAccount(
			"account"+string(rune('1'+i)),
			configpkg.AccountConfig{
				Token: "token-" + string(rune('1'+i)) + "-concurrent",
				Label: "Concurrent Account " + string(rune('1'+i)),
			},
		)
		require.NoError(b, err, "Adding account should succeed")
	}

	err = manager.Save()
	require.NoError(b, err, "Saving config should succeed")

	b.Run("ConcurrentConfigGet", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				config := manager.GetConfig()
				if config == nil {
					b.Error("Config should not be nil")
				}
			}
		})
	})

	b.Run("ConcurrentReloadAndGet", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// 90% reads, 10% reloads for realistic usage pattern
				if b.N%10 == 0 {
					err := manager.Reload()
					if err != nil {
						b.Errorf("Config reload failed: %v", err)
					}
				} else {
					config := manager.GetConfig()
					if config == nil {
						b.Error("Config should not be nil")
					}
				}
			}
		})
	})

	b.Run("MemoryPerConcurrentAccess", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				config := manager.GetConfig()
				if config == nil {
					b.Error("Config should not be nil")
					continue
				}

				// Simulate realistic configuration usage
				_ = len(config.Accounts)
				_ = config.System.DefaultAccount
			}
		})
	})
}
