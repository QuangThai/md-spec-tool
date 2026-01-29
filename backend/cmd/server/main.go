package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/yourorg/md-spec-tool/internal/config"
	"github.com/yourorg/md-spec-tool/internal/http"
)

func main() {
	godotenv.Load()

	cfg := config.LoadConfig()
	log.Printf("Starting server on %s:%s", cfg.Host, cfg.Port)

	// Setup router
	router := http.SetupRouter(cfg)

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
