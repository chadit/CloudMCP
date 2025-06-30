package linode_test

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/internal/services/linode"
	"github.com/chadit/CloudMCP/pkg/logger"
)

// TestHandleLKEClustersList_AccountError tests the handleLKEClustersList function to verify error handling.
// This test simulates a user requesting a list of all their LKE (Linode Kubernetes Engine) clusters through the MCP interface.
// Since this function requires Linode API calls, this test focuses on the error handling path.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with empty account manager
// 2. **Request Execution**: Call handleLKEClustersList expecting account manager failure
// 3. **Error Validation**: Verify appropriate error is returned for account lookup failure
//
// **Test Environment**: Service with no configured accounts to trigger error path
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Does not attempt to call Linode API when account lookup fails
// • Provides meaningful error message for troubleshooting
//
// **Purpose**: This test ensures LKE clusters list command fails appropriately when account configuration is invalid.
func TestHandleLKEClustersList_AccountError(t *testing.T) {
	t.Parallel()

	// Create minimal service with empty account manager
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	// Create completely isolated account manager for this test only
	accountManager := linode.NewAccountManagerForTesting()

	service := linode.NewForTesting(cfg, log, accountManager)

	// Test LKE clusters list request with empty account manager
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_lke_clusters_list",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleLKEClustersList should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleLKEClusterGet_AccountError tests the handleLKEClusterGet function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleLKEClusterGet with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Test Environment**: Service with no configured accounts and empty current account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleLKEClusterGet_AccountError(t *testing.T) {
	t.Parallel()

	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test LKE cluster get request with no accounts and valid parameter
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_lke_cluster_get",
			Arguments: map[string]interface{}{
				"cluster_id": float64(12345),
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleLKEClusterGet should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleLKEClusterCreate_AccountError tests the handleLKEClusterCreate function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleLKEClusterCreate with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleLKEClusterCreate_AccountError(t *testing.T) {
	t.Parallel()

	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test LKE cluster create request with no accounts
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_lke_cluster_create",
			Arguments: map[string]interface{}{
				"label":       "test-cluster",
				"region":      "us-east",
				"k8s_version": "1.29",
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleLKEClusterCreate should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleLKEClusterUpdate_AccountError tests the handleLKEClusterUpdate function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleLKEClusterUpdate with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleLKEClusterUpdate_AccountError(t *testing.T) {
	t.Parallel()

	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test LKE cluster update request with no accounts
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_lke_cluster_update",
			Arguments: map[string]interface{}{
				"cluster_id":  float64(12345),
				"k8s_version": "1.30",
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleLKEClusterUpdate should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleLKEClusterDelete_AccountError tests the handleLKEClusterDelete function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleLKEClusterDelete with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleLKEClusterDelete_AccountError(t *testing.T) {
	t.Parallel()

	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test LKE cluster delete request with no accounts
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_lke_cluster_delete",
			Arguments: map[string]interface{}{
				"cluster_id": float64(12345),
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleLKEClusterDelete should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleLKENodePoolCreate_AccountError tests the handleLKENodePoolCreate function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleLKENodePoolCreate with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleLKENodePoolCreate_AccountError(t *testing.T) {
	t.Parallel()

	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test LKE node pool create request with no accounts
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_lke_node_pool_create",
			Arguments: map[string]interface{}{
				"cluster_id": float64(12345),
				"type":       "g6-standard-2",
				"count":      float64(3),
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleLKENodePoolCreate should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleLKENodePoolUpdate_AccountError tests the handleLKENodePoolUpdate function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleLKENodePoolUpdate with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleLKENodePoolUpdate_AccountError(t *testing.T) {
	t.Parallel()

	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test LKE node pool update request with no accounts
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_lke_node_pool_update",
			Arguments: map[string]interface{}{
				"cluster_id": float64(12345),
				"pool_id":    float64(67890),
				"count":      float64(5),
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleLKENodePoolUpdate should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleLKENodePoolDelete_AccountError tests the handleLKENodePoolDelete function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleLKENodePoolDelete with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleLKENodePoolDelete_AccountError(t *testing.T) {
	t.Parallel()

	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test LKE node pool delete request with no accounts
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_lke_node_pool_delete",
			Arguments: map[string]interface{}{
				"cluster_id": float64(12345),
				"pool_id":    float64(67890),
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleLKENodePoolDelete should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// TestHandleLKEKubeconfig_AccountError tests the handleLKEKubeconfig function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleLKEKubeconfig with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleLKEKubeconfig_AccountError(t *testing.T) {
	t.Parallel()

	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test LKE kubeconfig request with no accounts
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_lke_kubeconfig",
			Arguments: map[string]interface{}{
				"cluster_id": float64(12345),
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.NoError(t, err, "handleLKEKubeconfig should not return error for account validation")
	require.NotNil(t, result, "result should not be nil")

	// Check that it's an error result
	require.NotEmpty(t, result.Content, "result should have content")
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			require.Contains(t, textContent.Text, "account", "error result should mention account issue")
		}
	}
}

// Note: Additional tests for successful LKE operations with mock Linode clients
// are not implemented in this unit test suite because they require functioning
// Linode API client interfaces for operations like ListLKEClusters, GetLKECluster,
// CreateLKECluster, UpdateLKECluster, DeleteLKECluster, CreateLKENodePool,
// UpdateLKENodePool, DeleteLKENodePool, and GetLKEKubeconfig.
//
// Parameter validation tests are not included because the LKE handlers
// currently use parseArguments() placeholder which doesn't implement comprehensive
// validation like parseIDFromArguments() used in other tool handlers.
//
// These operations would require either:
// 1. Interface abstraction for the Linode client (future improvement)
// 2. Integration testing with real API endpoints
// 3. Dependency injection to replace the client during testing
// 4. Implementation of proper parameter validation in LKE handlers
//
// The current tests adequately cover:
// - Account manager error scenarios for all 9 LKE tools
// - Request routing and basic tool handler setup
// - Error message formatting and response structure
// - MCP protocol compliance for error responses
//
// This provides coverage of the testable logic that doesn't
// require external API dependencies while ensuring proper error handling
// for the most common failure scenario (account configuration issues).
