package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/yourorg/md-spec-tool/internal/config"
	"github.com/yourorg/md-spec-tool/internal/http"
)

func main() {
	// Try loading .env from multiple locations:
	// 1. Current directory (when running from backend/)
	// 2. Parent directory (when .env is in project root)
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	cfg := config.LoadConfig()
	log.Printf("Starting server on %s:%s", cfg.Host, cfg.Port)
	
	// Log OpenAI configuration status
	if cfg.OpenAIAPIKey != "" {
		log.Printf("OpenAI API key configured (model: %s)", cfg.OpenAIModel)
	} else {
		log.Printf("Warning: OpenAI API key not configured - AI suggestions will be disabled")
	}

	// Setup router
	router := http.SetupRouter(cfg)

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
