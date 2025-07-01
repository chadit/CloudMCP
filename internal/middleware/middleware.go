// Package middleware provides a comprehensive middleware system for CloudMCP tool execution.
// It implements the chain of responsibility pattern to handle cross-cutting concerns
// like logging, metrics collection, rate limiting, authentication, and error handling.
package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/chadit/CloudMCP/pkg/interfaces"
	pkglogger "github.com/chadit/CloudMCP/pkg/logger"
)

// ToolHandler represents a function that executes an MCP tool.
// This is the core handler signature that middleware wraps around.
type ToolHandler func(ctx context.Context, tool interfaces.Tool, params map[string]interface{}) (interface{}, error)

// Middleware represents a middleware component in the execution chain.
// Each middleware can modify the request, response, or handle cross-cutting concerns.
type Middleware interface {
	// Name returns the unique identifier for this middleware.
	Name() string

	// Execute processes the request and delegates to the next handler in the chain.
	Execute(ctx context.Context, tool interfaces.Tool, params map[string]interface{}, next ToolHandler) (interface{}, error)

	// Priority returns the execution priority (lower numbers execute first).
	Priority() int
}

// Chain represents a middleware execution chain.
// It manages the ordered execution of middleware components around tool handlers.
type Chain struct {
	middlewares []Middleware
	logger      pkglogger.Logger
}

// NewChain creates a new middleware chain with optional logger.
func NewChain(logger pkglogger.Logger) *Chain {
	if logger == nil {
		logger = pkglogger.New("info")
	}

	return &Chain{
		middlewares: make([]Middleware, 0),
		logger:      logger,
	}
}

// Add adds a middleware to the chain.
// Middlewares are automatically sorted by priority when executed.
func (c *Chain) Add(middleware Middleware) *Chain {
	c.middlewares = append(c.middlewares, middleware)

	return c
}

// Execute runs the middleware chain around the given tool handler.
// Middlewares are executed in priority order (lowest priority first).
func (c *Chain) Execute(ctx context.Context, tool interfaces.Tool, params map[string]interface{}, handler ToolHandler) (interface{}, error) {
	if len(c.middlewares) == 0 {
		return handler(ctx, tool, params)
	}

	// Sort middlewares by priority
	sortedMiddlewares := c.getSortedMiddlewares()

	// Build the execution chain from the end backwards
	finalHandler := handler

	for i := len(sortedMiddlewares) - 1; i >= 0; i-- {
		middleware := sortedMiddlewares[i]
		currentHandler := finalHandler

		finalHandler = func(ctx context.Context, tool interfaces.Tool, params map[string]interface{}) (interface{}, error) {
			return middleware.Execute(ctx, tool, params, currentHandler)
		}
	}

	return finalHandler(ctx, tool, params)
}

// getSortedMiddlewares returns middlewares sorted by priority.
func (c *Chain) getSortedMiddlewares() []Middleware {
	sorted := make([]Middleware, len(c.middlewares))
	copy(sorted, c.middlewares)

	// Simple bubble sort by priority (sufficient for small numbers of middleware)
	for i := range sorted {
		for j := range len(sorted) - 1 - i {
			if sorted[j].Priority() > sorted[j+1].Priority() {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}

// List returns all registered middlewares sorted by priority.
func (c *Chain) List() []Middleware {
	return c.getSortedMiddlewares()
}

// Remove removes a middleware by name from the chain.
func (c *Chain) Remove(name string) bool {
	for i, middleware := range c.middlewares {
		if middleware.Name() == name {
			c.middlewares = append(c.middlewares[:i], c.middlewares[i+1:]...)

			return true
		}
	}

	return false
}

// Clear removes all middlewares from the chain.
func (c *Chain) Clear() {
	c.middlewares = make([]Middleware, 0)
}

// Count returns the number of middlewares in the chain.
func (c *Chain) Count() int {
	return len(c.middlewares)
}

// Has checks if a middleware with the given name exists in the chain.
func (c *Chain) Has(name string) bool {
	for _, middleware := range c.middlewares {
		if middleware.Name() == name {
			return true
		}
	}

	return false
}

// ExecutionContext provides context information about the current tool execution.
// This is passed through the middleware chain to provide execution metadata.
type ExecutionContext struct {
	// StartTime when the execution began
	StartTime time.Time

	// ToolName being executed
	ToolName string

	// Provider name (e.g., "linode", "aws")
	Provider string

	// RequestID for tracing
	RequestID string

	// User or client identifier
	UserID string

	// Additional metadata
	Metadata map[string]interface{}
}

// NewExecutionContext creates a new execution context with default values.
func NewExecutionContext(toolName, provider string) *ExecutionContext {
	return &ExecutionContext{
		StartTime: time.Now(),
		ToolName:  toolName,
		Provider:  provider,
		RequestID: generateRequestID(),
		Metadata:  make(map[string]interface{}),
	}
}

// Duration returns how long the execution has been running.
func (ec *ExecutionContext) Duration() time.Duration {
	return time.Since(ec.StartTime)
}

// WithMetadata adds metadata to the execution context.
func (ec *ExecutionContext) WithMetadata(key string, value interface{}) *ExecutionContext {
	ec.Metadata[key] = value

	return ec
}

// GetMetadata retrieves metadata from the execution context.
func (ec *ExecutionContext) GetMetadata(key string) (interface{}, bool) {
	value, exists := ec.Metadata[key]

	return value, exists
}

// ContextKey is used for storing execution context in the standard context.
type ContextKey string

const (
	// ExecutionContextKey is the key for storing ExecutionContext in context.Context.
	ExecutionContextKey ContextKey = "execution_context"
)

// WithExecutionContext adds an ExecutionContext to the standard context.
func WithExecutionContext(ctx context.Context, execCtx *ExecutionContext) context.Context {
	return context.WithValue(ctx, ExecutionContextKey, execCtx)
}

// GetExecutionContext retrieves an ExecutionContext from the standard context.
func GetExecutionContext(ctx context.Context) (*ExecutionContext, bool) {
	execCtx, ok := ctx.Value(ExecutionContextKey).(*ExecutionContext)

	return execCtx, ok
}

// generateRequestID creates a simple request ID for tracing.
// In production, you might want to use a more sophisticated ID generator.
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// Config provides configuration options for middleware components.
type Config struct {
	// Enabled controls whether the middleware is active
	Enabled bool

	// Priority sets the execution order (lower numbers execute first)
	Priority int

	// Config holds middleware-specific configuration
	Config map[string]interface{}
}

// NewConfig creates a default middleware configuration.
func NewConfig() *Config {
	const defaultMiddlewarePriority = 100

	return &Config{
		Enabled:  true,
		Priority: defaultMiddlewarePriority, // Default priority
		Config:   make(map[string]interface{}),
	}
}

// WithPriority sets the middleware priority.
func (mc *Config) WithPriority(priority int) *Config {
	mc.Priority = priority

	return mc
}

// WithConfig sets a configuration value.
func (mc *Config) WithConfig(key string, value interface{}) *Config {
	mc.Config[key] = value

	return mc
}

// GetConfig retrieves a configuration value.
func (mc *Config) GetConfig(key string) (interface{}, bool) {
	value, exists := mc.Config[key]

	return value, exists
}

// GetConfigString retrieves a string configuration value.
func (mc *Config) GetConfigString(key string) string {
	if value, exists := mc.Config[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}

	return ""
}

// GetConfigInt retrieves an integer configuration value.
func (mc *Config) GetConfigInt(key string) int {
	if value, exists := mc.Config[key]; exists {
		if i, ok := value.(int); ok {
			return i
		}
	}

	return 0
}

// GetConfigBool retrieves a boolean configuration value.
func (mc *Config) GetConfigBool(key string) bool {
	if value, exists := mc.Config[key]; exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}

	return false
}

// BaseMiddleware provides common functionality for middleware implementations.
// Other middleware can embed this to inherit standard behavior.
type BaseMiddleware struct {
	name   string
	config *Config
	logger pkglogger.Logger
}

// NewBaseMiddleware creates a new base middleware with the given name and config.
func NewBaseMiddleware(name string, config *Config, logger pkglogger.Logger) *BaseMiddleware {
	if config == nil {
		config = NewConfig()
	}

	if logger == nil {
		logger = pkglogger.New("info")
	}

	return &BaseMiddleware{
		name:   name,
		config: config,
		logger: logger,
	}
}

// Name returns the middleware name.
func (bm *BaseMiddleware) Name() string {
	return bm.name
}

// Priority returns the middleware priority.
func (bm *BaseMiddleware) Priority() int {
	return bm.config.Priority
}

// Config returns the middleware configuration.
func (bm *BaseMiddleware) Config() *Config {
	return bm.config
}

// Logger returns the middleware logger.
func (bm *BaseMiddleware) Logger() pkglogger.Logger { //nolint:ireturn // Getter method should return interface
	return bm.logger
}

// IsEnabled returns whether this middleware is enabled.
func (bm *BaseMiddleware) IsEnabled() bool {
	return bm.config.Enabled
}
