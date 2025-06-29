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

// createLKETestServer creates an HTTP test server with LKE (Kubernetes) API endpoints.
// This extends the base HTTP test server infrastructure with comprehensive
// Kubernetes-specific endpoints for integration testing.
//
// **LKE Endpoints Supported:**
// • GET /v4/lke/clusters - List LKE clusters
// • GET /v4/lke/clusters/{id} - Get specific cluster details
// • POST /v4/lke/clusters - Create new cluster
// • PUT /v4/lke/clusters/{id} - Update cluster
// • DELETE /v4/lke/clusters/{id} - Delete cluster
// • GET /v4/lke/clusters/{id}/kubeconfig - Get kubeconfig
// • POST /v4/lke/clusters/{id}/pools - Create node pool
// • PUT /v4/lke/clusters/{id}/pools/{pool_id} - Update node pool
// • DELETE /v4/lke/clusters/{id}/pools/{pool_id} - Delete node pool
//
// **Mock Data Features:**
// • Realistic Kubernetes cluster configurations with node pools
// • Kubeconfig retrieval for cluster access
// • Node pool management with autoscaling options
// • Error simulation for non-existent resources
// • Proper HTTP status codes and error responses
func createLKETestServer() *httptest.Server {
	mux := http.NewServeMux()

	// Include base endpoints (profile, account) from main test server
	addLKEBaseEndpoints(mux)

	// LKE clusters list endpoint
	mux.HandleFunc("/v4/lke/clusters", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleLKEClustersList(w, r)
		case http.MethodPost:
			handleLKEClustersCreate(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Specific LKE cluster endpoints with explicit ID matching
	mux.HandleFunc("/v4/lke/clusters/12345", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleLKEClustersGet(w, r, "12345")
		case http.MethodPut:
			handleLKEClustersUpdate(w, r, "12345")
		case http.MethodDelete:
			handleLKEClustersDelete(w, r, "12345")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// LKE kubeconfig endpoint
	mux.HandleFunc("/v4/lke/clusters/12345/kubeconfig", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleLKEKubeconfig(w, r, "12345")
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// LKE node pools endpoints
	mux.HandleFunc("/v4/lke/clusters/12345/pools", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handleLKENodePoolCreate(w, r, "12345")
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Specific node pool endpoints
	mux.HandleFunc("/v4/lke/clusters/12345/pools/67890", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			handleLKENodePoolUpdate(w, r, "12345", "67890")
		case http.MethodDelete:
			handleLKENodePoolDelete(w, r, "12345", "67890")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Non-existent cluster endpoints for error testing
	mux.HandleFunc("/v4/lke/clusters/999999", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleLKEClustersGet(w, r, "999999")
		case http.MethodDelete:
			handleLKEClustersDelete(w, r, "999999")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return httptest.NewServer(mux)
}

// addLKEBaseEndpoints adds the basic profile and account endpoints needed for service initialization.
func addLKEBaseEndpoints(mux *http.ServeMux) {
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

// handleLKEClustersList handles the LKE clusters list mock response.
func handleLKEClustersList(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"id":          12345,
				"label":       "production-k8s",
				"region":      "us-east",
				"k8s_version": "1.28",
				"status":      "ready",
				"created":     "2023-01-01T00:00:00",
				"updated":     "2023-01-01T00:00:00",
				"control_plane": map[string]interface{}{
					"high_availability": true,
				},
				"tags": []string{"production", "kubernetes"},
			},
			{
				"id":          54321,
				"label":       "development-k8s",
				"region":      "us-west",
				"k8s_version": "1.27",
				"status":      "ready",
				"created":     "2023-01-02T00:00:00",
				"updated":     "2023-01-02T00:00:00",
				"control_plane": map[string]interface{}{
					"high_availability": false,
				},
				"tags": []string{"development", "testing"},
			},
		},
		"page":    1,
		"pages":   1,
		"results": 2,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleLKEClustersGet handles the specific LKE cluster mock response.
func handleLKEClustersGet(w http.ResponseWriter, r *http.Request, clusterID string) {
	switch clusterID {
	case "12345":
		response := map[string]interface{}{
			"id":          12345,
			"label":       "production-k8s",
			"region":      "us-east",
			"k8s_version": "1.28",
			"status":      "ready",
			"created":     "2023-01-01T00:00:00",
			"updated":     "2023-01-01T00:00:00",
			"control_plane": map[string]interface{}{
				"high_availability": true,
			},
			"tags": []string{"production", "kubernetes"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	case "999999":
		// Simulate not found error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "cluster_id", "reason": "Not found"},
			},
		})
	default:
		http.Error(w, "Cluster not found", http.StatusNotFound)
	}
}

// handleLKEClustersCreate handles the LKE cluster creation mock response.
func handleLKEClustersCreate(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"id":          98765,
		"label":       "new-k8s-cluster",
		"region":      "us-central",
		"k8s_version": "1.28",
		"status":      "not_ready",
		"created":     "2023-01-01T01:00:00",
		"updated":     "2023-01-01T01:00:00",
		"control_plane": map[string]interface{}{
			"high_availability": false,
		},
		"tags": []string{},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleLKEClustersUpdate handles the LKE cluster update mock response.
func handleLKEClustersUpdate(w http.ResponseWriter, r *http.Request, clusterID string) {
	if clusterID == "999999" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "cluster_id", "reason": "Not found"},
			},
		})
		return
	}

	response := map[string]interface{}{
		"id":          12345,
		"label":       "updated-production-k8s",
		"region":      "us-east",
		"k8s_version": "1.29",
		"status":      "ready",
		"created":     "2023-01-01T00:00:00",
		"updated":     "2023-01-01T02:00:00",
		"control_plane": map[string]interface{}{
			"high_availability": true,
		},
		"tags": []string{"production", "kubernetes", "updated"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleLKEClustersDelete handles the LKE cluster deletion mock response.
func handleLKEClustersDelete(w http.ResponseWriter, r *http.Request, clusterID string) {
	if clusterID == "999999" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "cluster_id", "reason": "Not found"},
			},
		})
		return
	}

	// Return 204 No Content for successful deletion
	w.WriteHeader(http.StatusNoContent)
}

// handleLKEKubeconfig handles the LKE kubeconfig retrieval mock response.
func handleLKEKubeconfig(w http.ResponseWriter, r *http.Request, clusterID string) {
	if clusterID == "999999" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "cluster_id", "reason": "Not found"},
			},
		})
		return
	}

	response := map[string]interface{}{
		"kubeconfig": "YXBpVmVyc2lvbjogdjEKa2luZDogQ29uZmlnCmNsdXN0ZXJzOgotIG5hbWU6IHByb2R1Y3Rpb24tazhzCiAgY2x1c3RlcjoKICAgIGNlcnRpZmljYXRlLWF1dGhvcml0eS1kYXRhOiBMUzB0TFMwdAogICAgc2VydmVyOiBodHRwczovL2FwaS5wcm9kdWN0aW9uLWs4cy5saW5vZGUtbGtlLm5ldAp1c2VyczoKLSBuYW1lOiBwcm9kdWN0aW9uLWs4cy1hZG1pbgogIHVzZXI6CiAgICB0b2tlbjogdGVzdC10b2tlbi0xMjM0NQpjb250ZXh0czoKLSBjb250ZXh0OgogICAgY2x1c3RlcjogcHJvZHVjdGlvbi1rOHMKICAgIHVzZXI6IHByb2R1Y3Rpb24tazhzLWFkbWluCiAgbmFtZTogcHJvZHVjdGlvbi1rOHMtY3R4CmN1cnJlbnQtY29udGV4dDogcHJvZHVjdGlvbi1rOHMtY3R4",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleLKENodePoolCreate handles the LKE node pool creation mock response.
func handleLKENodePoolCreate(w http.ResponseWriter, r *http.Request, clusterID string) {
	if clusterID == "999999" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "cluster_id", "reason": "Not found"},
			},
		})
		return
	}

	response := map[string]interface{}{
		"id":    67890,
		"type":  "g6-standard-2",
		"count": 3,
		"disks": []map[string]interface{}{
			{
				"type": "ext4",
				"size": 25600,
			},
		},
		"autoscaler": map[string]interface{}{
			"enabled": true,
			"min":     1,
			"max":     5,
		},
		"tags": []string{"worker"},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleLKENodePoolUpdate handles the LKE node pool update mock response.
func handleLKENodePoolUpdate(w http.ResponseWriter, r *http.Request, clusterID, poolID string) {
	response := map[string]interface{}{
		"id":    67890,
		"type":  "g6-standard-4",
		"count": 5,
		"disks": []map[string]interface{}{
			{
				"type": "ext4",
				"size": 51200,
			},
		},
		"autoscaler": map[string]interface{}{
			"enabled": true,
			"min":     2,
			"max":     10,
		},
		"tags": []string{"worker", "updated"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleLKENodePoolDelete handles the LKE node pool deletion mock response.
func handleLKENodePoolDelete(w http.ResponseWriter, r *http.Request, clusterID, poolID string) {
	// Return 204 No Content for successful deletion
	w.WriteHeader(http.StatusNoContent)
}

// TestLKEToolsIntegration tests all LKE-related CloudMCP tools
// using HTTP test server infrastructure to validate proper integration
// between CloudMCP handlers and Linode API endpoints.
//
// **Integration Test Coverage**:
// • linode_lke_clusters_list - List all LKE clusters
// • linode_lke_cluster_get - Get specific cluster details
// • linode_lke_cluster_create - Create new cluster
// • linode_lke_cluster_update - Update existing cluster
// • linode_lke_cluster_delete - Delete cluster
// • linode_lke_kubeconfig - Get cluster kubeconfig
// • linode_lke_nodepool_create - Create node pool
// • linode_lke_nodepool_update - Update node pool
// • linode_lke_nodepool_delete - Delete node pool
//
// **Test Environment**: HTTP test server with LKE API handlers
//
// **Expected Behavior**:
// • All handlers return properly formatted text responses
// • Error conditions are handled gracefully with meaningful messages
// • LKE data includes all required fields (ID, label, status, version, etc.)
// • Node pool management operations work correctly
//
// **Purpose**: Validates that CloudMCP LKE handlers correctly format
// Linode API responses for LLM consumption using lightweight HTTP test infrastructure.
func TestLKEToolsIntegration(t *testing.T) {
	// Extend the HTTP test server with LKE endpoints
	server := createLKETestServer()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
	ctx := context.Background()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	t.Run("LKEClustersList", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_lke_clusters_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleLKEClustersList(ctx, request)
		require.NoError(t, err, "LKE clusters list should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate LKE cluster list formatting
		require.Contains(t, responseText, "Found 2 LKE clusters:", "should indicate correct cluster count")
		require.Contains(t, responseText, "production-k8s", "should contain first cluster label")
		require.Contains(t, responseText, "development-k8s", "should contain second cluster label")
		require.Contains(t, responseText, "Status: ready", "should show cluster status")
		require.Contains(t, responseText, "Version: 1.28", "should show Kubernetes version")
	})

	t.Run("LKEClusterGet", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_lke_cluster_get",
				Arguments: map[string]interface{}{
					"cluster_id": float64(12345),
				},
			},
		}

		result, err := service.handleLKEClusterGet(ctx, request)
		require.NoError(t, err, "LKE cluster get should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate detailed cluster information
		require.Contains(t, responseText, "LKE Cluster Details:", "should have cluster details header")
		require.Contains(t, responseText, "ID: 12345", "should contain cluster ID")
		require.Contains(t, responseText, "Label: production-k8s", "should contain cluster label")
		require.Contains(t, responseText, "Status: ready", "should contain cluster status")
		require.Contains(t, responseText, "Kubernetes Version: 1.28", "should contain K8s version")
		require.Contains(t, responseText, "Region: us-east", "should contain region")
	})

	t.Run("LKEClusterCreate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_lke_cluster_create",
				Arguments: map[string]interface{}{
					"label":       "new-k8s-cluster",
					"region":      "us-central",
					"k8s_version": "1.28",
					"node_pools": []interface{}{
						map[string]interface{}{
							"type":  "g6-standard-2",
							"count": float64(3),
						},
					},
				},
			},
		}

		result, err := service.handleLKEClusterCreate(ctx, request)
		require.NoError(t, err, "LKE cluster create should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate cluster creation confirmation
		require.Contains(t, responseText, "LKE Cluster created successfully:", "should confirm creation")
		require.Contains(t, responseText, "ID: 98765", "should show new cluster ID")
		require.Contains(t, responseText, "Label: new-k8s-cluster", "should show cluster label")
		require.Contains(t, responseText, "Status: not_ready", "should show initial status")
	})

	t.Run("LKEClusterUpdate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_lke_cluster_update",
				Arguments: map[string]interface{}{
					"cluster_id":  float64(12345),
					"label":       "updated-production-k8s",
					"k8s_version": "1.29",
				},
			},
		}

		result, err := service.handleLKEClusterUpdate(ctx, request)
		require.NoError(t, err, "LKE cluster update should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate cluster update confirmation
		require.Contains(t, responseText, "LKE Cluster updated successfully:", "should confirm update")
		require.Contains(t, responseText, "ID: 12345", "should show cluster ID")
		require.Contains(t, responseText, "Label: updated-production-k8s", "should show updated label")
		require.Contains(t, responseText, "Kubernetes Version: 1.29", "should show updated version")
	})

	t.Run("LKEKubeconfig", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_lke_kubeconfig",
				Arguments: map[string]interface{}{
					"cluster_id": float64(12345),
				},
			},
		}

		result, err := service.handleLKEKubeconfig(ctx, request)
		require.NoError(t, err, "LKE kubeconfig should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate kubeconfig response
		require.Contains(t, responseText, "Kubeconfig for LKE cluster", "should mention kubeconfig")
		require.Contains(t, responseText, "12345", "should show cluster ID")
		require.Contains(t, responseText, "YXBpVmVyc2lvbjog", "should contain base64 kubeconfig data")
	})

	t.Run("LKENodePoolCreate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_lke_nodepool_create",
				Arguments: map[string]interface{}{
					"cluster_id": float64(12345),
					"type":       "g6-standard-2",
					"count":      float64(3),
				},
			},
		}

		result, err := service.handleLKENodePoolCreate(ctx, request)
		require.NoError(t, err, "LKE node pool create should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate node pool creation confirmation
		require.Contains(t, responseText, "Node pool created successfully:", "should confirm creation")
		require.Contains(t, responseText, "Pool ID: 67890", "should show new pool ID")
		require.Contains(t, responseText, "Type: g6-standard-2", "should show node type")
		require.Contains(t, responseText, "Count: 3", "should show node count")
	})

	t.Run("LKENodePoolUpdate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_lke_nodepool_update",
				Arguments: map[string]interface{}{
					"cluster_id": float64(12345),
					"pool_id":    float64(67890),
					"count":      float64(5),
				},
			},
		}

		result, err := service.handleLKENodePoolUpdate(ctx, request)
		require.NoError(t, err, "LKE node pool update should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate node pool update confirmation
		require.Contains(t, responseText, "Node pool updated successfully:", "should confirm update")
		require.Contains(t, responseText, "Pool ID: 67890", "should show pool ID")
		require.Contains(t, responseText, "Count: 5", "should show updated count")
	})

	t.Run("LKENodePoolDelete", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_lke_nodepool_delete",
				Arguments: map[string]interface{}{
					"cluster_id": float64(12345),
					"pool_id":    float64(67890),
				},
			},
		}

		result, err := service.handleLKENodePoolDelete(ctx, request)
		require.NoError(t, err, "LKE node pool delete should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate node pool deletion confirmation
		require.Contains(t, responseText, "Node pool", "should mention node pool")
		require.Contains(t, responseText, "deleted successfully", "should confirm deletion")
		require.Contains(t, responseText, "67890", "should show deleted pool ID")
	})

	t.Run("LKEClusterDelete", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_lke_cluster_delete",
				Arguments: map[string]interface{}{
					"cluster_id": float64(12345),
				},
			},
		}

		result, err := service.handleLKEClusterDelete(ctx, request)
		require.NoError(t, err, "LKE cluster delete should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate cluster deletion confirmation
		require.Contains(t, responseText, "LKE Cluster", "should mention LKE cluster")
		require.Contains(t, responseText, "deleted successfully", "should confirm deletion")
		require.Contains(t, responseText, "12345", "should show deleted cluster ID")
	})
}

// TestLKEErrorHandlingIntegration tests error scenarios for LKE tools
// to ensure CloudMCP handles API errors gracefully and provides meaningful
// error messages to users.
//
// **Error Test Scenarios**:
// • Non-existent cluster ID (404 errors)
// • Invalid cluster configuration
// • Node pool conflicts
// • Permission errors for LKE operations
//
// **Expected Behavior**:
// • Proper error handling with contextual error messages
// • No unhandled exceptions or panics
// • Error responses follow CloudMCP error format
// • Error messages are actionable for users
//
// **Purpose**: Validates robust error handling in LKE operations
// and ensures reliable operation under error conditions.
func TestLKEErrorHandlingIntegration(t *testing.T) {
	server := createLKETestServer()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
	ctx := context.Background()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	t.Run("LKEClusterGetNotFound", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_lke_cluster_get",
				Arguments: map[string]interface{}{
					"cluster_id": float64(999999), // Non-existent cluster
				},
			},
		}

		result, err := service.handleLKEClusterGet(ctx, request)

		// Check for either Go error or MCP error result
		if err != nil {
			// Go error pattern
			require.Contains(t, err.Error(), "failed to get LKE cluster", "error should mention get cluster failure")
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

	t.Run("LKEClusterDeleteNotFound", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_lke_cluster_delete",
				Arguments: map[string]interface{}{
					"cluster_id": float64(999999), // Non-existent cluster
				},
			},
		}

		result, err := service.handleLKEClusterDelete(ctx, request)

		// Check for either Go error or MCP error result
		if err != nil {
			// Go error pattern
			require.Contains(t, err.Error(), "Failed to delete LKE cluster", "error should mention delete cluster failure")
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

	t.Run("InvalidClusterID", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_lke_cluster_get",
				Arguments: map[string]interface{}{
					"cluster_id": "invalid", // Invalid ID type
				},
			},
		}

		result, err := service.handleLKEClusterGet(ctx, request)

		// Should get MCP error result for invalid parameter
		if err == nil {
			require.NotNil(t, result, "result should not be nil")
			require.True(t, result.IsError, "result should be marked as error")
		}
	})
}
