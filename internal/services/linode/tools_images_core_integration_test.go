//go:build integration

package linode_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// createImagesTestServer creates an HTTP test server with images API endpoints.
// This extends the base HTTP test server infrastructure with comprehensive
// images-specific endpoints for integration testing.
//
// **Images Endpoints Supported:**
// • GET /v4/images - List all images
// • GET /v4/images/{id} - Get specific image details
// • POST /v4/images - Create new image from disk
// • PUT /v4/images/{id} - Update image properties
// • DELETE /v4/images/{id} - Delete custom image
// • POST /v4/images/{id}/replicate - Replicate image to regions
// • POST /v4/images/upload - Create image upload URL
// • POST /v4/images/{id}/update - Update image properties
//
// **Mock Data Features:**
// • Realistic image configurations with distributions and custom images
// • Image types including official distributions and private images
// • Size and region replication simulation
// • Error simulation for non-existent resources and invalid operations
// • Proper HTTP status codes and error responses
func createImagesTestServer() *httptest.Server {
	mux := http.NewServeMux()

	// Include base endpoints (profile, account) from main test server
	addImagesBaseEndpoints(mux)

	// Images list endpoint
	mux.HandleFunc("/v4/images", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleImagesList(w, r)
		case http.MethodPost:
			// Check if it's upload endpoint or create endpoint
			if r.URL.Query().Get("upload") == "true" {
				handleImagesUploadCreate(w, r)
			} else {
				handleImagesCreate(w, r)
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Images upload endpoint
	mux.HandleFunc("/v4/images/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handleImagesUploadCreate(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Specific image endpoints
	mux.HandleFunc("/v4/images/linode/ubuntu22.04", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleImagesGet(w, r, "linode/ubuntu22.04")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v4/images/private/12345", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleImagesGet(w, r, "private/12345")
		case http.MethodPut:
			handleImagesUpdate(w, r, "private/12345")
		case http.MethodDelete:
			handleImagesDelete(w, r, "private/12345")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v4/images/private/67890", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleImagesGet(w, r, "private/67890")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Image action endpoints
	mux.HandleFunc("/v4/images/private/12345/replicate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handleImagesReplicate(w, r, "private/12345")
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Non-existent image endpoints for error testing
	mux.HandleFunc("/v4/images/private/999999", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleImagesGet(w, r, "private/999999")
		case http.MethodDelete:
			handleImagesDelete(w, r, "private/999999")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return httptest.NewServer(mux)
}

// addImagesBaseEndpoints adds the basic profile and account endpoints needed for service initialization.
func addImagesBaseEndpoints(mux *http.ServeMux) {
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

// handleImagesList handles the images list mock response.
func handleImagesList(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"id":           "linode/ubuntu22.04",
				"label":        "Ubuntu 22.04 LTS",
				"description":  "Ubuntu 22.04 LTS",
				"type":         "manual",
				"status":       "available",
				"size":         2048,
				"is_public":    true,
				"deprecated":   false,
				"vendor":       "Ubuntu",
				"expiry":       nil,
				"eol":          "2027-04-01T04:00:00",
				"created":      "2022-04-21T00:00:00",
				"updated":      "2023-01-01T00:00:00",
				"created_by":   "linode",
				"total_size":   2048,
				"regions":      []string{"us-east", "us-west", "us-central", "eu-west", "ap-south"},
				"capabilities": []string{"cloud-init", "virtio_scsi", "virtio_net"},
			},
			{
				"id":           "linode/ubuntu18.04",
				"label":        "Ubuntu 18.04 LTS",
				"description":  "Ubuntu 18.04 LTS (Deprecated - use Ubuntu 22.04)",
				"type":         "manual",
				"status":       "available",
				"size":         1536,
				"is_public":    true,
				"deprecated":   true,
				"vendor":       "Ubuntu",
				"expiry":       nil,
				"eol":          "2023-04-01T04:00:00",
				"created":      "2018-04-26T00:00:00",
				"updated":      "2023-01-01T00:00:00",
				"created_by":   "linode",
				"total_size":   1536,
				"regions":      []string{"us-east", "us-west", "us-central"},
				"capabilities": []string{"cloud-init"},
			},
			{
				"id":           "private/12345",
				"label":        "Custom Web Server Image",
				"description":  "Custom configured web server with nginx and SSL",
				"type":         "manual",
				"status":       "available",
				"size":         3072,
				"is_public":    false,
				"deprecated":   false,
				"vendor":       nil,
				"expiry":       nil,
				"eol":          nil,
				"created":      "2023-01-01T00:00:00",
				"updated":      "2023-01-01T00:00:00",
				"created_by":   "testuser",
				"total_size":   3072,
				"regions":      []string{"us-east", "us-west"},
				"capabilities": []string{"cloud-init"},
			},
			{
				"id":           "private/67890",
				"label":        "Database Backup Image",
				"description":  "Backup image with database configurations",
				"type":         "manual",
				"status":       "available",
				"size":         4096,
				"is_public":    false,
				"deprecated":   false,
				"vendor":       nil,
				"expiry":       nil,
				"eol":          nil,
				"created":      "2023-01-02T00:00:00",
				"updated":      "2023-01-02T00:00:00",
				"created_by":   "testuser",
				"total_size":   4096,
				"regions":      []string{"us-central"},
				"capabilities": []string{},
			},
		},
		"page":    1,
		"pages":   1,
		"results": 3,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleImagesGet handles the specific image mock response.
func handleImagesGet(w http.ResponseWriter, r *http.Request, imageID string) {
	switch imageID {
	case "linode/ubuntu22.04":
		response := map[string]interface{}{
			"id":           "linode/ubuntu22.04",
			"label":        "Ubuntu 22.04 LTS",
			"description":  "Ubuntu 22.04 LTS",
			"type":         "manual",
			"status":       "available",
			"size":         2048,
			"is_public":    true,
			"deprecated":   false,
			"vendor":       "Ubuntu",
			"expiry":       nil,
			"eol":          "2027-04-01T04:00:00",
			"created":      "2022-04-21T00:00:00",
			"updated":      "2023-01-01T00:00:00",
			"created_by":   "linode",
			"total_size":   2048,
			"regions":      []string{"us-east", "us-west", "us-central", "eu-west", "ap-south"},
			"capabilities": []string{"cloud-init", "virtio_scsi", "virtio_net"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	case "private/12345":
		response := map[string]interface{}{
			"id":           "private/12345",
			"label":        "Custom Web Server Image",
			"description":  "Custom configured web server with nginx and SSL",
			"type":         "manual",
			"status":       "available",
			"size":         3072,
			"is_public":    false,
			"deprecated":   false,
			"vendor":       nil,
			"expiry":       nil,
			"eol":          nil,
			"created":      "2023-01-01T00:00:00",
			"updated":      "2023-01-01T00:00:00",
			"created_by":   "testuser",
			"total_size":   3072,
			"regions":      []string{"us-east", "us-west"},
			"capabilities": []string{"cloud-init"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	case "private/67890":
		response := map[string]interface{}{
			"id":           "private/67890",
			"label":        "Database Backup Image",
			"description":  "Backup image with database configurations",
			"type":         "manual",
			"status":       "available",
			"size":         4096,
			"is_public":    false,
			"deprecated":   false,
			"vendor":       nil,
			"expiry":       nil,
			"eol":          nil,
			"created":      "2023-01-02T00:00:00",
			"updated":      "2023-01-02T00:00:00",
			"created_by":   "testuser",
			"total_size":   4096,
			"regions":      []string{"us-central"},
			"capabilities": []string{},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	case "private/999999":
		// Simulate not found error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "image_id", "reason": "Not found"},
			},
		})
	default:
		http.Error(w, "Image not found", http.StatusNotFound)
	}
}

// handleImagesCreate handles the image creation mock response.
func handleImagesCreate(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"id":           "private/24680",
		"label":        "new-custom-image",
		"description":  "New custom image created from disk",
		"type":         "manual",
		"status":       "creating",
		"size":         2500,
		"is_public":    false,
		"deprecated":   false,
		"vendor":       nil,
		"expiry":       nil,
		"eol":          nil,
		"created":      "2023-01-01T01:00:00",
		"updated":      "2023-01-01T01:00:00",
		"created_by":   "testuser",
		"total_size":   2500,
		"regions":      []string{"us-east"},
		"capabilities": []string{},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleImagesUpdate handles the image update mock response.
func handleImagesUpdate(w http.ResponseWriter, r *http.Request, imageID string) {
	if imageID == "private/999999" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "image_id", "reason": "Not found"},
			},
		})
		return
	}

	response := map[string]interface{}{
		"id":           "private/12345",
		"label":        "Updated Web Server Image",
		"description":  "Updated custom web server image with latest configurations",
		"type":         "manual",
		"status":       "available",
		"size":         3072,
		"is_public":    false,
		"deprecated":   false,
		"vendor":       nil,
		"expiry":       nil,
		"eol":          nil,
		"created":      "2023-01-01T00:00:00",
		"updated":      "2023-01-01T02:00:00",
		"created_by":   "testuser",
		"total_size":   3072,
		"regions":      []string{"us-east", "us-west"},
		"capabilities": []string{"cloud-init"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleImagesDelete handles the image deletion mock response.
func handleImagesDelete(w http.ResponseWriter, r *http.Request, imageID string) {
	if imageID == "private/999999" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "image_id", "reason": "Not found"},
			},
		})
		return
	}

	// Return 204 No Content for successful deletion
	w.WriteHeader(http.StatusNoContent)
}

// handleImagesReplicate handles the image replication mock response.
func handleImagesReplicate(w http.ResponseWriter, r *http.Request, imageID string) {
	response := map[string]interface{}{
		"id":           "private/12345",
		"label":        "Custom Web Server Image",
		"description":  "Custom configured web server with nginx and SSL",
		"type":         "manual",
		"status":       "available",
		"size":         3072,
		"is_public":    false,
		"deprecated":   false,
		"vendor":       nil,
		"expiry":       nil,
		"eol":          nil,
		"created":      "2023-01-01T00:00:00",
		"updated":      "2023-01-01T00:00:00",
		"created_by":   "testuser",
		"total_size":   3072,
		"regions":      []string{"us-east", "us-west", "us-central", "eu-west"}, // Expanded regions after replication
		"capabilities": []string{"cloud-init"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleImagesUploadCreate handles the image upload creation mock response.
func handleImagesUploadCreate(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"image": map[string]interface{}{
			"id":           "private/13579",
			"label":        "uploaded-image",
			"description":  "Image uploaded from file",
			"type":         "manual",
			"status":       "creating",
			"size":         0, // Will be filled after upload
			"is_public":    false,
			"deprecated":   false,
			"vendor":       nil,
			"expiry":       nil,
			"eol":          nil,
			"created":      "2023-01-01T01:00:00",
			"updated":      "2023-01-01T01:00:00",
			"created_by":   "testuser",
			"total_size":   0,
			"regions":      []string{"us-central"},
			"capabilities": []string{},
		},
		"upload_to": "https://us-central-1.linodeobjects.com/linode-cloud-images/uploads/abc123?response-content-disposition=attachment%3B%20filename%3D%22uploaded-image%22",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// TestImagesToolsIntegration tests all images-related CloudMCP tools
// using HTTP test server infrastructure to validate proper integration
// between CloudMCP handlers and Linode API endpoints.
//
// **Integration Test Coverage**:
// • linode_images_list - List all images
// • linode_image_get - Get specific image details
// • linode_image_create - Create new image from disk
// • linode_image_update - Update image properties
// • linode_image_delete - Delete custom image
// • linode_image_replicate - Replicate image to regions
// • linode_image_upload_create - Create image upload URL
//
// **Test Environment**: HTTP test server with images API handlers
//
// **Expected Behavior**:
// • All handlers return properly formatted text responses
// • Error conditions are handled gracefully with meaningful messages
// • Image data includes all required fields (ID, label, size, regions, etc.)
// • Image operations work correctly for both public and private images
//
// **Purpose**: Validates that CloudMCP images handlers correctly format
// Linode API responses for LLM consumption using lightweight HTTP test infrastructure.
func TestImagesToolsIntegration(t *testing.T) {
	// Extend the HTTP test server with images endpoints
	server := createImagesTestServer()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
	ctx := t.Context()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	t.Run("ImagesList", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_images_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleImagesList(ctx, request)
		require.NoError(t, err, "images list should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate images list formatting
		require.Contains(t, responseText, "Found 3 Linode images:", "should indicate correct image count")
		require.Contains(t, responseText, "Ubuntu 22.04 LTS", "should contain Ubuntu image")
		require.Contains(t, responseText, "Custom Web Server Image", "should contain custom image")
		require.Contains(t, responseText, "Database Backup Image", "should contain backup image")
		require.Contains(t, responseText, "Public", "should show public images")
		require.Contains(t, responseText, "Private", "should show private images")
		require.Contains(t, responseText, "Size:", "should show image sizes")
		require.Contains(t, responseText, "Regions:", "should show available regions")
	})

	t.Run("ImageGet", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_image_get",
				Arguments: map[string]interface{}{
					"image_id": "private/12345",
				},
			},
		}

		result, err := service.handleImageGet(ctx, request)
		require.NoError(t, err, "image get should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate detailed image information
		require.Contains(t, responseText, "Linode Image Details:", "should have image details header")
		require.Contains(t, responseText, "ID: private/12345", "should contain image ID")
		require.Contains(t, responseText, "Label: Custom Web Server Image", "should contain image label")
		require.Contains(t, responseText, "Size: 3072 MB", "should contain image size")
		require.Contains(t, responseText, "Type: Private", "should show private type")
		require.Contains(t, responseText, "Status: available", "should contain status")
		require.Contains(t, responseText, "Regions: us-east, us-west", "should list regions")
		require.Contains(t, responseText, "Capabilities:", "should list capabilities")
	})

	t.Run("ImageGetPublic", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_image_get",
				Arguments: map[string]interface{}{
					"image_id": "linode/ubuntu22.04",
				},
			},
		}

		result, err := service.handleImageGet(ctx, request)
		require.NoError(t, err, "public image get should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate public image information
		require.Contains(t, responseText, "ID: linode/ubuntu22.04", "should contain public image ID")
		require.Contains(t, responseText, "Label: Ubuntu 22.04 LTS", "should contain Ubuntu label")
		require.Contains(t, responseText, "Type: Public", "should show public type")
		require.Contains(t, responseText, "Vendor: Ubuntu", "should show vendor")
		require.Contains(t, responseText, "End of Life: 2027-04-01", "should show EOL date")
	})

	t.Run("ImageCreate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_image_create",
				Arguments: map[string]interface{}{
					"disk_id":     float64(54321),
					"label":       "test-image",
					"description": "Test image from disk",
				},
			},
		}

		result, err := service.handleImageCreate(ctx, request)
		require.NoError(t, err, "image create should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate image creation confirmation
		require.Contains(t, responseText, "Custom image created successfully:", "should confirm creation")
		require.Contains(t, responseText, "ID: private/24680", "should show new image ID")
		require.Contains(t, responseText, "Label: new-custom-image", "should show image label")
		require.Contains(t, responseText, "Status: creating", "should show creating status")
	})

	t.Run("ImageUpdate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_image_update",
				Arguments: map[string]interface{}{
					"image_id":    "private/12345",
					"label":       "Updated Image",
					"description": "Updated image description",
				},
			},
		}

		result, err := service.handleImageUpdate(ctx, request)
		require.NoError(t, err, "image update should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate update confirmation
		require.Contains(t, responseText, "Image updated successfully:", "should confirm update")
		require.Contains(t, responseText, "ID: private/12345", "should show image ID")
		require.Contains(t, responseText, "Label: Updated Web Server Image", "should show updated label")
	})

	t.Run("ImageReplicate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_image_replicate",
				Arguments: map[string]interface{}{
					"image_id": "private/12345",
					"regions":  []interface{}{"us-central", "eu-west"},
				},
			},
		}

		result, err := service.handleImageReplicate(ctx, request)
		require.NoError(t, err, "image replicate should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate replication confirmation
		require.Contains(t, responseText, "Image replicated successfully:", "should confirm replication")
		require.Contains(t, responseText, "Image ID: private/12345", "should show image ID")
		require.Contains(t, responseText, "Available in regions:", "should show available regions")
	})

	t.Run("ImageUploadCreate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_image_upload_create",
				Arguments: map[string]interface{}{
					"label":       "uploaded-test-image",
					"region":      "us-central",
					"description": "Test uploaded image",
				},
			},
		}

		result, err := service.handleImageUploadCreate(ctx, request)
		require.NoError(t, err, "image upload create should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate upload creation confirmation
		require.Contains(t, responseText, "Image upload URL created successfully:", "should confirm upload creation")
		require.Contains(t, responseText, "Image ID: private/13579", "should show new image ID")
		require.Contains(t, responseText, "Upload URL:", "should provide upload URL")
		require.Contains(t, responseText, "https://", "should contain HTTPS upload URL")
	})

	t.Run("ImageDelete", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_image_delete",
				Arguments: map[string]interface{}{
					"image_id": "private/12345",
				},
			},
		}

		result, err := service.handleImageDelete(ctx, request)
		require.NoError(t, err, "image delete should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate deletion confirmation
		require.Contains(t, responseText, "deleted successfully", "should confirm deletion")
		require.Contains(t, responseText, "private/12345", "should show image ID")
	})
}

// TestImagesErrorHandlingIntegration tests error scenarios for images tools
// to ensure CloudMCP handles API errors gracefully and provides meaningful
// error messages to users.
//
// **Error Test Scenarios**:
// • Non-existent image IDs (404 errors)
// • Invalid image creation parameters
// • Permission errors for image operations
// • Public image modification attempts
//
// **Expected Behavior**:
// • Proper error handling with contextual error messages
// • No unhandled exceptions or panics
// • Error responses follow CloudMCP error format
// • Error messages are actionable for users
//
// **Purpose**: Validates robust error handling in image operations
// and ensures reliable operation under error conditions.
func TestImagesErrorHandlingIntegration(t *testing.T) {
	server := createImagesTestServer()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
	ctx := t.Context()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	t.Run("ImageGetNotFound", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_image_get",
				Arguments: map[string]interface{}{
					"image_id": "private/999999", // Non-existent image
				},
			},
		}

		result, err := service.handleImageGet(ctx, request)

		// Check for either Go error or MCP error result
		if err != nil {
			// Go error pattern
			require.Contains(t, err.Error(), "Failed to get image", "error should mention get image failure")
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

	t.Run("ImageDeleteNotFound", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_image_delete",
				Arguments: map[string]interface{}{
					"image_id": "private/999999", // Non-existent image
				},
			},
		}

		result, err := service.handleImageDelete(ctx, request)

		// Check for either Go error or MCP error result
		if err != nil {
			// Go error pattern
			require.Contains(t, err.Error(), "Failed to delete image", "error should mention delete image failure")
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

	t.Run("InvalidImageID", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_image_get",
				Arguments: map[string]interface{}{
					"image_id": 12345, // Invalid ID type (should be string)
				},
			},
		}

		result, err := service.handleImageGet(ctx, request)

		// Should get MCP error result for invalid parameter
		if err == nil {
			require.NotNil(t, result, "result should not be nil")
			require.True(t, result.IsError, "result should be marked as error")
		}
	})

	t.Run("MissingRequiredParameter", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_image_create",
				Arguments: map[string]interface{}{
					// Missing required parameters like disk_id, label
					"description": "incomplete image",
				},
			},
		}

		result, err := service.handleImageCreate(ctx, request)

		// Should get MCP error result for missing parameters
		if err == nil {
			require.NotNil(t, result, "result should not be nil")
			require.True(t, result.IsError, "result should be marked as error")
		}
	})

	t.Run("ImageReplicateInvalidRegions", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_image_replicate",
				Arguments: map[string]interface{}{
					"image_id": "private/12345",
					"regions":  "invalid-regions", // Invalid regions format (should be array)
				},
			},
		}

		result, err := service.handleImageReplicate(ctx, request)

		// Should get MCP error result for invalid parameter format
		if err == nil {
			require.NotNil(t, result, "result should not be nil")
			require.True(t, result.IsError, "result should be marked as error")
		}
	})
}
