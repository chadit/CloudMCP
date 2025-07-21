// Package server provides the minimal CloudMCP server implementation.
package server

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/internal/tools"
	"github.com/chadit/CloudMCP/pkg/contracts"
)

// Server represents a minimal CloudMCP server with simple tools.
type Server struct {
	config *config.Config
	mcp    *server.MCPServer
	tools  []contracts.Tool
}

// Static errors for err113 compliance.
var (
	ErrConfigNil             = errors.New("config cannot be nil")
	ErrExecuteNotImplemented = errors.New("execute method not implemented for wrapper")
)

// New creates a new minimal CloudMCP server with hello and version tools.
func New(cfg *config.Config) (*Server, error) {
	if cfg == nil {
		return nil, ErrConfigNil
	}

	// Create MCP server
	mcpServer := server.NewMCPServer(
		cfg.ServerName,
		"0.1.0",
		server.WithToolCapabilities(true),
	)

	// Create server instance
	s := &Server{
		config: cfg,
		mcp:    mcpServer,
		tools:  make([]contracts.Tool, 0),
	}

	// Register simple tools
	if err := s.registerTools(); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}

	return s, nil
}

// toolWrapper wraps mcp.Tool to implement our contracts.Tool for compatibility.
type toolWrapper struct {
	tool mcp.Tool
}

func (tw *toolWrapper) Name() string {
	return tw.tool.Name
}

func (tw *toolWrapper) Description() string {
	return tw.tool.Description
}

func (tw *toolWrapper) InputSchema() any {
	return tw.tool.InputSchema
}

func (tw *toolWrapper) Execute(_ context.Context, _ map[string]any) (*mcp.CallToolResult, error) {
	// This method is not used by the MCP server, but needed for interface compatibility
	return nil, ErrExecuteNotImplemented
}

// Start starts the minimal CloudMCP server.
func (s *Server) Start(_ context.Context) error {
	log.Printf("Starting CloudMCP minimal server with %d tools", len(s.tools))

	// Log registered tools
	for _, tool := range s.tools {
		log.Printf("Registered tool: %s - %s", tool.Name(), tool.Description())
	}

	// Start MCP server (blocks until context is cancelled or error occurs)
	log.Printf("CloudMCP server started successfully")
	return server.ServeStdio(s.mcp)
}

// GetToolCount returns the number of registered tools.
func (s *Server) GetToolCount() int {
	return len(s.tools)
}

// registerTools registers the simple hello and version tools.
func (s *Server) registerTools() error {
	// Create and register hello tool
	helloTool, helloHandler := tools.NewHelloTool()
	s.mcp.AddTool(helloTool, helloHandler)
	// Create a wrapper tool to maintain interface compatibility
	s.tools = append(s.tools, &toolWrapper{tool: helloTool})

	// Create and register version tool
	versionTool, versionHandler := tools.NewVersionTool()
	s.mcp.AddTool(versionTool, versionHandler)
	// Create a wrapper tool to maintain interface compatibility
	s.tools = append(s.tools, &toolWrapper{tool: versionTool})

	return nil
}
