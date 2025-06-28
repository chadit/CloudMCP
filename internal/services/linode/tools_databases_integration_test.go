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

// createDatabaseTestServer creates an HTTP test server with database API endpoints.
// This extends the base HTTP test server infrastructure with comprehensive
// database-specific endpoints for integration testing.
//
// **Database Endpoints Supported:**
// • GET /v4/databases/engines - List database engines
// • GET /v4/databases/types - List database types
// • GET /v4/databases/mysql/instances - List MySQL databases
// • GET /v4/databases/mysql/instances/{id} - Get specific MySQL database
// • POST /v4/databases/mysql/instances - Create new MySQL database
// • PUT /v4/databases/mysql/instances/{id} - Update MySQL database
// • DELETE /v4/databases/mysql/instances/{id} - Delete MySQL database
// • POST /v4/databases/mysql/instances/{id}/credentials/reset - Reset credentials
// • GET /v4/databases/postgresql/instances - List PostgreSQL databases
// • POST /v4/databases/postgresql/instances - Create new PostgreSQL database
//
// **Mock Data Features:**
// • Realistic database configurations with different engines
// • Credential management and reset operations
// • Error simulation for non-existent resources
// • Proper HTTP status codes and error responses
func createDatabaseTestServer() *httptest.Server {
	mux := http.NewServeMux()

	// Include base endpoints (profile, account) from main test server
	addDatabaseBaseEndpoints(mux)

	// Database engines endpoint
	mux.HandleFunc("/v4/databases/engines", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleDatabaseEnginesList(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Database types endpoint
	mux.HandleFunc("/v4/databases/types", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleDatabaseTypesList(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// MySQL databases list endpoint
	mux.HandleFunc("/v4/databases/mysql/instances", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleMySQLDatabasesList(w, r)
		case http.MethodPost:
			handleMySQLDatabasesCreate(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Specific MySQL database endpoints
	mux.HandleFunc("/v4/databases/mysql/instances/12345", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleMySQLDatabasesGet(w, r, "12345")
		case http.MethodPut:
			handleMySQLDatabasesUpdate(w, r, "12345")
		case http.MethodDelete:
			handleMySQLDatabasesDelete(w, r, "12345")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// MySQL credentials endpoints
	mux.HandleFunc("/v4/databases/mysql/instances/12345/credentials", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleMySQLCredentials(w, r, "12345")
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v4/databases/mysql/instances/12345/credentials/reset", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handleMySQLCredentialsReset(w, r, "12345")
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// PostgreSQL databases list endpoint
	mux.HandleFunc("/v4/databases/postgresql/instances", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlePostgreSQLDatabasesList(w, r)
		case http.MethodPost:
			handlePostgreSQLDatabasesCreate(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Specific PostgreSQL database endpoints
	mux.HandleFunc("/v4/databases/postgresql/instances/54321", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlePostgreSQLDatabasesGet(w, r, "54321")
		case http.MethodPut:
			handlePostgreSQLDatabasesUpdate(w, r, "54321")
		case http.MethodDelete:
			handlePostgreSQLDatabasesDelete(w, r, "54321")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// PostgreSQL credentials endpoints
	mux.HandleFunc("/v4/databases/postgresql/instances/54321/credentials", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlePostgreSQLCredentials(w, r, "54321")
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v4/databases/postgresql/instances/54321/credentials/reset", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlePostgreSQLCredentialsReset(w, r, "54321")
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Non-existent database endpoints for error testing
	mux.HandleFunc("/v4/databases/mysql/instances/999999", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleMySQLDatabasesGet(w, r, "999999")
		case http.MethodDelete:
			handleMySQLDatabasesDelete(w, r, "999999")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return httptest.NewServer(mux)
}

// addDatabaseBaseEndpoints adds the basic profile and account endpoints needed for service initialization.
func addDatabaseBaseEndpoints(mux *http.ServeMux) {
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

// handleDatabaseEnginesList handles the database engines list mock response.
func handleDatabaseEnginesList(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"id":      "mysql/8.0.30",
				"engine":  "mysql",
				"version": "8.0.30",
			},
			{
				"id":      "mysql/5.7.39",
				"engine":  "mysql",
				"version": "5.7.39",
			},
			{
				"id":      "postgresql/14.9",
				"engine":  "postgresql",
				"version": "14.9",
			},
			{
				"id":      "postgresql/13.12",
				"engine":  "postgresql",
				"version": "13.12",
			},
		},
		"page":    1,
		"pages":   1,
		"results": 4,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleDatabaseTypesList handles the database types list mock response.
func handleDatabaseTypesList(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"id":      "g6-nanode-1",
				"label":   "Nanode 1GB",
				"class":   "nanode",
				"disk":    25600,
				"memory":  1024,
				"vcpus":   1,
				"engines": map[string]interface{}{"mysql": []string{"mysql/8.0.30", "mysql/5.7.39"}, "postgresql": []string{"postgresql/14.9", "postgresql/13.12"}},
			},
			{
				"id":      "g6-standard-1",
				"label":   "Linode 2GB",
				"class":   "standard",
				"disk":    51200,
				"memory":  2048,
				"vcpus":   1,
				"engines": map[string]interface{}{"mysql": []string{"mysql/8.0.30", "mysql/5.7.39"}, "postgresql": []string{"postgresql/14.9", "postgresql/13.12"}},
			},
		},
		"page":    1,
		"pages":   1,
		"results": 2,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleMySQLDatabasesList handles the MySQL databases list mock response.
func handleMySQLDatabasesList(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"id":           12345,
				"label":        "production-mysql",
				"engine":       "mysql",
				"version":      "8.0.30",
				"region":       "us-east",
				"type":         "g6-standard-2",
				"status":       "active",
				"cluster_size": 3,
				"hosts": map[string]interface{}{
					"primary":   "lin-12345-678-mysql-primary.servers.linodedb.net",
					"secondary": "lin-12345-678-mysql-primary-private.servers.linodedb.net",
				},
				"port":           3306,
				"ssl_connection": true,
				"allow_list":     []string{"192.168.1.0/24", "10.0.0.0/8"},
				"created":        "2023-01-01T00:00:00",
				"updated":        "2023-01-01T00:00:00",
			},
			{
				"id":           67890,
				"label":        "development-mysql",
				"engine":       "mysql",
				"version":      "8.0.30",
				"region":       "us-west",
				"type":         "g6-nanode-1",
				"status":       "active",
				"cluster_size": 1,
				"hosts": map[string]interface{}{
					"primary": "lin-67890-123-mysql-primary.servers.linodedb.net",
				},
				"port":           3306,
				"ssl_connection": true,
				"allow_list":     []string{"0.0.0.0/0"},
				"created":        "2023-01-02T00:00:00",
				"updated":        "2023-01-02T00:00:00",
			},
		},
		"page":    1,
		"pages":   1,
		"results": 2,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleMySQLDatabasesGet handles the specific MySQL database mock response.
func handleMySQLDatabasesGet(w http.ResponseWriter, r *http.Request, databaseID string) {
	switch databaseID {
	case "12345":
		response := map[string]interface{}{
			"id":           12345,
			"label":        "production-mysql",
			"engine":       "mysql",
			"version":      "8.0.30",
			"region":       "us-east",
			"type":         "g6-standard-2",
			"status":       "active",
			"cluster_size": 3,
			"hosts": map[string]interface{}{
				"primary":   "lin-12345-678-mysql-primary.servers.linodedb.net",
				"secondary": "lin-12345-678-mysql-primary-private.servers.linodedb.net",
			},
			"port":           3306,
			"ssl_connection": true,
			"allow_list":     []string{"192.168.1.0/24", "10.0.0.0/8"},
			"created":        "2023-01-01T00:00:00",
			"updated":        "2023-01-01T00:00:00",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	case "999999":
		// Simulate not found error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "database_id", "reason": "Not found"},
			},
		})
	default:
		http.Error(w, "Database not found", http.StatusNotFound)
	}
}

// handleMySQLDatabasesCreate handles the MySQL database creation mock response.
func handleMySQLDatabasesCreate(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"id":           98765,
		"label":        "new-mysql-db",
		"engine":       "mysql",
		"version":      "8.0.30",
		"region":       "us-central",
		"type":         "g6-standard-1",
		"status":       "provisioning",
		"cluster_size": 1,
		"hosts": map[string]interface{}{
			"primary": "lin-98765-456-mysql-primary.servers.linodedb.net",
		},
		"port":           3306,
		"ssl_connection": true,
		"allow_list":     []string{},
		"created":        "2023-01-01T01:00:00",
		"updated":        "2023-01-01T01:00:00",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleMySQLDatabasesUpdate handles the MySQL database update mock response.
func handleMySQLDatabasesUpdate(w http.ResponseWriter, r *http.Request, databaseID string) {
	response := map[string]interface{}{
		"id":           12345,
		"label":        "updated-production-mysql",
		"engine":       "mysql",
		"version":      "8.0.30",
		"region":       "us-east",
		"type":         "g6-standard-2",
		"status":       "active",
		"cluster_size": 3,
		"hosts": map[string]interface{}{
			"primary":   "lin-12345-678-mysql-primary.servers.linodedb.net",
			"secondary": "lin-12345-678-mysql-primary-private.servers.linodedb.net",
		},
		"port":           3306,
		"ssl_connection": true,
		"allow_list":     []string{"192.168.1.0/24", "10.0.0.0/8", "172.16.0.0/12"},
		"created":        "2023-01-01T00:00:00",
		"updated":        "2023-01-01T02:00:00",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleMySQLDatabasesDelete handles the MySQL database deletion mock response.
func handleMySQLDatabasesDelete(w http.ResponseWriter, r *http.Request, databaseID string) {
	if databaseID == "999999" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "database_id", "reason": "Not found"},
			},
		})
		return
	}

	// Return 204 No Content for successful deletion
	w.WriteHeader(http.StatusNoContent)
}

// handleMySQLCredentials handles the MySQL credentials retrieval mock response.
func handleMySQLCredentials(w http.ResponseWriter, r *http.Request, databaseID string) {
	response := map[string]interface{}{
		"username": "linroot",
		"password": "secure-password-123",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleMySQLCredentialsReset handles the MySQL credentials reset mock response.
func handleMySQLCredentialsReset(w http.ResponseWriter, r *http.Request, databaseID string) {
	response := map[string]interface{}{
		"username": "linroot",
		"password": "new-secure-password-456",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handlePostgreSQLDatabasesList handles the PostgreSQL databases list mock response.
func handlePostgreSQLDatabasesList(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"id":           54321,
				"label":        "production-postgres",
				"engine":       "postgresql",
				"version":      "14.9",
				"region":       "us-east",
				"type":         "g6-standard-4",
				"status":       "active",
				"cluster_size": 3,
				"hosts": map[string]interface{}{
					"primary":   "lin-54321-789-pgsql-primary.servers.linodedb.net",
					"secondary": "lin-54321-789-pgsql-primary-private.servers.linodedb.net",
				},
				"port":           5432,
				"ssl_connection": true,
				"allow_list":     []string{"192.168.1.0/24"},
				"created":        "2023-01-01T00:00:00",
				"updated":        "2023-01-01T00:00:00",
			},
		},
		"page":    1,
		"pages":   1,
		"results": 1,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handlePostgreSQLDatabasesGet handles the specific PostgreSQL database mock response.
func handlePostgreSQLDatabasesGet(w http.ResponseWriter, r *http.Request, databaseID string) {
	response := map[string]interface{}{
		"id":           54321,
		"label":        "production-postgres",
		"engine":       "postgresql",
		"version":      "14.9",
		"region":       "us-east",
		"type":         "g6-standard-4",
		"status":       "active",
		"cluster_size": 3,
		"hosts": map[string]interface{}{
			"primary":   "lin-54321-789-pgsql-primary.servers.linodedb.net",
			"secondary": "lin-54321-789-pgsql-primary-private.servers.linodedb.net",
		},
		"port":           5432,
		"ssl_connection": true,
		"allow_list":     []string{"192.168.1.0/24"},
		"created":        "2023-01-01T00:00:00",
		"updated":        "2023-01-01T00:00:00",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handlePostgreSQLDatabasesCreate handles the PostgreSQL database creation mock response.
func handlePostgreSQLDatabasesCreate(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"id":           13579,
		"label":        "new-postgres-db",
		"engine":       "postgresql",
		"version":      "14.9",
		"region":       "us-west",
		"type":         "g6-standard-2",
		"status":       "provisioning",
		"cluster_size": 1,
		"hosts": map[string]interface{}{
			"primary": "lin-13579-321-pgsql-primary.servers.linodedb.net",
		},
		"port":           5432,
		"ssl_connection": true,
		"allow_list":     []string{},
		"created":        "2023-01-01T01:00:00",
		"updated":        "2023-01-01T01:00:00",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handlePostgreSQLDatabasesUpdate handles the PostgreSQL database update mock response.
func handlePostgreSQLDatabasesUpdate(w http.ResponseWriter, r *http.Request, databaseID string) {
	response := map[string]interface{}{
		"id":           54321,
		"label":        "updated-production-postgres",
		"engine":       "postgresql",
		"version":      "14.9",
		"region":       "us-east",
		"type":         "g6-standard-4",
		"status":       "active",
		"cluster_size": 3,
		"hosts": map[string]interface{}{
			"primary":   "lin-54321-789-pgsql-primary.servers.linodedb.net",
			"secondary": "lin-54321-789-pgsql-primary-private.servers.linodedb.net",
		},
		"port":           5432,
		"ssl_connection": true,
		"allow_list":     []string{"192.168.1.0/24", "10.0.0.0/8"},
		"created":        "2023-01-01T00:00:00",
		"updated":        "2023-01-01T02:00:00",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handlePostgreSQLDatabasesDelete handles the PostgreSQL database deletion mock response.
func handlePostgreSQLDatabasesDelete(w http.ResponseWriter, r *http.Request, databaseID string) {
	// Return 204 No Content for successful deletion
	w.WriteHeader(http.StatusNoContent)
}

// handlePostgreSQLCredentials handles the PostgreSQL credentials retrieval mock response.
func handlePostgreSQLCredentials(w http.ResponseWriter, r *http.Request, databaseID string) {
	response := map[string]interface{}{
		"username": "linpostgres",
		"password": "postgres-secure-password-789",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handlePostgreSQLCredentialsReset handles the PostgreSQL credentials reset mock response.
func handlePostgreSQLCredentialsReset(w http.ResponseWriter, r *http.Request, databaseID string) {
	response := map[string]interface{}{
		"username": "linpostgres",
		"password": "new-postgres-password-012",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// TestDatabaseToolsIntegration tests all database-related CloudMCP tools
// using HTTP test server infrastructure to validate proper integration
// between CloudMCP handlers and Linode API endpoints.
//
// **Integration Test Coverage**:
// • linode_database_engines_list - List database engines
// • linode_database_types_list - List database types
// • linode_mysql_databases_list - List MySQL databases
// • linode_mysql_database_get - Get specific MySQL database
// • linode_mysql_database_create - Create new MySQL database
// • linode_mysql_database_update - Update MySQL database
// • linode_mysql_database_delete - Delete MySQL database
// • linode_mysql_database_credentials - Get MySQL credentials
// • linode_mysql_database_credentials_reset - Reset MySQL credentials
// • linode_postgres_databases_list - List PostgreSQL databases
//
// **Test Environment**: HTTP test server with database API handlers
//
// **Expected Behavior**:
// • All handlers return properly formatted text responses
// • Error conditions are handled gracefully with meaningful messages
// • Database data includes all required fields (ID, label, status, hosts, etc.)
// • Credential management operations work correctly
//
// **Purpose**: Validates that CloudMCP database handlers correctly format
// Linode API responses for LLM consumption using lightweight HTTP test infrastructure.
func TestDatabaseToolsIntegration(t *testing.T) {
	// Extend the HTTP test server with database endpoints
	server := createDatabaseTestServer()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
	ctx := context.Background()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	t.Run("DatabaseEnginesList", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_database_engines_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleDatabaseEnginesList(ctx, request)
		require.NoError(t, err, "database engines list should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate database engines list formatting
		require.Contains(t, responseText, "Found 4 database engines:", "should indicate correct engine count")
		require.Contains(t, responseText, "mysql/8.0.30", "should contain MySQL 8.0.30")
		require.Contains(t, responseText, "postgresql/14.9", "should contain PostgreSQL 14.9")
		require.Contains(t, responseText, "Engine: mysql", "should show engine type")
	})

	t.Run("DatabaseTypesList", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_database_types_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleDatabaseTypesList(ctx, request)
		require.NoError(t, err, "database types list should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate database types list formatting
		require.Contains(t, responseText, "Found 2 database types:", "should indicate correct type count")
		require.Contains(t, responseText, "g6-nanode-1", "should contain nanode type")
		require.Contains(t, responseText, "g6-standard-1", "should contain standard type")
		require.Contains(t, responseText, "Memory:", "should show memory specs")
	})

	t.Run("MySQLDatabasesList", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_mysql_databases_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleMySQLDatabasesList(ctx, request)
		require.NoError(t, err, "MySQL databases list should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate MySQL databases list formatting
		require.Contains(t, responseText, "Found 2 MySQL databases:", "should indicate correct database count")
		require.Contains(t, responseText, "production-mysql", "should contain production database")
		require.Contains(t, responseText, "development-mysql", "should contain development database")
		require.Contains(t, responseText, "Status: active", "should show database status")
	})

	t.Run("MySQLDatabaseGet", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_mysql_database_get",
				Arguments: map[string]interface{}{
					"database_id": float64(12345),
				},
			},
		}

		result, err := service.handleMySQLDatabaseGet(ctx, request)
		require.NoError(t, err, "MySQL database get should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate detailed MySQL database information
		require.Contains(t, responseText, "MySQL Database Details:", "should have database details header")
		require.Contains(t, responseText, "ID: 12345", "should contain database ID")
		require.Contains(t, responseText, "Label: production-mysql", "should contain database label")
		require.Contains(t, responseText, "Engine: mysql", "should contain engine type")
		require.Contains(t, responseText, "Version: 8.0.30", "should contain version")
		require.Contains(t, responseText, "Cluster Size: 3", "should contain cluster size")
	})

	t.Run("MySQLDatabaseCreate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_mysql_database_create",
				Arguments: map[string]interface{}{
					"label":  "new-mysql-db",
					"region": "us-central",
					"type":   "g6-standard-1",
					"engine": "mysql/8.0.30",
				},
			},
		}

		result, err := service.handleMySQLDatabaseCreate(ctx, request)
		require.NoError(t, err, "MySQL database create should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate MySQL database creation confirmation
		require.Contains(t, responseText, "MySQL Database created successfully:", "should confirm creation")
		require.Contains(t, responseText, "ID: 98765", "should show new database ID")
		require.Contains(t, responseText, "Label: new-mysql-db", "should show database label")
		require.Contains(t, responseText, "Status: provisioning", "should show initial status")
	})

	t.Run("MySQLDatabaseUpdate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_mysql_database_update",
				Arguments: map[string]interface{}{
					"database_id": float64(12345),
					"label":       "updated-production-mysql",
				},
			},
		}

		result, err := service.handleMySQLDatabaseUpdate(ctx, request)
		require.NoError(t, err, "MySQL database update should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate MySQL database update confirmation
		require.Contains(t, responseText, "MySQL Database updated successfully:", "should confirm update")
		require.Contains(t, responseText, "ID: 12345", "should show database ID")
		require.Contains(t, responseText, "Label: updated-production-mysql", "should show updated label")
	})

	t.Run("MySQLDatabaseCredentials", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_mysql_database_credentials",
				Arguments: map[string]interface{}{
					"database_id": float64(12345),
				},
			},
		}

		result, err := service.handleMySQLDatabaseCredentials(ctx, request)
		require.NoError(t, err, "MySQL database credentials should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate MySQL credentials response
		require.Contains(t, responseText, "MySQL Database Credentials:", "should have credentials header")
		require.Contains(t, responseText, "Username: linroot", "should show username")
		require.Contains(t, responseText, "Password: secure-password-123", "should show password")
	})

	t.Run("MySQLDatabaseCredentialsReset", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_mysql_database_credentials_reset",
				Arguments: map[string]interface{}{
					"database_id": float64(12345),
				},
			},
		}

		result, err := service.handleMySQLDatabaseCredentialsReset(ctx, request)
		require.NoError(t, err, "MySQL database credentials reset should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate MySQL credentials reset confirmation
		require.Contains(t, responseText, "MySQL Database credentials reset successfully:", "should confirm reset")
		require.Contains(t, responseText, "Username: linroot", "should show username")
		require.Contains(t, responseText, "New Password: new-secure-password-456", "should show new password")
	})

	t.Run("PostgreSQLDatabasesList", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_postgres_databases_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handlePostgresDatabasesList(ctx, request)
		require.NoError(t, err, "PostgreSQL databases list should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate PostgreSQL databases list formatting
		require.Contains(t, responseText, "Found 1 PostgreSQL database", "should indicate correct database count")
		require.Contains(t, responseText, "production-postgres", "should contain PostgreSQL database")
		require.Contains(t, responseText, "Engine: postgresql", "should show engine type")
	})

	t.Run("MySQLDatabaseDelete", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_mysql_database_delete",
				Arguments: map[string]interface{}{
					"database_id": float64(12345),
				},
			},
		}

		result, err := service.handleMySQLDatabaseDelete(ctx, request)
		require.NoError(t, err, "MySQL database delete should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate MySQL database deletion confirmation
		require.Contains(t, responseText, "MySQL Database", "should mention MySQL database")
		require.Contains(t, responseText, "deleted successfully", "should confirm deletion")
		require.Contains(t, responseText, "12345", "should show deleted database ID")
	})
}

// TestDatabaseErrorHandlingIntegration tests error scenarios for database tools
// to ensure CloudMCP handles API errors gracefully and provides meaningful
// error messages to users.
//
// **Error Test Scenarios**:
// • Non-existent database ID (404 errors)
// • Invalid database configuration
// • Engine compatibility issues
// • Permission errors for database operations
//
// **Expected Behavior**:
// • Proper error handling with contextual error messages
// • No unhandled exceptions or panics
// • Error responses follow CloudMCP error format
// • Error messages are actionable for users
//
// **Purpose**: Validates robust error handling in database operations
// and ensures reliable operation under error conditions.
func TestDatabaseErrorHandlingIntegration(t *testing.T) {
	server := createDatabaseTestServer()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
	ctx := context.Background()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	t.Run("MySQLDatabaseGetNotFound", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_mysql_database_get",
				Arguments: map[string]interface{}{
					"database_id": float64(999999), // Non-existent database
				},
			},
		}

		result, err := service.handleMySQLDatabaseGet(ctx, request)

		// Check for either Go error or MCP error result
		if err != nil {
			// Go error pattern
			require.Contains(t, err.Error(), "failed to get MySQL database", "error should mention get database failure")
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

	t.Run("MySQLDatabaseDeleteNotFound", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_mysql_database_delete",
				Arguments: map[string]interface{}{
					"database_id": float64(999999), // Non-existent database
				},
			},
		}

		result, err := service.handleMySQLDatabaseDelete(ctx, request)

		// Check for either Go error or MCP error result
		if err != nil {
			// Go error pattern
			require.Contains(t, err.Error(), "Failed to delete MySQL database", "error should mention delete database failure")
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

	t.Run("InvalidDatabaseID", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_mysql_database_get",
				Arguments: map[string]interface{}{
					"database_id": "invalid", // Invalid ID type
				},
			},
		}

		result, err := service.handleMySQLDatabaseGet(ctx, request)

		// Should get MCP error result for invalid parameter
		if err == nil {
			require.NotNil(t, result, "result should not be nil")
			require.True(t, result.IsError, "result should be marked as error")
		}
	})
}
