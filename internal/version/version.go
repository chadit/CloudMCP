// Package version provides build-time version information for CloudMCP.
package version

import (
	"fmt"
	"runtime"
)

const (
	// Version is the semantic version of CloudMCP.
	Version = "1.0.0"

	// APIVersion is the version of the MCP protocol.
	APIVersion = "1.0"

	// BuildDate can be set at build time using ldflags.
	BuildDate = ""
)

var (
	// GitCommit can be set at build time using ldflags.
	GitCommit = "dev" //nolint:gochecknoglobals // Build-time variable set via ldflags

	// GitBranch can be set at build time using ldflags.
	GitBranch = "main" //nolint:gochecknoglobals // Build-time variable set via ldflags
)

// Info contains version and build information.
//
//nolint:tagliatelle // JSON field names maintain API compatibility with snake_case.
type Info struct {
	Version    string            `json:"version"`
	APIVersion string            `json:"api_version"` // Maintaining API compatibility
	BuildDate  string            `json:"build_date"`  // Maintaining API compatibility
	GitCommit  string            `json:"git_commit"`  // Maintaining API compatibility
	GitBranch  string            `json:"git_branch"`  // Maintaining API compatibility
	GoVersion  string            `json:"go_version"`  // Maintaining API compatibility
	Platform   string            `json:"platform"`
	Features   map[string]string `json:"features"`
}

// Get returns the current version information.
func Get() Info {
	buildDate := BuildDate

	if buildDate == "" {
		buildDate = "unknown"
	}

	return Info{
		Version:    Version,
		APIVersion: APIVersion,
		BuildDate:  buildDate,
		GitCommit:  GitCommit,
		GitBranch:  GitBranch,
		GoVersion:  runtime.Version(),
		Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		Features: map[string]string{
			"health_check": "enabled",
			"metrics":      "prometheus",
			"logging":      "structured",
			"protocol":     "mcp",
		},
	}
}

// String returns a human-readable version string.
func (i Info) String() string {
	return fmt.Sprintf("CloudMCP v%s (MCP: v%s, %s, %s)",
		i.Version, i.APIVersion, i.Platform, i.GitCommit)
}

// BuildInfo returns detailed build information.
func (i Info) BuildInfo() string {
	return fmt.Sprintf(`CloudMCP Build Information:
  Version: %s
  MCP Protocol: %s
  Build Date: %s
  Git Commit: %s
  Git Branch: %s
  Go Version: %s
  Platform: %s
  Health Check: %s
  Features: Health Check, Metrics, Structured Logging`,
		i.Version, i.APIVersion, i.BuildDate, i.GitCommit,
		i.GitBranch, i.GoVersion, i.Platform, i.Features["health_check"])
}
