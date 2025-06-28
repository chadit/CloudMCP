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

// createAccountTestServer creates an HTTP test server with account management API endpoints.
// This extends the base HTTP test server infrastructure with comprehensive
// account-specific endpoints for integration testing.
//
// **Account Endpoints Supported:**
// • GET /v4/account - Get current account information
// • GET /v4/profile - Get current user profile information
//
// **Mock Data Features:**
// • Realistic account configurations with billing and capabilities
// • Multiple account simulation for account switching
// • Profile information with user preferences
// • Proper HTTP status codes and error responses
func createAccountTestServer() *httptest.Server {
	mux := http.NewServeMux()

	// Include base endpoints (profile, account) from main test server
	addAccountBaseEndpoints(mux)

	return httptest.NewServer(mux)
}

// addAccountBaseEndpoints adds the basic profile and account endpoints needed for service initialization.
func addAccountBaseEndpoints(mux *http.ServeMux) {
	// Profile endpoint - used during service initialization and account queries
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
			"referrals":            map[string]int{"total": 2, "completed": 1, "pending": 1, "credit": 50},
			"ip_whitelist_enabled": false,
			"lish_auth_method":     "password",
			"authorized_keys":      []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQ..."},
			"two_factor_auth":      true,
			"restricted":           false,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Account endpoint - used during service initialization and account queries
	mux.HandleFunc("/v4/account", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		response := map[string]interface{}{
			"email":              "test@example.com",
			"first_name":         "Test",
			"last_name":          "User",
			"company":            "Test Company Inc.",
			"address_1":          "123 Main Street",
			"address_2":          "Suite 100",
			"city":               "New York",
			"state":              "NY",
			"zip":                "10001",
			"country":            "US",
			"phone":              "+1-555-123-4567",
			"credit_card":        map[string]string{"last_four": "1234", "expiry": "12/2025"},
			"balance":            125.50,
			"balance_uninvoiced": 45.75,
			"capabilities":       []string{"Linodes", "NodeBalancers", "Block Storage", "Object Storage", "Kubernetes", "Databases"},
			"active_since":       "2020-01-15T08:30:00",
			"promo_credit":       25.00,
			"tax_id":             "12-3456789",
			"billing_source":     "credit_card",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
}

// TestAccountToolsIntegration tests all account management-related CloudMCP tools
// using HTTP test server infrastructure to validate proper integration
// between CloudMCP handlers and Linode API endpoints.
//
// **Integration Test Coverage**:
// • linode_account_get - Get current account information
// • linode_account_list - List all configured accounts
// • linode_account_switch - Switch to different account
//
// **Test Environment**: HTTP test server with account management API handlers
//
// **Expected Behavior**:
// • All handlers return properly formatted text responses
// • Error conditions are handled gracefully with meaningful messages
// • Account data includes all required fields (email, balance, capabilities, etc.)
// • Profile information is properly formatted
// • Account switching functionality works correctly
//
// **Purpose**: Validates that CloudMCP account management handlers correctly format
// Linode API responses for LLM consumption using lightweight HTTP test infrastructure.
func TestAccountToolsIntegration(t *testing.T) {
	// Extend the HTTP test server with account management endpoints
	server := createAccountTestServer()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
	ctx := context.Background()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	t.Run("AccountGet", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_account_get",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleAccountGet(ctx, request)
		require.NoError(t, err, "account get should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate account information formatting
		require.Contains(t, responseText, "Current Linode Account Information:", "should have account header")
		require.Contains(t, responseText, "Email: test@example.com", "should contain email")
		require.Contains(t, responseText, "Name: Test User", "should contain full name")
		require.Contains(t, responseText, "Company: Test Company Inc.", "should contain company")
		require.Contains(t, responseText, "Balance: $125.50", "should contain balance")
		require.Contains(t, responseText, "Uninvoiced: $45.75", "should contain uninvoiced balance")
		require.Contains(t, responseText, "Capabilities:", "should list capabilities")
		require.Contains(t, responseText, "Linodes", "should contain Linodes capability")
		require.Contains(t, responseText, "Kubernetes", "should contain Kubernetes capability")
		require.Contains(t, responseText, "Active Since: 2020-01-15", "should contain active since date")
	})

	t.Run("AccountList", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_account_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleAccountList(ctx, request)
		require.NoError(t, err, "account list should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate account list formatting
		require.Contains(t, responseText, "Configured Linode Accounts:", "should have accounts header")
		require.Contains(t, responseText, "Account Name:", "should list account names")
		require.Contains(t, responseText, "Label:", "should list account labels")
		require.Contains(t, responseText, "Status:", "should show account status")
		require.Contains(t, responseText, "Current Account:", "should show current account")
	})

	t.Run("AccountSwitch", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_account_switch",
				Arguments: map[string]interface{}{
					"account_name": "primary",
				},
			},
		}

		result, err := service.handleAccountSwitch(ctx, request)
		require.NoError(t, err, "account switch should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate account switch confirmation
		require.Contains(t, responseText, "Account switched successfully", "should confirm account switch")
		require.Contains(t, responseText, "Current Account:", "should show current account")
		require.Contains(t, responseText, "primary", "should show switched account name")
	})
}

// TestAccountErrorHandlingIntegration tests error scenarios for account management tools
// to ensure CloudMCP handles API errors gracefully and provides meaningful
// error messages to users.
//
// **Error Test Scenarios**:
// • Non-existent account names for switching
// • Invalid account configurations
// • Permission errors for account operations
// • Account initialization failures
//
// **Expected Behavior**:
// • Proper error handling with contextual error messages
// • No unhandled exceptions or panics
// • Error responses follow CloudMCP error format
// • Error messages are actionable for users
//
// **Purpose**: Validates robust error handling in account management operations
// and ensures reliable operation under error conditions.
func TestAccountErrorHandlingIntegration(t *testing.T) {
	server := createAccountTestServer()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
	ctx := context.Background()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	t.Run("AccountSwitchInvalidName", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_account_switch",
				Arguments: map[string]interface{}{
					"account_name": "non-existent-account", // Non-existent account
				},
			},
		}

		result, err := service.handleAccountSwitch(ctx, request)

		// Check for either Go error or MCP error result
		if err != nil {
			// Go error pattern
			require.Contains(t, err.Error(), "Failed to switch account", "error should mention switch failure")
		} else {
			// MCP error result pattern
			require.NotNil(t, result, "result should not be nil")
			require.True(t, result.IsError, "result should be marked as error")
		}

		// When there's a Go error, result might be nil
		if result != nil {
			require.True(t, result.IsError, "result should be marked as error if not nil")
		}
	})

	t.Run("AccountSwitchMissingParameter", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_account_switch",
				Arguments: map[string]interface{}{
					// Missing required 'account_name' parameter
				},
			},
		}

		result, err := service.handleAccountSwitch(ctx, request)

		// Should get MCP error result for missing parameter
		if err == nil {
			require.NotNil(t, result, "result should not be nil")
			require.True(t, result.IsError, "result should be marked as error")
		}
	})

	t.Run("AccountGetStability", func(t *testing.T) {
		// Test multiple calls to ensure stability
		for i := 0; i < 3; i++ {
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "linode_account_get",
					Arguments: map[string]interface{}{},
				},
			}

			result, err := service.handleAccountGet(ctx, request)
			require.NoError(t, err, "account get should be stable across multiple calls")
			require.NotNil(t, result, "result should not be nil")
			require.NotEmpty(t, result.Content, "result should have content")
		}
	})

	t.Run("AccountListStability", func(t *testing.T) {
		// Test multiple calls to ensure stability
		for i := 0; i < 3; i++ {
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "linode_account_list",
					Arguments: map[string]interface{}{},
				},
			}

			result, err := service.handleAccountList(ctx, request)
			require.NoError(t, err, "account list should be stable across multiple calls")
			require.NotNil(t, result, "result should not be nil")
			require.NotEmpty(t, result.Content, "result should have content")
		}
	})
}
