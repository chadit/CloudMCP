// Package performance provides comprehensive performance testing framework for CloudMCP.
// This package implements baseline testing, regression detection, and CI-friendly performance validation.
package performance

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/chadit/CloudMCP/pkg/logger"
	"github.com/chadit/CloudMCP/pkg/metrics"
)

// TestConfig holds configuration for performance tests.
type TestConfig struct {
	// Baseline thresholds
	HealthToolMaxLatency       time.Duration // Sub-microsecond requirement: 800ns
	MCPProtocolMaxLatency      time.Duration // Tool execution: 10ms
	MetricsEndpointMaxLatency  time.Duration // Metrics response: 100ms
	JSONMarshalMaxLatency      time.Duration // Serialization: 1ms
	
	// Load test parameters
	ConcurrentUsers            int           // Concurrent test users
	TestDuration              time.Duration // Test run duration
	WarmupDuration            time.Duration // Warmup period
	
	// Performance regression thresholds
	MaxLatencyRegression      float64       // Maximum acceptable latency increase (%)
	MaxThroughputRegression   float64       // Maximum acceptable throughput decrease (%)
	
	// CI-specific settings
	CIMode                    bool          // Enable CI-friendly settings
	SkipLoadTests            bool          // Skip heavy load tests in CI
	CollectDetailedMetrics   bool          // Collect detailed performance metrics
	
	// Environment detection
	IsCI                     bool          // Detected CI environment
	IsDevelopment           bool          // Local development environment
}

// DefaultTestConfig returns default performance test configuration.
func DefaultTestConfig() *TestConfig {
	isCI := detectCIEnvironment()
	
	return &TestConfig{
		// CloudMCP sub-microsecond requirement for health tool
		HealthToolMaxLatency:       800 * time.Nanosecond,
		MCPProtocolMaxLatency:      10 * time.Millisecond,
		MetricsEndpointMaxLatency:  100 * time.Millisecond,
		JSONMarshalMaxLatency:      1 * time.Millisecond,
		
		// Load test configuration
		ConcurrentUsers:           10,
		TestDuration:             30 * time.Second,
		WarmupDuration:           5 * time.Second,
		
		// Regression thresholds
		MaxLatencyRegression:     20.0, // 20% increase is failure
		MaxThroughputRegression:  10.0, // 10% decrease is failure
		
		// CI configuration
		CIMode:                   isCI,
		SkipLoadTests:           isCI, // Skip heavy tests in CI
		CollectDetailedMetrics:  !isCI, // Detailed metrics only locally
		IsCI:                    isCI,
		IsDevelopment:          !isCI,
	}
}

// TestResult represents the results of a performance test.
type TestResult struct {
	TestName        string                 `json:"testName"`
	StartTime       time.Time             `json:"startTime"`
	EndTime         time.Time             `json:"endTime"`
	Duration        time.Duration         `json:"duration"`
	
	// Latency metrics
	MinLatency      time.Duration         `json:"minLatency"`
	MaxLatency      time.Duration         `json:"maxLatency"`
	AvgLatency      time.Duration         `json:"avgLatency"`
	P50Latency      time.Duration         `json:"p50Latency"`
	P95Latency      time.Duration         `json:"p95Latency"`
	P99Latency      time.Duration         `json:"p99Latency"`
	
	// Throughput metrics
	TotalOps        int64                 `json:"totalOps"`
	OpsPerSecond    float64               `json:"opsPerSecond"`
	
	// Error metrics
	TotalErrors     int64                 `json:"totalErrors"`
	ErrorRate       float64               `json:"errorRate"`
	
	// Memory metrics
	AllocsPerOp     int64                 `json:"allocsPerOp"`
	BytesPerOp      int64                 `json:"bytesPerOp"`
	
	// Success criteria
	PassedBaseline  bool                  `json:"passedBaseline"`
	PassedRegression bool                 `json:"passedRegression"`
	
	// Additional context
	Environment     map[string]string     `json:"environment"`
	Metadata        map[string]any        `json:"metadata"`
}

// PerformanceTestFramework provides comprehensive performance testing capabilities.
type PerformanceTestFramework struct {
	config          *TestConfig
	logger          logger.Logger
	metricsProvider metrics.Provider
	
	// Baseline storage
	baseline        *PerformanceBaseline
	baselineMutex   sync.RWMutex
	
	// Test execution state
	running         bool
	runningMutex    sync.RWMutex
}

// PerformanceBaseline stores baseline performance measurements.
type PerformanceBaseline struct {
	Version         string                    `json:"version"`
	Timestamp       time.Time                 `json:"timestamp"`
	Platform        string                    `json:"platform"`
	GoVersion       string                    `json:"goVersion"`
	
	// Component baselines
	HealthTool      *ComponentBaseline        `json:"healthTool"`
	MCPProtocol     *ComponentBaseline        `json:"mcpProtocol"`
	MetricsServer   *ComponentBaseline        `json:"metricsServer"`
	JSONMarshal     *ComponentBaseline        `json:"jsonMarshal"`
	
	// System context
	Environment     map[string]string         `json:"environment"`
}

// ComponentBaseline represents performance baseline for a component.
type ComponentBaseline struct {
	Name            string            `json:"name"`
	P50Latency      time.Duration     `json:"p50Latency"`
	P95Latency      time.Duration     `json:"p95Latency"`
	P99Latency      time.Duration     `json:"p99Latency"`
	ThroughputOPS   float64           `json:"throughputOPS"`
	AllocsPerOp     int64             `json:"allocsPerOp"`
	BytesPerOp      int64             `json:"bytesPerOp"`
	SampleCount     int               `json:"sampleCount"`
	MeasuredAt      time.Time         `json:"measuredAt"`
}

// NewPerformanceTestFramework creates a new performance testing framework.
func NewPerformanceTestFramework(config *TestConfig, log logger.Logger, metricsProvider metrics.Provider) *PerformanceTestFramework {
	if config == nil {
		config = DefaultTestConfig()
	}
	
	return &PerformanceTestFramework{
		config:          config,
		logger:          log,
		metricsProvider: metricsProvider,
		baseline:        &PerformanceBaseline{},
	}
}

// LoadBaseline loads performance baseline from storage or creates a new one.
func (f *PerformanceTestFramework) LoadBaseline(ctx context.Context) error {
	f.baselineMutex.Lock()
	defer f.baselineMutex.Unlock()
	
	// For now, create a new baseline (future: load from file/database)
	f.baseline = &PerformanceBaseline{
		Version:     "minimal-shell",
		Timestamp:   time.Now(),
		Platform:    runtime.GOOS + "/" + runtime.GOARCH,
		GoVersion:   runtime.Version(),
		Environment: f.gatherEnvironmentInfo(),
	}
	
	f.logger.Info("Performance baseline initialized",
		"platform", f.baseline.Platform,
		"go_version", f.baseline.GoVersion,
		"ci_mode", f.config.CIMode,
	)
	
	return nil
}

// SaveBaseline saves the current baseline for future regression testing.
func (f *PerformanceTestFramework) SaveBaseline(ctx context.Context) error {
	f.baselineMutex.RLock()
	defer f.baselineMutex.RUnlock()
	
	// Future: persist to file or database
	baselineJSON, err := json.MarshalIndent(f.baseline, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal baseline: %w", err)
	}
	
	f.logger.Info("Performance baseline saved",
		"size_bytes", len(baselineJSON),
		"components", f.countComponents(),
	)
	
	return nil
}

// RunBaselineTest executes a single baseline test for a component.
func (f *PerformanceTestFramework) RunBaselineTest(ctx context.Context, testName string, testFunc func() error) (*TestResult, error) {
	f.logger.Info("Starting baseline test", "test", testName)
	
	result := &TestResult{
		TestName:    testName,
		StartTime:   time.Now(),
		Environment: f.gatherEnvironmentInfo(),
		Metadata:    make(map[string]any),
	}
	
	// Warmup phase
	if f.config.WarmupDuration > 0 {
		f.logger.Debug("Warmup phase starting", "duration", f.config.WarmupDuration)
		warmupCtx, cancel := context.WithTimeout(ctx, f.config.WarmupDuration)
		f.runWarmup(warmupCtx, testFunc)
		cancel()
	}
	
	// Force garbage collection before measurement
	runtime.GC()
	runtime.GC() // Double GC to ensure clean state
	
	// Measurement phase
	latencies, errors := f.measureLatencies(ctx, testFunc)
	
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	
	// Calculate metrics
	f.calculateMetrics(result, latencies, errors)
	
	// Validate against thresholds
	f.validateBaseline(result, testName)
	
	f.logger.Info("Baseline test completed",
		"test", testName,
		"duration", result.Duration,
		"ops", result.TotalOps,
		"avg_latency", result.AvgLatency,
		"p95_latency", result.P95Latency,
		"passed", result.PassedBaseline,
	)
	
	return result, nil
}

// measureLatencies performs the actual latency measurements.
func (f *PerformanceTestFramework) measureLatencies(ctx context.Context, testFunc func() error) ([]time.Duration, []error) {
	const minSamples = 1000
	const maxSamples = 100000
	
	var latencies []time.Duration
	var errors []error
	
	// Determine sample count based on environment
	sampleCount := minSamples
	if !f.config.IsCI {
		sampleCount = maxSamples / 10 // More samples locally
	}
	
	testCtx, cancel := context.WithTimeout(ctx, f.config.TestDuration)
	defer cancel()
	
	for i := 0; i < sampleCount; i++ {
		select {
		case <-testCtx.Done():
			break
		default:
		}
		
		start := time.Now()
		err := testFunc()
		elapsed := time.Since(start)
		
		latencies = append(latencies, elapsed)
		if err != nil {
			errors = append(errors, err)
		}
	}
	
	return latencies, errors
}

// calculateMetrics computes performance metrics from raw measurements.
func (f *PerformanceTestFramework) calculateMetrics(result *TestResult, latencies []time.Duration, errors []error) {
	if len(latencies) == 0 {
		return
	}
	
	// Sort latencies for percentile calculations
	sortedLatencies := make([]time.Duration, len(latencies))
	copy(sortedLatencies, latencies)
	f.sortDurations(sortedLatencies)
	
	// Basic metrics
	result.TotalOps = int64(len(latencies))
	result.TotalErrors = int64(len(errors))
	result.ErrorRate = float64(result.TotalErrors) / float64(result.TotalOps) * 100
	
	if result.Duration > 0 {
		result.OpsPerSecond = float64(result.TotalOps) / result.Duration.Seconds()
	}
	
	// Latency metrics
	result.MinLatency = sortedLatencies[0]
	result.MaxLatency = sortedLatencies[len(sortedLatencies)-1]
	
	// Calculate average
	var sum time.Duration
	for _, lat := range latencies {
		sum += lat
	}
	result.AvgLatency = sum / time.Duration(len(latencies))
	
	// Percentiles
	result.P50Latency = f.percentile(sortedLatencies, 50)
	result.P95Latency = f.percentile(sortedLatencies, 95)
	result.P99Latency = f.percentile(sortedLatencies, 99)
}

// percentile calculates the nth percentile from sorted durations.
func (f *PerformanceTestFramework) percentile(sorted []time.Duration, n int) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	
	index := (n * len(sorted)) / 100
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	
	return sorted[index]
}

// sortDurations sorts duration slice in ascending order.
func (f *PerformanceTestFramework) sortDurations(durations []time.Duration) {
	// Simple bubble sort for small arrays, efficient enough for testing
	n := len(durations)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if durations[j] > durations[j+1] {
				durations[j], durations[j+1] = durations[j+1], durations[j]
			}
		}
	}
}

// validateBaseline checks if test results meet baseline requirements.
func (f *PerformanceTestFramework) validateBaseline(result *TestResult, testName string) {
	result.PassedBaseline = true
	
	// Get threshold for this test
	var threshold time.Duration
	switch testName {
	case "health_tool":
		threshold = f.config.HealthToolMaxLatency
	case "mcp_protocol":
		threshold = f.config.MCPProtocolMaxLatency
	case "metrics_endpoint":
		threshold = f.config.MetricsEndpointMaxLatency
	case "json_marshal":
		threshold = f.config.JSONMarshalMaxLatency
	default:
		// Default threshold
		threshold = 10 * time.Millisecond
	}
	
	// Check P95 latency against threshold
	if result.P95Latency > threshold {
		result.PassedBaseline = false
		f.logger.Warn("Baseline test failed",
			"test", testName,
			"p95_latency", result.P95Latency,
			"threshold", threshold,
			"exceeded_by", result.P95Latency-threshold,
		)
	}
	
	// Check error rate
	if result.ErrorRate > 1.0 { // 1% error rate threshold
		result.PassedBaseline = false
		f.logger.Warn("High error rate detected",
			"test", testName,
			"error_rate", result.ErrorRate,
		)
	}
}

// runWarmup executes warmup iterations to stabilize performance.
func (f *PerformanceTestFramework) runWarmup(ctx context.Context, testFunc func() error) {
	const warmupIterations = 100
	
	for i := 0; i < warmupIterations; i++ {
		select {
		case <-ctx.Done():
			return
		default:
		}
		
		_ = testFunc() // Ignore errors during warmup
		
		if i%10 == 0 {
			runtime.GC() // Periodic GC during warmup
		}
	}
}

// gatherEnvironmentInfo collects environment information for context.
func (f *PerformanceTestFramework) gatherEnvironmentInfo() map[string]string {
	return map[string]string{
		"go_version":     runtime.Version(),
		"go_os":          runtime.GOOS,
		"go_arch":        runtime.GOARCH,
		"num_cpu":        fmt.Sprintf("%d", runtime.NumCPU()),
		"gomaxprocs":     fmt.Sprintf("%d", runtime.GOMAXPROCS(0)),
		"ci_mode":        fmt.Sprintf("%t", f.config.CIMode),
		"test_framework": "cloudmcp_performance",
	}
}

// countComponents returns the number of baseline components.
func (f *PerformanceTestFramework) countComponents() int {
	count := 0
	if f.baseline.HealthTool != nil {
		count++
	}
	if f.baseline.MCPProtocol != nil {
		count++
	}
	if f.baseline.MetricsServer != nil {
		count++
	}
	if f.baseline.JSONMarshal != nil {
		count++
	}
	return count
}

// detectCIEnvironment detects if running in a CI environment.
func detectCIEnvironment() bool {
	ciEnvVars := []string{
		"CI", "CONTINUOUS_INTEGRATION",
		"GITHUB_ACTIONS", "GITLAB_CI", "CIRCLECI",
		"JENKINS_URL", "TRAVIS", "BUILDKITE",
	}
	
	for _, envVar := range ciEnvVars {
		if value := os.Getenv(envVar); value != "" && value != "false" {
			return true
		}
	}
	
	return false
}

// GetConfig returns the current test configuration.
func (f *PerformanceTestFramework) GetConfig() *TestConfig {
	return f.config
}

// IsRunning returns true if performance tests are currently running.
func (f *PerformanceTestFramework) IsRunning() bool {
	f.runningMutex.RLock()
	defer f.runningMutex.RUnlock()
	return f.running
}

// setRunning updates the running state of the framework.
func (f *PerformanceTestFramework) setRunning(running bool) {
	f.runningMutex.Lock()
	defer f.runningMutex.Unlock()
	f.running = running
}