// Package interfaces defines core interfaces for CloudMCP extensibility.
// It provides the foundation for multi-provider support and plugin architecture,
// enabling clean separation between different cloud provider implementations.
package interfaces

import (
	"context"
	"errors"

	"github.com/mark3labs/mcp-go/mcp"
)

// CloudProvider defines the interface that all cloud provider implementations must satisfy.
// This interface enables CloudMCP to support multiple cloud providers through a unified API.
type CloudProvider interface {
	// Name returns the unique identifier for this cloud provider (e.g., "linode", "aws", "gcp").
	Name() string

	// Initialize sets up the provider with the given configuration.
	// This method is called once during CloudMCP startup to establish
	// authentication, load configuration, and prepare resources.
	Initialize(ctx context.Context, config Config) error

	// RegisterTools registers all MCP tools provided by this cloud provider.
	// Each provider should register its complete set of management tools
	// using the provided MCP server instance.
	RegisterTools(server MCPServer) error

	// GetCapabilities returns the list of capabilities supported by this provider.
	// This allows CloudMCP to discover what features are available and
	// present appropriate options to users.
	GetCapabilities() []Capability

	// ValidateConfig validates the provider-specific configuration.
	// This method is called before Initialize to ensure all required
	// configuration parameters are present and valid.
	ValidateConfig(config Config) error

	// Shutdown gracefully shuts down the provider and cleans up resources.
	// This method is called during CloudMCP shutdown to ensure proper cleanup
	// of connections, caches, and other provider-specific resources.
	Shutdown(ctx context.Context) error

	// HealthCheck returns the current health status of the provider.
	// This enables monitoring and ensures the provider is functioning correctly.
	HealthCheck(ctx context.Context) error
}

// Config represents provider-specific configuration data.
// Each provider implementation defines its own configuration structure
// that satisfies this interface.
type Config interface {
	// GetString retrieves a string configuration value by key.
	GetString(key string) string

	// GetBool retrieves a boolean configuration value by key.
	GetBool(key string) bool

	// GetInt retrieves an integer configuration value by key.
	GetInt(key string) int

	// GetStringMap retrieves a map of string values by key.
	GetStringMap(key string) map[string]string

	// IsSet checks if a configuration key has been set.
	IsSet(key string) bool

	// Validate validates the entire configuration.
	Validate() error
}

// MCPServer defines the interface for registering MCP tools.
// This abstracts the MCP server implementation to allow for testing
// and different MCP server versions.
type MCPServer interface {
	// RegisterTool registers a single MCP tool with the server.
	RegisterTool(tool Tool) error

	// RegisterTools registers multiple MCP tools at once.
	RegisterTools(tools []Tool) error

	// GetRegisteredTools returns the list of all registered tools.
	GetRegisteredTools() []Tool
}

// Tool represents an MCP tool that can be registered with the server.
// This interface encapsulates the tool definition and execution logic.
type Tool interface {
	// Definition returns the MCP tool definition including name, description,
	// and input schema. This is used by the MCP protocol for tool discovery.
	Definition() mcp.Tool

	// Execute handles the actual tool execution with the provided parameters.
	// The context allows for cancellation and timeout handling.
	Execute(ctx context.Context, params map[string]any) (any, error)

	// Validate validates the tool parameters before execution.
	// This allows for early parameter validation and better error messages.
	Validate(params map[string]any) error
}

// Capability represents a feature or capability provided by a cloud provider.
// This enables feature discovery and helps users understand what each provider supports.
type Capability struct {
	// Name is the unique identifier for this capability (e.g., "compute", "storage", "networking").
	Name string `json:"name"`

	// Description provides a human-readable description of the capability.
	Description string `json:"description"`

	// Version indicates the version of this capability implementation.
	Version string `json:"version"`

	// Category groups related capabilities together (e.g., "infrastructure", "security", "monitoring").
	Category string `json:"category"`

	// Dependencies lists other capabilities that this capability requires.
	Dependencies []string `json:"dependencies,omitempty"`

	// Experimental indicates if this capability is experimental and may change.
	Experimental bool `json:"experimental,omitempty"`
}

// ProviderMetadata contains metadata about a cloud provider implementation.
// This information is used for provider discovery and management.
type ProviderMetadata struct {
	// Name is the unique identifier for the provider.
	Name string `json:"name"`

	// DisplayName is the human-readable name of the provider.
	DisplayName string `json:"displayName"`

	// Version is the version of the provider implementation.
	Version string `json:"version"`

	// Description provides details about what this provider supports.
	Description string `json:"description"`

	// Author information for the provider implementation.
	Author string `json:"author"`

	// Homepage URL for documentation and support.
	Homepage string `json:"homepage,omitempty"`

	// License under which the provider is distributed.
	License string `json:"license,omitempty"`

	// RequiredConfig lists the configuration keys required by this provider.
	RequiredConfig []string `json:"requiredConfig"`

	// OptionalConfig lists the optional configuration keys supported by this provider.
	OptionalConfig []string `json:"optionalConfig,omitempty"`

	// Capabilities lists all capabilities provided by this provider.
	Capabilities []Capability `json:"capabilities"`
}

// ProviderFactory creates new instances of a cloud provider.
// This factory pattern allows for proper initialization and configuration
// of provider instances without exposing implementation details.
type ProviderFactory interface {
	// CreateProvider creates a new instance of the cloud provider.
	// The returned provider is not yet initialized and should be configured
	// before calling Initialize().
	CreateProvider() CloudProvider

	// GetMetadata returns metadata about the provider type.
	// This information is used for provider discovery and validation.
	GetMetadata() ProviderMetadata

	// ValidateConfig validates configuration specific to this provider type.
	// This is called before creating a provider instance to ensure
	// the configuration is compatible.
	ValidateConfig(config Config) error
}

// Common errors used by providers.
var (
	// ErrProviderNotInitialized is returned when an operation is attempted on an uninitialized provider.
	ErrProviderNotInitialized = errors.New("provider is not initialized")

	// ErrProviderAlreadyInitialized is returned when Initialize is called on an already initialized provider.
	ErrProviderAlreadyInitialized = errors.New("provider is already initialized")

	// ErrConfigValidationFailed is returned when configuration validation fails.
	ErrConfigValidationFailed = errors.New("configuration validation failed")
)

// ProviderRegistry manages the registration and lifecycle of cloud providers.
// It provides a central point for provider discovery and management.
type ProviderRegistry interface {
	// RegisterProvider registers a provider factory with the registry.
	// The provider can then be created and used by CloudMCP.
	RegisterProvider(name string, factory ProviderFactory) error

	// GetProvider retrieves a provider instance by name.
	// Returns an error if the provider is not registered or fails to create.
	GetProvider(name string) (CloudProvider, error)

	// ListProviders returns the names of all registered providers.
	ListProviders() []string

	// GetProviderMetadata returns metadata for a specific provider.
	GetProviderMetadata(name string) (ProviderMetadata, error)

	// IsRegistered checks if a provider with the given name is registered.
	IsRegistered(name string) bool
}
