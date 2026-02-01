package main

import (
	"log"

	"search-engine-go/internal/config"
	"search-engine-go/internal/infrastructure/cache"
	"search-engine-go/internal/infrastructure/database"
	"search-engine-go/internal/infrastructure/logger"

	"go.uber.org/zap"
)

type Infrastructure struct {
	Logger *zap.Logger
	DB     *database.Postgres
	Cache  cache.Cache
}

func initializeInfrastructure(cfg *config.Config) (*Infrastructure, error) {
	zapLogger, err := logger.New(cfg.Log.Level, cfg.Log.Output)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	db, err := database.NewPostgres(cfg.Database, zapLogger)
	if err != nil {
		zapLogger.Error("Failed to connect to database", zap.Error(err))
		return nil, err
	}

	cacheClient, err := cache.NewRedis(cfg.Cache)
	if err != nil {
		zapLogger.Warn("Failed to connect to cache, continuing without cache", zap.Error(err))
		cacheClient = cache.NewInMemory()
	}

	return &Infrastructure{
		Logger: zapLogger,
		DB:     db,
		Cache:  cacheClient,
	}, nil
}
