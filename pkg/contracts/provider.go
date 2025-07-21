// Package contracts defines core interfaces for CloudMCP tool integration.
// This package provides the Tool interface contract that all MCP tools must implement,
// enabling a pluggable architecture for cloud provider extensions.
package contracts

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
