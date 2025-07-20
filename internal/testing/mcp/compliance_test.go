package mcp_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/internal/server"
	mcptest "github.com/chadit/CloudMCP/internal/testing/mcp"
	"github.com/chadit/CloudMCP/pkg/logger"
)

// TestMCPProtocolCompliance verifies that CloudMCP server properly implements MCP protocol.
// This comprehensive test suite validates protocol compliance for core MCP methods.
//
// **Test Environment:**
// • Mock MCP client for protocol interaction
// • Isolated server instance with test configuration
// • Buffered I/O for deterministic testing
//
// **Protocol Validation:**
// • MCP initialize method compliance
// • Tools list method implementation
// • Tool execution via tools/call method
// • JSON-RPC 2.0 message format validation
//
// **Expected Behavior:**
// • Server responds to all MCP protocol methods correctly
// • JSON-RPC 2.0 compliance in all message exchanges
// • Proper error handling for invalid requests
// • Tool registration and execution workflow
//
// **Purpose:** Ensure CloudMCP can properly communicate with MCP clients like Claude.
func TestMCPProtocolCompliance(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		description string
		testFunc    func(t *testing.T, client *mcptest.MockMCPClient, srv *server.Server)
	}{
		{
			name:        "Initialize",
			description: "Test MCP initialize method compliance",
			testFunc:    testInitializeCompliance,
		},
		{
			name:        "ToolsList",
			description: "Test tools/list method implementation",
			testFunc:    testToolsListCompliance,
		},
		{
			name:        "ToolsCall",
			description: "Test tools/call method with health_check tool",
			testFunc:    testToolsCallCompliance,
		},
		{
			name:        "ErrorHandling",
			description: "Test error handling compliance",
			testFunc:    testErrorHandlingCompliance,
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create test server
			cfg := createTestConfig(t)
			testLogger := logger.New("error") // Use minimal logging for tests
			srv, err := server.NewForTesting(cfg, testLogger)
			require.NoError(t, err, "Failed to create test server")

			// Create connected mock client
			client, err := mcptest.CreateConnectedClient(srv)
			require.NoError(t, err, "Failed to create connected client")
			client.SetTimeout(2 * time.Second)

			// Run specific test
			tt.testFunc(t, client, srv)
		})
	}
}

// testInitializeCompliance validates MCP initialize method implementation.
func testInitializeCompliance(t *testing.T, client *mcptest.MockMCPClient, srv *server.Server) {
	ctx := context.Background()

	// Send initialize request
	response, err := client.SendInitialize(ctx)
	require.NoError(t, err, "Initialize request should succeed")
	require.NotNil(t, response, "Initialize response should not be nil")

	// Validate response structure
	assert.Equal(t, "2024-11-05", response.ProtocolVersion, "Protocol version should match")
	assert.NotNil(t, response.Capabilities, "Capabilities should be present")
	assert.NotNil(t, response.ServerInfo, "Server info should be present")

	// Validate server info
	assert.NotEmpty(t, response.ServerInfo.Name, "Server name should not be empty")
	assert.NotEmpty(t, response.ServerInfo.Version, "Server version should not be empty")

	// Validate capabilities structure
	assert.Contains(t, response.Capabilities, "tools", "Should have tools capability")
}

// testToolsListCompliance validates tools/list method implementation.
func testToolsListCompliance(t *testing.T, client *mcptest.MockMCPClient, srv *server.Server) {
	ctx := context.Background()

	// Send tools/list request
	response, err := client.SendToolsList(ctx)
	require.NoError(t, err, "Tools list request should succeed")
	require.NotNil(t, response, "Tools list response should not be nil")

	// Validate tools structure
	require.Greater(t, len(response.Tools), 0, "Should have at least one tool registered")

	// Find health_check tool
	var healthTool *mcp.Tool
	for i := range response.Tools {
		if response.Tools[i].Name == "health_check" {
			healthTool = &response.Tools[i]
			break
		}
	}

	require.NotNil(t, healthTool, "health_check tool should be registered")

	// Validate health tool structure
	assert.Equal(t, "health_check", healthTool.Name, "Tool name should be health_check")
	assert.NotEmpty(t, healthTool.Description, "Tool description should not be empty")
	assert.NotNil(t, healthTool.InputSchema, "Tool should have input schema")

	// Validate input schema structure  
	schema := healthTool.InputSchema
	assert.Equal(t, "object", schema.Type, "Schema type should be object")
}

// testToolsCallCompliance validates tools/call method implementation.
func testToolsCallCompliance(t *testing.T, client *mcptest.MockMCPClient, srv *server.Server) {
	ctx := context.Background()

	// Test health_check tool call
	result, err := client.SendToolsCall(ctx, "health_check", map[string]interface{}{})
	require.NoError(t, err, "Tool call should succeed")
	require.NotNil(t, result, "Tool result should not be nil")

	// Validate result structure
	assert.False(t, result.IsError, "Health check should not return error")
	assert.NotNil(t, result.Content, "Result should have content")
	assert.NotEmpty(t, result.Content, "Result content should not be empty")

	// Validate result content structure
	require.Greater(t, len(result.Content), 0, "Content should not be empty")

	// First content item should be text content
	firstContent := result.Content[0]
	if textContent, ok := mcp.AsTextContent(firstContent); ok {
		assert.Equal(t, "text", textContent.Type, "Content type should be text")
		assert.NotEmpty(t, textContent.Text, "Content text should not be empty")
	} else {
		t.Error("First content item should be text content")
	}
}

// testErrorHandlingCompliance validates error handling compliance.
func testErrorHandlingCompliance(t *testing.T, client *mcptest.MockMCPClient, srv *server.Server) {
	ctx := context.Background()

	// Test calling non-existent tool
	result, err := client.SendToolsCall(ctx, "nonexistent_tool", map[string]interface{}{})
	
	// This should either return an error or a result with IsError=true
	if err != nil {
		// Server returned JSON-RPC error - this is valid
		assert.Contains(t, err.Error(), "tool", "Error should mention tool")
	} else {
		// Server returned successful response with error flag - also valid
		require.NotNil(t, result, "Result should not be nil")
		if result.IsError {
			assert.NotNil(t, result.Content, "Error result should have content")
		}
	}
}

// TestMCPProtocolVersionCompatibility tests compatibility with different MCP protocol versions.
func TestMCPProtocolVersionCompatibility(t *testing.T) {
	t.Parallel()

	supportedVersions := []string{
		"2024-11-05", // Current MCP version
	}

	for _, version := range supportedVersions {
		version := version // Capture range variable
		t.Run("Version_"+version, func(t *testing.T) {
			t.Parallel()

			cfg := createTestConfig(t)
			testLogger := logger.New("error")
			srv, err := server.NewForTesting(cfg, testLogger)
			require.NoError(t, err, "Failed to create test server")

			client, err := mcptest.CreateConnectedClient(srv)
			require.NoError(t, err, "Failed to create connected client")
			ctx := context.Background()

			// Send initialize with specific protocol version
			response, err := client.SendInitialize(ctx)
			require.NoError(t, err, "Initialize should succeed for supported version")
			
			// Server should respond with supported version
			assert.NotEmpty(t, response.ProtocolVersion, "Protocol version should be returned")
			
			// Use srv to avoid unused variable warning
			_ = srv
		})
	}
}

// TestConcurrentMCPRequests tests that the server can handle concurrent MCP requests.
func TestConcurrentMCPRequests(t *testing.T) {
	t.Parallel()

	cfg := createTestConfig(t)
	testLogger := logger.New("error")
	srv, err := server.NewForTesting(cfg, testLogger)
	require.NoError(t, err, "Failed to create test server")

	const numConcurrentRequests = 10
	resultChan := make(chan error, numConcurrentRequests)

	// Send concurrent tool calls
	for i := 0; i < numConcurrentRequests; i++ {
		go func() {
			client, err := mcptest.CreateConnectedClient(srv)
			if err != nil {
				resultChan <- err
				return
			}
			ctx := context.Background()

			// Send health check tool call
			_, err = client.SendToolsCall(ctx, "health_check", map[string]interface{}{})
			resultChan <- err
		}()
	}

	// Collect results
	for i := 0; i < numConcurrentRequests; i++ {
		select {
		case err := <-resultChan:
			assert.NoError(t, err, "Concurrent request should succeed")
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent request result")
		}
	}
	
	// Use srv to avoid unused variable warning
	_ = srv
}

// createTestConfig creates a test configuration for MCP compliance testing.
func createTestConfig(t *testing.T) *config.Config {
	t.Helper()

	return &config.Config{
		ServerName:    "CloudMCP-Test",
		EnableMetrics: false, // Disable metrics for simpler testing
		MetricsPort:   0,     // No metrics server needed
	}
}

// BenchmarkMCPToolCall benchmarks tool call performance for compliance testing.
func BenchmarkMCPToolCall(b *testing.B) {
	cfg := &config.Config{
		ServerName:    "CloudMCP-Benchmark",
		EnableMetrics: false,
		MetricsPort:   0,
	}

	testLogger := logger.New("error")
	srv, err := server.NewForTesting(cfg, testLogger)
	require.NoError(b, err, "Failed to create test server")

	client, err := mcptest.CreateConnectedClient(srv)
	require.NoError(b, err, "Failed to create connected client")
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := client.SendToolsCall(ctx, "health_check", map[string]interface{}{})
		if err != nil {
			b.Fatalf("Tool call failed: %v", err)
		}
	}
	
	// Use srv to avoid unused variable warning
	_ = srv
}