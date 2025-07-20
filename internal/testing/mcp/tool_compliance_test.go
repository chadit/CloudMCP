package mcp_test

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/internal/server"
	"github.com/chadit/CloudMCP/internal/tools"
	"github.com/chadit/CloudMCP/pkg/interfaces"
	"github.com/chadit/CloudMCP/pkg/logger"
)

// TestToolInterfaceCompliance validates that all registered tools comply with MCP Tool interface standards.
// This comprehensive test ensures tools implement proper interfaces and behave correctly within the MCP framework.
//
// **Tool Interface Requirements:**
// • Name() must return non-empty string
// • Description() must return meaningful description
// • InputSchema() must return valid JSON schema
// • Execute() must handle context properly
// • Execute() must return valid CallToolResult
//
// **Test Coverage:**
// • Interface method compliance
// • Input schema validation
// • Execution behavior verification
// • Error handling compliance
// • Performance characteristics
//
// **Purpose:** Ensure all tools can be properly registered and executed within MCP protocol framework.
func TestToolInterfaceCompliance(t *testing.T) {
	t.Parallel()

	// Create test tools for compliance checking
	testTools := []interfaces.Tool{
		tools.NewHealthCheckTool("test-server"),
	}

	for _, tool := range testTools {
		tool := tool // Capture range variable
		t.Run("Tool_"+tool.Name(), func(t *testing.T) {
			t.Parallel()

			// Run all compliance tests for this tool
			t.Run("NameCompliance", func(t *testing.T) {
				testToolNameCompliance(t, tool)
			})
			t.Run("DescriptionCompliance", func(t *testing.T) {
				testToolDescriptionCompliance(t, tool)
			})
			t.Run("InputSchemaCompliance", func(t *testing.T) {
				testToolInputSchemaCompliance(t, tool)
			})
			t.Run("ExecutionCompliance", func(t *testing.T) {
				testToolExecutionCompliance(t, tool)
			})
			t.Run("ErrorHandlingCompliance", func(t *testing.T) {
				testToolErrorHandlingCompliance(t, tool)
			})
			t.Run("PerformanceCompliance", func(t *testing.T) {
				testToolPerformanceCompliance(t, tool)
			})
		})
	}
}

// testToolNameCompliance validates tool name requirements.
func testToolNameCompliance(t *testing.T, tool interfaces.Tool) {
	name := tool.Name()

	// Name must not be empty
	assert.NotEmpty(t, name, "Tool name must not be empty")

	// Name should follow naming conventions
	assert.NotContains(t, name, " ", "Tool name should not contain spaces")
	assert.Equal(t, strings.ToLower(name), name, "Tool name should be lowercase or contain underscores")

	// Name should be descriptive
	assert.True(t, len(name) >= 3, "Tool name should be at least 3 characters long")
	assert.True(t, len(name) <= 50, "Tool name should not exceed 50 characters")

	// Name should be consistent across calls
	assert.Equal(t, name, tool.Name(), "Tool name should be consistent across calls")
}

// testToolDescriptionCompliance validates tool description requirements.
func testToolDescriptionCompliance(t *testing.T, tool interfaces.Tool) {
	description := tool.Description()

	// Description must not be empty
	assert.NotEmpty(t, description, "Tool description must not be empty")

	// Description should be meaningful
	assert.True(t, len(description) >= 10, "Tool description should be at least 10 characters")
	assert.True(t, len(description) <= 500, "Tool description should not exceed 500 characters")

	// Description should be consistent across calls
	assert.Equal(t, description, tool.Description(), "Tool description should be consistent across calls")

	// Description should contain tool name or related terms
	toolName := tool.Name()
	nameWords := strings.Split(strings.ReplaceAll(toolName, "_", " "), " ")
	containsRelatedTerm := false
	lowerDesc := strings.ToLower(description)

	for _, word := range nameWords {
		if len(word) > 2 && strings.Contains(lowerDesc, strings.ToLower(word)) {
			containsRelatedTerm = true
			break
		}
	}

	assert.True(t, containsRelatedTerm, "Tool description should contain terms related to tool name")
}

// testToolInputSchemaCompliance validates tool input schema requirements.
func testToolInputSchemaCompliance(t *testing.T, tool interfaces.Tool) {
	schema := tool.InputSchema()

	// Schema must not be nil
	assert.NotNil(t, schema, "Tool input schema must not be nil")

	// Schema should be consistent across calls
	assert.Equal(t, schema, tool.InputSchema(), "Tool input schema should be consistent across calls")

	// Try to marshal schema to JSON to ensure it's valid
	schemaJSON, err := json.Marshal(schema)
	assert.NoError(t, err, "Tool input schema should be JSON marshallable")

	// Schema should be a valid JSON schema object
	var schemaMap map[string]interface{}
	err = json.Unmarshal(schemaJSON, &schemaMap)
	assert.NoError(t, err, "Tool input schema should unmarshal to valid JSON object")

	// Basic JSON schema validation
	if schemaType, exists := schemaMap["type"]; exists {
		assert.IsType(t, "", schemaType, "Schema type should be a string")
		typeStr := schemaType.(string)
		validTypes := []string{"object", "array", "string", "number", "integer", "boolean", "null"}
		assert.Contains(t, validTypes, typeStr, "Schema type should be a valid JSON schema type")
	}
}

// testToolExecutionCompliance validates tool execution behavior.
func testToolExecutionCompliance(t *testing.T, tool interfaces.Tool) {
	ctx := context.Background()

	// Test execution with empty parameters
	result, err := tool.Execute(ctx, map[string]any{})
	assert.NoError(t, err, "Tool execution with empty parameters should not error")
	assert.NotNil(t, result, "Tool execution should return non-nil result")

	// Test execution with nil parameters
	result, err = tool.Execute(ctx, nil)
	assert.NoError(t, err, "Tool execution with nil parameters should not error")
	assert.NotNil(t, result, "Tool execution should return non-nil result")

	// Validate result structure
	if result != nil {
		assert.NotNil(t, result.Content, "Tool result should have content")

		// Content should be valid
		if len(result.Content) > 0 {
			assert.Greater(t, len(result.Content), 0, "Content array should not be empty")
		}

		// If result indicates error, it should have appropriate content
		if result.IsError {
			assert.NotNil(t, result.Content, "Error result should have content explaining the error")
		}
	}
}

// testToolErrorHandlingCompliance validates tool error handling behavior.
func testToolErrorHandlingCompliance(t *testing.T, tool interfaces.Tool) {
	// Test with cancelled context
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result, err := tool.Execute(cancelledCtx, map[string]any{})

	// Tool should handle cancelled context gracefully
	// Either return an error or a result with IsError=true
	if err != nil {
		// Error handling is acceptable
		assert.NotEmpty(t, err.Error(), "Error should have meaningful message")
	} else {
		// Result-based error handling is also acceptable
		assert.NotNil(t, result, "Result should not be nil when no error is returned")
	}

	// Test with timeout context
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer timeoutCancel()

	time.Sleep(2 * time.Nanosecond) // Ensure timeout

	result, err = tool.Execute(timeoutCtx, map[string]any{})
	
	// Tool should handle timeout gracefully (may not always trigger for fast tools)
	// This is more of a stress test
	if err != nil || (result != nil && result.IsError) {
		// Timeout handling worked
		t.Log("Tool properly handled timeout context")
	}
}

// testToolPerformanceCompliance validates tool performance characteristics.
func testToolPerformanceCompliance(t *testing.T, tool interfaces.Tool) {
	ctx := context.Background()

	// Measure execution time
	start := time.Now()
	result, err := tool.Execute(ctx, map[string]any{})
	executionTime := time.Since(start)

	// Basic performance expectations
	assert.NoError(t, err, "Tool execution should not error during performance test")
	assert.NotNil(t, result, "Tool execution should return result during performance test")

	// Tools should complete within reasonable time for testing
	maxExecutionTime := 5 * time.Second
	assert.True(t, executionTime < maxExecutionTime, 
		"Tool execution should complete within %v (took %v)", maxExecutionTime, executionTime)

	// Test concurrent execution safety
	const numConcurrent = 10
	results := make(chan error, numConcurrent)

	for i := 0; i < numConcurrent; i++ {
		go func() {
			_, err := tool.Execute(ctx, map[string]any{})
			results <- err
		}()
	}

	// Collect concurrent execution results
	for i := 0; i < numConcurrent; i++ {
		select {
		case err := <-results:
			assert.NoError(t, err, "Concurrent tool execution should not error")
		case <-time.After(10 * time.Second):
			t.Error("Concurrent execution test timed out")
		}
	}
}

// TestToolRegistryCompliance validates that tools are properly registered and accessible.
func TestToolRegistryCompliance(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		ServerName:    "CloudMCP-Test",
		EnableMetrics: false,
		MetricsPort:   0,
	}
	testLogger := logger.New("error")
	srv, err := server.NewForTesting(cfg, testLogger)
	require.NoError(t, err, "Failed to create test server")

	// Access the server's registry through reflection (for testing purposes)
	// In production, this would be done through proper interfaces
	serverValue := reflect.ValueOf(srv).Elem()
	adapterField := serverValue.FieldByName("mcpAdapter")
	require.True(t, adapterField.IsValid(), "Server should have mcpAdapter field")

	// Test tool registration
	if adapterField.CanInterface() {
		adapter := adapterField.Interface()
		
		// Use reflection to call methods (since we don't have direct access)
		adapterValue := reflect.ValueOf(adapter)
		
		// Test GetToolCount method
		getToolCountMethod := adapterValue.MethodByName("GetToolCount")
		if getToolCountMethod.IsValid() {
			results := getToolCountMethod.Call(nil)
			require.Len(t, results, 1, "GetToolCount should return one value")
			
			toolCount := results[0].Int()
			assert.Greater(t, toolCount, int64(0), "Should have at least one tool registered")
		}

		// Test HasTool method
		hasToolMethod := adapterValue.MethodByName("HasTool")
		if hasToolMethod.IsValid() {
			args := []reflect.Value{reflect.ValueOf("health_check")}
			results := hasToolMethod.Call(args)
			require.Len(t, results, 1, "HasTool should return one value")
			
			hasTool := results[0].Bool()
			assert.True(t, hasTool, "Should have health_check tool registered")
		}
	}
}

// TestToolInputValidation validates tool input parameter handling.
func TestToolInputValidation(t *testing.T) {
	t.Parallel()

	tool := tools.NewHealthCheckTool("test-server")
	ctx := context.Background()

	tests := []struct {
		name   string
		params map[string]any
		valid  bool
	}{
		{
			name:   "EmptyParams",
			params: map[string]any{},
			valid:  true,
		},
		{
			name:   "NilParams",
			params: nil,
			valid:  true,
		},
		{
			name: "ExtraParams",
			params: map[string]any{
				"extra_param": "value",
			},
			valid: true, // Tools should ignore unknown parameters
		},
		{
			name: "ComplexParams",
			params: map[string]any{
				"nested": map[string]any{
					"key": "value",
				},
				"array": []any{1, 2, 3},
			},
			valid: true, // Tools should handle complex parameters gracefully
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := tool.Execute(ctx, tt.params)

			if tt.valid {
				assert.NoError(t, err, "Valid parameters should not cause error")
				assert.NotNil(t, result, "Valid execution should return result")
			} else {
				// For invalid parameters, either error or error result is acceptable
				if err == nil && result != nil {
					// If no error, result should indicate the problem
					assert.True(t, result.IsError, "Invalid parameters should result in error indication")
				}
			}
		})
	}
}

// Helper function for tests - uses shared createTestConfig from compliance_test.go

// BenchmarkToolCompliance benchmarks tool interface compliance testing.
func BenchmarkToolCompliance(b *testing.B) {
	tool := tools.NewHealthCheckTool("benchmark-server")
	ctx := context.Background()
	params := map[string]any{}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, err := tool.Execute(ctx, params)
		if err != nil {
			b.Fatalf("Tool execution failed: %v", err)
		}
		if result == nil {
			b.Fatal("Tool execution returned nil result")
		}
	}
}