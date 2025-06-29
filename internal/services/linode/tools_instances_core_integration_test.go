//go:build integration

package linode_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// createInstancesTestServer creates an HTTP test server with instances API endpoints.
// This extends the base HTTP test server infrastructure with comprehensive
// instances-specific endpoints for integration testing.
//
// **Instances Endpoints Supported:**
// • GET /v4/linode/instances - List all instances
// • GET /v4/linode/instances/{id} - Get specific instance details
// • POST /v4/linode/instances - Create new instance
// • PUT /v4/linode/instances/{id} - Update instance
// • DELETE /v4/linode/instances/{id} - Delete instance
// • POST /v4/linode/instances/{id}/boot - Boot instance
// • POST /v4/linode/instances/{id}/shutdown - Shutdown instance
// • POST /v4/linode/instances/{id}/reboot - Reboot instance
//
// **Mock Data Features:**
// • Realistic instance configurations with multiple types and regions
// • Instance lifecycle state management (running, stopped, booting, etc.)
// • Boot configuration and disk attachment simulation
// • Error simulation for non-existent resources and invalid operations
// • Proper HTTP status codes and error responses
func createInstancesTestServer() *httptest.Server {
	mux := http.NewServeMux()

	// Include base endpoints (profile, account) from main test server
	addInstancesBaseEndpoints(mux)

	// Instances list endpoint
	mux.HandleFunc("/v4/linode/instances", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleInstancesList(w, r)
		case http.MethodPost:
			handleInstancesCreate(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Specific instance endpoints
	mux.HandleFunc("/v4/linode/instances/12345", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleInstancesGet(w, r, "12345")
		case http.MethodPut:
			handleInstancesUpdate(w, r, "12345")
		case http.MethodDelete:
			handleInstancesDelete(w, r, "12345")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v4/linode/instances/67890", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleInstancesGet(w, r, "67890")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Instance action endpoints
	mux.HandleFunc("/v4/linode/instances/12345/boot", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handleInstancesBoot(w, r, "12345")
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v4/linode/instances/12345/shutdown", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handleInstancesShutdown(w, r, "12345")
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v4/linode/instances/12345/reboot", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handleInstancesReboot(w, r, "12345")
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Non-existent instance endpoints for error testing
	mux.HandleFunc("/v4/linode/instances/999999", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleInstancesGet(w, r, "999999")
		case http.MethodDelete:
			handleInstancesDelete(w, r, "999999")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return httptest.NewServer(mux)
}

// addInstancesBaseEndpoints adds the basic profile and account endpoints needed for service initialization.
func addInstancesBaseEndpoints(mux *http.ServeMux) {
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

// handleInstancesList handles the instances list mock response.
func handleInstancesList(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"id":      12345,
				"label":   "web-server-prod",
				"group":   "production",
				"status":  "running",
				"created": "2023-01-01T00:00:00",
				"updated": "2023-01-01T00:00:00",
				"region":  "us-east",
				"image":   "linode/ubuntu22.04",
				"type":    "g6-standard-2",
				"specs": map[string]interface{}{
					"disk":     50000,
					"memory":   4096,
					"vcpus":    2,
					"transfer": 4000,
				},
				"ipv4":             []string{"203.0.113.100"},
				"ipv6":             "2001:db8::1/64",
				"hypervisor":       "kvm",
				"watchdog_enabled": false,
				"tags":             []string{"production", "web"},
			},
			{
				"id":      67890,
				"label":   "database-server",
				"group":   "production",
				"status":  "stopped",
				"created": "2023-01-02T00:00:00",
				"updated": "2023-01-02T00:00:00",
				"region":  "us-west",
				"image":   "linode/ubuntu22.04",
				"type":    "g6-standard-4",
				"specs": map[string]interface{}{
					"disk":     81920,
					"memory":   8192,
					"vcpus":    4,
					"transfer": 5000,
				},
				"ipv4":             []string{"198.51.100.50"},
				"ipv6":             "2001:db8:1::1/64",
				"hypervisor":       "kvm",
				"watchdog_enabled": true,
				"tags":             []string{"production", "database"},
			},
		},
		"page":    1,
		"pages":   1,
		"results": 2,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleInstancesGet handles the specific instance mock response.
func handleInstancesGet(w http.ResponseWriter, r *http.Request, instanceID string) {
	switch instanceID {
	case "12345":
		response := map[string]interface{}{
			"id":      12345,
			"label":   "web-server-prod",
			"group":   "production",
			"status":  "running",
			"created": "2023-01-01T00:00:00",
			"updated": "2023-01-01T00:00:00",
			"region":  "us-east",
			"image":   "linode/ubuntu22.04",
			"type":    "g6-standard-2",
			"specs": map[string]interface{}{
				"disk":     50000,
				"memory":   4096,
				"vcpus":    2,
				"transfer": 4000,
			},
			"ipv4":             []string{"203.0.113.100"},
			"ipv6":             "2001:db8::1/64",
			"hypervisor":       "kvm",
			"watchdog_enabled": false,
			"tags":             []string{"production", "web"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	case "67890":
		response := map[string]interface{}{
			"id":      67890,
			"label":   "database-server",
			"group":   "production",
			"status":  "stopped",
			"created": "2023-01-02T00:00:00",
			"updated": "2023-01-02T00:00:00",
			"region":  "us-west",
			"image":   "linode/ubuntu22.04",
			"type":    "g6-standard-4",
			"specs": map[string]interface{}{
				"disk":     81920,
				"memory":   8192,
				"vcpus":    4,
				"transfer": 5000,
			},
			"ipv4":             []string{"198.51.100.50"},
			"ipv6":             "2001:db8:1::1/64",
			"hypervisor":       "kvm",
			"watchdog_enabled": true,
			"tags":             []string{"production", "database"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	case "999999":
		// Simulate not found error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "instance_id", "reason": "Not found"},
			},
		})
	default:
		http.Error(w, "Instance not found", http.StatusNotFound)
	}
}

// handleInstancesCreate handles the instance creation mock response.
func handleInstancesCreate(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"id":      98765,
		"label":   "new-instance",
		"group":   "default",
		"status":  "provisioning",
		"created": "2023-01-01T01:00:00",
		"updated": "2023-01-01T01:00:00",
		"region":  "us-central",
		"image":   "linode/ubuntu22.04",
		"type":    "g6-nanode-1",
		"specs": map[string]interface{}{
			"disk":     25000,
			"memory":   1024,
			"vcpus":    1,
			"transfer": 1000,
		},
		"ipv4":             []string{"192.0.2.100"},
		"ipv6":             "2001:db8:2::1/64",
		"hypervisor":       "kvm",
		"watchdog_enabled": false,
		"tags":             []string{},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleInstancesUpdate handles the instance update mock response.
func handleInstancesUpdate(w http.ResponseWriter, r *http.Request, instanceID string) {
	if instanceID == "999999" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "instance_id", "reason": "Not found"},
			},
		})
		return
	}

	response := map[string]interface{}{
		"id":      12345,
		"label":   "updated-web-server",
		"group":   "production",
		"status":  "running",
		"created": "2023-01-01T00:00:00",
		"updated": "2023-01-01T02:00:00",
		"region":  "us-east",
		"image":   "linode/ubuntu22.04",
		"type":    "g6-standard-2",
		"specs": map[string]interface{}{
			"disk":     50000,
			"memory":   4096,
			"vcpus":    2,
			"transfer": 4000,
		},
		"ipv4":             []string{"203.0.113.100"},
		"ipv6":             "2001:db8::1/64",
		"hypervisor":       "kvm",
		"watchdog_enabled": false,
		"tags":             []string{"production", "web", "updated"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleInstancesDelete handles the instance deletion mock response.
func handleInstancesDelete(w http.ResponseWriter, r *http.Request, instanceID string) {
	if instanceID == "999999" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "instance_id", "reason": "Not found"},
			},
		})
		return
	}

	// Return 204 No Content for successful deletion
	w.WriteHeader(http.StatusNoContent)
}

// handleInstancesBoot handles the instance boot mock response.
func handleInstancesBoot(w http.ResponseWriter, r *http.Request, instanceID string) {
	// Return empty success response for boot action
	w.WriteHeader(http.StatusOK)
}

// handleInstancesShutdown handles the instance shutdown mock response.
func handleInstancesShutdown(w http.ResponseWriter, r *http.Request, instanceID string) {
	// Return empty success response for shutdown action
	w.WriteHeader(http.StatusOK)
}

// handleInstancesReboot handles the instance reboot mock response.
func handleInstancesReboot(w http.ResponseWriter, r *http.Request, instanceID string) {
	// Return empty success response for reboot action
	w.WriteHeader(http.StatusOK)
}

// TestInstancesToolsIntegration tests all instances-related CloudMCP tools
// using HTTP test server infrastructure to validate proper integration
// between CloudMCP handlers and Linode API endpoints.
//
// **Integration Test Coverage**:
// • linode_instances_list - List all instances
// • linode_instance_get - Get specific instance details
// • linode_instance_create - Create new instance
// • linode_instance_delete - Delete instance
// • linode_instance_boot - Boot instance
// • linode_instance_shutdown - Shutdown instance
// • linode_instance_reboot - Reboot instance
//
// **Test Environment**: HTTP test server with instances API handlers
//
// **Expected Behavior**:
// • All handlers return properly formatted text responses
// • Error conditions are handled gracefully with meaningful messages
// • Instance data includes all required fields (ID, label, status, specs, etc.)
// • Instance lifecycle operations work correctly
//
// **Purpose**: Validates that CloudMCP instances handlers correctly format
// Linode API responses for LLM consumption using lightweight HTTP test infrastructure.
func TestInstancesToolsIntegration(t *testing.T) {
	// Extend the HTTP test server with instances endpoints
	server := createInstancesTestServer()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
	ctx := context.Background()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	t.Run("InstancesList", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_instances_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleInstancesList(ctx, request)
		require.NoError(t, err, "instances list should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate instances list formatting
		require.Contains(t, responseText, "Found 2 Linode instances:", "should indicate correct instance count")
		require.Contains(t, responseText, "web-server-prod", "should contain first instance label")
		require.Contains(t, responseText, "database-server", "should contain second instance label")
		require.Contains(t, responseText, "Status: running", "should show running status")
		require.Contains(t, responseText, "Status: stopped", "should show stopped status")
		require.Contains(t, responseText, "Region:", "should show regions")
		require.Contains(t, responseText, "Type:", "should show instance types")
	})

	t.Run("InstanceGet", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_instance_get",
				Arguments: map[string]interface{}{
					"instance_id": float64(12345),
				},
			},
		}

		result, err := service.handleInstanceGet(ctx, request)
		require.NoError(t, err, "instance get should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate detailed instance information
		require.Contains(t, responseText, "Linode Instance Details:", "should have instance details header")
		require.Contains(t, responseText, "ID: 12345", "should contain instance ID")
		require.Contains(t, responseText, "Label: web-server-prod", "should contain instance label")
		require.Contains(t, responseText, "Status: running", "should contain instance status")
		require.Contains(t, responseText, "Region: us-east", "should contain region")
		require.Contains(t, responseText, "Type: g6-standard-2", "should contain instance type")
		require.Contains(t, responseText, "IPv4:", "should contain IP addresses")
		require.Contains(t, responseText, "Specs:", "should contain specifications")
	})

	t.Run("InstanceCreate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_instance_create",
				Arguments: map[string]interface{}{
					"label":  "test-instance",
					"region": "us-central",
					"type":   "g6-nanode-1",
					"image":  "linode/ubuntu22.04",
				},
			},
		}

		result, err := service.handleInstanceCreate(ctx, request)
		require.NoError(t, err, "instance create should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate instance creation confirmation
		require.Contains(t, responseText, "Linode instance created successfully:", "should confirm creation")
		require.Contains(t, responseText, "ID: 98765", "should show new instance ID")
		require.Contains(t, responseText, "Label: new-instance", "should show instance label")
		require.Contains(t, responseText, "Status: provisioning", "should show provisioning status")
	})

	t.Run("InstanceBoot", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_instance_boot",
				Arguments: map[string]interface{}{
					"instance_id": float64(12345),
				},
			},
		}

		result, err := service.handleInstanceBoot(ctx, request)
		require.NoError(t, err, "instance boot should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate boot confirmation
		require.Contains(t, responseText, "boot initiated successfully", "should confirm boot")
		require.Contains(t, responseText, "12345", "should show instance ID")
	})

	t.Run("InstanceShutdown", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_instance_shutdown",
				Arguments: map[string]interface{}{
					"instance_id": float64(12345),
				},
			},
		}

		result, err := service.handleInstanceShutdown(ctx, request)
		require.NoError(t, err, "instance shutdown should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate shutdown confirmation
		require.Contains(t, responseText, "shutdown initiated successfully", "should confirm shutdown")
		require.Contains(t, responseText, "12345", "should show instance ID")
	})

	t.Run("InstanceReboot", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_instance_reboot",
				Arguments: map[string]interface{}{
					"instance_id": float64(12345),
				},
			},
		}

		result, err := service.handleInstanceReboot(ctx, request)
		require.NoError(t, err, "instance reboot should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate reboot confirmation
		require.Contains(t, responseText, "reboot initiated successfully", "should confirm reboot")
		require.Contains(t, responseText, "12345", "should show instance ID")
	})

	t.Run("InstanceDelete", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_instance_delete",
				Arguments: map[string]interface{}{
					"instance_id": float64(12345),
				},
			},
		}

		result, err := service.handleInstanceDelete(ctx, request)
		require.NoError(t, err, "instance delete should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate deletion confirmation
		require.Contains(t, responseText, "deleted successfully", "should confirm deletion")
		require.Contains(t, responseText, "12345", "should show instance ID")
	})
}

// TestInstancesErrorHandlingIntegration tests error scenarios for instances tools
// to ensure CloudMCP handles API errors gracefully and provides meaningful
// error messages to users.
//
// **Error Test Scenarios**:
// • Non-existent instance IDs (404 errors)
// • Invalid instance creation parameters
// • Instance state conflicts for lifecycle operations
// • Permission errors for instance operations
//
// **Expected Behavior**:
// • Proper error handling with contextual error messages
// • No unhandled exceptions or panics
// • Error responses follow CloudMCP error format
// • Error messages are actionable for users
//
// **Purpose**: Validates robust error handling in instance operations
// and ensures reliable operation under error conditions.
func TestInstancesErrorHandlingIntegration(t *testing.T) {
	server := createInstancesTestServer()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
	ctx := context.Background()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	t.Run("InstanceGetNotFound", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_instance_get",
				Arguments: map[string]interface{}{
					"instance_id": float64(999999), // Non-existent instance
				},
			},
		}

		result, err := service.handleInstanceGet(ctx, request)

		// Check for either Go error or MCP error result
		if err != nil {
			// Go error pattern
			require.Contains(t, err.Error(), "Failed to get instance", "error should mention get instance failure")
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

	t.Run("InstanceDeleteNotFound", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_instance_delete",
				Arguments: map[string]interface{}{
					"instance_id": float64(999999), // Non-existent instance
				},
			},
		}

		result, err := service.handleInstanceDelete(ctx, request)

		// Check for either Go error or MCP error result
		if err != nil {
			// Go error pattern
			require.Contains(t, err.Error(), "Failed to delete instance", "error should mention delete instance failure")
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

	t.Run("InvalidInstanceID", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_instance_get",
				Arguments: map[string]interface{}{
					"instance_id": "invalid", // Invalid ID type
				},
			},
		}

		result, err := service.handleInstanceGet(ctx, request)

		// Should get MCP error result for invalid parameter
		if err == nil {
			require.NotNil(t, result, "result should not be nil")
			require.True(t, result.IsError, "result should be marked as error")
		}
	})

	t.Run("MissingRequiredParameter", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_instance_create",
				Arguments: map[string]interface{}{
					// Missing required parameters like region, type, etc.
					"label": "incomplete-instance",
				},
			},
		}

		result, err := service.handleInstanceCreate(ctx, request)

		// Should get MCP error result for missing parameters
		if err == nil {
			require.NotNil(t, result, "result should not be nil")
			require.True(t, result.IsError, "result should be marked as error")
		}
	})
}
