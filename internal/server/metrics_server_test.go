package server_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/server"
	"github.com/chadit/CloudMCP/pkg/logger"
	"github.com/chadit/CloudMCP/pkg/metrics"
)

var (
	// Test-specific errors.
	ErrTestLoggerNil      = errors.New("logger cannot be nil")
	ErrMetricsProviderNil = errors.New("metrics provider is nil")
	ErrInvalidMetricsPort = errors.New("invalid metrics port")
)

// createTestMetricsProvider creates a metrics provider for testing.
//
//nolint:ireturn // createTestMetricsProvider returns interface to allow multiple metric implementations.
func createTestMetricsProvider() metrics.Provider {
	provider, err := metrics.NewProvider(&metrics.ProviderConfig{
		Enabled:   true,
		Namespace: "cloudmcp_test",
		Subsystem: "server",
		Backend:   metrics.BackendPrometheus,
	})
	if err != nil {
		panic(err) // This should never happen in tests
	}

	return provider
}

func TestNewMetricsServer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		config        *server.MetricsServerConfig
		logger        logger.Logger
		provider      metrics.Provider
		expectedError error
		description   string
	}{
		{
			name:          "valid_config",
			config:        server.DefaultMetricsServerConfig(),
			logger:        logger.New("debug"),
			provider:      createTestMetricsProvider(),
			expectedError: nil,
			description:   "Valid configuration should succeed",
		},
		{
			name:          "nil_logger",
			config:        server.DefaultMetricsServerConfig(),
			logger:        nil,
			provider:      createTestMetricsProvider(),
			expectedError: server.ErrLoggerNil,
			description:   "Nil logger should return error",
		},
		{
			name:          "nil_provider",
			config:        server.DefaultMetricsServerConfig(),
			logger:        logger.New("debug"),
			provider:      nil,
			expectedError: server.ErrMetricsProviderNil,
			description:   "Nil provider should return error",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server, err := server.NewMetricsServer(testCase.config, testCase.logger, testCase.provider)

			if testCase.expectedError != nil {
				require.Error(t, err, testCase.description)
				require.ErrorIs(t, err, testCase.expectedError, "Error should match expected")
				require.Nil(t, server, "Server should be nil on error")
			} else {
				require.NoError(t, err, testCase.description)
				require.NotNil(t, server, "Server should not be nil on success")
			}
		})
	}
}

func TestMetricsServer_StartStop(t *testing.T) {
	t.Parallel()

	config := server.DefaultMetricsServerConfig()
	config.Port = 0 // Let system assign port
	testLogger := logger.New("debug")
	metricsProvider := createTestMetricsProvider()

	server, err := server.NewMetricsServer(config, testLogger, metricsProvider)
	require.NoError(t, err, "Server creation should succeed")
	require.NotNil(t, server, "Server should not be nil")

	// Test start
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	startCh := make(chan error, 1)
	go func() {
		startCh <- server.Start()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test stop
	err = server.Stop(ctx)
	require.NoError(t, err, "Server stop should succeed")

	// Wait for start to complete
	select {
	case err := <-startCh:
		// Server should stop gracefully
		require.True(t, err == nil || errors.Is(err, http.ErrServerClosed), "Start should complete gracefully")
	case <-ctx.Done():
		t.Fatal("Test timeout")
	}
}

func TestDefaultMetricsServerConfig(t *testing.T) {
	t.Parallel()

	config := server.DefaultMetricsServerConfig()

	require.NotNil(t, config, "Default config should not be nil")
	require.Positive(t, config.Port, "Port should be positive")
	require.True(t, config.EnableRateLimit, "Rate limiting should be enabled by default")
	require.Greater(t, config.RateLimitPerSecond, 0.0, "Rate limit should be positive")
}
