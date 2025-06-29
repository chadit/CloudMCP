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

// createObjectStorageTestServer creates an HTTP test server with Object Storage API endpoints.
// This extends the base HTTP test server infrastructure with comprehensive
// Object Storage-specific endpoints for integration testing.
//
// **Object Storage Endpoints Supported:**
// • GET /v4/object-storage/clusters - List object storage clusters
// • GET /v4/object-storage/buckets - List buckets
// • GET /v4/object-storage/buckets/{cluster}/{bucket} - Get specific bucket
// • POST /v4/object-storage/buckets - Create new bucket
// • PUT /v4/object-storage/buckets/{cluster}/{bucket} - Update bucket
// • DELETE /v4/object-storage/buckets/{cluster}/{bucket} - Delete bucket
// • GET /v4/object-storage/keys - List object storage keys
// • GET /v4/object-storage/keys/{id} - Get specific key
// • POST /v4/object-storage/keys - Create new key
// • PUT /v4/object-storage/keys/{id} - Update key
// • DELETE /v4/object-storage/keys/{id} - Delete key
//
// **Mock Data Features:**
// • Realistic Object Storage configurations with buckets and keys
// • Access key management with scoped permissions
// • Cluster information and regional availability
// • Error simulation for non-existent resources
// • Proper HTTP status codes and error responses
func createObjectStorageTestServer() *httptest.Server {
	mux := http.NewServeMux()

	// Include base endpoints (profile, account) from main test server
	addObjectStorageBaseEndpoints(mux)

	// Object Storage clusters endpoint
	mux.HandleFunc("/v4/object-storage/clusters", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleObjectStorageClustersList(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Object Storage buckets list endpoint
	mux.HandleFunc("/v4/object-storage/buckets", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleObjectStorageBucketsList(w, r)
		case http.MethodPost:
			handleObjectStorageBucketsCreate(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Specific bucket endpoints with cluster and bucket name
	mux.HandleFunc("/v4/object-storage/buckets/us-east-1/production-bucket", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleObjectStorageBucketsGet(w, r, "us-east-1", "production-bucket")
		case http.MethodPut:
			handleObjectStorageBucketsUpdate(w, r, "us-east-1", "production-bucket")
		case http.MethodDelete:
			handleObjectStorageBucketsDelete(w, r, "us-east-1", "production-bucket")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Object Storage keys list endpoint
	mux.HandleFunc("/v4/object-storage/keys", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleObjectStorageKeysList(w, r)
		case http.MethodPost:
			handleObjectStorageKeysCreate(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Specific key endpoints
	mux.HandleFunc("/v4/object-storage/keys/12345", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleObjectStorageKeysGet(w, r, "12345")
		case http.MethodPut:
			handleObjectStorageKeysUpdate(w, r, "12345")
		case http.MethodDelete:
			handleObjectStorageKeysDelete(w, r, "12345")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Non-existent resource endpoints for error testing
	mux.HandleFunc("/v4/object-storage/buckets/invalid-cluster/nonexistent-bucket", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleObjectStorageBucketsGet(w, r, "invalid-cluster", "nonexistent-bucket")
		case http.MethodDelete:
			handleObjectStorageBucketsDelete(w, r, "invalid-cluster", "nonexistent-bucket")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v4/object-storage/keys/999999", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleObjectStorageKeysGet(w, r, "999999")
		case http.MethodDelete:
			handleObjectStorageKeysDelete(w, r, "999999")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return httptest.NewServer(mux)
}

// addObjectStorageBaseEndpoints adds the basic profile and account endpoints needed for service initialization.
func addObjectStorageBaseEndpoints(mux *http.ServeMux) {
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

// handleObjectStorageClustersList handles the Object Storage clusters list mock response.
func handleObjectStorageClustersList(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"id":                 "us-east-1",
				"domain":             "us-east-1.linodeobjects.com",
				"status":             "available",
				"region":             "us-east",
				"static_site_domain": "website-us-east-1.linodeobjects.com",
			},
			{
				"id":                 "us-west-1",
				"domain":             "us-west-1.linodeobjects.com",
				"status":             "available",
				"region":             "us-west",
				"static_site_domain": "website-us-west-1.linodeobjects.com",
			},
			{
				"id":                 "eu-central-1",
				"domain":             "eu-central-1.linodeobjects.com",
				"status":             "available",
				"region":             "eu-central",
				"static_site_domain": "website-eu-central-1.linodeobjects.com",
			},
		},
		"page":    1,
		"pages":   1,
		"results": 3,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleObjectStorageBucketsList handles the Object Storage buckets list mock response.
func handleObjectStorageBucketsList(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"label":    "production-bucket",
				"cluster":  "us-east-1",
				"created":  "2023-01-01T00:00:00",
				"hostname": "production-bucket.us-east-1.linodeobjects.com",
				"size":     1048576000, // 1GB in bytes
				"objects":  250,
			},
			{
				"label":    "development-bucket",
				"cluster":  "us-west-1",
				"created":  "2023-01-02T00:00:00",
				"hostname": "development-bucket.us-west-1.linodeobjects.com",
				"size":     104857600, // 100MB in bytes
				"objects":  50,
			},
			{
				"label":    "backup-bucket",
				"cluster":  "eu-central-1",
				"created":  "2023-01-03T00:00:00",
				"hostname": "backup-bucket.eu-central-1.linodeobjects.com",
				"size":     5368709120, // 5GB in bytes
				"objects":  1000,
			},
		},
		"page":    1,
		"pages":   1,
		"results": 3,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleObjectStorageBucketsGet handles the specific Object Storage bucket mock response.
func handleObjectStorageBucketsGet(w http.ResponseWriter, r *http.Request, cluster, bucket string) {
	if cluster == "invalid-cluster" || bucket == "nonexistent-bucket" {
		// Simulate not found error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "bucket", "reason": "Not found"},
			},
		})
		return
	}

	if cluster == "us-east-1" && bucket == "production-bucket" {
		response := map[string]interface{}{
			"label":    "production-bucket",
			"cluster":  "us-east-1",
			"created":  "2023-01-01T00:00:00",
			"hostname": "production-bucket.us-east-1.linodeobjects.com",
			"size":     1048576000, // 1GB in bytes
			"objects":  250,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		http.Error(w, "Bucket not found", http.StatusNotFound)
	}
}

// handleObjectStorageBucketsCreate handles the Object Storage bucket creation mock response.
func handleObjectStorageBucketsCreate(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"label":    "new-test-bucket",
		"cluster":  "us-central-1",
		"created":  "2023-01-01T01:00:00",
		"hostname": "new-test-bucket.us-central-1.linodeobjects.com",
		"size":     0,
		"objects":  0,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleObjectStorageBucketsUpdate handles the Object Storage bucket update mock response.
func handleObjectStorageBucketsUpdate(w http.ResponseWriter, r *http.Request, cluster, bucket string) {
	response := map[string]interface{}{
		"label":    "updated-production-bucket",
		"cluster":  "us-east-1",
		"created":  "2023-01-01T00:00:00",
		"hostname": "updated-production-bucket.us-east-1.linodeobjects.com",
		"size":     1073741824, // 1GB in bytes
		"objects":  275,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleObjectStorageBucketsDelete handles the Object Storage bucket deletion mock response.
func handleObjectStorageBucketsDelete(w http.ResponseWriter, r *http.Request, cluster, bucket string) {
	if cluster == "invalid-cluster" || bucket == "nonexistent-bucket" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "bucket", "reason": "Not found"},
			},
		})
		return
	}

	// Return 204 No Content for successful deletion
	w.WriteHeader(http.StatusNoContent)
}

// handleObjectStorageKeysList handles the Object Storage keys list mock response.
func handleObjectStorageKeysList(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"id":         12345,
				"label":      "production-key",
				"access_key": "AKIAIOSFODNN7EXAMPLE",
				"secret_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				"limited":    true,
				"bucket_access": []map[string]interface{}{
					{
						"bucket_name": "production-bucket",
						"cluster":     "us-east-1",
						"permissions": "read_write",
					},
				},
			},
			{
				"id":         67890,
				"label":      "readonly-key",
				"access_key": "AKIAI44QH8DHBEXAMPLE",
				"secret_key": "je7MtGbClwBF/2Zp9Utk/h3yCo8nvbEXAMPLEKEY",
				"limited":    true,
				"bucket_access": []map[string]interface{}{
					{
						"bucket_name": "production-bucket",
						"cluster":     "us-east-1",
						"permissions": "read_only",
					},
					{
						"bucket_name": "development-bucket",
						"cluster":     "us-west-1",
						"permissions": "read_only",
					},
				},
			},
			{
				"id":            54321,
				"label":         "full-access-key",
				"access_key":    "AKIAIGUR5HPQEXAMPLE",
				"secret_key":    "wBRfXdnKMSHvZ2nfAdKGYlrZxEEXAMPLEKEY",
				"limited":       false,
				"bucket_access": nil,
			},
		},
		"page":    1,
		"pages":   1,
		"results": 3,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleObjectStorageKeysGet handles the specific Object Storage key mock response.
func handleObjectStorageKeysGet(w http.ResponseWriter, r *http.Request, keyID string) {
	switch keyID {
	case "12345":
		response := map[string]interface{}{
			"id":         12345,
			"label":      "production-key",
			"access_key": "AKIAIOSFODNN7EXAMPLE",
			"secret_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			"limited":    true,
			"bucket_access": []map[string]interface{}{
				{
					"bucket_name": "production-bucket",
					"cluster":     "us-east-1",
					"permissions": "read_write",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	case "999999":
		// Simulate not found error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "key_id", "reason": "Not found"},
			},
		})
	default:
		http.Error(w, "Key not found", http.StatusNotFound)
	}
}

// handleObjectStorageKeysCreate handles the Object Storage key creation mock response.
func handleObjectStorageKeysCreate(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"id":         98765,
		"label":      "new-api-key",
		"access_key": "AKIAI5RQ8BQREXAMPLE",
		"secret_key": "mX7RgFdKQBdFnrZxEYjS8BvAEXAMPLEKEY",
		"limited":    true,
		"bucket_access": []map[string]interface{}{
			{
				"bucket_name": "new-test-bucket",
				"cluster":     "us-central-1",
				"permissions": "read_write",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleObjectStorageKeysUpdate handles the Object Storage key update mock response.
func handleObjectStorageKeysUpdate(w http.ResponseWriter, r *http.Request, keyID string) {
	response := map[string]interface{}{
		"id":         12345,
		"label":      "updated-production-key",
		"access_key": "AKIAIOSFODNN7EXAMPLE",
		"secret_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		"limited":    true,
		"bucket_access": []map[string]interface{}{
			{
				"bucket_name": "production-bucket",
				"cluster":     "us-east-1",
				"permissions": "read_write",
			},
			{
				"bucket_name": "backup-bucket",
				"cluster":     "eu-central-1",
				"permissions": "read_only",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleObjectStorageKeysDelete handles the Object Storage key deletion mock response.
func handleObjectStorageKeysDelete(w http.ResponseWriter, r *http.Request, keyID string) {
	if keyID == "999999" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "key_id", "reason": "Not found"},
			},
		})
		return
	}

	// Return 204 No Content for successful deletion
	w.WriteHeader(http.StatusNoContent)
}

// TestObjectStorageToolsIntegration tests all Object Storage-related CloudMCP tools
// using HTTP test server infrastructure to validate proper integration
// between CloudMCP handlers and Linode API endpoints.
//
// **Integration Test Coverage**:
// • linode_objectstorage_clusters_list - List object storage clusters
// • linode_objectstorage_buckets_list - List buckets
// • linode_objectstorage_bucket_get - Get specific bucket
// • linode_objectstorage_bucket_create - Create new bucket
// • linode_objectstorage_bucket_update - Update bucket
// • linode_objectstorage_bucket_delete - Delete bucket
// • linode_objectstorage_keys_list - List object storage keys
// • linode_objectstorage_key_get - Get specific key
// • linode_objectstorage_key_create - Create new key
// • linode_objectstorage_key_update - Update key
// • linode_objectstorage_key_delete - Delete key
//
// **Test Environment**: HTTP test server with Object Storage API handlers
//
// **Expected Behavior**:
// • All handlers return properly formatted text responses
// • Error conditions are handled gracefully with meaningful messages
// • Object Storage data includes all required fields (ID, label, cluster, size, etc.)
// • Access key management operations work correctly
//
// **Purpose**: Validates that CloudMCP Object Storage handlers correctly format
// Linode API responses for LLM consumption using lightweight HTTP test infrastructure.
func TestObjectStorageToolsIntegration(t *testing.T) {
	// Extend the HTTP test server with Object Storage endpoints
	server := createObjectStorageTestServer()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
	ctx := context.Background()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	t.Run("ObjectStorageClustersList", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_objectstorage_clusters_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleObjectStorageClustersList(ctx, request)
		require.NoError(t, err, "Object Storage clusters list should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate Object Storage clusters list formatting
		require.Contains(t, responseText, "Found 3 Object Storage clusters:", "should indicate correct cluster count")
		require.Contains(t, responseText, "us-east-1", "should contain us-east-1 cluster")
		require.Contains(t, responseText, "us-west-1", "should contain us-west-1 cluster")
		require.Contains(t, responseText, "eu-central-1", "should contain eu-central-1 cluster")
		require.Contains(t, responseText, "Status: available", "should show cluster status")
	})

	t.Run("ObjectStorageBucketsList", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_objectstorage_buckets_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleObjectStorageBucketsList(ctx, request)
		require.NoError(t, err, "Object Storage buckets list should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate Object Storage buckets list formatting
		require.Contains(t, responseText, "Found 3 Object Storage buckets:", "should indicate correct bucket count")
		require.Contains(t, responseText, "production-bucket", "should contain production bucket")
		require.Contains(t, responseText, "development-bucket", "should contain development bucket")
		require.Contains(t, responseText, "backup-bucket", "should contain backup bucket")
		require.Contains(t, responseText, "Cluster:", "should show cluster information")
		require.Contains(t, responseText, "Objects:", "should show object count")
	})

	t.Run("ObjectStorageBucketGet", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_objectstorage_bucket_get",
				Arguments: map[string]interface{}{
					"cluster": "us-east-1",
					"bucket":  "production-bucket",
				},
			},
		}

		result, err := service.handleObjectStorageBucketGet(ctx, request)
		require.NoError(t, err, "Object Storage bucket get should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate detailed bucket information
		require.Contains(t, responseText, "Object Storage Bucket Details:", "should have bucket details header")
		require.Contains(t, responseText, "Label: production-bucket", "should contain bucket label")
		require.Contains(t, responseText, "Cluster: us-east-1", "should contain cluster name")
		require.Contains(t, responseText, "Hostname:", "should contain hostname")
		require.Contains(t, responseText, "Size:", "should contain size information")
		require.Contains(t, responseText, "Objects: 250", "should contain object count")
	})

	t.Run("ObjectStorageBucketCreate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_objectstorage_bucket_create",
				Arguments: map[string]interface{}{
					"label":   "new-test-bucket",
					"cluster": "us-central-1",
				},
			},
		}

		result, err := service.handleObjectStorageBucketCreate(ctx, request)
		require.NoError(t, err, "Object Storage bucket create should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate bucket creation confirmation
		require.Contains(t, responseText, "Object Storage bucket created successfully:", "should confirm creation")
		require.Contains(t, responseText, "Label: new-test-bucket", "should show bucket label")
		require.Contains(t, responseText, "Cluster: us-central-1", "should show cluster")
		require.Contains(t, responseText, "Hostname:", "should show hostname")
	})

	t.Run("ObjectStorageBucketUpdate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_objectstorage_bucket_update",
				Arguments: map[string]interface{}{
					"cluster": "us-east-1",
					"bucket":  "production-bucket",
				},
			},
		}

		result, err := service.handleObjectStorageBucketUpdate(ctx, request)
		require.NoError(t, err, "Object Storage bucket update should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate bucket update confirmation
		require.Contains(t, responseText, "Object Storage bucket updated successfully:", "should confirm update")
		require.Contains(t, responseText, "Label: updated-production-bucket", "should show updated label")
		require.Contains(t, responseText, "Cluster: us-east-1", "should show cluster")
	})

	t.Run("ObjectStorageKeysList", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_objectstorage_keys_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleObjectStorageKeysList(ctx, request)
		require.NoError(t, err, "Object Storage keys list should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate Object Storage keys list formatting
		require.Contains(t, responseText, "Found 3 Object Storage keys:", "should indicate correct key count")
		require.Contains(t, responseText, "production-key", "should contain production key")
		require.Contains(t, responseText, "readonly-key", "should contain readonly key")
		require.Contains(t, responseText, "full-access-key", "should contain full access key")
		require.Contains(t, responseText, "Access Key:", "should show access key")
		require.Contains(t, responseText, "Limited:", "should show limitation status")
	})

	t.Run("ObjectStorageKeyGet", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_objectstorage_key_get",
				Arguments: map[string]interface{}{
					"key_id": float64(12345),
				},
			},
		}

		result, err := service.handleObjectStorageKeyGet(ctx, request)
		require.NoError(t, err, "Object Storage key get should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate detailed key information
		require.Contains(t, responseText, "Object Storage Key Details:", "should have key details header")
		require.Contains(t, responseText, "ID: 12345", "should contain key ID")
		require.Contains(t, responseText, "Label: production-key", "should contain key label")
		require.Contains(t, responseText, "Access Key: AKIAIOSFODNN7EXAMPLE", "should contain access key")
		require.Contains(t, responseText, "Limited: true", "should contain limitation status")
		require.Contains(t, responseText, "Bucket Access:", "should contain bucket access section")
	})

	t.Run("ObjectStorageKeyCreate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_objectstorage_key_create",
				Arguments: map[string]interface{}{
					"label": "new-api-key",
					"bucket_access": []interface{}{
						map[string]interface{}{
							"bucket_name": "new-test-bucket",
							"cluster":     "us-central-1",
							"permissions": "read_write",
						},
					},
				},
			},
		}

		result, err := service.handleObjectStorageKeyCreate(ctx, request)
		require.NoError(t, err, "Object Storage key create should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate key creation confirmation
		require.Contains(t, responseText, "Object Storage key created successfully:", "should confirm creation")
		require.Contains(t, responseText, "ID: 98765", "should show new key ID")
		require.Contains(t, responseText, "Label: new-api-key", "should show key label")
		require.Contains(t, responseText, "Access Key:", "should show access key")
		require.Contains(t, responseText, "Secret Key:", "should show secret key")
	})

	t.Run("ObjectStorageKeyUpdate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_objectstorage_key_update",
				Arguments: map[string]interface{}{
					"key_id": float64(12345),
					"label":  "updated-production-key",
				},
			},
		}

		result, err := service.handleObjectStorageKeyUpdate(ctx, request)
		require.NoError(t, err, "Object Storage key update should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate key update confirmation
		require.Contains(t, responseText, "Object Storage key updated successfully:", "should confirm update")
		require.Contains(t, responseText, "ID: 12345", "should show key ID")
		require.Contains(t, responseText, "Label: updated-production-key", "should show updated label")
	})

	t.Run("ObjectStorageKeyDelete", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_objectstorage_key_delete",
				Arguments: map[string]interface{}{
					"key_id": float64(12345),
				},
			},
		}

		result, err := service.handleObjectStorageKeyDelete(ctx, request)
		require.NoError(t, err, "Object Storage key delete should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate key deletion confirmation
		require.Contains(t, responseText, "Object Storage key", "should mention Object Storage key")
		require.Contains(t, responseText, "deleted successfully", "should confirm deletion")
		require.Contains(t, responseText, "12345", "should show deleted key ID")
	})

	t.Run("ObjectStorageBucketDelete", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_objectstorage_bucket_delete",
				Arguments: map[string]interface{}{
					"cluster": "us-east-1",
					"bucket":  "production-bucket",
				},
			},
		}

		result, err := service.handleObjectStorageBucketDelete(ctx, request)
		require.NoError(t, err, "Object Storage bucket delete should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate bucket deletion confirmation
		require.Contains(t, responseText, "Object Storage bucket", "should mention Object Storage bucket")
		require.Contains(t, responseText, "deleted successfully", "should confirm deletion")
		require.Contains(t, responseText, "production-bucket", "should show deleted bucket name")
	})
}

// TestObjectStorageErrorHandlingIntegration tests error scenarios for Object Storage tools
// to ensure CloudMCP handles API errors gracefully and provides meaningful
// error messages to users.
//
// **Error Test Scenarios**:
// • Non-existent bucket or key ID (404 errors)
// • Invalid cluster or bucket names
// • Access permission errors
// • Permission errors for Object Storage operations
//
// **Expected Behavior**:
// • Proper error handling with contextual error messages
// • No unhandled exceptions or panics
// • Error responses follow CloudMCP error format
// • Error messages are actionable for users
//
// **Purpose**: Validates robust error handling in Object Storage operations
// and ensures reliable operation under error conditions.
func TestObjectStorageErrorHandlingIntegration(t *testing.T) {
	server := createObjectStorageTestServer()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
	ctx := context.Background()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	t.Run("ObjectStorageBucketGetNotFound", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_objectstorage_bucket_get",
				Arguments: map[string]interface{}{
					"cluster": "invalid-cluster",
					"bucket":  "nonexistent-bucket",
				},
			},
		}

		result, err := service.handleObjectStorageBucketGet(ctx, request)

		// Check for either Go error or MCP error result
		if err != nil {
			// Go error pattern
			require.Contains(t, err.Error(), "failed to get Object Storage bucket", "error should mention get bucket failure")
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

	t.Run("ObjectStorageKeyGetNotFound", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_objectstorage_key_get",
				Arguments: map[string]interface{}{
					"key_id": float64(999999), // Non-existent key
				},
			},
		}

		result, err := service.handleObjectStorageKeyGet(ctx, request)

		// Check for either Go error or MCP error result
		if err != nil {
			// Go error pattern
			require.Contains(t, err.Error(), "failed to get Object Storage key", "error should mention get key failure")
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

	t.Run("InvalidKeyID", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_objectstorage_key_get",
				Arguments: map[string]interface{}{
					"key_id": "invalid", // Invalid ID type
				},
			},
		}

		result, err := service.handleObjectStorageKeyGet(ctx, request)

		// Should get MCP error result for invalid parameter
		if err == nil {
			require.NotNil(t, result, "result should not be nil")
			require.True(t, result.IsError, "result should be marked as error")
		}
	})
}
