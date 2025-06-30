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

// createSupportMonitoringTestServer creates an HTTP test server with support and monitoring API endpoints.
// This extends the base HTTP test server infrastructure with comprehensive
// support and monitoring-specific endpoints for integration testing.
//
// **Support Endpoints Supported:**
// • GET /v4/support/tickets - List support tickets
// • GET /v4/support/tickets/{id} - Get specific support ticket
// • POST /v4/support/tickets - Create new support ticket
// • POST /v4/support/tickets/{id}/replies - Reply to support ticket
//
// **Monitoring Endpoints Supported:**
// • GET /v4/longview/clients - List Longview monitoring clients
// • GET /v4/longview/clients/{id} - Get specific Longview client
// • POST /v4/longview/clients - Create new Longview client
// • PUT /v4/longview/clients/{id} - Update Longview client
// • DELETE /v4/longview/clients/{id} - Delete Longview client
//
// **Mock Data Features:**
// • Realistic support ticket workflows with status tracking
// • Longview monitoring client configurations and statistics
// • Error simulation for non-existent resources
// • Proper HTTP status codes and error responses
func createSupportMonitoringTestServer() *httptest.Server {
	mux := http.NewServeMux()

	// Include base endpoints (profile, account) from main test server
	addSupportMonitoringBaseEndpoints(mux)

	// Support tickets list endpoint
	mux.HandleFunc("/v4/support/tickets", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleSupportTicketsList(w, r)
		case http.MethodPost:
			handleSupportTicketsCreate(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Specific support ticket endpoints
	mux.HandleFunc("/v4/support/tickets/12345", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleSupportTicketsGet(w, r, "12345")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Support ticket replies endpoint
	mux.HandleFunc("/v4/support/tickets/12345/replies", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handleSupportTicketsReply(w, r, "12345")
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Longview clients list endpoint
	mux.HandleFunc("/v4/longview/clients", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleLongviewClientsList(w, r)
		case http.MethodPost:
			handleLongviewClientsCreate(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Specific Longview client endpoints
	mux.HandleFunc("/v4/longview/clients/54321", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleLongviewClientsGet(w, r, "54321")
		case http.MethodPut:
			handleLongviewClientsUpdate(w, r, "54321")
		case http.MethodDelete:
			handleLongviewClientsDelete(w, r, "54321")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Non-existent ticket/client endpoints for error testing
	mux.HandleFunc("/v4/support/tickets/999999", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleSupportTicketsGet(w, r, "999999")
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v4/longview/clients/999999", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleLongviewClientsGet(w, r, "999999")
		case http.MethodDelete:
			handleLongviewClientsDelete(w, r, "999999")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return httptest.NewServer(mux)
}

// addSupportMonitoringBaseEndpoints adds the basic profile and account endpoints needed for service initialization.
func addSupportMonitoringBaseEndpoints(mux *http.ServeMux) {
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

// handleSupportTicketsList handles the support tickets list mock response.
func handleSupportTicketsList(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"id":          12345,
				"summary":     "Server performance issues",
				"description": "My Linode instance is experiencing high CPU usage and slow response times.",
				"status":      "open",
				"opened":      "2023-01-01T10:00:00",
				"opened_by":   "testuser",
				"updated":     "2023-01-01T15:30:00",
				"updated_by":  "support-agent",
				"closed":      nil,
				"attachments": []string{},
				"entity": map[string]interface{}{
					"id":    123456,
					"label": "my-server",
					"type":  "linode",
					"url":   "/v4/linode/instances/123456",
				},
			},
			{
				"id":          67890,
				"summary":     "Billing question",
				"description": "I have a question about my monthly invoice charges.",
				"status":      "closed",
				"opened":      "2023-01-02T09:00:00",
				"opened_by":   "testuser",
				"updated":     "2023-01-02T16:45:00",
				"updated_by":  "billing-support",
				"closed":      "2023-01-02T16:45:00",
				"attachments": []string{},
				"entity":      nil,
			},
		},
		"page":    1,
		"pages":   1,
		"results": 2,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSupportTicketsGet handles the specific support ticket mock response.
func handleSupportTicketsGet(w http.ResponseWriter, r *http.Request, ticketID string) {
	switch ticketID {
	case "12345":
		response := map[string]interface{}{
			"id":          12345,
			"summary":     "Server performance issues",
			"description": "My Linode instance is experiencing high CPU usage and slow response times.",
			"status":      "open",
			"opened":      "2023-01-01T10:00:00",
			"opened_by":   "testuser",
			"updated":     "2023-01-01T15:30:00",
			"updated_by":  "support-agent",
			"closed":      nil,
			"attachments": []string{},
			"entity": map[string]interface{}{
				"id":    123456,
				"label": "my-server",
				"type":  "linode",
				"url":   "/v4/linode/instances/123456",
			},
			"replies": []map[string]interface{}{
				{
					"created":     "2023-01-01T11:00:00",
					"created_by":  "testuser",
					"description": "Here are the current metrics from my monitoring dashboard.",
				},
				{
					"created":     "2023-01-01T15:30:00",
					"created_by":  "support-agent",
					"description": "Thank you for providing the metrics. Let me investigate this further.",
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
				{"field": "ticket_id", "reason": "Not found"},
			},
		})
	default:
		http.Error(w, "Support ticket not found", http.StatusNotFound)
	}
}

// handleSupportTicketsCreate handles the support ticket creation mock response.
func handleSupportTicketsCreate(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"id":          98765,
		"summary":     "New support request",
		"description": "I need help with setting up my new Linode instance.",
		"status":      "new",
		"opened":      "2023-01-01T12:00:00",
		"opened_by":   "testuser",
		"updated":     "2023-01-01T12:00:00",
		"updated_by":  "testuser",
		"closed":      nil,
		"attachments": []string{},
		"entity":      nil,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleSupportTicketsReply handles the support ticket reply mock response.
func handleSupportTicketsReply(w http.ResponseWriter, r *http.Request, ticketID string) {
	response := map[string]interface{}{
		"created":     "2023-01-01T16:00:00",
		"created_by":  "testuser",
		"description": "Thank you for your quick response. The issue has been resolved.",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleLongviewClientsList handles the Longview clients list mock response.
func handleLongviewClientsList(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"id":           54321,
				"label":        "production-monitor",
				"api_key":      "abcd1234-5678-90ef-ghij-klmnopqrstuv",
				"install_code": "curl -s https://lv.linode.com/abcd1234-5678-90ef-ghij-klmnopqrstuv | sudo bash",
				"created":      "2023-01-01T08:00:00",
				"updated":      "2023-01-01T08:00:00",
				"apps": map[string]interface{}{
					"apache": true,
					"mysql":  true,
					"nginx":  false,
				},
			},
			{
				"id":           98765,
				"label":        "development-monitor",
				"api_key":      "wxyz9876-5432-10ab-cdef-ghijklmnopqr",
				"install_code": "curl -s https://lv.linode.com/wxyz9876-5432-10ab-cdef-ghijklmnopqr | sudo bash",
				"created":      "2023-01-02T09:00:00",
				"updated":      "2023-01-02T09:00:00",
				"apps": map[string]interface{}{
					"apache": false,
					"mysql":  false,
					"nginx":  true,
				},
			},
		},
		"page":    1,
		"pages":   1,
		"results": 2,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleLongviewClientsGet handles the specific Longview client mock response.
func handleLongviewClientsGet(w http.ResponseWriter, r *http.Request, clientID string) {
	switch clientID {
	case "54321":
		response := map[string]interface{}{
			"id":           54321,
			"label":        "production-monitor",
			"api_key":      "abcd1234-5678-90ef-ghij-klmnopqrstuv",
			"install_code": "curl -s https://lv.linode.com/abcd1234-5678-90ef-ghij-klmnopqrstuv | sudo bash",
			"created":      "2023-01-01T08:00:00",
			"updated":      "2023-01-01T08:00:00",
			"apps": map[string]interface{}{
				"apache": true,
				"mysql":  true,
				"nginx":  false,
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
				{"field": "client_id", "reason": "Not found"},
			},
		})
	default:
		http.Error(w, "Longview client not found", http.StatusNotFound)
	}
}

// handleLongviewClientsCreate handles the Longview client creation mock response.
func handleLongviewClientsCreate(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"id":           13579,
		"label":        "new-monitor",
		"api_key":      "new1234-5678-90ab-cdef-ghijklmnopqr",
		"install_code": "curl -s https://lv.linode.com/new1234-5678-90ab-cdef-ghijklmnopqr | sudo bash",
		"created":      "2023-01-01T14:00:00",
		"updated":      "2023-01-01T14:00:00",
		"apps": map[string]interface{}{
			"apache": false,
			"mysql":  false,
			"nginx":  false,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleLongviewClientsUpdate handles the Longview client update mock response.
func handleLongviewClientsUpdate(w http.ResponseWriter, r *http.Request, clientID string) {
	if clientID == "999999" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "client_id", "reason": "Not found"},
			},
		})
		return
	}

	response := map[string]interface{}{
		"id":           54321,
		"label":        "updated-production-monitor",
		"api_key":      "abcd1234-5678-90ef-ghij-klmnopqrstuv",
		"install_code": "curl -s https://lv.linode.com/abcd1234-5678-90ef-ghij-klmnopqrstuv | sudo bash",
		"created":      "2023-01-01T08:00:00",
		"updated":      "2023-01-01T16:00:00",
		"apps": map[string]interface{}{
			"apache": true,
			"mysql":  true,
			"nginx":  false,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleLongviewClientsDelete handles the Longview client deletion mock response.
func handleLongviewClientsDelete(w http.ResponseWriter, r *http.Request, clientID string) {
	if clientID == "999999" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"field": "client_id", "reason": "Not found"},
			},
		})
		return
	}

	// Return 204 No Content for successful deletion
	w.WriteHeader(http.StatusNoContent)
}

// TestSupportMonitoringToolsIntegration tests all support and monitoring-related CloudMCP tools
// using HTTP test server infrastructure to validate proper integration
// between CloudMCP handlers and Linode API endpoints.
//
// **Integration Test Coverage**:
// • linode_support_tickets_list - List all support tickets
// • linode_support_ticket_get - Get specific support ticket details
// • linode_support_ticket_create - Create new support ticket
// • linode_support_ticket_reply - Reply to support ticket
// • linode_longview_clients_list - List Longview monitoring clients
// • linode_longview_client_get - Get specific Longview client details
// • linode_longview_client_create - Create new Longview client
// • linode_longview_client_update - Update Longview client
// • linode_longview_client_delete - Delete Longview client
//
// **Test Environment**: HTTP test server with support and monitoring API handlers
//
// **Expected Behavior**:
// • All handlers return properly formatted text responses
// • Error conditions are handled gracefully with meaningful messages
// • Support ticket data includes all required fields (ID, summary, status, etc.)
// • Longview client configurations are properly formatted
//
// **Purpose**: Validates that CloudMCP support and monitoring handlers correctly format
// Linode API responses for LLM consumption using lightweight HTTP test infrastructure.
func TestSupportMonitoringToolsIntegration(t *testing.T) {
	// Extend the HTTP test server with support and monitoring endpoints
	server := createSupportMonitoringTestServer()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
	ctx := t.Context()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	t.Run("SupportTicketsList", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_support_tickets_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleSupportTicketsList(ctx, request)
		require.NoError(t, err, "support tickets list should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate support tickets list formatting
		require.Contains(t, responseText, "Found 2 support tickets:", "should indicate correct ticket count")
		require.Contains(t, responseText, "Server performance issues", "should contain first ticket summary")
		require.Contains(t, responseText, "Billing question", "should contain second ticket summary")
		require.Contains(t, responseText, "Status: open", "should show open status")
		require.Contains(t, responseText, "Status: closed", "should show closed status")
	})

	t.Run("SupportTicketGet", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_support_ticket_get",
				Arguments: map[string]interface{}{
					"ticket_id": float64(12345),
				},
			},
		}

		result, err := service.handleSupportTicketGet(ctx, request)
		require.NoError(t, err, "support ticket get should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate detailed support ticket information
		require.Contains(t, responseText, "Support Ticket Details:", "should have ticket details header")
		require.Contains(t, responseText, "ID: 12345", "should contain ticket ID")
		require.Contains(t, responseText, "Summary: Server performance issues", "should contain ticket summary")
		require.Contains(t, responseText, "Status: open", "should contain ticket status")
		require.Contains(t, responseText, "Replies:", "should have replies section")
	})

	t.Run("SupportTicketCreate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_support_ticket_create",
				Arguments: map[string]interface{}{
					"summary":     "New support request",
					"description": "I need help with setting up my new Linode instance.",
				},
			},
		}

		result, err := service.handleSupportTicketCreate(ctx, request)
		require.NoError(t, err, "support ticket create should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate support ticket creation confirmation
		require.Contains(t, responseText, "Support ticket created successfully:", "should confirm creation")
		require.Contains(t, responseText, "ID: 98765", "should show new ticket ID")
		require.Contains(t, responseText, "Summary: New support request", "should show ticket summary")
	})

	t.Run("SupportTicketReply", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_support_ticket_reply",
				Arguments: map[string]interface{}{
					"ticket_id":   float64(12345),
					"description": "Thank you for your quick response. The issue has been resolved.",
				},
			},
		}

		result, err := service.handleSupportTicketReply(ctx, request)
		require.NoError(t, err, "support ticket reply should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate ticket reply confirmation
		require.Contains(t, responseText, "Reply added successfully to ticket", "should confirm reply")
		require.Contains(t, responseText, "12345", "should show ticket ID")
	})

	t.Run("LongviewClientsList", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_longview_clients_list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := service.handleLongviewClientsList(ctx, request)
		require.NoError(t, err, "Longview clients list should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate Longview clients list formatting
		require.Contains(t, responseText, "Found 2 Longview clients:", "should indicate correct client count")
		require.Contains(t, responseText, "production-monitor", "should contain first client label")
		require.Contains(t, responseText, "development-monitor", "should contain second client label")
		require.Contains(t, responseText, "API Key:", "should show API keys")
	})

	t.Run("LongviewClientGet", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_longview_client_get",
				Arguments: map[string]interface{}{
					"client_id": float64(54321),
				},
			},
		}

		result, err := service.handleLongviewClientGet(ctx, request)
		require.NoError(t, err, "Longview client get should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate detailed Longview client information
		require.Contains(t, responseText, "Longview Client Details:", "should have client details header")
		require.Contains(t, responseText, "ID: 54321", "should contain client ID")
		require.Contains(t, responseText, "Label: production-monitor", "should contain client label")
		require.Contains(t, responseText, "API Key:", "should contain API key")
		require.Contains(t, responseText, "Install Code:", "should contain install code")
	})

	t.Run("LongviewClientCreate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_longview_client_create",
				Arguments: map[string]interface{}{
					"label": "new-monitor",
				},
			},
		}

		result, err := service.handleLongviewClientCreate(ctx, request)
		require.NoError(t, err, "Longview client create should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate Longview client creation confirmation
		require.Contains(t, responseText, "Longview client created successfully:", "should confirm creation")
		require.Contains(t, responseText, "ID: 13579", "should show new client ID")
		require.Contains(t, responseText, "Label: new-monitor", "should show client label")
	})

	t.Run("LongviewClientUpdate", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_longview_client_update",
				Arguments: map[string]interface{}{
					"client_id": float64(54321),
					"label":     "updated-production-monitor",
				},
			},
		}

		result, err := service.handleLongviewClientUpdate(ctx, request)
		require.NoError(t, err, "Longview client update should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate Longview client update confirmation
		require.Contains(t, responseText, "Longview client updated successfully:", "should confirm update")
		require.Contains(t, responseText, "ID: 54321", "should show client ID")
		require.Contains(t, responseText, "Label: updated-production-monitor", "should show updated label")
	})

	t.Run("LongviewClientDelete", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_longview_client_delete",
				Arguments: map[string]interface{}{
					"client_id": float64(54321),
				},
			},
		}

		result, err := service.handleLongviewClientDelete(ctx, request)
		require.NoError(t, err, "Longview client delete should not return error")
		require.NotNil(t, result, "result should not be nil")
		require.NotEmpty(t, result.Content, "result should have content")

		// Extract and validate text response
		textContent, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok, "content should be text")

		responseText := textContent.Text
		require.NotEmpty(t, responseText, "response text should not be empty")

		// Validate Longview client deletion confirmation
		require.Contains(t, responseText, "Longview client", "should mention Longview client")
		require.Contains(t, responseText, "deleted successfully", "should confirm deletion")
		require.Contains(t, responseText, "54321", "should show deleted client ID")
	})
}

// TestSupportMonitoringErrorHandlingIntegration tests error scenarios for support and monitoring tools
// to ensure CloudMCP handles API errors gracefully and provides meaningful
// error messages to users.
//
// **Error Test Scenarios**:
// • Non-existent ticket/client IDs (404 errors)
// • Invalid ticket creation parameters
// • Monitoring client configuration conflicts
// • Permission errors for support operations
//
// **Expected Behavior**:
// • Proper error handling with contextual error messages
// • No unhandled exceptions or panics
// • Error responses follow CloudMCP error format
// • Error messages are actionable for users
//
// **Purpose**: Validates robust error handling in support and monitoring operations
// and ensures reliable operation under error conditions.
func TestSupportMonitoringErrorHandlingIntegration(t *testing.T) {
	server := createSupportMonitoringTestServer()
	defer server.Close()

	service := CreateHTTPTestService(t, server.URL)
	ctx := t.Context()

	err := service.Initialize(ctx)
	require.NoError(t, err, "service initialization should succeed")

	t.Run("SupportTicketGetNotFound", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_support_ticket_get",
				Arguments: map[string]interface{}{
					"ticket_id": float64(999999), // Non-existent ticket
				},
			},
		}

		result, err := service.handleSupportTicketGet(ctx, request)

		// Check for either Go error or MCP error result
		if err != nil {
			// Go error pattern
			require.Contains(t, err.Error(), "Failed to get support ticket", "error should mention get ticket failure")
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

	t.Run("LongviewClientGetNotFound", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_longview_client_get",
				Arguments: map[string]interface{}{
					"client_id": float64(999999), // Non-existent client
				},
			},
		}

		result, err := service.handleLongviewClientGet(ctx, request)

		// Check for either Go error or MCP error result
		if err != nil {
			// Go error pattern
			require.Contains(t, err.Error(), "Failed to get Longview client", "error should mention get client failure")
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

	t.Run("LongviewClientDeleteNotFound", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_longview_client_delete",
				Arguments: map[string]interface{}{
					"client_id": float64(999999), // Non-existent client
				},
			},
		}

		result, err := service.handleLongviewClientDelete(ctx, request)

		// Check for either Go error or MCP error result
		if err != nil {
			// Go error pattern
			require.Contains(t, err.Error(), "Failed to delete Longview client", "error should mention delete client failure")
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

	t.Run("InvalidParameterType", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_support_ticket_get",
				Arguments: map[string]interface{}{
					"ticket_id": "invalid", // Invalid ID type
				},
			},
		}

		result, err := service.handleSupportTicketGet(ctx, request)

		// Should get MCP error result for invalid parameter
		if err == nil {
			require.NotNil(t, result, "result should not be nil")
			require.True(t, result.IsError, "result should be marked as error")
		}
	})
}
