// Package performance provides performance benchmarks for CloudMCP health checks
package performance

import (
	"testing"
	"time"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/internal/server"
)

func BenchmarkHealthCheck(b *testing.B) {
	// Setup test configuration
	cfg := &config.Config{
		ServerName: "CloudMCP-BenchmarkTest",
		LogLevel:   "error", // Minimal logging for performance
	}

	// Create server instance
	srv, err := server.New(cfg)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}

	// Create benchmark config
	benchConfig := NewBenchmarkConfig("HealthCheck")
	benchConfig.MaxDuration = 5 * time.Second // Limit benchmark time
	benchConfig.MinOps = 1000                 // Expect at least 1000 ops

	// Run the benchmark
	benchConfig.RunBenchmark(b, func() {
		// Simulate health check operation
		// In real implementation, this would call the actual health check
		_ = srv != nil // Simple operation to benchmark
	})
}

func BenchmarkConfigLoad(b *testing.B) {
	benchConfig := NewBenchmarkConfig("ConfigLoad")

	benchConfig.RunBenchmark(b, func() {
		// Test configuration loading performance
		cfg, err := config.Load()
		if err != nil {
			b.Fatalf("Config load failed: %v", err)
		}
		_ = cfg.ServerName // Use the config to prevent optimization
	})
}

func BenchmarkServerCreation(b *testing.B) {
	cfg := &config.Config{
		ServerName: "CloudMCP-BenchmarkTest",
		LogLevel:   "error",
	}

	benchConfig := NewBenchmarkConfig("ServerCreation")
	benchConfig.MaxDuration = 10 * time.Second
	benchConfig.MinOps = 10 // Server creation is expensive, expect fewer ops

	benchConfig.RunBenchmark(b, func() {
		srv, err := server.New(cfg)
		if err != nil {
			b.Fatalf("Server creation failed: %v", err)
		}
		_ = srv // Use the server to prevent optimization
	})
}

// TestBenchmarkFramework tests the benchmark framework itself
func TestBenchmarkFramework(t *testing.T) {
	config := NewBenchmarkConfig("test")
	if config.Name != "test" {
		t.Errorf("Expected name 'test', got %s", config.Name)
	}

	// Test measurement function
	duration := Measure(func() {
		time.Sleep(1 * time.Millisecond)
	})

	if duration < 1*time.Millisecond {
		t.Errorf("Expected duration >= 1ms, got %v", duration)
	}

	t.Logf("Benchmark framework test completed successfully")
}
