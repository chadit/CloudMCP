// Package metrics provides CloudMCP metrics collection functionality.
//
// This package encapsulates all Prometheus metrics used throughout the CloudMCP
// application, eliminating global variables and providing a clean interface
// for metrics collection.
//
// Key components:
//   - Collector: Main metrics collection struct that holds all Prometheus metrics
//   - Config: Configuration for metrics collection including namespace and subsystem
//   - Timers: Convenient timing utilities for tool execution and API requests
//   - Middleware: Integration with execution pipelines for automatic metrics collection
//
// Example usage:
//
//	config := metrics.DefaultConfig()
//	config.Namespace = "myapp"
//	collector := metrics.NewCollector(config)
//
//	// Record a tool execution
//	timer := collector.NewToolExecutionTimer("list_instances", "prod")
//	// ... execute tool ...
//	timer.Finish("success")
//
//	// Record API request
//	collector.RecordAPIRequest("GET", "/instances", "200", time.Second).
package metrics
