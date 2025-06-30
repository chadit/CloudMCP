package linode_test

import (
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/internal/services/linode"
	"github.com/chadit/CloudMCP/internal/version"
	"github.com/chadit/CloudMCP/pkg/logger"
)

// Helper function to extract text content from CallToolResult.
func getTextContent(t *testing.T, result *mcp.CallToolResult) string {
	require.NotNil(t, result, "result should not be nil")
	require.NotEmpty(t, result.Content, "result should have content")
	require.Equal(t, 1, len(result.Content), "result should have exactly one content item")

	textContent, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok, "result content should be text content")
	require.NotEmpty(t, textContent.Text, "result text should not be empty")

	return textContent.Text
}

// TestSystemVersionTool tests the cloudmcp_version tool through the exported MCP interface.
// This test verifies that version information is returned in human-readable format through proper API boundaries.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with mock account manager
// 2. **Tool Execution**: Call cloudmcp_version tool through exported CallToolForTesting method
// 3. **Response Validation**: Verify version information format and content
// 4. **Content Verification**: Check for presence of all required version fields
// 5. **Account Information**: Ensure current account name is included
//
// **Test Environment**: Mock account manager with test account "test-account (Test Account)"
//
// **Expected Behavior**:
// • Returns successful tool result with formatted text output
// • Includes CloudMCP version, API version, build information
// • Contains Git commit, branch, Go version, and platform details
// • Lists all feature flags (API coverage, multi-account, metrics, etc.)
// • Shows current account name from account manager
//
// **Purpose**: This test ensures the version tool provides complete diagnostic information through exported API.
func TestSystemVersionTool(t *testing.T) {
	t.Parallel()

	// Create isolated test service
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "test-account",
		LinodeAccounts: map[string]config.LinodeAccount{
			"test-account": {
				Label: "Test Account",
				Token: "test-token",
			},
		},
	}

	// Create account manager for testing
	accountManager := linode.NewAccountManagerForTesting()
	testAccount := linode.NewAccountForTesting("test-account", "Test Account")
	accountManager.AddAccountForTesting(testAccount)
	accountManager.SetCurrentAccountForTesting("test-account")

	// Create service with proper testing constructor
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test system version request
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "cloudmcp_version",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "cloudmcp_version tool should not return an error")

	// Verify version information content
	versionText := getTextContent(t, result)
	require.Contains(t, versionText, "CloudMCP Version Information", "should contain version header")
	require.Contains(t, versionText, "Version: 1.0.0", "should contain CloudMCP version")
	require.Contains(t, versionText, "API Version: 4.0", "should contain API version")
	require.Contains(t, versionText, "Linode API Coverage: 100%", "should show 100% API coverage")
	require.Contains(t, versionText, "Multi-Account Support: enabled", "should show multi-account support")
	require.Contains(t, versionText, "Current Account: test-account (Test Account)", "should show current account")

	// Verify all required sections are present
	require.Contains(t, versionText, "Git Commit:", "should contain git commit")
	require.Contains(t, versionText, "Go Version:", "should contain Go version")
	require.Contains(t, versionText, "Platform:", "should contain platform info")
	require.Contains(t, versionText, "Features:", "should contain features section")

	// Verify account manager was used
	account, err := accountManager.GetCurrentAccount()
	require.NoError(t, err, "should be able to get current account")
	require.Equal(t, "test-account", account.Name, "should have correct account name")
}

// TestSystemVersionJSONTool tests the cloudmcp_version_json tool through the exported MCP interface.
// This test verifies that version information is returned in structured JSON format through proper API boundaries.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with mock account manager
// 2. **Tool Execution**: Call cloudmcp_version_json tool through exported CallToolForTesting method
// 3. **JSON Parsing**: Unmarshal response to verify valid JSON structure
// 4. **Data Validation**: Check all required fields and nested objects
// 5. **Account Verification**: Ensure current account information is included
//
// **Test Environment**: Mock account manager with test account "test-account (Test Account)"
//
// **Expected Behavior**:
// • Returns successful tool result with valid JSON output
// • Contains structured "cloudmcp" object with all version fields
// • Includes "current_account" field with account name
// • Specifies "service" field as "linode"
// • JSON structure matches expected schema for programmatic consumption
//
// **Purpose**: This test ensures JSON version tool provides machine-readable data through exported API.
func TestSystemVersionJSONTool(t *testing.T) {
	t.Parallel()

	// Create isolated test service
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "test-account",
		LinodeAccounts: map[string]config.LinodeAccount{
			"test-account": {
				Label: "Test Account",
				Token: "test-token",
			},
		},
	}

	// Create account manager for testing
	accountManager := linode.NewAccountManagerForTesting()
	testAccount := linode.NewAccountForTesting("test-account", "Test Account")
	accountManager.AddAccountForTesting(testAccount)
	accountManager.SetCurrentAccountForTesting("test-account")

	// Create service with proper testing constructor
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test system version JSON request
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "cloudmcp_version_json",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "cloudmcp_version_json tool should not return an error")

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

	// Verify cloudmcp version object
	cloudmcpObj, isObject := jsonResponse["cloudmcp"].(map[string]interface{})
	require.True(t, isObject, "cloudmcp should be an object")

	// Verify version fields
	require.Equal(t, "1.0.0", cloudmcpObj["version"], "should contain correct version")
	require.Equal(t, "4.0", cloudmcpObj["api_version"], "should contain correct API version")
	require.Contains(t, cloudmcpObj, "build_date", "should contain build_date")
	require.Contains(t, cloudmcpObj, "git_commit", "should contain git_commit")
	require.Contains(t, cloudmcpObj, "git_branch", "should contain git_branch")
	require.Contains(t, cloudmcpObj, "go_version", "should contain go_version")
	require.Contains(t, cloudmcpObj, "platform", "should contain platform")

	// Verify features object
	features, ok := cloudmcpObj["features"].(map[string]interface{})
	require.True(t, ok, "features should be an object")
	require.Equal(t, "100%", features["linode_api_coverage"], "should show 100% API coverage")
	require.Equal(t, "enabled", features["multi_account"], "should show multi-account enabled")
	require.Equal(t, "prometheus", features["metrics"], "should show prometheus metrics")
	require.Equal(t, "structured", features["logging"], "should show structured logging")
	require.Equal(t, "mcp", features["protocol"], "should show MCP protocol")

	// Verify account manager was used
	account, err := accountManager.GetCurrentAccount()
	require.NoError(t, err, "should be able to get current account")
	require.Equal(t, "test-account", account.Name, "should have correct account name")
}

// TestSystemVersionTool_AccountError tests the cloudmcp_version tool robustness through exported API.
// This test verifies the tool continues to work even with account lookup issues.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Tool Execution**: Call cloudmcp_version tool with no configured accounts
// 3. **Response Validation**: Verify tool provides version info despite account issues
//
// **Expected Behavior**:
// • Returns successful tool result regardless of account state
// • Provides all version information correctly
// • Handles account lookup gracefully with fallback
//
// **Purpose**: This test ensures version tool robustness through exported API.
func TestSystemVersionTool_AccountError(t *testing.T) {
	t.Parallel()

	// Create minimal service with empty account manager
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	// Create empty account manager for error testing
	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test system version request with empty account manager
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "cloudmcp_version",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "cloudmcp_version tool should not return an error even with empty accounts")

	// Verify robustness - should still provide version info
	versionText := getTextContent(t, result)
	require.Contains(t, versionText, "CloudMCP Version Information",
		"should still contain version header")
	require.Contains(t, versionText, "Version: 1.0.0",
		"should still contain CloudMCP version")

	// Note: Account will show as "unknown" when no accounts are configured
}

// TestVersionInfoConsistencyThroughAPI tests that version information is consistent between text and JSON formats through exported API.
// This test ensures both version tools return the same underlying data in different formats.
//
// **Test Workflow**:
// 1. **Dual Execution**: Call both version tools with same service instance
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
// **Purpose**: This test ensures version information consistency across different output formats through exported API.
func TestVersionInfoConsistencyThroughAPI(t *testing.T) {
	t.Parallel()

	// Create isolated test service
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "test-account",
		LinodeAccounts: map[string]config.LinodeAccount{
			"test-account": {
				Label: "Test Account",
				Token: "test-token",
			},
		},
	}

	// Create account manager for testing
	accountManager := linode.NewAccountManagerForTesting()
	testAccount := linode.NewAccountForTesting("test-account", "Test Account")
	accountManager.AddAccountForTesting(testAccount)
	accountManager.SetCurrentAccountForTesting("test-account")

	// Create service with proper testing constructor
	service := linode.NewForTesting(cfg, log, accountManager)
	ctx := t.Context()

	// Get text version
	textRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "cloudmcp_version",
			Arguments: map[string]interface{}{},
		},
	}
	textResult, err := service.CallToolForTesting(ctx, textRequest)
	require.NoError(t, err, "text version should succeed")

	// Get JSON version
	jsonRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "cloudmcp_version_json",
			Arguments: map[string]interface{}{},
		},
	}
	jsonResult, err := service.CallToolForTesting(ctx, jsonRequest)
	require.NoError(t, err, "JSON version should succeed")

	// Parse JSON to compare with text
	jsonText := getTextContent(t, jsonResult)

	var jsonResponse map[string]interface{}
	err = json.Unmarshal([]byte(jsonText), &jsonResponse)
	require.NoError(t, err, "JSON should be valid")

	cloudmcpObj, isMap := jsonResponse["cloudmcp"].(map[string]interface{})
	require.True(t, isMap, "cloudmcp field should be a map")

	// Verify version consistency
	textContent := getTextContent(t, textResult)
	version, isVersionString := cloudmcpObj["version"].(string)
	require.True(t, isVersionString, "version should be a string")
	require.Contains(t, textContent, version,
		"text should contain same version as JSON")

	apiVersion, isAPIVersionString := cloudmcpObj["api_version"].(string)
	require.True(t, isAPIVersionString, "api_version should be a string")
	require.Contains(t, textContent, apiVersion,
		"text should contain same API version as JSON")

	// Verify account consistency
	expectedAccount, isAccountString := jsonResponse["current_account"].(string)
	require.True(t, isAccountString, "current_account should be a string")
	require.Contains(t, textContent, expectedAccount,
		"text should contain same account as JSON")

	// Verify feature consistency
	features, isFeaturesMap := cloudmcpObj["features"].(map[string]interface{})
	require.True(t, isFeaturesMap, "features should be a map")

	linodeCoverage, isCoverageString := features["linode_api_coverage"].(string)
	require.True(t, isCoverageString, "linode_api_coverage should be a string")
	require.Contains(t, textContent, linodeCoverage,
		"text should contain same API coverage as JSON")
}

// TestVersionConstants tests that the version package provides expected constants through exported API.
// This test verifies that version information is properly configured.
//
// **Test Workflow**:
// 1. **Version Retrieval**: Get version info from version package
// 2. **Constant Verification**: Check that required constants are set
// 3. **Feature Validation**: Ensure all expected features are present
//
// **Expected Behavior**:
// • Version is set to "1.0.0"
// • API version is set to "4.0"
// • All required features are configured
// • No empty or missing critical fields
//
// **Purpose**: This test ensures version constants are properly configured for accurate reporting.
func TestVersionConstants(t *testing.T) {
	t.Parallel()

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
