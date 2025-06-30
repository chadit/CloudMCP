// Package linode implements the CloudMCP provider interface for Linode cloud services.
// This package provides a plugin-based implementation that integrates with the CloudMCP
// provider registry and can be used alongside other cloud provider implementations.
package linode

import (
	"context"
	"errors"
	"fmt"

	"github.com/chadit/CloudMCP/pkg/interfaces"
	"github.com/chadit/CloudMCP/pkg/logger"
)

// Provider implements the CloudProvider interface for Linode.
// This is a simplified implementation that demonstrates the plugin architecture
// without requiring tight integration with the existing service.
type Provider struct {
	config      interfaces.Config
	logger      logger.Logger
	initialized bool
}

// Static errors for err113 compliance.
var (
	ErrConfigKeyMissing         = errors.New("required configuration key is missing")
	ErrDefaultAccountEmpty      = errors.New("default_linode_account cannot be empty")
	ErrDefaultAccountTokenEmpty = errors.New("no token found for default account")
	ErrNoDefaultAccount         = errors.New("no default account configured")
	ErrNoTokenConfigured        = errors.New("no token configured for default account")
)

// NewProvider creates a new Linode provider instance.
// The provider is not initialized and must be configured before use.
func NewProvider() *Provider {
	return &Provider{
		initialized: false,
	}
}

// Name returns the unique identifier for this cloud provider.
func (p *Provider) Name() string {
	return "linode"
}

// Initialize sets up the provider with the given configuration.
func (p *Provider) Initialize(_ context.Context, config interfaces.Config) error {
	if p.initialized {
		return interfaces.ErrProviderAlreadyInitialized
	}

	// Validate configuration before proceeding
	if err := p.ValidateConfig(config); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Store configuration for later use
	p.config = config

	// Create a simple logger (using basic implementation)
	p.logger = logger.New("info")

	p.initialized = true
	p.logger.Info("Linode provider initialized successfully")

	return nil
}

// RegisterTools registers all MCP tools provided by this cloud provider.
func (p *Provider) RegisterTools(server interfaces.MCPServer) error {
	if !p.initialized {
		return interfaces.ErrProviderNotInitialized
	}

	// For this demonstration, create a few sample tools
	tools := p.createSampleTools()

	p.logger.Info("Registering Linode tools", "count", len(tools))

	for _, tool := range tools {
		if err := server.RegisterTool(tool); err != nil {
			return fmt.Errorf("failed to register tool %q: %w", tool.Definition().Name, err)
		}
	}

	p.logger.Info("Successfully registered all Linode tools")

	return nil
}

// createSampleTools creates sample tools for demonstration purposes.
func (p *Provider) createSampleTools() []interfaces.Tool {
	// Create sample tools that demonstrate the plugin architecture
	return []interfaces.Tool{
		p.createAccountInfoTool(),
		p.createInstanceListTool(),
	}
}

// createAccountInfoTool creates a sample account info tool.
func (p *Provider) createAccountInfoTool() interfaces.Tool { //nolint:ireturn // Factory method should return interface
	return &SampleTool{
		name:        "linode_account_info",
		description: "Get Linode account information",
		config:      p.config,
		logger:      p.logger,
	}
}

// createInstanceListTool creates a sample instance list tool.
func (p *Provider) createInstanceListTool() interfaces.Tool { //nolint:ireturn // Factory method should return interface
	return &SampleTool{
		name:        "linode_instances_list",
		description: "List Linode instances",
		config:      p.config,
		logger:      p.logger,
	}
}

// GetCapabilities returns the list of capabilities supported by this provider.
func (p *Provider) GetCapabilities() []interfaces.Capability {
	return []interfaces.Capability{
		{
			Name:        "compute",
			Description: "Linode compute instance management",
			Version:     "1.0.0",
			Category:    "infrastructure",
		},
		{
			Name:        "storage",
			Description: "Linode block storage and object storage management",
			Version:     "1.0.0",
			Category:    "infrastructure",
		},
		{
			Name:        "networking",
			Description: "Linode networking services including VPCs, firewalls, and load balancers",
			Version:     "1.0.0",
			Category:    "infrastructure",
		},
		{
			Name:        "dns",
			Description: "Linode DNS domain and record management",
			Version:     "1.0.0",
			Category:    "infrastructure",
		},
		{
			Name:         "kubernetes",
			Description:  "Linode Kubernetes Engine (LKE) cluster management",
			Version:      "1.0.0",
			Category:     "infrastructure",
			Dependencies: []string{"compute", "networking"},
		},
		{
			Name:        "databases",
			Description: "Linode managed database services",
			Version:     "1.0.0",
			Category:    "infrastructure",
		},
		{
			Name:        "monitoring",
			Description: "Linode monitoring and alerting services",
			Version:     "1.0.0",
			Category:    "observability",
		},
		{
			Name:        "support",
			Description: "Linode support ticket management",
			Version:     "1.0.0",
			Category:    "support",
		},
		{
			Name:        "account",
			Description: "Multi-account support and account switching",
			Version:     "1.0.0",
			Category:    "management",
		},
	}
}

// ValidateConfig validates the provider-specific configuration.
func (p *Provider) ValidateConfig(config interfaces.Config) error {
	// Check required configuration keys
	requiredKeys := []string{
		"default_linode_account",
	}

	for _, key := range requiredKeys {
		if !config.IsSet(key) {
			return fmt.Errorf("%w: %q", ErrConfigKeyMissing, key)
		}
	}

	// Validate account configuration
	defaultAccount := config.GetString("default_linode_account")
	if defaultAccount == "" {
		return ErrDefaultAccountEmpty
	}

	// Check for at least one account token
	accountTokenKey := fmt.Sprintf("linode_accounts_%s_token", defaultAccount)
	if !config.IsSet(accountTokenKey) {
		return fmt.Errorf("%w %q (expected key: %q)", ErrDefaultAccountTokenEmpty, defaultAccount, accountTokenKey)
	}

	return nil
}

// Shutdown gracefully shuts down the provider and cleans up resources.
func (p *Provider) Shutdown(_ context.Context) error {
	if !p.initialized {
		return nil // Already shut down or never initialized
	}

	p.logger.Info("Shutting down Linode provider")

	// The current service implementation doesn't have a shutdown method,
	// but if it did, we would call it here

	p.initialized = false
	p.logger.Info("Linode provider shutdown complete")

	return nil
}

// HealthCheck returns the current health status of the provider.
func (p *Provider) HealthCheck(_ context.Context) error {
	if !p.initialized {
		return interfaces.ErrProviderNotInitialized
	}

	// Perform a basic health check by validating configuration
	if !p.config.IsSet("default_linode_account") {
		return ErrNoDefaultAccount
	}

	defaultAccount := p.config.GetString("default_linode_account")
	tokenKey := fmt.Sprintf("linode_accounts_%s_token", defaultAccount)

	if !p.config.IsSet(tokenKey) {
		return fmt.Errorf("%w %q", ErrNoTokenConfigured, defaultAccount)
	}

	return nil
}

// IsInitialized returns whether the provider has been initialized.
func (p *Provider) IsInitialized() bool {
	return p.initialized
}

// Factory implements the ProviderFactory interface for the Linode provider.
type Factory struct{}

// NewFactory creates a new Linode provider factory.
func NewFactory() *Factory {
	return &Factory{}
}

// CreateProvider creates a new instance of the Linode cloud provider.
func (f *Factory) CreateProvider() interfaces.CloudProvider { //nolint:ireturn // Factory method should return interface
	return NewProvider()
}

// GetMetadata returns metadata about the Linode provider.
func (f *Factory) GetMetadata() interfaces.ProviderMetadata {
	return interfaces.ProviderMetadata{
		Name:        "linode",
		DisplayName: "Linode Cloud",
		Version:     "1.0.0",
		Description: "Complete Linode cloud infrastructure management with multi-account support",
		Author:      "CloudMCP Team",
		Homepage:    "https://github.com/chadit/CloudMCP",
		License:     "MIT",
		RequiredConfig: []string{
			"default_linode_account",
			"linode_accounts_{account}_token",
		},
		OptionalConfig: []string{
			"linode_accounts_{account}_label",
			"enable_metrics",
			"metrics_port",
			"log_level",
			"log_format",
			"server_name",
		},
		Capabilities: NewProvider().GetCapabilities(),
	}
}

// ValidateConfig validates configuration specific to the Linode provider.
func (f *Factory) ValidateConfig(config interfaces.Config) error {
	provider := NewProvider()

	return provider.ValidateConfig(config)
}
