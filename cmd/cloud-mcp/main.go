// Package main provides the CloudMCP minimal server application entry point.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/internal/server"
	"github.com/chadit/CloudMCP/internal/version"
)

func main() {
	// Load minimal configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Log startup information
	versionInfo := version.Get()
	log.Printf("Starting CloudMCP Minimal Server")
	log.Printf("Version: %s", versionInfo.Version)
	log.Printf("Server: %s", cfg.ServerName)
	log.Printf("Platform: %s", versionInfo.Platform)
	log.Printf("Git Commit: %s", versionInfo.GitCommit)

	// Create context with cancellation support
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Printf("Shutdown signal received")
		cancel()
	}()

	// Create and start minimal server
	srv, err := server.New(cfg)
	if err != nil {
		log.Printf("Failed to create server: %v", err)
		os.Exit(1)
	}

	if err := srv.Start(ctx); err != nil {
		log.Printf("Server error: %v", err)
		os.Exit(1)
	}

	log.Printf("Server shutdown complete")
}