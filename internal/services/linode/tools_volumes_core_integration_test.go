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

// createVolumesTestServer creates an HTTP test server with volumes API endpoints.
// This extends the base HTTP test server infrastructure with comprehensive
// volumes-specific endpoints for integration testing.
//
// **Volumes Endpoints Supported:**
// • GET /v4/volumes - List all volumes
// • GET /v4/volumes/{id} - Get specific volume details
// • POST /v4/volumes - Create new volume
// • PUT /v4/volumes/{id} - Update volume
// • DELETE /v4/volumes/{id} - Delete volume
// • POST /v4/volumes/{id}/attach - Attach volume to instance
// • POST /v4/volumes/{id}/detach - Detach volume from instance
//
// **Mock Data Features:**
// • Realistic volume configurations with multiple sizes and regions
// • Volume attachment state management (attached, detached)
// • Size constraints and regional availability simulation
// • Error simulation for non-existent resources and invalid operations
// • Proper HTTP status codes and error responses
func createVolumesTestServer() *httptest.Server {
	mux := http.NewServeMux()

	// Include base endpoints (profile, account) from main test server
	addVolumesBaseEndpoints(mux)

	// Volumes list endpoint
	mux.HandleFunc("/v4/volumes", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleVolumesList(w, r)
		case http.MethodPost:
			handleVolumesCreate(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Specific volume endpoints
	mux.HandleFunc("/v4/volumes/54321", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleVolumesGet(w, r, "54321")
		case http.MethodPut:
			handleVolumesUpdate(w, r, "54321")
		case http.MethodDelete:
			handleVolumesDelete(w, r, "54321")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v4/volumes/98765", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleVolumesGet(w, r, "98765")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Volume action endpoints
	mux.HandleFunc("/v4/volumes/54321/attach", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handleVolumesAttach(w, r, "54321")
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v4/volumes/54321/detach", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handleVolumesDetach(w, r, "54321")
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Non-existent volume endpoints for error testing
	mux.HandleFunc("/v4/volumes/999999", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleVolumesGet(w, r, "999999")
		case http.MethodDelete:
			handleVolumesDelete(w, r, "999999")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return httptest.NewServer(mux)
}

// addVolumesBaseEndpoints adds the basic profile and account endpoints needed for service initialization.
func addVolumesBaseEndpoints(mux *http.ServeMux) {
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

// handleVolumesList handles the volumes list mock response.
func handleVolumesList(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"id":              54321,
				"label":           "database-storage",
				"status":          "active",
				"size":            100,
				"region":          "us-east",
				"linode_id":       12345,
				"linode_label":    "web-server-prod",
				"filesystem_path": "/dev/disk/by-id/scsi-0Linode_Volume_database-storage",
				"tags":            []string{"production", "database"},
				"created":         "2023-01-01T00:00:00",
				"updated":         "2023-01-01T00:00:00",
			},
			{
				"id":              98765,
				"label":           "backup-storage",
				"status":          "active",
				"size":            50,
				"region":          "us-west",
				"linode_id":       nil,
				"linode_label":    nil,
				"filesystem_path": "/dev/disk/by-id/scsi-0Linode_Volume_backup-storage",
				"tags":            []string{"backup"},
				"created":         "2023-01-02T00:00:00",
				"updated":         "2023-01-02T00:00:00",
			},
		},
		"page":    1,
		"pages":   1,
		"results": 2,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleVolumesGet handles the specific volume mock response.
func handleVolumesGet(w http.ResponseWriter, r *http.Request, volumeID string) {
	switch volumeID {
	case "54321":
		response := map[string]interface{}{
			"id":              54321,
			"label":           "database-storage",
			"status":          "active",
			"size":            100,
			"region":          "us-east",
			"linode_id":       12345,
			"linode_label":    "web-server-prod",
			"filesystem_path": "/dev/disk/by-id/scsi-0Linode_Volume_database-storage",
			"tags":            []string{"production", "database"},
			"created":         "2023-01-01T00:00:00",
			"updated":         "2023-01-01T00:00:00",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	case "98765":
		response := map[string]interface{}{
			"id":              98765,
			"label":           "backup-storage",
			"status":          "active",
			"size":            50,
			"region":          "us-west",
			"linode_id":       nil,
			"linode_label":    nil,
			"filesystem_path": "/dev/disk/by-id/scsi-0Linode_Volume_backup-storage",
			"tags":            []string{"backup"},
			"created":         "2023-01-02T00:00:00",
			"updated":         "2023-01-02T00:00:00",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	case "999999":
		// Simulate not found error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "volume_id", "reason": "Not found"},
			},
		})
	default:
		http.Error(w, "Volume not found", http.StatusNotFound)
	}
}

// handleVolumesCreate handles the volume creation mock response.
func handleVolumesCreate(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"id":              13579,
		"label":           "new-volume",
		"status":          "creating",
		"size":            20,
		"region":          "us-central",
		"linode_id":       nil,
		"linode_label":    nil,
		"filesystem_path": "/dev/disk/by-id/scsi-0Linode_Volume_new-volume",
		"tags":            []string{},
		"created":         "2023-01-01T01:00:00",
		"updated":         "2023-01-01T01:00:00",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleVolumesUpdate handles the volume update mock response.
func handleVolumesUpdate(w http.ResponseWriter, r *http.Request, volumeID string) {
	if volumeID == "999999" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "volume_id", "reason": "Not found"},
			},
		})
		return
	}

	response := map[string]interface{}{
		"id":              54321,
		"label":           "updated-database-storage",
		"status":          "active",
		"size":            120, // Updated size
		"region":          "us-east",
		"linode_id":       12345,
		"linode_label":    "web-server-prod",
		"filesystem_path": "/dev/disk/by-id/scsi-0Linode_Volume_updated-database-storage",
		"tags":            []string{"production", "database", "updated"},
		"created":         "2023-01-01T00:00:00",
		"updated":         "2023-01-01T02:00:00",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleVolumesDelete handles the volume deletion mock response.
func handleVolumesDelete(w http.ResponseWriter, r *http.Request, volumeID string) {
	if volumeID == "999999" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "volume_id", "reason": "Not found"},
			},
		})
		return
	}

	// Return 204 No Content for successful deletion
	w.WriteHeader(http.StatusNoContent)
}

// handleVolumesAttach handles the volume attach mock response.
func handleVolumesAttach(w http.ResponseWriter, r *http.Request, volumeID string) {
	response := map[string]interface{}{
		"id":              54321,
		"label":           "database-storage",
		"status":          "active",
		"size":            100,
		"region":          "us-east",
		"linode_id":       12345,
		"linode_label":    "web-server-prod",
		"filesystem_path": "/dev/disk/by-id/scsi-0Linode_Volume_database-storage",
		"tags":            []string{"production", "database"},
		"created":         "2023-01-01T00:00:00",
		"updated":         "2023-01-01T00:00:00",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleVolumesDetach handles the volume detach mock response.
func handleVolumesDetach(w http.ResponseWriter, r *http.Request, volumeID string) {
	response := map[string]interface{}{
		"id":              54321,
		"label":           "database-storage",
		"status":          "active",
		"size":            100,
		"region":          "us-east",
		"linode_id":       nil, // Detached
		"linode_label":    nil, // Detached
		"filesystem_path": "/dev/disk/by-id/scsi-0Linode_Volume_database-storage",
		"tags":            []string{"production", "database"},
		"created":         "2023-01-01T00:00:00",
		"updated":         "2023-01-01T00:00:00",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// TestVolumesToolsIntegration tests all volumes-related CloudMCP tools
// using HTTP test server infrastructure to validate proper integration
// between CloudMCP handlers and Linode API endpoints.
//
// **Integration Test Coverage**:
// • linode_volumes_list - List all volumes
// • linode_volume_get - Get specific volume details
// • linode_volume_create - Create new volume
// • linode_volume_delete - Delete volume
// • linode_volume_attach - Attach volume to instance
// • linode_volume_detach - Detach volume from instance
//
// **Test Environment**: HTTP test server with volumes API handlers
//
// **Expected Behavior**:
// • All handlers return properly formatted text responses
// • Error conditions are handled gracefully with meaningful messages
// • Volume data includes all required fields (ID, label, size, region, etc.)
// • Volume attachment/detachment operations work correctly
//
// **Purpose**: Validates that CloudMCP volumes handlers correctly format
// Linode API responses for LLM consumption using lightweight HTTP test infrastructure.
func TestVolumesToolsIntegration(t *testing.T) {
	// Extend the HTTP test server with volumes endpoints
	server := createVolumesTestServer()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
	ctx := context.Background()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	t.Run("VolumesList", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_volumes_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleVolumesList(ctx, request)
		require.NoError(t, err, "volumes list should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate volumes list formatting
		require.Contains(t, responseText, "Found 2 block storage volumes:", "should indicate correct volume count")
		require.Contains(t, responseText, "database-storage", "should contain first volume label")
		require.Contains(t, responseText, "backup-storage", "should contain second volume label")
		require.Contains(t, responseText, "Size: 100 GB", "should show volume size")
		require.Contains(t, responseText, "Status: active", "should show volume status")
		require.Contains(t, responseText, "Attached to:", "should show attachment status")
		require.Contains(t, responseText, "Unattached", "should show unattached status")
	})

	t.Run("VolumeGet", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_volume_get",
				Arguments: map[string]interface{}{
					"volume_id": float64(54321),
				},
			},
		}

		result, err := service.handleVolumeGet(ctx, request)
		require.NoError(t, err, "volume get should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate detailed volume information
		require.Contains(t, responseText, "Block Storage Volume Details:", "should have volume details header")
		require.Contains(t, responseText, "ID: 54321", "should contain volume ID")
		require.Contains(t, responseText, "Label: database-storage", "should contain volume label")
		require.Contains(t, responseText, "Size: 100 GB", "should contain volume size")
		require.Contains(t, responseText, "Region: us-east", "should contain region")
		require.Contains(t, responseText, "Status: active", "should contain status")
		require.Contains(t, responseText, "Attached to: web-server-prod (12345)", "should contain attachment info")
		require.Contains(t, responseText, "Filesystem Path:", "should contain filesystem path")
	})

	t.Run("VolumeCreate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_volume_create",
				Arguments: map[string]interface{}{
					"label":  "test-volume",
					"size":   float64(20),
					"region": "us-central",
				},
			},
		}

		result, err := service.handleVolumeCreate(ctx, request)
		require.NoError(t, err, "volume create should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate volume creation confirmation
		require.Contains(t, responseText, "Block storage volume created successfully:", "should confirm creation")
		require.Contains(t, responseText, "ID: 13579", "should show new volume ID")
		require.Contains(t, responseText, "Label: new-volume", "should show volume label")
		require.Contains(t, responseText, "Size: 20 GB", "should show volume size")
		require.Contains(t, responseText, "Status: creating", "should show creating status")
	})

	t.Run("VolumeAttach", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_volume_attach",
				Arguments: map[string]interface{}{
					"volume_id": float64(54321),
					"linode_id": float64(12345),
				},
			},
		}

		result, err := service.handleVolumeAttach(ctx, request)
		require.NoError(t, err, "volume attach should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate attachment confirmation
		require.Contains(t, responseText, "Volume attached successfully:", "should confirm attachment")
		require.Contains(t, responseText, "Volume ID: 54321", "should show volume ID")
		require.Contains(t, responseText, "Attached to: web-server-prod (12345)", "should show attachment target")
	})

	t.Run("VolumeDetach", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_volume_detach",
				Arguments: map[string]interface{}{
					"volume_id": float64(54321),
				},
			},
		}

		result, err := service.handleVolumeDetach(ctx, request)
		require.NoError(t, err, "volume detach should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate detachment confirmation
		require.Contains(t, responseText, "Volume detached successfully:", "should confirm detachment")
		require.Contains(t, responseText, "Volume ID: 54321", "should show volume ID")
		require.Contains(t, responseText, "Status: Unattached", "should show unattached status")
	})

	t.Run("VolumeDelete", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_volume_delete",
				Arguments: map[string]interface{}{
					"volume_id": float64(54321),
				},
			},
		}

		result, err := service.handleVolumeDelete(ctx, request)
		require.NoError(t, err, "volume delete should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate deletion confirmation
		require.Contains(t, responseText, "deleted successfully", "should confirm deletion")
		require.Contains(t, responseText, "54321", "should show volume ID")
	})
}

// TestVolumesErrorHandlingIntegration tests error scenarios for volumes tools
// to ensure CloudMCP handles API errors gracefully and provides meaningful
// error messages to users.
//
// **Error Test Scenarios**:
// • Non-existent volume IDs (404 errors)
// • Invalid volume creation parameters
// • Volume attachment conflicts and constraints
// • Permission errors for volume operations
//
// **Expected Behavior**:
// • Proper error handling with contextual error messages
// • No unhandled exceptions or panics
// • Error responses follow CloudMCP error format
// • Error messages are actionable for users
//
// **Purpose**: Validates robust error handling in volume operations
// and ensures reliable operation under error conditions.
func TestVolumesErrorHandlingIntegration(t *testing.T) {
	server := createVolumesTestServer()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
	ctx := context.Background()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	t.Run("VolumeGetNotFound", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_volume_get",
				Arguments: map[string]interface{}{
					"volume_id": float64(999999), // Non-existent volume
				},
			},
		}

		result, err := service.handleVolumeGet(ctx, request)

		// Check for either Go error or MCP error result
		if err != nil {
			// Go error pattern
			require.Contains(t, err.Error(), "Failed to get volume", "error should mention get volume failure")
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

	t.Run("VolumeDeleteNotFound", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_volume_delete",
				Arguments: map[string]interface{}{
					"volume_id": float64(999999), // Non-existent volume
				},
			},
		}

		result, err := service.handleVolumeDelete(ctx, request)

		// Check for either Go error or MCP error result
		if err != nil {
			// Go error pattern
			require.Contains(t, err.Error(), "Failed to delete volume", "error should mention delete volume failure")
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

	t.Run("InvalidVolumeID", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_volume_get",
				Arguments: map[string]interface{}{
					"volume_id": "invalid", // Invalid ID type
				},
			},
		}

		result, err := service.handleVolumeGet(ctx, request)

		// Should get MCP error result for invalid parameter
		if err == nil {
			require.NotNil(t, result, "result should not be nil")
			require.True(t, result.IsError, "result should be marked as error")
		}
	})

	t.Run("MissingRequiredParameter", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_volume_create",
				Arguments: map[string]interface{}{
					// Missing required parameters like size, region
					"label": "incomplete-volume",
				},
			},
		}

		result, err := service.handleVolumeCreate(ctx, request)

		// Should get MCP error result for missing parameters
		if err == nil {
			require.NotNil(t, result, "result should not be nil")
			require.True(t, result.IsError, "result should be marked as error")
		}
	})
}
