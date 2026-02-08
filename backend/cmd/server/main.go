package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/yourorg/md-spec-tool/internal/config"
	httphandler "github.com/yourorg/md-spec-tool/internal/http"
)

func main() {
	// Try loading .env from multiple locations:
	// 1. Current directory (when running from backend/)
	// 2. Parent directory (when .env is in project root)
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	cfg := config.LoadConfig()
	slog.Info("Starting server", "host", cfg.Host, "port", cfg.Port, "ai_enabled", cfg.AIEnabled)

	// Setup router
	router := httphandler.SetupRouter(cfg)

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	
	// Create HTTP server with proper timeouts
	server := &http.Server{
		Addr:           addr,
		Handler:        router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Start server in goroutine
	go func() {
		slog.Info("HTTP server starting", "addr", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "err", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal (SIGINT, SIGTERM)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	slog.Info("Shutting down server...")
	
	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server shutdown error", "err", err)
		os.Exit(1)
	}

	slog.Info("Server shutdown complete")
}
