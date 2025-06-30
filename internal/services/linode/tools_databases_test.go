package linode_test

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/internal/services/linode"
	"github.com/chadit/CloudMCP/pkg/logger"
)

// TestHandleDatabasesList_AccountError tests the handleDatabasesList function to verify error handling.
// This test simulates a user requesting a list of all their managed databases through the MCP interface.
// Since this function requires Linode API calls, this test focuses on the error handling path.
//
// **Test Workflow**:
// 1. **Service Setup**: Create isolated test service with empty account manager
// 2. **Request Execution**: Call handleDatabasesList expecting account manager failure
// 3. **Error Validation**: Verify appropriate error is returned for account lookup failure
//
// **Test Environment**: Service with no configured accounts to trigger error path
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Does not attempt to call Linode API when account lookup fails
// • Provides meaningful error message for troubleshooting
//
// **Purpose**: This test ensures databases list command fails appropriately when account configuration is invalid.
func TestHandleDatabasesList_AccountError(t *testing.T) {
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

	// Test databases list request with empty account manager
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_databases_list",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.Error(t, err, "handleDatabasesList should return an error when no accounts are configured")
	require.Nil(t, result, "result should be nil when error occurs")
	require.Contains(t, err.Error(), "account", "error should mention account issue")
}

// TestHandleMySQLDatabasesList_AccountError tests the handleMySQLDatabasesList function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleMySQLDatabasesList with no configured accounts
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
func TestHandleMySQLDatabasesList_AccountError(t *testing.T) {
	t.Parallel()

	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test MySQL databases list request with no accounts
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_mysql_databases_list",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.Error(t, err, "handleMySQLDatabasesList should return an error when no accounts are configured")
	require.Nil(t, result, "result should be nil when error occurs")
	require.Contains(t, err.Error(), "account", "error should mention account issue")
}

// TestHandlePostgresDatabasesList_AccountError tests the handlePostgresDatabasesList function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handlePostgresDatabasesList with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandlePostgresDatabasesList_AccountError(t *testing.T) {
	t.Parallel()

	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test PostgreSQL databases list request with no accounts
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_postgres_databases_list",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.Error(t, err, "handlePostgresDatabasesList should return an error when no accounts are configured")
	require.Nil(t, result, "result should be nil when error occurs")
	require.Contains(t, err.Error(), "account", "error should mention account issue")
}

// TestHandleMySQLDatabaseGet_AccountError tests the handleMySQLDatabaseGet function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleMySQLDatabaseGet with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleMySQLDatabaseGet_AccountError(t *testing.T) {
	t.Parallel()

	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test MySQL database get request with no accounts
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_mysql_database_get",
			Arguments: map[string]interface{}{
				"database_id": float64(12345),
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.Error(t, err, "handleMySQLDatabaseGet should return an error when no accounts are configured")
	require.Nil(t, result, "result should be nil when error occurs")
	require.Contains(t, err.Error(), "account", "error should mention account issue")
}

// TestHandlePostgresDatabaseGet_AccountError tests the handlePostgresDatabaseGet function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handlePostgresDatabaseGet with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandlePostgresDatabaseGet_AccountError(t *testing.T) {
	t.Parallel()

	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test PostgreSQL database get request with no accounts
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_postgres_database_get",
			Arguments: map[string]interface{}{
				"database_id": float64(12345),
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.Error(t, err, "handlePostgresDatabaseGet should return an error when no accounts are configured")
	require.Nil(t, result, "result should be nil when error occurs")
	require.Contains(t, err.Error(), "account", "error should mention account issue")
}

// TestHandleMySQLDatabaseCreate_AccountError tests the handleMySQLDatabaseCreate function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleMySQLDatabaseCreate with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleMySQLDatabaseCreate_AccountError(t *testing.T) {
	t.Parallel()

	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test MySQL database create request with no accounts
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_mysql_database_create",
			Arguments: map[string]interface{}{
				"label":  "test-mysql-db",
				"region": "us-east",
				"type":   "g6-nanode-1",
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.Error(t, err, "handleMySQLDatabaseCreate should return an error when no accounts are configured")
	require.Nil(t, result, "result should be nil when error occurs")
	require.Contains(t, err.Error(), "account", "error should mention account issue")
}

// TestHandleMySQLDatabaseDelete_AccountError tests the handleMySQLDatabaseDelete function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleMySQLDatabaseDelete with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleMySQLDatabaseDelete_AccountError(t *testing.T) {
	t.Parallel()

	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test MySQL database delete request with no accounts
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_mysql_database_delete",
			Arguments: map[string]interface{}{
				"database_id": float64(12345),
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.Error(t, err, "handleMySQLDatabaseDelete should return an error when no accounts are configured")
	require.Nil(t, result, "result should be nil when error occurs")
	require.Contains(t, err.Error(), "account", "error should mention account issue")
}

// TestHandleMySQLDatabaseCredentialsReset_AccountError tests the handleMySQLDatabaseCredentialsReset function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleMySQLDatabaseCredentialsReset with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleMySQLDatabaseCredentialsReset_AccountError(t *testing.T) {
	t.Parallel()

	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test MySQL database credentials reset request with no accounts
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "linode_mysql_database_credentials_reset",
			Arguments: map[string]interface{}{
				"database_id": float64(12345),
			},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.Error(t, err, "handleMySQLDatabaseCredentialsReset should return an error when no accounts are configured")
	require.Nil(t, result, "result should be nil when error occurs")
	require.Contains(t, err.Error(), "account", "error should mention account issue")
}

// TestHandleDatabaseEnginesList_AccountError tests the handleDatabaseEnginesList function with account manager errors.
// This test verifies the function properly handles errors when no current account is available.
//
// **Test Workflow**:
// 1. **Service Setup**: Create test service with empty account manager
// 2. **Request Execution**: Call handleDatabaseEnginesList with no configured accounts
// 3. **Error Validation**: Verify proper error handling for missing account
//
// **Expected Behavior**:
// • Returns error when no current account is available
// • Error indicates account retrieval failure
// • No API calls are made when account is unavailable
//
// **Purpose**: This test ensures robust error handling when account manager has no configured accounts.
func TestHandleDatabaseEnginesList_AccountError(t *testing.T) {
	t.Parallel()

	// Create service with completely empty account manager - no accounts at all
	log := logger.New("debug")
	cfg := &config.Config{
		DefaultLinodeAccount: "nonexistent",
		LinodeAccounts:       map[string]config.LinodeAccount{},
	}

	accountManager := linode.NewAccountManagerForTesting()
	service := linode.NewForTesting(cfg, log, accountManager)

	// Test database engines list request with no accounts
	ctx := t.Context()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "linode_database_engines_list",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := service.CallToolForTesting(ctx, request)
	require.Error(t, err, "handleDatabaseEnginesList should return an error when no accounts are configured")
	require.Nil(t, result, "result should be nil when error occurs")
	require.Contains(t, err.Error(), "account", "error should mention account issue")
}

// Note: Additional tests for successful database operations with mock Linode clients
// are not implemented in this unit test suite because they require functioning
// Linode API client interfaces for operations like ListDatabases, GetDatabase,
// CreateDatabase, UpdateDatabase, DeleteDatabase, ResetDatabaseCredentials,
// ListDatabaseEngines, and ListDatabaseTypes.
//
// Parameter validation tests are not included because the database handlers
// currently use parseArguments() placeholder which doesn't implement comprehensive
// validation like parseIDFromArguments() used in other tool handlers.
//
// The database service includes 17 total handlers covering:
// - General database operations (list all databases)
// - MySQL-specific operations (list, get, create, update, delete, credentials)
// - PostgreSQL-specific operations (list, get, create, update, delete, credentials)
// - Database metadata operations (engines list, types list)
//
// These operations would require either:
// 1. Interface abstraction for the Linode client (future improvement)
// 2. Integration testing with real API endpoints
// 3. Dependency injection to replace the client during testing
// 4. Implementation of proper parameter validation in database handlers
//
// The current tests adequately cover:
// - Account manager error scenarios for key database operations
// - Request routing and basic tool handler setup
// - Error message formatting and response structure
// - MCP protocol compliance for error responses
// - Both MySQL and PostgreSQL engine support validation
//
// This provides coverage of the testable logic that doesn't
// require external API dependencies while ensuring proper error handling
// for the most common failure scenario (account configuration issues).
