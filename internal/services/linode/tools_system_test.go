package linode

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/internal/version"
	"github.com/chadit/CloudMCP/pkg/logger"
)

// Helper function to extract text content from CallToolResul
func getTextContent(t *testing.T, result *mcp.CallToolResult) string {
	require.NotNil(t, result, "result should not be nil")
	require.NotEmpty(t, result.Content, "result should have content")
	require.Equal(t, 1, len(result.Content), "result should have exactly one content item")

	textContent, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok, "result content should be text content")
	require.NotEmpty(t, textContent.Text, "result text should not be empty")

	return textContent.Text
}

// TestHandleSystemVersion tests the handleSystemVersion function to verify it returns comprehensive version information in human-readable format.
// This test simulates a user requesting CloudMCP version details through the MCP interface.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with mock account manager
// 2. **Request Execution**: Call handleSystemVersion with empty reques
// 3. **Response Validation**: Verify version information format and conten
// 4. **Content Verification**: Check for presence of all required version fields
// 5. **Account Information**: Ensure current account name is included
//
// **Test Environment**: Mock account manager with test account "test-account (Test Account)"
//
// **Expected Behavior**:
// • Returns successful tool result with formatted text outpu
// • Includes CloudMCP version, API version, build information
// • Contains Git commit, branch, Go version, and platform details
// • Lists all feature flags (API coverage, multi-account, metrics, etc.)
// • Shows current account name from account manager
//
// **Purpose**: This test ensures version command provides complete diagnostic information for troubleshooting and verification.
func TestHandleSystemVersion(t *testing.T) {
	// Create test service
	service, mockAccountManager, _ := CreateTestService()

	// Test system version reques
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "cloudmcp_version",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.handleSystemVersion(ctx, request)
	require.NoError(t, err, "handleSystemVersion should not return an error")

	// Verify version information conten
	versionText := getTextContent(t, result)
	require.Contains(t, versionText, "CloudMCP Version Information", "should contain version header")
	require.Contains(t, versionText, "Version: 1.0.0", "should contain CloudMCP version")
	require.Contains(t, versionText, "API Version: 4.0", "should contain API version")
	require.Contains(t, versionText, "Linode API Coverage: 100%", "should show 100% API coverage")
	require.Contains(t, versionText, "Multi-Account Support: enabled", "should show multi-account support")
	require.Contains(t, versionText, "Current Account: test-account (Test Account)", "should show current account")

	// Verify all required sections are presen
	require.Contains(t, versionText, "Git Commit:", "should contain git commit")
	require.Contains(t, versionText, "Go Version:", "should contain Go version")
	require.Contains(t, versionText, "Platform:", "should contain platform info")
	require.Contains(t, versionText, "Features:", "should contain features section")

	// Verify account manager was used
	account, err := mockAccountManager.GetCurrentAccount()
	require.NoError(t, err, "should be able to get current account")
	require.Equal(t, "test-account", account.Name, "should have correct account name")
}

// TestHandleSystemVersionJSON tests the handleSystemVersionJSON function to verify it returns version information in structured JSON format.
// This test simulates an API client requesting machine-readable version data for automated processing.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with mock account manager
// 2. **Request Execution**: Call handleSystemVersionJSON with empty reques
// 3. **JSON Parsing**: Unmarshal response to verify valid JSON structure
// 4. **Data Validation**: Check all required fields and nested objects
// 5. **Account Verification**: Ensure current account information is included
//
// **Test Environment**: Mock account manager with test account "test-account (Test Account)"
//
// **Expected Behavior**:
// • Returns successful tool result with valid JSON outpu
// • Contains structured "cloudmcp" object with all version fields
// • Includes "current_account" field with account name
// • Specifies "service" field as "linode"
// • JSON structure matches expected schema for programmatic consumption
//
// **Purpose**: This test ensures JSON version command provides machine-readable data for integration and monitoring tools.
func TestHandleSystemVersionJSON(t *testing.T) {
	// Create test service
	service, mockAccountManager, _ := CreateTestService()

	// Test system version JSON reques
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "cloudmcp_version_json",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.handleSystemVersionJSON(ctx, request)
	require.NoError(t, err, "handleSystemVersionJSON should not return an error")

	// Parse JSON response
	jsonText := getTextContent(t, result)
	var jsonResponse map[string]interface{}
	err = json.Unmarshal([]byte(jsonText), &jsonResponse)
	require.NoError(t, err, "response should be valid JSON")

	// Verify top-level structure
	require.Contains(t, jsonResponse, "cloudmcp", "should contain cloudmcp object")
	require.Contains(t, jsonResponse, "current_account", "should contain current_account field")
	require.Contains(t, jsonResponse, "service", "should contain service field")
	require.Equal(t, "linode", jsonResponse["service"], "service should be linode")
	require.Equal(t, "test-account (Test Account)", jsonResponse["current_account"], "should show current account")

	// Verify cloudmcp version objec
	cloudmcpObj, ok := jsonResponse["cloudmcp"].(map[string]interface{})
	require.True(t, ok, "cloudmcp should be an object")

	// Verify version fields
	require.Equal(t, "1.0.0", cloudmcpObj["version"], "should contain correct version")
	require.Equal(t, "4.0", cloudmcpObj["api_version"], "should contain correct API version")
	require.Contains(t, cloudmcpObj, "build_date", "should contain build_date")
	require.Contains(t, cloudmcpObj, "git_commit", "should contain git_commit")
	require.Contains(t, cloudmcpObj, "git_branch", "should contain git_branch")
	require.Contains(t, cloudmcpObj, "go_version", "should contain go_version")
	require.Contains(t, cloudmcpObj, "platform", "should contain platform")

	// Verify features objec
	features, ok := cloudmcpObj["features"].(map[string]interface{})
	require.True(t, ok, "features should be an object")
	require.Equal(t, "100%", features["linode_api_coverage"], "should show 100% API coverage")
	require.Equal(t, "enabled", features["multi_account"], "should show multi-account enabled")
	require.Equal(t, "prometheus", features["metrics"], "should show prometheus metrics")
	require.Equal(t, "structured", features["logging"], "should show structured logging")
	require.Equal(t, "mcp", features["protocol"], "should show MCP protocol")

	// Verify account manager was used
	account, err := mockAccountManager.GetCurrentAccount()
	require.NoError(t, err, "should be able to get current account")
	require.Equal(t, "test-account", account.Name, "should have correct account name")
}

// TestHandleSystemVersion_AccountError tests the handleSystemVersion function robustness.
// This test verifies the function continues to work even with account lookup issues.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleSystemVersion with empty account manager
// 3. **Response Validation**: Verify function provides version info despite account issues
//
// **Expected Behavior**:
// • Returns successful tool result regardless of account state
// • Provides all version information correctly
// • Handles account lookup gracefully with fallback
//
// **Purpose**: This test ensures version command robustness.
func TestHandleSystemVersion_AccountError(t *testing.T) {
	// Create minimal service for robustness testing
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	service := &Service{
		config: cfg,
		logger: log,
		accountManager: &AccountManager{
			accounts:       make(map[string]*Account),
			currentAccount: "",
			mu:             sync.RWMutex{},
		},
	}

	// Test system version request with empty account manager
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "cloudmcp_version",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.handleSystemVersion(ctx, request)
	require.NoError(t, err, "handleSystemVersion should not return an error even with empty accounts")

	// Verify robustness - should still provide version info
	versionText := getTextContent(t, result)
	require.Contains(t, versionText, "CloudMCP Version Information",
		"should still contain version header")
	require.Contains(t, versionText, "Version: 1.0.0",
		"should still contain CloudMCP version")
	// Note: Account will show as "unknown" when no accounts are configured
}

// TestHandleSystemVersionJSON_AccountError tests the handleSystemVersionJSON function robustness.
// This test verifies JSON output remains valid even with account lookup issues.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleSystemVersionJSON with empty account manager
// 3. **JSON Validation**: Parse and verify JSON structure remains valid
// 4. **Data Verification**: Ensure all version data is presen
//
// **Expected Behavior**:
// • Returns successful tool result with valid JSON regardless of account state
// • All version fields remain accurate and properly structured
// • JSON parsing continues to work without issues
//
// **Purpose**: This test ensures JSON version command robustness.
func TestHandleSystemVersionJSON_AccountError(t *testing.T) {
	// Create minimal service for robustness testing
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	service := &Service{
		config: cfg,
		logger: log,
		accountManager: &AccountManager{
			accounts:       make(map[string]*Account),
			currentAccount: "",
			mu:             sync.RWMutex{},
		},
	}

	// Test system version JSON request with empty account manager
	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "cloudmcp_version_json",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.handleSystemVersionJSON(ctx, request)
	require.NoError(t, err, "handleSystemVersionJSON should not return an error even with empty accounts")

	// Parse JSON response
	jsonText := getTextContent(t, result)
	var jsonResponse map[string]interface{}
	err = json.Unmarshal([]byte(jsonText), &jsonResponse)
	require.NoError(t, err, "response should still be valid JSON with empty accounts")

	// Verify robustness - should still provide complete version info
	require.Contains(t, jsonResponse, "cloudmcp", "should still contain cloudmcp object")
	require.Equal(t, "linode", jsonResponse["service"], "service should still be linode")

	// Verify cloudmcp version object is still valid
	cloudmcpObj, ok := jsonResponse["cloudmcp"].(map[string]interface{})
	require.True(t, ok, "cloudmcp should still be an object")
	require.Equal(t, "1.0.0", cloudmcpObj["version"], "should still contain correct version")
	// Note: current_account will show as "unknown" when no accounts are configured
}

// TestGetCurrentAccountName tests the getCurrentAccountName helper function for various account manager states.
// This test verifies the helper function's behavior in both success and error scenarios.
//
// **Test Workflow**:
// 1. **Success Case**: Test with valid account information
// 2. **Error Case**: Test with account manager error
// 3. **Format Verification**: Ensure proper account name formatting
//
// **Expected Behavior**:
// • Returns formatted "name (label)" for successful account retrieval
// • Returns "unknown" when account manager encounters error
// • Maintains consistent behavior for error handling
//
// **Purpose**: This test ensures the helper function provides reliable account name information for version display.
func TestGetCurrentAccountName(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Create test service with valid accoun
		service, mockAccountManager, _ := CreateTestService()

		accountName := service.getCurrentAccountName()

		require.Equal(t, "test-account (Test Account)", accountName,
			"should format account name with label")

		// Verify account manager was used correctly
		account, err := mockAccountManager.GetCurrentAccount()
		require.NoError(t, err, "should be able to get current account")
		require.Equal(t, "test-account", account.Name, "should have correct account name")
	})

	t.Run("Error", func(t *testing.T) {
		// Create minimal service with empty account manager
		log := logger.New("debug")
		cfg := &config.Config{
			DefaultLinodeAccount: "nonexistent",
			LinodeAccounts:       map[string]config.LinodeAccount{},
		}

		service := &Service{
			config: cfg,
			logger: log,
			accountManager: &AccountManager{
				accounts:       make(map[string]*Account),
				currentAccount: "",
				mu:             sync.RWMutex{},
			},
		}

		accountName := service.getCurrentAccountName()

		require.Equal(t, "unknown", accountName,
			"should return 'unknown' when account manager has no accounts")
	})
}

// TestVersionInfoConsistency tests that version information is consistent between text and JSON formats.
// This test ensures both version commands return the same underlying data in different formats.
//
// **Test Workflow**:
// 1. **Dual Execution**: Call both version functions with same service instance
// 2. **JSON Parsing**: Extract version data from JSON response
// 3. **Text Parsing**: Extract version data from text response
// 4. **Consistency Check**: Verify all version fields match between formats
//
// **Expected Behavior**:
// • Same version numbers in both text and JSON responses
// • Identical feature flags and build information
// • Consistent account name display
// • No data discrepancies between output formats
//
// **Purpose**: This test ensures version information consistency across different output formats for reliable diagnostics.
func TestVersionInfoConsistency(t *testing.T) {
	// Create test service
	service, _, _ := CreateTestService()
	ctx := context.Background()

	// Get text version
	textRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "cloudmcp_version",
			Arguments: map[string]interface{}{},
		},
	}
	textResult, err := service.handleSystemVersion(ctx, textRequest)
	require.NoError(t, err, "text version should succeed")

	// Get JSON version
	jsonRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "cloudmcp_version_json",
			Arguments: map[string]interface{}{},
		},
	}
	jsonResult, err := service.handleSystemVersionJSON(ctx, jsonRequest)
	require.NoError(t, err, "JSON version should succeed")

	// Parse JSON to compare with tex
	jsonText := getTextContent(t, jsonResult)
	var jsonResponse map[string]interface{}
	err = json.Unmarshal([]byte(jsonText), &jsonResponse)
	require.NoError(t, err, "JSON should be valid")

	cloudmcpObj := jsonResponse["cloudmcp"].(map[string]interface{})

	// Verify version consistency
	textContent := getTextContent(t, textResult)
	require.Contains(t, textContent, cloudmcpObj["version"].(string),
		"text should contain same version as JSON")
	require.Contains(t, textContent, cloudmcpObj["api_version"].(string),
		"text should contain same API version as JSON")

	// Verify account consistency
	expectedAccount := jsonResponse["current_account"].(string)
	require.Contains(t, textContent, expectedAccount,
		"text should contain same account as JSON")

	// Verify feature consistency
	features := cloudmcpObj["features"].(map[string]interface{})
	require.Contains(t, textContent, features["linode_api_coverage"].(string),
		"text should contain same API coverage as JSON")
}

// TestVersionConstants tests that the version package provides expected constants.
// This test verifies that version information is properly configured.
//
// **Test Workflow**:
// 1. **Version Retrieval**: Get version info from version package
// 2. **Constant Verification**: Check that required constants are se
// 3. **Feature Validation**: Ensure all expected features are presen
//
// **Expected Behavior**:
// • Version is set to "1.0.0"
// • API version is set to "4.0"
// • All required features are configured
// • No empty or missing critical fields
//
// **Purpose**: This test ensures version constants are properly configured for accurate reporting.
func TestVersionConstants(t *testing.T) {
	versionInfo := version.Get()

	// Verify core version information
	require.Equal(t, "1.0.0", versionInfo.Version, "should have correct version")
	require.Equal(t, "4.0", versionInfo.APIVersion, "should have correct API version")
	require.NotEmpty(t, versionInfo.GoVersion, "should have Go version")
	require.NotEmpty(t, versionInfo.Platform, "should have platform information")

	// Verify required features
	require.Contains(t, versionInfo.Features, "linode_api_coverage", "should have API coverage feature")
	require.Contains(t, versionInfo.Features, "multi_account", "should have multi-account feature")
	require.Contains(t, versionInfo.Features, "metrics", "should have metrics feature")
	require.Contains(t, versionInfo.Features, "logging", "should have logging feature")
	require.Contains(t, versionInfo.Features, "protocol", "should have protocol feature")

	// Verify feature values
	require.Equal(t, "100%", versionInfo.Features["linode_api_coverage"], "should show 100% API coverage")
	require.Equal(t, "enabled", versionInfo.Features["multi_account"], "should show multi-account enabled")
	require.Equal(t, "prometheus", versionInfo.Features["metrics"], "should show prometheus metrics")
	require.Equal(t, "structured", versionInfo.Features["logging"], "should show structured logging")
	require.Equal(t, "mcp", versionInfo.Features["protocol"], "should show MCP protocol")
}
