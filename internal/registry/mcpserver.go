package registry

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"

	"github.com/chadit/CloudMCP/pkg/interfaces"
)

// Package-level errors.
var (
	ErrToolCannotBeNil = errors.New("tool cannot be nil")
	ErrToolNotFound    = errors.New("tool not found")
)

// MCPServerAdapter adapts the mark3labs MCP server to the interfaces.MCPServer interface.
// This allows providers to register tools without being tightly coupled to a specific
// MCP server implementation.
type MCPServerAdapter struct {
	server *mcpserver.MCPServer
	tools  []interfaces.Tool
	mu     sync.RWMutex
}

// NewMCPServerAdapter creates a new MCP server adapter.
func NewMCPServerAdapter(mcpServer *mcpserver.MCPServer) *MCPServerAdapter {
	return &MCPServerAdapter{
		server: mcpServer,
		tools:  make([]interfaces.Tool, 0),
	}
}

// RegisterTool registers a single MCP tool with the server.
func (m *MCPServerAdapter) RegisterTool(tool interfaces.Tool) error {
	if tool == nil {
		return ErrToolCannotBeNil
	}

	// Get the tool definition
	definition := tool.Definition()

	// Create a handler function that wraps the tool execution
	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Get arguments from the request
		args := request.GetArguments()

		// Validate parameters first
		if err := tool.Validate(args); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("parameter validation failed: %v", err)), nil
		}

		// Execute the tool
		result, err := tool.Execute(ctx, args)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("tool execution failed: %v", err)), nil
		}

		// Convert result to MCP tool result
		return mcp.NewToolResultText(fmt.Sprintf("%v", result)), nil
	}

	// Register with the underlying MCP server using AddTool method
	m.server.AddTool(definition, handler)

	// Track the tool in our internal list
	m.mu.Lock()
	m.tools = append(m.tools, tool)
	m.mu.Unlock()

	return nil
}

// RegisterTools registers multiple MCP tools at once.
func (m *MCPServerAdapter) RegisterTools(tools []interfaces.Tool) error {
	for _, tool := range tools {
		if err := m.RegisterTool(tool); err != nil {
			return err
		}
	}

	return nil
}

// GetRegisteredTools returns the list of all registered tools.
func (m *MCPServerAdapter) GetRegisteredTools() []interfaces.Tool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]interfaces.Tool, len(m.tools))
	copy(result, m.tools)

	return result
}

// GetToolCount returns the number of registered tools.
func (m *MCPServerAdapter) GetToolCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.tools)
}

// HasTool checks if a tool with the given name is registered.
func (m *MCPServerAdapter) HasTool(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, tool := range m.tools {
		if tool.Definition().Name == name {
			return true
		}
	}

	return false
}

// GetTool retrieves a registered tool by name.
func (m *MCPServerAdapter) GetTool(name string) (interfaces.Tool, error) { // Adapter method should return interface
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, tool := range m.tools {
		if tool.Definition().Name == name {
			return tool, nil
		}
	}

	return nil, fmt.Errorf("%w: %q", ErrToolNotFound, name)
}

// ListToolNames returns the names of all registered tools.
func (m *MCPServerAdapter) ListToolNames() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, len(m.tools))
	for i, tool := range m.tools {
		names[i] = tool.Definition().Name
	}

	return names
}

// GetToolDefinitions returns the MCP tool definitions for all registered tools.
func (m *MCPServerAdapter) GetToolDefinitions() []mcp.Tool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	definitions := make([]mcp.Tool, len(m.tools))
	for i, tool := range m.tools {
		definitions[i] = tool.Definition()
	}

	return definitions
}

// Clear removes all registered tools.
// This is primarily useful for testing purposes.
func (m *MCPServerAdapter) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tools = make([]interfaces.Tool, 0)
}

// GetUnderlyingServer returns the underlying MCP server instance.
// This allows access to advanced features not exposed through the adapter interface.
func (m *MCPServerAdapter) GetUnderlyingServer() *mcpserver.MCPServer {
	return m.server
}

// ToolWrapper wraps a function to implement the interfaces.Tool interface.
// This is useful for creating simple tools without implementing the full interface.
type ToolWrapper struct {
	definition mcp.Tool
	handler    func(map[string]interface{}) (interface{}, error)
	validator  func(map[string]interface{}) error
}

// NewToolWrapper creates a new tool wrapper with the provided definition and handler.
func NewToolWrapper(
	definition mcp.Tool,
	handler func(map[string]interface{}) (interface{}, error),
	validator func(map[string]interface{}) error,
) *ToolWrapper {
	if validator == nil {
		// Default validator that always passes
		validator = func(map[string]interface{}) error {
			return nil
		}
	}

	return &ToolWrapper{
		definition: definition,
		handler:    handler,
		validator:  validator,
	}
}

// Definition returns the MCP tool definition.
func (tw *ToolWrapper) Definition() mcp.Tool {
	return tw.definition
}

// Execute handles the tool execution.
func (tw *ToolWrapper) Execute(_ context.Context, params map[string]interface{}) (interface{}, error) {
	return tw.handler(params)
}

// Validate validates the tool parameters.
func (tw *ToolWrapper) Validate(params map[string]interface{}) error {
	return tw.validator(params)
}

// SimpleToolBuilder helps build simple tools with a fluent interface.
type SimpleToolBuilder struct {
	tool *ToolWrapper
}

// NewSimpleToolBuilder creates a new simple tool builder.
func NewSimpleToolBuilder(name, description string) *SimpleToolBuilder {
	definition := mcp.Tool{
		Name:        name,
		Description: description,
	}

	return &SimpleToolBuilder{
		tool: &ToolWrapper{
			definition: definition,
			validator: func(map[string]interface{}) error {
				return nil
			},
		},
	}
}

// WithInputSchema sets the input schema for the tool.
func (stb *SimpleToolBuilder) WithInputSchema(schema mcp.ToolInputSchema) *SimpleToolBuilder {
	stb.tool.definition.InputSchema = schema

	return stb
}

// WithHandler sets the execution handler for the tool.
func (stb *SimpleToolBuilder) WithHandler(handler func(map[string]interface{}) (interface{}, error)) *SimpleToolBuilder {
	stb.tool.handler = handler

	return stb
}

// WithValidator sets the parameter validator for the tool.
func (stb *SimpleToolBuilder) WithValidator(validator func(map[string]interface{}) error) *SimpleToolBuilder {
	stb.tool.validator = validator

	return stb
}

// Build returns the constructed tool.
func (stb *SimpleToolBuilder) Build() interfaces.Tool { // Builder method should return interface
	return stb.tool
}
