package tools

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/chadit/CloudMCP/internal/version"
)

// NewVersionTool creates a new version tool with handler.
func NewVersionTool() (mcp.Tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)) {
	tool := mcp.NewTool("version",
		mcp.WithDescription("Returns CloudMCP server version and build information"),
	)

	handler := func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		versionInfo := version.Get()

		// Return as formatted JSON for readability
		jsonResponse, err := json.MarshalIndent(versionInfo, "", "  ")
		if err != nil {
			// Fallback to simple string format
			return mcp.NewToolResultText(versionInfo.String()), nil
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}

	return tool, handler
}
