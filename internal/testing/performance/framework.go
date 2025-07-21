// Package performance provides performance testing utilities for CloudMCP
package performance

import (
	"testing"
	"time"
)

// BenchmarkConfig holds configuration for performance tests
type BenchmarkConfig struct {
	MaxDuration time.Duration
	MinOps      int
	Name        string
}

// NewBenchmarkConfig creates a new benchmark configuration
func NewBenchmarkConfig(name string) *BenchmarkConfig {
	return &BenchmarkConfig{
		MaxDuration: 10 * time.Second,
		MinOps:      100,
		Name:        name,
	}
}

// RunBenchmark executes a benchmark with the given configuration
func (c *BenchmarkConfig) RunBenchmark(b *testing.B, fn func()) {
	b.ResetTimer()
	start := time.Now()

	for i := 0; i < b.N; i++ {
		fn()

		// Check if we've exceeded max duration
		if time.Since(start) > c.MaxDuration {
			break
		}
	}

	b.StopTimer()

	// Report if we didn't meet minimum ops
	if b.N < c.MinOps {
		b.Logf("Warning: Only completed %d operations (expected minimum %d)", b.N, c.MinOps)
	}
}

// Measure executes a function and returns its execution duration
func Measure(fn func()) time.Duration {
	start := time.Now()
	fn()
	return time.Since(start)
}
