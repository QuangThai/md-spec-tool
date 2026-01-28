package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/yourorg/md-spec-tool/internal/config"
	"github.com/yourorg/md-spec-tool/internal/database"
	"github.com/yourorg/md-spec-tool/internal/http"
)

func main() {
	godotenv.Load()

	cfg := config.LoadConfig()
	log.Printf("Starting server on %s:%s", cfg.Host, cfg.Port)

	// Connect to database
	pool, err := database.New(cfg.DBDsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Run migrations
	if err := database.RunMigrations(pool); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Setup router
	router := http.SetupRouter(cfg, pool)

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
