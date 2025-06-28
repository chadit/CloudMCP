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

// TestHandleAccountGet_AccountError tests the handleAccountGet function robustness with account manager errors.
// This test verifies the function handles account lookup failures gracefully.
// Since handleAccountGet requires a real Linode API client, this test focuses on the testable error path.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager (no accounts configured)
// 2. **Request Execution**: Call handleAccountGet expecting account manager failure
// 3. **Error Validation**: Verify appropriate error is returned
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Does not attempt to call Linode API when account lookup fails
// • Provides meaningful error message for troubleshooting
//
// **Purpose**: This test ensures account command fails appropriately when account configuration is invalid.
// Note: Full integration testing with mock Linode client requires interface abstraction (future improvement).
func TestHandleAccountGet_AccountError(t *testing.T) {
	// Create minimal service with empty account manager
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	// Create completely isolated account manager for this test only
	accountManager := linode.NewAccountManagerForTesting()

	service := linode.NewForTesting(cfg, log, accountManager)

	// Test account get request with empty account manager
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_account_get",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.Error(t, err, "handleAccountGet should return an error when no accounts are configured")
	require.Nil(t, result, "result should be nil when error occurs")
}

// TestHandleAccountList tests the handleAccountList function to verify it returns all configured accounts.
// This test simulates a user requesting a list of all available accounts through the MCP interface.
// This function doesn't require API calls, making it fully testable with mocks.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with multiple mock accounts
// 2. **Request Execution**: Call handleAccountList with empty request
// 3. **Response Validation**: Verify account list format and content
// 4. **Current Account Verification**: Check that current account is properly marked
//
// **Test Environment**: Isolated mock account manager with test accounts "test-account", "dev-account"
//
// **Expected Behavior**:
// • Returns successful tool result with formatted text output
// • Lists all configured accounts with names and labels
// • Marks current account with "(current)" indicator
// • Shows current account name at the top of the output
//
// **Purpose**: This test ensures account list command provides complete multi-account visibility.
func TestHandleAccountList(t *testing.T) {
	// Create completely isolated test service for this test only
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "test-account",
		LinodeAccounts: map[string]config.LinodeAccount{
			"test-account": {Label: "Test Account", Token: "test-token"},
			"dev-account":  {Label: "Development Account", Token: "dev-token"},
		},
	}

	// Create completely isolated account manager - no shared state
	accountManager := linode.NewAccountManagerForTesting()

	// Add test accounts
	testAccount := linode.NewAccountForTesting("test-account", "Test Account")
	devAccount := linode.NewAccountForTesting("dev-account", "Development Account")

	accountManager.AddAccountForTesting(testAccount)
	accountManager.AddAccountForTesting(devAccount)
	accountManager.SetCurrentAccountForTesting("test-account")

	// Create isolated service instance
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test account list request
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_account_list",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleAccountList should not return an error")

	// Verify account list content
	listText := linode.GetTextContentForTesting(t, result)
	require.Contains(t, listText, "Current account: test-account", "should show current account")
	require.Contains(t, listText, "Configured accounts:", "should have accounts section header")
	require.Contains(t, listText, "test-account: Test Account (current)", "should mark current account")
	require.Contains(t, listText, "dev-account: Development Account", "should list additional accounts")

	// Verify account manager state (isolated to this test)
	accounts := accountManager.ListAccounts()
	require.Len(t, accounts, 2, "should have both test accounts")
	require.Contains(t, accounts, "test-account", "should contain primary test account")
	require.Contains(t, accounts, "dev-account", "should contain additional test account")
}

// TestHandleAccountList_SingleAccount tests the handleAccountList function with only one account configured.
// This test verifies the function works correctly in single-account scenarios.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with single mock account
// 2. **Request Execution**: Call handleAccountList with single account configuration
// 3. **Response Validation**: Verify single account is listed and marked as current
//
// **Expected Behavior**:
// • Returns successful tool result with single account information
// • Marks the only account as current
// • Displays proper formatting for single-account case
//
// **Purpose**: This test ensures account list works correctly in minimal configuration scenarios.
func TestHandleAccountList_SingleAccount(t *testing.T) {
	// Create completely isolated test service for this test only
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "single-account",
		LinodeAccounts: map[string]config.LinodeAccount{
			"single-account": {Label: "Single Account", Token: "single-token"},
		},
	}

	// Create completely isolated account manager - no shared state
	accountManager := linode.NewAccountManagerForTesting()

	// Add single test account
	singleAccount := linode.NewAccountForTesting("single-account", "Single Account")
	accountManager.AddAccountForTesting(singleAccount)
	accountManager.SetCurrentAccountForTesting("single-account")

	// Create isolated service instance
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test account list request
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_account_list",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleAccountList should not return an error")

	// Verify single account content
	listText := linode.GetTextContentForTesting(t, result)
	require.Contains(t, listText, "Current account: single-account", "should show current account")
	require.Contains(t, listText, "single-account: Single Account (current)", "should mark single account as current")

	// Verify account manager state (isolated to this test)
	accounts := accountManager.ListAccounts()
	require.Len(t, accounts, 1, "should have single account")
}

// TestHandleAccountSwitch_InvalidAccount tests the handleAccountSwitch function with invalid account names.
// This test verifies the function handles switch requests to non-existent accounts gracefully.
// Note: This test focuses on parameter validation and error handling, avoiding API calls.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with known accounts
// 2. **Invalid Switch**: Call handleAccountSwitch with non-existent account name
// 3. **Error Validation**: Verify appropriate error response is returned
// 4. **State Preservation**: Check that current account remains unchanged
//
// **Expected Behavior**:
// • Returns error result (not error exception) for invalid account names
// • Provides meaningful error message for troubleshooting
// • Preserves current account state when switch fails
// • Does not attempt to call Linode API for verification
//
// **Purpose**: This test ensures account switching fails gracefully with invalid input.
func TestHandleAccountSwitch_InvalidAccount(t *testing.T) {
	// Create completely isolated test service for this test only
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "original-account",
		LinodeAccounts: map[string]config.LinodeAccount{
			"original-account": {Label: "Original Account", Token: "original-token"},
		},
	}

	// Create completely isolated account manager - no shared state
	accountManager := linode.NewAccountManagerForTesting()

	// Add original test account
	originalAccount := linode.NewAccountForTesting("original-account", "Original Account")
	accountManager.AddAccountForTesting(originalAccount)
	accountManager.SetCurrentAccountForTesting("original-account")

	// Create isolated service instance
	service := linode.NewForTesting(cfg, log, accountManager)

	// Remember original current account
	currentAccountBeforeSwitch, err := accountManager.GetCurrentAccount()
	require.NoError(t, err, "should be able to get original current account")

	// Test account switch with invalid account name
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_account_switch",
			Arguments: map[string]interface{}{
				"account_name": "nonexistent-account",
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleAccountSwitch should not return exception for invalid account")
	require.NotNil(t, result, "result should not be nil")

	// Verify error response content
	switchText := linode.GetTextContentForTesting(t, result)
	require.Contains(t, switchText, "Failed to switch account", "should contain error message")

	// Verify current account was not changed
	currentAccountAfterSwitch, err := accountManager.GetCurrentAccount()
	require.NoError(t, err, "should be able to get current account after failed switch")
	require.Equal(t, currentAccountBeforeSwitch.Name, currentAccountAfterSwitch.Name, "current account should remain unchanged")
}

// TestHandleAccountSwitch_MissingParameter tests the handleAccountSwitch function with missing required parameters.
// This test verifies the function handles requests with missing account_name parameter.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid configuration
// 2. **Missing Parameter**: Call handleAccountSwitch without account_name parameter
// 3. **Error Validation**: Verify appropriate parameter error is returned
// 4. **State Preservation**: Check that current account remains unchanged
//
// **Expected Behavior**:
// • Returns error result for missing required parameters
// • Provides specific error message about missing account_name
// • Preserves current account state when parameters are invalid
// • Does not attempt account manager operations with invalid input
//
// **Purpose**: This test ensures account switching validates required parameters properly.
func TestHandleAccountSwitch_MissingParameter(t *testing.T) {
	// Create completely isolated test service for this test only
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "param-test-account",
		LinodeAccounts: map[string]config.LinodeAccount{
			"param-test-account": {Label: "Parameter Test Account", Token: "param-token"},
		},
	}

	// Create completely isolated account manager - no shared state
	accountManager := linode.NewAccountManagerForTesting()

	// Add parameter test account
	paramTestAccount := linode.NewAccountForTesting("param-test-account", "Parameter Test Account")
	accountManager.AddAccountForTesting(paramTestAccount)
	accountManager.SetCurrentAccountForTesting("param-test-account")

	// Create isolated service instance
	service := linode.NewForTesting(cfg, log, accountManager)

	// Remember original current account
	originalAccount, err := accountManager.GetCurrentAccount()
	require.NoError(t, err, "should be able to get original current account")

	// Test account switch with missing account_name parameter
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_account_switch",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleAccountSwitch should not return exception for missing parameter")
	require.NotNil(t, result, "result should not be nil")

	// Verify parameter error content
	switchText := linode.GetTextContentForTesting(t, result)
	require.Contains(t, switchText, "account_name is required", "should contain parameter error message")

	// Verify current account was not changed
	currentAccount, err := accountManager.GetCurrentAccount()
	require.NoError(t, err, "should be able to get current account after failed switch")
	require.Equal(t, originalAccount.Name, currentAccount.Name, "current account should remain unchanged")
}

// TestHandleAccountSwitch_EmptyParameter tests the handleAccountSwitch function with empty account_name parameter.
// This test verifies the function handles requests with empty string account names.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid configuration
// 2. **Empty Parameter**: Call handleAccountSwitch with empty account_name string
// 3. **Error Validation**: Verify appropriate parameter error is returned
//
// **Expected Behavior**:
// • Returns error result for empty account_name parameter
// • Treats empty string as missing required parameter
// • Provides same error message as missing parameter case
//
// **Purpose**: This test ensures account switching validates parameter content as well as presence.
func TestHandleAccountSwitch_EmptyParameter(t *testing.T) {
	// Create completely isolated test service for this test only
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "empty-test-account",
		LinodeAccounts: map[string]config.LinodeAccount{
			"empty-test-account": {Label: "Empty Test Account", Token: "empty-token"},
		},
	}

	// Create completely isolated account manager - no shared state
	accountManager := linode.NewAccountManagerForTesting()

	// Add empty test account
	emptyTestAccount := linode.NewAccountForTesting("empty-test-account", "Empty Test Account")
	accountManager.AddAccountForTesting(emptyTestAccount)
	accountManager.SetCurrentAccountForTesting("empty-test-account")

	// Create isolated service instance
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test account switch with empty account_name parameter
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_account_switch",
			Arguments: map[string]interface{}{
				"account_name": "",
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleAccountSwitch should not return exception for empty parameter")
	require.NotNil(t, result, "result should not be nil")

	// Verify parameter error content
	switchText := linode.GetTextContentForTesting(t, result)
	require.Contains(t, switchText, "account_name is required", "should contain parameter error message")
}

// Note: TestHandleAccountSwitch_ValidSwitch is not implemented in this unit test suite
// because it requires a functioning Linode API client for the verification step.
// The handleAccountSwitch function calls account.Client.GetProfile(ctx) to verify
// the switched account, which cannot be easily mocked with the current architecture.
//
// This test would require either:
// 1. Interface abstraction for the Linode client (future improvement)
// 2. Integration testing with real API endpoints
// 3. Dependency injection to replace the client during testing
//
// The other account switching tests (invalid account, missing parameters, etc.)
// adequately cover the parameter validation and error handling logic that can
// be unit tested without API dependencies.
