package middleware

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/chadit/CloudMCP/pkg/interfaces"
	pkglogger "github.com/chadit/CloudMCP/pkg/logger"
)

const (
	// Priority constants for middleware ordering.
	defaultRecoveryPriority        = 1
	defaultErrorEnrichmentPriority = 50
	defaultRetryPriority           = 40
	defaultCircuitBreakerPriority  = 45

	// Retry configuration constants.
	recoveryMaxRetries     = 3
	recoveryTimeoutSeconds = 30
	recoveryBackoffFactor  = 2.0

	// Circuit breaker configuration constants.
	recoveryFailureThreshold = 5
	recoveryRecoveryTimeout  = 60
	recoverySuccessThreshold = 3
	stackTraceBufferSize     = 4096
)

var (
	// ErrCircuitBreakerOpen indicates that the circuit breaker is open for a tool.
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
	// ErrToolExecutionPanic indicates that tool execution panicked.
	ErrToolExecutionPanic = errors.New("tool execution panic")
	// ErrToolExecutionFailed indicates that tool execution failed with enriched context.
	ErrToolExecutionFailed = errors.New("tool execution failed")
)

// RecoveryMiddleware provides panic recovery and error handling.
type RecoveryMiddleware struct {
	*BaseMiddleware
}

// NewRecoveryMiddleware creates a new recovery middleware.
func NewRecoveryMiddleware(config *Config, logger pkglogger.Logger) *RecoveryMiddleware {
	if config == nil {
		config = NewConfig().WithPriority(defaultRecoveryPriority) // Highest priority - should wrap everything
	}

	return &RecoveryMiddleware{
		BaseMiddleware: NewBaseMiddleware("recovery", config, logger),
	}
}

// Execute implements the Middleware interface for panic recovery.
func (rm *RecoveryMiddleware) Execute(ctx context.Context, tool interfaces.Tool, params map[string]any, next ToolHandler) (any, error) {
	if !rm.IsEnabled() {
		return next(ctx, tool, params)
	}

	// Execute the next handler with panic recovery
	var result any

	var err error

	func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				// Capture stack trace
				stackTrace := make([]byte, stackTraceBufferSize)
				stackSize := runtime.Stack(stackTrace, false)

				// Get execution context
				execCtx, hasCtx := GetExecutionContext(ctx)
				requestID := "unknown"

				if hasCtx {
					requestID = execCtx.RequestID
				}

				// Log the panic
				rm.logger.Error("Panic recovered during tool execution",
					"tool", tool.Definition().Name,
					"request_id", requestID,
					"panic", recovered,
					"stack_trace", string(stackTrace[:stackSize]),
				)

				// Convert panic to error
				err = fmt.Errorf("%w: %v", ErrToolExecutionPanic, recovered)
				result = nil
			}
		}()

		result, err = next(ctx, tool, params)
	}()

	return result, err
}

// ErrorEnrichmentMiddleware enriches errors with additional context.
type ErrorEnrichmentMiddleware struct {
	*BaseMiddleware
}

// NewErrorEnrichmentMiddleware creates a new error enrichment middleware.
func NewErrorEnrichmentMiddleware(config *Config, logger pkglogger.Logger) *ErrorEnrichmentMiddleware {
	if config == nil {
		config = NewConfig().WithPriority(defaultErrorEnrichmentPriority) // After execution, before response
	}

	return &ErrorEnrichmentMiddleware{
		BaseMiddleware: NewBaseMiddleware("error_enrichment", config, logger),
	}
}

// Execute implements the Middleware interface for error enrichment.
func (eem *ErrorEnrichmentMiddleware) Execute(ctx context.Context, tool interfaces.Tool, params map[string]any, next ToolHandler) (any, error) {
	if !eem.IsEnabled() {
		return next(ctx, tool, params)
	}

	// Execute the tool
	result, err := next(ctx, tool, params)
	// Enrich error if one occurred
	if err != nil {
		enrichedErr := eem.enrichError(ctx, tool, err)

		return result, enrichedErr
	}

	return result, nil
}

// enrichError adds additional context to errors.
func (eem *ErrorEnrichmentMiddleware) enrichError(ctx context.Context, tool interfaces.Tool, originalErr error) error {
	// Get execution context
	execCtx, hasCtx := GetExecutionContext(ctx)
	if !hasCtx {
		return originalErr
	}

	// Create enriched error message
	enrichedMsg := fmt.Sprintf("Tool execution failed: %v (tool: %s, provider: %s, request_id: %s)",
		originalErr,
		tool.Definition().Name,
		execCtx.Provider,
		execCtx.RequestID,
	)

	return fmt.Errorf("%w: %s", ErrToolExecutionFailed, enrichedMsg)
}

// RetryMiddleware provides automatic retry functionality for failed operations.
type RetryMiddleware struct {
	*BaseMiddleware
}

// NewRetryMiddleware creates a new retry middleware.
func NewRetryMiddleware(config *Config, logger pkglogger.Logger) *RetryMiddleware {
	if config == nil {
		config = NewConfig().WithPriority(defaultRetryPriority)
		// Default retry configuration
		config.WithConfig("max_retries", recoveryMaxRetries)
		config.WithConfig("base_delay", time.Second)
		config.WithConfig("max_delay", recoveryTimeoutSeconds*time.Second)
		config.WithConfig("backoff_factor", recoveryBackoffFactor)
	}

	return &RetryMiddleware{
		BaseMiddleware: NewBaseMiddleware("retry", config, logger),
	}
}

// Execute implements the Middleware interface for retry logic.
func (rm *RetryMiddleware) Execute(ctx context.Context, tool interfaces.Tool, params map[string]any, next ToolHandler) (any, error) {
	if !rm.IsEnabled() {
		return next(ctx, tool, params)
	}

	maxRetries := rm.config.GetConfigInt("max_retries")
	if maxRetries <= 0 {
		return next(ctx, tool, params)
	}

	baseDelay := time.Second

	if delay, exists := rm.config.GetConfig("base_delay"); exists {
		if d, ok := delay.(time.Duration); ok {
			baseDelay = d
		}
	}

	maxDelay := recoveryTimeoutSeconds * time.Second

	if delay, exists := rm.config.GetConfig("max_delay"); exists {
		if d, ok := delay.(time.Duration); ok {
			maxDelay = d
		}
	}

	backoffFactor := 2.0

	if factor, exists := rm.config.GetConfig("backoff_factor"); exists {
		if f, ok := factor.(float64); ok {
			backoffFactor = f
		}
	}

	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Execute the tool
		result, err := next(ctx, tool, params)

		// Success - return immediately
		if err == nil {
			if attempt > 0 {
				rm.logger.Info("Tool execution succeeded after retry",
					"tool", tool.Definition().Name,
					"attempt", attempt+1,
					"total_attempts", attempt+1,
				)
			}

			return result, nil
		}

		lastErr = err

		// Check if error is retryable
		if !rm.isRetryableError(err) {
			rm.logger.Debug("Error is not retryable, giving up",
				"tool", tool.Definition().Name,
				"error", err,
			)

			break
		}

		// Don't delay after the last attempt
		if attempt < maxRetries {
			// Calculate delay with exponential backoff
			delay := rm.calculateDelay(attempt, baseDelay, maxDelay, backoffFactor)

			rm.logger.Warn("Tool execution failed, retrying",
				"tool", tool.Definition().Name,
				"attempt", attempt+1,
				"max_attempts", maxRetries+1,
				"delay", delay,
				"error", err,
			)

			// Wait before retry
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during retry delay: %w", ctx.Err())
			case <-time.After(delay):
			}
		}
	}

	// All retries exhausted
	rm.logger.Error("Tool execution failed after all retries",
		"tool", tool.Definition().Name,
		"attempts", maxRetries+1,
		"final_error", lastErr,
	)

	return nil, fmt.Errorf("tool execution failed after %d attempts: %w", maxRetries+1, lastErr)
}

// isRetryableError determines if an error should trigger a retry.
func (rm *RetryMiddleware) isRetryableError(err error) bool {
	errorStr := err.Error()

	// Errors that should NOT be retried
	nonRetryableErrors := []string{
		"authentication",
		"unauthorized",
		"forbidden",
		"not found",
		"bad request",
		"validation",
		"invalid",
		"rate limit", // Rate limits should be handled by rate limiting middleware
	}

	for _, nonRetryable := range nonRetryableErrors {
		if containsAny(errorStr, []string{nonRetryable}) {
			return false
		}
	}

	// Errors that SHOULD be retried
	retryableErrors := []string{
		"timeout",
		"deadline",
		"network",
		"connection",
		"server error",
		"503", // Service unavailable
		"502", // Bad gateway
		"500", // Internal server error
	}

	for _, retryable := range retryableErrors {
		if containsAny(errorStr, []string{retryable}) {
			return true
		}
	}

	// Default: don't retry unless specifically identified as retryable
	return false
}

// calculateDelay calculates the delay for the next retry using exponential backoff.
func (rm *RetryMiddleware) calculateDelay(attempt int, baseDelay, maxDelay time.Duration, backoffFactor float64) time.Duration {
	// Calculate exponential backoff
	delay := time.Duration(float64(baseDelay) * pow(backoffFactor, float64(attempt)))

	// Cap at maximum delay
	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}

// pow calculates base^exp for floating point numbers (simple implementation).
func pow(base, exp float64) float64 {
	if exp == 0 {
		return 1
	}

	if exp == 1 {
		return base
	}

	result := 1.0
	for range int(exp) {
		result *= base
	}

	return result
}

// CircuitBreakerMiddleware implements the circuit breaker pattern.
type CircuitBreakerMiddleware struct {
	*BaseMiddleware
	state            CircuitState
	failureCount     int
	lastFailureTime  time.Time
	successCount     int
	failureThreshold int
	recoveryTimeout  time.Duration
	successThreshold int
}

// CircuitState represents the state of a circuit breaker.
type CircuitState int

const (
	// CircuitClosed means the circuit is closed and requests are allowed.
	CircuitClosed CircuitState = iota
	// CircuitOpen means the circuit is open and requests are rejected.
	CircuitOpen
	// CircuitHalfOpen means the circuit is testing if it should close.
	CircuitHalfOpen
)

// NewCircuitBreakerMiddleware creates a new circuit breaker middleware.
func NewCircuitBreakerMiddleware(config *Config, logger pkglogger.Logger) *CircuitBreakerMiddleware {
	if config == nil {
		config = NewConfig().WithPriority(defaultCircuitBreakerPriority)
		// Default circuit breaker configuration
		config.WithConfig("failure_threshold", recoveryFailureThreshold)
		config.WithConfig("recovery_timeout", recoveryRecoveryTimeout*time.Second)
		config.WithConfig("success_threshold", recoverySuccessThreshold)
	}

	return &CircuitBreakerMiddleware{
		BaseMiddleware:   NewBaseMiddleware("circuit_breaker", config, logger),
		state:            CircuitClosed,
		failureThreshold: config.GetConfigInt("failure_threshold"),
		recoveryTimeout:  recoveryRecoveryTimeout * time.Second,
		successThreshold: config.GetConfigInt("success_threshold"),
	}
}

// Execute implements the Middleware interface for circuit breaker functionality.
func (cbm *CircuitBreakerMiddleware) Execute(ctx context.Context, tool interfaces.Tool, params map[string]any, next ToolHandler) (any, error) {
	if !cbm.IsEnabled() {
		return next(ctx, tool, params)
	}

	// Check circuit state
	if !cbm.allowRequest() {
		cbm.logger.Warn("Circuit breaker is open, rejecting request",
			"tool", tool.Definition().Name,
			"failure_count", cbm.failureCount,
		)

		return nil, fmt.Errorf("%w for tool %s", ErrCircuitBreakerOpen, tool.Definition().Name)
	}

	// Execute the tool
	result, err := next(ctx, tool, params)

	// Update circuit state based on result
	if err != nil {
		cbm.onFailure()
	} else {
		cbm.onSuccess()
	}

	return result, err
}

// allowRequest determines if a request should be allowed based on circuit state.
func (cbm *CircuitBreakerMiddleware) allowRequest() bool {
	switch cbm.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		// Check if recovery timeout has passed
		if time.Since(cbm.lastFailureTime) >= cbm.recoveryTimeout {
			cbm.state = CircuitHalfOpen
			cbm.successCount = 0

			return true
		}

		return false
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

// onFailure updates the circuit state when a failure occurs.
func (cbm *CircuitBreakerMiddleware) onFailure() {
	cbm.failureCount++
	cbm.lastFailureTime = time.Now()
	cbm.successCount = 0

	if cbm.state == CircuitClosed && cbm.failureCount >= cbm.failureThreshold {
		cbm.state = CircuitOpen
		cbm.logger.Warn("Circuit breaker opened due to failures",
			"failure_count", cbm.failureCount,
			"threshold", cbm.failureThreshold,
		)
	} else if cbm.state == CircuitHalfOpen {
		cbm.state = CircuitOpen
		cbm.logger.Warn("Circuit breaker reopened due to failure in half-open state")
	}
}

// onSuccess updates the circuit state when a success occurs.
func (cbm *CircuitBreakerMiddleware) onSuccess() {
	cbm.successCount++

	if cbm.state == CircuitHalfOpen && cbm.successCount >= cbm.successThreshold {
		cbm.state = CircuitClosed
		cbm.failureCount = 0
		cbm.logger.Info("Circuit breaker closed after successful recovery",
			"success_count", cbm.successCount,
		)
	}
}
