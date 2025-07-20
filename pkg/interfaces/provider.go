// Package interfaces defines core interfaces for CloudMCP.
// Provides minimal foundation for tool management and future extensibility.
package interfaces

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

// Tool represents an MCP tool that can be registered with the server.
type Tool interface {
	// Name returns the tool name.
	Name() string

	// Description returns the tool description.
	Description() string

	// InputSchema returns the tool's input schema.
	InputSchema() any

	// Execute handles the actual tool execution with the provided parameters.
	Execute(ctx context.Context, params map[string]any) (*mcp.CallToolResult, error)
}
