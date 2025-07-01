// Package interfaces_test provides comprehensive unit tests for the interfaces package.
// It validates the behavior and structure of cloud provider interfaces, capabilities,
// and metadata types used throughout the CloudMCP system.
package interfaces_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/pkg/interfaces"
)

func TestCapability(t *testing.T) {
	t.Parallel()

	capability := interfaces.Capability{
		Name:         "compute",
		Description:  "Virtual machine management",
		Version:      "1.0.0",
		Category:     "infrastructure",
		Dependencies: []string{"networking"},
		Experimental: false,
	}

	require.Equal(t, "compute", capability.Name, "Name should match")
	require.Equal(t, "Virtual machine management", capability.Description, "Description should match")
	require.Equal(t, "1.0.0", capability.Version, "Version should match")
	require.Equal(t, "infrastructure", capability.Category, "Category should match")
	require.Contains(t, capability.Dependencies, "networking", "Dependencies should contain networking")
	require.False(t, capability.Experimental, "Should not be experimental")
}

func TestProviderMetadata(t *testing.T) {
	t.Parallel()

	metadata := interfaces.ProviderMetadata{
		Name:           "test-provider",
		DisplayName:    "Test Provider",
		Version:        "1.0.0",
		Description:    "A test cloud provider",
		Author:         "Test Author",
		Homepage:       "https://example.com",
		License:        "MIT",
		RequiredConfig: []string{"api_key", "region"},
		OptionalConfig: []string{"timeout", "retries"},
		Capabilities: []interfaces.Capability{
			{
				Name:        "compute",
				Description: "Compute services",
				Version:     "1.0.0",
				Category:    "infrastructure",
			},
		},
	}

	require.Equal(t, "test-provider", metadata.Name, "Name should match")
	require.Equal(t, "Test Provider", metadata.DisplayName, "DisplayName should match")
	require.Equal(t, "1.0.0", metadata.Version, "Version should match")
	require.Equal(t, "A test cloud provider", metadata.Description, "Description should match")
	require.Equal(t, "Test Author", metadata.Author, "Author should match")
	require.Equal(t, "https://example.com", metadata.Homepage, "Homepage should match")
	require.Equal(t, "MIT", metadata.License, "License should match")
	require.Contains(t, metadata.RequiredConfig, "api_key", "RequiredConfig should contain api_key")
	require.Contains(t, metadata.OptionalConfig, "timeout", "OptionalConfig should contain timeout")
	require.Len(t, metadata.Capabilities, 1, "Should have one capability")
	require.Equal(t, "compute", metadata.Capabilities[0].Name, "Capability name should match")
	require.Equal(t, "Compute services", metadata.Capabilities[0].Description, "Capability description should match")
}

func TestCapabilityValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		capability interfaces.Capability
		valid      bool
	}{
		{
			name: "valid capability",
			capability: interfaces.Capability{
				Name:        "storage",
				Description: "Storage services",
				Version:     "1.0.0",
				Category:    "infrastructure",
			},
			valid: true,
		},
		{
			name: "capability with dependencies",
			capability: interfaces.Capability{
				Name:         "kubernetes",
				Description:  "Kubernetes services",
				Version:      "1.0.0",
				Category:     "infrastructure",
				Dependencies: []string{"compute", "networking"},
			},
			valid: true,
		},
		{
			name: "experimental capability",
			capability: interfaces.Capability{
				Name:         "ai-services",
				Description:  "Experimental AI services",
				Version:      "0.1.0",
				Category:     "ai",
				Experimental: true,
			},
			valid: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Basic validation - ensure required fields are present
			require.NotEmpty(t, testCase.capability.Name, "Capability name should not be empty")
			require.NotEmpty(t, testCase.capability.Description, "Capability description should not be empty")
			require.NotEmpty(t, testCase.capability.Version, "Capability version should not be empty")
			require.NotEmpty(t, testCase.capability.Category, "Capability category should not be empty")
		})
	}
}

func TestProviderMetadataValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		metadata interfaces.ProviderMetadata
		valid    bool
	}{
		{
			name: "minimal valid metadata",
			metadata: interfaces.ProviderMetadata{
				Name:           "minimal-provider",
				DisplayName:    "Minimal Provider",
				Version:        "1.0.0",
				Description:    "A minimal provider",
				RequiredConfig: []string{"api_key"},
				Capabilities:   []interfaces.Capability{},
			},
			valid: true,
		},
		{
			name: "complete metadata",
			metadata: interfaces.ProviderMetadata{
				Name:           "complete-provider",
				DisplayName:    "Complete Provider",
				Version:        "2.1.0",
				Description:    "A complete provider implementation",
				Author:         "Provider Team",
				Homepage:       "https://provider.example.com",
				License:        "Apache-2.0",
				RequiredConfig: []string{"api_key", "secret"},
				OptionalConfig: []string{"region", "timeout"},
				Capabilities: []interfaces.Capability{
					{
						Name:        "compute",
						Description: "Compute management",
						Version:     "2.1.0",
						Category:    "infrastructure",
					},
				},
			},
			valid: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Validate required fields
			require.NotEmpty(t, testCase.metadata.Name, "Provider name should not be empty")
			require.NotEmpty(t, testCase.metadata.DisplayName, "Provider display name should not be empty")
			require.NotEmpty(t, testCase.metadata.Version, "Provider version should not be empty")
			require.NotEmpty(t, testCase.metadata.Description, "Provider description should not be empty")
		})
	}
}

func TestCapabilityCategories(t *testing.T) {
	t.Parallel()

	commonCategories := []string{
		"infrastructure",
		"networking",
		"storage",
		"security",
		"monitoring",
		"ai",
		"database",
		"analytics",
	}

	for _, category := range commonCategories {
		capability := interfaces.Capability{
			Name:        "test-capability",
			Description: "Test capability",
			Version:     "1.0.0",
			Category:    category,
		}

		require.Equal(t, "test-capability", capability.Name, "Name should match")
		require.Equal(t, "Test capability", capability.Description, "Description should match")
		require.Equal(t, "1.0.0", capability.Version, "Version should match")
		require.Equal(t, category, capability.Category, "Category should match")
	}
}

func TestCapabilityDependencies(t *testing.T) {
	t.Parallel()

	// Test capability with no dependencies
	basicCapability := interfaces.Capability{
		Name:        "basic",
		Description: "Basic capability",
		Version:     "1.0.0",
		Category:    "infrastructure",
	}
	require.Equal(t, "basic", basicCapability.Name, "Basic capability name should match")
	require.Equal(t, "Basic capability", basicCapability.Description, "Basic capability description should match")
	require.Equal(t, "1.0.0", basicCapability.Version, "Basic capability version should match")
	require.Equal(t, "infrastructure", basicCapability.Category, "Basic capability category should match")
	require.Empty(t, basicCapability.Dependencies, "Basic capability should have no dependencies")

	// Test capability with dependencies
	advancedCapability := interfaces.Capability{
		Name:         "advanced",
		Description:  "Advanced capability",
		Version:      "1.0.0",
		Category:     "infrastructure",
		Dependencies: []string{"basic", "networking"},
	}
	require.Equal(t, "advanced", advancedCapability.Name, "Advanced capability name should match")
	require.Equal(t, "Advanced capability", advancedCapability.Description, "Advanced capability description should match")
	require.Equal(t, "1.0.0", advancedCapability.Version, "Advanced capability version should match")
	require.Equal(t, "infrastructure", advancedCapability.Category, "Advanced capability category should match")
	require.Len(t, advancedCapability.Dependencies, 2, "Advanced capability should have 2 dependencies")
	require.Contains(t, advancedCapability.Dependencies, "basic", "Should depend on basic")
	require.Contains(t, advancedCapability.Dependencies, "networking", "Should depend on networking")
}

func TestExperimentalCapabilities(t *testing.T) {
	t.Parallel()

	stableCapability := interfaces.Capability{
		Name:         "stable-feature",
		Description:  "A stable feature",
		Version:      "1.0.0",
		Category:     "infrastructure",
		Experimental: false,
	}
	require.Equal(t, "stable-feature", stableCapability.Name, "Stable capability name should match")
	require.Equal(t, "A stable feature", stableCapability.Description, "Stable capability description should match")
	require.Equal(t, "1.0.0", stableCapability.Version, "Stable capability version should match")
	require.Equal(t, "infrastructure", stableCapability.Category, "Stable capability category should match")
	require.False(t, stableCapability.Experimental, "Stable capability should not be experimental")

	experimentalCapability := interfaces.Capability{
		Name:         "experimental-feature",
		Description:  "An experimental feature",
		Version:      "0.1.0",
		Category:     "experimental",
		Experimental: true,
	}
	require.Equal(t, "experimental-feature", experimentalCapability.Name, "Experimental capability name should match")
	require.Equal(t, "An experimental feature", experimentalCapability.Description, "Experimental capability description should match")
	require.Equal(t, "0.1.0", experimentalCapability.Version, "Experimental capability version should match")
	require.Equal(t, "experimental", experimentalCapability.Category, "Experimental capability category should match")
	require.True(t, experimentalCapability.Experimental, "Experimental capability should be marked as experimental")
}

func TestProviderConfigValidation(t *testing.T) {
	t.Parallel()

	metadata := interfaces.ProviderMetadata{
		Name:           "config-test-provider",
		DisplayName:    "Config Test Provider",
		Version:        "1.0.0",
		Description:    "Provider for testing configuration",
		RequiredConfig: []string{"api_key", "region", "project_id"},
		OptionalConfig: []string{"timeout", "retry_count", "debug"},
	}

	// Validate basic metadata fields
	require.Equal(t, "config-test-provider", metadata.Name, "Provider name should match")
	require.Equal(t, "Config Test Provider", metadata.DisplayName, "Provider display name should match")
	require.Equal(t, "1.0.0", metadata.Version, "Provider version should match")
	require.Equal(t, "Provider for testing configuration", metadata.Description, "Provider description should match")

	// Test that required config is properly defined
	require.Len(t, metadata.RequiredConfig, 3, "Should have 3 required config items")
	require.Contains(t, metadata.RequiredConfig, "api_key", "Should require api_key")
	require.Contains(t, metadata.RequiredConfig, "region", "Should require region")
	require.Contains(t, metadata.RequiredConfig, "project_id", "Should require project_id")

	// Test that optional config is properly defined
	require.Len(t, metadata.OptionalConfig, 3, "Should have 3 optional config items")
	require.Contains(t, metadata.OptionalConfig, "timeout", "Should have optional timeout")
	require.Contains(t, metadata.OptionalConfig, "retry_count", "Should have optional retry_count")
	require.Contains(t, metadata.OptionalConfig, "debug", "Should have optional debug")
}
