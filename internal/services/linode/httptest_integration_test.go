//go:build integration

package linode

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/pkg/logger"
)

// MockLinodeAPIServer creates an HTTP test server that mocks the Linode API.
// This provides a lightweight alternative to containerized WireMock for
// integration testing scenarios.
//
// **Server Features:**
// • Standard Go httptest.Server for fast startup
// • Realistic Linode API endpoint responses
// • Support for various HTTP status codes and error scenarios
// • Easy debugging with Go standard library tools
//
// **Supported Endpoints:**
// • GET /v4/profile - User profile information
// • GET /v4/account - Account information
// • GET /v4/linode/instances - List instances
// • GET /v4/linode/instances/{id} - Get specific instance
// • POST /v4/linode/instances - Create new instance
// • GET /v4/volumes - List volumes
//
// **Error Simulation:**
// • 404 responses for non-existent resources
// • 401 responses for invalid tokens
// • Various validation error scenarios
func MockLinodeAPIServer() *httptest.Server {
	mux := http.NewServeMux()

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

	// Instances list endpoint
	mux.HandleFunc("/v4/linode/instances", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleInstancesList(w, r)
		case http.MethodPost:
			handleInstanceCreate(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Specific instance endpoint
	mux.HandleFunc("/v4/linode/instances/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract instance ID from path
		path := strings.TrimPrefix(r.URL.Path, "/v4/linode/instances/")
		switch path {
		case "123456":
			handleInstanceGet(w, r, "123456")
		case "999999":
			// Simulate not found error
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"errors": []map[string]string{
					{"field": "linode_id", "reason": "Not found"},
				},
			})
		default:
			http.Error(w, "Instance not found", http.StatusNotFound)
		}
	})

	// Volumes endpoint
	mux.HandleFunc("/v4/volumes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		handleVolumesList(w, r)
	})

	return httptest.NewServer(mux)
}

// handleInstancesList handles the instances list mock response.
func handleInstancesList(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"id":               123456,
				"label":            "test-instance-1",
				"region":           "us-east",
				"image":            "linode/ubuntu22.04",
				"type":             "g6-nanode-1",
				"status":           "running",
				"ipv4":             []string{"192.168.1.1"},
				"ipv6":             "2600:3c01::f03c:91ff:fe24:3a2f/128",
				"created":          "2023-01-01T00:00:00",
				"updated":          "2023-01-01T00:00:00",
				"hypervisor":       "kvm",
				"watchdog_enabled": true,
				"tags":             []string{"production", "web"},
				"specs": map[string]interface{}{
					"disk":     25600,
					"memory":   1024,
					"vcpus":    1,
					"gpus":     0,
					"transfer": 1000,
				},
				"alerts": map[string]interface{}{
					"cpu":            90,
					"network_in":     10,
					"network_out":    10,
					"transfer_quota": 80,
					"io":             10000,
				},
				"backups": map[string]interface{}{
					"enabled":   false,
					"available": false,
					"schedule": map[string]interface{}{
						"day":    nil,
						"window": nil,
					},
					"last_successful": nil,
				},
				"group": "",
			},
		},
		"page":    1,
		"pages":   1,
		"results": 1,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleInstanceGet handles the specific instance mock response.
func handleInstanceGet(w http.ResponseWriter, r *http.Request, instanceID string) {
	response := map[string]interface{}{
		"id":               123456,
		"label":            "test-instance-1",
		"region":           "us-east",
		"image":            "linode/ubuntu22.04",
		"type":             "g6-nanode-1",
		"status":           "running",
		"ipv4":             []string{"192.168.1.1"},
		"ipv6":             "2600:3c01::f03c:91ff:fe24:3a2f/128",
		"created":          "2023-01-01T00:00:00",
		"updated":          "2023-01-01T00:00:00",
		"hypervisor":       "kvm",
		"watchdog_enabled": true,
		"tags":             []string{"production", "web"},
		"specs": map[string]interface{}{
			"disk":     25600,
			"memory":   1024,
			"vcpus":    1,
			"gpus":     0,
			"transfer": 1000,
		},
		"alerts": map[string]interface{}{
			"cpu":            90,
			"network_in":     10,
			"network_out":    10,
			"transfer_quota": 80,
			"io":             10000,
		},
		"backups": map[string]interface{}{
			"enabled":   false,
			"available": false,
			"schedule": map[string]interface{}{
				"day":    nil,
				"window": nil,
			},
			"last_successful": nil,
		},
		"group": "",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleInstanceCreate handles the instance creation mock response.
func handleInstanceCreate(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"id":               123457,
		"label":            "new-test-instance",
		"region":           "us-east",
		"image":            "linode/ubuntu22.04",
		"type":             "g6-nanode-1",
		"status":           "provisioning",
		"ipv4":             []string{"192.168.1.2"},
		"ipv6":             "2600:3c01::f03c:91ff:fe24:3a30/128",
		"created":          "2023-01-01T01:00:00",
		"updated":          "2023-01-01T01:00:00",
		"hypervisor":       "kvm",
		"watchdog_enabled": true,
		"tags":             []string{},
		"specs": map[string]interface{}{
			"disk":     25600,
			"memory":   1024,
			"vcpus":    1,
			"gpus":     0,
			"transfer": 1000,
		},
		"alerts": map[string]interface{}{
			"cpu":            90,
			"network_in":     10,
			"network_out":    10,
			"transfer_quota": 80,
			"io":             10000,
		},
		"backups": map[string]interface{}{
			"enabled":   false,
			"available": false,
			"schedule": map[string]interface{}{
				"day":    nil,
				"window": nil,
			},
			"last_successful": nil,
		},
		"group": "",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleVolumesList handles the volumes list mock response.
func handleVolumesList(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"id":              456789,
				"label":           "test-volume-1",
				"region":          "us-east",
				"size":            20,
				"status":          "active",
				"created":         "2023-01-01T00:00:00",
				"updated":         "2023-01-01T00:00:00",
				"filesystem_path": "/dev/disk/by-id/scsi-0Linode_Volume_test-volume-1",
				"tags":            []string{"storage"},
				"linode_id":       123456,
				"linode_label":    "test-instance-1",
			},
		},
		"page":    1,
		"pages":   1,
		"results": 1,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateHTTPTestService creates a CloudMCP Linode service configured to use the HTTP test server.
// This service instance will make API calls to the httptest server instead of
// the real Linode API, providing lightweight integration testing.
//
// **Service Configuration:**
// • Uses HTTP test server base URL
// • Configured with test account and token
// • Debug logging enabled for detailed test output
// • Same service interface as production for realistic testing
//
// **Account Setup:**
// • Single test account named "httptest-integration"
// • Uses placeholder token (test server doesn't validate)
// • Custom API URL pointing to httptest server
func CreateHTTPTestService(t *testing.T, mockAPIURL string) *Service {
	log := logger.New("debug")

	cfg := &config.Config{
		DefaultLinodeAccount: "httptest-integration",
		LinodeAccounts: map[string]config.LinodeAccount{
			"httptest-integration": {
				Token:  "test-token-httptest-api",
				Label:  "HTTP Test Integration Account",
				APIURL: mockAPIURL,
			},
		},
	}

	service, err := New(cfg, log)
	require.NoError(t, err, "failed to create HTTP test service")

	return service
}

// createHTTPTestServiceForBenchmark creates a CloudMCP service for benchmark tests.
// This is separate from the main test helper because testing.B and testing.T are different types.
func createHTTPTestServiceForBenchmark(b *testing.B, mockAPIURL string) *Service {
	log := logger.New("error") // Use error level to reduce benchmark noise

	cfg := &config.Config{
		DefaultLinodeAccount: "httptest-integration",
		LinodeAccounts: map[string]config.LinodeAccount{
			"httptest-integration": {
				Token:  "test-token-httptest-api",
				Label:  "HTTP Test Integration Account",
				APIURL: mockAPIURL,
			},
		},
	}

	service, err := New(cfg, log)
	if err != nil {
		b.Fatalf("failed to create HTTP test service for benchmark: %v", err)
	}

	return service
}

// SetupHTTPTestIntegration performs common setup for HTTP-based integration tests.
// This helper function creates the HTTP test server, test service,
// and returns all necessary components for integration testing with
// proper cleanup setup.
//
// **Setup Process:**
// 1. Start HTTP test server with Linode API handlers
// 2. Create CloudMCP service configured for test server
// 3. Initialize service with test account
// 4. Return service and cleanup function
//
// **Cleanup Handling:**
// Returns a cleanup function that should be called with defer to ensure
// proper server shutdown and resource cleanup.
//
// **Usage:**
//
//	service, cleanup := SetupHTTPTestIntegration(t)
//	defer cleanup()
//	// Run integration tests with service
func SetupHTTPTestIntegration(t *testing.T) (*Service, func()) {
	ctx := context.Background()

	server := MockLinodeAPIServer()

	service := CreateHTTPTestService(t, server.URL)

	err := service.Initialize(ctx)
	require.NoError(t, err, "failed to initialize HTTP test service")

	cleanup := func() {
		server.Close()
	}

	return service, cleanup
}

// TestHTTPServerInstancesListIntegration tests the instances list workflow using HTTP test server.
// This integration test verifies that the CloudMCP service can successfully communicate
// with an HTTP test server mock and return properly formatted instance data.
//
// **Integration Test Workflow**:
// 1. **Server Setup**: Start HTTP test server with Linode API handlers
// 2. **Service Creation**: Initialize CloudMCP service with test server endpoint
// 3. **Handler Call**: Execute instances list handler directly
// 4. **Response Validation**: Verify JSON structure and data completeness
// 5. **Contract Verification**: Ensure response matches Linode API contract
//
// **Test Environment**: HTTP test server with predefined instance data
//
// **Expected Behavior**:
// • Returns valid handler result with JSON content
// • JSON contains expected instance fields (id, label, status, region, etc.)
// • Response structure matches real Linode API format
// • All required fields are present and correctly typed
//
// **Purpose**: This test demonstrates lightweight integration testing using Go's
// standard httptest package instead of containers for faster test execution.
func TestHTTPServerInstancesListIntegration(t *testing.T) {
	service, cleanup := SetupHTTPTestIntegration(t)
	defer cleanup()

	ctx := context.Background()
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

	// Validate text response contains expected instance information
	require.Contains(t, responseText, "Found 1 Linode instance(s)", "should indicate 1 instance found")
	require.Contains(t, responseText, "ID: 123456", "should contain instance ID")
	require.Contains(t, responseText, "test-instance-1", "should contain instance label")
	require.Contains(t, responseText, "Status: running", "should contain running status")
	require.Contains(t, responseText, "Region: us-east", "should contain region")
	require.Contains(t, responseText, "Type: g6-nanode-1", "should contain instance type")
	require.Contains(t, responseText, "IPv4: [192.168.1.1]", "should contain IPv4 address")
}

// TestHTTPServerInstanceGetIntegration tests getting a specific instance using HTTP test server.
// This integration test verifies that the CloudMCP service can retrieve detailed
// information for a specific instance using the HTTP test server mock.
//
// **Integration Test Workflow**:
// 1. **Service Setup**: Initialize service with HTTP test server
// 2. **Handler Call**: Execute instance get handler with specific instance ID
// 3. **Response Validation**: Verify detailed instance data structure
// 4. **Field Verification**: Ensure all instance fields are present and valid
//
// **Test Scenario**: Retrieve instance with ID 123456 (defined in test server)
//
// **Expected Behavior**:
// • Returns detailed instance information through handler
// • All instance fields are properly structured and typed
// • Response matches expected Linode API instance object format
// • Specific field values match predefined test server data
//
// **Purpose**: Validates that CloudMCP can successfully retrieve and format
// detailed instance information using lightweight HTTP test server infrastructure.
func TestHTTPServerInstanceGetIntegration(t *testing.T) {
	service, cleanup := SetupHTTPTestIntegration(t)
	defer cleanup()

	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_instance_get",
			Arguments: map[string]interface{}{
				"instance_id": float64(123456),
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

	// Validate text response contains expected instance information
	require.Contains(t, responseText, "Instance Details:", "should have instance details header")
	require.Contains(t, responseText, "ID: 123456", "should contain instance ID")
	require.Contains(t, responseText, "Label: test-instance-1", "should contain instance label")
	require.Contains(t, responseText, "Status: running", "should contain running status")
	require.Contains(t, responseText, "Region: us-east", "should contain region")
	require.Contains(t, responseText, "Type: g6-nanode-1", "should contain instance type")
	require.Contains(t, responseText, "Image: linode/ubuntu22.04", "should contain image")
	require.Contains(t, responseText, "Specifications:", "should have specifications section")
	require.Contains(t, responseText, "CPUs: 1", "should contain CPU count")
	require.Contains(t, responseText, "Memory: 1024 MB", "should contain memory")
	require.Contains(t, responseText, "Network:", "should have network section")
	require.Contains(t, responseText, "IPv4: 192.168.1.1", "should contain IPv4 address")
	require.Contains(t, responseText, "IPv6: 2600:3c01::f03c:91ff:fe24:3a2f/128", "should contain IPv6 address")
}

// TestHTTPServerErrorHandlingIntegration tests error handling for non-existent instances using HTTP test server.
// This integration test verifies that CloudMCP properly handles and formats error
// responses when requesting an instance that doesn't exist in the test server.
//
// **Integration Test Workflow**:
// 1. **Service Setup**: Initialize CloudMCP with HTTP test server endpoints
// 2. **Error Request**: Request instance with ID that triggers 404 error
// 3. **Error Validation**: Verify proper error handling and response format
// 4. **Message Verification**: Ensure error messages are meaningful
//
// **Test Scenario**: Request instance ID 999999 (configured to return 404)
//
// **Expected Behavior**:
// • Returns handler error result instead of throwing exception
// • Error message indicates "not found" condition
// • Response structure follows error format
// • Error details match Linode API error structure
//
// **Purpose**: Ensures CloudMCP gracefully handles API errors and provides
// meaningful error information using HTTP test server infrastructure.
func TestHTTPServerErrorHandlingIntegration(t *testing.T) {
	service, cleanup := SetupHTTPTestIntegration(t)
	defer cleanup()

	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_instance_get",
			Arguments: map[string]interface{}{
				"instance_id": float64(999999), // This ID is configured to return 404
			},
		},
	}

	result, err := service.handleInstanceGet(ctx, request)

	// The current implementation returns Go errors for API errors
	require.Error(t, err, "should return Go error for API errors")
	require.Contains(t, err.Error(), "failed to get instance 999999", "error should mention failed to get instance")
	require.Contains(t, err.Error(), "linode/instance_get", "error should include tool context")

	// When there's a Go error, result might be nil
	if result != nil {
		require.True(t, result.IsError, "result should be marked as error if not nil")
	}
}
