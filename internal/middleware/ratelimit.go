package middleware

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/chadit/CloudMCP/pkg/interfaces"
	pkglogger "github.com/chadit/CloudMCP/pkg/logger"
)

const (
	// Default priorities for rate limiting middleware.
	defaultRateLimitPriority         = 30
	defaultAdaptiveRateLimitPriority = 35

	// Default rate limiting values.
	defaultRateLimit    = 100
	adaptiveRateLimit   = 50
	adaptiveBurstSize   = 10
	loadThresholdMedium = 0.3
	loadFactorHigh      = 0.8
	loadFactorMedium    = 0.5
	retryDelaySeconds   = 30
	gracePeriodSeconds  = 60
)

var (
	// ErrRateLimitExceeded indicates that the rate limit has been exceeded.
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	// ErrSystemLoadHigh indicates that the rate limit is enforced due to high system load.
	ErrSystemLoadHigh = errors.New("rate limit exceeded due to high system load")
)

// RateLimiter defines the interface for rate limiting implementations.
type RateLimiter interface {
	// Allow checks if a request should be allowed based on the key and current rate.
	Allow(key string) bool

	// Reserve reserves a slot for the given key, returning the time to wait.
	Reserve(key string) time.Duration

	// Reset clears the rate limit state for the given key.
	Reset(key string)

	// GetLimit returns the current rate limit for the key.
	GetLimit(key string) (requests int, window time.Duration)
}

// TokenBucket implements a token bucket rate limiter.
type TokenBucket struct {
	mu       sync.RWMutex
	buckets  map[string]*bucket
	rate     int           // tokens per window
	window   time.Duration // time window
	capacity int           // maximum tokens
}

// bucket represents a single token bucket for rate limiting.
type bucket struct {
	tokens   int
	lastFill time.Time
}

// NewTokenBucket creates a new token bucket rate limiter.
func NewTokenBucket(rate int, window time.Duration, capacity int) *TokenBucket {
	if capacity <= 0 {
		capacity = rate
	}

	return &TokenBucket{
		buckets:  make(map[string]*bucket),
		rate:     rate,
		window:   window,
		capacity: capacity,
	}
}

// Allow checks if a request should be allowed.
func (tb *TokenBucket) Allow(key string) bool {
	return tb.Reserve(key) == 0
}

// Reserve reserves a token and returns the time to wait.
func (tb *TokenBucket) Reserve(key string) time.Duration {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	bucketData, exists := tb.buckets[key]

	if !exists {
		bucketData = &bucket{
			tokens:   tb.capacity,
			lastFill: now,
		}
		tb.buckets[key] = bucketData
	}

	// Fill tokens based on elapsed time
	elapsed := now.Sub(bucketData.lastFill)
	tokensToAdd := int(elapsed.Nanoseconds() * int64(tb.rate) / tb.window.Nanoseconds())

	if tokensToAdd > 0 {
		bucketData.tokens += tokensToAdd
		if bucketData.tokens > tb.capacity {
			bucketData.tokens = tb.capacity
		}

		bucketData.lastFill = now
	}

	// Try to consume a token
	if bucketData.tokens > 0 {
		bucketData.tokens--

		return 0
	}

	// Calculate wait time for next token
	timeToNextToken := tb.window / time.Duration(tb.rate)

	return timeToNextToken
}

// Reset clears the rate limit state for a key.
func (tb *TokenBucket) Reset(key string) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	delete(tb.buckets, key)
}

// GetLimit returns the rate limit configuration.
func (tb *TokenBucket) GetLimit(_ string) (int, time.Duration) {
	return tb.rate, tb.window
}

// SlidingWindow implements a sliding window rate limiter.
type SlidingWindow struct {
	mu      sync.RWMutex
	windows map[string]*windowData
	limit   int
	window  time.Duration
}

// windowData tracks requests in a sliding window.
type windowData struct {
	requests []time.Time
}

// NewSlidingWindow creates a new sliding window rate limiter.
func NewSlidingWindow(limit int, window time.Duration) *SlidingWindow {
	return &SlidingWindow{
		windows: make(map[string]*windowData),
		limit:   limit,
		window:  window,
	}
}

// Allow checks if a request should be allowed.
func (sw *SlidingWindow) Allow(key string) bool {
	return sw.Reserve(key) == 0
}

// Reserve reserves a slot and returns wait time.
func (sw *SlidingWindow) Reserve(key string) time.Duration {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-sw.window)

	winData, exists := sw.windows[key]

	if !exists {
		winData = &windowData{
			requests: make([]time.Time, 0),
		}
		sw.windows[key] = winData
	}

	// Remove old requests outside the window
	validRequests := make([]time.Time, 0, len(winData.requests))

	for _, reqTime := range winData.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}

	winData.requests = validRequests

	// Check if we can add a new request
	if len(winData.requests) < sw.limit {
		winData.requests = append(winData.requests, now)

		return 0
	}

	// Calculate wait time until oldest request expires
	oldest := winData.requests[0]
	waitUntil := oldest.Add(sw.window)

	if waitUntil.After(now) {
		return waitUntil.Sub(now)
	}

	return 0
}

// Reset clears the rate limit state for a key.
func (sw *SlidingWindow) Reset(key string) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	delete(sw.windows, key)
}

// GetLimit returns the rate limit configuration.
func (sw *SlidingWindow) GetLimit(_ string) (int, time.Duration) {
	return sw.limit, sw.window
}

// RateLimitMiddleware provides rate limiting for tool execution.
type RateLimitMiddleware struct {
	*BaseMiddleware
	limiter RateLimiter
}

// NewRateLimitMiddleware creates a new rate limiting middleware.
func NewRateLimitMiddleware(config *Config, logger pkglogger.Logger, limiter RateLimiter) *RateLimitMiddleware {
	if config == nil {
		config = NewConfig().WithPriority(defaultRateLimitPriority) // Before expensive operations
	}

	if limiter == nil {
		// Default: 100 requests per minute per tool
		limiter = NewTokenBucket(defaultRateLimit, time.Minute, defaultRateLimit)
	}

	return &RateLimitMiddleware{
		BaseMiddleware: NewBaseMiddleware("rate_limit", config, logger),
		limiter:        limiter,
	}
}

// Execute implements the Middleware interface for rate limiting.
func (rlm *RateLimitMiddleware) Execute(ctx context.Context, tool interfaces.Tool, params map[string]any, next ToolHandler) (any, error) {
	if !rlm.IsEnabled() {
		return next(ctx, tool, params)
	}

	// Create rate limit key
	key := rlm.createRateLimitKey(ctx, tool)

	// Check rate limit
	waitTime := rlm.limiter.Reserve(key)
	if waitTime > 0 {
		rlm.logger.Warn("Rate limit exceeded",
			"tool", tool.Definition().Name,
			"wait_time", waitTime,
			"key", key,
		)

		// Return rate limit error
		return nil, fmt.Errorf("%w for tool %q, please wait %v", ErrRateLimitExceeded, tool.Definition().Name, waitTime)
	}

	// Execute the tool
	return next(ctx, tool, params)
}

// createRateLimitKey creates a unique key for rate limiting.
func (rlm *RateLimitMiddleware) createRateLimitKey(ctx context.Context, tool interfaces.Tool) string {
	toolName := tool.Definition().Name

	// Get execution context for additional key components
	execCtx, hasCtx := GetExecutionContext(ctx)
	if !hasCtx {
		return toolName
	}

	// Create key based on configuration
	keyStrategy := rlm.config.GetConfigString("key_strategy")
	switch keyStrategy {
	case "per_user":
		if execCtx.UserID != "" {
			return fmt.Sprintf("%s:%s", execCtx.UserID, toolName)
		}

		return toolName
	case "per_provider":
		return fmt.Sprintf("%s:%s", execCtx.Provider, toolName)
	case "per_user_provider":
		if execCtx.UserID != "" {
			return fmt.Sprintf("%s:%s:%s", execCtx.UserID, execCtx.Provider, toolName)
		}

		return fmt.Sprintf("%s:%s", execCtx.Provider, toolName)
	case "global":
		return "global"
	default:
		// Default: per tool
		return toolName
	}
}

// AdaptiveRateLimitMiddleware provides adaptive rate limiting based on system load.
type AdaptiveRateLimitMiddleware struct {
	*BaseMiddleware
	baseLimiter   RateLimiter
	loadThreshold float64
	reductionRate float64
}

// NewAdaptiveRateLimitMiddleware creates a new adaptive rate limiting middleware.
func NewAdaptiveRateLimitMiddleware(config *Config, logger pkglogger.Logger, baseLimiter RateLimiter) *AdaptiveRateLimitMiddleware {
	if config == nil {
		config = NewConfig().WithPriority(defaultAdaptiveRateLimitPriority) // After basic rate limiting
	}

	if baseLimiter == nil {
		baseLimiter = NewTokenBucket(adaptiveRateLimit, time.Minute, adaptiveRateLimit)
	}

	return &AdaptiveRateLimitMiddleware{
		BaseMiddleware: NewBaseMiddleware("adaptive_rate_limit", config, logger),
		baseLimiter:    baseLimiter,
		loadThreshold:  loadFactorHigh,   // 80% system load threshold
		reductionRate:  loadFactorMedium, // Reduce rate limit by 50% when under load
	}
}

// Execute implements the Middleware interface for adaptive rate limiting.
func (arlm *AdaptiveRateLimitMiddleware) Execute(ctx context.Context, tool interfaces.Tool, params map[string]any, next ToolHandler) (any, error) {
	if !arlm.IsEnabled() {
		return next(ctx, tool, params)
	}

	// Check system load (simplified implementation)
	systemLoad := arlm.getSystemLoad()

	// Create rate limit key
	key := arlm.createAdaptiveKey(ctx, tool, systemLoad)

	// Apply adaptive rate limiting
	if systemLoad > arlm.loadThreshold {
		// System under load - apply stricter rate limiting
		waitTime := arlm.baseLimiter.Reserve(key)
		if waitTime > 0 {
			// Add additional delay based on load
			additionalDelay := time.Duration(float64(waitTime) * (systemLoad - arlm.loadThreshold))
			totalWait := waitTime + additionalDelay

			arlm.logger.Warn("Adaptive rate limit applied due to system load",
				"tool", tool.Definition().Name,
				"system_load", systemLoad,
				"wait_time", totalWait,
			)

			return nil, fmt.Errorf("%w (%.2f), please wait %v", ErrSystemLoadHigh, systemLoad, totalWait)
		}
	} else {
		// Normal load - use base rate limiting
		waitTime := arlm.baseLimiter.Reserve(key)
		if waitTime > 0 {
			return nil, fmt.Errorf("%w for tool %q, please wait %v", ErrRateLimitExceeded, tool.Definition().Name, waitTime)
		}
	}

	return next(ctx, tool, params)
}

// getSystemLoad returns a simplified system load metric (0.0 to 1.0).
// In a real implementation, this would check actual system metrics.
func (arlm *AdaptiveRateLimitMiddleware) getSystemLoad() float64 {
	// Simplified implementation - in practice, this would check:
	// - CPU usage
	// - Memory usage
	// - Active goroutines
	// - Response times
	return loadThresholdMedium // Return 30% load as example
}

// createAdaptiveKey creates a rate limit key that considers system load.
func (arlm *AdaptiveRateLimitMiddleware) createAdaptiveKey(ctx context.Context, tool interfaces.Tool, load float64) string {
	baseKey := tool.Definition().Name

	// Get execution context
	execCtx, hasCtx := GetExecutionContext(ctx)
	if hasCtx && execCtx.UserID != "" {
		baseKey = fmt.Sprintf("%s:%s", execCtx.UserID, baseKey)
	}

	// Add load bucket to the key
	const loadScale = 10
	loadBucket := int(load * loadScale) // 0-10 scale

	return fmt.Sprintf("%s:load_%d", baseKey, loadBucket)
}
