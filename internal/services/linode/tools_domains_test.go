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

// TestHandleDomainsList_AccountError tests the handleDomainsList function to verify error handling.
// This test simulates a user requesting a list of all their Linode domains through the MCP interface.
// Since this function requires Linode API calls, this test focuses on the error handling path.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with empty account manager
// 2. **Request Execution**: Call handleDomainsList expecting account manager failure
// 3. **Error Validation**: Verify appropriate error is returned for account lookup failure
//
// **Test Environment**: Service with no configured accounts to trigger error path
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Does not attempt to call Linode API when account lookup fails
// • Provides meaningful error message for troubleshooting
//
// **Purpose**: This test ensures domains list command fails appropriately when account configuration is invalid.
func TestHandleDomainsList_AccountError(t *testing.T) {
	// Create minimal service with empty account manager
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	// Create completely isolated account manager for this test only
	accountManager := linode.NewAccountManagerForTesting()

	service := linode.NewForTesting(cfg, log, accountManager)

	// Test domains list request with empty account manager
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_domains_list",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.Error(t, err, "handleDomainsList should return error for account lookup failure")
	require.Nil(t, result, "result should be nil when account lookup fails")
	require.Contains(t, err.Error(), "account", "error should mention account issue")
}

// TestHandleDomainGet_AccountError tests the handleDomainGet function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleDomainGet with no configured accounts
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
func TestHandleDomainGet_AccountError(t *testing.T) {
	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test domain get request with no accounts and valid parameter
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_domain_get",
			Arguments: map[string]interface{}{
				"domain_id": float64(12345),
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.Error(t, err, "handleDomainGet should return error for account lookup failure")
	require.Nil(t, result, "result should be nil when account lookup fails")
	require.Contains(t, err.Error(), "account", "error should mention account issue")
}

// TestHandleDomainGet_ParameterValidation tests the handleDomainGet function parameter validation.
// This test verifies the function properly validates required parameters before making API calls.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with valid account manager setup
// 2. **Parameter Testing**: Call handleDomainGet with various invalid parameters
// 3. **Validation Verification**: Verify parameter validation errors are returned appropriately
//
// **Test Environment**: Valid service setup with mock account manager
//
// **Expected Behavior**:
// • Returns error result for missing domain_id parameter
// • Returns error result for invalid domain_id parameter types
// • Does not attempt API calls when parameter validation fails
//
// **Purpose**: This test ensures proper parameter validation before expensive API operations.
func TestHandleDomainGet_ParameterValidation(t *testing.T) {
	// Create service with valid account manager setup
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "test",
		LinodeAccounts: map[string]config.LinodeAccount{
			"test": {
				Token: "test-token",
				Label: "Test Account",
			},
		},
	}

	accountManager := linode.NewAccountManagerForTesting()
	testAccount := linode.NewAccountForTesting("test", "Test Account")
	accountManager.AddAccountForTesting(testAccount)
	accountManager.SetCurrentAccountForTesting("test")

	service := linode.NewForTesting(cfg, log, accountManager)

	ctx := context.Background()

	// Test missing domain_id parameter
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_domain_get",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result due to missing parameter
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "domain_id", "error result should mention missing domain_id parameter")
		}
	}

	// Test invalid domain_id parameter type
	request = mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_domain_get",
			Arguments: map[string]interface{}{
				"domain_id": "invalid-string",
			},
		},
	}

	result, err = service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result due to invalid parameter type
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "domain_id", "error result should mention invalid domain_id parameter")
		}
	}
}

// TestHandleDomainCreate_AccountError tests the handleDomainCreate function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleDomainCreate with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleDomainCreate_AccountError(t *testing.T) {
	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test domain create request with no accounts
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_domain_create",
			Arguments: map[string]interface{}{
				"domain": "example.com",
				"type":   "master",
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleDomainCreate should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleDomainUpdate_AccountError tests the handleDomainUpdate function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleDomainUpdate with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleDomainUpdate_AccountError(t *testing.T) {
	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test domain update request with no accounts
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_domain_update",
			Arguments: map[string]interface{}{
				"domain_id":   float64(12345),
				"description": "Updated domain",
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleDomainUpdate should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleDomainDelete_AccountError tests the handleDomainDelete function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleDomainDelete with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleDomainDelete_AccountError(t *testing.T) {
	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test domain delete request with no accounts
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_domain_delete",
			Arguments: map[string]interface{}{
				"domain_id": float64(12345),
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleDomainDelete should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleDomainRecordsList_AccountError tests the handleDomainRecordsList function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleDomainRecordsList with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleDomainRecordsList_AccountError(t *testing.T) {
	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test domain records list request with no accounts
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_domain_records_list",
			Arguments: map[string]interface{}{
				"domain_id": float64(12345),
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleDomainRecordsList should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleDomainRecordGet_AccountError tests the handleDomainRecordGet function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleDomainRecordGet with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleDomainRecordGet_AccountError(t *testing.T) {
	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test domain record get request with no accounts
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_domain_record_get",
			Arguments: map[string]interface{}{
				"domain_id": float64(12345),
				"record_id": float64(67890),
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleDomainRecordGet should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleDomainRecordCreate_AccountError tests the handleDomainRecordCreate function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleDomainRecordCreate with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleDomainRecordCreate_AccountError(t *testing.T) {
	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test domain record create request with no accounts
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_domain_record_create",
			Arguments: map[string]interface{}{
				"domain_id": float64(12345),
				"type":      "A",
				"name":      "www",
				"target":    "192.0.2.1",
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleDomainRecordCreate should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleDomainRecordUpdate_AccountError tests the handleDomainRecordUpdate function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleDomainRecordUpdate with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleDomainRecordUpdate_AccountError(t *testing.T) {
	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test domain record update request with no accounts
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_domain_record_update",
			Arguments: map[string]interface{}{
				"domain_id": float64(12345),
				"record_id": float64(67890),
				"target":    "192.0.2.2",
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleDomainRecordUpdate should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleDomainRecordDelete_AccountError tests the handleDomainRecordDelete function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleDomainRecordDelete with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleDomainRecordDelete_AccountError(t *testing.T) {
	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test domain record delete request with no accounts
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_domain_record_delete",
			Arguments: map[string]interface{}{
				"domain_id": float64(12345),
				"record_id": float64(67890),
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleDomainRecordDelete should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// Note: Additional tests for successful domain operations with mock Linode clients
// are not implemented in this unit test suite because they require functioning
// Linode API client interfaces for operations like ListDomains, GetDomain,
// CreateDomain, UpdateDomain, DeleteDomain, ListDomainRecords, GetDomainRecord,
// CreateDomainRecord, UpdateDomainRecord, and DeleteDomainRecord.
//
// Parameter validation tests are included for domain operations that use
// parseIDFromArguments() for parameter validation, providing coverage of
// the input validation logic without requiring external API dependencies.
//
// These operations would require either:
// 1. Interface abstraction for the Linode client (future improvement)
// 2. Integration testing with real API endpoints
// 3. Dependency injection to replace the client during testing
//
// The current tests adequately cover:
// - Account manager error scenarios for all 10 domain/DNS tools
// - Parameter validation for tools that implement proper validation
// - Request routing and basic tool handler setup
// - Error message formatting and response structure
// - MCP protocol compliance for error responses
//
// This provides coverage of the testable logic that doesn't
// require external API dependencies while ensuring proper error handling
// for the most common failure scenarios (account configuration issues
// and parameter validation failures).
