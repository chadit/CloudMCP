package linode_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/services/linode"
)

func TestNewMetricsCollector(t *testing.T) {
	t.Parallel()
	// Test enabled collector

	collector := linode.NewMetricsCollector(true)
	require.NotNil(t, collector, "Collector should not be nil")

	// Test disabled collector
	disabledCollector := linode.NewMetricsCollector(false)
	require.NotNil(t, disabledCollector, "Disabled collector should not be nil")
}

func TestMetricsCollector_RecordToolExecution(t *testing.T) {
	t.Parallel()

	collector := linode.NewMetricsCollector(true)

	// Record a successful tool execution
	collector.RecordToolExecution("instances_list", "primary", "success", 100*time.Millisecond)

	// Record another execution with different status
	collector.RecordToolExecution("instances_list", "primary", "error", 50*time.Millisecond)

	// Basic verification that the methods don't panic
	require.NotNil(t, collector, "Collector should remain valid after recording metrics")
}

func TestMetricsCollector_RecordToolExecution_Disabled(t *testing.T) {
	t.Parallel()

	collector := linode.NewMetricsCollector(false)

	// Record tool execution on disabled collector
	collector.RecordToolExecution("instances_list", "primary", "success", 100*time.Millisecond)

	// Should not panic when disabled
	require.NotNil(t, collector, "Collector should remain valid when disabled")
}

func TestMetricsCollector_RecordAPIRequest(t *testing.T) {
	t.Parallel()

	collector := linode.NewMetricsCollector(true)

	// Record an API request
	collector.RecordAPIRequest("GET", "/linode/instances", "200", 200*time.Millisecond)

	// Record a failed request
	collector.RecordAPIRequest("GET", "/linode/instances", "500", 300*time.Millisecond)

	require.NotNil(t, collector, "Collector should remain valid after recording API metrics")
}

func TestMetricsCollector_RecordCache(t *testing.T) {
	t.Parallel()

	collector := linode.NewMetricsCollector(true)

	// Record cache hit
	collector.RecordCacheHit("regions", "primary")

	// Record cache miss
	collector.RecordCacheMiss("types", "primary")

	require.NotNil(t, collector, "Collector should remain valid after recording cache metrics")
}

func TestMetricsCollector_RecordAccountSwitch(t *testing.T) {
	t.Parallel()

	collector := linode.NewMetricsCollector(true)

	// Record successful account switch
	collector.RecordAccountSwitch("primary", "development", "success")

	// Record failed account switch
	collector.RecordAccountSwitch("primary", "nonexistent", "error")

	require.NotNil(t, collector, "Collector should remain valid after recording account switch metrics")
}

func TestMetricsCollector_UpdateActiveConnections(t *testing.T) {
	t.Parallel()

	collector := linode.NewMetricsCollector(true)

	// Update active connections
	collector.UpdateActiveConnections("primary", 5)

	// Update to different value
	collector.UpdateActiveConnections("primary", 3)

	require.NotNil(t, collector, "Collector should remain valid after updating active connections")
}

func TestMetricsCollector_UpdateResourceCount(t *testing.T) {
	t.Parallel()

	collector := linode.NewMetricsCollector(true)

	// Update resource count
	collector.UpdateResourceCount("instances", "primary", 10)

	// Update different resource type
	collector.UpdateResourceCount("volumes", "primary", 5)

	require.NotNil(t, collector, "Collector should remain valid after updating resource counts")
}

func TestToolExecutionTimer(t *testing.T) {
	t.Parallel()

	collector := linode.NewMetricsCollector(true)

	// Create and use timer
	timer := collector.NewToolExecutionTimer("instances_list", "primary")
	require.NotNil(t, timer, "Timer should not be nil")

	// Simulate some work
	time.Sleep(10 * time.Millisecond)

	// Finish timer
	timer.Finish("success")

	// Timer should complete without errors
	require.NotNil(t, timer, "Timer should remain valid after completion")
}

func TestAPIRequestTimer(t *testing.T) {
	t.Parallel()

	collector := linode.NewMetricsCollector(true)

	// Create and use timer
	timer := collector.NewAPIRequestTimer("GET", "/linode/instances")
	require.NotNil(t, timer, "Timer should not be nil")

	// Simulate some work
	time.Sleep(10 * time.Millisecond)

	// Finish timer
	timer.Finish("200")

	require.NotNil(t, timer, "Timer should remain valid after completion")
}

func TestMetricsMiddleware_Success(t *testing.T) {
	t.Parallel()

	collector := linode.NewMetricsCollector(true)

	// Create mock next function that succeeds
	var executedTool, executedAccount string

	nextFunc := func(_ context.Context, tool string, account string) error {
		executedTool = tool
		executedAccount = account

		time.Sleep(10 * time.Millisecond) // Simulate work

		return nil
	}

	middleware := linode.NewMetricsMiddleware(collector, nextFunc)
	require.NotNil(t, middleware, "Middleware should not be nil")

	// Execute through middleware
	ctx := t.Context()
	err := middleware.Execute(ctx, "test_tool", "test_account")

	// Verify execution completed successfully
	require.NoError(t, err, "Middleware execution should not error")
	require.Equal(t, "test_tool", executedTool, "Tool should be passed through")
	require.Equal(t, "test_account", executedAccount, "Account should be passed through")
}

var ErrTest = errors.New("test error")

func TestMetricsMiddleware_Error(t *testing.T) {
	t.Parallel()

	collector := linode.NewMetricsCollector(true)

	// Create mock next function that fails
	expectedError := ErrTest
	nextFunc := func(_ context.Context, _ string, _ string) error {
		time.Sleep(10 * time.Millisecond) // Simulate work

		return expectedError
	}

	middleware := linode.NewMetricsMiddleware(collector, nextFunc)

	// Execute through middleware
	ctx := t.Context()
	err := middleware.Execute(ctx, "test_tool", "test_account")

	// Verify error was returned
	require.Error(t, err, "Middleware should return error")
	require.Equal(t, expectedError, err, "Middleware should return original error")
}

func TestGetMetricsRegistry(t *testing.T) {
	t.Parallel()

	registry := linode.GetMetricsRegistry()

	require.NotNil(t, registry, "Registry should not be nil")
}

func TestMetricsCollector_DisabledOperations(t *testing.T) {
	t.Parallel()

	// Test that all operations work correctly when disabled

	collector := linode.NewMetricsCollector(false)

	// These should all execute without panics or errors
	collector.RecordToolExecution("tool", "account", "status", time.Second)
	collector.RecordAPIRequest("GET", "/endpoint", "200", time.Second)
	collector.RecordCacheHit("type", "account")
	collector.RecordCacheMiss("type", "account")
	collector.RecordAccountSwitch("from", "to", "success")
	collector.UpdateActiveConnections("account", 5)
	collector.UpdateResourceCount("type", "account", 10)

	// Timers should also work when disabled
	toolTimer := collector.NewToolExecutionTimer("tool", "account")
	require.NotNil(t, toolTimer, "Tool timer should be created even when disabled")
	toolTimer.Finish("success")

	apiTimer := collector.NewAPIRequestTimer("GET", "/endpoint")
	require.NotNil(t, apiTimer, "API timer should be created even when disabled")
	apiTimer.Finish("200")

	// Middleware should work when disabled
	nextFunc := func(_ context.Context, _ string, _ string) error {
		return nil
	}
	middleware := linode.NewMetricsMiddleware(collector, nextFunc)

	err := middleware.Execute(t.Context(), "tool", "account")

	require.NoError(t, err, "Middleware should work when metrics are disabled")
}
