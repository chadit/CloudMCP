package linode_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/services/linode"
)

func TestDefaultHTTPClientConfig(t *testing.T) {
	t.Parallel()
	config := linode.DefaultHTTPClientConfig()

	// Verify default values are set appropriately
	require.Equal(t, 100, config.MaxIdleConns, "MaxIdleConns should be set to default")
	require.Equal(t, 10, config.MaxIdleConnsPerHost, "MaxIdleConnsPerHost should be set to default")
	require.Equal(t, 90*time.Second, config.IdleConnTimeout, "IdleConnTimeout should be set to default")
	require.Equal(t, 30*time.Second, config.Timeout, "Timeout should be set to default")
	require.Equal(t, 10*time.Second, config.DialTimeout, "DialTimeout should be set to default")
	require.Equal(t, 10*time.Second, config.TLSHandshakeTimeout, "TLSHandshakeTimeout should be set to default")
	require.Equal(t, 10*time.Second, config.ResponseHeaderTimeout, "ResponseHeaderTimeout should be set to default")
	require.Equal(t, 30*time.Second, config.KeepAlive, "KeepAlive should be set to default")
	require.False(t, config.DisableKeepAlives, "DisableKeepAlives should be false by default")
	require.False(t, config.DisableCompression, "DisableCompression should be false by default")
	require.False(t, config.InsecureSkipVerify, "InsecureSkipVerify should be false by default")
	require.Equal(t, 30, config.MaxConnsPerHost, "MaxConnsPerHost should be set to default")
	require.Equal(t, 1*time.Second, config.ExpectContinueTimeout, "ExpectContinueTimeout should be set to default")
}

func TestCreateOptimizedLinodeClient(t *testing.T) {
	t.Parallel()
	config := linode.DefaultHTTPClientConfig()
	client := linode.CreateOptimizedLinodeClient("test-token", config)

	require.NotNil(t, client, "Client should not be nil")
	// Verify the client was created successfully
	// Note: We can't easily test the internal configuration without exposing internals
	// but we can verify the client is functional by checking it's not nil
}

func TestCreateStandardLinodeClient(t *testing.T) {
	t.Parallel()
	client := linode.CreateStandardLinodeClient("test-token")

	require.NotNil(t, client, "Standard client should not be nil")
}

func TestGetHTTPClientStats(t *testing.T) {
	t.Parallel()
	config := linode.DefaultHTTPClientConfig()
	client := linode.CreateOptimizedLinodeClient("test-token", config)

	stats := linode.GetHTTPClientStats(client)

	require.Equal(t, "http.Transport", stats.TransportType, "Transport type should be http.Transport")
	require.True(t, stats.KeepAliveEnabled, "Keep alive should be enabled")
	require.True(t, stats.CompressionEnabled, "Compression should be enabled")
	require.True(t, stats.HTTP2Enabled, "HTTP2 should be enabled")
}

func TestClientValidator_ValidateConfig(t *testing.T) {
	t.Parallel()
	validator := &linode.ClientValidator{}

	tests := []struct {
		name         string
		config       linode.HTTPClientConfig
		wantWarnings int
		description  string
	}{
		{
			name:         "default config should have no warnings",
			config:       linode.DefaultHTTPClientConfig(),
			wantWarnings: 0,
			description:  "Default configuration should be optimized",
		},
		{
			name: "low timeout should generate warning",
			config: linode.HTTPClientConfig{
				Timeout:         2 * time.Second, // Very low
				DialTimeout:     5 * time.Second,
				IdleConnTimeout: 60 * time.Second, // Normal to avoid additional warnings
			},
			wantWarnings: 1,
			description:  "Low timeout should generate warning",
		},
		{
			name: "very low dial timeout should generate warning",
			config: linode.HTTPClientConfig{
				Timeout:         10 * time.Second,
				DialTimeout:     500 * time.Millisecond, // Very low
				IdleConnTimeout: 60 * time.Second,       // Normal to avoid additional warnings
			},
			wantWarnings: 1,
			description:  "Very low dial timeout should generate warning",
		},
		{
			name: "high connection limits should generate warnings",
			config: linode.HTTPClientConfig{
				Timeout:             10 * time.Second,
				DialTimeout:         5 * time.Second,
				IdleConnTimeout:     60 * time.Second, // Normal to avoid additional warnings
				MaxIdleConns:        1500,             // Very high
				MaxIdleConnsPerHost: 150,              // Very high
			},
			wantWarnings: 2,
			description:  "High connection limits should generate warnings",
		},
		{
			name: "disabled optimizations should generate warnings",
			config: linode.HTTPClientConfig{
				Timeout:            10 * time.Second,
				DialTimeout:        5 * time.Second,
				IdleConnTimeout:    60 * time.Second, // Normal to avoid additional warnings
				DisableKeepAlives:  true,             // Poor for performance
				DisableCompression: true,             // Increases bandwidth
			},
			wantWarnings: 2,
			description:  "Disabled optimizations should generate warnings",
		},
		{
			name: "insecure settings should generate warning",
			config: linode.HTTPClientConfig{
				Timeout:            10 * time.Second,
				DialTimeout:        5 * time.Second,
				IdleConnTimeout:    60 * time.Second, // Normal to avoid additional warnings
				InsecureSkipVerify: true,             // Security risk
			},
			wantWarnings: 1,
			description:  "Insecure TLS settings should generate warning",
		},
		{
			name: "short idle timeout should generate warning",
			config: linode.HTTPClientConfig{
				Timeout:         10 * time.Second,
				DialTimeout:     5 * time.Second,
				IdleConnTimeout: 15 * time.Second, // Very short
			},
			wantWarnings: 1,
			description:  "Short idle connection timeout should generate warning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			warnings := validator.ValidateConfig(tt.config)
			require.Len(t, warnings, tt.wantWarnings, "Should have expected number of warnings: %s", tt.description)
		})
	}
}

func TestClientValidator_RecommendConfig(t *testing.T) {
	t.Parallel()

	validator := &linode.ClientValidator{}

	tests := []struct {
		usage    string
		testFunc func(*testing.T, linode.HTTPClientConfig)
	}{
		{
			usage: "high-throughput",
			testFunc: func(t *testing.T, config linode.HTTPClientConfig) {
				require.Equal(t, 200, config.MaxIdleConns, "High-throughput should have many idle connections")
				require.Equal(t, 20, config.MaxIdleConnsPerHost, "High-throughput should have many connections per host")
				require.Equal(t, 50, config.MaxConnsPerHost, "High-throughput should allow many concurrent connections")
				require.Equal(t, 60*time.Second, config.Timeout, "High-throughput should have longer timeout")
			},
		},
		{
			usage: "low-latency",
			testFunc: func(t *testing.T, config linode.HTTPClientConfig) {
				require.Equal(t, 3*time.Second, config.DialTimeout, "Low-latency should have fast dial timeout")
				require.Equal(t, 3*time.Second, config.TLSHandshakeTimeout, "Low-latency should have fast TLS timeout")
				require.Equal(t, 15*time.Second, config.Timeout, "Low-latency should have short overall timeout")
			},
		},
		{
			usage: "resource-constrained",
			testFunc: func(t *testing.T, config linode.HTTPClientConfig) {
				require.Equal(t, 20, config.MaxIdleConns, "Resource-constrained should limit connections")
				require.Equal(t, 2, config.MaxIdleConnsPerHost, "Resource-constrained should limit per-host connections")
				require.Equal(t, 5, config.MaxConnsPerHost, "Resource-constrained should limit concurrent connections")
			},
		},
		{
			usage: "batch-processing",
			testFunc: func(t *testing.T, config linode.HTTPClientConfig) {
				require.Equal(t, 300, config.MaxIdleConns, "Batch processing should allow many connections")
				require.Equal(t, 100, config.MaxConnsPerHost, "Batch processing should allow many concurrent connections")
				require.Equal(t, 120*time.Second, config.Timeout, "Batch processing should have long timeout")
			},
		},
		{
			usage: "unknown-usage",
			testFunc: func(t *testing.T, config linode.HTTPClientConfig) {
				defaultConfig := linode.DefaultHTTPClientConfig()
				require.Equal(t, defaultConfig.MaxIdleConns, config.MaxIdleConns, "Unknown usage should return default config")
				require.Equal(t, defaultConfig.Timeout, config.Timeout, "Unknown usage should return default config")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.usage, func(t *testing.T) {
			t.Parallel()
			config := validator.RecommendConfig(tt.usage)
			tt.testFunc(t, config)
		})
	}
}

func TestHTTPClientConfig_EdgeCases(t *testing.T) {
	t.Parallel()
	// Test various edge cases and boundary conditions

	t.Run("zero values", func(t *testing.T) {
		t.Parallel()
		config := linode.HTTPClientConfig{}
		validator := &linode.ClientValidator{}

		warnings := validator.ValidateConfig(config)
		// Zero values should generate multiple warnings
		require.Greater(t, len(warnings), 0, "Zero values should generate warnings")
	})

	t.Run("negative timeouts", func(t *testing.T) {
		t.Parallel()

		config := linode.HTTPClientConfig{
			Timeout:     -1 * time.Second,
			DialTimeout: -1 * time.Second,
		}
		validator := &linode.ClientValidator{}

		warnings := validator.ValidateConfig(config)
		// Negative timeouts should be caught as very low
		require.Greater(t, len(warnings), 0, "Negative timeouts should generate warnings")
	})
}

func TestHTTPClientCreation_TokenHandling(t *testing.T) {
	t.Parallel()
	// Test with various token formats
	tokens := []string{
		"test-token",
		"pat_abcdef123456789",
		"", // Empty token should still create client
	}

	config := linode.DefaultHTTPClientConfig()

	for _, token := range tokens {
		t.Run("token_"+token, func(t *testing.T) {
			t.Parallel()

			client := linode.CreateOptimizedLinodeClient(token, config)
			require.NotNil(t, client, "Client should be created regardless of token format")

			standardClient := linode.CreateStandardLinodeClient(token)
			require.NotNil(t, standardClient, "Standard client should be created regardless of token format")
		})
	}
}

func TestPerformanceStructs(t *testing.T) {
	t.Parallel()
	// Test that performance-related structs are properly defined

	t.Run("PerformanceTest struct", func(t *testing.T) {
		t.Parallel()

		test := linode.PerformanceTest{
			RequestCount:    100,
			ConcurrentUsers: 10,
			TestDuration:    5 * time.Minute,
			Endpoint:        "/linode/instances",
		}

		require.Equal(t, 100, test.RequestCount, "RequestCount should be set")
		require.Equal(t, 10, test.ConcurrentUsers, "ConcurrentUsers should be set")
		require.Equal(t, 5*time.Minute, test.TestDuration, "TestDuration should be set")
		require.Equal(t, "/linode/instances", test.Endpoint, "Endpoint should be set")
	})

	t.Run("BenchmarkResult struct", func(t *testing.T) {
		t.Parallel()

		result := linode.BenchmarkResult{
			TotalRequests:      100,
			SuccessfulRequests: 95,
			FailedRequests:     5,
			AverageLatency:     100 * time.Millisecond,
			MinLatency:         50 * time.Millisecond,
			MaxLatency:         500 * time.Millisecond,
			RequestsPerSecond:  50.0,
			TestDuration:       2 * time.Second,
			ConcurrentUsers:    10,
			ErrorRate:          5.0,
		}

		require.Equal(t, 100, result.TotalRequests, "TotalRequests should be set")
		require.Equal(t, 95, result.SuccessfulRequests, "SuccessfulRequests should be set")
		require.Equal(t, 5, result.FailedRequests, "FailedRequests should be set")
		require.Equal(t, 5.0, result.ErrorRate, "ErrorRate should be set")
	})
}
