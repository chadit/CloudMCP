// Package tools provides simple built-in tools for CloudMCP minimal shell mode.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/chadit/CloudMCP/internal/version"
)

// HealthCheckTool provides server health and service discovery information.
// This tool serves as the minimal interface for a CloudMCP shell instance,
// providing essential status information and readiness for future provider registration.
type HealthCheckTool struct {
	serverName string
	startTime  time.Time
	toolsCount int
}

// HealthCheckResponse represents the structured response from the health check tool.
type HealthCheckResponse struct {
	Status     string        `json:"status"`
	Message    string        `json:"message"`
	ServerInfo ServerInfo    `json:"serverInfo"`
	Services   ServicesInfo  `json:"availableServices"`
	Providers  ProvidersInfo `json:"providers"`
	Timestamp  string        `json:"timestamp"`
	Uptime     string        `json:"uptime"`
	Metrics    MetricsInfo   `json:"metrics"`
}

// ServerInfo contains basic server information.
type ServerInfo struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Mode       string `json:"mode"`
	APIVersion string `json:"apiVersion"`
	Platform   string `json:"platform"`
	GitCommit  string `json:"gitCommit"`
}

// ServicesInfo contains information about available services.
type ServicesInfo struct {
	ToolsRegistered int      `json:"toolsRegistered"`
	ToolNames       []string `json:"toolNames"`
	Capabilities    []string `json:"capabilities"`
}

// ProvidersInfo contains information about cloud providers.
type ProvidersInfo struct {
	Registered int      `json:"registered"`
	Available  []string `json:"available"`
	Status     string   `json:"status"`
}

// MetricsInfo contains metrics system information.
type MetricsInfo struct {
	Enabled  bool   `json:"enabled"`
	Endpoint string `json:"endpoint,omitempty"`
	Backend  string `json:"backend"`
}

// NewHealthCheckTool creates a new health check tool instance.
func NewHealthCheckTool(serverName string) *HealthCheckTool {
	return &HealthCheckTool{
		serverName: serverName,
		startTime:  time.Now(),
		toolsCount: 1, // The health tool itself
	}
}

// Name returns the tool name.
func (h *HealthCheckTool) Name() string {
	return "health_check"
}

// Description returns the tool description.
func (h *HealthCheckTool) Description() string {
	return "Check server health and list available services for CloudMCP shell"
}

// InputSchema returns the tool's input schema.
func (h *HealthCheckTool) InputSchema() any {
	return mcp.ToolInputSchema{
		Type:       "object",
		Properties: map[string]any{},
	}
}

// Execute performs the health check and returns service discovery information.
func (h *HealthCheckTool) Execute(_ context.Context, _ map[string]any) (*mcp.CallToolResult, error) {
	// Get version information
	versionInfo := version.Get()

	// Calculate uptime
	uptime := time.Since(h.startTime)

	// Prepare response
	response := HealthCheckResponse{
		Status:  "healthy",
		Message: "CloudMCP server running - minimal shell configuration",
		ServerInfo: ServerInfo{
			Name:       h.serverName,
			Version:    versionInfo.Version + "-minimal",
			Mode:       "minimal_shell",
			APIVersion: versionInfo.APIVersion,
			Platform:   versionInfo.Platform,
			GitCommit:  versionInfo.GitCommit,
		},
		Services: ServicesInfo{
			ToolsRegistered: h.toolsCount,
			ToolNames:       []string{"health_check"},
			Capabilities:    []string{"health_monitoring", "service_discovery", "ready_for_provider_expansion"},
		},
		Providers: ProvidersInfo{
			Registered: 0,
			Available:  []string{"ready_for_future_providers"},
			Status:     "minimal_mode",
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Uptime:    formatDuration(uptime),
		Metrics: MetricsInfo{
			Enabled:  true,
			Endpoint: "/metrics",
			Backend:  "prometheus",
		},
	}

	// Return as JSON for better formatting in MCP responses
	jsonResponse, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("%+v", response)), fmt.Errorf("failed to marshal health response: %w", err)
	}

	return mcp.NewToolResultText(string(jsonResponse)), nil
}

// UpdateToolsCount updates the count of registered tools.
func (h *HealthCheckTool) UpdateToolsCount(count int) {
	h.toolsCount = count
}

const (
	secondsPerMinute = 60
	minutesPerHour   = 60
)

// formatDuration formats a duration in a human-readable format.
// Uses ISO 8601 duration format for consistency with standards.
func formatDuration(duration time.Duration) string {
	if duration < time.Minute {
		return fmt.Sprintf("PT%dS", int(duration.Seconds()))
	}

	if duration < time.Hour {
		minutes := int(duration.Minutes())
		seconds := int(duration.Seconds()) % secondsPerMinute

		if seconds == 0 {
			return fmt.Sprintf("PT%dM", minutes)
		}

		return fmt.Sprintf("PT%dM%dS", minutes, seconds)
	}

	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % minutesPerHour

	if minutes == 0 {
		return fmt.Sprintf("PT%dH", hours)
	}

	return fmt.Sprintf("PT%dH%dM", hours, minutes)
}
