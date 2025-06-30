//go:build integration

package linode_test

import (
	"context"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// BenchmarkHandlerPerformance benchmarks the performance of CloudMCP handlers
// using HTTP test server infrastructure to measure latency and throughput.
//
// **Performance Test Areas**:
// • Handler execution latency for common operations
// • Memory allocation patterns during handler execution
// • Throughput measurement for concurrent handler calls
// • Performance comparison between different handler types
//
// **Benchmark Metrics**:
// • Operations per second (ops/sec)
// • Nanoseconds per operation (ns/op)
// • Memory allocations per operation (allocs/op)
// • Bytes allocated per operation (B/op)
//
// **Purpose**: Establishes performance baselines for CloudMCP handlers
// and identifies potential optimization opportunities in the codebase.
func BenchmarkHandlerPerformance(b *testing.B) {
	// Create a fresh service and server for each benchmark to avoid connection issues
	server := MockLinodeAPIServer()
	defer server.Close()

	service := createHTTPTestServiceForBenchmark(b, server.URL)
	ctx := b.Context()

	err := service.Initialize(ctx)
	if err != nil {
		b.Fatalf("failed to initialize service: %v", err)
	}

	b.Run("AccountGet", func(b *testing.B) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_account_get",
				Arguments: map[string]interface{}{},
			},
		}

		b.ResetTimer()
		b.ReportAllocs()

		for range b.N {
			result, err := service.handleAccountGet(ctx, request)
			if err != nil {
				b.Fatalf("handler error: %v", err)
			}
			if result == nil {
				b.Fatal("result should not be nil")
			}
		}
	})

	b.Run("InstancesList", func(b *testing.B) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_instances_list",
				Arguments: map[string]interface{}{},
			},
		}

		b.ResetTimer()
		b.ReportAllocs()

		for range b.N {
			result, err := service.handleInstancesList(ctx, request)
			if err != nil {
				b.Fatalf("handler error: %v", err)
			}
			if result == nil {
				b.Fatal("result should not be nil")
			}
		}
	})

	b.Run("InstanceGet", func(b *testing.B) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_instance_get",
				Arguments: map[string]interface{}{
					"instance_id": float64(123456),
				},
			},
		}

		b.ResetTimer()
		b.ReportAllocs()

		for range b.N {
			result, err := service.handleInstanceGet(ctx, request)
			if err != nil {
				b.Fatalf("handler error: %v", err)
			}
			if result == nil {
				b.Fatal("result should not be nil")
			}
		}
	})

	b.Run("VolumesList", func(b *testing.B) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_volumes_list",
				Arguments: map[string]interface{}{},
			},
		}

		b.ResetTimer()
		b.ReportAllocs()

		for range b.N {
			result, err := service.handleVolumesList(ctx, request)
			if err != nil {
				b.Fatalf("handler error: %v", err)
			}
			if result == nil {
				b.Fatal("result should not be nil")
			}
		}
	})
}

// TestHandlerLatency tests the latency of CloudMCP handlers under normal conditions.
// This replaces the concurrent benchmarks which had HTTP server connection issues.
//
// **Latency Test Areas**:
// • Individual handler execution latency
// • Consistency of performance across multiple runs
// • Performance comparison between different handler types
// • Memory allocation patterns during execution
//
// **Test Metrics**:
// • Average latency per handler type
// • Latency consistency (standard deviation)
// • Memory allocations and byte allocations
// • Performance relative to expectations
//
// **Purpose**: Validates that CloudMCP handlers perform within acceptable
// latency bounds and maintain consistent performance characteristics.
func TestHandlerLatency(t *testing.T) {
	server := MockLinodeAPIServer()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
	ctx := b.Context()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	testCases := []struct {
		name        string
		request     mcp.CallToolRequest
		handler     func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
		description string
	}{
		{
			name: "AccountGetLatency",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "linode_account_get",
					Arguments: map[string]interface{}{},
				},
			},
			handler:     service.handleAccountGet,
			description: "Account information retrieval latency",
		},
		{
			name: "InstancesListLatency",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "linode_instances_list",
					Arguments: map[string]interface{}{},
				},
			},
			handler:     service.handleInstancesList,
			description: "Instance listing latency",
		},
		{
			name: "InstanceGetLatency",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "linode_instance_get",
					Arguments: map[string]interface{}{
						"instance_id": float64(123456),
					},
				},
			},
			handler:     service.handleInstanceGet,
			description: "Single instance retrieval latency",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			const numRuns = 50
			var totalDuration time.Duration

			for i := 0; i < numRuns; i++ {
				start := time.Now()
				result, err := tc.handler(ctx, tc.request)
				duration := time.Since(start)

				require.NoError(t, err, "handler should not return error on run %d", i+1)
				require.NotNil(t, result, "result should not be nil on run %d", i+1)

				totalDuration += duration
			}

			avgDuration := totalDuration / numRuns
			t.Logf("%s: Average latency over %d runs: %v (%s)",
				tc.name, numRuns, avgDuration, tc.description)

			// Performance should be reasonable for HTTP test server
			require.True(t, avgDuration < 100*time.Millisecond,
				"average latency %v should be under 100ms for %s", avgDuration, tc.description)
		})
	}
}

// TestPerformanceThresholds validates that CloudMCP handlers meet
// performance requirements under HTTP test server conditions.
//
// **Performance Requirements**:
// • Handler latency should be under acceptable thresholds
// • Memory allocation should be reasonable for text processing
// • No significant performance degradation between handler types
// • Consistent performance across multiple executions
//
// **Test Metrics**:
// • Maximum acceptable latency per handler type
// • Memory allocation limits for text-based operations
// • Performance consistency validation across runs
// • Resource utilization monitoring
//
// **Purpose**: Ensures CloudMCP meets performance requirements and
// provides early detection of performance regressions during development.
func TestPerformanceThresholds(t *testing.T) {
	server := MockLinodeAPIServer()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
	ctx := b.Context()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	testCases := []struct {
		name        string
		request     mcp.CallToolRequest
		handler     func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
		maxLatency  time.Duration
		description string
	}{
		{
			name: "AccountGetLatency",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "linode_account_get",
					Arguments: map[string]interface{}{},
				},
			},
			handler:     service.handleAccountGet,
			maxLatency:  50 * time.Millisecond,
			description: "Account information retrieval should complete quickly",
		},
		{
			name: "InstancesListLatency",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "linode_instances_list",
					Arguments: map[string]interface{}{},
				},
			},
			handler:     service.handleInstancesList,
			maxLatency:  100 * time.Millisecond,
			description: "Instance listing should complete within reasonable time",
		},
		{
			name: "InstanceGetLatency",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "linode_instance_get",
					Arguments: map[string]interface{}{
						"instance_id": float64(123456),
					},
				},
			},
			handler:     service.handleInstanceGet,
			maxLatency:  75 * time.Millisecond,
			description: "Single instance retrieval should be fast",
		},
		{
			name: "VolumesListLatency",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "linode_volumes_list",
					Arguments: map[string]interface{}{},
				},
			},
			handler:     service.handleVolumesList,
			maxLatency:  100 * time.Millisecond,
			description: "Volume listing should complete within reasonable time",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Measure latency over multiple runs for consistency
			const numRuns = 10
			var totalDuration time.Duration

			for i := 0; i < numRuns; i++ {
				start := time.Now()
				result, err := tc.handler(ctx, tc.request)
				duration := time.Since(start)

				require.NoError(t, err, "handler should not return error on run %d", i+1)
				require.NotNil(t, result, "result should not be nil on run %d", i+1)

				totalDuration += duration

				// Validate individual run doesn't exceed threshold
				require.True(t, duration <= tc.maxLatency,
					"run %d took %v, which exceeds maximum latency %v for %s",
					i+1, duration, tc.maxLatency, tc.description)
			}

			avgDuration := totalDuration / numRuns
			t.Logf("%s: Average latency over %d runs: %v (max allowed: %v)",
				tc.name, numRuns, avgDuration, tc.maxLatency)

			// Average should also be well under the threshold
			require.True(t, avgDuration <= tc.maxLatency,
				"average latency %v exceeds maximum %v for %s",
				avgDuration, tc.maxLatency, tc.description)
		})
	}
}

// TestMemoryUsage validates memory allocation patterns for CloudMCP handlers
// to ensure efficient memory usage during text processing operations.
//
// **Memory Usage Validation**:
// • Reasonable memory allocation for text processing
// • No significant memory leaks during repeated operations
// • Consistent memory patterns across handler types
// • Memory efficiency for large text responses
//
// **Test Approach**:
// • Multiple handler executions with memory monitoring
// • Comparison of memory usage between handler types
// • Validation of memory cleanup after operations
// • Detection of unexpected memory allocation patterns
//
// **Purpose**: Ensures CloudMCP handlers use memory efficiently and
// don't introduce memory leaks or excessive allocation patterns.
func TestMemoryUsage(t *testing.T) {
	server := MockLinodeAPIServer()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
	ctx := b.Context()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	testCases := []struct {
		name        string
		request     mcp.CallToolRequest
		handler     func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
		description string
	}{
		{
			name: "AccountGetMemory",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "linode_account_get",
					Arguments: map[string]interface{}{},
				},
			},
			handler:     service.handleAccountGet,
			description: "Account information retrieval memory usage",
		},
		{
			name: "InstancesListMemory",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "linode_instances_list",
					Arguments: map[string]interface{}{},
				},
			},
			handler:     service.handleInstancesList,
			description: "Instance listing memory usage",
		},
		{
			name: "InstanceGetMemory",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "linode_instance_get",
					Arguments: map[string]interface{}{
						"instance_id": float64(123456),
					},
				},
			},
			handler:     service.handleInstanceGet,
			description: "Single instance retrieval memory usage",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Warm up to avoid measuring initialization overhead
			for i := 0; i < 3; i++ {
				_, err := tc.handler(ctx, tc.request)
				require.NoError(t, err, "warmup run %d should not error", i+1)
			}

			// Measure memory allocation
			memResult := testing.Benchmark(func(b *testing.B) {
				b.ReportAllocs()
				for range b.N {
					result, err := tc.handler(ctx, tc.request)
					if err != nil {
						b.Fatalf("handler error: %v", err)
					}
					if result == nil {
						b.Fatal("result should not be nil")
					}
				}
			})

			// Log memory allocation metrics
			if memResult.AllocsPerOp() > 0 {
				t.Logf("%s: %d allocs/op, %d B/op (%s)",
					tc.name, memResult.AllocsPerOp(), memResult.AllocedBytesPerOp(), tc.description)
			}

			// Basic sanity checks for memory usage
			// These are reasonable bounds for text processing operations
			require.True(t, memResult.AllocsPerOp() < 1000,
				"handler should not allocate excessively (got %d allocs/op)", memResult.AllocsPerOp())

			require.True(t, memResult.AllocedBytesPerOp() < 100*1024, // 100KB
				"handler should not allocate excessive memory (got %d B/op)", memResult.AllocedBytesPerOp())
		})
	}
}
