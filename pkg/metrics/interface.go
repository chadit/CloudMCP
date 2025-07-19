package metrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Backend type constants.
	BackendPrometheus = "prometheus"
	BackendNoOp       = "noop"
	BackendLog        = "log"
)

// CoreMetrics defines basic metrics operations available to all components.
type CoreMetrics interface {
	// IsEnabled returns whether metrics collection is enabled for this provider.
	IsEnabled() bool

	// Generic metrics operations for flexible metric collection.
	IncrementCounter(name string, tags map[string]string)
	IncrementCounterBy(name string, value float64, tags map[string]string)
	SetGauge(name string, value float64, tags map[string]string)
	RecordHistogram(name string, value float64, tags map[string]string)
	RecordTiming(name string, duration time.Duration, tags map[string]string)
}

// ToolMetrics defines metrics operations specific to tool execution tracking.
type ToolMetrics interface {
	// Tool execution metrics - record tool performance and outcomes.
	RecordToolExecution(tool, account, status string, duration time.Duration)
	NewToolExecutionTimer(tool, account string) Timer
}

// CloudMetrics defines metrics operations specific to cloud provider interactions.
type CloudMetrics interface {
	// API request metrics - record cloud provider API interactions.
	RecordAPIRequest(method, endpoint, status string, duration time.Duration)
	NewAPIRequestTimer(method, endpoint string) Timer

	// Cache metrics - record cache hit/miss ratios.
	RecordCacheHit(cacheType, account string)
	RecordCacheMiss(cacheType, account string)

	// Account management metrics - record account switching operations.
	RecordAccountSwitch(fromAccount, toAccount, status string)

	// Connection tracking - monitor active connections per account.
	UpdateActiveConnections(account string, count int)

	// Resource tracking - monitor cloud resource counts by type and account.
	UpdateResourceCount(resourceType, account string, count int)
}

// LifecycleMetrics defines metrics operations for tracking provider lifecycle and health.
type LifecycleMetrics interface {
	// Provider lifecycle metrics - track provider initialization and health.
	RecordProviderLifecycle(provider, event, status string, duration time.Duration)
	UpdateProviderHealth(provider string, healthy bool)
}

// Provider defines a comprehensive metrics interface through composition of focused interfaces.
// This design follows the Interface Segregation Principle for better maintainability.
type Provider interface {
	CoreMetrics
	ToolMetrics
	CloudMetrics
	LifecycleMetrics
}

// Timer provides a unified interface for timing operations.
// It encapsulates the start time and automatically records metrics when finished.
// This interface serves both tool executions and API requests to eliminate redundancy.
type Timer interface {
	// Finish records the operation duration with the specified status.
	// For tool executions: "success", "error", "timeout", "cancelled"
	// For API requests: HTTP status codes or "error" for network issues
	Finish(status string)

	// FinishWithTags records the operation with additional context tags.
	FinishWithTags(status string, tags map[string]string)
}

// ToolTimer is an alias for Timer to maintain backward compatibility for tool timing.
type ToolTimer = Timer

// APITimer is an alias for Timer to maintain backward compatibility for API timing.
type APITimer = Timer

// Backend defines the interface for different metrics backend implementations.
// This allows the metrics system to support multiple backends (Prometheus, StatsD, etc.).
type Backend interface {
	// Backend identification.
	Name() string
	IsEnabled() bool

	// Core metric operations that backends must implement.
	Counter(name string, tags map[string]string) CounterMetric
	Gauge(name string, tags map[string]string) GaugeMetric
	Histogram(name string, tags map[string]string) HistogramMetric
	Timer(name string, tags map[string]string) TimerMetric

	// Lifecycle management.
	Start() error
	Stop() error
	Health() error
}

// CounterMetric represents a monotonically increasing counter.
type CounterMetric interface {
	Inc()
	Add(value float64)
}

// GaugeMetric represents a gauge that can go up or down.
type GaugeMetric interface {
	Set(value float64)
	Inc()
	Dec()
	Add(value float64)
	Sub(value float64)
}

// HistogramMetric represents a histogram for recording distributions.
type HistogramMetric interface {
	Observe(value float64)
	ObserveDuration(duration time.Duration)
}

// TimerMetric represents a timer for measuring durations.
type TimerMetric interface {
	Start() TimerInstance
	Record(duration time.Duration)
}

// TimerInstance represents an active timing measurement.
type TimerInstance interface {
	Stop()
	Elapsed() time.Duration
}

// ProviderConfig holds configuration for metrics providers.
type ProviderConfig struct {
	// Enabled controls whether metrics collection is active.
	Enabled bool

	// Namespace is the top-level namespace for all metrics (e.g., "cloudmcp").
	Namespace string

	// Subsystem is an optional subsystem name (e.g., "provider", "api").
	Subsystem string

	// Backend specifies the metrics backend to use.
	Backend string

	// Tags are default tags applied to all metrics from this provider.
	Tags map[string]string

	// BackendConfig contains backend-specific configuration.
	BackendConfig map[string]any
}

// DefaultProviderConfig returns a default metrics provider configuration.
func DefaultProviderConfig() *ProviderConfig {
	return &ProviderConfig{
		Enabled:       true,
		Namespace:     "cloudmcp",
		Subsystem:     "",
		Backend:       BackendPrometheus,
		Tags:          make(map[string]string),
		BackendConfig: make(map[string]any),
	}
}

// PrometheusBackendConfig holds Prometheus-specific configuration options.
type PrometheusBackendConfig struct {
	// Registry is the Prometheus registry to use (optional, defaults to default registry).
	Registry prometheus.Registerer

	// DefaultBuckets are the histogram buckets to use for timing metrics.
	DefaultBuckets []float64

	// EnableDefaultMetrics controls whether to register default Go metrics.
	EnableDefaultMetrics bool

	// MetricsPath is the HTTP path for metrics endpoint (default: "/metrics").
	MetricsPath string

	// MetricsPort is the port for the metrics HTTP server (0 = disabled).
	MetricsPort int
}

// DefaultPrometheusBackendConfig returns default Prometheus configuration.
func DefaultPrometheusBackendConfig() *PrometheusBackendConfig {
	return &PrometheusBackendConfig{
		Registry:             prometheus.DefaultRegisterer,
		DefaultBuckets:       prometheus.DefBuckets,
		EnableDefaultMetrics: true,
		MetricsPath:          "/metrics",
		MetricsPort:          0, // Disabled by default
	}
}

// TestPrometheusBackendConfig returns a Prometheus configuration suitable for testing.
func TestPrometheusBackendConfig() *PrometheusBackendConfig {
	return &PrometheusBackendConfig{
		Registry:             prometheus.NewRegistry(),
		DefaultBuckets:       prometheus.DefBuckets,
		EnableDefaultMetrics: false,
		MetricsPath:          "/metrics",
		MetricsPort:          0,
	}
}

// ProviderFactory creates metrics providers with the specified configuration.
type ProviderFactory interface {
	// CreateProvider creates a new metrics provider with the given configuration.
	CreateProvider(config *ProviderConfig) (Provider, error)

	// SupportedBackends returns a list of supported backend names.
	SupportedBackends() []string

	// ValidateConfig validates the provider configuration for the specified backend.
	ValidateConfig(config *ProviderConfig) error
}

// BackendFactory creates metrics backends for specific implementations.
type BackendFactory interface {
	// CreateBackend creates a new metrics backend with the given configuration.
	CreateBackend(config *ProviderConfig) (Backend, error)

	// BackendName returns the name of the backend this factory creates.
	BackendName() string

	// ValidateConfig validates the configuration for this backend type.
	ValidateConfig(config *ProviderConfig) error
}

// NewProvider creates a new metrics provider using the specified configuration.
// This is the main entry point for creating metrics providers in the application.
//
//nolint:ireturn // NewProvider implements Abstract Factory pattern - must return interface for backend abstraction.
func NewProvider(config *ProviderConfig) (Provider, error) {
	return NewProviderWithFactory(config, DefaultProviderFactory())
}

// NewProviderWithFactory creates a new metrics provider using the specified configuration and factory.
// This allows explicit factory management without global state.
//
//nolint:ireturn // NewProviderWithFactory enables Dependency Injection pattern - interface return required for testability.
func NewProviderWithFactory(config *ProviderConfig, factory ProviderFactory) (Provider, error) {
	provider, err := factory.CreateProvider(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics provider: %w", err)
	}

	return provider, nil
}

// NewPrometheusProvider creates a Prometheus-based metrics provider with the given configuration.
// This is a convenience function for creating Prometheus providers directly.
//
//nolint:ireturn // NewPrometheusProvider is a specialized factory function - returns interface for backend polymorphism.
func NewPrometheusProvider(config *ProviderConfig) (Provider, error) {
	if config == nil {
		config = DefaultProviderConfig()
	}

	config.Backend = BackendPrometheus

	return NewProvider(config)
}

// NewNoOpProvider creates a no-operation metrics provider that discards all metrics.
// This is useful for testing or when metrics collection should be disabled.
//
//nolint:ireturn // NewNoOpProvider implements Null Object pattern - interface return maintains API consistency.
func NewNoOpProvider() Provider {
	config := &ProviderConfig{
		Enabled: false,
		Backend: BackendNoOp,
	}

	provider, _ := NewProvider(config)

	return provider
}

// GetProviderFactory returns the default provider factory.
// This function creates a new default factory each time to avoid global state.
//
//nolint:ireturn // GetProviderFactory provides factory abstraction - interface return enables multiple factory implementations.
func GetProviderFactory() ProviderFactory {
	return DefaultProviderFactory()
}

// SetProviderFactory is deprecated in favor of explicit factory management.
// Use NewProviderWithFactory or dependency injection instead.
// This function is maintained for backward compatibility but does nothing.
func SetProviderFactory(_ ProviderFactory) {
	// No-op to maintain backward compatibility while eliminating global state.
	// Users should migrate to explicit factory management patterns.
}
