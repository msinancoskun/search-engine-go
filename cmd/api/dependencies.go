package main

import (
	"search-engine-go/internal/api/handler"
	"search-engine-go/internal/api/middleware"
	"search-engine-go/internal/config"
	"search-engine-go/internal/repository"
	"search-engine-go/internal/service"
	"search-engine-go/pkg/adapter"

	"go.uber.org/zap"
)

type Dependencies struct {
	ProviderService *service.ProviderService
	ScoringService  *service.ScoringService
	ContentService  *service.ContentService
	JWTService      *service.JWTService

	AuthHandler      *handler.AuthHandler
	ContentHandler   *handler.ContentHandler
	DashboardHandler *handler.DashboardHandler

	RateLimiter *middleware.RateLimiter
	Logger *zap.Logger
}

func initializeDependencies(infra *Infrastructure, adapters *adapter.AdapterRegistry, cfg *config.Config) (*Dependencies, error) {
	contentRepo := repository.NewContentRepository(infra.DB.GetDB())

	providerService := service.NewProviderService(adapters, infra.Logger)
	scoringService := service.NewScoringService()
	contentService := service.NewContentService(contentRepo, providerService, scoringService, infra.Cache, infra.Logger)

	jwtService := service.NewJWTService(cfg.Auth, infra.Logger)
	authHandler := handler.NewAuthHandler(jwtService, infra.Logger)
	contentHandler := handler.NewContentHandler(contentService, infra.Logger)
	dashboardHandler := handler.NewDashboardHandler(contentService, infra.Logger)

	rateLimiter := middleware.NewRateLimiter(cfg.Server.RateLimit, infra.Logger)

	return &Dependencies{
		ProviderService:  providerService,
		ScoringService:   scoringService,
		ContentService:   contentService,
		JWTService:       jwtService,
		AuthHandler:      authHandler,
		ContentHandler:   contentHandler,
		DashboardHandler: dashboardHandler,
		RateLimiter:      rateLimiter,
		Logger:           infra.Logger,
	}, nil
}
