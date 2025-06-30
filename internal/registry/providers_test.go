package registry_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/registry"
	"github.com/chadit/CloudMCP/pkg/interfaces"
)

// MockProvider implements the CloudProvider interface for testing.
type MockProvider struct {
	name         string
	initialized  bool
	capabilities []interfaces.Capability
}

func (mp *MockProvider) Name() string {
	return mp.name
}

func (mp *MockProvider) Initialize(_ context.Context, _ interfaces.Config) error {
	mp.initialized = true

	return nil
}

func (mp *MockProvider) RegisterTools(_ interfaces.MCPServer) error {
	return nil
}

func (mp *MockProvider) GetCapabilities() []interfaces.Capability {
	return mp.capabilities
}

func (mp *MockProvider) ValidateConfig(_ interfaces.Config) error {
	return nil
}

func (mp *MockProvider) Shutdown(_ context.Context) error {
	mp.initialized = false

	return nil
}

func (mp *MockProvider) HealthCheck(_ context.Context) error {
	if !mp.initialized {
		return interfaces.ErrProviderNotInitialized
	}

	return nil
}

// MockProviderFactory implements the ProviderFactory interface for testing.
type MockProviderFactory struct {
	metadata interfaces.ProviderMetadata
}

func (mpf *MockProviderFactory) CreateProvider() interfaces.CloudProvider {
	return &MockProvider{
		name: mpf.metadata.Name,
		capabilities: []interfaces.Capability{
			{
				Name:        "test-capability",
				Description: "Test capability",
				Version:     "1.0.0",
				Category:    "test",
			},
		},
	}
}

func (mpf *MockProviderFactory) GetMetadata() interfaces.ProviderMetadata {
	return mpf.metadata
}

func (mpf *MockProviderFactory) ValidateConfig(_ interfaces.Config) error {
	return nil
}

func TestNewRegistry(t *testing.T) {
	t.Parallel()

	registry := registry.NewRegistry()
	require.NotNil(t, registry, "Registry should not be nil")
	require.Equal(t, 0, registry.Count(), "New registry should be empty")
}

func TestRegistry_RegisterProvider(t *testing.T) {
	t.Parallel()

	registry := registry.NewRegistry()

	metadata := interfaces.ProviderMetadata{
		Name:        "test-provider",
		DisplayName: "Test Provider",
		Version:     "1.0.0",
		Description: "A test provider",
	}

	factory := &MockProviderFactory{metadata: metadata}

	// Test successful registration
	err := registry.RegisterProvider("test-provider", factory)
	require.NoError(t, err, "Registration should succeed")
	require.Equal(t, 1, registry.Count(), "Registry should have one provider")

	// Test duplicate registration
	err = registry.RegisterProvider("test-provider", factory)
	require.Error(t, err, "Duplicate registration should fail")
	require.Contains(t, err.Error(), "already registered", "Error should mention already registered")
}

func TestRegistry_RegisterProvider_InvalidInputs(t *testing.T) {
	t.Parallel()

	registry := registry.NewRegistry()
	factory := &MockProviderFactory{}

	// Test empty name
	err := registry.RegisterProvider("", factory)
	require.Error(t, err, "Empty name should fail")
	require.Contains(t, err.Error(), "cannot be empty", "Error should mention empty name")

	// Test nil factory
	err = registry.RegisterProvider("test-provider", nil)
	require.Error(t, err, "Nil factory should fail")
	require.Contains(t, err.Error(), "cannot be nil", "Error should mention nil factory")
}

func TestRegistry_GetProvider(t *testing.T) {
	t.Parallel()

	registry := registry.NewRegistry()

	metadata := interfaces.ProviderMetadata{
		Name:        "test-provider",
		DisplayName: "Test Provider",
		Version:     "1.0.0",
		Description: "A test provider",
	}

	factory := &MockProviderFactory{metadata: metadata}

	// Register provider
	err := registry.RegisterProvider("test-provider", factory)
	require.NoError(t, err, "Registration should succeed")

	// Test successful retrieval
	provider, err := registry.GetProvider("test-provider")
	require.NoError(t, err, "Should retrieve provider successfully")
	require.NotNil(t, provider, "Provider should not be nil")
	require.Equal(t, "test-provider", provider.Name(), "Provider name should match")

	// Test retrieval of non-existent provider
	_, err = registry.GetProvider("non-existent")
	require.Error(t, err, "Should fail to retrieve non-existent provider")
	require.Contains(t, err.Error(), "not registered", "Error should mention not registered")
}

func TestRegistry_ListProviders(t *testing.T) {
	t.Parallel()

	registry := registry.NewRegistry()

	// Test empty registry
	providers := registry.ListProviders()
	require.Empty(t, providers, "Empty registry should return empty list")

	// Register multiple providers
	for _, name := range []string{"provider-a", "provider-b", "provider-c"} {
		metadata := interfaces.ProviderMetadata{
			Name:        name,
			DisplayName: name + " Provider",
			Version:     "1.0.0",
			Description: "A test provider",
		}
		factory := &MockProviderFactory{metadata: metadata}

		err := registry.RegisterProvider(name, factory)
		require.NoError(t, err, "Registration should succeed for %s", name)
	}

	// Test listing
	providers = registry.ListProviders()
	require.Len(t, providers, 3, "Should have 3 providers")
	require.Contains(t, providers, "provider-a", "Should contain provider-a")
	require.Contains(t, providers, "provider-b", "Should contain provider-b")
	require.Contains(t, providers, "provider-c", "Should contain provider-c")
}

func TestRegistry_GetProviderMetadata(t *testing.T) {
	t.Parallel()

	registry := registry.NewRegistry()

	metadata := interfaces.ProviderMetadata{
		Name:           "test-provider",
		DisplayName:    "Test Provider",
		Version:        "1.0.0",
		Description:    "A test provider",
		Author:         "Test Author",
		RequiredConfig: []string{"api_key"},
	}

	factory := &MockProviderFactory{metadata: metadata}

	// Register provider
	err := registry.RegisterProvider("test-provider", factory)
	require.NoError(t, err, "Registration should succeed")

	// Test successful metadata retrieval
	retrievedMetadata, err := registry.GetProviderMetadata("test-provider")
	require.NoError(t, err, "Should retrieve metadata successfully")
	require.Equal(t, metadata.Name, retrievedMetadata.Name, "Name should match")
	require.Equal(t, metadata.DisplayName, retrievedMetadata.DisplayName, "DisplayName should match")
	require.Equal(t, metadata.Version, retrievedMetadata.Version, "Version should match")
	require.Equal(t, metadata.Author, retrievedMetadata.Author, "Author should match")

	// Test metadata retrieval for non-existent provider
	_, err = registry.GetProviderMetadata("non-existent")
	require.Error(t, err, "Should fail to retrieve metadata for non-existent provider")
}

func TestRegistry_IsRegistered(t *testing.T) {
	t.Parallel()

	registry := registry.NewRegistry()

	// Test non-registered provider
	require.False(t, registry.IsRegistered("test-provider"), "Should not be registered initially")

	// Register provider
	metadata := interfaces.ProviderMetadata{
		Name:        "test-provider",
		DisplayName: "Test Provider",
		Version:     "1.0.0",
		Description: "A test provider",
	}
	factory := &MockProviderFactory{metadata: metadata}

	err := registry.RegisterProvider("test-provider", factory)
	require.NoError(t, err, "Registration should succeed")

	// Test registered provider
	require.True(t, registry.IsRegistered("test-provider"), "Should be registered after registration")
	require.False(t, registry.IsRegistered("other-provider"), "Other provider should not be registered")
}

func TestRegistry_GetAllMetadata(t *testing.T) {
	t.Parallel()

	registry := registry.NewRegistry()

	// Test empty registry
	metadata := registry.GetAllMetadata()
	require.Empty(t, metadata, "Empty registry should return empty metadata")

	// Register multiple providers
	providerNames := []string{"provider-a", "provider-b"}
	for _, name := range providerNames {
		providerMetadata := interfaces.ProviderMetadata{
			Name:        name,
			DisplayName: name + " Provider",
			Version:     "1.0.0",
			Description: "A test provider",
		}
		factory := &MockProviderFactory{metadata: providerMetadata}

		err := registry.RegisterProvider(name, factory)
		require.NoError(t, err, "Registration should succeed for %s", name)
	}

	// Test getting all metadata
	allMetadata := registry.GetAllMetadata()
	require.Len(t, allMetadata, 2, "Should have metadata for 2 providers")

	for _, name := range providerNames {
		require.Contains(t, allMetadata, name, "Should contain metadata for %s", name)
		require.Equal(t, name, allMetadata[name].Name, "Name should match for %s", name)
	}
}

func TestRegistry_ValidateProvider(t *testing.T) {
	t.Parallel()

	reg := registry.NewRegistry()
	config := registry.NewConfigAdapter(map[string]interface{}{
		"api_key": "test-key",
		"region":  "us-east",
	})

	// Test validation for non-existent provider
	err := reg.ValidateProvider("non-existent", config)
	require.Error(t, err, "Should fail to validate non-existent provider")

	// Register provider
	metadata := interfaces.ProviderMetadata{
		Name:        "test-provider",
		DisplayName: "Test Provider",
		Version:     "1.0.0",
		Description: "A test provider",
	}
	factory := &MockProviderFactory{metadata: metadata}

	err = reg.RegisterProvider("test-provider", factory)
	require.NoError(t, err, "Registration should succeed")

	// Test successful validation
	err = reg.ValidateProvider("test-provider", config)
	require.NoError(t, err, "Validation should succeed")
}

func TestRegistry_Reset(t *testing.T) {
	t.Parallel()

	registry := registry.NewRegistry()

	// Register a provider
	metadata := interfaces.ProviderMetadata{
		Name:        "test-provider",
		DisplayName: "Test Provider",
		Version:     "1.0.0",
		Description: "A test provider",
	}
	factory := &MockProviderFactory{metadata: metadata}

	err := registry.RegisterProvider("test-provider", factory)
	require.NoError(t, err, "Registration should succeed")
	require.Equal(t, 1, registry.Count(), "Should have one provider")

	// Reset registry
	registry.Reset()
	require.Equal(t, 0, registry.Count(), "Should be empty after reset")
	require.False(t, registry.IsRegistered("test-provider"), "Provider should not be registered after reset")
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	registry := registry.NewRegistry()

	// Test concurrent registration
	const numProviders = 10
	errors := make(chan error, numProviders)

	for i := range numProviders {
		go func(index int) {
			metadata := interfaces.ProviderMetadata{
				Name:        fmt.Sprintf("provider-%d", index),
				DisplayName: fmt.Sprintf("Provider %d", index),
				Version:     "1.0.0",
				Description: "A test provider",
			}
			factory := &MockProviderFactory{metadata: metadata}

			err := registry.RegisterProvider(fmt.Sprintf("provider-%d", index), factory)
			errors <- err
		}(i)
	}

	// Wait for all registrations to complete
	for range numProviders {
		err := <-errors
		require.NoError(t, err, "Concurrent registration should succeed")
	}

	require.Equal(t, numProviders, registry.Count(), "Should have all providers registered")
}

func TestGlobalRegistryFunctions(t *testing.T) {
	t.Parallel()

	// Note: These tests interact with the global registry, so they may interfere
	// with each other if run in parallel. In a real implementation, you might
	// want to provide a way to use a custom registry for testing.

	// Get the default registry and reset it for testing
	defaultRegistry := registry.GetDefaultRegistry()
	defaultRegistry.Reset()

	metadata := interfaces.ProviderMetadata{
		Name:        "global-test-provider",
		DisplayName: "Global Test Provider",
		Version:     "1.0.0",
		Description: "A test provider for global functions",
	}
	factory := &MockProviderFactory{metadata: metadata}

	// Test global registration
	err := registry.RegisterProvider("global-test-provider", factory)
	require.NoError(t, err, "Global registration should succeed")

	// Test global listing
	providers := registry.ListProviders()
	require.Contains(t, providers, "global-test-provider", "Global list should contain the provider")

	// Test global retrieval
	provider, err := registry.GetProvider("global-test-provider")
	require.NoError(t, err, "Global retrieval should succeed")
	require.Equal(t, "global-test-provider", provider.Name(), "Provider name should match")

	// Test global metadata retrieval
	retrievedMetadata, err := registry.GetProviderMetadata("global-test-provider")
	require.NoError(t, err, "Global metadata retrieval should succeed")
	require.Equal(t, metadata.Name, retrievedMetadata.Name, "Metadata should match")

	// Test global registration check
	require.True(t, registry.IsRegistered("global-test-provider"), "Should be registered globally")

	// Clean up
	defaultRegistry.Reset()
}
