// Package registry provides provider registration and lifecycle management.
// It implements the core provider registry functionality that enables
// CloudMCP to discover, create, and manage multiple cloud provider implementations.
package registry

import (
	"errors"
	"fmt"
	"sync"

	"github.com/chadit/CloudMCP/pkg/interfaces"
)

// Registry-specific errors.
var (
	ErrProviderNameEmpty          = errors.New("provider name cannot be empty")
	ErrProviderFactoryNil         = errors.New("provider factory cannot be nil")
	ErrProviderAlreadyRegistered  = errors.New("provider is already registered")
	ErrProviderNotRegistered      = errors.New("provider is not registered")
	ErrProviderFactoryReturnedNil = errors.New("factory returned nil provider")
)

// Registry implements the ProviderRegistry interface and manages
// the registration and creation of cloud provider instances.
type Registry struct {
	factories map[string]interfaces.ProviderFactory
	mu        sync.RWMutex
}

// NewRegistry creates a new provider registry instance.
// The registry starts empty and providers must be registered before use.
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]interfaces.ProviderFactory),
	}
}

// RegisterProvider registers a provider factory with the registry.
// The provider name must be unique, and the factory will be used to create
// provider instances when requested.
func (r *Registry) RegisterProvider(name string, factory interfaces.ProviderFactory) error {
	if name == "" {
		return ErrProviderNameEmpty
	}

	if factory == nil {
		return ErrProviderFactoryNil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("%w: %q", ErrProviderAlreadyRegistered, name)
	}

	r.factories[name] = factory

	return nil
}

// GetProvider retrieves a provider instance by name.
// Creates a new provider instance using the registered factory.
// Returns an error if the provider is not registered or creation fails.
func (r *Registry) GetProvider(name string) (interfaces.CloudProvider, error) { //nolint:ireturn // Registry method should return interface
	r.mu.RLock()
	factory, exists := r.factories[name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("%w: %q", ErrProviderNotRegistered, name)
	}

	provider := factory.CreateProvider()
	if provider == nil {
		return nil, fmt.Errorf("%w: %q", ErrProviderFactoryReturnedNil, name)
	}

	return provider, nil
}

// ListProviders returns the names of all registered providers.
// The returned slice is a copy and safe to modify.
func (r *Registry) ListProviders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]string, 0, len(r.factories))
	for name := range r.factories {
		providers = append(providers, name)
	}

	return providers
}

// GetProviderMetadata returns metadata for a specific provider.
// Returns an error if the provider is not registered.
func (r *Registry) GetProviderMetadata(name string) (interfaces.ProviderMetadata, error) {
	r.mu.RLock()
	factory, exists := r.factories[name]
	r.mu.RUnlock()

	if !exists {
		return interfaces.ProviderMetadata{}, fmt.Errorf("%w: %q", ErrProviderNotRegistered, name)
	}

	return factory.GetMetadata(), nil
}

// IsRegistered checks if a provider with the given name is registered.
func (r *Registry) IsRegistered(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.factories[name]

	return exists
}

// GetAllMetadata returns metadata for all registered providers.
// This is useful for provider discovery and listing capabilities.
func (r *Registry) GetAllMetadata() map[string]interfaces.ProviderMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metadata := make(map[string]interfaces.ProviderMetadata)
	for name, factory := range r.factories {
		metadata[name] = factory.GetMetadata()
	}

	return metadata
}

// ValidateProvider validates that a provider can be created and properly configured.
// This performs validation without actually creating a provider instance.
func (r *Registry) ValidateProvider(name string, config interfaces.Config) error {
	r.mu.RLock()
	factory, exists := r.factories[name]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("%w: %q", ErrProviderNotRegistered, name)
	}

	// Validate the configuration with the provider factory
	if err := factory.ValidateConfig(config); err != nil {
		return fmt.Errorf("configuration validation failed for provider %q: %w", name, err)
	}

	return nil
}

// Reset clears all registered providers.
// This is primarily useful for testing purposes.
func (r *Registry) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.factories = make(map[string]interfaces.ProviderFactory)
}

// Count returns the number of registered providers.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.factories)
}

// defaultRegistry is the global registry instance used by CloudMCP.
// It provides convenient package-level functions for provider registration.
var defaultRegistry = NewRegistry() //nolint:gochecknoglobals // Package-level registry is intentional

// RegisterProvider registers a provider with the default global registry.
// This is a convenience function for the most common use case.
func RegisterProvider(name string, factory interfaces.ProviderFactory) error {
	return defaultRegistry.RegisterProvider(name, factory)
}

// GetProvider retrieves a provider from the default global registry.
// This is a convenience function for the most common use case.
func GetProvider(name string) (interfaces.CloudProvider, error) { //nolint:ireturn // Registry function should return interface
	return defaultRegistry.GetProvider(name)
}

// ListProviders returns all providers from the default global registry.
// This is a convenience function for the most common use case.
func ListProviders() []string {
	return defaultRegistry.ListProviders()
}

// GetProviderMetadata returns metadata from the default global registry.
// This is a convenience function for the most common use case.
func GetProviderMetadata(name string) (interfaces.ProviderMetadata, error) {
	return defaultRegistry.GetProviderMetadata(name)
}

// IsRegistered checks the default global registry.
// This is a convenience function for the most common use case.
func IsRegistered(name string) bool {
	return defaultRegistry.IsRegistered(name)
}

// GetDefaultRegistry returns the default global registry instance.
// This allows for advanced usage while maintaining the convenience functions.
func GetDefaultRegistry() *Registry {
	return defaultRegistry
}
