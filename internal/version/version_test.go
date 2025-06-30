package version_test

import (
	"encoding/json"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/version"
)

func TestGet_DefaultValues(t *testing.T) {
	t.Parallel()
	info := version.Get()

	// Test required fields are not empty
	require.NotEmpty(t, info.Version, "version.Version should not be empty")
	require.NotEmpty(t, info.APIVersion, "version.APIVersion should not be empty")
	require.NotEmpty(t, info.GitCommit, "version.GitCommit should not be empty")
	require.NotEmpty(t, info.GitBranch, "version.GitBranch should not be empty")
	require.NotEmpty(t, info.GoVersion, "GoVersion should not be empty")
	require.NotEmpty(t, info.Platform, "Platform should not be empty")
	require.NotEmpty(t, info.Features, "Features should not be empty")

	// Test specific default values
	require.Equal(t, version.Version, info.Version, "version.Version should match constant")
	require.Equal(t, version.APIVersion, info.APIVersion, "APIVersion should match constant")
	require.Equal(t, version.GitCommit, info.GitCommit, "version.GitCommit should match variable")
	require.Equal(t, version.GitBranch, info.GitBranch, "version.GitBranch should match variable")
	require.Equal(t, runtime.Version(), info.GoVersion, "GoVersion should match runtime version")

	// Test build date handling when empty
	if version.BuildDate == "" {
		require.Equal(t, "unknown", info.BuildDate, "version.BuildDate should be 'unknown' when empty")
	} else {
		require.Equal(t, version.BuildDate, info.BuildDate, "version.BuildDate should match constant when set")
	}
}

func TestGet_Platform(t *testing.T) {
	t.Parallel()
	info := version.Get()

	expectedPlatform := runtime.GOOS + "/" + runtime.GOARCH
	require.Equal(t, expectedPlatform, info.Platform, "Platform should be GOOS/GOARCH")
	require.Contains(t, info.Platform, "/", "Platform should contain separator")
}

func TestGet_Features(t *testing.T) {
	t.Parallel()
	info := version.Get()

	// Test that required features are present
	expectedFeatures := map[string]string{
		"linode_api_coverage": "100%",
		"multi_account":       "enabled",
		"metrics":             "prometheus",
		"logging":             "structured",
		"protocol":            "mcp",
	}

	for key, expectedValue := range expectedFeatures {
		actualValue, exists := info.Features[key]
		require.True(t, exists, "Feature %s should exist", key)
		require.Equal(t, expectedValue, actualValue, "Feature %s should have correct value", key)
	}
}

func TestInfo_String(t *testing.T) {
	t.Parallel()
	info := version.Get()
	str := info.String()

	// Test that string contains expected components
	require.Contains(t, str, "CloudMCP", "String should contain product name")
	require.Contains(t, str, info.Version, "String should contain version")
	require.Contains(t, str, info.APIVersion, "String should contain API version")
	require.Contains(t, str, info.Platform, "String should contain platform")
	require.Contains(t, str, info.GitCommit, "String should contain git commit")

	// Test string format
	require.Contains(t, str, "v"+info.Version, "version.Version should be prefixed with 'v'")
	require.Contains(t, str, "API: v"+info.APIVersion, "API version should be formatted correctly")
}

func TestInfo_BuildInfo(t *testing.T) {
	t.Parallel()
	info := version.Get()
	buildInfo := info.BuildInfo()

	// Test that build info contains all expected components
	require.Contains(t, buildInfo, "CloudMCP Build Information", "Should contain header")
	require.Contains(t, buildInfo, "Version: "+info.Version, "Should contain version")
	require.Contains(t, buildInfo, "API Version: "+info.APIVersion, "Should contain API version")
	require.Contains(t, buildInfo, "Build Date: "+info.BuildDate, "Should contain build date")
	require.Contains(t, buildInfo, "Git Commit: "+info.GitCommit, "Should contain git commit")
	require.Contains(t, buildInfo, "Git Branch: "+info.GitBranch, "Should contain git branch")
	require.Contains(t, buildInfo, "Go Version: "+info.GoVersion, "Should contain Go version")
	require.Contains(t, buildInfo, "Platform: "+info.Platform, "Should contain platform")
	require.Contains(t, buildInfo, "Linode API Coverage: 100%", "Should contain API coverage")
	require.Contains(t, buildInfo, "Features:", "Should contain features section")
	require.Contains(t, buildInfo, "Multi-account", "Should mention multi-account feature")
	require.Contains(t, buildInfo, "Metrics", "Should mention metrics feature")
	require.Contains(t, buildInfo, "Structured Logging", "Should mention logging feature")
}

func TestInfo_JSONSerialization(t *testing.T) {
	t.Parallel()
	info := version.Get()

	// Test that version.Info can be serialized to JSON
	jsonData, err := json.Marshal(info)
	require.NoError(t, err, "version.Info should be serializable to JSON")
	require.NotEmpty(t, jsonData, "JSON data should not be empty")

	// Test that JSON can be deserialized back to version.Info
	var deserialized version.Info
	err = json.Unmarshal(jsonData, &deserialized)
	require.NoError(t, err, "JSON should be deserializable to version.Info")

	// Test that deserialized info matches original
	require.Equal(t, info.Version, deserialized.Version, "version.Version should match after JSON round-trip")
	require.Equal(t, info.APIVersion, deserialized.APIVersion, "version.APIVersion should match after JSON round-trip")
	require.Equal(t, info.BuildDate, deserialized.BuildDate, "version.BuildDate should match after JSON round-trip")
	require.Equal(t, info.GitCommit, deserialized.GitCommit, "version.GitCommit should match after JSON round-trip")
	require.Equal(t, info.GitBranch, deserialized.GitBranch, "version.GitBranch should match after JSON round-trip")
	require.Equal(t, info.GoVersion, deserialized.GoVersion, "GoVersion should match after JSON round-trip")
	require.Equal(t, info.Platform, deserialized.Platform, "Platform should match after JSON round-trip")
	require.Equal(t, info.Features, deserialized.Features, "Features should match after JSON round-trip")
}

func TestInfo_JSONStructure(t *testing.T) {
	t.Parallel()
	info := version.Get()

	jsonData, err := json.Marshal(info)
	require.NoError(t, err, "Should marshal to JSON")

	// Parse JSON to verify structure
	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	require.NoError(t, err, "Should unmarshal JSON")

	// Test expected JSON fields
	expectedFields := []string{
		"version", "api_version", "build_date", "git_commit",
		"git_branch", "go_version", "platform", "features",
	}

	for _, field := range expectedFields {
		require.Contains(t, jsonMap, field, "JSON should contain field: %s", field)
	}

	// Test features is an object
	features, ok := jsonMap["features"].(map[string]interface{})
	require.True(t, ok, "Features should be an object in JSON")
	require.NotEmpty(t, features, "Features object should not be empty")
}

func TestConstants(t *testing.T) {
	t.Parallel()
	// Test that constants have expected values
	require.Equal(t, "1.0.0", version.Version, "version.Version constant should match expected value")
	require.Equal(t, "4.0", version.APIVersion, "version.APIVersion constant should match expected value")
	require.Equal(t, "", version.BuildDate, "version.BuildDate should be empty by default")

	// Test that variables have expected default values
	require.Equal(t, "dev", version.GitCommit, "version.GitCommit should have default value")
	require.Equal(t, "main", version.GitBranch, "version.GitBranch should have default value")
}

func TestInfo_StringFormat(t *testing.T) {
	t.Parallel()
	info := version.Get()
	str := info.String()

	// Test string format more precisely
	expectedFormat := "CloudMCP v" + info.Version + " (API: v" + info.APIVersion + ", " + info.Platform + ", " + info.GitCommit + ")"
	require.Equal(t, expectedFormat, str, "String format should match expected pattern")
}

func TestInfo_BuildInfoFormat(t *testing.T) {
	t.Parallel()
	info := version.Get()
	buildInfo := info.BuildInfo()

	// Test that build info has proper structure
	lines := strings.Split(buildInfo, "\n")
	require.GreaterOrEqual(t, len(lines), 8, "Build info should have multiple lines")

	// Test specific line content
	require.Contains(t, lines[0], "CloudMCP Build Information:", "First line should be header")

	// Find and test specific information lines
	foundVersion := false
	foundAPIVersion := false
	foundPlatform := false

	for _, line := range lines {
		if strings.Contains(line, "Version: "+info.Version) {
			foundVersion = true
		}

		if strings.Contains(line, "API Version: "+info.APIVersion) {
			foundAPIVersion = true
		}

		if strings.Contains(line, "Platform: "+info.Platform) {
			foundPlatform = true
		}
	}

	require.True(t, foundVersion, "Build info should contain version line")
	require.True(t, foundAPIVersion, "Build info should contain API version line")
	require.True(t, foundPlatform, "Build info should contain platform line")
}

func TestGet_Consistency(t *testing.T) {
	t.Parallel()
	// Test that multiple calls to version.Get() return consistent data
	info1 := version.Get()
	info2 := version.Get()

	require.Equal(t, info1.Version, info2.Version, "version.Version should be consistent")
	require.Equal(t, info1.APIVersion, info2.APIVersion, "version.APIVersion should be consistent")
	require.Equal(t, info1.BuildDate, info2.BuildDate, "version.BuildDate should be consistent")
	require.Equal(t, info1.GitCommit, info2.GitCommit, "version.GitCommit should be consistent")
	require.Equal(t, info1.GitBranch, info2.GitBranch, "version.GitBranch should be consistent")
	require.Equal(t, info1.GoVersion, info2.GoVersion, "Goversion.Version should be consistent")
	require.Equal(t, info1.Platform, info2.Platform, "Platform should be consistent")
	require.Equal(t, info1.Features, info2.Features, "Features should be consistent")
}
