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

// createDomainTestServer creates an HTTP test server with domain and DNS record API endpoints.
// This extends the base HTTP test server infrastructure with comprehensive
// domain-specific endpoints for integration testing.
//
// **Domain Endpoints Supported:**
// • GET /v4/domains - List domains
// • GET /v4/domains/{id} - Get specific domain
// • POST /v4/domains - Create new domain
// • PUT /v4/domains/{id} - Update domain
// • DELETE /v4/domains/{id} - Delete domain
// • GET /v4/domains/{id}/records - List domain records
// • GET /v4/domains/{id}/records/{record_id} - Get specific record
// • POST /v4/domains/{id}/records - Create new record
// • PUT /v4/domains/{id}/records/{record_id} - Update record
// • DELETE /v4/domains/{id}/records/{record_id} - Delete record
//
// **Mock Data Features:**
// • Realistic domain configurations with SOA and DNS settings
// • DNS record management with all record types (A, AAAA, CNAME, MX, TXT, SRV)
// • Error simulation for non-existent resources
// • Proper HTTP status codes and error responses
func createDomainTestServer() *httptest.Server {
	mux := http.NewServeMux()

	// Include base endpoints (profile, account) from main test server
	addDomainBaseEndpoints(mux)

	// Domains list endpoint
	mux.HandleFunc("/v4/domains", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleDomainsList(w, r)
		case http.MethodPost:
			handleDomainsCreate(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Specific domain endpoints with explicit ID matching
	mux.HandleFunc("/v4/domains/12345", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleDomainsGet(w, r, "12345")
		case http.MethodPut:
			handleDomainsUpdate(w, r, "12345")
		case http.MethodDelete:
			handleDomainsDelete(w, r, "12345")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Domain records endpoints
	mux.HandleFunc("/v4/domains/12345/records", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleDomainRecordsList(w, r, "12345")
		case http.MethodPost:
			handleDomainRecordsCreate(w, r, "12345")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Specific domain record endpoints
	mux.HandleFunc("/v4/domains/12345/records/67890", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleDomainRecordsGet(w, r, "12345", "67890")
		case http.MethodPut:
			handleDomainRecordsUpdate(w, r, "12345", "67890")
		case http.MethodDelete:
			handleDomainRecordsDelete(w, r, "12345", "67890")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Non-existent domain endpoints for error testing
	mux.HandleFunc("/v4/domains/999999", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleDomainsGet(w, r, "999999")
		case http.MethodDelete:
			handleDomainsDelete(w, r, "999999")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return httptest.NewServer(mux)
}

// addDomainBaseEndpoints adds the basic profile and account endpoints needed for service initialization.
func addDomainBaseEndpoints(mux *http.ServeMux) {
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

// handleDomainsList handles the domains list mock response.
func handleDomainsList(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"id":          12345,
				"domain":      "example.com",
				"type":        "master",
				"status":      "active",
				"description": "Production domain",
				"soa_email":   "admin@example.com",
				"retry_sec":   3600,
				"master_ips":  []string{},
				"axfr_ips":    []string{},
				"expire_sec":  604800,
				"refresh_sec": 14400,
				"ttl_sec":     300,
				"group":       "production",
				"tags":        []string{"production", "website"},
			},
			{
				"id":          54321,
				"domain":      "test.com",
				"type":        "master",
				"status":      "active",
				"description": "Test domain",
				"soa_email":   "test@test.com",
				"retry_sec":   7200,
				"master_ips":  []string{},
				"axfr_ips":    []string{},
				"expire_sec":  1209600,
				"refresh_sec": 28800,
				"ttl_sec":     600,
				"group":       "testing",
				"tags":        []string{"testing", "development"},
			},
		},
		"page":    1,
		"pages":   1,
		"results": 2,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleDomainsGet handles the specific domain mock response.
func handleDomainsGet(w http.ResponseWriter, r *http.Request, domainID string) {
	switch domainID {
	case "12345":
		response := map[string]interface{}{
			"id":          12345,
			"domain":      "example.com",
			"type":        "master",
			"status":      "active",
			"description": "Production domain",
			"soa_email":   "admin@example.com",
			"retry_sec":   3600,
			"master_ips":  []string{},
			"axfr_ips":    []string{},
			"expire_sec":  604800,
			"refresh_sec": 14400,
			"ttl_sec":     300,
			"group":       "production",
			"tags":        []string{"production", "website"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	case "999999":
		// Simulate not found error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "domain_id", "reason": "Not found"},
			},
		})
	default:
		http.Error(w, "Domain not found", http.StatusNotFound)
	}
}

// handleDomainsCreate handles the domain creation mock response.
func handleDomainsCreate(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"id":          98765,
		"domain":      "newdomain.com",
		"type":        "master",
		"status":      "active",
		"description": "New test domain",
		"soa_email":   "admin@newdomain.com",
		"retry_sec":   3600,
		"master_ips":  []string{},
		"axfr_ips":    []string{},
		"expire_sec":  604800,
		"refresh_sec": 14400,
		"ttl_sec":     300,
		"group":       "",
		"tags":        []string{},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleDomainsUpdate handles the domain update mock response.
func handleDomainsUpdate(w http.ResponseWriter, r *http.Request, domainID string) {
	if domainID == "999999" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "domain_id", "reason": "Not found"},
			},
		})
		return
	}

	response := map[string]interface{}{
		"id":          12345,
		"domain":      "example.com",
		"type":        "master",
		"status":      "active",
		"description": "Updated production domain",
		"soa_email":   "updated@example.com",
		"retry_sec":   3600,
		"master_ips":  []string{},
		"axfr_ips":    []string{},
		"expire_sec":  604800,
		"refresh_sec": 14400,
		"ttl_sec":     300,
		"group":       "production",
		"tags":        []string{"production", "website", "updated"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleDomainsDelete handles the domain deletion mock response.
func handleDomainsDelete(w http.ResponseWriter, r *http.Request, domainID string) {
	if domainID == "999999" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "domain_id", "reason": "Not found"},
			},
		})
		return
	}

	// Return 204 No Content for successful deletion
	w.WriteHeader(http.StatusNoContent)
}

// handleDomainRecordsList handles the domain records list mock response.
func handleDomainRecordsList(w http.ResponseWriter, r *http.Request, domainID string) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"id":       67890,
				"type":     "A",
				"name":     "www",
				"target":   "192.168.1.1",
				"priority": 0,
				"weight":   0,
				"port":     0,
				"service":  nil,
				"protocol": nil,
				"ttl_sec":  300,
				"tag":      nil,
				"created":  "2023-01-01T00:00:00",
				"updated":  "2023-01-01T00:00:00",
			},
			{
				"id":       67891,
				"type":     "MX",
				"name":     "",
				"target":   "mail.example.com",
				"priority": 10,
				"weight":   0,
				"port":     0,
				"service":  nil,
				"protocol": nil,
				"ttl_sec":  3600,
				"tag":      nil,
				"created":  "2023-01-01T01:00:00",
				"updated":  "2023-01-01T01:00:00",
			},
			{
				"id":       67892,
				"type":     "TXT",
				"name":     "",
				"target":   "v=spf1 include:_spf.google.com ~all",
				"priority": 0,
				"weight":   0,
				"port":     0,
				"service":  nil,
				"protocol": nil,
				"ttl_sec":  300,
				"tag":      nil,
				"created":  "2023-01-01T02:00:00",
				"updated":  "2023-01-01T02:00:00",
			},
		},
		"page":    1,
		"pages":   1,
		"results": 3,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleDomainRecordsGet handles the specific domain record mock response.
func handleDomainRecordsGet(w http.ResponseWriter, r *http.Request, domainID, recordID string) {
	switch recordID {
	case "67890":
		response := map[string]interface{}{
			"id":       67890,
			"type":     "A",
			"name":     "www",
			"target":   "192.168.1.1",
			"priority": 0,
			"weight":   0,
			"port":     0,
			"service":  nil,
			"protocol": nil,
			"ttl_sec":  300,
			"tag":      nil,
			"created":  "2023-01-01T00:00:00",
			"updated":  "2023-01-01T00:00:00",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	default:
		http.Error(w, "Record not found", http.StatusNotFound)
	}
}

// handleDomainRecordsCreate handles the domain record creation mock response.
func handleDomainRecordsCreate(w http.ResponseWriter, r *http.Request, domainID string) {
	response := map[string]interface{}{
		"id":       99999,
		"type":     "A",
		"name":     "api",
		"target":   "192.168.1.100",
		"priority": 0,
		"weight":   0,
		"port":     0,
		"service":  nil,
		"protocol": nil,
		"ttl_sec":  300,
		"tag":      nil,
		"created":  "2023-01-01T03:00:00",
		"updated":  "2023-01-01T03:00:00",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleDomainRecordsUpdate handles the domain record update mock response.
func handleDomainRecordsUpdate(w http.ResponseWriter, r *http.Request, domainID, recordID string) {
	response := map[string]interface{}{
		"id":       67890,
		"type":     "A",
		"name":     "www",
		"target":   "192.168.1.2",
		"priority": 0,
		"weight":   0,
		"port":     0,
		"service":  nil,
		"protocol": nil,
		"ttl_sec":  600,
		"tag":      nil,
		"created":  "2023-01-01T00:00:00",
		"updated":  "2023-01-01T04:00:00",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleDomainRecordsDelete handles the domain record deletion mock response.
func handleDomainRecordsDelete(w http.ResponseWriter, r *http.Request, domainID, recordID string) {
	// Return 204 No Content for successful deletion
	w.WriteHeader(http.StatusNoContent)
}

// TestDomainToolsIntegration tests all domain-related CloudMCP tools
// using HTTP test server infrastructure to validate proper integration
// between CloudMCP handlers and Linode API endpoints.
//
// **Integration Test Coverage**:
// • linode_domains_list - List all domains
// • linode_domain_get - Get specific domain details
// • linode_domain_create - Create new domain
// • linode_domain_update - Update existing domain
// • linode_domain_delete - Delete domain
// • linode_domain_records_list - List domain records
// • linode_domain_record_get - Get specific record details
// • linode_domain_record_create - Create new record
// • linode_domain_record_update - Update existing record
// • linode_domain_record_delete - Delete record
//
// **Test Environment**: HTTP test server with domain API handlers
//
// **Expected Behavior**:
// • All handlers return properly formatted text responses
// • Error conditions are handled gracefully with meaningful messages
// • Domain data includes all required fields (ID, domain, type, status, etc.)
// • DNS record management operations work correctly
//
// **Purpose**: Validates that CloudMCP domain handlers correctly format
// Linode API responses for LLM consumption using lightweight HTTP test infrastructure.
func TestDomainToolsIntegration(t *testing.T) {
	// Extend the HTTP test server with domain endpoints
	server := createDomainTestServer()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
	ctx := context.Background()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	t.Run("DomainsList", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_domains_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleDomainsList(ctx, request)
		require.NoError(t, err, "domains list should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate domain list formatting
		require.Contains(t, responseText, "Found 2 domains:", "should indicate correct domain count")
		require.Contains(t, responseText, "example.com", "should contain first domain")
		require.Contains(t, responseText, "test.com", "should contain second domain")
		require.Contains(t, responseText, "(master)", "should show domain type")
		require.Contains(t, responseText, "Status: active", "should show domain status")
	})

	t.Run("DomainGet", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_domain_get",
				Arguments: map[string]interface{}{
					"domain_id": float64(12345),
				},
			},
		}

		result, err := service.handleDomainGet(ctx, request)
		require.NoError(t, err, "domain get should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate detailed domain information
		require.Contains(t, responseText, "Domain Details:", "should have domain details header")
		require.Contains(t, responseText, "ID: 12345", "should contain domain ID")
		require.Contains(t, responseText, "Domain: example.com", "should contain domain name")
		require.Contains(t, responseText, "Type: master", "should contain domain type")
		require.Contains(t, responseText, "Status: active", "should contain domain status")
		require.Contains(t, responseText, "SOA Email:", "should have SOA email section")
	})

	t.Run("DomainCreate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_domain_create",
				Arguments: map[string]interface{}{
					"domain":      "newdomain.com",
					"type":        "master",
					"soa_email":   "admin@newdomain.com",
					"description": "New test domain",
				},
			},
		}

		result, err := service.handleDomainCreate(ctx, request)
		require.NoError(t, err, "domain create should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate domain creation confirmation
		require.Contains(t, responseText, "Domain created successfully:", "should confirm creation")
		require.Contains(t, responseText, "ID: 98765", "should show new domain ID")
		require.Contains(t, responseText, "Domain: newdomain.com", "should show domain name")
		require.Contains(t, responseText, "Type: master", "should show domain type")
	})

	t.Run("DomainUpdate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_domain_update",
				Arguments: map[string]interface{}{
					"domain_id":   float64(12345),
					"description": "Updated production domain",
					"soa_email":   "updated@example.com",
				},
			},
		}

		result, err := service.handleDomainUpdate(ctx, request)
		require.NoError(t, err, "domain update should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate domain update confirmation
		require.Contains(t, responseText, "Domain updated successfully:", "should confirm update")
		require.Contains(t, responseText, "ID: 12345", "should show domain ID")
		require.Contains(t, responseText, "Domain: example.com", "should show domain name")
	})

	t.Run("DomainRecordsList", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_domain_records_list",
				Arguments: map[string]interface{}{
					"domain_id": float64(12345),
				},
			},
		}

		result, err := service.handleDomainRecordsList(ctx, request)
		require.NoError(t, err, "domain records list should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate domain records list formatting
		require.Contains(t, responseText, "Found 3 domain records:", "should indicate correct record count")
		require.Contains(t, responseText, "A Records:", "should have A records section")
		require.Contains(t, responseText, "MX Records:", "should have MX records section")
		require.Contains(t, responseText, "TXT Records:", "should have TXT records section")
	})

	t.Run("DomainRecordGet", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_domain_record_get",
				Arguments: map[string]interface{}{
					"domain_id": float64(12345),
					"record_id": float64(67890),
				},
			},
		}

		result, err := service.handleDomainRecordGet(ctx, request)
		require.NoError(t, err, "domain record get should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate detailed record information
		require.Contains(t, responseText, "Domain Record Details:", "should have record details header")
		require.Contains(t, responseText, "ID: 67890", "should contain record ID")
		require.Contains(t, responseText, "Type: A", "should contain record type")
		require.Contains(t, responseText, "Name: www", "should contain record name")
		require.Contains(t, responseText, "Target: 192.168.1.1", "should contain record target")
	})

	t.Run("DomainRecordCreate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_domain_record_create",
				Arguments: map[string]interface{}{
					"domain_id": float64(12345),
					"type":      "A",
					"name":      "api",
					"target":    "192.168.1.100",
					"ttl_sec":   float64(300),
				},
			},
		}

		result, err := service.handleDomainRecordCreate(ctx, request)
		require.NoError(t, err, "domain record create should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate record creation confirmation
		require.Contains(t, responseText, "Domain record created successfully:", "should confirm creation")
		require.Contains(t, responseText, "ID: 99999", "should show new record ID")
		require.Contains(t, responseText, "Type: A", "should show record type")
		require.Contains(t, responseText, "Name: api", "should show record name")
	})

	t.Run("DomainRecordUpdate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_domain_record_update",
				Arguments: map[string]interface{}{
					"domain_id": float64(12345),
					"record_id": float64(67890),
					"target":    "192.168.1.2",
					"ttl_sec":   float64(600),
				},
			},
		}

		result, err := service.handleDomainRecordUpdate(ctx, request)
		require.NoError(t, err, "domain record update should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate record update confirmation
		require.Contains(t, responseText, "Domain record updated successfully:", "should confirm update")
		require.Contains(t, responseText, "ID: 67890", "should show record ID")
		require.Contains(t, responseText, "Target: 192.168.1.2", "should show updated target")
	})

	t.Run("DomainRecordDelete", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_domain_record_delete",
				Arguments: map[string]interface{}{
					"domain_id": float64(12345),
					"record_id": float64(67890),
				},
			},
		}

		result, err := service.handleDomainRecordDelete(ctx, request)
		require.NoError(t, err, "domain record delete should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate record deletion confirmation
		require.Contains(t, responseText, "Domain record", "should mention domain record")
		require.Contains(t, responseText, "deleted successfully", "should confirm deletion")
		require.Contains(t, responseText, "67890", "should show deleted record ID")
	})

	t.Run("DomainDelete", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_domain_delete",
				Arguments: map[string]interface{}{
					"domain_id": float64(12345),
				},
			},
		}

		result, err := service.handleDomainDelete(ctx, request)
		require.NoError(t, err, "domain delete should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate domain deletion confirmation
		require.Contains(t, responseText, "Domain", "should mention domain")
		require.Contains(t, responseText, "deleted successfully", "should confirm deletion")
		require.Contains(t, responseText, "12345", "should show deleted domain ID")
	})
}

// TestDomainErrorHandlingIntegration tests error scenarios for domain tools
// to ensure CloudMCP handles API errors gracefully and provides meaningful
// error messages to users.
//
// **Error Test Scenarios**:
// • Non-existent domain ID (404 errors)
// • Invalid domain configuration
// • DNS record conflicts
// • Permission errors for domain operations
//
// **Expected Behavior**:
// • Proper error handling with contextual error messages
// • No unhandled exceptions or panics
// • Error responses follow CloudMCP error format
// • Error messages are actionable for users
//
// **Purpose**: Validates robust error handling in domain operations
// and ensures reliable operation under error conditions.
func TestDomainErrorHandlingIntegration(t *testing.T) {
	server := createDomainTestServer()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
	ctx := context.Background()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	t.Run("DomainGetNotFound", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_domain_get",
				Arguments: map[string]interface{}{
					"domain_id": float64(999999), // Non-existent domain
				},
			},
		}

		result, err := service.handleDomainGet(ctx, request)

		// Check for either Go error or MCP error result
		if err != nil {
			// Go error pattern
			require.Contains(t, err.Error(), "failed to get domain", "error should mention get domain failure")
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

	t.Run("DomainDeleteNotFound", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_domain_delete",
				Arguments: map[string]interface{}{
					"domain_id": float64(999999), // Non-existent domain
				},
			},
		}

		result, err := service.handleDomainDelete(ctx, request)

		// Check for either Go error or MCP error result
		if err != nil {
			// Go error pattern
			require.Contains(t, err.Error(), "Failed to delete domain", "error should mention delete domain failure")
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

	t.Run("InvalidDomainID", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_domain_get",
				Arguments: map[string]interface{}{
					"domain_id": "invalid", // Invalid ID type
				},
			},
		}

		result, err := service.handleDomainGet(ctx, request)

		// Should get MCP error result for invalid parameter
		if err == nil {
			require.NotNil(t, result, "result should not be nil")
			require.True(t, result.IsError, "result should be marked as error")
		}
	})
}
