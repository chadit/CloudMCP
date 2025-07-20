// Package mcp provides MCP protocol compliance testing utilities for CloudMCP.
// This package implements mock clients and validation frameworks to ensure proper
// MCP protocol implementation and JSON-RPC 2.0 compliance.
package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// MockMCPClient simulates an MCP client for protocol testing.
// It provides methods to send MCP requests and validate responses according to
// the MCP specification and JSON-RPC 2.0 standard.
type MockMCPClient struct {
	reader    io.Reader
	writer    io.Writer
	responses []json.RawMessage
	requests  []json.RawMessage
	timeout   time.Duration
}

// JSONRPCRequest represents a JSON-RPC 2.0 request structure.
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response structure.
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error structure.
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCPInitializeRequest represents an MCP initialize request.
type MCPInitializeRequest struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      ClientInfo             `json:"clientInfo"`
}

// MCPInitializeResponse represents an MCP initialize response.
type MCPInitializeResponse struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ServerInfo      ServerInfo             `json:"serverInfo"`
}

// ClientInfo represents MCP client information.
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ServerInfo represents MCP server information.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ToolsListResponse represents the response from tools/list method.
type ToolsListResponse struct {
	Tools []mcp.Tool `json:"tools"`
}

// ToolsCallRequest represents a tools/call request.
type ToolsCallRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// NewMockMCPClient creates a new mock MCP client for testing.
func NewMockMCPClient(reader io.Reader, writer io.Writer) *MockMCPClient {
	return &MockMCPClient{
		reader:    reader,
		writer:    writer,
		responses: make([]json.RawMessage, 0),
		requests:  make([]json.RawMessage, 0),
		timeout:   5 * time.Second,
	}
}

// SetTimeout sets the timeout for client operations.
func (c *MockMCPClient) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// SendInitialize sends an MCP initialize request.
func (c *MockMCPClient) SendInitialize(ctx context.Context) (*MCPInitializeResponse, error) {
	request := MCPInitializeRequest{
		ProtocolVersion: "2024-11-05",
		Capabilities: map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		ClientInfo: ClientInfo{
			Name:    "CloudMCP-Test-Client",
			Version: "1.0.0",
		},
	}

	response, err := c.sendRequest(ctx, "initialize", request)
	if err != nil {
		return nil, fmt.Errorf("failed to send initialize request: %w", err)
	}

	var initResponse MCPInitializeResponse
	if err := json.Unmarshal(response.Result, &initResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal initialize response: %w", err)
	}

	return &initResponse, nil
}

// SendToolsList sends a tools/list request.
func (c *MockMCPClient) SendToolsList(ctx context.Context) (*ToolsListResponse, error) {
	response, err := c.sendRequest(ctx, "tools/list", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to send tools/list request: %w", err)
	}

	var toolsResponse ToolsListResponse
	if err := json.Unmarshal(response.Result, &toolsResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tools/list response: %w", err)
	}

	return &toolsResponse, nil
}

// SendToolsCall sends a tools/call request.
func (c *MockMCPClient) SendToolsCall(ctx context.Context, toolName string, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	request := ToolsCallRequest{
		Name:      toolName,
		Arguments: arguments,
	}

	response, err := c.sendRequest(ctx, "tools/call", request)
	if err != nil {
		return nil, fmt.Errorf("failed to send tools/call request: %w", err)
	}

	var callResponse mcp.CallToolResult
	if err := json.Unmarshal(response.Result, &callResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tools/call response: %w", err)
	}

	return &callResponse, nil
}

// sendRequest sends a JSON-RPC request and returns the response.
func (c *MockMCPClient) sendRequest(ctx context.Context, method string, params interface{}) (*JSONRPCResponse, error) {
	// Create JSON-RPC request
	requestID := time.Now().UnixNano()
	jsonrpcReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      requestID,
		Method:  method,
		Params:  params,
	}

	// Marshal request
	requestData, err := json.Marshal(jsonrpcReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Store request for validation
	c.requests = append(c.requests, requestData)

	// Send request
	if _, err := c.writer.Write(append(requestData, '\n')); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	// Read response with timeout
	responseCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	responseCh := make(chan []byte, 1)
	errCh := make(chan error, 1)

	go func() {
		buffer := make([]byte, 4096)
		n, err := c.reader.Read(buffer)
		if err != nil {
			errCh <- err
			return
		}
		responseCh <- buffer[:n]
	}()

	var responseData []byte
	select {
	case responseData = <-responseCh:
		// Response received
	case err := <-errCh:
		return nil, fmt.Errorf("failed to read response: %w", err)
	case <-responseCtx.Done():
		return nil, fmt.Errorf("request timeout: %w", responseCtx.Err())
	}

	// Store response for validation
	c.responses = append(c.responses, responseData)

	// Parse JSON-RPC response
	var jsonrpcResp JSONRPCResponse
	if err := json.Unmarshal(responseData, &jsonrpcResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Validate JSON-RPC response
	if err := c.validateJSONRPCResponse(&jsonrpcResp, requestID); err != nil {
		return nil, fmt.Errorf("invalid JSON-RPC response: %w", err)
	}

	return &jsonrpcResp, nil
}

// validateJSONRPCResponse validates a JSON-RPC 2.0 response.
func (c *MockMCPClient) validateJSONRPCResponse(response *JSONRPCResponse, expectedID interface{}) error {
	// Validate JSON-RPC version
	if response.JSONRPC != "2.0" {
		return fmt.Errorf("invalid JSON-RPC version: expected '2.0', got '%s'", response.JSONRPC)
	}

	// Validate ID matches request
	if response.ID != expectedID {
		return fmt.Errorf("response ID mismatch: expected %v, got %v", expectedID, response.ID)
	}

	// Validate either result or error is present (but not both)
	hasResult := len(response.Result) > 0
	hasError := response.Error != nil

	if hasResult && hasError {
		return fmt.Errorf("response cannot have both result and error")
	}

	if !hasResult && !hasError {
		return fmt.Errorf("response must have either result or error")
	}

	// Validate error structure if present
	if hasError {
		if response.Error.Code == 0 {
			return fmt.Errorf("error code cannot be zero")
		}
		if response.Error.Message == "" {
			return fmt.Errorf("error message cannot be empty")
		}
	}

	return nil
}

// GetStoredRequests returns all stored requests for validation.
func (c *MockMCPClient) GetStoredRequests() []json.RawMessage {
	return c.requests
}

// GetStoredResponses returns all stored responses for validation.
func (c *MockMCPClient) GetStoredResponses() []json.RawMessage {
	return c.responses
}

// CreateStdioPipe creates a pipe for stdio communication testing.
func CreateStdioPipe() (io.Reader, io.Writer, io.Reader, io.Writer) {
	// Create two pipes: client->server and server->client
	clientReader, serverWriter := io.Pipe()
	serverReader, clientWriter := io.Pipe()

	return clientReader, clientWriter, serverReader, serverWriter
}

// ValidateJSONRPCMessage validates that a message conforms to JSON-RPC 2.0.
func ValidateJSONRPCMessage(data []byte) error {
	var message map[string]interface{}
	if err := json.Unmarshal(data, &message); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Check JSON-RPC version
	jsonrpc, ok := message["jsonrpc"].(string)
	if !ok || jsonrpc != "2.0" {
		return fmt.Errorf("missing or invalid jsonrpc field: expected '2.0'")
	}

	// Check if it's a request or response
	if method, hasMethod := message["method"]; hasMethod {
		// It's a request
		if _, ok := method.(string); !ok {
			return fmt.Errorf("method must be a string")
		}
		// ID is optional for notifications
	} else {
		// It's a response
		if _, hasID := message["id"]; !hasID {
			return fmt.Errorf("response must have id field")
		}

		hasResult := message["result"] != nil
		hasError := message["error"] != nil

		if hasResult && hasError {
			return fmt.Errorf("response cannot have both result and error")
		}
		if !hasResult && !hasError {
			return fmt.Errorf("response must have either result or error")
		}
	}

	return nil
}

// CreateBufferedClient creates a mock client with buffered I/O for testing.
func CreateBufferedClient() (*MockMCPClient, *bytes.Buffer, *bytes.Buffer) {
	clientToServer := &bytes.Buffer{}
	serverToClient := &bytes.Buffer{}

	client := NewMockMCPClient(serverToClient, clientToServer)
	return client, clientToServer, serverToClient
}