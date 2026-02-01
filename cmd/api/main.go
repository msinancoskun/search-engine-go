package main

import (
	"log"

	"search-engine-go/internal/config"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	infra, err := initializeInfrastructure(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize infrastructure: %v", err)
	}
	defer infra.Logger.Sync()
	defer infra.DB.Close()

	adapters, err := setupProviders(cfg.Providers, infra.Logger)
	if err != nil {
		infra.Logger.Fatal("Failed to setup providers", zap.Error(err))
	}

	deps, err := initializeDependencies(infra, adapters, cfg)
	if err != nil {
		infra.Logger.Fatal("Failed to initialize dependencies", zap.Error(err))
	}
	defer deps.RateLimiter.Shutdown()

	router := setupRouter(cfg, deps)
	server := createServer(cfg.Server, router)

	if err := startServer(server, infra.Logger, deps, infra, cfg.Server.ShutdownTimeout); err != nil {
		infra.Logger.Fatal("Failed to start server", zap.Error(err))
	}
}
