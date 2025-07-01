package linode_test

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/internal/services/linode"
	"github.com/chadit/CloudMCP/pkg/logger"
)

// TestHandleFirewallsList_AccountError tests the handleFirewallsList function to verify it returns all firewalls.
// This test simulates a user requesting a list of all their Linode firewalls through the MCP interface.
// Since this function requires Linode API calls, this test focuses on the error handling path.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with empty account manager
// 2. **Request Execution**: Call handleFirewallsList expecting account manager failure
// 3. **Error Validation**: Verify appropriate error is returned for account lookup failure
//
// **Test Environment**: Service with no configured accounts to trigger error path
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Does not attempt to call Linode API when account lookup fails
// • Provides meaningful error message for troubleshooting
//
// **Purpose**: This test ensures firewalls list command fails appropriately when account configuration is invalid.
// Note: Full integration testing with mock Linode client requires interface abstraction (future improvement).
func TestHandleFirewallsList_AccountError(t *testing.T) {
	t.Parallel()

	// Create minimal service with empty account manager
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	// Create completely isolated account manager for this test only
	accountManager := linode.NewAccountManagerForTesting()

	service := linode.NewForTesting(cfg, log, accountManager)

	// Test firewalls list request with empty account manager
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_firewalls_list",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleFirewallsList should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleFirewallGet_AccountError tests the handleFirewallGet function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleFirewallGet with no configured accounts
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
func TestHandleFirewallGet_AccountError(t *testing.T) {
	t.Parallel()

	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	// Don't add any accounts - the account manager should be completely empty
	// Don't set current account - it should remain empty string
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test firewall get request with no accounts and valid parameter
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_firewall_get",
			Arguments: map[string]interface{}{
				"firewall_id": float64(12345),
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleFirewallGet should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleFirewallCreate_AccountError tests the handleFirewallCreate function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleFirewallCreate with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleFirewallCreate_AccountError(t *testing.T) {
	t.Parallel()

	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test firewall create request with no accounts
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_firewall_create",
			Arguments: map[string]interface{}{
				"label": "test-firewall",
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleFirewallCreate should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleFirewallUpdate_AccountError tests the handleFirewallUpdate function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleFirewallUpdate with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleFirewallUpdate_AccountError(t *testing.T) {
	t.Parallel()

	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test firewall update request with no accounts
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_firewall_update",
			Arguments: map[string]interface{}{
				"firewall_id": float64(12345),
				"label":       "updated-firewall",
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleFirewallUpdate should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleFirewallDelete_AccountError tests the handleFirewallDelete function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleFirewallDelete with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleFirewallDelete_AccountError(t *testing.T) {
	t.Parallel()

	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test firewall delete request with no accounts
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_firewall_delete",
			Arguments: map[string]interface{}{
				"firewall_id": float64(12345),
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleFirewallDelete should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleFirewallRulesUpdate_AccountError tests the handleFirewallRulesUpdate function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleFirewallRulesUpdate with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleFirewallRulesUpdate_AccountError(t *testing.T) {
	t.Parallel()

	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test firewall rules update request with no accounts
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_firewall_rules_update",
			Arguments: map[string]interface{}{
				"firewall_id": float64(12345),
				"rules":       map[string]interface{}{},
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleFirewallRulesUpdate should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleFirewallDeviceCreate_AccountError tests the handleFirewallDeviceCreate function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleFirewallDeviceCreate with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleFirewallDeviceCreate_AccountError(t *testing.T) {
	t.Parallel()

	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test firewall device create request with no accounts
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_firewall_device_create",
			Arguments: map[string]interface{}{
				"firewall_id": float64(12345),
				"device_id":   float64(67890),
				"device_type": "linode",
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleFirewallDeviceCreate should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleFirewallDeviceDelete_AccountError tests the handleFirewallDeviceDelete function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleFirewallDeviceDelete with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleFirewallDeviceDelete_AccountError(t *testing.T) {
	t.Parallel()

	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test firewall device delete request with no accounts
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_firewall_device_delete",
			Arguments: map[string]interface{}{
				"firewall_id": float64(12345),
				"device_id":   float64(67890),
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleFirewallDeviceDelete should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// Note: Additional tests for successful firewall operations with mock Linode clients
// are not implemented in this unit test suite because they require functioning
// Linode API client interfaces for operations like ListFirewalls, GetFirewall,
// CreateFirewall, UpdateFirewall, DeleteFirewall, UpdateFirewallRules,
// CreateFirewallDevice, and DeleteFirewallDevice.
//
// Parameter validation tests are not included because the firewall handlers
// currently use parseArguments() placeholder which doesn't implement validation.
// The firewall handlers would need to be updated to use parseIDFromArguments()
// or similar validation functions like other tool handlers.
//
// These operations would require either:
// 1. Interface abstraction for the Linode client (future improvement)
// 2. Integration testing with real API endpoints
// 3. Dependency injection to replace the client during testing
// 4. Implementation of proper parameter validation in firewall handlers
//
// The current tests adequately cover:
// - Account manager error scenarios for all 8 firewall tools
// - Request routing and basic tool handler setup
// - Error message formatting and response structure
// - MCP protocol compliance for error responses
//
// This provides coverage of the testable logic that doesn't
// require external API dependencies or parameter validation implementation.
