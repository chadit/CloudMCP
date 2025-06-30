package linode

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Tool execution metrics.
	toolExecutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cloudmcp_tool_execution_duration_seconds",
			Help:    "Duration of tool execution",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"tool", "account", "status"},
	)

	toolExecutionTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cloudmcp_tool_execution_total",
			Help: "Total number of tool executions",
		},
		[]string{"tool", "account", "status"},
	)

	// API request metrics.
	apiRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cloudmcp_linode_api_duration_seconds",
			Help:    "Duration of Linode API requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status"},
	)

	apiRequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cloudmcp_linode_api_requests_total",
			Help: "Total number of Linode API requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// Cache metrics.
	cacheHitTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cloudmcp_cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_type", "account"},
	)

	cacheMissTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cloudmcp_cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_type", "account"},
	)

	// Account switching metrics.
	accountSwitchTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cloudmcp_account_switches_total",
			Help: "Total number of account switches",
		},
		[]string{"from_account", "to_account", "status"},
	)

	// Active connections.
	activeConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cloudmcp_active_connections",
			Help: "Number of active connections per account",
		},
		[]string{"account"},
	)

	// Resource count metrics.
	resourceCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cloudmcp_resources",
			Help: "Number of resources by type and account",
		},
		[]string{"resource_type", "account"},
	)
)

// MetricsCollector provides methods for recording various CloudMCP metrics.
type MetricsCollector struct {
	enabled bool
}

// NewMetricsCollector creates a new metrics collector.
func NewMetricsCollector(enabled bool) *MetricsCollector {
	return &MetricsCollector{
		enabled: enabled,
	}
}

// RecordToolExecution records metrics for tool execution.
func (m *MetricsCollector) RecordToolExecution(tool, account, status string, duration time.Duration) {
	if !m.enabled {
		return
	}

	toolExecutionDuration.WithLabelValues(tool, account, status).Observe(duration.Seconds())
	toolExecutionTotal.WithLabelValues(tool, account, status).Inc()
}

// RecordAPIRequest records metrics for Linode API requests.
func (m *MetricsCollector) RecordAPIRequest(method, endpoint, status string, duration time.Duration) {
	if !m.enabled {
		return
	}

	apiRequestDuration.WithLabelValues(method, endpoint, status).Observe(duration.Seconds())
	apiRequestTotal.WithLabelValues(method, endpoint, status).Inc()
}

// RecordCacheHit records a cache hit.
func (m *MetricsCollector) RecordCacheHit(cacheType, account string) {
	if !m.enabled {
		return
	}

	cacheHitTotal.WithLabelValues(cacheType, account).Inc()
}

// RecordCacheMiss records a cache miss.
func (m *MetricsCollector) RecordCacheMiss(cacheType, account string) {
	if !m.enabled {
		return
	}

	cacheMissTotal.WithLabelValues(cacheType, account).Inc()
}

// RecordAccountSwitch records an account switch operation.
func (m *MetricsCollector) RecordAccountSwitch(fromAccount, toAccount, status string) {
	if !m.enabled {
		return
	}

	accountSwitchTotal.WithLabelValues(fromAccount, toAccount, status).Inc()
}

// UpdateActiveConnections updates the number of active connections for an account.
func (m *MetricsCollector) UpdateActiveConnections(account string, count int) {
	if !m.enabled {
		return
	}

	activeConnections.WithLabelValues(account).Set(float64(count))
}

// UpdateResourceCount updates the count of resources for a specific type and account.
func (m *MetricsCollector) UpdateResourceCount(resourceType, account string, count int) {
	if !m.enabled {
		return
	}

	resourceCount.WithLabelValues(resourceType, account).Set(float64(count))
}

// ToolExecutionTimer provides a convenient way to time tool executions.
type ToolExecutionTimer struct {
	metrics   *MetricsCollector
	tool      string
	account   string
	startTime time.Time
}

// NewToolExecutionTimer creates a new timer for tool execution.
func (m *MetricsCollector) NewToolExecutionTimer(tool, account string) *ToolExecutionTimer {
	return &ToolExecutionTimer{
		metrics:   m,
		tool:      tool,
		account:   account,
		startTime: time.Now(),
	}
}

// Finish records the tool execution metrics with the specified status.
func (t *ToolExecutionTimer) Finish(status string) {
	duration := time.Since(t.startTime)
	t.metrics.RecordToolExecution(t.tool, t.account, status, duration)
}

// APIRequestTimer provides a convenient way to time API requests.
type APIRequestTimer struct {
	metrics   *MetricsCollector
	method    string
	endpoint  string
	startTime time.Time
}

// NewAPIRequestTimer creates a new timer for API requests.
func (m *MetricsCollector) NewAPIRequestTimer(method, endpoint string) *APIRequestTimer {
	return &APIRequestTimer{
		metrics:   m,
		method:    method,
		endpoint:  endpoint,
		startTime: time.Now(),
	}
}

// Finish records the API request metrics with the specified status.
func (t *APIRequestTimer) Finish(status string) {
	duration := time.Since(t.startTime)
	t.metrics.RecordAPIRequest(t.method, t.endpoint, status, duration)
}

// MetricsMiddleware wraps tool execution with metrics collection.
type MetricsMiddleware struct {
	metrics *MetricsCollector
	next    func(ctx context.Context, tool string, account string) error
}

// NewMetricsMiddleware creates a middleware that collects metrics for tool execution.
func NewMetricsMiddleware(metrics *MetricsCollector, next func(ctx context.Context, tool string, account string) error) *MetricsMiddleware {
	return &MetricsMiddleware{
		metrics: metrics,
		next:    next,
	}
}

// Execute wraps the execution with metrics collection.
func (m *MetricsMiddleware) Execute(ctx context.Context, tool string, account string) error {
	timer := m.metrics.NewToolExecutionTimer(tool, account)

	err := m.next(ctx, tool, account)

	status := "success"
	if err != nil {
		status = "error"
	}

	timer.Finish(status)

	return err
}

// GetMetricsRegistry returns the default Prometheus registry for external access.
func GetMetricsRegistry() *prometheus.Registry {
	return prometheus.DefaultRegisterer.(*prometheus.Registry)
}
