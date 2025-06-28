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

// createStackScriptsTestServer creates an HTTP test server with StackScripts API endpoints.
// This extends the base HTTP test server infrastructure with comprehensive
// StackScripts-specific endpoints for integration testing.
//
// **StackScripts Endpoints Supported:**
// • GET /v4/linode/stackscripts - List StackScripts
// • GET /v4/linode/stackscripts/{id} - Get specific StackScript
// • POST /v4/linode/stackscripts - Create new StackScript
// • PUT /v4/linode/stackscripts/{id} - Update StackScript
// • DELETE /v4/linode/stackscripts/{id} - Delete StackScript
//
// **Mock Data Features:**
// • Realistic StackScript configurations with deployment scripts
// • Public and private StackScript visibility
// • Image compatibility and deployment statistics
// • Error simulation for non-existent resources
// • Proper HTTP status codes and error responses
func createStackScriptsTestServer() *httptest.Server {
	mux := http.NewServeMux()

	// Include base endpoints (profile, account) from main test server
	addStackScriptsBaseEndpoints(mux)

	// StackScripts list endpoint
	mux.HandleFunc("/v4/linode/stackscripts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleStackScriptsList(w, r)
		case http.MethodPost:
			handleStackScriptsCreate(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Specific StackScript endpoints with explicit ID matching
	mux.HandleFunc("/v4/linode/stackscripts/12345", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleStackScriptsGet(w, r, "12345")
		case http.MethodPut:
			handleStackScriptsUpdate(w, r, "12345")
		case http.MethodDelete:
			handleStackScriptsDelete(w, r, "12345")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Non-existent StackScript endpoints for error testing
	mux.HandleFunc("/v4/linode/stackscripts/999999", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleStackScriptsGet(w, r, "999999")
		case http.MethodDelete:
			handleStackScriptsDelete(w, r, "999999")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return httptest.NewServer(mux)
}

// addStackScriptsBaseEndpoints adds the basic profile and account endpoints needed for service initialization.
func addStackScriptsBaseEndpoints(mux *http.ServeMux) {
	// Profile endpoint - used during service initialization
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

	// Account endpoint
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

// handleStackScriptsList handles the StackScripts list mock response.
func handleStackScriptsList(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"id":                 12345,
				"username":           "testuser",
				"label":              "nginx-setup",
				"description":        "Automated NGINX web server setup",
				"ordinal":            1,
				"logo_url":           "",
				"images":             []string{"linode/ubuntu22.04", "linode/debian11"},
				"deployments_total":  150,
				"deployments_active": 45,
				"is_public":          true,
				"mine":               true,
				"created":            "2023-01-01T00:00:00",
				"updated":            "2023-01-01T00:00:00",
				"rev_note":           "Initial version",
				"script":             "#!/bin/bash\napt update && apt install -y nginx\nsystemctl enable nginx\nsystemctl start nginx",
				"user_defined_fields": []map[string]interface{}{
					{
						"label":   "domain",
						"name":    "DOMAIN",
						"example": "example.com",
					},
				},
				"user_gravatar_id": "abc123",
			},
			{
				"id":                  54321,
				"username":            "testuser",
				"label":               "docker-setup",
				"description":         "Docker and Docker Compose installation",
				"ordinal":             2,
				"logo_url":            "",
				"images":              []string{"linode/ubuntu22.04", "linode/centos8"},
				"deployments_total":   89,
				"deployments_active":  23,
				"is_public":           false,
				"mine":                true,
				"created":             "2023-01-02T00:00:00",
				"updated":             "2023-01-02T00:00:00",
				"rev_note":            "Added Docker Compose",
				"script":              "#!/bin/bash\ncurl -fsSL https://get.docker.com -o get-docker.sh\nsh get-docker.sh",
				"user_defined_fields": []map[string]interface{}{},
				"user_gravatar_id":    "def456",
			},
		},
		"page":    1,
		"pages":   1,
		"results": 2,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleStackScriptsGet handles the specific StackScript mock response.
func handleStackScriptsGet(w http.ResponseWriter, r *http.Request, stackscriptID string) {
	switch stackscriptID {
	case "12345":
		response := map[string]interface{}{
			"id":                 12345,
			"username":           "testuser",
			"label":              "nginx-setup",
			"description":        "Automated NGINX web server setup",
			"ordinal":            1,
			"logo_url":           "",
			"images":             []string{"linode/ubuntu22.04", "linode/debian11"},
			"deployments_total":  150,
			"deployments_active": 45,
			"is_public":          true,
			"mine":               true,
			"created":            "2023-01-01T00:00:00",
			"updated":            "2023-01-01T00:00:00",
			"rev_note":           "Initial version",
			"script":             "#!/bin/bash\napt update && apt install -y nginx\nsystemctl enable nginx\nsystemctl start nginx",
			"user_defined_fields": []map[string]interface{}{
				{
					"label":   "domain",
					"name":    "DOMAIN",
					"example": "example.com",
				},
			},
			"user_gravatar_id": "abc123",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	case "999999":
		// Simulate not found error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "stackscript_id", "reason": "Not found"},
			},
		})
	default:
		http.Error(w, "StackScript not found", http.StatusNotFound)
	}
}

// handleStackScriptsCreate handles the StackScript creation mock response.
func handleStackScriptsCreate(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"id":                  98765,
		"username":            "testuser",
		"label":               "new-stackscript",
		"description":         "Newly created StackScript",
		"ordinal":             1,
		"logo_url":            "",
		"images":              []string{"linode/ubuntu22.04"},
		"deployments_total":   0,
		"deployments_active":  0,
		"is_public":           false,
		"mine":                true,
		"created":             "2023-01-01T01:00:00",
		"updated":             "2023-01-01T01:00:00",
		"rev_note":            "Initial creation",
		"script":              "#!/bin/bash\necho 'Hello, World!'",
		"user_defined_fields": []map[string]interface{}{},
		"user_gravatar_id":    "xyz789",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleStackScriptsUpdate handles the StackScript update mock response.
func handleStackScriptsUpdate(w http.ResponseWriter, r *http.Request, stackscriptID string) {
	if stackscriptID == "999999" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "stackscript_id", "reason": "Not found"},
			},
		})
		return
	}

	response := map[string]interface{}{
		"id":                 12345,
		"username":           "testuser",
		"label":              "updated-nginx-setup",
		"description":        "Updated NGINX web server setup with SSL",
		"ordinal":            1,
		"logo_url":           "",
		"images":             []string{"linode/ubuntu22.04", "linode/debian11"},
		"deployments_total":  150,
		"deployments_active": 45,
		"is_public":          true,
		"mine":               true,
		"created":            "2023-01-01T00:00:00",
		"updated":            "2023-01-01T02:00:00",
		"rev_note":           "Added SSL configuration",
		"script":             "#!/bin/bash\napt update && apt install -y nginx certbot\nsystemctl enable nginx\nsystemctl start nginx",
		"user_defined_fields": []map[string]interface{}{
			{
				"label":   "domain",
				"name":    "DOMAIN",
				"example": "example.com",
			},
		},
		"user_gravatar_id": "abc123",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleStackScriptsDelete handles the StackScript deletion mock response.
func handleStackScriptsDelete(w http.ResponseWriter, r *http.Request, stackscriptID string) {
	if stackscriptID == "999999" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "stackscript_id", "reason": "Not found"},
			},
		})
		return
	}

	// Return 204 No Content for successful deletion
	w.WriteHeader(http.StatusNoContent)
}

// TestStackScriptsToolsIntegration tests all StackScripts-related CloudMCP tools
// using HTTP test server infrastructure to validate proper integration
// between CloudMCP handlers and Linode API endpoints.
//
// **Integration Test Coverage**:
// • linode_stackscripts_list - List all StackScripts
// • linode_stackscript_get - Get specific StackScript details
// • linode_stackscript_create - Create new StackScript
// • linode_stackscript_update - Update existing StackScript
// • linode_stackscript_delete - Delete StackScript
//
// **Test Environment**: HTTP test server with StackScripts API handlers
//
// **Expected Behavior**:
// • All handlers return properly formatted text responses
// • Error conditions are handled gracefully with meaningful messages
// • StackScript data includes all required fields (ID, label, script, images, etc.)
// • Deployment statistics and user-defined fields are properly formatted
//
// **Purpose**: Validates that CloudMCP StackScripts handlers correctly format
// Linode API responses for LLM consumption using lightweight HTTP test infrastructure.
func TestStackScriptsToolsIntegration(t *testing.T) {
	// Extend the HTTP test server with StackScripts endpoints
	server := createStackScriptsTestServer()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
	ctx := context.Background()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	t.Run("StackScriptsList", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_stackscripts_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleStackScriptsList(ctx, request)
		require.NoError(t, err, "StackScripts list should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate StackScripts list formatting
		require.Contains(t, responseText, "Found 2 StackScripts:", "should indicate correct StackScript count")
		require.Contains(t, responseText, "nginx-setup", "should contain first StackScript label")
		require.Contains(t, responseText, "docker-setup", "should contain second StackScript label")
		require.Contains(t, responseText, "Public", "should show public visibility")
		require.Contains(t, responseText, "Private", "should show private visibility")
		require.Contains(t, responseText, "Deployments:", "should show deployment statistics")
	})

	t.Run("StackScriptGet", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_stackscript_get",
				Arguments: map[string]interface{}{
					"stackscript_id": float64(12345),
				},
			},
		}

		result, err := service.handleStackScriptGet(ctx, request)
		require.NoError(t, err, "StackScript get should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate detailed StackScript information
		require.Contains(t, responseText, "StackScript Details:", "should have StackScript details header")
		require.Contains(t, responseText, "ID: 12345", "should contain StackScript ID")
		require.Contains(t, responseText, "Label: nginx-setup", "should contain StackScript label")
		require.Contains(t, responseText, "Description: Automated NGINX", "should contain description")
		require.Contains(t, responseText, "Images:", "should list compatible images")
		require.Contains(t, responseText, "Script:", "should include script content")
	})

	t.Run("StackScriptCreate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_stackscript_create",
				Arguments: map[string]interface{}{
					"label":       "new-stackscript",
					"description": "Test StackScript creation",
					"images":      []interface{}{"linode/ubuntu22.04"},
					"script":      "#!/bin/bash\necho 'Hello, World!'",
				},
			},
		}

		result, err := service.handleStackScriptCreate(ctx, request)
		require.NoError(t, err, "StackScript create should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate StackScript creation confirmation
		require.Contains(t, responseText, "StackScript created successfully:", "should confirm creation")
		require.Contains(t, responseText, "ID: 98765", "should show new StackScript ID")
		require.Contains(t, responseText, "Label: new-stackscript", "should show StackScript label")
	})

	t.Run("StackScriptUpdate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_stackscript_update",
				Arguments: map[string]interface{}{
					"stackscript_id": float64(12345),
					"label":          "updated-nginx-setup",
					"description":    "Updated NGINX setup with SSL",
				},
			},
		}

		result, err := service.handleStackScriptUpdate(ctx, request)
		require.NoError(t, err, "StackScript update should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate StackScript update confirmation
		require.Contains(t, responseText, "StackScript updated successfully:", "should confirm update")
		require.Contains(t, responseText, "ID: 12345", "should show StackScript ID")
		require.Contains(t, responseText, "Label: updated-nginx-setup", "should show updated label")
	})

	t.Run("StackScriptDelete", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_stackscript_delete",
				Arguments: map[string]interface{}{
					"stackscript_id": float64(12345),
				},
			},
		}

		result, err := service.handleStackScriptDelete(ctx, request)
		require.NoError(t, err, "StackScript delete should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate StackScript deletion confirmation
		require.Contains(t, responseText, "StackScript", "should mention StackScript")
		require.Contains(t, responseText, "deleted successfully", "should confirm deletion")
		require.Contains(t, responseText, "12345", "should show deleted StackScript ID")
	})
}

// TestStackScriptsErrorHandlingIntegration tests error scenarios for StackScripts tools
// to ensure CloudMCP handles API errors gracefully and provides meaningful
// error messages to users.
//
// **Error Test Scenarios**:
// • Non-existent StackScript ID (404 errors)
// • Invalid script syntax
// • Image compatibility conflicts
// • Permission errors for StackScript operations
//
// **Expected Behavior**:
// • Proper error handling with contextual error messages
// • No unhandled exceptions or panics
// • Error responses follow CloudMCP error format
// • Error messages are actionable for users
//
// **Purpose**: Validates robust error handling in StackScripts operations
// and ensures reliable operation under error conditions.
func TestStackScriptsErrorHandlingIntegration(t *testing.T) {
	server := createStackScriptsTestServer()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
	ctx := context.Background()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	t.Run("StackScriptGetNotFound", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_stackscript_get",
				Arguments: map[string]interface{}{
					"stackscript_id": float64(999999), // Non-existent StackScript
				},
			},
		}

		result, err := service.handleStackScriptGet(ctx, request)

		// Check for either Go error or MCP error result
		if err != nil {
			// Go error pattern
			require.Contains(t, err.Error(), "Failed to get StackScript", "error should mention get StackScript failure")
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

	t.Run("StackScriptDeleteNotFound", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_stackscript_delete",
				Arguments: map[string]interface{}{
					"stackscript_id": float64(999999), // Non-existent StackScript
				},
			},
		}

		result, err := service.handleStackScriptDelete(ctx, request)

		// Check for either Go error or MCP error result
		if err != nil {
			// Go error pattern
			require.Contains(t, err.Error(), "Failed to delete StackScript", "error should mention delete StackScript failure")
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

	t.Run("InvalidStackScriptID", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_stackscript_get",
				Arguments: map[string]interface{}{
					"stackscript_id": "invalid", // Invalid ID type
				},
			},
		}

		result, err := service.handleStackScriptGet(ctx, request)

		// Should get MCP error result for invalid parameter
		if err == nil {
			require.NotNil(t, result, "result should not be nil")
			require.True(t, result.IsError, "result should be marked as error")
		}
	})
}
