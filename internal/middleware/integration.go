// Package middleware provides integration examples showing how to use the middleware
// system with the CloudMCP plugin architecture and MCP server.
package middleware

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/chadit/CloudMCP/internal/registry"
	"github.com/chadit/CloudMCP/pkg/interfaces"
	pkglogger "github.com/chadit/CloudMCP/pkg/logger"
)

const (
	unknownProvider = "unknown"

	// Middleware priority constants.
	priorityRecovery        = 1
	prioritySecurity        = 5
	priorityLogging         = 10
	priorityStructured      = 15
	priorityMetrics         = 20
	priorityUsage           = 25
	priorityRateLimit       = 30
	priorityAdaptive        = 35
	priorityRetry           = 40
	priorityCircuitBreaker  = 45
	priorityErrorEnrichment = 50

	// Configuration constants.
	integrationDefaultRateLimit = 100
	defaultRetryAttempts        = 3
	defaultTimeoutSeconds       = 30
	defaultBackoffMultiplier    = 2.0
	retryBackoffMultiplier      = 2.0
	circuitBreakerThreshold     = 5
	circuitBreakerTimeout       = 60
	adaptiveLoadThreshold       = 50
	adaptiveMaxRequests         = 1000

	// Production environment constants.
	productionRateLimit               = 50
	productionRateLimitRefill         = 50
	productionCircuitFailureThreshold = 3
	productionCircuitRecoveryTimeout  = 30
	productionCircuitSuccessThreshold = 2

	// Development environment constants.
	developmentRateLimit       = 1000
	developmentRateLimitRefill = 1000
)

// ErrToolNotFound indicates that a tool was not found in a provider.
var ErrToolNotFound = errors.New("tool not found")

// Manager manages the complete middleware stack for CloudMCP.
// It provides a high-level interface for configuring and executing middleware
// around tool operations.
type Manager struct {
	chain     *Chain
	collector MetricsCollector
	logger    pkglogger.Logger
}

// NewManager creates a new middleware manager with default configuration.
func NewManager(logger pkglogger.Logger, collector MetricsCollector) *Manager {
	if logger == nil {
		logger = pkglogger.New("info")
	}

	if collector == nil {
		collector = NewLogBasedMetricsCollector(logger)
	}

	return &Manager{
		chain:     NewChain(logger),
		collector: collector,
		logger:    logger,
	}
}

// ConfigureDefault sets up a default middleware stack with common middleware components.
func (mm *Manager) ConfigureDefault() *Manager {
	// Recovery middleware (highest priority - wraps everything)
	recoveryConfig := NewConfig().WithPriority(priorityRecovery)
	mm.chain.Add(NewRecoveryMiddleware(recoveryConfig, mm.logger))

	// Security logging (very high priority)
	securityConfig := NewConfig().WithPriority(prioritySecurity)
	mm.chain.Add(NewSecurityLoggingMiddleware(securityConfig, mm.logger))

	// Basic logging
	loggingConfig := NewConfig().WithPriority(priorityLogging)
	loggingConfig.WithConfig("log_parameters", false) // Don't log parameters by default
	loggingConfig.WithConfig("log_results", false)    // Don't log results by default
	mm.chain.Add(NewLoggingMiddleware(loggingConfig, mm.logger))

	// Structured logging
	structuredConfig := NewConfig().WithPriority(priorityStructured)
	mm.chain.Add(NewStructuredLoggingMiddleware(structuredConfig, mm.logger))

	// Metrics collection
	metricsConfig := NewConfig().WithPriority(priorityMetrics)
	mm.chain.Add(NewMetricsMiddleware(metricsConfig, mm.logger, mm.collector))

	// Usage metrics
	usageConfig := NewConfig().WithPriority(priorityUsage)
	mm.chain.Add(NewUsageMetricsMiddleware(usageConfig, mm.logger, mm.collector))

	// Rate limiting (before expensive operations)
	rateLimitConfig := NewConfig().WithPriority(priorityRateLimit)
	rateLimitConfig.WithConfig("key_strategy", "per_tool") // Rate limit per tool

	limiter := NewTokenBucket(defaultRateLimit, time.Minute, defaultRateLimit) // 100 requests per minute per tool
	mm.chain.Add(NewRateLimitMiddleware(rateLimitConfig, mm.logger, limiter))

	// Adaptive rate limiting
	adaptiveConfig := NewConfig().WithPriority(priorityAdaptive)
	adaptiveLimiter := NewTokenBucket(adaptiveLoadThreshold, time.Minute, adaptiveLoadThreshold) // More restrictive base limit
	mm.chain.Add(NewAdaptiveRateLimitMiddleware(adaptiveConfig, mm.logger, adaptiveLimiter))

	// Retry logic
	retryConfig := NewConfig().WithPriority(priorityRetry)
	retryConfig.WithConfig("max_retries", defaultRetryAttempts)
	retryConfig.WithConfig("base_delay", time.Second)
	retryConfig.WithConfig("max_delay", defaultTimeoutSeconds*time.Second)
	retryConfig.WithConfig("backoff_factor", retryBackoffMultiplier)
	mm.chain.Add(NewRetryMiddleware(retryConfig, mm.logger))

	// Circuit breaker
	circuitConfig := NewConfig().WithPriority(priorityCircuitBreaker)
	circuitConfig.WithConfig("failure_threshold", circuitBreakerThreshold)
	circuitConfig.WithConfig("recovery_timeout", circuitBreakerTimeout*time.Second)
	circuitConfig.WithConfig("success_threshold", defaultRetryAttempts)
	mm.chain.Add(NewCircuitBreakerMiddleware(circuitConfig, mm.logger))

	// Error enrichment (after execution)
	errorConfig := NewConfig().WithPriority(priorityErrorEnrichment)
	mm.chain.Add(NewErrorEnrichmentMiddleware(errorConfig, mm.logger))

	return mm
}

// ConfigureProduction sets up a production-ready middleware stack.
func (mm *Manager) ConfigureProduction() *Manager {
	// Start with default configuration
	mm.ConfigureDefault()

	// Override specific settings for production

	// More restrictive rate limiting for production
	mm.chain.Remove("rate_limit")

	rateLimitConfig := NewConfig().WithPriority(priorityRateLimit)
	rateLimitConfig.WithConfig("key_strategy", "per_user_provider") // Rate limit per user and provider

	limiter := NewTokenBucket(productionRateLimit, time.Minute, productionRateLimitRefill)
	mm.chain.Add(NewRateLimitMiddleware(rateLimitConfig, mm.logger, limiter))

	// More aggressive circuit breaker for production
	mm.chain.Remove("circuit_breaker")

	circuitConfig := NewConfig().WithPriority(priorityCircuitBreaker)
	circuitConfig.WithConfig("failure_threshold", productionCircuitFailureThreshold) // Lower threshold
	circuitConfig.WithConfig("recovery_timeout", productionCircuitRecoveryTimeout*time.Second)
	circuitConfig.WithConfig("success_threshold", productionCircuitSuccessThreshold)
	mm.chain.Add(NewCircuitBreakerMiddleware(circuitConfig, mm.logger))

	// Enable security logging for sensitive operations
	mm.chain.Remove("security_logging")

	securityConfig := NewConfig().WithPriority(prioritySecurity)
	sensitiveOps := []string{"instance_delete", "volume_delete", "database_delete"}
	securityConfig.WithConfig("sensitive_operations", sensitiveOps)
	mm.chain.Add(NewSecurityLoggingMiddleware(securityConfig, mm.logger))

	return mm
}

// ConfigureDevelopment sets up a development-friendly middleware stack.
func (mm *Manager) ConfigureDevelopment() *Manager {
	// Recovery middleware
	recoveryConfig := NewConfig().WithPriority(priorityRecovery)
	mm.chain.Add(NewRecoveryMiddleware(recoveryConfig, mm.logger))

	// Verbose logging for development
	loggingConfig := NewConfig().WithPriority(priorityLogging)
	loggingConfig.WithConfig("log_parameters", true) // Log parameters in dev
	loggingConfig.WithConfig("log_results", true)    // Log results in dev
	mm.chain.Add(NewLoggingMiddleware(loggingConfig, mm.logger))

	// Basic metrics
	metricsConfig := NewConfig().WithPriority(priorityMetrics)
	mm.chain.Add(NewMetricsMiddleware(metricsConfig, mm.logger, mm.collector))

	// Lenient rate limiting for development
	rateLimitConfig := NewConfig().WithPriority(priorityRateLimit)
	rateLimitConfig.WithConfig("key_strategy", "per_tool")

	limiter := NewTokenBucket(developmentRateLimit, time.Minute, developmentRateLimitRefill) // Very high limit for dev
	mm.chain.Add(NewRateLimitMiddleware(rateLimitConfig, mm.logger, limiter))

	// No retry in development (fail fast)
	// No circuit breaker in development (let errors through)

	return mm
}

// ExecuteTool executes a tool through the complete middleware chain.
func (mm *Manager) ExecuteTool(ctx context.Context, tool interfaces.Tool, params map[string]any) (any, error) {
	// Create execution context if not present
	_, hasCtx := GetExecutionContext(ctx)
	if !hasCtx {
		toolName := tool.Definition().Name
		// Try to determine provider from tool name or context
		provider := unknownProvider

		if len(toolName) > 0 {
			// Simple heuristic: look for provider prefix
			if len(toolName) > 6 && toolName[:6] == "linode" {
				provider = "linode"
			}
		}

		newExecCtx := NewExecutionContext(toolName, provider)
		ctx = WithExecutionContext(ctx, newExecCtx)
	}

	// Define the base handler that actually executes the tool
	baseHandler := func(ctx context.Context, tool interfaces.Tool, params map[string]any) (any, error) {
		// Validate parameters
		if err := tool.Validate(params); err != nil {
			return nil, fmt.Errorf("tool validation failed: %w", err)
		}

		// Execute the tool
		return tool.Execute(ctx, params)
	}

	// Execute through middleware chain
	return mm.chain.Execute(ctx, tool, params, baseHandler)
}

// AddMiddleware adds a custom middleware to the chain.
func (mm *Manager) AddMiddleware(middleware Middleware) *Manager {
	mm.chain.Add(middleware)

	return mm
}

// RemoveMiddleware removes a middleware by name.
func (mm *Manager) RemoveMiddleware(name string) bool {
	return mm.chain.Remove(name)
}

// GetMiddlewareList returns all registered middlewares.
func (mm *Manager) GetMiddlewareList() []Middleware {
	return mm.chain.List()
}

// Clear removes all middlewares.
func (mm *Manager) Clear() *Manager {
	mm.chain.Clear()

	return mm
}

// ToolExecutorAdapter adapts the middleware manager to work with the MCP server adapter.
// This allows tools to be executed through the middleware chain when called via MCP.
type ToolExecutorAdapter struct {
	manager *Manager
}

// NewToolExecutorAdapter creates a new tool executor adapter.
func NewToolExecutorAdapter(manager *Manager) *ToolExecutorAdapter {
	return &ToolExecutorAdapter{
		manager: manager,
	}
}

// WrapMCPServer wraps an MCP server adapter to use middleware for tool execution.
func (tea *ToolExecutorAdapter) WrapMCPServer(server *registry.MCPServerAdapter) *WrappedMCPServer {
	return &WrappedMCPServer{
		server:  server,
		adapter: tea,
	}
}

// WrappedMCPServer wraps an MCP server to execute tools through middleware.
type WrappedMCPServer struct {
	server  *registry.MCPServerAdapter
	adapter *ToolExecutorAdapter
}

// ExecuteToolWithMiddleware executes a tool by name through the middleware chain.
func (wms *WrappedMCPServer) ExecuteToolWithMiddleware(ctx context.Context, toolName string, params map[string]any) (any, error) {
	// Get the tool from the server
	tool, err := wms.server.GetTool(toolName)
	if err != nil {
		return nil, fmt.Errorf("failed to get tool %q: %w", toolName, err)
	}

	// Execute through middleware
	return wms.adapter.manager.ExecuteTool(ctx, tool, params)
}

// GetUnderlyingServer returns the underlying MCP server adapter.
func (wms *WrappedMCPServer) GetUnderlyingServer() *registry.MCPServerAdapter {
	return wms.server
}

// ProviderMiddlewareIntegration shows how to integrate middleware with the provider system.
type ProviderMiddlewareIntegration struct {
	registry *registry.Registry
	manager  *Manager
	logger   pkglogger.Logger
}

// NewProviderMiddlewareIntegration creates a new provider-middleware integration.
func NewProviderMiddlewareIntegration(registry *registry.Registry, manager *Manager, logger pkglogger.Logger) *ProviderMiddlewareIntegration {
	return &ProviderMiddlewareIntegration{
		registry: registry,
		manager:  manager,
		logger:   logger,
	}
}

// ExecuteProviderTool executes a tool from a specific provider through middleware.
func (pmi *ProviderMiddlewareIntegration) ExecuteProviderTool(ctx context.Context, providerName, toolName string, params map[string]any) (any, error) {
	// Get the provider
	provider, err := pmi.registry.GetProvider(providerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider %q: %w", providerName, err)
	}

	// Create a mock MCP server to register tools
	mockServer := &mockMCPServerForIntegration{
		tools: make(map[string]interfaces.Tool),
	}

	// Register provider tools
	if err := provider.RegisterTools(mockServer); err != nil {
		return nil, fmt.Errorf("failed to register tools for provider %q: %w", providerName, err)
	}

	// Get the specific tool
	tool, exists := mockServer.tools[toolName]
	if !exists {
		return nil, fmt.Errorf("%w %q in provider %q", ErrToolNotFound, toolName, providerName)
	}

	// Create execution context with provider information
	execCtx := NewExecutionContext(toolName, providerName)
	ctx = WithExecutionContext(ctx, execCtx)

	// Execute through middleware
	return pmi.manager.ExecuteTool(ctx, tool, params)
}

// mockMCPServerForIntegration implements interfaces.MCPServer for integration testing.
type mockMCPServerForIntegration struct {
	tools map[string]interfaces.Tool
}

func (m *mockMCPServerForIntegration) RegisterTool(tool interfaces.Tool) error {
	m.tools[tool.Definition().Name] = tool

	return nil
}

func (m *mockMCPServerForIntegration) RegisterTools(tools []interfaces.Tool) error {
	for _, tool := range tools {
		if err := m.RegisterTool(tool); err != nil {
			return err
		}
	}

	return nil
}

func (m *mockMCPServerForIntegration) GetRegisteredTools() []interfaces.Tool {
	tools := make([]interfaces.Tool, 0, len(m.tools))
	for _, tool := range m.tools {
		tools = append(tools, tool)
	}

	return tools
}

// Example configurations for different environments.

// GetDefaultConfig returns a default middleware configuration.
func GetDefaultConfig() map[string]*Config {
	return map[string]*Config{
		"recovery":         NewConfig().WithPriority(priorityRecovery),
		"security_logging": NewConfig().WithPriority(prioritySecurity),
		"logging": NewConfig().WithPriority(priorityLogging).
			WithConfig("log_parameters", false).
			WithConfig("log_results", false),
		"metrics": NewConfig().WithPriority(priorityMetrics),
		"rate_limit": NewConfig().WithPriority(priorityRateLimit).
			WithConfig("key_strategy", "per_tool"),
		"retry": NewConfig().WithPriority(priorityRetry).
			WithConfig("max_retries", defaultRetryAttempts).
			WithConfig("base_delay", time.Second),
		"circuit_breaker": NewConfig().WithPriority(priorityCircuitBreaker).
			WithConfig("failure_threshold", circuitBreakerThreshold).
			WithConfig("recovery_timeout", circuitBreakerTimeout*time.Second),
		"error_enrichment": NewConfig().WithPriority(priorityErrorEnrichment),
	}
}

// GetProductionConfig returns a production-optimized middleware configuration.
func GetProductionConfig() map[string]*Config {
	config := GetDefaultConfig()

	// More restrictive rate limiting
	config["rate_limit"].WithConfig("key_strategy", "per_user_provider")

	// Lower circuit breaker threshold
	config["circuit_breaker"].WithConfig("failure_threshold", productionCircuitFailureThreshold)

	// Shorter recovery timeout
	config["circuit_breaker"].WithConfig("recovery_timeout", productionCircuitRecoveryTimeout*time.Second)

	return config
}

// GetDevelopmentConfig returns a development-friendly middleware configuration.
func GetDevelopmentConfig() map[string]*Config {
	return map[string]*Config{
		"recovery": NewConfig().WithPriority(priorityRecovery),
		"logging": NewConfig().WithPriority(priorityLogging).
			WithConfig("log_parameters", true).
			WithConfig("log_results", true),
		"metrics": NewConfig().WithPriority(priorityMetrics),
		"rate_limit": NewConfig().WithPriority(priorityRateLimit).
			WithConfig("key_strategy", "global"), // Less restrictive for dev
	}
}
