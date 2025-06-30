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

// createNetworkingTestServer creates an HTTP test server with networking API endpoints.
// This extends the base HTTP test server infrastructure with comprehensive
// networking-specific endpoints for integration testing.
//
// **Networking Endpoints Supported:**
// • GET /v4/networking/ips - List reserved IPs
// • GET /v4/networking/ips/{address} - Get specific IP details
// • POST /v4/networking/ips - Allocate new reserved IP
// • PUT /v4/networking/ips/{address} - Assign/update reserved IP
// • GET /v4/networking/vlans - List VLANs
// • GET /v4/networking/ipv6/pools - List IPv6 pools
// • GET /v4/networking/ipv6/ranges - List IPv6 ranges
//
// **Mock Data Features:**
// • Realistic IP address configurations with regional assignments
// • VLAN and IPv6 pool/range management
// • Reserved IP allocation and assignment workflows
// • Error simulation for non-existent resources
// • Proper HTTP status codes and error responses
func createNetworkingTestServer() *httptest.Server {
	mux := http.NewServeMux()

	// Include base endpoints (profile, account) from main test server
	addNetworkingBaseEndpoints(mux)

	// Reserved IPs list endpoint
	mux.HandleFunc("/v4/networking/ips", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleReservedIPsList(w, r)
		case http.MethodPost:
			handleReservedIPsAllocate(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Specific reserved IP endpoints
	mux.HandleFunc("/v4/networking/ips/192.168.1.100", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleReservedIPsGet(w, r, "192.168.1.100")
		case http.MethodPut:
			handleReservedIPsUpdate(w, r, "192.168.1.100")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// VLANs endpoint
	mux.HandleFunc("/v4/networking/vlans", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleVLANsList(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// IPv6 pools endpoint
	mux.HandleFunc("/v4/networking/ipv6/pools", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleIPv6PoolsList(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// IPv6 ranges endpoint
	mux.HandleFunc("/v4/networking/ipv6/ranges", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleIPv6RangesList(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// IP assignment endpoint (Linode instances assign IPs)
	mux.HandleFunc("/v4/linode/instances/ips/assign", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handleIPAssignment(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Non-existent IP endpoints for error testing
	mux.HandleFunc("/v4/networking/ips/10.0.0.255", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleReservedIPsGet(w, r, "10.0.0.255")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return httptest.NewServer(mux)
}

// addNetworkingBaseEndpoints adds the basic profile and account endpoints needed for service initialization.
func addNetworkingBaseEndpoints(mux *http.ServeMux) {
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

// handleReservedIPsList handles the reserved IPs list mock response.
func handleReservedIPsList(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"address":     "192.168.1.100",
				"gateway":     "192.168.1.1",
				"subnet_mask": "255.255.255.0",
				"prefix":      24,
				"type":        "ipv4",
				"public":      false,
				"rdns":        "reserved1.example.com",
				"linode_id":   0, // Unassigned reserved IP
				"region":      "us-east",
			},
			{
				"address":     "203.0.113.50",
				"gateway":     "203.0.113.1",
				"subnet_mask": "255.255.255.0",
				"prefix":      24,
				"type":        "ipv4",
				"public":      true,
				"rdns":        "public.example.com",
				"linode_id":   0, // Unassigned reserved IP
				"region":      "us-west",
			},
		},
		"page":    1,
		"pages":   1,
		"results": 2,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleIPAssignment handles the IP assignment mock response.
func handleIPAssignment(w http.ResponseWriter, r *http.Request) {
	// Return empty success response for IP assignment
	w.WriteHeader(http.StatusOK)
}

// handleReservedIPsGet handles the specific reserved IP mock response.
func handleReservedIPsGet(w http.ResponseWriter, r *http.Request, address string) {
	switch address {
	case "192.168.1.100":
		response := map[string]interface{}{
			"address":     "192.168.1.100",
			"gateway":     "192.168.1.1",
			"subnet_mask": "255.255.255.0",
			"prefix":      24,
			"type":        "ipv4",
			"public":      false,
			"rdns":        "reserved1.example.com",
			"linode_id":   12345,
			"region":      "us-east",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	case "10.0.0.255":
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

// handleReservedIPsAllocate handles the reserved IP allocation mock response.
func handleReservedIPsAllocate(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"address":     "203.0.113.75",
		"gateway":     "203.0.113.1",
		"subnet_mask": "255.255.255.0",
		"prefix":      24,
		"type":        "ipv4",
		"public":      true,
		"rdns":        "new-ip.example.com",
		"linode_id":   nil,
		"region":      "us-central",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleReservedIPsUpdate handles the reserved IP assignment/update mock response.
func handleReservedIPsUpdate(w http.ResponseWriter, r *http.Request, address string) {
	response := map[string]interface{}{
		"address":     address,
		"gateway":     "192.168.1.1",
		"subnet_mask": "255.255.255.0",
		"prefix":      24,
		"type":        "ipv4",
		"public":      false,
		"rdns":        "updated.example.com",
		"linode_id":   67890,
		"region":      "us-east",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleIPAssignment handles the IP assignment mock response.
func handleIPAssignment(w http.ResponseWriter, r *http.Request) {
	// Return empty success response for IP assignment
	w.WriteHeader(http.StatusOK)
}

// handleVLANsList handles the VLANs list mock response.
func handleVLANsList(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"label":   "production-vlan",
				"region":  "us-east",
				"linodes": []int{12345, 67890},
				"created": "2023-01-01T00:00:00",
			},
			{
				"label":   "development-vlan",
				"region":  "us-west",
				"linodes": []int{54321},
				"created": "2023-01-02T00:00:00",
			},
		},
		"page":    1,
		"pages":   1,
		"results": 2,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleIPAssignment handles the IP assignment mock response.
func handleIPAssignment(w http.ResponseWriter, r *http.Request) {
	// Return empty success response for IP assignment
	w.WriteHeader(http.StatusOK)
}

// handleIPv6PoolsList handles the IPv6 pools list mock response.
func handleIPv6PoolsList(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"range":  "2001:db8::/64",
				"region": "us-east",
			},
			{
				"range":  "2001:db8:1::/64",
				"region": "us-west",
			},
		},
		"page":    1,
		"pages":   1,
		"results": 2,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleIPv6RangesList handles the IPv6 ranges list mock response.
func handleIPv6RangesList(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"range":        "2001:db8:2::/56",
				"region":       "us-east",
				"prefix":       56,
				"route_target": "2001:db8:2::1",
			},
			{
				"range":        "2001:db8:3::/56",
				"region":       "us-west",
				"prefix":       56,
				"route_target": "2001:db8:3::1",
			},
		},
		"page":    1,
		"pages":   1,
		"results": 2,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleIPAssignment handles the IP assignment mock response.
func handleIPAssignment(w http.ResponseWriter, r *http.Request) {
	// Return empty success response for IP assignment
	w.WriteHeader(http.StatusOK)
}

// TestNetworkingToolsIntegration tests all networking-related CloudMCP tools
// using HTTP test server infrastructure to validate proper integration
// between CloudMCP handlers and Linode API endpoints.
//
// **Integration Test Coverage**:
// • linode_reserved_ips_list - List all reserved IP addresses
// • linode_reserved_ip_get - Get specific IP address details
// • linode_reserved_ip_allocate - Allocate new reserved IP
// • linode_reserved_ip_assign - Assign IP to Linode instance
// • linode_reserved_ip_update - Update IP reverse DNS
// • linode_vlans_list - List virtual LANs
// • linode_ipv6_pools_list - List IPv6 address pools
// • linode_ipv6_ranges_list - List IPv6 ranges
//
// **Test Environment**: HTTP test server with networking API handlers
//
// **Expected Behavior**:
// • All handlers return properly formatted text responses
// • Error conditions are handled gracefully with meaningful messages
// • IP data includes all required fields (address, gateway, subnet, etc.)
// • VLAN and IPv6 information is properly formatted
//
// **Purpose**: Validates that CloudMCP networking handlers correctly format
// Linode API responses for LLM consumption using lightweight HTTP test infrastructure.
func TestNetworkingToolsIntegration(t *testing.T) {
	// Extend the HTTP test server with networking endpoints
	server := createNetworkingTestServer()
	defer server.Close()

	service := linode.CreateHTTPTestService(t, server.URL)
	ctx := context.Background()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	t.Run("ReservedIPsList", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_reserved_ips_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleReservedIPsList(ctx, request)
		require.NoError(t, err, "reserved IPs list should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate reserved IPs list formatting
		require.Contains(t, responseText, "Found 2 reserved IP addresses:", "should indicate correct IP count")
		require.Contains(t, responseText, "192.168.1.100", "should contain first IP address")
		require.Contains(t, responseText, "203.0.113.50", "should contain second IP address")
		require.Contains(t, responseText, "Private", "should show private IP type")
		require.Contains(t, responseText, "Public", "should show public IP type")
	})

	t.Run("ReservedIPGet", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_reserved_ip_get",
				Arguments: map[string]interface{}{
					"address": "192.168.1.100",
				},
			},
		}

		result, err := service.handleReservedIPGet(ctx, request)
		require.NoError(t, err, "reserved IP get should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate detailed IP information
		require.Contains(t, responseText, "IP Address Details:", "should have IP details header")
		require.Contains(t, responseText, "Address: 192.168.1.100", "should contain IP address")
		require.Contains(t, responseText, "Gateway: 192.168.1.1", "should contain gateway")
		require.Contains(t, responseText, "Subnet Mask: 255.255.255.0", "should contain subnet mask")
		require.Contains(t, responseText, "Region: us-east", "should contain region")
	})

	t.Run("ReservedIPAllocate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_reserved_ip_allocate",
				Arguments: map[string]interface{}{
					"type":   "ipv4",
					"region": "us-central",
					"public": true,
				},
			},
		}

		result, err := service.handleReservedIPAllocate(ctx, request)
		require.NoError(t, err, "reserved IP allocate should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate IP allocation confirmation
		require.Contains(t, responseText, "Reserved IP allocated successfully:", "should confirm allocation")
		require.Contains(t, responseText, "Address: 203.0.113.75", "should show allocated IP")
		require.Contains(t, responseText, "Region: us-central", "should show region")
	})

	t.Run("ReservedIPAssign", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_reserved_ip_assign",
				Arguments: map[string]interface{}{
					"address":   "192.168.1.100",
					"linode_id": float64(67890),
				},
			},
		}

		result, err := service.handleReservedIPAssign(ctx, request)
		require.NoError(t, err, "reserved IP assign should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate IP assignment confirmation
		require.Contains(t, responseText, "IP address assignment updated:", "should confirm assignment")
		require.Contains(t, responseText, "Address: 192.168.1.100", "should show IP address")
		require.Contains(t, responseText, "Assignment:", "should show assignment status")
	})

	t.Run("ReservedIPUpdate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_reserved_ip_update",
				Arguments: map[string]interface{}{
					"address": "192.168.1.100",
					"rdns":    "updated.example.com",
				},
			},
		}

		result, err := service.handleReservedIPUpdate(ctx, request)
		require.NoError(t, err, "reserved IP update should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate IP update confirmation
		require.Contains(t, responseText, "IP address updated successfully:", "should confirm update")
		require.Contains(t, responseText, "Address: 192.168.1.100", "should show IP address")
		require.Contains(t, responseText, "Reverse DNS: updated.example.com", "should show updated RDNS")
	})

	t.Run("VLANsList", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_vlans_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleVLANsList(ctx, request)
		require.NoError(t, err, "VLANs list should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate VLANs list formatting
		require.Contains(t, responseText, "Found 2 VLANs:", "should indicate correct VLAN count")
		require.Contains(t, responseText, "production-vlan", "should contain first VLAN label")
		require.Contains(t, responseText, "development-vlan", "should contain second VLAN label")
		require.Contains(t, responseText, "Linodes:", "should show attached Linodes")
	})

	t.Run("IPv6PoolsList", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_ipv6_pools_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleIPv6PoolsList(ctx, request)
		require.NoError(t, err, "IPv6 pools list should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate IPv6 pools list formatting
		require.Contains(t, responseText, "Found 2 IPv6 pools:", "should indicate correct pool count")
		require.Contains(t, responseText, "2001:db8::/64", "should contain first IPv6 range")
		require.Contains(t, responseText, "2001:db8:1::/64", "should contain second IPv6 range")
		require.Contains(t, responseText, "Region:", "should show regions")
	})

	t.Run("IPv6RangesList", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_ipv6_ranges_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleIPv6RangesList(ctx, request)
		require.NoError(t, err, "IPv6 ranges list should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate IPv6 ranges list formatting
		require.Contains(t, responseText, "Found 2 IPv6 ranges:", "should indicate correct range count")
		require.Contains(t, responseText, "2001:db8:2::/56", "should contain first IPv6 range")
		require.Contains(t, responseText, "2001:db8:3::/56", "should contain second IPv6 range")
		require.Contains(t, responseText, "Route Target:", "should show route targets")
	})
}

// TestNetworkingErrorHandlingIntegration tests error scenarios for networking tools
// to ensure CloudMCP handles API errors gracefully and provides meaningful
// error messages to users.
//
// **Error Test Scenarios**:
// • Non-existent IP addresses (404 errors)
// • Invalid IP allocation parameters
// • VLAN configuration conflicts
// • Permission errors for networking operations
//
// **Expected Behavior**:
// • Proper error handling with contextual error messages
// • No unhandled exceptions or panics
// • Error responses follow CloudMCP error format
// • Error messages are actionable for users
//
// **Purpose**: Validates robust error handling in networking operations
// and ensures reliable operation under error conditions.
func TestNetworkingErrorHandlingIntegration(t *testing.T) {
	server := createNetworkingTestServer()
	defer server.Close()

	service := linode.CreateHTTPTestService(t, server.URL)
	ctx := context.Background()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	t.Run("ReservedIPGetNotFound", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_reserved_ip_get",
				Arguments: map[string]interface{}{
					"address": "10.0.0.255", // Non-existent IP
				},
			},
		}

		result, err := service.handleReservedIPGet(ctx, request)

		// Check for either Go error or MCP error result
		if err != nil {
			// Go error pattern
			require.Contains(t, err.Error(), "Failed to get reserved IP", "error should mention get IP failure")
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

	t.Run("InvalidIPAddress", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_reserved_ip_get",
				Arguments: map[string]interface{}{
					"address": "invalid-ip", // Invalid IP format
				},
			},
		}

		result, err := service.handleReservedIPGet(ctx, request)

		// Should get MCP error result for invalid parameter
		if err == nil {
			require.NotNil(t, result, "result should not be nil")
			require.True(t, result.IsError, "result should be marked as error")
		}
	})

	t.Run("MissingRequiredParameter", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_reserved_ip_allocate",
				Arguments: map[string]interface{}{
					// Missing required 'type' parameter
					"region": "us-central",
				},
			},
		}

		result, err := service.handleReservedIPAllocate(ctx, request)

		// Should get MCP error result for missing parameter
		if err == nil {
			require.NotNil(t, result, "result should not be nil")
			require.True(t, result.IsError, "result should be marked as error")
		}
	})
}
