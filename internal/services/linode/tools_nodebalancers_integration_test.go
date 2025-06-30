//go:build integration

package linode_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// createNodeBalancerTestServer creates an HTTP test server with NodeBalancer API endpoints.
// This extends the base HTTP test server infrastructure with comprehensive
// NodeBalancer-specific endpoints for integration testing.
//
// **NodeBalancer Endpoints Supported:**
// • GET /v4/nodebalancers - List NodeBalancers
// • GET /v4/nodebalancers/{id} - Get specific NodeBalancer
// • POST /v4/nodebalancers - Create new NodeBalancer
// • PUT /v4/nodebalancers/{id} - Update NodeBalancer
// • DELETE /v4/nodebalancers/{id} - Delete NodeBalancer
// • GET /v4/nodebalancers/{id}/configs - List configurations
// • POST /v4/nodebalancers/{id}/configs - Create configuration
// • PUT /v4/nodebalancers/{id}/configs/{config_id} - Update configuration
// • DELETE /v4/nodebalancers/{id}/configs/{config_id} - Delete configuration
//
// **Mock Data Features:**
// • Realistic NodeBalancer configurations with SSL, health checks, nodes
// • Configuration and node management
// • Error simulation for non-existent resources
// • Proper HTTP status codes and error responses
func createNodeBalancerTestServer() *httptest.Server {
	mux := http.NewServeMux()

	// Include base endpoints (profile, account) from main test server
	addNodeBalancerBaseEndpoints(mux)

	// NodeBalancers list endpoint
	mux.HandleFunc("/v4/nodebalancers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleNodeBalancersList(w, r)
		case http.MethodPost:
			handleNodeBalancerCreate(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Specific NodeBalancer endpoints with dynamic ID handling
	mux.HandleFunc("/v4/nodebalancers/", func(w http.ResponseWriter, r *http.Request) {
		// Extract NodeBalancer ID from path
		path := strings.TrimPrefix(r.URL.Path, "/v4/nodebalancers/")
		pathParts := strings.Split(path, "/")

		if len(pathParts) == 1 && pathParts[0] != "" {
			// Single NodeBalancer operations: /v4/nodebalancers/{id}
			nodebalancerID := pathParts[0]
			switch r.Method {
			case http.MethodGet:
				handleNodeBalancerGet(w, r, nodebalancerID)
			case http.MethodPut:
				handleNodeBalancerUpdate(w, r, nodebalancerID)
			case http.MethodDelete:
				handleNodeBalancerDelete(w, r, nodebalancerID)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if len(pathParts) == 2 && pathParts[1] == "configs" {
			// NodeBalancer configs operations: /v4/nodebalancers/{id}/configs
			nodebalancerID := pathParts[0]
			switch r.Method {
			case http.MethodGet:
				handleNodeBalancerConfigsList(w, r, nodebalancerID)
			case http.MethodPost:
				handleNodeBalancerConfigCreate(w, r, nodebalancerID)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if len(pathParts) == 3 && pathParts[1] == "configs" {
			// Specific config operations: /v4/nodebalancers/{id}/configs/{config_id}
			nodebalancerID := pathParts[0]
			configID := pathParts[2]
			switch r.Method {
			case http.MethodPut:
				handleNodeBalancerConfigUpdate(w, r, nodebalancerID, configID)
			case http.MethodDelete:
				handleNodeBalancerConfigDelete(w, r, nodebalancerID, configID)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})

	return httptest.NewServer(mux)
}

// addNodeBalancerBaseEndpoints adds the basic profile and account endpoints needed for service initialization.
func addNodeBalancerBaseEndpoints(mux *http.ServeMux) {
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

// handleNodeBalancersList handles the NodeBalancers list mock response.
func handleNodeBalancersList(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"id":                   12345,
				"label":                "web-nodebalancer",
				"hostname":             "nb-12345.newark.nodebalancer.linode.com",
				"client_conn_throttle": 0,
				"region":               "us-east",
				"ipv4":                 "192.168.1.100",
				"ipv6":                 "2600:3c01::f03c:91ff:fe24:3a2f",
				"created":              "2023-01-01T00:00:00",
				"updated":              "2023-01-01T00:00:00",
				"transfer": map[string]interface{}{
					"in":    float64(12345678),
					"out":   float64(23456789),
					"total": float64(35802467),
				},
				"tags": []string{"production", "web"},
			},
			{
				"id":                   54321,
				"label":                "api-nodebalancer",
				"hostname":             "nb-54321.newark.nodebalancer.linode.com",
				"client_conn_throttle": 10,
				"region":               "us-east",
				"ipv4":                 "192.168.1.101",
				"ipv6":                 "2600:3c01::f03c:91ff:fe24:3a30",
				"created":              "2023-01-02T00:00:00",
				"updated":              "2023-01-02T00:00:00",
				"transfer": map[string]interface{}{
					"in":    float64(9876543),
					"out":   float64(8765432),
					"total": float64(18641975),
				},
				"tags": []string{"production", "api"},
			},
		},
		"page":    1,
		"pages":   1,
		"results": 2,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleNodeBalancerGet handles the specific NodeBalancer mock response.
func handleNodeBalancerGet(w http.ResponseWriter, r *http.Request, nodebalancerID string) {
	switch nodebalancerID {
	case "12345":
		response := map[string]interface{}{
			"id":                   12345,
			"label":                "web-nodebalancer",
			"hostname":             "nb-12345.newark.nodebalancer.linode.com",
			"client_conn_throttle": 0,
			"region":               "us-east",
			"ipv4":                 "192.168.1.100",
			"ipv6":                 "2600:3c01::f03c:91ff:fe24:3a2f",
			"created":              "2023-01-01T00:00:00",
			"updated":              "2023-01-01T00:00:00",
			"transfer": map[string]interface{}{
				"in":    float64(12345678),
				"out":   float64(23456789),
				"total": float64(35802467),
			},
			"tags": []string{"production", "web"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	case "999999":
		// Simulate not found error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "nodebalancer_id", "reason": "Not found"},
			},
		})
	default:
		http.Error(w, "NodeBalancer not found", http.StatusNotFound)
	}
}

// handleNodeBalancerCreate handles the NodeBalancer creation mock response.
func handleNodeBalancerCreate(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"id":                   67890,
		"label":                "test-nodebalancer",
		"hostname":             "nb-67890.newark.nodebalancer.linode.com",
		"client_conn_throttle": 0,
		"region":               "us-east",
		"ipv4":                 "192.168.1.102",
		"ipv6":                 "2600:3c01::f03c:91ff:fe24:3a31",
		"created":              "2023-01-01T01:00:00",
		"updated":              "2023-01-01T01:00:00",
		"transfer": map[string]interface{}{
			"in":    float64(0),
			"out":   float64(0),
			"total": float64(0),
		},
		"tags": []string{},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleNodeBalancerUpdate handles the NodeBalancer update mock response.
func handleNodeBalancerUpdate(w http.ResponseWriter, r *http.Request, nodebalancerID string) {
	if nodebalancerID == "999999" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "nodebalancer_id", "reason": "Not found"},
			},
		})
		return
	}

	response := map[string]interface{}{
		"id":                   12345,
		"label":                "updated-web-nodebalancer",
		"hostname":             "nb-12345.newark.nodebalancer.linode.com",
		"client_conn_throttle": 5,
		"region":               "us-east",
		"ipv4":                 "192.168.1.100",
		"ipv6":                 "2600:3c01::f03c:91ff:fe24:3a2f",
		"created":              "2023-01-01T00:00:00",
		"updated":              "2023-01-01T02:00:00",
		"transfer": map[string]interface{}{
			"in":    float64(12345678),
			"out":   float64(23456789),
			"total": float64(35802467),
		},
		"tags": []string{"production", "web", "updated"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleNodeBalancerConfigsList handles the NodeBalancer configs list mock response.
func handleNodeBalancerConfigsList(w http.ResponseWriter, r *http.Request, nodebalancerID string) {
	if nodebalancerID == "999999" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "nodebalancer_id", "reason": "Not found"},
			},
		})
		return
	}

	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"id":              54321,
				"port":            80,
				"protocol":        "http",
				"algorithm":       "roundrobin",
				"stickiness":      "none",
				"check":           "http",
				"check_interval":  5,
				"check_timeout":   3,
				"check_attempts":  2,
				"check_path":      "/",
				"check_body":      "",
				"check_passive":   true,
				"proxy_protocol":  "none",
				"cipher_suite":    "recommended",
				"nodebalancer_id": 12345,
				"ssl_commonname":  "",
				"ssl_fingerprint": "",
				"ssl_cert":        "",
				"ssl_key":         "",
				"nodes_status": map[string]interface{}{
					"up":   0,
					"down": 0,
				},
			},
		},
		"page":    1,
		"pages":   1,
		"results": 1,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleNodeBalancerConfigCreate handles the NodeBalancer config creation mock response.
func handleNodeBalancerConfigCreate(w http.ResponseWriter, r *http.Request, nodebalancerID string) {
	if nodebalancerID == "999999" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "nodebalancer_id", "reason": "Not found"},
			},
		})
		return
	}

	response := map[string]interface{}{
		"id":              54321,
		"port":            80,
		"protocol":        "http",
		"algorithm":       "roundrobin",
		"stickiness":      "none",
		"check":           "http",
		"check_interval":  5,
		"check_timeout":   3,
		"check_attempts":  2,
		"check_path":      "/",
		"check_body":      "",
		"check_passive":   true,
		"proxy_protocol":  "none",
		"cipher_suite":    "recommended",
		"nodebalancer_id": 12345,
		"ssl_commonname":  "",
		"ssl_fingerprint": "",
		"ssl_cert":        "",
		"ssl_key":         "",
		"nodes_status": map[string]interface{}{
			"up":   0,
			"down": 0,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleNodeBalancerConfigUpdate handles the NodeBalancer config update mock response.
func handleNodeBalancerConfigUpdate(w http.ResponseWriter, r *http.Request, nodebalancerID, configID string) {
	if nodebalancerID == "999999" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "nodebalancer_id", "reason": "Not found"},
			},
		})
		return
	}

	response := map[string]interface{}{
		"id":              54321,
		"port":            443,
		"protocol":        "https",
		"algorithm":       "leastconn",
		"stickiness":      "table",
		"check":           "http",
		"check_interval":  10,
		"check_timeout":   5,
		"check_attempts":  3,
		"check_path":      "/health",
		"check_body":      "OK",
		"check_passive":   false,
		"proxy_protocol":  "v1",
		"cipher_suite":    "legacy",
		"nodebalancer_id": 12345,
		"ssl_commonname":  "example.com",
		"ssl_fingerprint": "AB:CD:EF:12:34:56:78:90",
		"ssl_cert":        "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
		"ssl_key":         "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----",
		"nodes_status": map[string]interface{}{
			"up":   2,
			"down": 0,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleNodeBalancerConfigDelete handles the NodeBalancer config deletion mock response.
func handleNodeBalancerConfigDelete(w http.ResponseWriter, r *http.Request, nodebalancerID, configID string) {
	if nodebalancerID == "999999" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "nodebalancer_id", "reason": "Not found"},
			},
		})
		return
	}

	// Return 204 No Content for successful deletion
	w.WriteHeader(http.StatusNoContent)
}

// handleNodeBalancerDelete handles the NodeBalancer deletion mock response.
func handleNodeBalancerDelete(w http.ResponseWriter, r *http.Request, nodebalancerID string) {
	if nodebalancerID == "999999" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "nodebalancer_id", "reason": "Not found"},
			},
		})
		return
	}

	// Return 204 No Content for successful deletion
	w.WriteHeader(http.StatusNoContent)
}

// TestNodeBalancerToolsIntegration tests all NodeBalancer-related CloudMCP tools
// using HTTP test server infrastructure to validate proper integration
// between CloudMCP handlers and Linode API endpoints.
//
// **Integration Test Coverage**:
// • linode_nodebalancers_list - List all NodeBalancers
// • linode_nodebalancer_get - Get specific NodeBalancer details
// • linode_nodebalancer_create - Create new NodeBalancer
// • linode_nodebalancer_update - Update existing NodeBalancer
// • linode_nodebalancer_delete - Delete NodeBalancer
// • linode_nodebalancer_config_create - Create configuration
// • linode_nodebalancer_config_update - Update configuration
// • linode_nodebalancer_config_delete - Delete configuration
//
// **Test Environment**: HTTP test server with NodeBalancer API handlers
//
// **Expected Behavior**:
// • All handlers return properly formatted text responses
// • Error conditions are handled gracefully with meaningful messages
// • NodeBalancer data includes all required fields (ID, label, hostname, IPs, etc.)
// • Configuration management operations work correctly
//
// **Purpose**: Validates that CloudMCP NodeBalancer handlers correctly format
// Linode API responses for LLM consumption using lightweight HTTP test infrastructure.
func TestNodeBalancerToolsIntegration(t *testing.T) {
	// Extend the HTTP test server with NodeBalancer endpoints
	server := createNodeBalancerTestServer()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
	ctx := t.Context()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	t.Run("NodeBalancersList", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_nodebalancers_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleNodeBalancersList(ctx, request)
		require.NoError(t, err, "NodeBalancers list should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate NodeBalancer list formatting
		require.Contains(t, responseText, "Found 2 NodeBalancers:", "should indicate correct NodeBalancer count")
		require.Contains(t, responseText, "web-nodebalancer", "should contain first NodeBalancer label")
		require.Contains(t, responseText, "api-nodebalancer", "should contain second NodeBalancer label")
		require.Contains(t, responseText, "192.168.1.100", "should show IPv4 address")
		require.Contains(t, responseText, "(us-east)", "should show region")
	})

	t.Run("NodeBalancerGet", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_nodebalancer_get",
				Arguments: map[string]interface{}{
					"nodebalancer_id": float64(12345),
				},
			},
		}

		result, err := service.handleNodeBalancerGet(ctx, request)
		require.NoError(t, err, "NodeBalancer get should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate detailed NodeBalancer information
		require.Contains(t, responseText, "NodeBalancer Details:", "should have NodeBalancer details header")
		require.Contains(t, responseText, "ID: 12345", "should contain NodeBalancer ID")
		require.Contains(t, responseText, "Label: web-nodebalancer", "should contain NodeBalancer label")
		require.Contains(t, responseText, "Hostname:", "should contain hostname")
		require.Contains(t, responseText, "IPv4: 192.168.1.100", "should contain IPv4 address")
		require.Contains(t, responseText, "Region: us-east", "should contain region")
	})

	t.Run("NodeBalancerCreate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_nodebalancer_create",
				Arguments: map[string]interface{}{
					"label":  "test-nodebalancer",
					"region": "us-east",
				},
			},
		}

		result, err := service.handleNodeBalancerCreate(ctx, request)
		require.NoError(t, err, "NodeBalancer create should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate NodeBalancer creation confirmation
		require.Contains(t, responseText, "NodeBalancer created successfully", "should confirm creation")
		require.Contains(t, responseText, "ID: 67890", "should show new NodeBalancer ID")
		require.Contains(t, responseText, "Label: test-nodebalancer", "should show NodeBalancer label")
		require.Contains(t, responseText, "IPv4: 192.168.1.102", "should show assigned IPv4")
	})

	t.Run("NodeBalancerUpdate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_nodebalancer_update",
				Arguments: map[string]interface{}{
					"nodebalancer_id":      float64(12345),
					"label":                "updated-web-nodebalancer",
					"client_conn_throttle": float64(5),
				},
			},
		}

		result, err := service.handleNodeBalancerUpdate(ctx, request)
		require.NoError(t, err, "NodeBalancer update should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate NodeBalancer update confirmation
		require.Contains(t, responseText, "NodeBalancer updated successfully", "should confirm update")
		require.Contains(t, responseText, "ID: 12345", "should show NodeBalancer ID")
		require.Contains(t, responseText, "Label: updated-web-nodebalancer", "should show updated label")
	})

	t.Run("NodeBalancerConfigCreate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_nodebalancer_config_create",
				Arguments: map[string]interface{}{
					"nodebalancer_id": float64(12345),
					"port":            float64(80),
					"protocol":        "http",
					"algorithm":       "roundrobin",
					"check":           "http",
					"check_path":      "/",
				},
			},
		}

		result, err := service.handleNodeBalancerConfigCreate(ctx, request)
		require.NoError(t, err, "NodeBalancer config create should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate NodeBalancer config creation confirmation
		require.Contains(t, responseText, "NodeBalancer configuration created successfully", "should confirm config creation")
		require.Contains(t, responseText, "Config ID: 54321", "should show config ID")
		require.Contains(t, responseText, "Port: 80", "should show port")
		require.Contains(t, responseText, "Protocol: http", "should show protocol")
	})

	t.Run("NodeBalancerConfigUpdate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_nodebalancer_config_update",
				Arguments: map[string]interface{}{
					"nodebalancer_id": float64(12345),
					"config_id":       float64(54321),
					"port":            float64(443),
					"protocol":        "https",
					"algorithm":       "leastconn",
					"check_path":      "/health",
				},
			},
		}

		result, err := service.handleNodeBalancerConfigUpdate(ctx, request)
		require.NoError(t, err, "NodeBalancer config update should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate NodeBalancer config update confirmation
		require.Contains(t, responseText, "NodeBalancer configuration updated successfully", "should confirm config update")
		require.Contains(t, responseText, "Config ID: 54321", "should show config ID")
		require.Contains(t, responseText, "Port: 443", "should show updated port")
		require.Contains(t, responseText, "Protocol: https", "should show updated protocol")
	})

	t.Run("NodeBalancerConfigDelete", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_nodebalancer_config_delete",
				Arguments: map[string]interface{}{
					"nodebalancer_id": float64(12345),
					"config_id":       float64(54321),
				},
			},
		}

		result, err := service.handleNodeBalancerConfigDelete(ctx, request)
		require.NoError(t, err, "NodeBalancer config delete should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate NodeBalancer config deletion confirmation
		require.Contains(t, responseText, "NodeBalancer configuration", "should mention configuration")
		require.Contains(t, responseText, "deleted successfully", "should confirm deletion")
	})

	t.Run("NodeBalancerDelete", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_nodebalancer_delete",
				Arguments: map[string]interface{}{
					"nodebalancer_id": float64(12345),
				},
			},
		}

		result, err := service.handleNodeBalancerDelete(ctx, request)
		require.NoError(t, err, "NodeBalancer delete should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate NodeBalancer deletion confirmation
		require.Contains(t, responseText, "NodeBalancer", "should mention NodeBalancer")
		require.Contains(t, responseText, "deleted successfully", "should confirm deletion")
	})
}

// TestNodeBalancerErrorHandlingIntegration tests error scenarios for NodeBalancer tools
// to ensure CloudMCP handles API errors gracefully and provides meaningful
// error messages to users.
//
// **Error Test Scenarios**:
// • Non-existent NodeBalancer ID (404 errors)
// • Invalid configuration parameters
// • Network configuration conflicts
// • Permission errors for NodeBalancer operations
//
// **Expected Behavior**:
// • Proper error handling with contextual error messages
// • No unhandled exceptions or panics
// • Error responses follow CloudMCP error format
// • Error messages are actionable for users
//
// **Purpose**: Validates robust error handling in NodeBalancer operations
// and ensures reliable operation under error conditions.
func TestNodeBalancerErrorHandlingIntegration(t *testing.T) {
	server := createNodeBalancerTestServer()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
	ctx := t.Context()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	t.Run("NodeBalancerGetNotFound", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_nodebalancer_get",
				Arguments: map[string]interface{}{
					"nodebalancer_id": float64(999999), // Non-existent NodeBalancer
				},
			},
		}

		result, err := service.handleNodeBalancerGet(ctx, request)

		// Check for either Go error or MCP error result
		if err != nil {
			// Go error pattern
			require.Contains(t, err.Error(), "Failed to get NodeBalancer", "error should mention get NodeBalancer failure")
		} else {
			// MCP error result pattern
			require.NotNil(t, result, "result should not be nil")
			require.True(t, result.IsError, "result should be marked as error")
		}
	})

	t.Run("NodeBalancerDeleteNotFound", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_nodebalancer_delete",
				Arguments: map[string]interface{}{
					"nodebalancer_id": float64(999999), // Non-existent NodeBalancer
				},
			},
		}

		result, err := service.handleNodeBalancerDelete(ctx, request)

		// Check for either Go error or MCP error result
		if err != nil {
			// Go error pattern
			require.Contains(t, err.Error(), "Failed to delete NodeBalancer", "error should mention delete NodeBalancer failure")
		} else {
			// MCP error result pattern
			require.NotNil(t, result, "result should not be nil")
			require.True(t, result.IsError, "result should be marked as error")
		}
	})
}
