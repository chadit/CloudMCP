package server

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/server"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/internal/services/linode"
	"github.com/chadit/CloudMCP/pkg/interfaces"
	"github.com/chadit/CloudMCP/pkg/logger"
)

type Server struct {
	config   *config.Config
	logger   logger.Logger
	mcp      *server.MCPServer
	services []interfaces.CloudService
}

func New(cfg *config.Config, log logger.Logger) (*Server, error) {
	mcpServer := server.NewMCPServer(
		cfg.ServerName,
		"0.1.0",
		server.WithToolCapabilities(true),
		server.WithLogging(),
	)

	s := &Server{
		config:   cfg,
		logger:   log,
		mcp:      mcpServer,
		services: make([]interfaces.CloudService, 0),
	}

	linodeSvc, err := linode.New(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create Linode service: %w", err)
	}
	s.services = append(s.services, linodeSvc)

	for _, svc := range s.services {
		if err := svc.RegisterTools(mcpServer); err != nil {
			return nil, fmt.Errorf("failed to register tools for %s: %w", svc.Name(), err)
		}
	}

	return s, nil
}

func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("Initializing services")

	for _, svc := range s.services {
		if err := svc.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize %s: %w", svc.Name(), err)
		}
		s.logger.Info("Service initialized", "service", svc.Name())
	}

	s.logger.Info("Starting MCP server")
	
	if err := server.ServeStdio(s.mcp); err != nil {
		return fmt.Errorf("MCP server error: %w", err)
	}

	defer func() {
		for _, svc := range s.services {
			if err := svc.Shutdown(context.Background()); err != nil {
				s.logger.Error("Failed to shutdown service",
					"service", svc.Name(),
					"error", err,
				)
			}
		}
	}()

	return nil
}