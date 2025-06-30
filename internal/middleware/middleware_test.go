package middleware_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/middleware"
	"github.com/chadit/CloudMCP/pkg/interfaces"
	pkglogger "github.com/chadit/CloudMCP/pkg/logger"
)

// Constants for repeated strings to fix goconst linting issues.
const (
	testToolName        = "test_tool"
	testToolDescription = "Test tool"
	middleware1Name     = "middleware1"
	middleware2Name     = "middleware2"
	middleware3Name     = "middleware3"
	mockResult          = "mock result"
	handlerResult       = "handler result"
	successResult       = "success"
	testError           = "test error"
	networkError        = "network error"
	authenticationError = "authentication failed"
	testPanic           = "test panic"
	logLevel            = "info"
	key1                = "key1"
	key2                = "key2"
	key3                = "key3"
	value1              = "value1"
	testValue           = "test_value"
	testKey             = "test_key"
	testProvider        = "test_provider"
	testMiddleware      = "test_middleware"
	accountSwitch       = "account_switch"
	switchAccount       = "Switch account"
	switched            = "switched"
	param1              = "param1"
	logParameters       = "log_parameters"
	logResults          = "log_results"
	maxRetries          = "max_retries"
	baseDelay           = "base_delay"
	failureThreshold    = "failure_threshold"
	recoveryTimeout     = "recovery_timeout"
	successThreshold    = "success_threshold"
	testToolCB          = "test_tool_cb"
	nonExistent         = "non_existent"
)

// Static errors to fix err113 linting issues.
var (
	errTest           = errors.New(testError)
	errNetwork        = errors.New(networkError)
	errAuthentication = errors.New(authenticationError)
)

// mockTool implements the interfaces.Tool interface for testing.
type mockTool struct {
	name        string
	description string
}

func (mt *mockTool) Definition() mcp.Tool {
	return mcp.Tool{
		Name:        mt.name,
		Description: mt.description,
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		},
	}
}

func (mt *mockTool) Execute(_ context.Context, _ map[string]any) (any, error) {
	return mockResult, nil
}

func (mt *mockTool) Validate(_ map[string]any) error {
	return nil
}

// mockMiddleware implements the middleware.Middleware interface for testing.
type mockMiddleware struct {
	name     string
	priority int
	executed bool
	beforeFn func()
	afterFn  func()
}

func (mm *mockMiddleware) Name() string {
	return mm.name
}

func (mm *mockMiddleware) Priority() int {
	return mm.priority
}

func (mm *mockMiddleware) Execute(ctx context.Context, tool interfaces.Tool, params map[string]any, next middleware.ToolHandler) (any, error) {
	mm.executed = true
	if mm.beforeFn != nil {
		mm.beforeFn()
	}

	result, err := next(ctx, tool, params)

	if mm.afterFn != nil {
		mm.afterFn()
	}

	return result, err
}

func TestChain_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		middlewares   []middleware.Middleware
		expectedOrder []string
	}{
		{
			name: "Single middleware",
			middlewares: []middleware.Middleware{
				&mockMiddleware{name: middleware1Name, priority: 10},
			},
			expectedOrder: []string{middleware1Name},
		},
		{
			name: "Multiple middlewares in order",
			middlewares: []middleware.Middleware{
				&mockMiddleware{name: middleware1Name, priority: 10},
				&mockMiddleware{name: middleware2Name, priority: 20},
				&mockMiddleware{name: middleware3Name, priority: 30},
			},
			expectedOrder: []string{middleware1Name, middleware2Name, middleware3Name},
		},
		{
			name: "Multiple middlewares out of order",
			middlewares: []middleware.Middleware{
				&mockMiddleware{name: middleware3Name, priority: 30},
				&mockMiddleware{name: middleware1Name, priority: 10},
				&mockMiddleware{name: middleware2Name, priority: 20},
			},
			expectedOrder: []string{middleware1Name, middleware2Name, middleware3Name},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			chain := middleware.NewChain(pkglogger.New(logLevel))

			var executionOrder []string

			// Set up middlewares to track execution order
			for _, middlewareItem := range testCase.middlewares {
				if mock, ok := middlewareItem.(*mockMiddleware); ok {
					mock.beforeFn = func() {
						executionOrder = append(executionOrder, mock.name)
					}
				}

				chain.Add(middlewareItem)
			}

			tool := &mockTool{name: testToolName, description: testToolDescription}
			params := map[string]any{}

			handler := func(_ context.Context, _ interfaces.Tool, _ map[string]any) (any, error) {
				return handlerResult, nil
			}

			result, err := chain.Execute(t.Context(), tool, params, handler)

			require.NoError(t, err, "chain execution should not fail")
			assert.Equal(t, handlerResult, result)
			assert.Equal(t, testCase.expectedOrder, executionOrder)

			// Verify all middlewares were executed
			for _, middlewareItem := range testCase.middlewares {
				if mock, ok := middlewareItem.(*mockMiddleware); ok {
					assert.True(t, mock.executed, "Middleware %s should have been executed", mock.name)
				}
			}
		})
	}
}

func TestChain_ManipulationMethods(t *testing.T) {
	t.Parallel()

	chain := middleware.NewChain(pkglogger.New(logLevel))

	middleware1 := &mockMiddleware{name: middleware1Name, priority: 10}
	middleware2 := &mockMiddleware{name: middleware2Name, priority: 20}

	// Test Add
	chain.Add(middleware1)
	assert.Equal(t, 1, chain.Count())
	assert.True(t, chain.Has(middleware1Name))
	assert.False(t, chain.Has(middleware2Name))

	chain.Add(middleware2)
	assert.Equal(t, 2, chain.Count())
	assert.True(t, chain.Has(middleware2Name))

	// Test List (should be sorted by priority)
	middlewares := chain.List()
	assert.Len(t, middlewares, 2)
	assert.Equal(t, middleware1Name, middlewares[0].Name())
	assert.Equal(t, middleware2Name, middlewares[1].Name())

	// Test Remove
	removed := chain.Remove(middleware1Name)
	assert.True(t, removed)
	assert.Equal(t, 1, chain.Count())
	assert.False(t, chain.Has(middleware1Name))

	// Test Remove non-existent
	removed = chain.Remove(nonExistent)
	assert.False(t, removed)
	assert.Equal(t, 1, chain.Count())

	// Test Clear
	chain.Clear()
	assert.Equal(t, 0, chain.Count())
}

func TestExecutionContext(t *testing.T) {
	t.Parallel()

	execCtx := middleware.NewExecutionContext(testToolName, testProvider)

	// Test basic properties
	assert.Equal(t, testToolName, execCtx.ToolName)
	assert.Equal(t, testProvider, execCtx.Provider)
	assert.NotEmpty(t, execCtx.RequestID)
	assert.True(t, execCtx.StartTime.Before(time.Now()) || execCtx.StartTime.Equal(time.Now()))

	// Test Duration
	time.Sleep(10 * time.Millisecond)

	duration := execCtx.Duration()
	assert.GreaterOrEqual(t, duration, 10*time.Millisecond)

	// Test metadata
	execCtx.WithMetadata(key1, value1)
	execCtx.WithMetadata(key2, 42)

	value, exists := execCtx.GetMetadata(key1)
	assert.True(t, exists)
	assert.Equal(t, value1, value)

	value, exists = execCtx.GetMetadata(key2)
	assert.True(t, exists)
	assert.Equal(t, 42, value)

	_, exists = execCtx.GetMetadata(nonExistent)
	assert.False(t, exists)
}

func TestContextManagement(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	execCtx := middleware.NewExecutionContext(testToolName, testProvider)

	// Test WithExecutionContext
	ctxWithExec := middleware.WithExecutionContext(ctx, execCtx)

	// Test GetExecutionContext
	retrievedCtx, exists := middleware.GetExecutionContext(ctxWithExec)
	assert.True(t, exists)
	assert.Equal(t, execCtx, retrievedCtx)

	// Test GetExecutionContext with context that doesn't have execution context
	_, exists = middleware.GetExecutionContext(ctx)
	assert.False(t, exists)
}

func TestConfig(t *testing.T) {
	t.Parallel()

	config := middleware.NewConfig()

	// Test defaults
	assert.True(t, config.Enabled)
	assert.Equal(t, 100, config.Priority)
	assert.NotNil(t, config.Config)

	// Test WithPriority
	config.WithPriority(50)
	assert.Equal(t, 50, config.Priority)

	// Test WithConfig
	config.WithConfig(key1, value1)
	config.WithConfig(key2, 42)
	config.WithConfig(key3, true)

	// Test GetConfig
	value, exists := config.GetConfig(key1)
	assert.True(t, exists)
	assert.Equal(t, value1, value)

	// Test typed getters
	assert.Equal(t, value1, config.GetConfigString(key1))
	assert.Equal(t, "", config.GetConfigString(nonExistent))

	assert.Equal(t, 42, config.GetConfigInt(key2))
	assert.Equal(t, 0, config.GetConfigInt(nonExistent))

	assert.True(t, config.GetConfigBool(key3))
	assert.False(t, config.GetConfigBool(nonExistent))
}

func TestBaseMiddleware(t *testing.T) {
	t.Parallel()

	config := middleware.NewConfig().WithPriority(50)
	config.WithConfig(testKey, testValue)

	logger := pkglogger.New(logLevel)
	baseMiddleware := middleware.NewBaseMiddleware(testMiddleware, config, logger)

	assert.Equal(t, testMiddleware, baseMiddleware.Name())
	assert.Equal(t, 50, baseMiddleware.Priority())
	assert.Equal(t, config, baseMiddleware.Config())
	assert.Equal(t, logger, baseMiddleware.Logger())
	assert.True(t, baseMiddleware.IsEnabled())

	// Test with disabled config
	config.Enabled = false

	assert.False(t, baseMiddleware.IsEnabled())
}

func TestLoggingMiddleware(t *testing.T) {
	t.Parallel()

	config := middleware.NewConfig()
	config.WithConfig(logParameters, true)
	config.WithConfig(logResults, true)

	logger := pkglogger.New(logLevel)
	middleware := middleware.NewLoggingMiddleware(config, logger)

	assert.Equal(t, "logging", middleware.Name())
	assert.Equal(t, 100, middleware.Priority())

	tool := &mockTool{name: testToolName, description: testToolDescription}
	params := map[string]any{param1: value1}

	handler := func(_ context.Context, _ interfaces.Tool, _ map[string]any) (any, error) {
		return handlerResult, nil
	}

	result, err := middleware.Execute(t.Context(), tool, params, handler)

	require.NoError(t, err, "logging middleware should not fail")
	assert.Equal(t, handlerResult, result)
}

func TestLoggingMiddleware_WithError(t *testing.T) {
	t.Parallel()

	middleware := middleware.NewLoggingMiddleware(nil, pkglogger.New(logLevel))

	tool := &mockTool{name: testToolName, description: testToolDescription}
	params := map[string]any{}

	handler := func(_ context.Context, _ interfaces.Tool, _ map[string]any) (any, error) {
		return nil, errTest
	}

	result, err := middleware.Execute(t.Context(), tool, params, handler)

	assert.Equal(t, errTest, err)
	assert.Nil(t, result)
}

func TestSecurityLoggingMiddleware(t *testing.T) {
	t.Parallel()

	middleware := middleware.NewSecurityLoggingMiddleware(nil, pkglogger.New(logLevel))

	assert.Equal(t, "security_logging", middleware.Name())
	assert.Equal(t, 5, middleware.Priority())

	// Test with security-sensitive tool
	tool := &mockTool{name: accountSwitch, description: switchAccount}
	params := map[string]any{}

	handler := func(_ context.Context, _ interfaces.Tool, _ map[string]any) (any, error) {
		return switched, nil
	}

	result, err := middleware.Execute(t.Context(), tool, params, handler)

	require.NoError(t, err, "security logging middleware should not fail")
	assert.Equal(t, switched, result)
}

func TestMetricsMiddleware(t *testing.T) {
	t.Parallel()

	collector := &middleware.NoOpMetricsCollector{}
	middleware := middleware.NewMetricsMiddleware(nil, pkglogger.New(logLevel), collector)

	assert.Equal(t, "metrics", middleware.Name())
	assert.Equal(t, 20, middleware.Priority())

	tool := &mockTool{name: testToolName, description: testToolDescription}
	params := map[string]any{param1: value1}

	handler := func(_ context.Context, _ interfaces.Tool, _ map[string]any) (any, error) {
		time.Sleep(10 * time.Millisecond) // Simulate some work

		return handlerResult, nil
	}

	result, err := middleware.Execute(t.Context(), tool, params, handler)

	require.NoError(t, err, "metrics middleware should not fail")
	assert.Equal(t, handlerResult, result)
}

func TestRecoveryMiddleware(t *testing.T) {
	t.Parallel()

	middleware := middleware.NewRecoveryMiddleware(nil, pkglogger.New(logLevel))

	assert.Equal(t, "recovery", middleware.Name())
	assert.Equal(t, 1, middleware.Priority())

	tool := &mockTool{name: testToolName, description: testToolDescription}
	params := map[string]any{}

	// Test panic recovery
	handler := func(_ context.Context, _ interfaces.Tool, _ map[string]any) (any, error) {
		panic(testPanic)
	}

	result, err := middleware.Execute(t.Context(), tool, params, handler)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "tool execution panic: "+testPanic)
	assert.Nil(t, result)
}

func TestRetryMiddleware(t *testing.T) {
	t.Parallel()

	config := middleware.NewConfig()
	config.WithConfig(maxRetries, 2)
	config.WithConfig(baseDelay, 10*time.Millisecond)

	middleware := middleware.NewRetryMiddleware(config, pkglogger.New(logLevel))

	assert.Equal(t, "retry", middleware.Name())
	assert.Equal(t, 100, middleware.Priority())

	tool := &mockTool{name: testToolName, description: testToolDescription}
	params := map[string]any{}

	// Test successful retry
	attempts := 0
	handler := func(_ context.Context, _ interfaces.Tool, _ map[string]any) (any, error) {
		attempts++
		if attempts < 2 {
			return nil, errNetwork // Retryable error
		}

		return successResult, nil
	}

	start := time.Now()
	result, err := middleware.Execute(t.Context(), tool, params, handler)
	duration := time.Since(start)

	require.NoError(t, err, "retry middleware should succeed after retry")
	assert.Equal(t, successResult, result)
	assert.Equal(t, 2, attempts)
	assert.GreaterOrEqual(t, duration, 10*time.Millisecond) // Should have delayed for retry
}

func TestRetryMiddleware_NonRetryableError(t *testing.T) {
	t.Parallel()

	config := middleware.NewConfig()
	config.WithConfig(maxRetries, 2)

	middleware := middleware.NewRetryMiddleware(config, pkglogger.New(logLevel))

	tool := &mockTool{name: testToolName, description: testToolDescription}
	params := map[string]any{}

	// Test non-retryable error
	attempts := 0
	handler := func(_ context.Context, _ interfaces.Tool, _ map[string]any) (any, error) {
		attempts++

		return nil, errAuthentication // Non-retryable error
	}

	result, err := middleware.Execute(t.Context(), tool, params, handler)

	require.Error(t, err)
	assert.Contains(t, err.Error(), authenticationError)
	assert.Nil(t, result)
	assert.Equal(t, 1, attempts) // Should not have retried
}

func TestRateLimitMiddleware(t *testing.T) {
	t.Parallel()

	limiter := middleware.NewTokenBucket(2, time.Second, 2) // 2 requests per second
	middleware := middleware.NewRateLimitMiddleware(nil, pkglogger.New(logLevel), limiter)

	assert.Equal(t, "rate_limit", middleware.Name())
	assert.Equal(t, 30, middleware.Priority())

	tool := &mockTool{name: testToolName, description: testToolDescription}
	params := map[string]any{}

	handler := func(_ context.Context, _ interfaces.Tool, _ map[string]any) (any, error) {
		return successResult, nil
	}

	// First two requests should succeed
	for range 2 {
		result, err := middleware.Execute(t.Context(), tool, params, handler)
		require.NoError(t, err, "rate limit should allow initial requests")
		assert.Equal(t, successResult, result)
	}

	// Third request should be rate limited
	result, err := middleware.Execute(t.Context(), tool, params, handler)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit exceeded")
	assert.Nil(t, result)
}

func TestCircuitBreakerMiddleware(t *testing.T) {
	t.Parallel()

	config := middleware.NewConfig()
	config.WithConfig(failureThreshold, 2)
	config.WithConfig(recoveryTimeout, 100*time.Millisecond)
	config.WithConfig(successThreshold, 1)

	middleware := middleware.NewCircuitBreakerMiddleware(config, pkglogger.New(logLevel))

	assert.Equal(t, "circuit_breaker", middleware.Name())
	assert.Equal(t, 100, middleware.Priority())

	// Test basic functionality - for now just test that it executes without panicking
	tool := &mockTool{name: testToolCB, description: testToolDescription}
	params := map[string]any{}

	// Handler that succeeds
	successHandler := func(_ context.Context, _ interfaces.Tool, _ map[string]any) (any, error) {
		return successResult, nil
	}

	result, err := middleware.Execute(t.Context(), tool, params, successHandler)
	require.NoError(t, err, "circuit breaker middleware should not fail")
	assert.Equal(t, successResult, result)
}
