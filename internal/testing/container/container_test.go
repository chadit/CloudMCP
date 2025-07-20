package container

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/chadit/CloudMCP/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestContainerFramework validates the container testing framework itself.
func TestContainerFramework(t *testing.T) {
	tests := []struct {
		name        string
		containerTag string
		baseURL     string
		expectError bool
	}{
		{
			name:        "valid_container_test",
			containerTag: "cloudmcp:ci",
			baseURL:     "http://localhost:8080",
			expectError: false,
		},
		{
			name:        "missing_container_tag",
			containerTag: "",
			baseURL:     "http://localhost:8080",
			expectError: false, // Framework should handle this gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test logger
			log := logger.New("debug")
			
			// Create test framework
			tf := NewTestFramework(log, "test-container", tt.baseURL)
			require.NotNil(t, tf, "TestFramework should not be nil")
			
			// Validate framework properties
			assert.Equal(t, "test-container", tf.containerName, "Container name should match")
			assert.Equal(t, tt.baseURL, tf.baseURL, "Base URL should match")
			assert.NotNil(t, tf.httpClient, "HTTP client should not be nil")
			assert.NotNil(t, tf.log, "Logger should not be nil")
		})
	}
}

// TestContainerTestSuite validates the complete container test suite execution.
func TestContainerTestSuite(t *testing.T) {
	// Skip if not in container testing environment
	if os.Getenv("CONTAINER_TEST_MODE") != "true" {
		t.Skip("Skipping container tests - set CONTAINER_TEST_MODE=true to run")
	}

	log := logger.New("debug")
	tf := NewTestFramework(log, "cloudmcp:ci", "http://localhost:8080")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	suite, err := tf.RunTestSuite(ctx, "cloudmcp:ci")
	require.NoError(t, err, "Test suite should execute without error")
	require.NotNil(t, suite, "Test suite result should not be nil")

	// Validate test suite structure
	assert.Equal(t, "CloudMCP Container Validation", suite.Name, "Suite name should match")
	assert.Equal(t, "cloudmcp:ci", suite.ContainerTag, "Container tag should match")
	assert.Greater(t, suite.TotalTests, 0, "Should have executed tests")
	assert.Equal(t, suite.TotalTests, suite.PassedTests+suite.FailedTests, "Test count should be consistent")
	assert.Greater(t, suite.Duration, time.Duration(0), "Suite should have measurable duration")

	// Log test results for debugging
	t.Logf("Container test suite completed:")
	t.Logf("  Total tests: %d", suite.TotalTests)
	t.Logf("  Passed: %d", suite.PassedTests)
	t.Logf("  Failed: %d", suite.FailedTests)
	t.Logf("  Duration: %v", suite.Duration)

	// Validate individual test results
	expectedTests := []string{
		"container_startup",
		"health_check",
		"security_scan",
		"mcp_protocol",
		"performance_metrics",
		"multi_architecture",
		"signal_handling",
		"resource_limits",
	}

	assert.Len(t, suite.Results, len(expectedTests), "Should have results for all expected tests")

	// Check that all expected tests were executed
	testNames := make(map[string]bool)
	for _, result := range suite.Results {
		testNames[result.TestName] = true
		
		// Each test result should have required fields
		assert.NotEmpty(t, result.TestName, "Test name should not be empty")
		assert.Greater(t, result.Duration, time.Duration(0), "Test duration should be positive")
		assert.False(t, result.Timestamp.IsZero(), "Test timestamp should be set")
		
		if !result.Passed {
			t.Logf("Test %s failed: %s", result.TestName, result.Error)
		}
	}

	for _, expectedTest := range expectedTests {
		assert.True(t, testNames[expectedTest], "Expected test %s should have been executed", expectedTest)
	}
}

// TestHealthCheckValidation tests the health check validation functionality.
func TestHealthCheckValidation(t *testing.T) {
	log := logger.New("debug")
	tf := NewTestFramework(log, "test-container", "http://localhost:8080")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result := tf.testHealthCheck(ctx)
	
	// Validate test result structure
	assert.Equal(t, "health_check", result.TestName, "Test name should match")
	assert.False(t, result.Timestamp.IsZero(), "Timestamp should be set")
	assert.Greater(t, result.Duration, time.Duration(0), "Duration should be positive")
	
	// If test failed, it should have an error message
	if !result.Passed {
		assert.NotEmpty(t, result.Error, "Failed test should have error message")
		t.Logf("Health check test failed (expected in unit test): %s", result.Error)
	}
}

// TestSecurityScanValidation tests the security scan functionality.
func TestSecurityScanValidation(t *testing.T) {
	log := logger.New("debug")
	tf := NewTestFramework(log, "test-container", "http://localhost:8080")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	result := tf.testSecurityScan(ctx)
	
	// Validate test result structure
	assert.Equal(t, "security_scan", result.TestName, "Test name should match")
	assert.False(t, result.Timestamp.IsZero(), "Timestamp should be set")
	assert.Greater(t, result.Duration, time.Duration(0), "Duration should be positive")
	
	// In unit test environment, security scan should pass (simulated)
	assert.True(t, result.Passed, "Security scan should pass in unit test")
	
	// Validate details structure if present
	if result.Details != nil {
		details, ok := result.Details.(*SecurityScanResult)
		if ok {
			assert.GreaterOrEqual(t, details.TotalFindings, 0, "Total findings should be non-negative")
			assert.GreaterOrEqual(t, details.HighSeverity, 0, "High severity count should be non-negative")
			assert.GreaterOrEqual(t, details.MediumSeverity, 0, "Medium severity count should be non-negative")
			assert.GreaterOrEqual(t, details.LowSeverity, 0, "Low severity count should be non-negative")
		}
	}
}

// TestMCPProtocolValidation tests the MCP protocol validation functionality.
func TestMCPProtocolValidation(t *testing.T) {
	log := logger.New("debug")
	tf := NewTestFramework(log, "test-container", "http://localhost:8080")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result := tf.testMCPProtocol(ctx)
	
	// Validate test result structure
	assert.Equal(t, "mcp_protocol", result.TestName, "Test name should match")
	assert.False(t, result.Timestamp.IsZero(), "Timestamp should be set")
	assert.Greater(t, result.Duration, time.Duration(0), "Duration should be positive")
	
	// In unit test environment, MCP protocol validation should pass (simulated)
	assert.True(t, result.Passed, "MCP protocol validation should pass in unit test")
	
	// Validate details structure if present
	if result.Details != nil {
		details, ok := result.Details.(*MCPProtocolResult)
		if ok {
			assert.NotEmpty(t, details.ProtocolVersion, "Protocol version should not be empty")
			assert.True(t, details.Compliant, "Protocol should be compliant")
			assert.Greater(t, details.ValidationTime, time.Duration(0), "Validation time should be positive")
		}
	}
}

// TestPerformanceMetricsCollection tests the performance metrics collection.
func TestPerformanceMetricsCollection(t *testing.T) {
	log := logger.New("debug")
	tf := NewTestFramework(log, "test-container", "http://localhost:8080")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result := tf.testPerformanceMetrics(ctx)
	
	// Validate test result structure
	assert.Equal(t, "performance_metrics", result.TestName, "Test name should match")
	assert.False(t, result.Timestamp.IsZero(), "Timestamp should be set")
	assert.Greater(t, result.Duration, time.Duration(0), "Duration should be positive")
	
	// In unit test environment, performance metrics should pass (simulated)
	assert.True(t, result.Passed, "Performance metrics should pass in unit test")
	
	// Validate details structure if present
	if result.Details != nil {
		details, ok := result.Details.(*PerformanceMetrics)
		if ok {
			assert.Greater(t, details.StartupTime, time.Duration(0), "Startup time should be positive")
			assert.Greater(t, details.MemoryUsage, int64(0), "Memory usage should be positive")
			assert.GreaterOrEqual(t, details.CPUUsage, 0.0, "CPU usage should be non-negative")
			assert.Greater(t, details.ResponseTime, time.Duration(0), "Response time should be positive")
			assert.Greater(t, details.RequestsPerSec, 0.0, "Requests per second should be positive")
			assert.Greater(t, details.ImageSize, int64(0), "Image size should be positive")
		}
	}
}

// TestTestResultSerialization validates that test results can be properly serialized.
func TestTestResultSerialization(t *testing.T) {
	// Create sample test result
	result := TestResult{
		TestName:    "test_serialization",
		Passed:      true,
		Duration:    1500 * time.Millisecond,
		Timestamp:   time.Now(),
		ContainerID: "test-container-123",
		Details: map[string]interface{}{
			"status": "success",
			"count":  42,
		},
	}

	// Serialize to JSON
	data, err := json.Marshal(result)
	require.NoError(t, err, "Should serialize without error")
	require.NotEmpty(t, data, "Serialized data should not be empty")

	// Deserialize from JSON
	var deserialized TestResult
	err = json.Unmarshal(data, &deserialized)
	require.NoError(t, err, "Should deserialize without error")

	// Validate deserialized data
	assert.Equal(t, result.TestName, deserialized.TestName, "Test name should match")
	assert.Equal(t, result.Passed, deserialized.Passed, "Passed status should match")
	assert.Equal(t, result.Duration, deserialized.Duration, "Duration should match")
	assert.Equal(t, result.ContainerID, deserialized.ContainerID, "Container ID should match")
	
	// Timestamps should be close (within 1 second due to JSON precision)
	assert.WithinDuration(t, result.Timestamp, deserialized.Timestamp, time.Second, "Timestamps should be close")
}

// TestTestSuiteSerialization validates that test suites can be properly serialized.
func TestTestSuiteSerialization(t *testing.T) {
	// Create sample test suite
	suite := TestSuite{
		Name:         "Test Suite Serialization",
		ContainerTag: "test:latest",
		TotalTests:   2,
		PassedTests:  1,
		FailedTests:  1,
		Duration:     5 * time.Second,
		Timestamp:    time.Now(),
		Results: []TestResult{
			{
				TestName:  "test_pass",
				Passed:    true,
				Duration:  1 * time.Second,
				Timestamp: time.Now(),
			},
			{
				TestName:  "test_fail",
				Passed:    false,
				Duration:  2 * time.Second,
				Error:     "test failed",
				Timestamp: time.Now(),
			},
		},
	}

	// Serialize to JSON
	data, err := json.Marshal(suite)
	require.NoError(t, err, "Should serialize without error")
	require.NotEmpty(t, data, "Serialized data should not be empty")

	// Deserialize from JSON
	var deserialized TestSuite
	err = json.Unmarshal(data, &deserialized)
	require.NoError(t, err, "Should deserialize without error")

	// Validate deserialized data
	assert.Equal(t, suite.Name, deserialized.Name, "Suite name should match")
	assert.Equal(t, suite.ContainerTag, deserialized.ContainerTag, "Container tag should match")
	assert.Equal(t, suite.TotalTests, deserialized.TotalTests, "Total tests should match")
	assert.Equal(t, suite.PassedTests, deserialized.PassedTests, "Passed tests should match")
	assert.Equal(t, suite.FailedTests, deserialized.FailedTests, "Failed tests should match")
	assert.Equal(t, suite.Duration, deserialized.Duration, "Duration should match")
	assert.Len(t, deserialized.Results, len(suite.Results), "Results count should match")
}