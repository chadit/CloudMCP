package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Collector provides methods for recording various CloudMCP metrics.
// All metrics are encapsulated within the collector to avoid global variables.
type Collector struct {
	enabled bool

	// Tool execution metrics.
	toolExecutionDuration *prometheus.HistogramVec
	toolExecutionTotal    *prometheus.CounterVec

	// API request metrics.
	apiRequestDuration *prometheus.HistogramVec
	apiRequestTotal    *prometheus.CounterVec

	// Cache metrics.
	cacheHitTotal  *prometheus.CounterVec
	cacheMissTotal *prometheus.CounterVec

	// Account switching metrics.
	accountSwitchTotal *prometheus.CounterVec

	// Active connections.
	activeConnections *prometheus.GaugeVec

	// Resource count metrics.
	resourceCount *prometheus.GaugeVec
}

// Config holds configuration options for metrics collection.
type Config struct {
	Enabled   bool
	Namespace string
	Subsystem string
	Registry  prometheus.Registerer // Optional custom registry (useful for testing)
}

// DefaultConfig returns a default metrics configuration.
func DefaultConfig() *Config {
	return &Config{
		Enabled:   true,
		Namespace: "cloudmcp",
		Subsystem: "",
		Registry:  prometheus.DefaultRegisterer,
	}
}

// TestConfig returns a configuration suitable for testing with an isolated registry.
func TestConfig() *Config {
	return &Config{
		Enabled:   true,
		Namespace: "cloudmcp_test",
		Subsystem: "",
		Registry:  prometheus.NewRegistry(),
	}
}

// TestConfigWithNamespace returns a test configuration with a custom namespace for better isolation.
func TestConfigWithNamespace(namespace string) *Config {
	return &Config{
		Enabled:   true,
		Namespace: namespace,
		Subsystem: "",
		Registry:  prometheus.NewRegistry(),
	}
}

// NewCollector creates a new metrics collector with the specified configuration.
// If enabled is false, all metric operations become no-ops.
func NewCollector(config *Config) *Collector {
	if config == nil {
		config = DefaultConfig()
	}

	collector := &Collector{
		enabled: config.Enabled,
	}

	// Only initialize metrics if enabled.
	if config.Enabled {
		collector.initializeMetrics(config)
	}

	return collector
}

// initializeMetrics creates all Prometheus metrics with the given configuration.
func (c *Collector) initializeMetrics(config *Config) {
	factory := promauto.With(config.Registry)

	// Tool execution metrics.
	c.toolExecutionDuration = factory.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "tool_execution_duration_seconds",
			Help:      "Duration of tool execution",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"tool", "account", "status"},
	)

	c.toolExecutionTotal = factory.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "tool_execution_total",
			Help:      "Total number of tool executions",
		},
		[]string{"tool", "account", "status"},
	)

	// API request metrics.
	c.apiRequestDuration = factory.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "api_duration_seconds",
			Help:      "Duration of API requests",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status"},
	)

	c.apiRequestTotal = factory.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "api_requests_total",
			Help:      "Total number of API requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// Cache metrics.
	c.cacheHitTotal = factory.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "cache_hits_total",
			Help:      "Total number of cache hits",
		},
		[]string{"cache_type", "account"},
	)

	c.cacheMissTotal = factory.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "cache_misses_total",
			Help:      "Total number of cache misses",
		},
		[]string{"cache_type", "account"},
	)

	// Account switching metrics.
	c.accountSwitchTotal = factory.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "account_switches_total",
			Help:      "Total number of account switches",
		},
		[]string{"from_account", "to_account", "status"},
	)

	// Active connections.
	c.activeConnections = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "active_connections",
			Help:      "Number of active connections per account",
		},
		[]string{"account"},
	)

	// Resource count metrics.
	c.resourceCount = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "resources",
			Help:      "Number of resources by type and account",
		},
		[]string{"resource_type", "account"},
	)
}

// IsEnabled returns whether metrics collection is enabled.
func (c *Collector) IsEnabled() bool {
	return c.enabled
}

// RecordToolExecution records metrics for tool execution.
func (c *Collector) RecordToolExecution(tool, account, status string, duration time.Duration) {
	if !c.enabled || c.toolExecutionDuration == nil {
		return
	}

	c.toolExecutionDuration.WithLabelValues(tool, account, status).Observe(duration.Seconds())
	c.toolExecutionTotal.WithLabelValues(tool, account, status).Inc()
}

// RecordAPIRequest records metrics for API requests.
func (c *Collector) RecordAPIRequest(method, endpoint, status string, duration time.Duration) {
	if !c.enabled || c.apiRequestDuration == nil {
		return
	}

	c.apiRequestDuration.WithLabelValues(method, endpoint, status).Observe(duration.Seconds())
	c.apiRequestTotal.WithLabelValues(method, endpoint, status).Inc()
}

// RecordCacheHit records a cache hit.
func (c *Collector) RecordCacheHit(cacheType, account string) {
	if !c.enabled || c.cacheHitTotal == nil {
		return
	}

	c.cacheHitTotal.WithLabelValues(cacheType, account).Inc()
}

// RecordCacheMiss records a cache miss.
func (c *Collector) RecordCacheMiss(cacheType, account string) {
	if !c.enabled || c.cacheMissTotal == nil {
		return
	}

	c.cacheMissTotal.WithLabelValues(cacheType, account).Inc()
}

// RecordAccountSwitch records an account switch operation.
func (c *Collector) RecordAccountSwitch(fromAccount, toAccount, status string) {
	if !c.enabled || c.accountSwitchTotal == nil {
		return
	}

	c.accountSwitchTotal.WithLabelValues(fromAccount, toAccount, status).Inc()
}

// UpdateActiveConnections updates the number of active connections for an account.
func (c *Collector) UpdateActiveConnections(account string, count int) {
	if !c.enabled || c.activeConnections == nil {
		return
	}

	c.activeConnections.WithLabelValues(account).Set(float64(count))
}

// UpdateResourceCount updates the count of resources for a specific type and account.
func (c *Collector) UpdateResourceCount(resourceType, account string, count int) {
	if !c.enabled || c.resourceCount == nil {
		return
	}

	c.resourceCount.WithLabelValues(resourceType, account).Set(float64(count))
}

// GetMetricsRegistry returns the default Prometheus registry for external access.
func GetMetricsRegistry() *prometheus.Registry {
	registry, ok := prometheus.DefaultRegisterer.(*prometheus.Registry)

	if !ok {
		// This should not happen in normal operation, but we provide a fallback.
		return prometheus.NewRegistry()
	}

	return registry
}
