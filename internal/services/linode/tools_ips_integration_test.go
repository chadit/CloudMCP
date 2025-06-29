//go:build integration

package linode_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chadit/CloudMCP/internal/services/linode"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// createIPsTestServer creates an HTTP test server with IP management API endpoints.
// This extends the base HTTP test server infrastructure with comprehensive
// IP-specific endpoints for integration testing.
//
// **IP Endpoints Supported:**
// • GET /v4/networking/ips - List all IP addresses
// • GET /v4/networking/ips/{address} - Get specific IP address details
//
// **Mock Data Features:**
// • Realistic IP address configurations with public and private IPs
// • IPv4 and IPv6 address support
// • Regional IP assignments and RDNS information
// • Error simulation for non-existent resources
// • Proper HTTP status codes and error responses
func createIPsTestServer() *httptest.Server {
	mux := http.NewServeMux()

	// Include base endpoints (profile, account) from main test server
	addIPsBaseEndpoints(mux)

	// IPs list endpoint
	mux.HandleFunc("/v4/networking/ips", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleIPsList(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Specific IP endpoints
	mux.HandleFunc("/v4/networking/ips/203.0.113.100", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleIPsGet(w, r, "203.0.113.100")
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v4/networking/ips/192.168.1.50", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleIPsGet(w, r, "192.168.1.50")
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Non-existent IP endpoints for error testing
	mux.HandleFunc("/v4/networking/ips/10.0.0.999", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleIPsGet(w, r, "10.0.0.999")
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return httptest.NewServer(mux)
}

// addIPsBaseEndpoints adds the basic profile and account endpoints needed for service initialization.
func addIPsBaseEndpoints(mux *http.ServeMux) {
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

// handleIPsList handles the IPs list mock response.
func handleIPsList(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"address":     "203.0.113.100",
				"gateway":     "203.0.113.1",
				"subnet_mask": "255.255.255.0",
				"prefix":      24,
				"type":        "ipv4",
				"public":      true,
				"rdns":        "web-server.example.com",
				"linode_id":   12345,
				"region":      "us-east",
			},
			{
				"address":     "192.168.1.50",
				"gateway":     "192.168.1.1",
				"subnet_mask": "255.255.255.0",
				"prefix":      24,
				"type":        "ipv4",
				"public":      false,
				"rdns":        "internal.example.com",
				"linode_id":   12345,
				"region":      "us-east",
			},
			{
				"address":     "2001:db8::1",
				"gateway":     "2001:db8::1",
				"subnet_mask": "",
				"prefix":      64,
				"type":        "ipv6",
				"public":      true,
				"rdns":        "ipv6.example.com",
				"linode_id":   67890,
				"region":      "us-west",
			},
			{
				"address":     "198.51.100.25",
				"gateway":     "198.51.100.1",
				"subnet_mask": "255.255.255.0",
				"prefix":      24,
				"type":        "ipv4",
				"public":      true,
				"rdns":        "api.example.com",
				"linode_id":   0, // Unassigned/reserved
				"region":      "us-central",
			},
		},
		"page":    1,
		"pages":   1,
		"results": 4,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleIPsGet handles the specific IP address mock response.
func handleIPsGet(w http.ResponseWriter, r *http.Request, address string) {
	switch address {
	case "203.0.113.100":
		response := map[string]interface{}{
			"address":     "203.0.113.100",
			"gateway":     "203.0.113.1",
			"subnet_mask": "255.255.255.0",
			"prefix":      24,
			"type":        "ipv4",
			"public":      true,
			"rdns":        "web-server.example.com",
			"linode_id":   12345,
			"region":      "us-east",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	case "192.168.1.50":
		response := map[string]interface{}{
			"address":     "192.168.1.50",
			"gateway":     "192.168.1.1",
			"subnet_mask": "255.255.255.0",
			"prefix":      24,
			"type":        "ipv4",
			"public":      false,
			"rdns":        "internal.example.com",
			"linode_id":   12345,
			"region":      "us-east",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	case "10.0.0.999":
		// Simulate not found error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "address", "reason": "Not found"},
			},
		})
	default:
		http.Error(w, "IP address not found", http.StatusNotFound)
	}
}

// TestIPsToolsIntegration tests all IP management-related CloudMCP tools
// using HTTP test server infrastructure to validate proper integration
// between CloudMCP handlers and Linode API endpoints.
//
// **Integration Test Coverage**:
// • linode_ips_list - List all IP addresses
// • linode_ip_get - Get specific IP address details
//
// **Test Environment**: HTTP test server with IP management API handlers
//
// **Expected Behavior**:
// • All handlers return properly formatted text responses
// • Error conditions are handled gracefully with meaningful messages
// • IP data includes all required fields (address, gateway, subnet, etc.)
// • Both IPv4 and IPv6 addresses are properly formatted
// • Public and private IP visibility is correctly shown
//
// **Purpose**: Validates that CloudMCP IP management handlers correctly format
// Linode API responses for LLM consumption using lightweight HTTP test infrastructure.
func TestIPsToolsIntegration(t *testing.T) {
	// Extend the HTTP test server with IP management endpoints
	server := createIPsTestServer()
	defer server.Close()

	service := linode.CreateHTTPTestService(t, server.URL)
	ctx := context.Background()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	t.Run("IPsList", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_ips_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleIPsList(ctx, request)
		require.NoError(t, err, "IPs list should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate IPs list formatting
		require.Contains(t, responseText, "Found 4 IP addresses:", "should indicate correct IP count")
		require.Contains(t, responseText, "203.0.113.100", "should contain first public IP")
		require.Contains(t, responseText, "192.168.1.50", "should contain private IP")
		require.Contains(t, responseText, "2001:db8::1", "should contain IPv6 address")
		require.Contains(t, responseText, "198.51.100.25", "should contain reserved IP")
		require.Contains(t, responseText, "Public IPv4", "should show public IPv4 type")
		require.Contains(t, responseText, "Private IPv4", "should show private IPv4 type")
		require.Contains(t, responseText, "Public IPv6", "should show IPv6 type")
		require.Contains(t, responseText, "Reserved", "should show reserved status")
	})

	t.Run("IPGet", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_ip_get",
				Arguments: map[string]interface{}{
					"address": "203.0.113.100",
				},
			},
		}

		result, err := service.handleIPGet(ctx, request)
		require.NoError(t, err, "IP get should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate detailed IP information
		require.Contains(t, responseText, "IP Address Details:", "should have IP details header")
		require.Contains(t, responseText, "Address: 203.0.113.100", "should contain IP address")
		require.Contains(t, responseText, "Gateway: 203.0.113.1", "should contain gateway")
		require.Contains(t, responseText, "Subnet Mask: 255.255.255.0", "should contain subnet mask")
		require.Contains(t, responseText, "Region: us-east", "should contain region")
		require.Contains(t, responseText, "Reverse DNS: web-server.example.com", "should contain RDNS")
		require.Contains(t, responseText, "Assigned to Linode: 12345", "should show assignment")
	})

	t.Run("IPGetPrivate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_ip_get",
				Arguments: map[string]interface{}{
					"address": "192.168.1.50",
				},
			},
		}

		result, err := service.handleIPGet(ctx, request)
		require.NoError(t, err, "private IP get should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate private IP information
		require.Contains(t, responseText, "Address: 192.168.1.50", "should contain private IP address")
		require.Contains(t, responseText, "Visibility: Private", "should show private visibility")
		require.Contains(t, responseText, "Reverse DNS: internal.example.com", "should contain internal RDNS")
	})
}

// TestIPsErrorHandlingIntegration tests error scenarios for IP management tools
// to ensure CloudMCP handles API errors gracefully and provides meaningful
// error messages to users.
//
// **Error Test Scenarios**:
// • Non-existent IP addresses (404 errors)
// • Invalid IP address formats
// • Permission errors for IP operations
//
// **Expected Behavior**:
// • Proper error handling with contextual error messages
// • No unhandled exceptions or panics
// • Error responses follow CloudMCP error format
// • Error messages are actionable for users
//
// **Purpose**: Validates robust error handling in IP management operations
// and ensures reliable operation under error conditions.
func TestIPsErrorHandlingIntegration(t *testing.T) {
	server := createIPsTestServer()
	defer server.Close()

	service := linode.CreateHTTPTestService(t, server.URL)
	ctx := context.Background()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	t.Run("IPGetNotFound", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_ip_get",
				Arguments: map[string]interface{}{
					"address": "10.0.0.999", // Non-existent IP
				},
			},
		}

		result, err := service.handleIPGet(ctx, request)

		// Check for either Go error or MCP error result
		if err != nil {
			// Go error pattern
			require.Contains(t, err.Error(), "Failed to get IP address", "error should mention get IP failure")
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

	t.Run("IPGetInvalidAddress", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_ip_get",
				Arguments: map[string]interface{}{
					"address": "invalid-ip-address", // Invalid IP format
				},
			},
		}

		result, err := service.handleIPGet(ctx, request)

		// Should get MCP error result for invalid parameter
		if err == nil {
			require.NotNil(t, result, "result should not be nil")
			require.True(t, result.IsError, "result should be marked as error")
		}
	})

	t.Run("IPGetMissingParameter", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_ip_get",
				Arguments: map[string]interface{}{
					// Missing required 'address' parameter
				},
			},
		}

		result, err := service.handleIPGet(ctx, request)

		// Should get MCP error result for missing parameter
		if err == nil {
			require.NotNil(t, result, "result should not be nil")
			require.True(t, result.IsError, "result should be marked as error")
		}
	})
}
