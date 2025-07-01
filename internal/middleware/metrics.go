package middleware

import (
	"context"
	"strconv"
	"time"

	"github.com/chadit/CloudMCP/pkg/interfaces"
	pkglogger "github.com/chadit/CloudMCP/pkg/logger"
)

const (
	// Default priorities for metrics middleware.
	defaultMetricsPriority      = 20
	defaultUsageMetricsPriority = 25
)

// MetricsCollector defines the interface for metrics collection.
// This allows different metrics backends (Prometheus, StatsD, etc.) to be used.
type MetricsCollector interface {
	// Counter operations
	IncrementCounter(name string, tags map[string]string)
	IncrementCounterBy(name string, value float64, tags map[string]string)

	// Gauge operations
	SetGauge(name string, value float64, tags map[string]string)

	// Histogram operations
	RecordHistogram(name string, value float64, tags map[string]string)

	// Timing operations
	RecordTiming(name string, duration time.Duration, tags map[string]string)
}

// NoOpMetricsCollector is a metrics collector that does nothing.
// Useful for testing or when metrics are disabled.
type NoOpMetricsCollector struct{}

// IncrementCounter implements MetricsCollector.
func (n *NoOpMetricsCollector) IncrementCounter(_ string, _ map[string]string) {}

// IncrementCounterBy implements MetricsCollector.
func (n *NoOpMetricsCollector) IncrementCounterBy(_ string, _ float64, _ map[string]string) {
}

// SetGauge implements MetricsCollector.
func (n *NoOpMetricsCollector) SetGauge(_ string, _ float64, _ map[string]string) {}

// RecordHistogram implements MetricsCollector.
func (n *NoOpMetricsCollector) RecordHistogram(_ string, _ float64, _ map[string]string) {}

// RecordTiming implements MetricsCollector.
func (n *NoOpMetricsCollector) RecordTiming(_ string, _ time.Duration, _ map[string]string) {
}

// LogBasedMetricsCollector implements MetricsCollector using structured logging.
// This is useful when a dedicated metrics system isn't available.
type LogBasedMetricsCollector struct {
	logger pkglogger.Logger
}

// NewLogBasedMetricsCollector creates a new log-based metrics collector.
func NewLogBasedMetricsCollector(logger pkglogger.Logger) *LogBasedMetricsCollector {
	return &LogBasedMetricsCollector{
		logger: logger,
	}
}

// IncrementCounter logs a counter increment.
func (l *LogBasedMetricsCollector) IncrementCounter(name string, tags map[string]string) {
	l.IncrementCounterBy(name, 1.0, tags)
}

// IncrementCounterBy logs a counter increment with a specific value.
func (l *LogBasedMetricsCollector) IncrementCounterBy(name string, value float64, tags map[string]string) {
	l.logger.Info("Metric counter",
		"metric_type", "counter",
		"metric_name", name,
		"value", value,
		"tags", tags,
	)
}

// SetGauge logs a gauge value.
func (l *LogBasedMetricsCollector) SetGauge(name string, value float64, tags map[string]string) {
	l.logger.Info("Metric gauge",
		"metric_type", "gauge",
		"metric_name", name,
		"value", value,
		"tags", tags,
	)
}

// RecordHistogram logs a histogram value.
func (l *LogBasedMetricsCollector) RecordHistogram(name string, value float64, tags map[string]string) {
	l.logger.Info("Metric histogram",
		"metric_type", "histogram",
		"metric_name", name,
		"value", value,
		"tags", tags,
	)
}

// RecordTiming logs a timing measurement.
func (l *LogBasedMetricsCollector) RecordTiming(name string, duration time.Duration, tags map[string]string) {
	l.logger.Info("Metric timing",
		"metric_type", "timing",
		"metric_name", name,
		"duration_ms", duration.Milliseconds(),
		"tags", tags,
	)
}

// MetricsMiddleware collects performance and usage metrics for tool execution.
type MetricsMiddleware struct {
	*BaseMiddleware
	collector MetricsCollector
}

// NewMetricsMiddleware creates a new metrics collection middleware.
func NewMetricsMiddleware(config *Config, logger pkglogger.Logger, collector MetricsCollector) *MetricsMiddleware {
	if config == nil {
		config = NewConfig().WithPriority(defaultMetricsPriority) // After logging
	}

	if collector == nil {
		collector = &NoOpMetricsCollector{}
	}

	return &MetricsMiddleware{
		BaseMiddleware: NewBaseMiddleware("metrics", config, logger),
		collector:      collector,
	}
}

// Execute implements the Middleware interface for metrics collection.
func (mm *MetricsMiddleware) Execute(ctx context.Context, tool interfaces.Tool, params map[string]any, next ToolHandler) (any, error) {
	if !mm.IsEnabled() {
		return next(ctx, tool, params)
	}

	toolName := tool.Definition().Name

	// Get execution context
	execCtx, hasCtx := GetExecutionContext(ctx)
	if !hasCtx {
		execCtx = NewExecutionContext(toolName, "unknown")
		ctx = WithExecutionContext(ctx, execCtx)
	}

	// Create metric tags
	tags := map[string]string{
		"tool":     toolName,
		"provider": execCtx.Provider,
	}

	// Add user context if available
	if execCtx.UserID != "" {
		tags["user_id"] = execCtx.UserID
	}

	// Record tool execution start
	mm.collector.IncrementCounter("cloudmcp.tool.executions.started", tags)

	// Record tool parameters count
	mm.collector.RecordHistogram("cloudmcp.tool.parameters.count", float64(len(params)), tags)

	// Execute and measure
	startTime := time.Now()
	result, err := next(ctx, tool, params)
	duration := time.Since(startTime)

	// Record execution time
	mm.collector.RecordTiming("cloudmcp.tool.execution.duration", duration, tags)

	// Record completion status
	if err != nil {
		mm.collector.IncrementCounter("cloudmcp.tool.executions.failed", tags)

		// Record error type if available
		errorTags := make(map[string]string)
		for k, v := range tags {
			errorTags[k] = v
		}

		errorTags["error_type"] = mm.categorizeError(err)

		mm.collector.IncrementCounter("cloudmcp.tool.errors", errorTags)
	} else {
		mm.collector.IncrementCounter("cloudmcp.tool.executions.completed", tags)
	}

	// Record performance categories
	mm.recordPerformanceCategory(duration, tags)

	return result, err
}

// categorizeError attempts to categorize the error type for metrics.
func (mm *MetricsMiddleware) categorizeError(err error) string {
	errorStr := err.Error()

	// Simple categorization based on error message
	switch {
	case containsAny(errorStr, []string{"timeout", "deadline", "context canceled"}):
		return "timeout"
	case containsAny(errorStr, []string{"authentication", "unauthorized", "forbidden"}):
		return "auth"
	case containsAny(errorStr, []string{"not found", "404"}):
		return "not_found"
	case containsAny(errorStr, []string{"rate limit", "too many requests", "429"}):
		return "rate_limit"
	case containsAny(errorStr, []string{"validation", "invalid", "bad request", "400"}):
		return "validation"
	case containsAny(errorStr, []string{"network", "connection", "dns"}):
		return "network"
	default:
		return "unknown"
	}
}

// recordPerformanceCategory records metrics based on execution duration.
func (mm *MetricsMiddleware) recordPerformanceCategory(duration time.Duration, baseTags map[string]string) {
	// Create performance category tags
	perfTags := make(map[string]string)
	for k, v := range baseTags {
		perfTags[k] = v
	}

	// Categorize by performance
	switch {
	case duration < 100*time.Millisecond:
		perfTags["performance"] = "fast"
	case duration < 1*time.Second:
		perfTags["performance"] = "normal"
	case duration < 5*time.Second:
		perfTags["performance"] = "slow"
	default:
		perfTags["performance"] = "very_slow"
	}

	mm.collector.IncrementCounter("cloudmcp.tool.performance.category", perfTags)
}

// containsAny checks if a string contains any of the given substrings.
func containsAny(s string, substrings []string) bool {
	for _, substring := range substrings {
		if len(s) >= len(substring) {
			for i := 0; i <= len(s)-len(substring); i++ {
				if s[i:i+len(substring)] == substring {
					return true
				}
			}
		}
	}

	return false
}

// UsageMetricsMiddleware tracks usage patterns and statistics.
type UsageMetricsMiddleware struct {
	*BaseMiddleware
	collector MetricsCollector
}

// NewUsageMetricsMiddleware creates a new usage metrics middleware.
func NewUsageMetricsMiddleware(config *Config, logger pkglogger.Logger, collector MetricsCollector) *UsageMetricsMiddleware {
	if config == nil {
		config = NewConfig().WithPriority(defaultUsageMetricsPriority) // After basic metrics
	}

	if collector == nil {
		collector = &NoOpMetricsCollector{}
	}

	return &UsageMetricsMiddleware{
		BaseMiddleware: NewBaseMiddleware("usage_metrics", config, logger),
		collector:      collector,
	}
}

// Execute implements the Middleware interface for usage metrics collection.
func (umm *UsageMetricsMiddleware) Execute(ctx context.Context, tool interfaces.Tool, params map[string]any, next ToolHandler) (any, error) {
	if !umm.IsEnabled() {
		return next(ctx, tool, params)
	}

	toolName := tool.Definition().Name

	// Get execution context
	execCtx, hasCtx := GetExecutionContext(ctx)
	if !hasCtx {
		execCtx = NewExecutionContext(toolName, "unknown")
		ctx = WithExecutionContext(ctx, execCtx)
	}

	// Create usage tags
	tags := map[string]string{
		"tool":     toolName,
		"provider": execCtx.Provider,
		"hour":     strconv.Itoa(time.Now().Hour()),
		"day":      time.Now().Weekday().String(),
	}

	// Record usage patterns
	umm.collector.IncrementCounter("cloudmcp.usage.tool.invocations", tags)

	// Execute the tool
	result, err := next(ctx, tool, params)

	// Record tool popularity
	popularityTags := map[string]string{
		"tool":     toolName,
		"provider": execCtx.Provider,
	}
	umm.collector.IncrementCounter("cloudmcp.usage.tool.popularity", popularityTags)

	// Record success rate
	if err == nil {
		umm.collector.IncrementCounter("cloudmcp.usage.tool.successes", popularityTags)
	}

	return result, err
}
