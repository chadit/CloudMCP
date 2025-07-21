// Package server provides the metrics HTTP server for CloudMCP.
// It exposes Prometheus metrics, health checks, and provider-specific health status.
package server

import (
	"context"
	"crypto/subtle"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"

	"github.com/chadit/CloudMCP/pkg/logger"
	"github.com/chadit/CloudMCP/pkg/metrics"
)

const (
	// Default values for metrics server configuration.
	defaultMetricsPort       = 8080
	defaultShutdownTimeout   = 30 * time.Second
	defaultReadTimeout       = 10 * time.Second
	defaultWriteTimeout      = 10 * time.Second
	defaultReadHeaderTimeout = 5 * time.Second
	defaultIdleTimeout       = 60 * time.Second

	// Health status constants.
	HealthStatusHealthy          = "healthy"
	HealthStatusDegraded         = "degraded"
	HealthStatusUnhealthy        = "unhealthy"
	defaultMaxHeaderBytes        = 1 << 20 // 1MB
	defaultHealthCheckTimeout    = 5 * time.Second
	defaultProviderHealthTimeout = 10 * time.Second

	// Security defaults.
	defaultRateLimitPerSecond = 10.0
	defaultRateLimitBurst     = 20
	defaultTLSMinVersion      = tls.VersionTLS13 // Use TLS 1.3 for enhanced security
)

// Static errors for error checking compliance.
var (
	ErrMetricsServerNil        = errors.New("metrics server is nil")
	ErrMetricsProviderNil      = errors.New("metrics provider is nil")
	ErrServerAlreadyRunning    = errors.New("metrics server is already running")
	ErrServerNotRunning        = errors.New("metrics server is not running")
	ErrInvalidPort             = errors.New("invalid metrics port")
	ErrHealthCheckFailed       = errors.New("health check failed")
	ErrProviderHealthFailed    = errors.New("provider health check failed")
	ErrShutdownTimeout         = errors.New("metrics server shutdown timeout")
	ErrMetricsProviderDisabled = errors.New("metrics provider is disabled")
	ErrUnauthorized            = errors.New("unauthorized access")
	ErrRateLimitExceeded       = errors.New("rate limit exceeded")
	ErrInvalidCredentials      = errors.New("invalid authentication credentials")
)

// MetricsServerConfig holds configuration for the metrics HTTP server.
type MetricsServerConfig struct {
	// Port is the HTTP port to listen on for metrics endpoints.
	Port int

	// ShutdownTimeout is the maximum time to wait for graceful shutdown.
	ShutdownTimeout time.Duration

	// ReadTimeout is the maximum duration for reading the entire request.
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out writes of the response.
	WriteTimeout time.Duration

	// ReadHeaderTimeout is the amount of time allowed to read request headers.
	ReadHeaderTimeout time.Duration

	// IdleTimeout is the maximum amount of time to wait for the next request.
	IdleTimeout time.Duration

	// MaxHeaderBytes controls the maximum number of bytes the server will read.
	MaxHeaderBytes int

	// HealthCheckTimeout is the timeout for individual health checks.
	HealthCheckTimeout time.Duration

	// ProviderHealthTimeout is the timeout for provider health checks.
	ProviderHealthTimeout time.Duration

	// Security configuration.

	// EnableTLS enables HTTPS with TLS encryption.
	EnableTLS bool

	// TLSCertFile is the path to the TLS certificate file.
	TLSCertFile string

	// TLSKeyFile is the path to the TLS private key file.
	TLSKeyFile string

	// TLSMinVersion is the minimum TLS version to support.
	TLSMinVersion uint16

	// BasicAuthUsername is the username for basic authentication (empty disables auth).
	BasicAuthUsername string

	// BasicAuthPassword is the password for basic authentication.
	BasicAuthPassword string

	// EnableRateLimit enables rate limiting for requests.
	EnableRateLimit bool

	// RateLimitPerSecond is the number of requests allowed per second.
	RateLimitPerSecond float64

	// RateLimitBurst is the burst size for rate limiting.
	RateLimitBurst int

	// EnableSecurityHeaders enables security HTTP headers.
	EnableSecurityHeaders bool
}

// DefaultMetricsServerConfig returns a default metrics server configuration.
func DefaultMetricsServerConfig() *MetricsServerConfig {
	return &MetricsServerConfig{
		Port:                  defaultMetricsPort,
		ShutdownTimeout:       defaultShutdownTimeout,
		ReadTimeout:           defaultReadTimeout,
		WriteTimeout:          defaultWriteTimeout,
		ReadHeaderTimeout:     defaultReadHeaderTimeout,
		IdleTimeout:           defaultIdleTimeout,
		MaxHeaderBytes:        defaultMaxHeaderBytes,
		HealthCheckTimeout:    defaultHealthCheckTimeout,
		ProviderHealthTimeout: defaultProviderHealthTimeout,

		// Security defaults.
		EnableTLS:             false,
		TLSMinVersion:         defaultTLSMinVersion,
		BasicAuthUsername:     "",
		BasicAuthPassword:     "",
		EnableRateLimit:       true,
		RateLimitPerSecond:    defaultRateLimitPerSecond,
		RateLimitBurst:        defaultRateLimitBurst,
		EnableSecurityHeaders: true,
	}
}

// HealthStatus represents the health status of a component.
type HealthStatus struct {
	Status    string    `json:"status"`    // HealthStatusHealthy, HealthStatusDegraded, HealthStatusUnhealthy
	Message   string    `json:"message"`   // Human-readable status message
	Timestamp time.Time `json:"timestamp"` // When the status was last checked
	Duration  string    `json:"duration"`  // How long the check took
}

// ProviderHealth represents the health status of a cloud provider.
type ProviderHealth struct {
	Name      string                 `json:"name"`      // Provider name (e.g., "linode")
	Status    string                 `json:"status"`    // Overall provider status
	Message   string                 `json:"message"`   // Overall status message
	Timestamp time.Time              `json:"timestamp"` // When the status was last checked
	Accounts  map[string]interface{} `json:"accounts"`  // Per-account health data
}

// OverallHealth represents the overall system health status.
type OverallHealth struct {
	Status     string                    `json:"status"`     // Overall system status
	Message    string                    `json:"message"`    // Overall status message
	Timestamp  time.Time                 `json:"timestamp"`  // When the status was last checked
	Components map[string]HealthStatus   `json:"components"` // Individual component health
	Providers  map[string]ProviderHealth `json:"providers"`  // Provider-specific health
	Metrics    map[string]interface{}    `json:"metrics"`    // Key metrics summary
}

// MetricsServer provides an HTTP server for exposing metrics and health endpoints.
type MetricsServer struct {
	config     *MetricsServerConfig
	logger     logger.Logger
	provider   metrics.Provider
	httpServer *http.Server
	mux        *http.ServeMux
	listener   net.Listener // Store listener to get actual port

	// Security components.
	rateLimiter *rate.Limiter

	mu      sync.RWMutex
	running bool
}

// NewMetricsServer creates a new metrics HTTP server with the specified configuration.
func NewMetricsServer(
	config *MetricsServerConfig,
	logger logger.Logger,
	provider metrics.Provider,
) (*MetricsServer, error) {
	if config == nil {
		config = DefaultMetricsServerConfig()
	}

	if logger == nil {
		return nil, ErrLoggerNil
	}

	if provider == nil {
		return nil, ErrMetricsProviderNil
	}

	if config.Port < 0 || config.Port > 65535 {
		return nil, ErrInvalidPort
	}

	if !provider.IsEnabled() {
		logger.Warn("Creating metrics server with disabled metrics provider")
	}

	mux := http.NewServeMux()

	server := &MetricsServer{
		config:   config,
		logger:   logger.With("component", "metrics_server"),
		provider: provider,
		mux:      mux,
	}

	// Initialize security components.
	if config.EnableRateLimit {
		server.rateLimiter = rate.NewLimiter(rate.Limit(config.RateLimitPerSecond), config.RateLimitBurst)
	}

	// Register HTTP handlers.
	server.registerHandlers()

	// Create HTTP server with timeouts and security settings.
	server.httpServer = &http.Server{
		Addr:              ":" + strconv.Itoa(config.Port),
		Handler:           server.securityMiddleware(mux),
		ReadTimeout:       config.ReadTimeout,
		WriteTimeout:      config.WriteTimeout,
		ReadHeaderTimeout: config.ReadHeaderTimeout,
		IdleTimeout:       config.IdleTimeout,
		MaxHeaderBytes:    config.MaxHeaderBytes,
	}

	// Configure TLS if enabled.
	if config.EnableTLS {
		// #nosec G402 -- TLS configuration is secure: defaults to TLS 1.3, uses strong cipher suites, configurable via environment
		server.httpServer.TLSConfig = &tls.Config{
			MinVersion: config.TLSMinVersion,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			},
			PreferServerCipherSuites: true,
		}
	}

	return server, nil
}

// Start starts the metrics HTTP server in a non-blocking way.
func (s *MetricsServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return ErrServerAlreadyRunning
	}

	s.logger.Info("Starting metrics server", "port", s.config.Port)

	// Create listener to handle port 0 (OS-assigned port).
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Port))
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	s.listener = listener

	// Update config with actual port if port 0 was used.
	if s.config.Port == 0 {
		if addr, ok := listener.Addr().(*net.TCPAddr); ok {
			s.config.Port = addr.Port
		}
	}

	// Start HTTP server in a goroutine.
	go func() {
		var err error

		if s.config.EnableTLS {
			// Start HTTPS server.
			if s.config.TLSCertFile == "" || s.config.TLSKeyFile == "" {
				s.logger.Error("TLS enabled but certificate or key file not specified")

				return
			}

			err = s.httpServer.ServeTLS(listener, s.config.TLSCertFile, s.config.TLSKeyFile)
		} else {
			// Start HTTP server.
			err = s.httpServer.Serve(listener)
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("Metrics server error", "error", err)
		}
	}()

	s.running = true

	protocol := "HTTP"

	if s.config.EnableTLS {
		protocol = "HTTPS"
	}

	s.logger.Info("Metrics server started successfully",
		"protocol", protocol,
		"port", s.config.Port,
		"tls_enabled", s.config.EnableTLS,
		"auth_enabled", s.isAuthEnabled(),
		"rate_limit_enabled", s.config.EnableRateLimit,
	)

	return nil
}

// Stop gracefully stops the metrics HTTP server.
func (s *MetricsServer) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return ErrServerNotRunning
	}

	s.logger.Info("Stopping metrics server")

	// Create shutdown context with timeout.
	shutdownCtx, cancel := context.WithTimeout(ctx, s.config.ShutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown.
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Failed to shutdown metrics server gracefully", "error", err)

		// Force close if graceful shutdown fails.
		if closeErr := s.httpServer.Close(); closeErr != nil {
			s.logger.Error("Failed to force close metrics server", "error", closeErr)
		}

		// Also close the listener.
		if s.listener != nil {
			if closeErr := s.listener.Close(); closeErr != nil {
				s.logger.Error("Failed to close listener", "error", closeErr)
			}
		}

		s.running = false

		return fmt.Errorf("metrics server shutdown: %w", err)
	}

	s.running = false
	s.logger.Info("Metrics server stopped successfully")

	return nil
}

// IsRunning returns whether the metrics server is currently running.
func (s *MetricsServer) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.running
}

// Port returns the port the metrics server is configured to listen on.
func (s *MetricsServer) Port() int {
	return s.config.Port
}

// URL returns the base URL for the metrics server.
func (s *MetricsServer) URL() string {
	protocol := "http"

	if s.config.EnableTLS {
		protocol = "https"
	}

	return fmt.Sprintf("%s://localhost:%d", protocol, s.config.Port)
}

// authenticateRequest performs basic authentication.
func (s *MetricsServer) authenticateRequest(r *http.Request) bool {
	auth := r.Header.Get("Authorization")

	if auth == "" {
		return false
	}

	// Check for Basic auth scheme.
	const prefix = "Basic "

	if !strings.HasPrefix(auth, prefix) {
		return false
	}

	// Decode base64 credentials.
	encoded := auth[len(prefix):]

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return false
	}

	// Split username:password.
	credentials := string(decoded)
	colonIndex := strings.Index(credentials, ":")

	if colonIndex == -1 {
		return false
	}

	username := credentials[:colonIndex]
	password := credentials[colonIndex+1:]

	// Use constant-time comparison to prevent timing attacks.
	usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(s.config.BasicAuthUsername)) == 1
	passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(s.config.BasicAuthPassword)) == 1

	return usernameMatch && passwordMatch
}

// registerHandlers registers all HTTP endpoints for the metrics server.
func (s *MetricsServer) registerHandlers() {
	// Prometheus metrics endpoint.
	s.mux.Handle("/metrics", s.metricsHandler())

	// Basic health check endpoint.
	s.mux.HandleFunc("/health", s.healthHandler)

	// Provider-specific health check endpoint.
	s.mux.HandleFunc("/provider/health", s.providerHealthHandler)

	// Root endpoint with basic server info.
	s.mux.HandleFunc("/", s.rootHandler)
}

// metricsHandler returns the Prometheus metrics handler with proper error handling.
func (s *MetricsServer) metricsHandler() http.Handler {
	if !s.provider.IsEnabled() {
		// Return a handler that explains metrics are disabled.
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusServiceUnavailable)

			_, err := w.Write([]byte("# Metrics collection is disabled\n"))
			if err != nil {
				s.logger.Error("Failed to write metrics disabled response", "error", err)
			}
		})
	}
	// Use Prometheus default handler.
	return promhttp.Handler()
}

// healthHandler provides a basic health check endpoint.
func (s *MetricsServer) healthHandler(responseWriter http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(request.Context(), s.config.HealthCheckTimeout)
	defer cancel()

	start := time.Now()

	// Perform basic health checks.
	health := s.performHealthCheck(ctx)
	health.Duration = time.Since(start).String()

	// Set response headers.
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	// Set status code based on health.
	statusCode := http.StatusOK

	switch health.Status {
	case HealthStatusDegraded:
		statusCode = http.StatusPartialContent
	case HealthStatusUnhealthy:
		statusCode = http.StatusServiceUnavailable
	}

	responseWriter.WriteHeader(statusCode)

	// Encode and send response.
	if err := json.NewEncoder(responseWriter).Encode(health); err != nil {
		s.logger.Error("Failed to encode health response", "error", err)
		http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)

		return
	}

	// Log health check.
	s.logger.Debug("Health check completed",
		"status", health.Status,
		"duration", health.Duration,
		"remote_addr", request.RemoteAddr,
	)
}

// providerHealthHandler provides detailed provider-specific health information.
func (s *MetricsServer) providerHealthHandler(responseWriter http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(request.Context(), s.config.ProviderHealthTimeout)
	defer cancel()

	start := time.Now()

	// Perform comprehensive health checks.
	health := s.performComprehensiveHealthCheck(ctx)

	// Set response headers.
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	// Set status code based on overall health.
	statusCode := http.StatusOK

	switch health.Status {
	case HealthStatusDegraded:
		statusCode = http.StatusPartialContent
	case HealthStatusUnhealthy:
		statusCode = http.StatusServiceUnavailable
	}

	responseWriter.WriteHeader(statusCode)

	// Encode and send response.
	if err := json.NewEncoder(responseWriter).Encode(health); err != nil {
		s.logger.Error("Failed to encode provider health response", "error", err)
		http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)

		return
	}

	// Log provider health check.
	duration := time.Since(start)
	s.logger.Debug("Provider health check completed",
		"status", health.Status,
		"duration", duration.String(),
		"providers", len(health.Providers),
		"remote_addr", request.RemoteAddr,
	)
}

// rootHandler provides basic server information.
func (s *MetricsServer) rootHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/" {
		http.NotFound(responseWriter, request)

		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")

	info := map[string]interface{}{
		"service":   "CloudMCP Metrics Server",
		"status":    "running",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"endpoints": map[string]string{
			"metrics":         "/metrics",
			"health":          "/health",
			"provider_health": "/provider/health",
		},
		"metrics_enabled": s.provider.IsEnabled(),
	}

	if err := json.NewEncoder(responseWriter).Encode(info); err != nil {
		s.logger.Error("Failed to encode server info response", "error", err)
		http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)

		return
	}
}

// performHealthCheck performs a basic health check of the metrics server.
func (s *MetricsServer) performHealthCheck(ctx context.Context) *HealthStatus {
	health := &HealthStatus{
		Timestamp: time.Now().UTC(),
	}

	// Check if metrics provider is enabled and healthy.
	if !s.provider.IsEnabled() {
		health.Status = HealthStatusDegraded
		health.Message = "Metrics collection is disabled"

		return health
	}

	// Try to record a test metric to verify provider functionality.
	select {
	case <-ctx.Done():
		health.Status = HealthStatusUnhealthy
		health.Message = "Health check timeout"

		return health
	default:
		// Quick health check - just verify provider responds.
		s.provider.IncrementCounter("health_check", map[string]string{
			"component": "metrics_server",
		})

		health.Status = HealthStatusHealthy
		health.Message = "All systems operational"
	}

	return health
}

// performComprehensiveHealthCheck performs detailed health checks including providers.
func (s *MetricsServer) performComprehensiveHealthCheck(ctx context.Context) *OverallHealth {
	start := time.Now()

	health := &OverallHealth{
		Timestamp:  start.UTC(),
		Components: make(map[string]HealthStatus),
		Providers:  make(map[string]ProviderHealth),
		Metrics:    make(map[string]interface{}),
	}

	var (
		waitGroup         sync.WaitGroup
		mutex             sync.Mutex
		componentStatuses = make([]string, 0)
	)

	// Check metrics provider health.
	waitGroup.Add(1)

	go func() {
		defer waitGroup.Done()

		componentHealth := HealthStatus{
			Timestamp: time.Now().UTC(),
		}

		if !s.provider.IsEnabled() {
			componentHealth.Status = HealthStatusDegraded
			componentHealth.Message = "Metrics collection disabled"
		} else {
			// Test metrics provider functionality.
			testStart := time.Now()

			s.provider.IncrementCounter("comprehensive_health_check", map[string]string{
				"component": "metrics_provider",
			})

			componentHealth.Duration = time.Since(testStart).String()

			componentHealth.Status = HealthStatusHealthy
			componentHealth.Message = "Metrics provider operational"
		}

		mutex.Lock()
		health.Components["metrics_provider"] = componentHealth
		componentStatuses = append(componentStatuses, componentHealth.Status)
		mutex.Unlock()
	}()

	// No providers in minimal shell mode

	// Wait for all checks to complete or timeout.
	done := make(chan struct{})
	go func() {
		waitGroup.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		health.Status = HealthStatusUnhealthy
		health.Message = "Health check timeout"

		return health
	case <-done:
	}

	// Determine overall status.
	health.Status = s.determineOverallStatus(componentStatuses)
	health.Message = s.generateOverallMessage(health.Status, len(health.Components), len(health.Providers))

	// Add summary metrics.
	health.Metrics["check_duration_seconds"] = time.Since(start).Seconds()
	health.Metrics["total_components"] = len(health.Components)
	health.Metrics["total_providers"] = len(health.Providers)
	health.Metrics["healthy_components"] = s.countHealthyComponents(health.Components)
	health.Metrics["healthy_providers"] = s.countHealthyProviders(health.Providers)

	return health
}

// determineOverallStatus determines the overall system status from component statuses.
func (s *MetricsServer) determineOverallStatus(statuses []string) string {
	hasUnhealthy := false
	hasDegraded := false

	for _, status := range statuses {
		switch status {
		case HealthStatusUnhealthy:
			hasUnhealthy = true
		case HealthStatusDegraded:
			hasDegraded = true
		}
	}

	switch {
	case hasUnhealthy:
		return HealthStatusUnhealthy
	case hasDegraded:
		return HealthStatusDegraded
	default:
		return HealthStatusHealthy
	}
}

// generateOverallMessage generates a human-readable overall status message.
func (s *MetricsServer) generateOverallMessage(status string, componentCount, providerCount int) string {
	switch status {
	case HealthStatusHealthy:
		return fmt.Sprintf("All %d components and %d providers are healthy", componentCount, providerCount)
	case HealthStatusDegraded:
		return fmt.Sprintf("Some components or providers are degraded (%d components, %d providers)", componentCount, providerCount)
	case HealthStatusUnhealthy:
		return fmt.Sprintf("One or more components or providers are unhealthy (%d components, %d providers)", componentCount, providerCount)
	default:
		return "Unknown status"
	}
}

// countHealthyComponents counts the number of healthy components.
func (s *MetricsServer) countHealthyComponents(components map[string]HealthStatus) int {
	count := 0

	for _, component := range components {
		if component.Status == HealthStatusHealthy {
			count++
		}
	}

	return count
}

// countHealthyProviders counts the number of healthy providers.
func (s *MetricsServer) countHealthyProviders(providers map[string]ProviderHealth) int {
	count := 0

	for _, provider := range providers {
		if provider.Status == HealthStatusHealthy {
			count++
		}
	}

	return count
}

// securityMiddleware applies security measures to all HTTP requests.
func (s *MetricsServer) securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		// Add security headers.
		if s.config.EnableSecurityHeaders {
			s.addSecurityHeaders(responseWriter)
		}

		// Rate limiting.
		if s.config.EnableRateLimit && s.rateLimiter != nil {
			if !s.rateLimiter.Allow() {
				s.logger.Warn("Rate limit exceeded", "remote_addr", request.RemoteAddr, "path", request.URL.Path)
				http.Error(responseWriter, "Rate limit exceeded", http.StatusTooManyRequests)

				return
			}
		}

		// Basic authentication (only for sensitive endpoints).
		if s.requiresAuth(request.URL.Path) && s.isAuthEnabled() {
			if !s.authenticateRequest(request) {
				responseWriter.Header().Set("WWW-Authenticate", `Basic realm="CloudMCP Metrics"`)
				http.Error(responseWriter, "Unauthorized", http.StatusUnauthorized)

				return
			}
		}

		// Continue to next handler.
		next.ServeHTTP(responseWriter, request)
	})
}

// addSecurityHeaders adds security headers to the response.
func (s *MetricsServer) addSecurityHeaders(responseWriter http.ResponseWriter) {
	responseWriter.Header().Set("X-Content-Type-Options", "nosniff")
	responseWriter.Header().Set("X-Frame-Options", "DENY")
	responseWriter.Header().Set("X-XSS-Protection", "1; mode=block")
	responseWriter.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	responseWriter.Header().Set("Content-Security-Policy", "default-src 'self'")
	responseWriter.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
}

// requiresAuth determines if an endpoint requires authentication.
func (s *MetricsServer) requiresAuth(path string) bool {
	// Require auth for metrics endpoint, but not for health checks.
	switch path {
	case "/metrics":
		return true
	case "/provider/health":
		return true
	default:
		return false
	}
}

// isAuthEnabled checks if basic authentication is configured.
func (s *MetricsServer) isAuthEnabled() bool {
	return s.config.BasicAuthUsername != "" && s.config.BasicAuthPassword != ""
}
