package main

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func shutdownGracefully(
	srv *http.Server,
	deps *Dependencies,
	infra *Infrastructure,
	timeout time.Duration,
	logger *zap.Logger,
) {
	logger.Info("Starting graceful shutdown", zap.Duration("timeout", timeout))

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	logger.Info("Shutting down HTTP server...")
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Error during server shutdown", zap.Error(err))
	} else {
		logger.Info("HTTP server stopped")
	}

	logger.Info("Shutting down rate limiter...")
	deps.RateLimiter.Shutdown()

	logger.Info("Closing cache connection...")
	if err := infra.Cache.Close(); err != nil {
		logger.Warn("Error closing cache", zap.Error(err))
	} else {
		logger.Info("Cache connection closed")
	}

	logger.Info("Closing database connection...")
	if err := infra.DB.Close(); err != nil {
		logger.Warn("Error closing database", zap.Error(err))
	} else {
		logger.Info("Database connection closed")
	}

	logger.Info("Graceful shutdown completed")
}
