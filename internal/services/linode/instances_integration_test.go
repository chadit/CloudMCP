//go:build integration

package linode_test

import (
	"github.com/chadit/CloudMCP/internal/services/linode"
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// TestInstancesListIntegration tests the complete instances list workflow through MCP.
// This integration test verifies that the CloudMCP service can successfully communicate
// with a mock Linode API server and return properly formatted instance data through
// the MCP protocol interface.
//
// **Integration Test Workflow**:
// 1. **Container Setup**: Start WireMock container with Linode API stubs
// 2. **Service Creation**: Initialize CloudMCP service with mock API endpoint
// 3. **MCP Request**: Execute instances list tool through MCP interface
// 4. **Response Validation**: Verify JSON structure and data completeness
// 5. **Contract Verification**: Ensure response matches Linode API contract
//
// **Test Environment**: Mock Linode API with predefined instance data
//
// **Expected Behavior**:
// • Returns valid MCP tool result with JSON content
// • JSON contains expected instance fields (id, label, status, region, etc.)
// • Response structure matches real Linode API format
// • All required fields are present and correctly typed
//
// **Purpose**: This test ensures end-to-end integration between CloudMCP's MCP
// interface and Linode API communication works correctly with realistic data.
func TestInstancesListIntegration(t *testing.T) {
	service, cleanup := linode.SetupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_instances_list",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.handleInstancesList(ctx, request)
	require.NoError(t, err, "instances list should not return error")
	require.NotNil(t, result, "result should not be nil")
	require.NotEmpty(t, result.Content, "result should have content")

	// Extract and validate JSON response
	textContent, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok, "content should be text")

	var response map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &response)
	require.NoError(t, err, "response should be valid JSON")

	// Validate response structure matches Linode API contract
	require.Contains(t, response, "data", "response should contain data field")
	require.Contains(t, response, "page", "response should contain page field")
	require.Contains(t, response, "pages", "response should contain pages field")
	require.Contains(t, response, "results", "response should contain results field")

	// Validate instance data structure
	data, ok := response["data"].([]interface{})
	require.True(t, ok, "data should be an array")
	require.Len(t, data, 1, "should have one instance")

	instance := data[0].(map[string]interface{})
	ValidateJSONStructure(t, instance, map[string]string{
		"id":      "number",
		"label":   "string",
		"region":  "string",
		"image":   "string",
		"type":    "string",
		"status":  "string",
		"ipv4":    "array",
		"ipv6":    "string",
		"created": "string",
		"updated": "string",
		"specs":   "object",
		"alerts":  "object",
		"backups": "object",
	})

	// Validate specific values from stub data
	require.Equal(t, float64(123456), instance["id"], "instance ID should match stub")
	require.Equal(t, "test-instance-1", instance["label"], "instance label should match stub")
	require.Equal(t, "running", instance["status"], "instance status should match stub")
}

// TestInstanceGetIntegration tests getting a specific instance through MCP interface.
// This integration test verifies that the CloudMCP service can retrieve detailed
// information for a specific instance using the mock Linode API.
//
// **Integration Test Workflow**:
// 1. **Service Setup**: Initialize service with mock API container
// 2. **MCP Request**: Execute instance get tool with specific instance ID
// 3. **Response Validation**: Verify detailed instance data structure
// 4. **Field Verification**: Ensure all instance fields are present and valid
//
// **Test Scenario**: Retrieve instance with ID 123456 (defined in stub data)
//
// **Expected Behavior**:
// • Returns detailed instance information in MCP format
// • All instance fields are properly structured and typed
// • Response matches expected Linode API instance object format
// • Specific field values match predefined stub data
//
// **Purpose**: Validates that CloudMCP can successfully retrieve and format
// detailed instance information through the MCP protocol interface.
func TestInstanceGetIntegration(t *testing.T) {
	service, cleanup := linode.SetupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_instance_get",
			Arguments: map[string]interface{}{
				"instance_id": float64(123456),
			},
		},
	}

	result, err := service.handleInstanceGet(ctx, request)
	require.NoError(t, err, "instance get should not return error")
	require.NotNil(t, result, "result should not be nil")
	require.NotEmpty(t, result.Content, "result should have content")

	// Extract and validate JSON response
	textContent, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok, "content should be text")

	var instance map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &instance)
	require.NoError(t, err, "response should be valid JSON")

	// Validate complete instance structure
	ValidateJSONStructure(t, instance, map[string]string{
		"id":               "number",
		"label":            "string",
		"region":           "string",
		"image":            "string",
		"type":             "string",
		"status":           "string",
		"ipv4":             "array",
		"ipv6":             "string",
		"created":          "string",
		"updated":          "string",
		"hypervisor":       "string",
		"watchdog_enabled": "bool",
		"tags":             "array",
		"specs":            "object",
		"alerts":           "object",
		"backups":          "object",
		"group":            "string",
	})

	// Validate nested objects
	specs := instance["specs"].(map[string]interface{})
	ValidateJSONStructure(t, specs, map[string]string{
		"disk":     "number",
		"memory":   "number",
		"vcpus":    "number",
		"gpus":     "number",
		"transfer": "number",
	})

	alerts := instance["alerts"].(map[string]interface{})
	ValidateJSONStructure(t, alerts, map[string]string{
		"cpu":            "number",
		"network_in":     "number",
		"network_out":    "number",
		"transfer_quota": "number",
		"io":             "number",
	})

	// Validate specific values
	require.Equal(t, float64(123456), instance["id"], "instance ID should match")
	require.Equal(t, "test-instance-1", instance["label"], "instance label should match")
	require.Equal(t, "us-east", instance["region"], "instance region should match")
	require.Equal(t, "running", instance["status"], "instance status should match")
}

// TestInstanceCreateIntegration tests instance creation through MCP interface.
// This integration test verifies that CloudMCP can successfully process instance
// creation requests and return appropriate responses using the mock API.
//
// **Integration Test Workflow**:
// 1. **Service Setup**: Initialize CloudMCP service with mock API
// 2. **Create Request**: Execute instance create tool with valid parameters
// 3. **Response Validation**: Verify new instance data structure and values
// 4. **Status Verification**: Ensure creation response has correct status
//
// **Test Scenario**: Create new instance with standard configuration
//
// **Expected Behavior**:
// • Returns new instance data with generated ID
// • Instance status shows "provisioning" for new instances
// • All required fields are present in creation response
// • Response structure matches Linode API instance format
//
// **Purpose**: Validates that CloudMCP can successfully handle instance creation
// workflows through the MCP interface with proper parameter handling.
func TestInstanceCreateIntegration(t *testing.T) {
	service, cleanup := linode.SetupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_instance_create",
			Arguments: map[string]interface{}{
				"region":     "us-east",
				"type":       "g6-nanode-1",
				"label":      "new-test-instance",
				"image":      "linode/ubuntu22.04",
				"root_pass":  "SecurePassword123!",
				"booted":     true,
				"private_ip": false,
			},
		},
	}

	result, err := service.handleInstanceCreate(ctx, request)
	require.NoError(t, err, "instance create should not return error")
	require.NotNil(t, result, "result should not be nil")
	require.NotEmpty(t, result.Content, "result should have content")

	// Extract and validate JSON response
	textContent, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok, "content should be text")

	var instance map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &instance)
	require.NoError(t, err, "response should be valid JSON")

	// Validate new instance structure
	ValidateJSONStructure(t, instance, map[string]string{
		"id":     "number",
		"label":  "string",
		"region": "string",
		"image":  "string",
		"type":   "string",
		"status": "string",
		"ipv4":   "array",
		"ipv6":   "string",
	})

	// Validate creation-specific values
	require.Equal(t, float64(123457), instance["id"], "new instance should have different ID")
	require.Equal(t, "new-test-instance", instance["label"], "instance label should match request")
	require.Equal(t, "provisioning", instance["status"], "new instance should be provisioning")
	require.Equal(t, "us-east", instance["region"], "instance region should match request")
	require.Equal(t, "g6-nanode-1", instance["type"], "instance type should match request")
}

// TestInstanceNotFoundErrorIntegration tests error handling for non-existent instances.
// This integration test verifies that CloudMCP properly handles and formats error
// responses when requesting an instance that doesn't exist in the mock API.
//
// **Integration Test Workflow**:
// 1. **Service Setup**: Initialize CloudMCP with mock API endpoints
// 2. **Error Request**: Request instance with ID that triggers 404 error
// 3. **Error Validation**: Verify proper error handling and response format
// 4. **Message Verification**: Ensure error messages are meaningful
//
// **Test Scenario**: Request instance ID 999999 (configured to return 404)
//
// **Expected Behavior**:
// • Returns MCP error result instead of throwing exception
// • Error message indicates "not found" condition
// • Response structure follows MCP error format
// • Error details match Linode API error structure
//
// **Purpose**: Ensures CloudMCP gracefully handles API errors and provides
// meaningful error information through the MCP interface.
func TestInstanceNotFoundErrorIntegration(t *testing.T) {
	service, cleanup := linode.SetupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_instance_get",
			Arguments: map[string]interface{}{
				"instance_id": float64(999999), // This ID is configured to return 404
			},
		},
	}

	result, err := service.handleInstanceGet(ctx, request)
	require.NoError(t, err, "should not return Go error for API errors")
	require.NotNil(t, result, "result should not be nil")

	// Should be an error result
	require.True(t, result.IsError, "result should be marked as error")
	require.NotEmpty(t, result.Content, "error result should have content")

	textContent, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok, "content should be text")
	require.Contains(t, textContent.Text, "not found", "error should mention not found")
}

// TestVolumesListIntegration tests volume listing through MCP interface.
// This integration test verifies that CloudMCP can successfully retrieve and
// format volume information using the mock Linode API.
//
// **Integration Test Workflow**:
// 1. **Service Setup**: Initialize CloudMCP service with mock API container
// 2. **Volume Request**: Execute volumes list tool through MCP interface
// 3. **Response Validation**: Verify volume data structure and content
// 4. **Contract Verification**: Ensure response matches Linode API format
//
// **Test Environment**: Mock API with predefined volume data
//
// **Expected Behavior**:
// • Returns paginated volume list in MCP format
// • Volume objects contain all required fields
// • Response structure matches Linode API contract
// • Attached volume information is properly formatted
//
// **Purpose**: Validates that CloudMCP can successfully handle volume operations
// through the MCP interface with proper data formatting.
func TestVolumesListIntegration(t *testing.T) {
	service, cleanup := linode.SetupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_volumes_list",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.handleVolumesList(ctx, request)
	require.NoError(t, err, "volumes list should not return error")
	require.NotNil(t, result, "result should not be nil")
	require.NotEmpty(t, result.Content, "result should have content")

	// Extract and validate JSON response
	textContent, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok, "content should be text")

	var response map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &response)
	require.NoError(t, err, "response should be valid JSON")

	// Validate response structure
	require.Contains(t, response, "data", "response should contain data field")
	data := response["data"].([]interface{})
	require.Len(t, data, 1, "should have one volume")

	volume := data[0].(map[string]interface{})
	ValidateJSONStructure(t, volume, map[string]string{
		"id":              "number",
		"label":           "string",
		"region":          "string",
		"size":            "number",
		"status":          "string",
		"created":         "string",
		"updated":         "string",
		"filesystem_path": "string",
		"tags":            "array",
		"linode_id":       "number",
		"linode_label":    "string",
	})

	// Validate specific values
	require.Equal(t, float64(456789), volume["id"], "volume ID should match stub")
	require.Equal(t, "test-volume-1", volume["label"], "volume label should match stub")
	require.Equal(t, "active", volume["status"], "volume status should match stub")
	require.Equal(t, float64(20), volume["size"], "volume size should match stub")
}
