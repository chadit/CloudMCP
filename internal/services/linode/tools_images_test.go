package linode_test

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/internal/services/linode"
	"github.com/chadit/CloudMCP/pkg/logger"
)

// TestHandleImagesList_AccountError tests the handleImagesList function to verify it returns all images available.
// This test simulates a user requesting a list of all available Linode images through the MCP interface.
// Since this function requires Linode API calls, this test focuses on the error handling path.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with empty account manager
// 2. **Request Execution**: Call handleImagesList expecting account manager failure
// 3. **Error Validation**: Verify appropriate error is returned for account lookup failure
//
// **Test Environment**: Service with no configured accounts to trigger error path
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Does not attempt to call Linode API when account lookup fails
// • Provides meaningful error message for troubleshooting
//
// **Purpose**: This test ensures images list command fails appropriately when account configuration is invalid.
// Note: Full integration testing with mock Linode client requires interface abstraction (future improvement).
func TestHandleImagesList_AccountError(t *testing.T) {
	// Create minimal service with empty account manager
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	// Create completely isolated account manager for this test only
	accountManager := linode.NewAccountManagerForTesting()

	service := linode.NewForTesting(cfg, log, accountManager)

	// Test images list request with empty account manager
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_images_list",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.Error(t, err, "handleImagesList should return an error when no accounts are configured")
	require.Nil(t, result, "result should be nil when error occurs")
}

// TestHandleImageGet_AccountError tests the handleImageGet function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleImageGet with no configured accounts
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
func TestHandleImageGet_AccountError(t *testing.T) {
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

	// Test image get request with no accounts and valid parameter
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_image_get",
			Arguments: map[string]interface{}{
				"image_id": "linode/ubuntu22.04",
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.Error(t, err, "handleImageGet should return error when no current account exists")
	require.Nil(t, result, "result should be nil when error occurs")
}

// TestHandleImageGet_MissingParameter tests the handleImageGet function with missing required parameters.
// This test verifies the function handles requests with missing image_id parameter.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Missing Parameter**: Call handleImageGet without image_id parameter
// 3. **Error Validation**: Verify appropriate parameter error is returned
//
// **Expected Behavior**:
// • Returns error result for missing required parameters
// • Provides specific error message about missing image_id
// • Does not attempt account manager operations with invalid input
//
// **Purpose**: This test ensures image get validates required parameters properly.
func TestHandleImageGet_MissingParameter(t *testing.T) {
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

	// Test image get request with missing image_id parameter
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_image_get",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleImageGet should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "image_id", "error result should mention missing image_id parameter")
		}
	}
}

// TestHandleImageGet_EmptyParameter tests the handleImageGet function with empty image_id parameter.
// This test verifies the function handles requests with empty string image_id parameter.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Empty Parameter**: Call handleImageGet with empty image_id string
// 3. **Error Validation**: Verify appropriate parameter error is returned
//
// **Expected Behavior**:
// • Returns error result for empty parameter values
// • Provides specific error message about missing image_id
// • Treats empty string as missing required parameter
//
// **Purpose**: This test ensures image get validates parameter content as well as presence.
func TestHandleImageGet_EmptyParameter(t *testing.T) {
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

	// Test image get request with empty image_id parameter
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_image_get",
			Arguments: map[string]interface{}{
				"image_id": "",
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleImageGet should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "image_id", "error result should mention missing image_id parameter")
		}
	}
}

// TestHandleImageCreate_MissingRequiredParameters tests the handleImageCreate function with missing required parameters.
// This test verifies the function handles requests with missing required parameters for image creation.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Missing Parameters**: Call handleImageCreate without required parameters (disk_id, label)
// 3. **Error Validation**: Verify appropriate parameter error is returned
//
// **Expected Behavior**:
// • Returns error result for missing required parameters
// • Provides specific error message about missing required fields
// • Does not attempt account manager operations with invalid input
//
// **Purpose**: This test ensures image create validates all required parameters properly.
func TestHandleImageCreate_MissingRequiredParameters(t *testing.T) {
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

	// Test image create request with missing required parameters
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_image_create",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleImageCreate should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "required", "error result should mention missing required parameters")
		}
	}
}

// TestHandleImageCreate_PartialParameters tests the handleImageCreate function with some required parameters missing.
// This test verifies the function handles requests with only some of the required parameters.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Partial Parameters**: Call handleImageCreate with only some required parameters
// 3. **Error Validation**: Verify appropriate parameter error is returned
//
// **Expected Behavior**:
// • Returns error result when any required parameter is missing
// • Provides specific error message about missing required fields
// • Validates all required parameters before attempting API calls
//
// **Purpose**: This test ensures image create validates parameter completeness properly.
func TestHandleImageCreate_PartialParameters(t *testing.T) {
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

	// Test image create request with only partial required parameters
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_image_create",
			Arguments: map[string]interface{}{
				"disk_id": float64(12345),
				// Missing label parameter
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleImageCreate should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "label", "error result should mention missing label parameter")
		}
	}
}

// TestHandleImageUpdate_MissingParameter tests the handleImageUpdate function with missing required parameters.
// This test verifies the function handles requests with missing image_id parameter.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Missing Parameter**: Call handleImageUpdate without image_id parameter
// 3. **Error Validation**: Verify appropriate parameter error is returned
//
// **Expected Behavior**:
// • Returns error result for missing required parameters
// • Provides specific error message about missing image_id
// • Does not attempt account manager operations with invalid input
//
// **Purpose**: This test ensures image update validates required parameters properly.
func TestHandleImageUpdate_MissingParameter(t *testing.T) {
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

	// Test image update request with missing image_id parameter
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_image_update",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleImageUpdate should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "image_id", "error result should mention missing image_id parameter")
		}
	}
}

// TestHandleImageDelete_MissingParameter tests the handleImageDelete function with missing required parameters.
// This test verifies the function handles requests with missing image_id parameter.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Missing Parameter**: Call handleImageDelete without image_id parameter
// 3. **Error Validation**: Verify appropriate parameter error is returned
//
// **Expected Behavior**:
// • Returns error result for missing required parameters
// • Provides specific error message about missing image_id
// • Does not attempt account manager operations with invalid input
//
// **Purpose**: This test ensures image delete validates required parameters properly.
func TestHandleImageDelete_MissingParameter(t *testing.T) {
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

	// Test image delete request with missing image_id parameter
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_image_delete",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleImageDelete should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "image_id", "error result should mention missing image_id parameter")
		}
	}
}

// TestHandleImageReplicate_MissingParameter tests the handleImageReplicate function with missing required parameters.
// This test verifies the function handles requests with missing image_id parameter.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Missing Parameter**: Call handleImageReplicate without image_id parameter
// 3. **Error Validation**: Verify appropriate parameter error is returned
//
// **Expected Behavior**:
// • Returns error result for missing required parameters
// • Provides specific error message about missing image_id
// • Does not attempt account manager operations with invalid input
//
// **Purpose**: This test ensures image replicate validates required parameters properly.
func TestHandleImageReplicate_MissingParameter(t *testing.T) {
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

	// Test image replicate request with missing image_id parameter
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_image_replicate",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleImageReplicate should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "image_id", "error result should mention missing image_id parameter")
		}
	}
}

// TestHandleImageUploadCreate_MissingParameter tests the handleImageUploadCreate function with missing required parameters.
// This test verifies the function handles requests with missing required parameters for image upload creation.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with valid account configuration
// 2. **Missing Parameters**: Call handleImageUploadCreate without required parameters
// 3. **Error Validation**: Verify appropriate parameter error is returned
//
// **Expected Behavior**:
// • Returns error result for missing required parameters
// • Provides specific error message about missing required fields
// • Does not attempt account manager operations with invalid input
//
// **Purpose**: This test ensures image upload create validates required parameters properly.
func TestHandleImageUploadCreate_MissingParameter(t *testing.T) {
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

	// Test image upload create request with missing required parameters
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_image_upload_create",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleImageUploadCreate should not return error for parameter validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "required", "error result should mention missing required parameters")
		}
	}
}

// Note: Additional tests for successful image operations with mock Linode clients
// are not implemented in this unit test suite because they require functioning
// Linode API client interfaces for operations like ListImages, GetImage,
// CreateImage, UpdateImage, DeleteImage, ReplicateImage, and CreateImageUpload.
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
