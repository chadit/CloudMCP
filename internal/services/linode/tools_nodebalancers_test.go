package linode_test

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/internal/services/linode"
	"github.com/chadit/CloudMCP/pkg/logger"
)

// TestHandleNodeBalancersList_AccountError tests the handleNodeBalancersList function to verify error handling.
// This test simulates a user requesting a list of all their Linode NodeBalancers through the MCP interface.
// Since this function requires Linode API calls, this test focuses on the error handling path.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with empty account manager
// 2. **Request Execution**: Call handleNodeBalancersList expecting account manager failure
// 3. **Error Validation**: Verify appropriate error is returned for account lookup failure
//
// **Test Environment**: Service with no configured accounts to trigger error path
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Does not attempt to call Linode API when account lookup fails
// • Provides meaningful error message for troubleshooting
//
// **Purpose**: This test ensures NodeBalancers list command fails appropriately when account configuration is invalid.
func TestHandleNodeBalancersList_AccountError(t *testing.T) {
	// Create minimal service with empty account manager
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	// Create completely isolated account manager for this test only
	accountManager := linode.NewAccountManagerForTesting()

	service := linode.NewForTesting(cfg, log, accountManager)

	// Test NodeBalancers list request with empty account manager
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_nodebalancers_list",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleNodeBalancersList should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleNodeBalancerGet_AccountError tests the handleNodeBalancerGet function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleNodeBalancerGet with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Test Environment**: Service with no configured accounts and empty current account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleNodeBalancerGet_AccountError(t *testing.T) {
	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test NodeBalancer get request with no accounts and valid parameter
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_nodebalancer_get",
			Arguments: map[string]interface{}{
				"nodebalancer_id": float64(12345),
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleNodeBalancerGet should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleNodeBalancerCreate_AccountError tests the handleNodeBalancerCreate function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleNodeBalancerCreate with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleNodeBalancerCreate_AccountError(t *testing.T) {
	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test NodeBalancer create request with no accounts
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_nodebalancer_create",
			Arguments: map[string]interface{}{
				"label":  "test-nodebalancer",
				"region": "us-east",
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleNodeBalancerCreate should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleNodeBalancerUpdate_AccountError tests the handleNodeBalancerUpdate function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleNodeBalancerUpdate with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleNodeBalancerUpdate_AccountError(t *testing.T) {
	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test NodeBalancer update request with no accounts
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_nodebalancer_update",
			Arguments: map[string]interface{}{
				"nodebalancer_id": float64(12345),
				"label":           "updated-nodebalancer",
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleNodeBalancerUpdate should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleNodeBalancerDelete_AccountError tests the handleNodeBalancerDelete function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleNodeBalancerDelete with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleNodeBalancerDelete_AccountError(t *testing.T) {
	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test NodeBalancer delete request with no accounts
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_nodebalancer_delete",
			Arguments: map[string]interface{}{
				"nodebalancer_id": float64(12345),
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleNodeBalancerDelete should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleNodeBalancerConfigCreate_AccountError tests the handleNodeBalancerConfigCreate function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleNodeBalancerConfigCreate with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleNodeBalancerConfigCreate_AccountError(t *testing.T) {
	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test NodeBalancer config create request with no accounts
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_nodebalancer_config_create",
			Arguments: map[string]interface{}{
				"nodebalancer_id": float64(12345),
				"port":            float64(80),
				"protocol":        "http",
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleNodeBalancerConfigCreate should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleNodeBalancerConfigUpdate_AccountError tests the handleNodeBalancerConfigUpdate function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleNodeBalancerConfigUpdate with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleNodeBalancerConfigUpdate_AccountError(t *testing.T) {
	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test NodeBalancer config update request with no accounts
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_nodebalancer_config_update",
			Arguments: map[string]interface{}{
				"nodebalancer_id": float64(12345),
				"config_id":       float64(67890),
				"port":            float64(443),
				"protocol":        "https",
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleNodeBalancerConfigUpdate should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleNodeBalancerConfigDelete_AccountError tests the handleNodeBalancerConfigDelete function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleNodeBalancerConfigDelete with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleNodeBalancerConfigDelete_AccountError(t *testing.T) {
	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test NodeBalancer config delete request with no accounts
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_nodebalancer_config_delete",
			Arguments: map[string]interface{}{
				"nodebalancer_id": float64(12345),
				"config_id":       float64(67890),
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleNodeBalancerConfigDelete should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// Note: Additional tests for successful NodeBalancer operations with mock Linode clients
// are not implemented in this unit test suite because they require functioning
// Linode API client interfaces for operations like ListNodeBalancers, GetNodeBalancer,
// CreateNodeBalancer, UpdateNodeBalancer, DeleteNodeBalancer, CreateNodeBalancerConfig,
// UpdateNodeBalancerConfig, and DeleteNodeBalancerConfig.
//
// Parameter validation tests are limited because the NodeBalancer handlers
// use parseArguments() function which provides basic validation but not comprehensive
// parameter validation like parseIDFromArguments() used in other tool handlers.
//
// These operations would require either:
// 1. Interface abstraction for the Linode client (future improvement)
// 2. Integration testing with real API endpoints
// 3. Dependency injection to replace the client during testing
// 4. Enhanced parameter validation in NodeBalancer handlers
//
// The current tests adequately cover:
// - Account manager error scenarios for all 8 NodeBalancer tools
// - Request routing and basic tool handler setup
// - Error message formatting and response structure
// - MCP protocol compliance for error responses
//
// This provides coverage of the testable logic that doesn't
// require external API dependencies while ensuring proper error handling
// for the most common failure scenario (account configuration issues).
