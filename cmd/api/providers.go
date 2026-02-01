package main

import (
	"search-engine-go/internal/config"
	"search-engine-go/pkg/adapter"

	"go.uber.org/zap"
)

func setupProviders(cfg config.ProvidersConfig, logger *zap.Logger) (*adapter.AdapterRegistry, error) {
	adapters := adapter.NewAdapterRegistry()

	provider1Adapter := adapter.NewJSONProviderAdapterWithRetry(
		"provider1",
		"mocks/json_provider.json",
		cfg.Provider1.RateLimit,
		cfg.Provider1.Timeout,
		cfg.Provider1.RetryCount,
		cfg.Provider1.RetryDelay,
	)
	adapters.Register("provider1", provider1Adapter)
	logger.Info("Registered provider", zap.String("name", "provider1"), zap.String("type", "JSON"), zap.String("source", "mock file"))

	provider2Adapter := adapter.NewXMLProviderAdapterWithRetry(
		"provider2",
		"mocks/xml_provider.xml",
		cfg.Provider2.RateLimit,
		cfg.Provider2.Timeout,
		cfg.Provider2.RetryCount,
		cfg.Provider2.RetryDelay,
	)
	adapters.Register("provider2", provider2Adapter)
	logger.Info("Registered provider", zap.String("name", "provider2"), zap.String("type", "XML"), zap.String("source", "mock file"))

	return adapters, nil
}
