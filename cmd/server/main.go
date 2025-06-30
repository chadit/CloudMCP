package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/internal/server"
	"github.com/chadit/CloudMCP/internal/version"
	"github.com/chadit/CloudMCP/pkg/logger"
)

const (
	defaultLogMaxSize    = 10 // 10MB
	defaultLogMaxBackups = 5  // Keep 5 files
	defaultLogMaxAge     = 30 // 30 days
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Set up enhanced logging with rotation
	logConfig := logger.LogConfig{
		Level:      cfg.LogLevel,
		FilePath:   config.GetLogPath(),
		MaxSize:    defaultLogMaxSize,    // 10MB
		MaxBackups: defaultLogMaxBackups, // Keep 5 files
		MaxAge:     defaultLogMaxAge,     // 30 days
	}

	// Ensure log directory exists
	if err := config.EnsureLogDir(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create log directory: %v\n", err)
		// Fall back to stderr logging
		logConfig.FilePath = ""
	}

	log := logger.NewWithConfig(logConfig)
	versionInfo := version.Get()
	log.Info("Starting CloudMCP Server",
		"version", versionInfo.Version,
		"server_name", cfg.ServerName,
		"log_level", cfg.LogLevel,
		"log_file", logConfig.FilePath,
		"api_version", versionInfo.APIVersion,
		"platform", versionInfo.Platform,
		"git_commit", versionInfo.GitCommit,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("Shutdown signal received")
		cancel()
	}()

	srv, err := server.New(cfg, log)
	if err != nil {
		log.Error("Failed to create server", "error", err)

		return
	}

	if err := srv.Start(ctx); err != nil {
		log.Error("Server error", "error", err)

		return
	}

	log.Info("Server shutdown complete")
}
