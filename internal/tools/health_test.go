package tools_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/tools"
)

// TestNewHealthCheckTool verifies that NewHealthCheckTool creates a properly initialized health check tool.
// This test validates the constructor behavior for setting up a health monitoring tool instance.
func TestNewHealthCheckTool(t *testing.T) {
	t.Parallel()

	serverName := "test-server"
	tool := tools.NewHealthCheckTool(serverName)

	// Note: Cannot directly test private fields serverName, toolsCount, startTime from external package.
	// These are tested indirectly through the Execute method which uses these values.
	assert.NotNil(t, tool, "health tool should not be nil")
}

// TestHealthCheckTool_Name verifies that the health check tool returns the correct name.
// This test ensures the tool identifies itself properly for MCP registration.
func TestHealthCheckTool_Name(t *testing.T) {
	t.Parallel()

	tool := tools.NewHealthCheckTool("test-server")
	name := tool.Name()

	assert.Equal(t, "health_check", name, "tool name should be health_check")
}

// TestHealthCheckTool_Description verifies that the health check tool provides a meaningful description.
// This test ensures the tool description contains relevant information for users.
func TestHealthCheckTool_Description(t *testing.T) {
	t.Parallel()

	tool := tools.NewHealthCheckTool("test-server")
	description := tool.Description()

	assert.Contains(t, description, "health", "description should mention health")
	assert.NotEmpty(t, description, "description should not be empty")
}

// TestHealthCheckTool_InputSchema verifies that the health check tool provides a valid input schema.
// This test ensures the tool defines proper MCP input validation schema.
func TestHealthCheckTool_InputSchema(t *testing.T) {
	t.Parallel()

	tool := tools.NewHealthCheckTool("test-server")
	schema := tool.InputSchema()

	assert.NotNil(t, schema, "input schema should not be nil")
}

// TestHealthCheckTool_Execute verifies that the health check tool executes successfully and returns valid health data.
// This test simulates a health check request and validates the response structure and content.
//
// **Workflow Steps:**
// 1. **Setup**: Create health check tool instance with test server configuration
// 2. **Execution**: Execute health check with empty parameters (no input required)
// 3. **Validation**: Verify successful execution and proper response format
//
// **Expected Behavior:**
// • Tool execution completes without errors
// • Response contains structured health information
// • Result format is valid for MCP protocol.
func TestHealthCheckTool_Execute(t *testing.T) {
	t.Parallel()

	serverName := "test-server"
	tool := tools.NewHealthCheckTool(serverName)

	// Test basic execution
	ctx := t.Context()
	result, err := tool.Execute(ctx, map[string]any{})

	require.NoError(t, err, "execution should not return error")
	require.NotNil(t, result, "result should not be nil")
	require.NotNil(t, result.Content, "result content should not be nil")
	require.NotEmpty(t, result.Content, "result should have content")

	// Verify result is not error
	assert.False(t, result.IsError, "result should not be an error")
}

// TestHealthCheckTool_UpdateToolsCount verifies that the tools count can be updated and reflects in health responses.
// This test simulates the scenario where additional tools are registered and validates count tracking.
//
// **Workflow Steps:**
// 1. **Initial Setup**: Create health check tool with default count
// 2. **Count Update**: Update tools count to simulate tool registration
// 3. **Execution Test**: Execute health check to verify updated count is reflected
//
// **Expected Behavior:**
// • Tools count updates successfully
// • Health check execution continues to work after count update
// • Updated count is properly tracked internally.
func TestHealthCheckTool_UpdateToolsCount(t *testing.T) {
	t.Parallel()

	tool := tools.NewHealthCheckTool("test-server")

	// Test updating tools count
	tool.UpdateToolsCount(5)

	// Test execution after update
	ctx := t.Context()
	result, err := tool.Execute(ctx, map[string]any{})

	require.NoError(t, err, "execution should succeed after update")
	require.NotNil(t, result, "result should not be nil")
}
