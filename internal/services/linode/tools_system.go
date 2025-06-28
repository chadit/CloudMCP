package linode

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/chadit/CloudMCP/internal/version"
)

// handleSystemVersion returns version and build information
func (s *Service) handleSystemVersion(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	versionInfo := version.Get()

	// Format as detailed text output
	output := fmt.Sprintf(`CloudMCP Version Information:

Version: %s
API Version: %s (Linode API v4)
Build Date: %s
Git Commit: %s
Git Branch: %s
Go Version: %s
Platform: %s

Features:
  • Linode API Coverage: %s
  • Multi-Account Support: %s
  • Metrics: %s
  • Logging: %s
  • Protocol: %s

Current Account: %s`,
		versionInfo.Version,
		versionInfo.APIVersion,
		versionInfo.BuildDate,
		versionInfo.GitCommit,
		versionInfo.GitBranch,
		versionInfo.GoVersion,
		versionInfo.Platform,
		versionInfo.Features["linode_api_coverage"],
		versionInfo.Features["multi_account"],
		versionInfo.Features["metrics"],
		versionInfo.Features["logging"],
		versionInfo.Features["protocol"],
		s.getCurrentAccountName(),
	)

	return mcp.NewToolResultText(output), nil
}

// handleSystemVersionJSON returns version information as JSON
func (s *Service) handleSystemVersionJSON(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	versionInfo := version.Get()

	// Add current account information
	output := map[string]interface{}{
		"cloudmcp":        versionInfo,
		"current_account": s.getCurrentAccountName(),
		"service":         "linode",
	}

	jsonOutput, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal version info: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonOutput)), nil
}

// getCurrentAccountName returns the current account name for version info
func (s *Service) getCurrentAccountName() string {
	account, err := s.accountManager.GetCurrentAccount()
	if err != nil {
		return "unknown"
	}
	return fmt.Sprintf("%s (%s)", account.Name, account.Label)
}
