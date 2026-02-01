package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"search-engine-go/internal/config"

	"go.uber.org/zap"
)

func createServer(cfg config.ServerConfig, router http.Handler) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}

func startServer(
	srv *http.Server,
	logger *zap.Logger,
	deps *Dependencies,
	infra *Infrastructure,
	shutdownTimeout time.Duration,
) error {
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("Starting server", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		logger.Fatal("Server error", zap.Error(err))
		return err
	case sig := <-quit:
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
	}

	shutdownGracefully(srv, deps, infra, shutdownTimeout, logger)
	return nil
}
