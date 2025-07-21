package server

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/internal/registry"
	"github.com/chadit/CloudMCP/internal/tools"
	"github.com/chadit/CloudMCP/pkg/logger"
	"github.com/chadit/CloudMCP/pkg/metrics"
)

// Security and metrics server constants.
const (
	defaultMetricsRateLimit      = 10.0
	defaultMetricsRateLimitBurst = 20
)

type Server struct {
	config          *config.Config
	logger          logger.Logger
	mcp             *server.MCPServer
	healthTool      *tools.HealthCheckTool
	mcpAdapter      *registry.MCPServerAdapter
	metricsServer   *MetricsServer
	metricsProvider metrics.Provider
}

// Static errors for err113 compliance.
var (
	ErrConfigNil               = errors.New("config cannot be nil")
	ErrLoggerNil               = errors.New("logger cannot be nil")
	ErrRegistryCreation        = errors.New("failed to create provider registry")
	ErrProviderSetup           = errors.New("failed to setup providers")
	ErrMetricsSetup            = errors.New("failed to setup metrics server")
	ErrHealthToolNotRegistered = errors.New("health check tool not properly registered")
)

func New(cfg *config.Config, log logger.Logger) (*Server, error) {
	if cfg == nil {
		return nil, ErrConfigNil
	}

	if log == nil {
		return nil, ErrLoggerNil
	}

	// Create MCP server.
	mcpServer := server.NewMCPServer(
		cfg.ServerName,
		"0.1.0",
		server.WithToolCapabilities(true),
	)

	// Initialize metrics provider.
	metricsProvider, err := metrics.NewProvider(&metrics.ProviderConfig{
		Enabled:   cfg.EnableMetrics,
		Namespace: "cloudmcp",
		Subsystem: "server",
		Backend:   metrics.BackendPrometheus,
		Tags: map[string]string{
			"service": cfg.ServerName,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics provider: %w", err)
	}

	// Create MCP adapter
	mcpAdapter := registry.NewMCPServerAdapter(mcpServer)

	// Create health tool
	healthTool := tools.NewHealthCheckTool(cfg.ServerName)

	// Create server instance.
	serverInstance := &Server{
		config:          cfg,
		logger:          log,
		mcp:             mcpServer,
		healthTool:      healthTool,
		mcpAdapter:      mcpAdapter,
		metricsProvider: metricsProvider,
	}

	// Register the health check tool
	if err := serverInstance.registerHealthTool(); err != nil {
		return nil, fmt.Errorf("failed to register health tool: %w", err)
	}

	// Setup metrics server if enabled (but don't start it yet).
	if err := serverInstance.setupMetricsServer(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMetricsSetup, err)
	}

	return serverInstance, nil
}

// NewForTesting creates a new server instance for testing with isolated metrics registries.
func NewForTesting(cfg *config.Config, log logger.Logger) (*Server, error) {
	if cfg == nil {
		return nil, ErrConfigNil
	}

	if log == nil {
		return nil, ErrLoggerNil
	}

	mcpServer := server.NewMCPServer(cfg.ServerName, "test-version")

	// Initialize test metrics provider.
	testMetricsProvider, err := metrics.NewProvider(&metrics.ProviderConfig{
		Enabled:   cfg.EnableMetrics,
		Namespace: "cloudmcp_test",
		Subsystem: "server",
		Backend:   metrics.BackendPrometheus,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create test metrics provider: %w", err)
	}

	// Create MCP adapter and health tool for testing
	mcpAdapter := registry.NewMCPServerAdapter(mcpServer)
	healthTool := tools.NewHealthCheckTool(cfg.ServerName)

	serverInstance := &Server{
		config:          cfg,
		logger:          log,
		mcp:             mcpServer,
		healthTool:      healthTool,
		mcpAdapter:      mcpAdapter,
		metricsProvider: testMetricsProvider,
	}

	// Register health tool for testing.
	if err := serverInstance.registerHealthTool(); err != nil {
		return nil, fmt.Errorf("failed to register health tool: %w", err)
	}

	return serverInstance, nil
}

// Start starts the CloudMCP server with minimal shell configuration.
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("Starting CloudMCP server", "mode", "minimal_shell")

	if err := s.setupMinimalShell(ctx); err != nil {
		return err
	}

	// Check if running in daemon mode (for containers)
	daemonMode := os.Getenv("DAEMON_MODE") == "true"

	if daemonMode {
		s.logger.Info("Running in daemon mode - HTTP server only")
		
		// Log server startup completion.
		s.logger.Info("CloudMCP server started successfully",
			"mode", "daemon",
			"tools_registered", s.mcpAdapter.GetToolCount(),
			"metrics_enabled", s.metricsServer != nil,
		)

		// In daemon mode, just wait for context cancellation
		<-ctx.Done()
		s.logger.Info("Context cancelled, shutting down")
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	s.logger.Info("Starting MCP protocol server")

	// Create a channel to signal when stdio server is done.
	errCh := make(chan error, 1)

	// Run ServeStdio in a goroutine.
	go func() {
		errCh <- server.ServeStdio(s.mcp)
	}()

	// Log server startup completion.
	s.logger.Info("CloudMCP server started successfully",
		"mode", "minimal_shell",
		"tools_registered", s.mcpAdapter.GetToolCount(),
		"metrics_enabled", s.metricsServer != nil,
	)

	// Wait for either context cancellation or server error.
	select {
	case <-ctx.Done():
		s.logger.Info("Context cancelled, shutting down")

		return fmt.Errorf("context cancelled: %w", ctx.Err())
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("MCP server error: %w", err)
		}

		return nil
	}
}

// registerHealthTool registers the health check tool with the MCP server.
func (s *Server) registerHealthTool() error {
	s.logger.Info("Registering health check tool")

	if err := s.mcpAdapter.RegisterTool(s.healthTool); err != nil {
		return fmt.Errorf("failed to register health tool: %w", err)
	}

	// Update health tool with current metrics
	s.healthTool.UpdateToolsCount(s.mcpAdapter.GetToolCount())

	s.logger.Info("Health check tool registered successfully", "tools_count", s.mcpAdapter.GetToolCount())

	return nil
}

// initializeMinimalShell initializes the minimal shell components.
func (s *Server) initializeMinimalShell(_ context.Context) {
	s.logger.Info("Initializing minimal shell mode")

	// Update health tool with final status
	s.healthTool.UpdateToolsCount(s.mcpAdapter.GetToolCount())

	s.logger.Info("Minimal shell initialization completed", "tools", s.mcpAdapter.GetToolCount())
}

// setupMetricsServer creates and configures the metrics HTTP server.
func (s *Server) setupMetricsServer() error {
	if !s.config.EnableMetrics {
		s.logger.Info("Metrics disabled, skipping metrics server setup")

		return nil
	}

	s.logger.Info("Setting up metrics HTTP server", "port", s.config.MetricsPort)

	// Check for authentication configuration.
	authConfigured := os.Getenv("METRICS_AUTH_USERNAME") != "" && os.Getenv("METRICS_AUTH_PASSWORD") != ""
	tlsConfigured := os.Getenv("METRICS_TLS_ENABLED") == "true"

	// Create metrics server configuration with security features enabled.
	metricsConfig := &MetricsServerConfig{
		Port:                  s.config.MetricsPort,
		EnableSecurityHeaders: true,
		EnableRateLimit:       true,
		RateLimitPerSecond:    defaultMetricsRateLimit,
		RateLimitBurst:        defaultMetricsRateLimitBurst,
		// TLS configuration via environment variables.
		EnableTLS:     os.Getenv("METRICS_TLS_ENABLED") == "true",
		TLSCertFile:   os.Getenv("METRICS_TLS_CERT_FILE"),
		TLSKeyFile:    os.Getenv("METRICS_TLS_KEY_FILE"),
		TLSMinVersion: defaultTLSMinVersion, // TLS 1.3
		// Basic auth credentials via environment variables.
		BasicAuthUsername: os.Getenv("METRICS_AUTH_USERNAME"),
		BasicAuthPassword: os.Getenv("METRICS_AUTH_PASSWORD"),
	}

	// Create metrics server for minimal shell mode (no providers).
	metricsServer, err := NewMetricsServer(
		metricsConfig,
		s.logger,
		s.metricsProvider,
	)
	if err != nil {
		return fmt.Errorf("failed to create metrics server: %w", err)
	}

	s.metricsServer = metricsServer
	s.logger.Info("Metrics server setup completed",
		"security_headers", true,
		"rate_limiting", true,
		"authentication", authConfigured,
		"tls_enabled", tlsConfigured,
	)

	return nil
}

// setupMinimalShell initializes the minimal shell configuration.
func (s *Server) setupMinimalShell(ctx context.Context) error {
	s.logger.Info("Setting up minimal shell configuration")

	// Initialize minimal shell components.
	s.initializeMinimalShell(ctx)

	// Start metrics server if enabled.
	if s.metricsServer != nil {
		// Update metrics server with empty providers list for minimal mode.
		if err := s.metricsServer.Start(); err != nil {
			return fmt.Errorf("failed to start metrics server: %w", err)
		}

		s.logger.Info("Metrics server started", "port", s.metricsServer.Port())
	}

	// Validate health tool registration.
	if !s.mcpAdapter.HasTool("health_check") {
		return ErrHealthToolNotRegistered
	}

	s.logger.Info("Minimal shell setup completed",
		"tools_registered", s.mcpAdapter.GetToolCount(),
		"health_tool_available", true,
	)

	return nil
}
