package linode

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/linode/linodego"
	"golang.org/x/oauth2"
)

// HTTPClientConfig contains configuration options for optimizing HTTP client performance.
type HTTPClientConfig struct {
	// Connection pool settings
	MaxIdleConns        int           // Maximum number of idle connections across all hosts
	MaxIdleConnsPerHost int           // Maximum number of idle connections per host
	IdleConnTimeout     time.Duration // Maximum time idle connections are kept alive

	// Timeout settings
	Timeout               time.Duration // Overall request timeout
	DialTimeout           time.Duration // TCP connection timeout
	TLSHandshakeTimeout   time.Duration // TLS handshake timeout
	ResponseHeaderTimeout time.Duration // Time to wait for response headers

	// Keep-alive settings
	KeepAlive         time.Duration // TCP keep-alive interval
	DisableKeepAlives bool          // Whether to disable HTTP keep-alives

	// Security and reliability settings
	DisableCompression bool // Whether to disable compression
	InsecureSkipVerify bool // Whether to skip TLS verification (not recommended)

	// Advanced settings
	MaxConnsPerHost       int           // Maximum connections per host
	ExpectContinueTimeout time.Duration // Time to wait for 100-continue response
}

const (
	// Connection limit thresholds for validation warnings.
	maxIdleConnsWarningThreshold = 1000
	maxIdleConnsPerHostThreshold = 100

	// Default configuration constants.
	defaultMaxIdleConns           = 100
	defaultMaxIdleConnsPerHost    = 10
	defaultIdleConnTimeoutSeconds = 90
	defaultTimeoutSeconds         = 30
	defaultDialTimeoutSeconds     = 10
	defaultTLSTimeoutSeconds      = 10
	defaultResponseTimeoutSeconds = 10
	defaultKeepAliveSeconds       = 30
	defaultMaxConnsPerHost        = 30

	// High-performance configuration constants.
	highPerfMaxIdleConns        = 200
	highPerfMaxIdleConnsPerHost = 20
	highPerfMaxConnsPerHost     = 50
	highPerfIdleConnTimeout     = 120
	highPerfTimeout             = 60

	// Low-latency configuration constants.
	lowLatencyDialTimeout     = 3
	lowLatencyTLSTimeout      = 3
	lowLatencyResponseTimeout = 5
	lowLatencyTimeoutSeconds  = 15
	lowLatencyMaxConnsPerHost = 10

	// Resource-constrained configuration constants.
	resourceConstrainedMaxIdleConns        = 20
	resourceConstrainedMaxIdleConnsPerHost = 2
	resourceConstrainedMaxConnsPerHost     = 5
	resourceConstrainedIdleTimeoutSeconds  = 30

	// Batch processing configuration constants.
	batchMaxIdleConns        = 300
	batchMaxIdleConnsPerHost = 30
	batchMaxConnsPerHost     = 100
	batchTimeoutSeconds      = 120
	batchIdleTimeoutSeconds  = 300
)

// DefaultHTTPClientConfig returns a default configuration optimized for Linode API usage.
func DefaultHTTPClientConfig() HTTPClientConfig {
	return HTTPClientConfig{
		// Connection pool settings - optimized for API usage
		MaxIdleConns:        defaultMaxIdleConns,                         // Allow many idle connections for reuse
		MaxIdleConnsPerHost: defaultMaxIdleConnsPerHost,                  // Reasonable per-host limit
		IdleConnTimeout:     defaultIdleConnTimeoutSeconds * time.Second, // Keep connections alive for reuse

		// Timeout settings - balanced for API responsiveness
		Timeout:               defaultTimeoutSeconds * time.Second,         // Overall request timeout
		DialTimeout:           defaultDialTimeoutSeconds * time.Second,     // TCP connection timeout
		TLSHandshakeTimeout:   defaultTLSTimeoutSeconds * time.Second,      // TLS handshake timeout
		ResponseHeaderTimeout: defaultResponseTimeoutSeconds * time.Second, // Response header timeout

		// Keep-alive settings - improve performance
		KeepAlive:         defaultKeepAliveSeconds * time.Second, // TCP keep-alive
		DisableKeepAlives: false,                                 // Enable keep-alives for performance

		// Compression and security
		DisableCompression: false, // Enable compression to reduce bandwidth
		InsecureSkipVerify: false, // Always verify TLS certificates

		// Advanced settings
		MaxConnsPerHost:       defaultMaxConnsPerHost, // Allow multiple concurrent connections
		ExpectContinueTimeout: 1 * time.Second,        // Quick 100-continue handling
	}
}

// CreateOptimizedLinodeClient creates a Linode client with optimized HTTP transport settings.
func CreateOptimizedLinodeClient(token string, config HTTPClientConfig) *linodego.Client {
	// Create custom transport with optimized settings
	transport := &http.Transport{
		// Connection pool settings
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		IdleConnTimeout:     config.IdleConnTimeout,
		MaxConnsPerHost:     config.MaxConnsPerHost,

		// Dialer settings for connection establishment
		DialContext: (&net.Dialer{
			Timeout:   config.DialTimeout,
			KeepAlive: config.KeepAlive,
		}).DialContext,

		// TLS settings
		TLSHandshakeTimeout: config.TLSHandshakeTimeout,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.InsecureSkipVerify, //nolint:gosec // Configurable for testing environments
		},

		// HTTP settings
		DisableKeepAlives:     config.DisableKeepAlives,
		DisableCompression:    config.DisableCompression,
		ResponseHeaderTimeout: config.ResponseHeaderTimeout,
		ExpectContinueTimeout: config.ExpectContinueTimeout,

		// Force HTTP/2 for better performance
		ForceAttemptHTTP2: true,
	}

	// Create OAuth2 token source
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: token,
	})

	// Create OAuth2 HTTP client with our optimized transport
	ctx := context.TODO()
	oauthClient := oauth2.NewClient(ctx, tokenSource)
	oauthClient.Transport = transport
	oauthClient.Timeout = config.Timeout

	// Create Linode client
	client := linodego.NewClient(oauthClient)

	return &client
}

// CreateStandardLinodeClient creates a Linode client with standard settings (for comparison).
func CreateStandardLinodeClient(token string) *linodego.Client {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: token,
	})

	ctx := context.TODO()
	oauthClient := oauth2.NewClient(ctx, tokenSource)
	client := linodego.NewClient(oauthClient)

	return &client
}

// HTTPClientStats provides statistics about HTTP client performance.
type HTTPClientStats struct {
	IdleConnections    int            `json:"idleConnections"`
	ActiveConnections  int            `json:"activeConnections"`
	TotalConnections   int            `json:"totalConnections"`
	ConnectionsPerHost map[string]int `json:"connectionsPerHost"`
	TransportType      string         `json:"transportType"`
	KeepAliveEnabled   bool           `json:"keepAliveEnabled"`
	CompressionEnabled bool           `json:"compressionEnabled"`
	HTTP2Enabled       bool           `json:"http2Enabled"`
}

// GetHTTPClientStats returns statistics about the HTTP client transport.
// Note: This is a simplified version as Go's http.Transport doesn't expose all internal stats.
func GetHTTPClientStats(_ *linodego.Client) HTTPClientStats {
	stats := HTTPClientStats{
		TransportType:      "http.Transport",
		KeepAliveEnabled:   true, // Default assumption
		CompressionEnabled: true, // Default assumption
		HTTP2Enabled:       true, // Our client forces HTTP/2
	}

	// Note: Go's standard http.Transport doesn't expose connection pool statistics
	// This would require custom instrumentation or metrics collection for detailed stats

	return stats
}

// PerformanceTest contains configuration for HTTP client performance testing.
type PerformanceTest struct {
	RequestCount    int           // Number of requests to make
	ConcurrentUsers int           // Number of concurrent users
	TestDuration    time.Duration // Maximum test duration
	Endpoint        string        // API endpoint to test
}

// BenchmarkResult contains the results of a performance benchmark.
type BenchmarkResult struct {
	TotalRequests      int           `json:"totalRequests"`
	SuccessfulRequests int           `json:"successfulRequests"`
	FailedRequests     int           `json:"failedRequests"`
	AverageLatency     time.Duration `json:"averageLatency"`
	MinLatency         time.Duration `json:"minLatency"`
	MaxLatency         time.Duration `json:"maxLatency"`
	RequestsPerSecond  float64       `json:"requestsPerSecond"`
	TestDuration       time.Duration `json:"testDuration"`
	ConcurrentUsers    int           `json:"concurrentUsers"`
	ErrorRate          float64       `json:"errorRate"`
}

// ClientValidator provides methods to validate HTTP client configuration.
type ClientValidator struct{}

// ValidateConfig validates an HTTP client configuration for potential issues.
func (v *ClientValidator) ValidateConfig(config HTTPClientConfig) []string {
	var warnings []string

	// Check for unreasonably low timeouts
	if config.Timeout < 5*time.Second {
		warnings = append(warnings, "Timeout is very low and may cause premature request failures")
	}

	if config.DialTimeout < 1*time.Second {
		warnings = append(warnings, "DialTimeout is very low and may cause connection failures")
	}

	// Check for unreasonably high connection limits
	if config.MaxIdleConns > maxIdleConnsWarningThreshold {
		warnings = append(warnings, "MaxIdleConns is very high and may consume excessive memory")
	}

	if config.MaxIdleConnsPerHost > maxIdleConnsPerHostThreshold {
		warnings = append(warnings, "MaxIdleConnsPerHost is very high for API usage")
	}

	// Check for disabled keep-alives
	if config.DisableKeepAlives {
		warnings = append(warnings, "DisableKeepAlives reduces performance for API usage")
	}

	// Check for disabled compression
	if config.DisableCompression {
		warnings = append(warnings, "DisableCompression may increase bandwidth usage")
	}

	// Check for insecure TLS settings
	if config.InsecureSkipVerify {
		warnings = append(warnings, "InsecureSkipVerify is a security risk and should not be used in production")
	}

	// Check for very short idle connection timeout
	if config.IdleConnTimeout < 30*time.Second {
		warnings = append(warnings, "IdleConnTimeout is short and may reduce connection reuse benefits")
	}

	return warnings
}

// RecommendConfig recommends optimal configuration based on usage patterns.
func (v *ClientValidator) RecommendConfig(usage string) HTTPClientConfig {
	config := DefaultHTTPClientConfig()

	switch usage {
	case "high-throughput":
		// Optimize for high-throughput scenarios
		config.MaxIdleConns = highPerfMaxIdleConns
		config.MaxIdleConnsPerHost = highPerfMaxIdleConnsPerHost
		config.MaxConnsPerHost = highPerfMaxConnsPerHost
		config.IdleConnTimeout = highPerfIdleConnTimeout * time.Second
		config.Timeout = highPerfTimeout * time.Second

	case "low-latency":
		// Optimize for low-latency scenarios
		config.DialTimeout = lowLatencyDialTimeout * time.Second
		config.TLSHandshakeTimeout = lowLatencyTLSTimeout * time.Second
		config.ResponseHeaderTimeout = lowLatencyResponseTimeout * time.Second
		config.Timeout = lowLatencyTimeoutSeconds * time.Second
		config.MaxConnsPerHost = lowLatencyMaxConnsPerHost

	case "resource-constrained":
		// Optimize for resource-constrained environments
		config.MaxIdleConns = resourceConstrainedMaxIdleConns
		config.MaxIdleConnsPerHost = resourceConstrainedMaxIdleConnsPerHost
		config.MaxConnsPerHost = resourceConstrainedMaxConnsPerHost
		config.IdleConnTimeout = resourceConstrainedIdleTimeoutSeconds * time.Second

	case "batch-processing":
		// Optimize for batch processing scenarios
		config.MaxIdleConns = batchMaxIdleConns
		config.MaxIdleConnsPerHost = batchMaxIdleConnsPerHost
		config.MaxConnsPerHost = batchMaxConnsPerHost
		config.Timeout = batchTimeoutSeconds * time.Second
		config.IdleConnTimeout = batchIdleTimeoutSeconds * time.Second
	default:
	}

	return config
}
