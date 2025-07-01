package middleware

import (
	"context"
	"time"

	"github.com/chadit/CloudMCP/pkg/interfaces"
	pkglogger "github.com/chadit/CloudMCP/pkg/logger"
)

const (
	// Default priorities for logging middleware.
	defaultLoggingPriority           = 10
	defaultSecurityLoggingPriority   = 5
	defaultStructuredLoggingPriority = 15
)

// LoggingMiddleware provides detailed logging for tool execution.
// It logs the start, completion, and any errors that occur during tool execution.
type LoggingMiddleware struct {
	*BaseMiddleware
}

// NewLoggingMiddleware creates a new logging middleware.
func NewLoggingMiddleware(config *Config, logger pkglogger.Logger) *LoggingMiddleware {
	if config == nil {
		config = NewConfig().WithPriority(defaultLoggingPriority) // High priority for logging
	}

	return &LoggingMiddleware{
		BaseMiddleware: NewBaseMiddleware("logging", config, logger),
	}
}

// Execute implements the Middleware interface for logging tool execution.
func (lm *LoggingMiddleware) Execute(ctx context.Context, tool interfaces.Tool, params map[string]any, next ToolHandler) (any, error) {
	if !lm.IsEnabled() {
		return next(ctx, tool, params)
	}

	toolName := tool.Definition().Name
	startTime := time.Now()

	// Get or create execution context
	execCtx, hasCtx := GetExecutionContext(ctx)
	if !hasCtx {
		execCtx = NewExecutionContext(toolName, "unknown")
		ctx = WithExecutionContext(ctx, execCtx)
	}

	// Log tool execution start
	lm.logger.Info("Tool execution started",
		"tool", toolName,
		"request_id", execCtx.RequestID,
		"provider", execCtx.Provider,
		"params_count", len(params),
	)

	// Log parameters if debug level is enabled
	if lm.config.GetConfigBool("log_parameters") {
		lm.logger.Debug("Tool parameters", "tool", toolName, "params", params)
	}

	// Execute the next handler
	result, err := next(ctx, tool, params)

	duration := time.Since(startTime)

	if err != nil {
		// Log error
		lm.logger.Error("Tool execution failed",
			"tool", toolName,
			"request_id", execCtx.RequestID,
			"duration", duration,
			"error", err,
		)
	} else {
		// Log successful completion
		lm.logger.Info("Tool execution completed",
			"tool", toolName,
			"request_id", execCtx.RequestID,
			"duration", duration,
		)

		// Log result if debug level is enabled
		if lm.config.GetConfigBool("log_results") {
			lm.logger.Debug("Tool result", "tool", toolName, "result", result)
		}
	}

	return result, err
}

// SecurityLoggingMiddleware provides security-focused logging for audit trails.
// It logs security-relevant events like authentication, authorization, and sensitive operations.
type SecurityLoggingMiddleware struct {
	*BaseMiddleware
}

// NewSecurityLoggingMiddleware creates a new security logging middleware.
func NewSecurityLoggingMiddleware(config *Config, logger pkglogger.Logger) *SecurityLoggingMiddleware {
	if config == nil {
		config = NewConfig().WithPriority(defaultSecurityLoggingPriority) // Very high priority for security logging
	}

	return &SecurityLoggingMiddleware{
		BaseMiddleware: NewBaseMiddleware("security_logging", config, logger),
	}
}

// Execute implements the Middleware interface for security logging.
func (slm *SecurityLoggingMiddleware) Execute(ctx context.Context, tool interfaces.Tool, params map[string]any, next ToolHandler) (any, error) {
	if !slm.IsEnabled() {
		return next(ctx, tool, params)
	}

	toolName := tool.Definition().Name

	// Get execution context
	execCtx, hasCtx := GetExecutionContext(ctx)
	if !hasCtx {
		execCtx = NewExecutionContext(toolName, "unknown")
		ctx = WithExecutionContext(ctx, execCtx)
	}

	// Check if this is a security-sensitive operation
	if slm.isSecuritySensitive(toolName) {
		slm.logger.Info("Security-sensitive operation initiated",
			"tool", toolName,
			"request_id", execCtx.RequestID,
			"user_id", execCtx.UserID,
			"provider", execCtx.Provider,
			"timestamp", time.Now().UTC(),
		)
	}

	// Execute the operation
	result, err := next(ctx, tool, params)

	// Log security events
	if slm.isSecuritySensitive(toolName) {
		if err != nil {
			slm.logger.Warn("Security-sensitive operation failed",
				"tool", toolName,
				"request_id", execCtx.RequestID,
				"user_id", execCtx.UserID,
				"error", err,
			)
		} else {
			slm.logger.Info("Security-sensitive operation completed",
				"tool", toolName,
				"request_id", execCtx.RequestID,
				"user_id", execCtx.UserID,
			)
		}
	}

	return result, err
}

// isSecuritySensitive determines if a tool operation is security-sensitive.
func (slm *SecurityLoggingMiddleware) isSecuritySensitive(toolName string) bool {
	// Define security-sensitive operations
	sensitiveOperations := map[string]bool{
		"account_switch":        true,
		"firewall_create":       true,
		"firewall_delete":       true,
		"firewall_rule_create":  true,
		"firewall_rule_delete":  true,
		"instance_delete":       true,
		"instance_boot":         true,
		"instance_shutdown":     true,
		"instance_reboot":       true,
		"volume_delete":         true,
		"database_delete":       true,
		"lke_cluster_delete":    true,
		"domain_delete":         true,
		"nodebalancer_delete":   true,
		"object_storage_delete": true,
	}

	// Check configuration for additional sensitive operations
	if additionalOps, exists := slm.config.GetConfig("sensitive_operations"); exists {
		if ops, ok := additionalOps.([]string); ok {
			for _, op := range ops {
				sensitiveOperations[op] = true
			}
		}
	}

	return sensitiveOperations[toolName]
}

// StructuredLoggingMiddleware provides structured logging with consistent field formatting.
type StructuredLoggingMiddleware struct {
	*BaseMiddleware
}

// NewStructuredLoggingMiddleware creates a new structured logging middleware.
func NewStructuredLoggingMiddleware(config *Config, logger pkglogger.Logger) *StructuredLoggingMiddleware {
	if config == nil {
		config = NewConfig().WithPriority(defaultStructuredLoggingPriority) // After basic logging
	}

	return &StructuredLoggingMiddleware{
		BaseMiddleware: NewBaseMiddleware("structured_logging", config, logger),
	}
}

// Execute implements the Middleware interface for structured logging.
func (slm *StructuredLoggingMiddleware) Execute(ctx context.Context, tool interfaces.Tool, params map[string]any, next ToolHandler) (any, error) {
	if !slm.IsEnabled() {
		return next(ctx, tool, params)
	}

	// Get execution context
	execCtx, hasCtx := GetExecutionContext(ctx)
	if !hasCtx {
		execCtx = NewExecutionContext(tool.Definition().Name, "unknown")
		ctx = WithExecutionContext(ctx, execCtx)
	}

	// Create structured log entry
	logEntry := map[string]any{
		"event_type":   "tool_execution",
		"tool_name":    tool.Definition().Name,
		"request_id":   execCtx.RequestID,
		"provider":     execCtx.Provider,
		"user_id":      execCtx.UserID,
		"timestamp":    time.Now().UTC(),
		"params_count": len(params),
	}

	// Add tool metadata
	if description := tool.Definition().Description; description != "" {
		logEntry["tool_description"] = description
	}

	// Execute and measure
	startTime := time.Now()
	result, err := next(ctx, tool, params)
	duration := time.Since(startTime)

	// Complete log entry
	logEntry["duration_ms"] = duration.Milliseconds()
	logEntry["success"] = err == nil

	if err != nil {
		logEntry["error_message"] = err.Error()
		slm.logger.Error("Structured tool execution log", "entry", logEntry)
	} else {
		slm.logger.Info("Structured tool execution log", "entry", logEntry)
	}

	return result, err
}
