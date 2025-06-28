//go:build integration

package linode

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// ContractTest represents a test case for CloudMCP handler contract validation.
// Contract tests ensure that CloudMCP's handlers provide consistent text-based
// responses with expected formatting and content structure.
type ContractTest struct {
	Name             string                   // Test case name
	ToolName         string                   // MCP tool name to test
	Arguments        map[string]interface{}   // Tool arguments
	ExpectedContains []string                 // Expected text content patterns
	Validator        func(*testing.T, string) // Custom validation function for text content
}

// TestCloudMCPHandlerContracts runs a comprehensive suite of CloudMCP handler contract tests.
// This test verifies that CloudMCP handlers correctly format text responses
// and maintain consistent output patterns across different operations.
//
// **Contract Validation Areas**:
// • Text response formatting and structure
// • Required information presence and accuracy
// • Consistent output patterns across handlers
// • Error message formatting and clarity
//
// **Test Coverage**:
// • Account information text formatting
// • Instance management text responses
// • Error condition text handling
// • Data consistency validation
//
// **Purpose**: Ensures that CloudMCP maintains consistent text-based output
// contracts for reliable integration with LLM interfaces.
func TestCloudMCPHandlerContracts(t *testing.T) {
	service, cleanup := SetupHTTPTestIntegration(t)
	defer cleanup()

	contractTests := []ContractTest{
		{
			Name:      "Account Information Text Contract",
			ToolName:  "linode_account_get",
			Arguments: map[string]interface{}{},
			ExpectedContains: []string{
				"Account:", "httptest-integration", "HTTP Test Integration Account",
				"Username:", "testuser",
				"Email:", "test@example.com",
				"UID:", "12345",
				"Restricted:", "false",
			},
			Validator: func(t *testing.T, text string) {
				require.Contains(t, text, "Account: httptest-integration (HTTP Test Integration Account)")
				require.Contains(t, text, "Username: testuser")
				require.Contains(t, text, "Email: test@example.com")
				require.Contains(t, text, "UID: 12345")
				require.Contains(t, text, "Restricted: false")
			},
		},
		{
			Name:      "Instance List Text Contract",
			ToolName:  "linode_instances_list",
			Arguments: map[string]interface{}{},
			ExpectedContains: []string{
				"Found 1 Linode instance(s):",
				"ID: 123456", "test-instance-1",
				"Status: running", "Region: us-east", "Type: g6-nanode-1",
				"IPv4: [192.168.1.1]",
			},
			Validator: func(t *testing.T, text string) {
				require.Contains(t, text, "Found 1 Linode instance(s):")
				require.Contains(t, text, "ID: 123456 | test-instance-1")
				require.Contains(t, text, "Status: running | Region: us-east | Type: g6-nanode-1")
				require.Contains(t, text, "IPv4: [192.168.1.1]")
			},
		},
		{
			Name:     "Instance Details Text Contract",
			ToolName: "linode_instance_get",
			Arguments: map[string]interface{}{
				"instance_id": float64(123456),
			},
			ExpectedContains: []string{
				"Instance Details:",
				"ID: 123456", "Label: test-instance-1",
				"Status: running", "Region: us-east", "Type: g6-nanode-1",
				"Image: linode/ubuntu22.04",
				"Specifications:",
				"CPUs: 1", "Memory: 1024 MB", "Transfer: 0 GB",
				"Network:",
				"IPv4: 192.168.1.1", "IPv6: 2600:3c01::f03c:91ff:fe24:3a2f/128",
				"Created:", "Updated:",
				"Backups:", "Watchdog:",
			},
			Validator: func(t *testing.T, text string) {
				require.Contains(t, text, "Instance Details:")
				require.Contains(t, text, "ID: 123456")
				require.Contains(t, text, "Label: test-instance-1")
				require.Contains(t, text, "Status: running")
				require.Contains(t, text, "Region: us-east")
				require.Contains(t, text, "Type: g6-nanode-1")
				require.Contains(t, text, "Image: linode/ubuntu22.04")
				require.Contains(t, text, "Specifications:")
				require.Contains(t, text, "CPUs: 1")
				require.Contains(t, text, "Memory: 1024 MB")
				require.Contains(t, text, "Network:")
				require.Contains(t, text, "IPv4: 192.168.1.1")
				require.Contains(t, text, "IPv6: 2600:3c01::f03c:91ff:fe24:3a2f/128")
			},
		},
	}

	for _, test := range contractTests {
		t.Run(test.Name, func(t *testing.T) {
			runContractTest(t, service, test)
		})
	}
}

// runContractTest executes a single contract test against the CloudMCP service.
// This function handles the handler call, text response validation, and custom
// validation logic for each contract test case.
func runContractTest(t *testing.T, service *Service, test ContractTest) {
	ctx := context.Background()

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      test.ToolName,
			Arguments: test.Arguments,
		},
	}

	var result *mcp.CallToolResult
	var err error

	// Call the appropriate handler method based on tool name
	switch test.ToolName {
	case "linode_account_get":
		result, err = service.handleAccountGet(ctx, request)
	case "linode_instances_list":
		result, err = service.handleInstancesList(ctx, request)
	case "linode_instance_get":
		result, err = service.handleInstanceGet(ctx, request)
	case "linode_volumes_list":
		result, err = service.handleVolumesList(ctx, request)
	default:
		t.Fatalf("unsupported tool name: %s", test.ToolName)
	}

	require.NoError(t, err, "tool call should not return error for %s", test.Name)
	require.NotNil(t, result, "result should not be nil for %s", test.Name)
	require.NotEmpty(t, result.Content, "result should have content for %s", test.Name)

	// Extract text response
	textContent, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok, "content should be text for %s", test.Name)

	responseText := textContent.Text
	require.NotEmpty(t, responseText, "response text should not be empty for %s", test.Name)

	// Validate expected text patterns
	for _, expectedPattern := range test.ExpectedContains {
		require.Contains(t, responseText, expectedPattern,
			"response should contain '%s' for %s", expectedPattern, test.Name)
	}

	// Run custom validator if provided
	if test.Validator != nil {
		test.Validator(t, responseText)
	}
}

// TestMCPProtocolCompliance validates CloudMCP's MCP protocol implementation.
// This test ensures that all handler responses follow the proper MCP format
// and contain required metadata and structure elements.
//
// **MCP Compliance Areas**:
// • Tool result format and structure
// • Text content type handling
// • Error response formatting patterns
// • Content structure validation
//
// **Validation Points**:
// • Result objects contain proper Content arrays
// • Text content is properly formatted
// • Successful results have IsError=false
// • Content arrays contain exactly one TextContent item
//
// **Purpose**: Ensures CloudMCP maintains full compatibility with the
// Model Context Protocol specification for text-based tool interactions.
func TestMCPProtocolCompliance(t *testing.T) {
	service, cleanup := SetupHTTPTestIntegration(t)
	defer cleanup()

	ctx := context.Background()

	// Test successful response format
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_account_get",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.handleAccountGet(ctx, request)
	require.NoError(t, err, "tool call should not return error")
	require.NotNil(t, result, "result should not be nil")

	// Validate MCP result structure
	require.False(t, result.IsError, "successful call should not be marked as error")
	require.NotEmpty(t, result.Content, "result should have content")
	require.Len(t, result.Content, 1, "result should have exactly one content item")

	// Validate content type
	content := result.Content[0]
	textContent, ok := content.(mcp.TextContent)
	require.True(t, ok, "content should be TextContent type")
	require.NotEmpty(t, textContent.Text, "text content should not be empty")

	// Validate that text content is well-formatted (contains expected patterns)
	require.Contains(t, textContent.Text, "Account:", "text should contain account information")
	require.Contains(t, textContent.Text, "Username:", "text should contain username")
	require.Contains(t, textContent.Text, "Email:", "text should contain email")

	// Test error response handling (current implementation returns Go errors)
	errorRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_instance_get",
			Arguments: map[string]interface{}{
				"instance_id": float64(999999), // Non-existent instance
			},
		},
	}

	errorResult, err := service.handleInstanceGet(ctx, errorRequest)

	// Current implementation returns Go errors for API failures
	require.Error(t, err, "error responses should return Go errors")
	require.Contains(t, err.Error(), "failed to get instance 999999", "error should contain meaningful message")
	require.Contains(t, err.Error(), "linode/instance_get", "error should include tool context")

	// Error result might be nil when Go error is returned
	if errorResult != nil {
		require.True(t, errorResult.IsError, "error result should be marked as error if not nil")
	}
}
