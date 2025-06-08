package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/internal/server"
	"github.com/chadit/CloudMCP/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.LogLevel)
	log.Info("Starting CloudMCP Server",
		"version", "0.1.0",
		"server_name", cfg.ServerName,
		"log_level", cfg.LogLevel,
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
		os.Exit(1)
	}

	if err := srv.Start(ctx); err != nil {
		log.Error("Server error", "error", err)
		os.Exit(1)
	}

	log.Info("Server shutdown complete")
}