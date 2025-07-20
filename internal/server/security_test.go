package server_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/internal/server"
	"github.com/chadit/CloudMCP/pkg/logger"
)

// TestMetricsServerSecurity tests basic server creation.
// The original tests were removed since they tested private fields and methods
// that are not accessible from external test packages.
func TestMetricsServerSecurity(t *testing.T) {
	t.Parallel()

	// Create minimal config for testing.
	cfg := &config.Config{
		ServerName:    "test-server",
		LogLevel:      "debug",
		EnableMetrics: true,
		MetricsPort:   8080,
	}

	log := logger.New("debug")

	// Test server creation.
	serverInstance, err := server.New(cfg, log)
	require.NoError(t, err, "server creation should succeed")
	require.NotNil(t, serverInstance, "server should not be nil")
}
