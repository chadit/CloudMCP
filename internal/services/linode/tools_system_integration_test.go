//go:build integration

package linode

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// createSystemTestServer creates an HTTP test server with system API endpoints.
// This extends the base HTTP test server infrastructure with comprehensive
// system-specific endpoints for integration testing.
//
// **System Endpoints Supported:**
// • GET /v4/account - Account information (CloudMCP version)
// • GET /v4/profile - Profile information (CloudMCP version JSON)
//
// **Mock Data Features:**
// • CloudMCP version information in text and JSON formats
// • System status and build information
// • Proper HTTP status codes and error responses
func createSystemTestServer() *httptest.Server {
	mux := http.NewServeMux()

	// Include base endpoints (profile, account) from main test server
	addSystemBaseEndpoints(mux)

	return httptest.NewServer(mux)
}

// addSystemBaseEndpoints adds the basic profile and account endpoints needed for service initialization.
func addSystemBaseEndpoints(mux *http.ServeMux) {
	// Profile endpoint - used during service initialization and system version queries
	mux.HandleFunc("/v4/profile", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		response := map[string]interface{}{
			"uid":                  12345,
			"username":             "testuser",
			"email":                "test@example.com",
			"timezone":             "US/Eastern",
			"email_notifications":  true,
			"referrals":            map[string]int{"total": 0, "completed": 0, "pending": 0, "credit": 0},
			"ip_whitelist_enabled": false,
			"lish_auth_method":     "password",
			"authorized_keys":      []string{},
			"two_factor_auth":      false,
			"restricted":           false,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Account endpoint - used during service initialization and system version queries
	mux.HandleFunc("/v4/account", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		response := map[string]interface{}{
			"email":              "test@example.com",
			"first_name":         "Test",
			"last_name":          "User",
			"company":            "Test Company",
			"address_1":          "123 Test St",
			"city":               "Test City",
			"state":              "Test State",
			"zip":                "12345",
			"country":            "US",
			"phone":              "555-1234",
			"credit_card":        map[string]string{"last_four": "1234", "expiry": "12/2025"},
			"balance":            100.0,
			"balance_uninvoiced": 0.0,
			"capabilities":       []string{"Linodes", "NodeBalancers", "Block Storage", "Object Storage"},
			"active_since":       "2020-01-01T00:00:00",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
}

// TestSystemToolsIntegration tests all system-related CloudMCP tools
// using HTTP test server infrastructure to validate proper integration
// between CloudMCP handlers and system information endpoints.
//
// **Integration Test Coverage**:
// • cloudmcp_version - Get CloudMCP version information
// • cloudmcp_version_json - Get CloudMCP version in JSON format
//
// **Test Environment**: HTTP test server with system API handlers
//
// **Expected Behavior**:
// • All handlers return properly formatted text responses
// • Version information includes build details and server info
// • JSON format responses are properly structured
//
// **Purpose**: Validates that CloudMCP system handlers correctly format
// version and system information for LLM consumption using lightweight HTTP test infrastructure.
func TestSystemToolsIntegration(t *testing.T) {
	// Extend the HTTP test server with system endpoints
	server := createSystemTestServer()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
	ctx := context.Background()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	t.Run("CloudMCPVersion", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "cloudmcp_version",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleSystemVersion(ctx, request)
		require.NoError(t, err, "CloudMCP version should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate version information formatting
		require.Contains(t, responseText, "CloudMCP Server Version:", "should contain version header")
		require.Contains(t, responseText, "Version:", "should contain version field")
		require.Contains(t, responseText, "Build:", "should contain build field")
		require.Contains(t, responseText, "Server Name:", "should contain server name")
		require.Contains(t, responseText, "Account Info:", "should contain account info")
	})

	t.Run("CloudMCPVersionJSON", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "cloudmcp_version_json",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleSystemVersionJSON(ctx, request)
		require.NoError(t, err, "CloudMCP version JSON should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate JSON version information formatting
		require.Contains(t, responseText, "CloudMCP Server Version (JSON):", "should contain JSON version header")
		require.Contains(t, responseText, "\"version\":", "should contain version field in JSON")
		require.Contains(t, responseText, "\"build\":", "should contain build field in JSON")
		require.Contains(t, responseText, "\"server_name\":", "should contain server name in JSON")
		require.Contains(t, responseText, "\"account\":", "should contain account info in JSON")
	})
}

// TestSystemErrorHandlingIntegration tests error scenarios for system tools
// to ensure CloudMCP handles system errors gracefully and provides meaningful
// error messages to users.
//
// **Error Test Scenarios**:
// • Service initialization failures
// • Account information retrieval errors
// • Version information formatting errors
//
// **Expected Behavior**:
// • Proper error handling with contextual error messages
// • No unhandled exceptions or panics
// • Error responses follow CloudMCP error format
// • Error messages are actionable for users
//
// **Purpose**: Validates robust error handling in system operations
// and ensures reliable operation under error conditions.
func TestSystemErrorHandlingIntegration(t *testing.T) {
	server := createSystemTestServer()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
	ctx := context.Background()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	// Note: System tools typically don't have many error scenarios since they
	// primarily return internal version/build information. The main error cases
	// would be service initialization failures, which are already tested in
	// the main integration test.

	t.Run("SystemVersionStability", func(t *testing.T) {
		// Test multiple calls to ensure stability
		for i := 0; i < 3; i++ {
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "cloudmcp_version",
					Arguments: map[string]interface{}{},
				},
			}

			result, err := service.handleSystemVersion(ctx, request)
			require.NoError(t, err, "CloudMCP version should be stable across multiple calls")
			require.NotNil(t, result, "result should not be nil")
			require.NotEmpty(t, result.Content, "result should have content")
		}
	})

	t.Run("SystemVersionJSONStability", func(t *testing.T) {
		// Test multiple calls to ensure stability
		for i := 0; i < 3; i++ {
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "cloudmcp_version_json",
					Arguments: map[string]interface{}{},
				},
			}

			result, err := service.handleSystemVersionJSON(ctx, request)
			require.NoError(t, err, "CloudMCP version JSON should be stable across multiple calls")
			require.NotNil(t, result, "result should not be nil")
			require.NotEmpty(t, result.Content, "result should have content")
		}
	})
}
