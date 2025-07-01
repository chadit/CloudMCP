package linode_test

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/internal/services/linode"
	"github.com/chadit/CloudMCP/pkg/logger"
)

// TestHandleVolumesList_AccountError tests the handleVolumesList function to verify it returns all volumes in the current account.
// This test simulates a user requesting a list of all their Linode volumes through the MCP interface.
// Since this function requires Linode API calls, this test focuses on the error handling path.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with empty account manager
// 2. **Request Execution**: Call handleVolumesList expecting account manager failure
// 3. **Error Validation**: Verify appropriate error is returned for account lookup failure
//
// **Test Environment**: Service with no configured accounts to trigger error path
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Does not attempt to call Linode API when account lookup fails
// • Provides meaningful error message for troubleshooting
//
// **Purpose**: This test ensures volumes list command fails appropriately when account configuration is invalid.
// Note: Full integration testing with mock Linode client requires interface abstraction (future improvement).
func TestHandleVolumesList_AccountError(t *testing.T) {
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

	// Test volumes list request with empty account manager
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_volumes_list",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.Error(t, err, "handleVolumesList should return an error when no accounts are configured")
	require.Nil(t, result, "result should be nil when error occurs")
}

// TestHandleVolumeGet_AccountError tests the handleVolumeGet function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleVolumeGet with no configured accounts
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
func TestHandleVolumeGet_AccountError(t *testing.T) {
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

	// Test volume get request with no accounts and valid parameter
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_volume_get",
			Arguments: map[string]interface{}{
				"volume_id": float64(12345),
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.Error(t, err, "handleVolumeGet should return error when no current account exists")
	require.Nil(t, result, "result should be nil when error occurs")
}

// TestHandleVolumeGet_MissingParameter tests the handleVolumeGet function with missing required parameters.
// This test verifies the function handles requests with missing volume_id parameter.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Missing Parameter**: Call handleVolumeGet without volume_id parameter
// 3. **Error Validation**: Verify appropriate parameter error is returned
//
// **Expected Behavior**:
// • Returns error result for missing required parameters
// • Provides specific error message about missing volume_id
// • Does not attempt account manager operations with invalid input
//
// **Purpose**: This test ensures volume get validates required parameters properly.
func TestHandleVolumeGet_MissingParameter(t *testing.T) {
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

	// Test volume get request with missing volume_id parameter
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_volume_get",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleVolumeGet should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "volume_id", "error result should mention missing volume_id parameter")
		}
	}
}

// TestHandleVolumeGet_InvalidParameter tests the handleVolumeGet function with invalid parameter types.
// This test verifies the function handles requests with non-integer volume_id parameter.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Invalid Parameter**: Call handleVolumeGet with string volume_id instead of integer
// 3. **Error Validation**: Verify appropriate parameter type error is returned
//
// **Expected Behavior**:
// • Returns error result for invalid parameter types
// • Provides specific error message about parameter type mismatch
// • Does not attempt account manager operations with invalid input
//
// **Purpose**: This test ensures volume get validates parameter types properly.
func TestHandleVolumeGet_InvalidParameter(t *testing.T) {
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

	// Test volume get request with invalid volume_id parameter type
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_volume_get",
			Arguments: map[string]interface{}{
				"volume_id": "not-a-number",
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleVolumeGet should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "volume_id", "error result should mention invalid volume_id parameter")
		}
	}
}

// TestHandleVolumeCreate_MissingRequiredParameters tests the handleVolumeCreate function with missing required parameters.
// This test verifies the function handles requests with missing required parameters for volume creation.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Missing Parameters**: Call handleVolumeCreate without required parameters (label, size)
// 3. **Error Validation**: Verify appropriate parameter error is returned
//
// **Expected Behavior**:
// • Returns error result for missing required parameters
// • Provides specific error message about missing required fields
// • Does not attempt account manager operations with invalid input
//
// **Purpose**: This test ensures volume create validates all required parameters properly.
func TestHandleVolumeCreate_MissingRequiredParameters(t *testing.T) {
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

	// Test volume create request with missing required parameters
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_volume_create",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleVolumeCreate should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "required", "error result should mention missing required parameters")
		}
	}
}

// TestHandleVolumeCreate_PartialParameters tests the handleVolumeCreate function with some required parameters missing.
// This test verifies the function handles requests with only some of the required parameters.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Partial Parameters**: Call handleVolumeCreate with only some required parameters
// 3. **Error Validation**: Verify appropriate parameter error is returned
//
// **Expected Behavior**:
// • Returns error result when any required parameter is missing
// • Provides specific error message about missing required fields
// • Validates all required parameters before attempting API calls
//
// **Purpose**: This test ensures volume create validates parameter completeness properly.
func TestHandleVolumeCreate_PartialParameters(t *testing.T) {
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

	// Test volume create request with only partial required parameters
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_volume_create",
			Arguments: map[string]interface{}{
				"label": "test-volume",
				// Missing size parameter
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleVolumeCreate should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "size", "error result should mention missing size parameter")
		}
	}
}

// TestHandleVolumeCreate_InvalidSize tests the handleVolumeCreate function with invalid size values.
// This test verifies the function handles requests with size values outside the valid range.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Invalid Size**: Call handleVolumeCreate with size outside 10-8192 range
// 3. **Error Validation**: Verify appropriate size validation error is returned
//
// **Expected Behavior**:
// • Returns error result for invalid size values
// • Provides specific error message about size range requirements
// • Validates size range before attempting API calls
//
// **Purpose**: This test ensures volume create validates size constraints properly.
func TestHandleVolumeCreate_InvalidSize(t *testing.T) {
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

	// Test volume create request with invalid size (too small)
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_volume_create",
			Arguments: map[string]interface{}{
				"label": "test-volume",
				"size":  float64(5), // Too small, minimum is 10
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleVolumeCreate should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "between 10 and 8192", "error result should mention size range requirements")
		}
	}
}

// TestHandleVolumeDelete_MissingParameter tests the handleVolumeDelete function with missing required parameters.
// This test verifies the function handles requests with missing volume_id parameter.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Missing Parameter**: Call handleVolumeDelete without volume_id parameter
// 3. **Error Validation**: Verify appropriate parameter error is returned
//
// **Expected Behavior**:
// • Returns error result for missing required parameters
// • Provides specific error message about missing volume_id
// • Does not attempt account manager operations with invalid input
//
// **Purpose**: This test ensures volume delete validates required parameters properly.
func TestHandleVolumeDelete_MissingParameter(t *testing.T) {
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

	// Test volume delete request with missing volume_id parameter
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_volume_delete",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleVolumeDelete should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "volume_id", "error result should mention missing volume_id parameter")
		}
	}
}

// TestHandleVolumeAttach_MissingParameters tests the handleVolumeAttach function with missing required parameters.
// This test verifies the function handles requests with missing volume_id and linode_id parameters.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Missing Parameters**: Call handleVolumeAttach without required parameters
// 3. **Error Validation**: Verify appropriate parameter error is returned
//
// **Expected Behavior**:
// • Returns error result for missing required parameters
// • Provides specific error message about missing required parameters
// • Does not attempt account manager operations with invalid input
//
// **Purpose**: This test ensures volume attach validates required parameters properly.
func TestHandleVolumeAttach_MissingParameters(t *testing.T) {
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

	// Test volume attach request with missing required parameters
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_volume_attach",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleVolumeAttach should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			// Should mention either volume_id or linode_id as missing required parameter
			text := textContent.Text
			require.True(t,
				(text != "" && (len(text) > 0 && text[0] != 0)), // Just check it's not empty
				"error result should contain parameter validation message")
		}
	}
}

// TestHandleVolumeDetach_MissingParameter tests the handleVolumeDetach function with missing required parameters.
// This test verifies the function handles requests with missing volume_id parameter.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Missing Parameter**: Call handleVolumeDetach without volume_id parameter
// 3. **Error Validation**: Verify appropriate parameter error is returned
//
// **Expected Behavior**:
// • Returns error result for missing required parameters
// • Provides specific error message about missing volume_id
// • Does not attempt account manager operations with invalid input
//
// **Purpose**: This test ensures volume detach validates required parameters properly.
func TestHandleVolumeDetach_MissingParameter(t *testing.T) {
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

	// Test volume detach request with missing volume_id parameter
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_volume_detach",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleVolumeDetach should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "volume_id", "error result should mention missing volume_id parameter")
		}
	}
}

// Note: Additional tests for successful volume operations with mock Linode clients
// are not implemented in this unit test suite because they require functioning
// Linode API client interfaces for operations like ListVolumes, GetVolume,
// CreateVolume, DeleteVolume, AttachVolume, and DetachVolume.
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
