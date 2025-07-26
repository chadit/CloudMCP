// Package tools provides simple built-in tools for CloudMCP minimal shell mode.
package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// NewHelloTool creates a new hello tool with handler.
func NewHelloTool() (mcp.Tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)) {
	tool := mcp.NewTool("hello",
		mcp.WithDescription("Responds with a friendly greeting message from CloudMCP"),
		mcp.WithString("name",
			mcp.Description("Name to include in the greeting (optional)"),
		),
	)

	handler := func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := request.GetString("name", "World")
		message := fmt.Sprintf("Hello, %s! CloudMCP server is running and ready to help.", name)

		return mcp.NewToolResultText(message), nil
	}

	return tool, handler
}
