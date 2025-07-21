// Package testing provides integration tests for CloudMCP
package testing

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/internal/server"
)

func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()
	os.Exit(code)
}

func TestServerIntegration(t *testing.T) {
	// Use test context with timeout
	ctx := context.Background()
	if testing.Testing() {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	// Create minimal test configuration
	cfg := &config.Config{
		ServerName: "CloudMCP-IntegrationTest",
		LogLevel:   "error",
	}

	// Test server creation
	srv, err := server.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test that server can be created without starting
	if srv == nil {
		t.Fatal("Server should not be nil")
	}

	t.Logf("Integration test completed successfully")
}

func TestConfigLoad(t *testing.T) {
	// Set test environment variables
	_ = os.Setenv("CLOUD_MCP_SERVER_NAME", "TestServer")
	_ = os.Setenv("LOG_LEVEL", "debug")
	defer func() {
		_ = os.Unsetenv("CLOUD_MCP_SERVER_NAME")
		_ = os.Unsetenv("LOG_LEVEL")
	}()

	// Test configuration loading
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.ServerName != "TestServer" {
		t.Errorf("Expected ServerName 'TestServer', got %s", cfg.ServerName)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("Expected LogLevel 'debug', got %s", cfg.LogLevel)
	}

	t.Logf("Config test completed successfully")
}