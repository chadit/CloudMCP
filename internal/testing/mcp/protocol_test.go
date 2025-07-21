package mcp_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/testing/mcp"
)

// TestJSONRPCMessageValidation validates JSON-RPC 2.0 message format compliance.
// This test ensures all messages exchanged follow the JSON-RPC 2.0 specification.
//
// **JSON-RPC 2.0 Requirements:**
// • All messages must have "jsonrpc": "2.0"
// • Requests must have "method" field
// • Responses must have either "result" or "error" (but not both)
// • Error objects must have "code" and "message" fields
// • ID field must match between request and response
//
// **Test Coverage:**
// • Valid request message validation
// • Valid response message validation
// • Valid error message validation
// • Invalid message detection
// • Edge cases and boundary conditions
//
// **Purpose:** Ensure CloudMCP exchanges messages that are fully JSON-RPC 2.0 compliant.
func TestJSONRPCMessageValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		description string
		message     string
		expectValid bool
		errorMsg    string
	}{
		{
			name:        "ValidRequest",
			description: "Test valid JSON-RPC 2.0 request message",
			message: `{
				"jsonrpc": "2.0",
				"method": "tools/list",
				"id": 1
			}`,
			expectValid: true,
		},
		{
			name:        "ValidRequestWithParams",
			description: "Test valid request with parameters",
			message: `{
				"jsonrpc": "2.0",
				"method": "tools/call",
				"params": {
					"name": "health_check",
					"arguments": {}
				},
				"id": 2
			}`,
			expectValid: true,
		},
		{
			name:        "ValidNotification",
			description: "Test valid notification (request without ID)",
			message: `{
				"jsonrpc": "2.0",
				"method": "initialized"
			}`,
			expectValid: true,
		},
		{
			name:        "ValidSuccessResponse",
			description: "Test valid success response",
			message: `{
				"jsonrpc": "2.0",
				"result": {
					"tools": []
				},
				"id": 1
			}`,
			expectValid: true,
		},
		{
			name:        "ValidErrorResponse",
			description: "Test valid error response",
			message: `{
				"jsonrpc": "2.0",
				"error": {
					"code": -32601,
					"message": "Method not found"
				},
				"id": 1
			}`,
			expectValid: true,
		},
		{
			name:        "InvalidMissingJSONRPC",
			description: "Test invalid message missing jsonrpc field",
			message: `{
				"method": "tools/list",
				"id": 1
			}`,
			expectValid: false,
			errorMsg:    "invalid JSON-RPC version",
		},
		{
			name:        "InvalidWrongJSONRPCVersion",
			description: "Test invalid message with wrong jsonrpc version",
			message: `{
				"jsonrpc": "1.0",
				"method": "tools/list",
				"id": 1
			}`,
			expectValid: false,
			errorMsg:    "invalid JSON-RPC version",
		},
		{
			name:        "InvalidRequestMissingMethod",
			description: "Test invalid request missing method field",
			message: `{
				"jsonrpc": "2.0",
				"id": 1
			}`,
			expectValid: false,
			errorMsg:    "response must have either result or error",
		},
		{
			name:        "InvalidResponseMissingResultAndError",
			description: "Test invalid response missing both result and error",
			message: `{
				"jsonrpc": "2.0",
				"id": 1
			}`,
			expectValid: false,
			errorMsg:    "response must have either result or error",
		},
		{
			name:        "InvalidResponseBothResultAndError",
			description: "Test invalid response with both result and error",
			message: `{
				"jsonrpc": "2.0",
				"result": {},
				"error": {
					"code": -1,
					"message": "Error"
				},
				"id": 1
			}`,
			expectValid: false,
			errorMsg:    "response cannot have both result and error",
		},
		{
			name:        "InvalidResponseMissingID",
			description: "Test invalid response missing ID field",
			message: `{
				"jsonrpc": "2.0",
				"result": {}
			}`,
			expectValid: false,
			errorMsg:    "response must have id field",
		},
		{
			name:        "InvalidMethodNotString",
			description: "Test invalid request with non-string method",
			message: `{
				"jsonrpc": "2.0",
				"method": 123,
				"id": 1
			}`,
			expectValid: false,
			errorMsg:    "method must be a string",
		},
		{
			name:        "InvalidJSON",
			description: "Test invalid JSON format",
			message:     `{"jsonrpc": "2.0", "method": "test", "id":}`,
			expectValid: false,
			errorMsg:    "invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := mcp.ValidateJSONRPCMessage([]byte(tt.message))

			if tt.expectValid {
				assert.NoError(t, err, "Message should be valid: %s", tt.description)
			} else {
				require.Error(t, err, "Message should be invalid: %s", tt.description)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg, "Error message should contain expected text")
				}
			}
		})
	}
}

// TestJSONRPCResponseValidation validates JSON-RPC response validation logic.
func TestJSONRPCResponseValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		response    *mcp.JSONRPCResponse
		expectedID  interface{}
		expectValid bool
		errorMsg    string
	}{
		{
			name: "ValidResponse",
			response: &mcp.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      123,
				Result:  json.RawMessage(`{"status": "ok"}`),
			},
			expectedID:  123,
			expectValid: true,
		},
		{
			name: "ValidErrorResponse",
			response: &mcp.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      "test-id",
				Error: &mcp.JSONRPCError{
					Code:    -32601,
					Message: "Method not found",
				},
			},
			expectedID:  "test-id",
			expectValid: true,
		},
		{
			name: "InvalidJSONRPCVersion",
			response: &mcp.JSONRPCResponse{
				JSONRPC: "1.0",
				ID:      123,
				Result:  json.RawMessage(`{}`),
			},
			expectedID:  123,
			expectValid: false,
			errorMsg:    "invalid JSON-RPC version",
		},
		{
			name: "InvalidIDMismatch",
			response: &mcp.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      456,
				Result:  json.RawMessage(`{}`),
			},
			expectedID:  123,
			expectValid: false,
			errorMsg:    "response ID mismatch",
		},
		{
			name: "InvalidBothResultAndError",
			response: &mcp.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      123,
				Result:  json.RawMessage(`{}`),
				Error: &mcp.JSONRPCError{
					Code:    -1,
					Message: "Error",
				},
			},
			expectedID:  123,
			expectValid: false,
			errorMsg:    "response cannot have both result and error",
		},
		{
			name: "InvalidNeitherResultNorError",
			response: &mcp.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      123,
			},
			expectedID:  123,
			expectValid: false,
			errorMsg:    "response must have either result or error",
		},
		{
			name: "InvalidErrorZeroCode",
			response: &mcp.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      123,
				Error: &mcp.JSONRPCError{
					Code:    0,
					Message: "Error",
				},
			},
			expectedID:  123,
			expectValid: false,
			errorMsg:    "error code cannot be zero",
		},
		{
			name: "InvalidErrorEmptyMessage",
			response: &mcp.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      123,
				Error: &mcp.JSONRPCError{
					Code:    -1,
					Message: "",
				},
			},
			expectedID:  123,
			expectValid: false,
			errorMsg:    "error message cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			responseJSON, err := json.Marshal(tt.response)
			require.NoError(t, err, "Failed to marshal test response")

			// Use the appropriate validation function based on whether we need ID validation
			if tt.expectedID != nil {
				// Use ValidateJSONRPCResponse for tests that require ID validation
				err = mcp.ValidateJSONRPCResponse(responseJSON, tt.expectedID)
			} else {
				// Use ValidateJSONRPCMessage for tests that don't require ID validation
				err = mcp.ValidateJSONRPCMessage(responseJSON)
			}

			if tt.expectValid {
				assert.NoError(t, err, "Response should be valid")
			} else {
				require.Error(t, err, "Response should be invalid")
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg, "Error message should contain expected text")
				}
			}
		})
	}
}

// TestJSONRPCErrorCodes validates standard JSON-RPC error codes.
func TestJSONRPCErrorCodes(t *testing.T) {
	t.Parallel()

	standardErrorCodes := map[int]string{
		-32700: "Parse error",
		-32600: "Invalid Request",
		-32601: "Method not found",
		-32602: "Invalid params",
		-32603: "Internal error",
	}

	for code, description := range standardErrorCodes {
		t.Run(description, func(t *testing.T) {
			t.Parallel()

			errorResponse := mcp.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      1,
				Error: &mcp.JSONRPCError{
					Code:    code,
					Message: description,
				},
			}

			responseJSON, err := json.Marshal(errorResponse)
			require.NoError(t, err, "Failed to marshal error response")

			err = mcp.ValidateJSONRPCMessage(responseJSON)
			assert.NoError(t, err, "Standard error codes should be valid")
		})
	}
}

// TestJSONRPCBatchRequests validates JSON-RPC batch request handling.
func TestJSONRPCBatchRequests(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		batch       string
		expectValid bool
		errorMsg    string
	}{
		{
			name: "ValidBatch",
			batch: `[
				{
					"jsonrpc": "2.0",
					"method": "tools/list",
					"id": 1
				},
				{
					"jsonrpc": "2.0",
					"method": "initialize",
					"params": {"protocolVersion": "2024-11-05"},
					"id": 2
				}
			]`,
			expectValid: true,
		},
		{
			name: "EmptyBatch",
			batch:       `[]`,
			expectValid: true, // Empty batch is technically valid JSON
		},
		{
			name: "InvalidBatchWithInvalidMessage",
			batch: `[
				{
					"jsonrpc": "2.0",
					"method": "tools/list",
					"id": 1
				},
				{
					"jsonrpc": "1.0",
					"method": "invalid",
					"id": 2
				}
			]`,
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var batch []json.RawMessage
			err := json.Unmarshal([]byte(tt.batch), &batch)
			require.NoError(t, err, "Failed to parse batch JSON")

			allValid := true
			var lastError error

			for _, message := range batch {
				if err := mcp.ValidateJSONRPCMessage(message); err != nil {
					allValid = false
					lastError = err
					break
				}
			}

			if tt.expectValid {
				assert.True(t, allValid, "All messages in batch should be valid")
			} else {
				assert.False(t, allValid, "Batch should contain invalid messages")
				if tt.errorMsg != "" && lastError != nil {
					assert.Contains(t, lastError.Error(), tt.errorMsg, "Error should contain expected message")
				}
			}
		})
	}
}

// BenchmarkJSONRPCValidation benchmarks JSON-RPC message validation performance.
func BenchmarkJSONRPCValidation(b *testing.B) {
	testMessage := []byte(`{
		"jsonrpc": "2.0",
		"method": "tools/call",
		"params": {
			"name": "health_check",
			"arguments": {}
		},
		"id": 1
	}`)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := mcp.ValidateJSONRPCMessage(testMessage)
		if err != nil {
			b.Fatalf("Validation failed: %v", err)
		}
	}
}