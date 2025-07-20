// Package metrics provides CloudMCP metrics collection functionality with a unified interface.
//
// This package consolidates metrics collection across different backends (Prometheus, logging, etc.).
// and provides both specific metrics for CloudMCP operations and generic metrics capabilities.
// It eliminates global variables and provides a clean, type-safe interface for metrics collection.
//
// # Architecture.
//
// The package is organized around these key interfaces:.
//   - Provider: Unified metrics collection interface supporting both specific and generic metrics.
//   - Backend: Abstract backend interface allowing multiple metrics implementations.
//   - ProviderFactory: Factory for creating providers with different backends.
//   - BackendFactory: Factory for creating specific backend implementations.
//
// # Supported Backends.
//
//   - Prometheus: Full-featured metrics with histograms, counters, and gauges.
//   - Log: Metrics output to structured logs for development and debugging.
//   - NoOp: Disabled metrics for testing or when metrics are not needed.
//
// # Key Features.
//
//   - Tool execution metrics with timing and status tracking.
//   - API request metrics for cloud provider interactions.
//   - Cache hit/miss ratio tracking.
//   - Account management and switching metrics.
//   - Connection and resource count monitoring.
//   - Provider lifecycle and health metrics.
//   - Generic counter, gauge, histogram, and timing metrics.
//   - Thread-safe operations with proper label handling.
//   - Timer utilities for convenient duration measurement.
//
// # Basic Usage.
//
//	// Create a Prometheus provider.
//	provider, err := metrics.NewPrometheusProvider(nil) // Uses default config.
//	if err != nil {.
//		log.Fatal(err).
//	}.
//
//	// Record tool execution.
//	timer := provider.NewToolExecutionTimer("list_instances", "prod").
//	// ... execute tool ...
//	timer.Finish("success").
//
//	// Record API request.
//	provider.RecordAPIRequest("GET", "/instances", "200", time.Second).
//
//	// Generic metrics.
//	provider.IncrementCounter("custom_events", map[string]string{.
//		"type": "user_action",.
//		"env":  "production",.
//	}).
//
// # Advanced Usage.
//
//	// Custom configuration.
//	config := &metrics.ProviderConfig{.
//		Enabled:   true,.
//		Namespace: "myapp",.
//		Subsystem: "api",.
//		Backend:   metrics.BackendPrometheus,.
//		Tags: map[string]string{.
//			"version": "1.0.0",.
//			"region":  "us-east-1",.
//		},.
//	}.
//
//	provider, err := metrics.NewProvider(config).
//	if err != nil {.
//		log.Fatal(err).
//	}.
//
//	// Use with additional context.
//	apiTimer := provider.NewAPIRequestTimer("POST", "/users").
//	defer apiTimer.FinishWithTags("201", map[string]string{.
//		"user_type": "admin",.
//		"source":    "web",.
//	}).
//
// # Testing.
//
// For testing, use the no-op provider or log provider:.
//
//	// No-op provider (discards all metrics).
//	provider := metrics.NewNoOpProvider().
//
//	// Log provider (outputs to logs for verification).
//	config := &metrics.ProviderConfig{.
//		Enabled: true,.
//		Backend: metrics.BackendLog,.
//	}.
//	provider, _ := metrics.NewProvider(config).
//
// # Thread Safety.
//
// All Provider implementations are thread-safe and can be used concurrently.
// from multiple goroutines. Backend implementations handle proper synchronization.
// for metric registration and updates.
//
// # Performance.
//
// The unified interface adds minimal overhead while providing maximum flexibility.
// Prometheus backends cache metric instances to avoid repeated lookups, and.
// no-op providers have zero overhead when metrics are disabled.
package metrics
