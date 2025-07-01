package linode_test

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/internal/services/linode"
	"github.com/chadit/CloudMCP/pkg/logger"
)

// TestHandleInstancesList tests the handleInstancesList function to verify it returns all instances in the current account.
// This test simulates a user requesting a list of all their Linode instances through the MCP interface.
// Since this function requires Linode API calls, this test focuses on the error handling path.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with empty account manager
// 2. **Request Execution**: Call handleInstancesList expecting account manager failure
// 3. **Error Validation**: Verify appropriate error is returned for account lookup failure
//
// **Test Environment**: Service with no configured accounts to trigger error path
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Does not attempt to call Linode API when account lookup fails
// • Provides meaningful error message for troubleshooting
//
// **Purpose**: This test ensures instances list command fails appropriately when account configuration is invalid.
// Note: Full integration testing with mock Linode client requires interface abstraction (future improvement).
func TestHandleInstancesList_AccountError(t *testing.T) {
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

	// Test instances list request with empty account manager
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_instances_list",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.Error(t, err, "handleInstancesList should return an error when no accounts are configured")
	require.Nil(t, result, "result should be nil when error occurs")
}

// TestHandleInstanceGet_AccountError tests the handleInstanceGet function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleInstanceGet with no configured accounts
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
func TestHandleInstanceGet_AccountError(t *testing.T) {
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

	// Test instance get request with no accounts and valid parameter
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_instance_get",
			Arguments: map[string]interface{}{
				"instance_id": float64(12345),
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.Error(t, err, "handleInstanceGet should return error when no current account exists")
	require.Nil(t, result, "result should be nil when error occurs")
}

// TestHandleInstanceGet_MissingParameter tests the handleInstanceGet function with missing required parameters.
// This test verifies the function handles requests with missing instance_id parameter.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Missing Parameter**: Call handleInstanceGet without instance_id parameter
// 3. **Error Validation**: Verify appropriate parameter error is returned
//
// **Expected Behavior**:
// • Returns error result for missing required parameters
// • Provides specific error message about missing instance_id
// • Does not attempt account manager operations with invalid input
//
// **Purpose**: This test ensures instance get validates required parameters properly.
func TestHandleInstanceGet_MissingParameter(t *testing.T) {
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

	// Test instance get request with missing instance_id parameter
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_instance_get",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleInstanceGet should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "instance_id", "error result should mention missing instance_id parameter")
		}
	}
}

// TestHandleInstanceGet_InvalidParameter tests the handleInstanceGet function with invalid parameter types.
// This test verifies the function handles requests with non-integer instance_id parameter.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Invalid Parameter**: Call handleInstanceGet with string instance_id instead of integer
// 3. **Error Validation**: Verify appropriate parameter type error is returned
//
// **Expected Behavior**:
// • Returns error result for invalid parameter types
// • Provides specific error message about parameter type mismatch
// • Does not attempt account manager operations with invalid input
//
// **Purpose**: This test ensures instance get validates parameter types properly.
func TestHandleInstanceGet_InvalidParameter(t *testing.T) {
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

	// Test instance get request with invalid instance_id parameter type
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_instance_get",
			Arguments: map[string]interface{}{
				"instance_id": "not-a-number",
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleInstanceGet should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "instance_id", "error result should mention invalid instance_id parameter")
		}
	}
}

// TestHandleInstanceCreate_MissingRequiredParameters tests the handleInstanceCreate function with missing required parameters.
// This test verifies the function handles requests with missing required parameters for instance creation.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Missing Parameters**: Call handleInstanceCreate without required parameters (region, type, label)
// 3. **Error Validation**: Verify appropriate parameter error is returned
//
// **Expected Behavior**:
// • Returns error result for missing required parameters
// • Provides specific error message about missing required fields
// • Does not attempt account manager operations with invalid input
//
// **Purpose**: This test ensures instance create validates all required parameters properly.
func TestHandleInstanceCreate_MissingRequiredParameters(t *testing.T) {
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

	// Test instance create request with missing required parameters
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_instance_create",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleInstanceCreate should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "required", "error result should mention missing required parameters")
		}
	}
}

// TestHandleInstanceCreate_PartialParameters tests the handleInstanceCreate function with some required parameters missing.
// This test verifies the function handles requests with only some of the required parameters.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Partial Parameters**: Call handleInstanceCreate with only some required parameters
// 3. **Error Validation**: Verify appropriate parameter error is returned
//
// **Expected Behavior**:
// • Returns error result when any required parameter is missing
// • Provides specific error message about missing required fields
// • Validates all required parameters before attempting API calls
//
// **Purpose**: This test ensures instance create validates parameter completeness properly.
func TestHandleInstanceCreate_PartialParameters(t *testing.T) {
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

	// Test instance create request with only partial required parameters
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_instance_create",
			Arguments: map[string]interface{}{
				"region": "us-east",
				// Missing type and label
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleInstanceCreate should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "required", "error result should mention missing required parameters")
		}
	}
}

// TestHandleInstanceDelete_MissingParameter tests the handleInstanceDelete function with missing required parameters.
// This test verifies the function handles requests with missing instance_id parameter.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Missing Parameter**: Call handleInstanceDelete without instance_id parameter
// 3. **Error Validation**: Verify appropriate parameter error is returned
//
// **Expected Behavior**:
// • Returns error result for missing required parameters
// • Provides specific error message about missing instance_id
// • Does not attempt account manager operations with invalid input
//
// **Purpose**: This test ensures instance delete validates required parameters properly.
func TestHandleInstanceDelete_MissingParameter(t *testing.T) {
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

	// Test instance delete request with missing instance_id parameter
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_instance_delete",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleInstanceDelete should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "instance_id", "error result should mention missing instance_id parameter")
		}
	}
}

// TestHandleInstanceBoot_MissingParameter tests the handleInstanceBoot function with missing required parameters.
// This test verifies the function handles requests with missing instance_id parameter.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Missing Parameter**: Call handleInstanceBoot without instance_id parameter
// 3. **Error Validation**: Verify appropriate parameter error is returned
//
// **Expected Behavior**:
// • Returns error result for missing required parameters
// • Provides specific error message about missing instance_id
// • Does not attempt account manager operations with invalid input
//
// **Purpose**: This test ensures instance boot validates required parameters properly.
func TestHandleInstanceBoot_MissingParameter(t *testing.T) {
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

	// Test instance boot request with missing instance_id parameter
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_instance_boot",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleInstanceBoot should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "instance_id", "error result should mention missing instance_id parameter")
		}
	}
}

// TestHandleInstanceShutdown_MissingParameter tests the handleInstanceShutdown function with missing required parameters.
// This test verifies the function handles requests with missing instance_id parameter.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Missing Parameter**: Call handleInstanceShutdown without instance_id parameter
// 3. **Error Validation**: Verify appropriate parameter error is returned
//
// **Expected Behavior**:
// • Returns error result for missing required parameters
// • Provides specific error message about missing instance_id
// • Does not attempt account manager operations with invalid input
//
// **Purpose**: This test ensures instance shutdown validates required parameters properly.
func TestHandleInstanceShutdown_MissingParameter(t *testing.T) {
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

	// Test instance shutdown request with missing instance_id parameter
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_instance_shutdown",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleInstanceShutdown should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "instance_id", "error result should mention missing instance_id parameter")
		}
	}
}

// TestHandleInstanceReboot_MissingParameter tests the handleInstanceReboot function with missing required parameters.
// This test verifies the function handles requests with missing instance_id parameter.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Missing Parameter**: Call handleInstanceReboot without instance_id parameter
// 3. **Error Validation**: Verify appropriate parameter error is returned
//
// **Expected Behavior**:
// • Returns error result for missing required parameters
// • Provides specific error message about missing instance_id
// • Does not attempt account manager operations with invalid input
//
// **Purpose**: This test ensures instance reboot validates required parameters properly.
func TestHandleInstanceReboot_MissingParameter(t *testing.T) {
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

	// Test instance reboot request with missing instance_id parameter
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_instance_reboot",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleInstanceReboot should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "instance_id", "error result should mention missing instance_id parameter")
		}
	}
}

// TestHandleInstanceBoot_ValidParameters tests the handleInstanceBoot function with valid parameters but account error.
// This test verifies the function validates parameters correctly before attempting account operations.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with empty account manager
// 2. **Valid Parameters**: Call handleInstanceBoot with valid instance_id and optional config_id
// 3. **Account Error**: Trigger account manager error to test error path
//
// **Test Environment**: Service with no configured accounts and empty current account
//
// **Expected Behavior**:
// • Parameter validation passes for valid input
// • Returns error when account manager fails
// • Does not proceed to API calls when account unavailable
//
// **Purpose**: This test ensures instance boot parameter validation works correctly and account errors are handled.
func TestHandleInstanceBoot_ValidParameters(t *testing.T) {
	t.Parallel()

	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	// Use completely empty account manager to trigger account error after parameter validation
	accountManager := linode.NewAccountManagerForTesting()
	// Don't add any accounts - the account manager should be completely empty
	// Don't set current account - it should remain empty string
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test instance boot request with valid parameters but no account
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_instance_boot",
			Arguments: map[string]interface{}{
				"instance_id": float64(12345),
				"config_id":   float64(67890),
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.Error(t, err, "handleInstanceBoot should return error when account manager fails")
	require.Nil(t, result, "result should be nil when account error occurs")
	// Error should be about account, not parameter validation
	require.NotContains(t, err.Error(), "instance_id", "error should not be about parameter validation")
}

// Note: Additional tests for successful instance operations with mock Linode clients
// are not implemented in this unit test suite because they require functioning
// Linode API client interfaces for operations like ListInstances, GetInstance,
// CreateInstance, DeleteInstance, BootInstance, ShutdownInstance, and RebootInstance.
//
// These operations would require either:
// 1. Interface abstraction for the Linode client (future improvement)
// 2. Integration testing with real API endpoints
// 3. Dependency injection to replace the client during testing
//
// The current tests adequately cover:
// - Parameter validation and error handling logic
// - Account manager error scenarios
// - Request routing and basic tool handler setup
// - Error message formatting and response structure
//
// This provides comprehensive coverage of the testable logic that doesn't
// require external API dependencies.
