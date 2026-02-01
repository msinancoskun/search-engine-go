package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"search-engine-go/internal/domain"
	"search-engine-go/pkg/adapter"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type MockProviderAdapter struct {
	name     string
	contents []*domain.Content
	err      error
	delay    time.Duration
}

func (m *MockProviderAdapter) GetName() string {
	return m.name
}

func (m *MockProviderAdapter) GetRateLimit() int {
	return 60
}

func (m *MockProviderAdapter) FetchContent(ctx context.Context, query string, contentType *domain.ContentType) ([]*domain.Content, error) {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.contents, nil
}

func TestProviderService_FetchFromAllProviders(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("Fetch from single provider successfully", func(t *testing.T) {
		registry := adapter.NewAdapterRegistry()
		mockAdapter := &MockProviderAdapter{
			name: "provider1",
			contents: []*domain.Content{
				{ProviderID: "p1_1", Provider: "provider1", Title: "Content 1", Type: domain.ContentTypeVideo, Views: 100, Likes: 10, CreatedAt: time.Now()},
			},
		}
		registry.Register("provider1", mockAdapter)

		service := NewProviderService(registry, logger)

		contents, err := service.FetchFromAllProviders(context.Background(), "test", nil)

		assert.NoError(t, err)
		assert.Len(t, contents, 1)
		assert.Equal(t, "Content 1", contents[0].Title)
	})

	t.Run("Fetch from multiple providers successfully", func(t *testing.T) {
		registry := adapter.NewAdapterRegistry()
		mockAdapter1 := &MockProviderAdapter{
			name: "provider1",
			contents: []*domain.Content{
				{ProviderID: "p1_1", Provider: "provider1", Title: "Content 1", Type: domain.ContentTypeVideo, Views: 100, Likes: 10, CreatedAt: time.Now()},
			},
		}
		mockAdapter2 := &MockProviderAdapter{
			name: "provider2",
			contents: []*domain.Content{
				{ProviderID: "p2_1", Provider: "provider2", Title: "Content 2", Type: domain.ContentTypeText, ReadingTime: 5, Reactions: 25, CreatedAt: time.Now()},
			},
		}
		registry.Register("provider1", mockAdapter1)
		registry.Register("provider2", mockAdapter2)

		service := NewProviderService(registry, logger)

		contents, err := service.FetchFromAllProviders(context.Background(), "test", nil)

		assert.NoError(t, err)
		assert.Len(t, contents, 2)
	})

	t.Run("Fetch with content type filter", func(t *testing.T) {
		registry := adapter.NewAdapterRegistry()
		contentType := domain.ContentTypeVideo
		mockAdapter := &MockProviderAdapter{
			name: "provider1",
			contents: []*domain.Content{
				{ProviderID: "p1_1", Provider: "provider1", Title: "Video Content", Type: domain.ContentTypeVideo, Views: 100, Likes: 10, CreatedAt: time.Now()},
			},
		}
		registry.Register("provider1", mockAdapter)

		service := NewProviderService(registry, logger)

		contents, err := service.FetchFromAllProviders(context.Background(), "test", &contentType)

		assert.NoError(t, err)
		assert.Len(t, contents, 1)
		assert.Equal(t, domain.ContentTypeVideo, contents[0].Type)
	})

	t.Run("Handle provider failure gracefully", func(t *testing.T) {
		registry := adapter.NewAdapterRegistry()
		mockAdapter1 := &MockProviderAdapter{
			name: "provider1",
			contents: []*domain.Content{
				{ProviderID: "p1_1", Provider: "provider1", Title: "Content 1", Type: domain.ContentTypeVideo, Views: 100, Likes: 10, CreatedAt: time.Now()},
			},
		}
		mockAdapter2 := &MockProviderAdapter{
			name: "provider2",
			err:  errors.New("provider error"),
		}
		registry.Register("provider1", mockAdapter1)
		registry.Register("provider2", mockAdapter2)

		service := NewProviderService(registry, logger)

		contents, err := service.FetchFromAllProviders(context.Background(), "test", nil)

		assert.NoError(t, err)
		assert.Len(t, contents, 1)
		assert.Equal(t, "Content 1", contents[0].Title)
	})

	t.Run("Return error when all providers fail", func(t *testing.T) {
		registry := adapter.NewAdapterRegistry()
		mockAdapter1 := &MockProviderAdapter{
			name: "provider1",
			err:  errors.New("provider1 error"),
		}
		mockAdapter2 := &MockProviderAdapter{
			name: "provider2",
			err:  errors.New("provider2 error"),
		}
		registry.Register("provider1", mockAdapter1)
		registry.Register("provider2", mockAdapter2)

		service := NewProviderService(registry, logger)

		contents, err := service.FetchFromAllProviders(context.Background(), "test", nil)

		assert.Error(t, err)
		assert.Nil(t, contents)
		assert.Contains(t, err.Error(), "all providers failed")
	})

	t.Run("Return empty when no adapters registered", func(t *testing.T) {
		registry := adapter.NewAdapterRegistry()
		service := NewProviderService(registry, logger)

		contents, err := service.FetchFromAllProviders(context.Background(), "test", nil)

		assert.NoError(t, err)
		assert.Len(t, contents, 0)
	})

	t.Run("Concurrent fetching from multiple providers", func(t *testing.T) {
		registry := adapter.NewAdapterRegistry()
		mockAdapter1 := &MockProviderAdapter{
			name:  "provider1",
			delay: 50 * time.Millisecond,
			contents: []*domain.Content{
				{ProviderID: "p1_1", Provider: "provider1", Title: "Content 1", Type: domain.ContentTypeVideo, Views: 100, Likes: 10, CreatedAt: time.Now()},
			},
		}
		mockAdapter2 := &MockProviderAdapter{
			name:  "provider2",
			delay: 50 * time.Millisecond,
			contents: []*domain.Content{
				{ProviderID: "p2_1", Provider: "provider2", Title: "Content 2", Type: domain.ContentTypeText, ReadingTime: 5, Reactions: 25, CreatedAt: time.Now()},
			},
		}
		mockAdapter3 := &MockProviderAdapter{
			name:  "provider3",
			delay: 50 * time.Millisecond,
			contents: []*domain.Content{
				{ProviderID: "p3_1", Provider: "provider3", Title: "Content 3", Type: domain.ContentTypeVideo, Views: 200, Likes: 20, CreatedAt: time.Now()},
			},
		}
		registry.Register("provider1", mockAdapter1)
		registry.Register("provider2", mockAdapter2)
		registry.Register("provider3", mockAdapter3)

		service := NewProviderService(registry, logger)

		start := time.Now()
		contents, err := service.FetchFromAllProviders(context.Background(), "test", nil)
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.Len(t, contents, 3)
		assert.Less(t, duration, 200*time.Millisecond, "Should fetch concurrently")
	})
}

func TestProviderService_getCircuitBreaker(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	registry := adapter.NewAdapterRegistry()
	service := NewProviderService(registry, logger)

	t.Run("Creates circuit breaker for new provider", func(t *testing.T) {
		cb1 := service.getCircuitBreaker("provider1")
		assert.NotNil(t, cb1)

		cb2 := service.getCircuitBreaker("provider1")
		assert.Equal(t, cb1, cb2, "Should return same circuit breaker instance")
	})

	t.Run("Creates separate circuit breakers for different providers", func(t *testing.T) {
		cb1 := service.getCircuitBreaker("provider1")
		cb2 := service.getCircuitBreaker("provider2")

		assert.NotEqual(t, cb1, cb2, "Should create separate circuit breakers")
	})
}
