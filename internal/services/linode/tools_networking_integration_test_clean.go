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

// createNetworkingTestServerClean creates an HTTP test server with networking API endpoints.
func createNetworkingTestServerClean() *httptest.Server {
	mux := http.NewServeMux()

	// Include base endpoints (profile, account) from main test server
	addNetworkingBaseEndpointsClean(mux)

	// Reserved IPs list endpoint
	mux.HandleFunc("/v4/networking/ips", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleReservedIPsListClean(w, r)
		case http.MethodPost:
			handleReservedIPsAllocateClean(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Specific reserved IP endpoints
	mux.HandleFunc("/v4/networking/ips/192.168.1.100", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleReservedIPsGetClean(w, r, "192.168.1.100")
		case http.MethodPut:
			handleReservedIPsUpdateClean(w, r, "192.168.1.100")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// IP assignment endpoint (Linode instances assign IPs)
	mux.HandleFunc("/v4/linode/instances/ips/assign", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handleIPAssignmentClean(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// VLANs endpoint
	mux.HandleFunc("/v4/networking/vlans", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleVLANsListClean(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// IPv6 pools endpoint
	mux.HandleFunc("/v4/networking/ipv6/pools", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleIPv6PoolsListClean(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// IPv6 ranges endpoint
	mux.HandleFunc("/v4/networking/ipv6/ranges", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleIPv6RangesListClean(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Non-existent IP endpoints for error testing
	mux.HandleFunc("/v4/networking/ips/10.0.0.255", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleReservedIPsGetClean(w, r, "10.0.0.255")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return httptest.NewServer(mux)
}

// addNetworkingBaseEndpointsClean adds the basic profile and account endpoints needed for service initialization.
func addNetworkingBaseEndpointsClean(mux *http.ServeMux) {
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

// handleReservedIPsListClean handles the reserved IPs list mock response.
func handleReservedIPsListClean(w http.ResponseWriter, r *http.Request) {
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

// handleReservedIPsGetClean handles the specific reserved IP mock response.
func handleReservedIPsGetClean(w http.ResponseWriter, r *http.Request, address string) {
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

// handleReservedIPsAllocateClean handles the reserved IP allocation mock response.
func handleReservedIPsAllocateClean(w http.ResponseWriter, r *http.Request) {
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

// handleReservedIPsUpdateClean handles the reserved IP assignment/update mock response.
func handleReservedIPsUpdateClean(w http.ResponseWriter, r *http.Request, address string) {
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

// handleIPAssignmentClean handles the IP assignment mock response.
func handleIPAssignmentClean(w http.ResponseWriter, r *http.Request) {
	// Return empty success response for IP assignment
	w.WriteHeader(http.StatusOK)
}

// handleVLANsListClean handles the VLANs list mock response.
func handleVLANsListClean(w http.ResponseWriter, r *http.Request) {
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

// handleIPv6PoolsListClean handles the IPv6 pools list mock response.
func handleIPv6PoolsListClean(w http.ResponseWriter, r *http.Request) {
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

// handleIPv6RangesListClean handles the IPv6 ranges list mock response.
func handleIPv6RangesListClean(w http.ResponseWriter, r *http.Request) {
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

// TestNetworkingToolsIntegrationClean tests all networking-related CloudMCP tools
func TestNetworkingToolsIntegrationClean(t *testing.T) {
	// Extend the HTTP test server with networking endpoints
	server := createNetworkingTestServerClean()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
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
}
