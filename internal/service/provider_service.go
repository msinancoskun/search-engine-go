package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"search-engine-go/internal/domain"
	"search-engine-go/internal/infrastructure/circuitbreaker"
	"search-engine-go/pkg/adapter"

	"go.uber.org/zap"
)

type ProviderService struct {
	registry        *adapter.AdapterRegistry
	log             *zap.Logger
	circuitBreakers map[string]*circuitbreaker.CircuitBreaker
	mu              sync.RWMutex
}

func NewProviderService(registry *adapter.AdapterRegistry, log *zap.Logger) *ProviderService {
	return &ProviderService{
		registry:        registry,
		log:             log,
		circuitBreakers: make(map[string]*circuitbreaker.CircuitBreaker),
	}
}

func (s *ProviderService) getCircuitBreaker(providerName string) *circuitbreaker.CircuitBreaker {
	s.mu.RLock()
	cb, exists := s.circuitBreakers[providerName]
	s.mu.RUnlock()

	if exists {
		return cb
	}

	cb = circuitbreaker.NewCircuitBreaker(5, 30*time.Second)

	s.mu.Lock()
	s.circuitBreakers[providerName] = cb
	s.mu.Unlock()

	return cb
}

func (s *ProviderService) FetchFromAllProviders(ctx context.Context, query string, contentType *domain.ContentType) ([]*domain.Content, error) {
	adapters := s.registry.GetAll()
	if s.hasNoAdapters(adapters) {
		return []*domain.Content{}, nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var allContents []*domain.Content
	var errors []error

	for name, adpt := range adapters {
		wg.Add(1)
		go func(providerName string, providerAdapter adapter.ProviderAdapter) {
			defer wg.Done()

			cb := s.getCircuitBreaker(providerName)

			var contents []*domain.Content
			var err error

			cbErr := cb.Execute(ctx, func() error {
				contents, err = providerAdapter.FetchContent(ctx, query, contentType)
				return err
			})

			if cbErr != nil {
				if cbErr.Error() == "circuit breaker is open" {
					s.log.Warn("Circuit breaker is open for provider",
						zap.String("provider", providerName),
						zap.String("state", "open"),
					)
					mu.Lock()
					errors = append(errors, fmt.Errorf("provider %s: circuit breaker is open", providerName))
					mu.Unlock()
					return
				}

				s.log.Warn("Failed to fetch from provider",
					zap.String("provider", providerName),
					zap.Error(cbErr),
					zap.String("circuit_state", cb.GetState().String()),
				)
				mu.Lock()
				errors = append(errors, fmt.Errorf("provider %s: %w", providerName, cbErr))
				mu.Unlock()
				return
			}

			mu.Lock()
			allContents = append(allContents, contents...)
			mu.Unlock()

			s.log.Debug("Fetched content from provider",
				zap.String("provider", providerName),
				zap.Int("count", len(contents)),
				zap.String("circuit_state", cb.GetState().String()),
			)
		}(name, adpt)
	}

	wg.Wait()

	if s.allProvidersFailed(allContents, errors) {
		return nil, fmt.Errorf("all providers failed: %v", errors)
	}

	return allContents, nil
}

func (s *ProviderService) hasNoAdapters(adapters map[string]adapter.ProviderAdapter) bool {
	return len(adapters) == 0
}

func (s *ProviderService) allProvidersFailed(contents []*domain.Content, errors []error) bool {
	return len(contents) == 0 && len(errors) > 0
}
