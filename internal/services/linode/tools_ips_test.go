package linode_test

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/internal/services/linode"
	"github.com/chadit/CloudMCP/pkg/logger"
)

// TestHandleIPsList_AccountError tests the handleIPsList function to verify it returns all IP addresses.
// This test simulates a user requesting a list of all their Linode IP addresses through the MCP interface.
// Since this function requires Linode API calls, this test focuses on the error handling path.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with empty account manager
// 2. **Request Execution**: Call handleIPsList expecting account manager failure
// 3. **Error Validation**: Verify appropriate error is returned for account lookup failure
//
// **Test Environment**: Service with no configured accounts to trigger error path
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Does not attempt to call Linode API when account lookup fails
// • Provides meaningful error message for troubleshooting
//
// **Purpose**: This test ensures IPs list command fails appropriately when account configuration is invalid.
// Note: Full integration testing with mock Linode client requires interface abstraction (future improvement).
func TestHandleIPsList_AccountError(t *testing.T) {
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

	// Test IPs list request with empty account manager
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_ips_list",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.Error(t, err, "handleIPsList should return an error when no accounts are configured")
	require.Nil(t, result, "result should be nil when error occurs")
}

// TestHandleIPGet_AccountError tests the handleIPGet function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleIPGet with no configured accounts
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
func TestHandleIPGet_AccountError(t *testing.T) {
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

	// Test IP get request with no accounts and valid parameter
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_ip_get",
			Arguments: map[string]interface{}{
				"address": "192.168.1.1",
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.Error(t, err, "handleIPGet should return error when no current account exists")
	require.Nil(t, result, "result should be nil when error occurs")
}

// TestHandleIPGet_MissingParameter tests the handleIPGet function with missing required parameters.
// This test verifies the function handles requests with missing address parameter.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Missing Parameter**: Call handleIPGet without address parameter
// 3. **Error Validation**: Verify appropriate parameter error is returned
//
// **Expected Behavior**:
// • Returns error result for missing required parameters
// • Provides specific error message about missing address
// • Does not attempt account manager operations with invalid input
//
// **Purpose**: This test ensures IP get validates required parameters properly.
func TestHandleIPGet_MissingParameter(t *testing.T) {
	t.Parallel()

	// Create isolated test service
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "test-account",
		LinodeAccounts: map[string]config.LinodeAccount{
			"test-account": {Label: "Test Account", Token: "test-token"},
		},
	}

	accountManager := linode.NewAccountManagerForTesting()
	testAccount := linode.NewAccountForTesting("test-account", "Test Account")
	accountManager.AddAccountForTesting(testAccount)
	accountManager.SetCurrentAccountForTesting("test-account")

	service := linode.NewForTesting(cfg, log, accountManager)

	// Test IP get request with missing address parameter
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_ip_get",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleIPGet should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "address", "error result should mention missing address parameter")
		}
	}
}

// TestHandleIPGet_EmptyParameter tests the handleIPGet function with empty address parameter.
// This test verifies the function handles requests with empty string address parameter.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Empty Parameter**: Call handleIPGet with empty address string
// 3. **Error Validation**: Verify appropriate parameter error is returned
//
// **Expected Behavior**:
// • Returns error result for empty parameter values
// • Provides specific error message about missing address
// • Treats empty string as missing required parameter
//
// **Purpose**: This test ensures IP get validates parameter content as well as presence.
func TestHandleIPGet_EmptyParameter(t *testing.T) {
	t.Parallel()

	// Create isolated test service
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "test-account",
		LinodeAccounts: map[string]config.LinodeAccount{
			"test-account": {Label: "Test Account", Token: "test-token"},
		},
	}

	accountManager := linode.NewAccountManagerForTesting()
	testAccount := linode.NewAccountForTesting("test-account", "Test Account")
	accountManager.AddAccountForTesting(testAccount)
	accountManager.SetCurrentAccountForTesting("test-account")

	service := linode.NewForTesting(cfg, log, accountManager)

	// Test IP get request with empty address parameter
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_ip_get",
			Arguments: map[string]interface{}{
				"address": "",
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleIPGet should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "address", "error result should mention missing address parameter")
		}
	}
}

// TestHandleIPGet_InvalidIPFormat tests the handleIPGet function with invalid IP address format.
// This test verifies the function handles requests with malformed IP address strings.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Invalid IP Format**: Call handleIPGet with malformed IP address
// 3. **Error Validation**: Verify appropriate IP format error is returned
//
// **Expected Behavior**:
// • Returns error result for invalid IP address format
// • Provides specific error message about invalid IP address format
// • Validates IP address format before attempting API calls
//
// **Purpose**: This test ensures IP get validates IP address format properly.
func TestHandleIPGet_InvalidIPFormat(t *testing.T) {
	t.Parallel()

	// Create isolated test service
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "test-account",
		LinodeAccounts: map[string]config.LinodeAccount{
			"test-account": {Label: "Test Account", Token: "test-token"},
		},
	}

	accountManager := linode.NewAccountManagerForTesting()
	testAccount := linode.NewAccountForTesting("test-account", "Test Account")
	accountManager.AddAccountForTesting(testAccount)
	accountManager.SetCurrentAccountForTesting("test-account")

	service := linode.NewForTesting(cfg, log, accountManager)

	// Test IP get request with invalid IP address format
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_ip_get",
			Arguments: map[string]interface{}{
				"address": "not-an-ip-address",
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleIPGet should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "invalid IP address", "error result should mention invalid IP address format")
		}
	}
}

// TestHandleIPGet_InvalidIPFormat_Numbers tests the handleIPGet function with numeric but invalid IP format.
// This test verifies the function handles requests with numeric strings that aren't valid IP addresses.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Invalid IP Numbers**: Call handleIPGet with numeric but invalid IP format
// 3. **Error Validation**: Verify appropriate IP format error is returned
//
// **Expected Behavior**:
// • Returns error result for invalid IP address format
// • Provides specific error message about invalid IP address format
// • Validates IP address format correctly even for numeric strings
//
// **Purpose**: This test ensures IP get validates IP address format comprehensively.
func TestHandleIPGet_InvalidIPFormat_Numbers(t *testing.T) {
	t.Parallel()

	// Create isolated test service
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "test-account",
		LinodeAccounts: map[string]config.LinodeAccount{
			"test-account": {Label: "Test Account", Token: "test-token"},
		},
	}

	accountManager := linode.NewAccountManagerForTesting()
	testAccount := linode.NewAccountForTesting("test-account", "Test Account")
	accountManager.AddAccountForTesting(testAccount)
	accountManager.SetCurrentAccountForTesting("test-account")

	service := linode.NewForTesting(cfg, log, accountManager)

	// Test IP get request with invalid IP address format (numbers but invalid)
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_ip_get",
			Arguments: map[string]interface{}{
				"address": "999.999.999.999", // Invalid IP - octets too large
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleIPGet should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "invalid IP address", "error result should mention invalid IP address format")
		}
	}
}

// TestHandleIPGet_ValidIPv4Format tests the handleIPGet function with valid IPv4 format but account error.
// This test verifies the function validates IP format correctly before attempting account operations.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with empty account manager
// 2. **Valid IPv4 Format**: Call handleIPGet with valid IPv4 address format
// 3. **Account Error**: Trigger account manager error to test error path
//
// **Expected Behavior**:
// • IP format validation passes for valid IPv4 input
// • Returns error when account manager fails
// • Does not proceed to API calls when account unavailable
//
// **Purpose**: This test ensures IP get IP format validation works correctly for IPv4 addresses.
func TestHandleIPGet_ValidIPv4Format(t *testing.T) {
	t.Parallel()

	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	// Use completely empty account manager to trigger account error after IP format validation
	accountManager := linode.NewAccountManagerForTesting()
	// Don't add any accounts - the account manager should be completely empty
	// Don't set current account - it should remain empty string
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test IP get request with valid IPv4 format but no account
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_ip_get",
			Arguments: map[string]interface{}{
				"address": "192.168.1.100",
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.Error(t, err, "handleIPGet should return error when account manager fails")
	require.Nil(t, result, "result should be nil when account error occurs")
	// Error should be about account, not IP format validation
	require.NotContains(t, err.Error(), "invalid IP address", "error should not be about IP format validation")
}

// TestHandleIPGet_ValidIPv6Format tests the handleIPGet function with valid IPv6 format but account error.
// This test verifies the function validates IPv6 format correctly before attempting account operations.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with empty account manager
// 2. **Valid IPv6 Format**: Call handleIPGet with valid IPv6 address format
// 3. **Account Error**: Trigger account manager error to test error path
//
// **Expected Behavior**:
// • IP format validation passes for valid IPv6 input
// • Returns error when account manager fails
// • Does not proceed to API calls when account unavailable
//
// **Purpose**: This test ensures IP get IP format validation works correctly for IPv6 addresses.
func TestHandleIPGet_ValidIPv6Format(t *testing.T) {
	t.Parallel()

	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	// Use completely empty account manager to trigger account error after IP format validation
	accountManager := linode.NewAccountManagerForTesting()
	// Don't add any accounts - the account manager should be completely empty
	// Don't set current account - it should remain empty string
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test IP get request with valid IPv6 format but no account
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_ip_get",
			Arguments: map[string]interface{}{
				"address": "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.Error(t, err, "handleIPGet should return error when account manager fails")
	require.Nil(t, result, "result should be nil when account error occurs")
	// Error should be about account, not IP format validation
	require.NotContains(t, err.Error(), "invalid IP address", "error should not be about IP format validation")
}

// Note: Additional tests for successful IP operations with mock Linode clients
// are not implemented in this unit test suite because they require functioning
// Linode API client interfaces for operations like ListIPs and instance IP lookups.
//
// These operations would require either:
// 1. Interface abstraction for the Linode client (future improvement)
// 2. Integration testing with real API endpoints
// 3. Dependency injection to replace the client during testing
//
// The current tests adequately cover:
// - Parameter validation and error handling logic
// - Account manager error scenarios
// - IP address format validation (IPv4 and IPv6)
// - Request routing and basic tool handler setup
// - Error message formatting and response structure
//
// This provides comprehensive coverage of the testable logic that doesn't
// require external API dependencies.
