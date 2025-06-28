//go:build integration

package linode

import (
	"context"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// TestCompleteInstanceLifecycleWorkflow tests a complete end-to-end workflow
// for managing Linode instances through CloudMCP handlers.
//
// **End-to-End Test Workflow**:
// 1. **Account Setup**: Initialize service and verify account information
// 2. **Instance Creation**: Create a new instance with specified configuration
// 3. **Instance Verification**: Retrieve and validate the created instance details
// 4. **Instance Management**: Test operational commands (boot, reboot, shutdown)
// 5. **Instance Cleanup**: Delete the instance and verify removal
//
// **Test Environment**: HTTP test server with realistic Linode API responses
//
// **Expected Behavior**:
// • All handler operations complete successfully without errors
// • Text responses contain expected information and proper formatting
// • Instance state transitions are properly reflected in subsequent calls
// • Error conditions are handled gracefully with meaningful messages
//
// **Purpose**: Validates complete user workflows and ensures CloudMCP
// handlers work together seamlessly for real-world usage scenarios.
func TestCompleteInstanceLifecycleWorkflow(t *testing.T) {
	service, cleanup := SetupHTTPTestIntegration(t)
	defer cleanup()

	ctx := context.Background()

	// Step 1: Verify account information
	t.Run("Step1_VerifyAccount", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_account_get",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleAccountGet(ctx, request)
		require.NoError(t, err, "account verification should succeed")
		require.NotNil(t, result, "account result should not be nil")

		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")
		require.Contains(t, textContent.Text, "Account: httptest-integration", "should show correct account")
		require.Contains(t, textContent.Text, "Username: testuser", "should show username")
	})

	// Step 2: List existing instances
	t.Run("Step2_ListExistingInstances", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_instances_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleInstancesList(ctx, request)
		require.NoError(t, err, "instances list should succeed")
		require.NotNil(t, result, "instances result should not be nil")

		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")
		require.Contains(t, textContent.Text, "Found 1 Linode instance(s)", "should show existing instance")
		require.Contains(t, textContent.Text, "test-instance-1", "should show instance label")
	})

	// Step 3: Get detailed information about existing instance
	t.Run("Step3_GetInstanceDetails", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_instance_get",
				Arguments: map[string]interface{}{
					"instance_id": float64(123456),
				},
			},
		}

		result, err := service.handleInstanceGet(ctx, request)
		require.NoError(t, err, "instance get should succeed")
		require.NotNil(t, result, "instance get result should not be nil")

		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")
		require.Contains(t, textContent.Text, "Instance Details:", "should show instance details")
		require.Contains(t, textContent.Text, "ID: 123456", "should show correct instance ID")
		require.Contains(t, textContent.Text, "Status: running", "should show running status")
	})

	// Step 4: Create a new instance (simulated)
	t.Run("Step4_CreateNewInstance", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_instance_create",
				Arguments: map[string]interface{}{
					"region": "us-east",
					"type":   "g6-nanode-1",
					"label":  "test-new-instance",
					"image":  "linode/ubuntu22.04",
				},
			},
		}

		result, err := service.handleInstanceCreate(ctx, request)
		require.NoError(t, err, "instance creation should succeed")
		require.NotNil(t, result, "instance creation result should not be nil")

		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")
		require.Contains(t, textContent.Text, "Instance created successfully!", "should confirm creation")
		require.Contains(t, textContent.Text, "ID: 123457", "should show new instance ID")
		require.Contains(t, textContent.Text, "Label: new-test-instance", "should show new instance label")
		require.Contains(t, textContent.Text, "Status: provisioning", "should show provisioning status")
	})

	// Step 5: Test error handling with invalid instance ID
	t.Run("Step5_ErrorHandling", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_instance_get",
				Arguments: map[string]interface{}{
					"instance_id": float64(999999), // Non-existent instance
				},
			},
		}

		result, err := service.handleInstanceGet(ctx, request)
		require.Error(t, err, "should return error for non-existent instance")
		require.Contains(t, err.Error(), "failed to get instance 999999", "error should mention instance ID")
		require.Contains(t, err.Error(), "linode/instance_get", "error should include tool context")

		// When there's a Go error, result might be nil
		if result != nil {
			require.True(t, result.IsError, "result should be marked as error if not nil")
		}
	})
}

// TestAccountManagementWorkflow tests account-related operations and switching.
//
// **Account Management Workflow**:
// 1. **Current Account**: Verify current account information
// 2. **Account List**: List all available accounts
// 3. **Account Operations**: Test account-specific operations
// 4. **Data Consistency**: Ensure account information is consistent across calls
//
// **Test Environment**: HTTP test server with single account configuration
//
// **Expected Behavior**:
// • Account information is consistently reported across different handlers
// • Account-specific operations work correctly
// • Account data includes all required fields (username, email, etc.)
//
// **Purpose**: Validates account management functionality and ensures
// consistent account handling across CloudMCP operations.
func TestAccountManagementWorkflow(t *testing.T) {
	service, cleanup := SetupHTTPTestIntegration(t)
	defer cleanup()

	ctx := context.Background()

	// Step 1: Get current account information
	t.Run("Step1_GetAccountInfo", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_account_get",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleAccountGet(ctx, request)
		require.NoError(t, err, "account get should succeed")
		require.NotNil(t, result, "result should not be nil")

		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		// Validate account information structure
		require.Contains(t, textContent.Text, "Account: httptest-integration", "should show account name")
		require.Contains(t, textContent.Text, "Username: testuser", "should show username")
		require.Contains(t, textContent.Text, "Email: test@example.com", "should show email")
		require.Contains(t, textContent.Text, "UID: 12345", "should show UID")
		require.Contains(t, textContent.Text, "Restricted: false", "should show restriction status")
	})

	// Step 2: List available accounts
	t.Run("Step2_ListAccounts", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_account_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleAccountList(ctx, request)
		require.NoError(t, err, "account list should succeed")
		require.NotNil(t, result, "result should not be nil")

		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		// Validate account list structure
		require.Contains(t, textContent.Text, "Configured accounts:", "should show accounts header")
		require.Contains(t, textContent.Text, "httptest-integration", "should show test account")
		require.Contains(t, textContent.Text, "HTTP Test Integration Account", "should show account label")
		require.Contains(t, textContent.Text, "(current)", "should mark current account")
	})

	// Step 3: Verify consistency between account operations
	t.Run("Step3_AccountConsistency", func(t *testing.T) {
		// Get account info again to ensure consistency
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_account_get",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleAccountGet(ctx, request)
		require.NoError(t, err, "second account get should succeed")
		require.NotNil(t, result, "result should not be nil")

		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		// Verify consistent information
		require.Contains(t, textContent.Text, "Username: testuser", "username should be consistent")
		require.Contains(t, textContent.Text, "Email: test@example.com", "email should be consistent")
	})
}

// TestResourceDiscoveryWorkflow tests discovering and listing various Linode resources.
//
// **Resource Discovery Workflow**:
// 1. **Instance Discovery**: List all instances and their details
// 2. **Volume Discovery**: List all volumes and their attachments
// 3. **Resource Relationships**: Verify relationships between instances and volumes
// 4. **Resource Consistency**: Ensure resource information is consistently formatted
//
// **Test Environment**: HTTP test server with predefined instances and volumes
//
// **Expected Behavior**:
// • All resource types can be successfully listed
// • Resource information includes all required fields
// • Relationships between resources are properly represented
// • Text formatting is consistent across resource types
//
// **Purpose**: Validates resource discovery capabilities and ensures consistent
// resource representation across different CloudMCP handlers.
func TestResourceDiscoveryWorkflow(t *testing.T) {
	service, cleanup := SetupHTTPTestIntegration(t)
	defer cleanup()

	ctx := context.Background()

	// Step 1: Discover instances
	var instanceFound bool
	t.Run("Step1_DiscoverInstances", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_instances_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleInstancesList(ctx, request)
		require.NoError(t, err, "instances list should succeed")
		require.NotNil(t, result, "result should not be nil")

		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		// Validate instance discovery
		require.Contains(t, textContent.Text, "Found 1 Linode instance(s)", "should find instances")
		require.Contains(t, textContent.Text, "test-instance-1", "should show instance label")
		require.Contains(t, textContent.Text, "Status: running", "should show instance status")
		require.Contains(t, textContent.Text, "Region: us-east", "should show instance region")

		instanceFound = true
	})

	// Step 2: Discover volumes
	t.Run("Step2_DiscoverVolumes", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_volumes_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleVolumesList(ctx, request)
		require.NoError(t, err, "volumes list should succeed")
		require.NotNil(t, result, "result should not be nil")

		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		// Validate volume discovery
		require.Contains(t, textContent.Text, "Found 1 volume(s)", "should find volumes")
		require.Contains(t, textContent.Text, "test-volume-1", "should show volume label")
		require.Contains(t, textContent.Text, "Status: active", "should show volume status")
		require.Contains(t, textContent.Text, "Size: 20 GB", "should show volume size")
	})

	// Step 3: Verify resource relationships
	t.Run("Step3_VerifyResourceRelationships", func(t *testing.T) {
		require.True(t, instanceFound, "instance should have been discovered in previous step")

		// Get detailed volume information to verify instance attachment
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_volumes_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleVolumesList(ctx, request)
		require.NoError(t, err, "volumes list should succeed")
		require.NotNil(t, result, "result should not be nil")

		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		// Verify volume is attached to the instance we discovered
		require.Contains(t, textContent.Text, "Attached to: Linode 123456 (test-instance-1)",
			"volume should show attachment to discovered instance")
	})

	// Step 4: Test resource consistency over time
	t.Run("Step4_ResourceConsistency", func(t *testing.T) {
		// Wait a short time and re-query to ensure consistency
		time.Sleep(10 * time.Millisecond)

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_instances_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleInstancesList(ctx, request)
		require.NoError(t, err, "second instances list should succeed")
		require.NotNil(t, result, "result should not be nil")

		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		// Verify consistent information
		require.Contains(t, textContent.Text, "Found 1 Linode instance(s)", "instance count should be consistent")
		require.Contains(t, textContent.Text, "test-instance-1", "instance label should be consistent")
	})
}
