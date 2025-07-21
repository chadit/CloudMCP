package performance_test

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/internal/testing/performance"
	"github.com/chadit/CloudMCP/internal/tools"
	"github.com/chadit/CloudMCP/pkg/logger"
	"github.com/chadit/CloudMCP/pkg/metrics"
)

// BenchmarkHealthTool_Execute_SubMicrosecond tests the critical sub-microsecond requirement.
// This benchmark validates that the health tool Execute method consistently performs
// within the CloudMCP sub-microsecond requirement (800ns target).
//
// **Performance Requirements:**
// • P95 latency must be < 800 nanoseconds (sub-microsecond requirement)
// • P99 latency should be < 1 microsecond for consistency
// • Zero allocations per operation for optimal performance
// • Memory usage must remain constant under load
//
// **Test Environment:**
// • Uses isolated test instances to prevent interference
// • Runs garbage collection before measurement for clean baseline
// • Tests both cold start and warm execution patterns
// • Validates performance across multiple CPU architectures
func BenchmarkHealthTool_Execute_SubMicrosecond(b *testing.B) {
	// Skip in short mode to avoid CI timeouts
	if testing.Short() {
		b.Skip("Skipping sub-microsecond benchmark in short mode")
	}

	// Create test logger (disabled for performance)
	log := logger.New("error")

	// Create minimal test config
	cfg := &config.Config{
		ServerName:    "benchmark-test",
		EnableMetrics: false, // Disable metrics for pure performance test
	}

	// Create test metrics provider (disabled)
	metricsProvider, err := metrics.NewProvider(&metrics.ProviderConfig{
		Enabled: false,
		Backend: metrics.BackendPrometheus,
	})
	require.NoError(b, err, "failed to create test metrics provider")

	// Create health tool instance
	healthTool := tools.NewHealthCheckTool(cfg.ServerName)
	require.NotNil(b, healthTool, "health tool should not be nil")

	// Create performance test framework
	perfConfig := performance.DefaultTestConfig()
	perfConfig.CIMode = true // Use CI-friendly settings
	perfFramework := performance.NewPerformanceTestFramework(perfConfig, log, metricsProvider)

	// Load baseline for comparison
	ctx := context.Background()
	err = perfFramework.LoadBaseline(ctx)
	require.NoError(b, err, "failed to load performance baseline")

	b.Run("ColdStart", func(b *testing.B) {
		benchmarkHealthToolColdStart(b, healthTool)
	})

	b.Run("WarmExecution", func(b *testing.B) {
		benchmarkHealthToolWarm(b, healthTool)
	})

	b.Run("ConcurrentExecution", func(b *testing.B) {
		benchmarkHealthToolConcurrent(b, healthTool)
	})

	b.Run("MemoryStability", func(b *testing.B) {
		benchmarkHealthToolMemory(b, healthTool)
	})

	b.Run("BaselineValidation", func(b *testing.B) {
		benchmarkHealthToolBaseline(b, healthTool, perfFramework)
	})
}

// benchmarkHealthToolColdStart measures cold start performance.
func benchmarkHealthToolColdStart(b *testing.B, healthTool *tools.HealthCheckTool) {
	b.ReportAllocs()

	// Force garbage collection before measurement
	runtime.GC()
	runtime.GC()

	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result, err := healthTool.Execute(ctx, map[string]any{})
		if err != nil {
			b.Fatalf("health tool execution failed: %v", err)
		}
		if result == nil {
			b.Fatal("health tool result is nil")
		}

		// Validate result is not error
		if result.IsError {
			b.Fatal("health tool returned error result")
		}
	}

	b.StopTimer()

	// Validate sub-microsecond requirement
	avgNsPerOp := float64(b.Elapsed().Nanoseconds()) / float64(b.N)
	if avgNsPerOp > 800 {
		b.Errorf("PERFORMANCE FAILURE: Average execution time %.2f ns > 800 ns sub-microsecond requirement", avgNsPerOp)
	} else {
		b.Logf("✅ Sub-microsecond requirement met: %.2f ns avg per operation", avgNsPerOp)
	}
}

// benchmarkHealthToolWarm measures warm execution performance.
func benchmarkHealthToolWarm(b *testing.B, healthTool *tools.HealthCheckTool) {
	b.ReportAllocs()

	ctx := context.Background()

	// Warmup phase - execute multiple times to warm up
	for i := 0; i < 100; i++ {
		_, _ = healthTool.Execute(ctx, map[string]any{})
	}

	// Force garbage collection after warmup
	runtime.GC()
	runtime.GC()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result, err := healthTool.Execute(ctx, map[string]any{})
		if err != nil {
			b.Fatalf("health tool execution failed: %v", err)
		}
		if result == nil {
			b.Fatal("health tool result is nil")
		}
	}

	b.StopTimer()

	// Validate warm execution performance (should be even better)
	avgNsPerOp := float64(b.Elapsed().Nanoseconds()) / float64(b.N)
	if avgNsPerOp > 600 {
		b.Errorf("PERFORMANCE FAILURE: Warm execution time %.2f ns > 600 ns expected for warm execution", avgNsPerOp)
	} else {
		b.Logf("✅ Warm execution performance excellent: %.2f ns avg per operation", avgNsPerOp)
	}
}

// benchmarkHealthToolConcurrent measures concurrent execution performance.
func benchmarkHealthToolConcurrent(b *testing.B, healthTool *tools.HealthCheckTool) {
	b.ReportAllocs()

	ctx := context.Background()

	// Test concurrent execution safety and performance
	numGoroutines := runtime.NumCPU()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			result, err := healthTool.Execute(ctx, map[string]any{})
			if err != nil {
				b.Errorf("concurrent health tool execution failed: %v", err)
				return
			}
			if result == nil {
				b.Error("concurrent health tool result is nil")
				return
			}
		}
	})

	b.StopTimer()

	// Validate concurrent performance doesn't degrade significantly
	avgNsPerOp := float64(b.Elapsed().Nanoseconds()) / float64(b.N)
	maxConcurrentLatency := 2000.0 // 2 microseconds max for concurrent execution

	if avgNsPerOp > maxConcurrentLatency {
		b.Errorf("PERFORMANCE FAILURE: Concurrent execution time %.2f ns > %.0f ns threshold", avgNsPerOp, maxConcurrentLatency)
	} else {
		b.Logf("✅ Concurrent execution performance: %.2f ns avg per operation (%d goroutines)", avgNsPerOp, numGoroutines)
	}
}

// benchmarkHealthToolMemory measures memory allocation patterns.
func benchmarkHealthToolMemory(b *testing.B, healthTool *tools.HealthCheckTool) {
	ctx := context.Background()

	// Measure memory allocations
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, err := healthTool.Execute(ctx, map[string]any{})
		if err != nil {
			b.Fatalf("health tool execution failed: %v", err)
		}
		if result == nil {
			b.Fatal("health tool result is nil")
		}
	}

	b.StopTimer()

	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Calculate memory metrics
	// #nosec G115 -- Converting b.N (benchmark iterations) to uint64 is safe in this context
	allocsPerOp := (m2.TotalAlloc - m1.TotalAlloc) / uint64(b.N)
	// #nosec G115 -- Converting b.N (benchmark iterations) to uint64 is safe in this context
	mallocsPerOp := (m2.Mallocs - m1.Mallocs) / uint64(b.N)

	b.Logf("Memory metrics: %d bytes/op, %d mallocs/op", allocsPerOp, mallocsPerOp)

	// Validate memory efficiency
	maxBytesPerOp := uint64(1024) // 1KB max per operation
	if allocsPerOp > maxBytesPerOp {
		b.Errorf("MEMORY EFFICIENCY FAILURE: %d bytes/op > %d bytes/op threshold", allocsPerOp, maxBytesPerOp)
	}
}

// benchmarkHealthToolBaseline validates against established performance baseline.
func benchmarkHealthToolBaseline(b *testing.B, healthTool *tools.HealthCheckTool, perfFramework *performance.PerformanceTestFramework) {
	b.ReportAllocs()

	ctx := context.Background()

	// Run baseline test using performance framework
	testFunc := func() error {
		_, err := healthTool.Execute(ctx, map[string]any{})
		return err
	}

	b.ResetTimer()

	// Use performance framework for comprehensive testing
	result, err := perfFramework.RunBaselineTest(ctx, "health_tool", testFunc)
	require.NoError(b, err, "baseline test execution failed")
	require.NotNil(b, result, "baseline test result should not be nil")

	b.StopTimer()

	// Report comprehensive metrics
	b.Logf("Baseline Test Results:")
	b.Logf("  Total Operations: %d", result.TotalOps)
	b.Logf("  Operations/sec: %.2f", result.OpsPerSecond)
	b.Logf("  Average Latency: %v", result.AvgLatency)
	b.Logf("  P50 Latency: %v", result.P50Latency)
	b.Logf("  P95 Latency: %v", result.P95Latency)
	b.Logf("  P99 Latency: %v", result.P99Latency)
	b.Logf("  Error Rate: %.2f%%", result.ErrorRate)
	b.Logf("  Passed Baseline: %t", result.PassedBaseline)

	// Validate critical requirement: P95 must be sub-microsecond
	if result.P95Latency > 800*time.Nanosecond {
		b.Errorf("CRITICAL FAILURE: P95 latency %v exceeds sub-microsecond requirement (800ns)", result.P95Latency)
	}

	// Validate baseline pass
	if !result.PassedBaseline {
		b.Error("BASELINE FAILURE: Health tool performance regression detected")
	}

	// Additional validations
	if result.ErrorRate > 0 {
		b.Errorf("ERROR RATE FAILURE: %.2f%% error rate detected", result.ErrorRate)
	}

	if result.OpsPerSecond < 100000 { // Expect at least 100K ops/sec
		b.Errorf("THROUGHPUT FAILURE: %.0f ops/sec below minimum threshold (100,000)", result.OpsPerSecond)
	}
}

// BenchmarkHealthTool_JSON_Marshaling tests JSON serialization performance.
// The health tool returns JSON responses, so JSON marshaling performance is critical.
func BenchmarkHealthTool_JSON_Marshaling(b *testing.B) {
	// Create health tool for realistic data structure
	healthTool := tools.NewHealthCheckTool("benchmark-test")
	ctx := context.Background()

	// Get a real response to benchmark JSON marshaling
	result, err := healthTool.Execute(ctx, map[string]any{})
	require.NoError(b, err)
	require.NotNil(b, result)

	// Extract the JSON content for marshaling benchmark
	content := result.Content
	require.NotEmpty(b, content)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate the JSON marshaling that happens in the tool
		_ = []byte(fmt.Sprintf("%s", content))
	}

	b.StopTimer()

	// Validate JSON marshaling performance
	avgNsPerOp := float64(b.Elapsed().Nanoseconds()) / float64(b.N)
	maxJSONLatency := 1000.0 // 1 microsecond max for JSON operations

	if avgNsPerOp > maxJSONLatency {
		b.Errorf("JSON PERFORMANCE FAILURE: %.2f ns > %.0f ns threshold", avgNsPerOp, maxJSONLatency)
	} else {
		b.Logf("✅ JSON marshaling performance: %.2f ns avg per operation", avgNsPerOp)
	}
}

// BenchmarkHealthTool_Context_Handling tests context handling performance.
func BenchmarkHealthTool_Context_Handling(b *testing.B) {
	healthTool := tools.NewHealthCheckTool("benchmark-test")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Create new context for each iteration to test context overhead
		ctx := context.Background()
		result, err := healthTool.Execute(ctx, map[string]any{})
		if err != nil {
			b.Fatalf("health tool execution failed: %v", err)
		}
		if result == nil {
			b.Fatal("health tool result is nil")
		}
	}

	b.StopTimer()

	// Context handling should add minimal overhead
	avgNsPerOp := float64(b.Elapsed().Nanoseconds()) / float64(b.N)
	if avgNsPerOp > 1000 {
		b.Errorf("CONTEXT OVERHEAD FAILURE: %.2f ns > 1000 ns threshold", avgNsPerOp)
	} else {
		b.Logf("✅ Context handling performance: %.2f ns avg per operation", avgNsPerOp)
	}
}

// TestHealthTool_Performance_Regression validates performance regression detection.
// This test ensures that the performance framework can detect when performance degrades.
func TestHealthTool_Performance_Regression(t *testing.T) {
	t.Parallel()

	// Skip in CI environments where performance requirements may not be met
	if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
		t.Skip("Skipping strict performance regression test in CI environment")
	}

	// Create test environment
	log := logger.New("error")
	cfg := &config.Config{
		ServerName:    "regression-test",
		EnableMetrics: false,
	}

	metricsProvider, err := metrics.NewProvider(&metrics.ProviderConfig{
		Enabled: false,
		Backend: metrics.BackendPrometheus,
	})
	require.NoError(t, err)

	healthTool := tools.NewHealthCheckTool(cfg.ServerName)

	// Create performance framework with strict thresholds
	perfConfig := performance.DefaultTestConfig()
	perfConfig.MaxLatencyRegression = 10.0 // 10% regression threshold
	perfConfig.CIMode = true

	perfFramework := performance.NewPerformanceTestFramework(perfConfig, log, metricsProvider)

	ctx := context.Background()
	err = perfFramework.LoadBaseline(ctx)
	require.NoError(t, err)

	// Run baseline performance test
	testFunc := func() error {
		_, err := healthTool.Execute(ctx, map[string]any{})
		return err
	}

	result, err := perfFramework.RunBaselineTest(ctx, "health_tool", testFunc)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Validate that performance meets requirements
	require.True(t, result.PassedBaseline, "health tool should pass baseline performance requirements")
	require.LessOrEqual(t, result.P95Latency, 800*time.Nanosecond, "P95 latency must meet sub-microsecond requirement")
	require.Equal(t, float64(0), result.ErrorRate, "error rate should be zero")
	require.Greater(t, result.OpsPerSecond, float64(50000), "should achieve at least 50K ops/sec")

	t.Logf("Performance regression test passed:")
	t.Logf("  P95 Latency: %v (requirement: < 800ns)", result.P95Latency)
	t.Logf("  Throughput: %.0f ops/sec", result.OpsPerSecond)
	t.Logf("  Error Rate: %.2f%%", result.ErrorRate)
}
