package version

import (
	"fmt"
	"runtime"
)

const (
	// Version is the semantic version of CloudMCP.
	Version = "1.0.0"

	// APIVersion is the version of the Linode API coverage.
	APIVersion = "4.0"

	// BuildDate can be set at build time using ldflags.
	BuildDate = ""
)

var (
	// GitCommit can be set at build time using ldflags. //nolint:gochecknoglobals // Build-time constants set via ldflags
	GitCommit = "dev"

	// GitBranch can be set at build time using ldflags. //nolint:gochecknoglobals // Build-time constants set via ldflags
	GitBranch = "main"
)

// Info contains version and build information.
type Info struct {
	Version    string            `json:"version"`
	APIVersion string            `json:"api_version"` //nolint:tagliatelle // Maintaining API compatibility
	BuildDate  string            `json:"build_date"`  //nolint:tagliatelle // Maintaining API compatibility  
	GitCommit  string            `json:"git_commit"` //nolint:tagliatelle // Maintaining API compatibility
	GitBranch  string            `json:"git_branch"` //nolint:tagliatelle // Maintaining API compatibility
	GoVersion  string            `json:"go_version"` //nolint:tagliatelle // Maintaining API compatibility
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
			"linode_api_coverage": "100%",
			"multi_account":       "enabled",
			"metrics":             "prometheus",
			"logging":             "structured",
			"protocol":            "mcp",
		},
	}
}

// String returns a human-readable version string.
func (i Info) String() string {
	return fmt.Sprintf("CloudMCP v%s (API: v%s, %s, %s)",
		i.Version, i.APIVersion, i.Platform, i.GitCommit)
}

// BuildInfo returns detailed build information.
func (i Info) BuildInfo() string {
	return fmt.Sprintf(`CloudMCP Build Information:
  Version: %s
  API Version: %s  
  Build Date: %s
  Git Commit: %s
  Git Branch: %s
  Go Version: %s
  Platform: %s
  Linode API Coverage: %s
  Features: Multi-account, Metrics, Structured Logging`,
		i.Version, i.APIVersion, i.BuildDate, i.GitCommit,
		i.GitBranch, i.GoVersion, i.Platform, i.Features["linode_api_coverage"])
}
