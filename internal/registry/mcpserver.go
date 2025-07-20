// Package registry provides MCP (Model Context Protocol) server adapters and tool management.
// It acts as an abstraction layer between CloudMCP's internal tool system and the underlying
// MCP server implementation, enabling loose coupling and easier testing.
//
// # Core Components
//
// The package centers around the MCPServerAdapter, which wraps the mark3labs/mcp-go server
// implementation and provides a clean interface for tool registration and management.
//
// # Tool Registration
//
// Tools implementing the interfaces.Tool interface can be registered with the MCP server:
//
//	adapter := registry.NewMCPServerAdapter(mcpServer)
//	healthTool := tools.NewHealthCheckTool("server-name")
//	err := adapter.RegisterTool(healthTool)
//
// # Thread Safety
//
// All adapter operations are thread-safe using read-write mutexes, allowing concurrent
// access to tool information while protecting against race conditions during registration.
//
// # Tool Management Features
//
//   - Tool registration with automatic MCP protocol integration
//   - Tool lookup by name with O(n) search
//   - Tool counting and listing capabilities
//   - Thread-safe concurrent access to tool registry
//   - Batch tool registration for efficient setup
//   - Clear functionality for testing scenarios
//
// # MCP Protocol Integration
//
// The adapter automatically handles:
//   - Converting tool definitions to MCP format
//   - Creating request handlers that bridge tool execution
//   - Managing tool input schema validation
//   - Wrapping tool responses in proper MCP result format
//
// # Usage in CloudMCP
//
// This package is primarily used by the server package to register and manage
// tools like health checks, with the adapter serving as the bridge between
// CloudMCP's tool interface and the MCP protocol requirements.
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
// This allows providers to register tools without being tightly coupled to a specific.
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

	// Create the tool definition.
	schema, _ := tool.InputSchema().(mcp.ToolInputSchema)
	definition := mcp.Tool{
		Name:        tool.Name(),
		Description: tool.Description(),
		InputSchema: schema,
	}

	// Create a handler function that wraps the tool execution.
	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Get arguments from the request.
		args := request.GetArguments()

		// Execute the tool directly (no separate validation step).
		return tool.Execute(ctx, args)
	}

	// Register with the underlying MCP server using AddTool method.
	m.server.AddTool(definition, handler)

	// Track the tool in our internal list.
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

	// Return a copy to prevent external modification.
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
		if tool.Name() == name {
			return true
		}
	}

	return false
}

// GetTool retrieves a registered tool by name.
//
//nolint:ireturn // GetTool returns interface to allow multiple metric implementations.
func (m *MCPServerAdapter) GetTool(name string) (interfaces.Tool, error) { // Adapter method should return interface
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, tool := range m.tools {
		if tool.Name() == name {
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
		names[i] = tool.Name()
	}

	return names
}

// GetToolDefinitions returns the MCP tool definitions for all registered tools.
func (m *MCPServerAdapter) GetToolDefinitions() []mcp.Tool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	definitions := make([]mcp.Tool, len(m.tools))

	for i, tool := range m.tools {
		schema, _ := tool.InputSchema().(mcp.ToolInputSchema)
		definitions[i] = mcp.Tool{
			Name:        tool.Name(),
			Description: tool.Description(),
			InputSchema: schema,
		}
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
