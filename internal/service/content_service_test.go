package service

import (
	"context"
	"testing"
	"time"

	"search-engine-go/internal/domain"
	"search-engine-go/internal/infrastructure/cache"
	"search-engine-go/internal/repository"
	"search-engine-go/pkg/adapter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type MockAdapter struct {
	name     string
	contents []*domain.Content
	err      error
}

func (m *MockAdapter) GetName() string {
	return m.name
}

func (m *MockAdapter) GetRateLimit() int {
	return 60
}

func (m *MockAdapter) FetchContent(ctx context.Context, query string, contentType *domain.ContentType) ([]*domain.Content, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.contents, nil
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&domain.Content{})
	require.NoError(t, err)

	return db
}

func TestContentService_Search(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	scoringService := NewScoringServiceWithTime(time.Now())

	t.Run("Search with cache hit", func(t *testing.T) {
		db := setupTestDB(t)
		repo := repository.NewContentRepository(db)
		cacheClient := cache.NewInMemory()
		defer cacheClient.Close()

		registry := adapter.NewAdapterRegistry()
		providerSvc := NewProviderService(registry, logger)
		service := NewContentService(repo, providerSvc, scoringService, cacheClient, logger)

		cachedContent := []*domain.Content{
			{ID: 1, Title: "Cached Content", Type: domain.ContentTypeVideo, Score: 10.5},
			{ID: 2, Title: "Another Cached", Type: domain.ContentTypeText, Score: 8.3},
		}
		req := &domain.SearchRequest{
			Query:    "test",
			Page:     1,
			PageSize: 20,
			SortBy:   "score",
		}
		cacheKey := service.generateCacheKey(req)
		err := cacheClient.Set(context.Background(), cacheKey, cachedContent, 5*time.Minute)
		require.NoError(t, err)

		response, err := service.Search(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 2, response.Total)
		assert.Len(t, response.Items, 2)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 20, response.PageSize)
		assert.Equal(t, 1, response.TotalPages)
	})

	t.Run("Search with cache miss - fetches from providers", func(t *testing.T) {
		db := setupTestDB(t)
		repo := repository.NewContentRepository(db)
		cacheClient := cache.NewInMemory()
		defer cacheClient.Close()

		registry := adapter.NewAdapterRegistry()
		now := time.Now()
		mockAdapter := &MockAdapter{
			name: "test-provider",
			contents: []*domain.Content{
				{
					ProviderID: "provider1_1",
					Provider:   "test-provider",
					Title:      "Test Video",
					Type:       domain.ContentTypeVideo,
					Views:      1000,
					Likes:      50,
					CreatedAt:  now.Add(-1 * time.Hour),
				},
				{
					ProviderID: "provider1_2",
					Provider:   "test-provider",
					Title:      "Test Article",
					Type:       domain.ContentTypeText,
					ReadingTime: 5,
					Reactions:  25,
					CreatedAt:  now.Add(-2 * time.Hour),
				},
			},
		}
		registry.Register("test-provider", mockAdapter)

		providerSvc := NewProviderService(registry, logger)
		service := NewContentService(repo, providerSvc, scoringService, cacheClient, logger)

		req := &domain.SearchRequest{
			Query:    "test",
			Page:     1,
			PageSize: 20,
			SortBy:   "score",
		}

		response, err := service.Search(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Greater(t, response.Total, 0)
		assert.Greater(t, len(response.Items), 0)

		var savedContent []*domain.Content
		db.Find(&savedContent)
		assert.Greater(t, len(savedContent), 0)

		for _, item := range response.Items {
			assert.Greater(t, item.Score, 0.0)
		}
	})

	t.Run("Search with pagination", func(t *testing.T) {
		db := setupTestDB(t)
		repo := repository.NewContentRepository(db)
		cacheClient := cache.NewInMemory()
		defer cacheClient.Close()

		registry := adapter.NewAdapterRegistry()
		now := time.Now()
		mockAdapter := &MockAdapter{
			name: "pagination-provider",
			contents: []*domain.Content{
				{ProviderID: "p1_1", Provider: "pagination-provider", Title: "Item 1", Type: domain.ContentTypeVideo, Views: 100, Likes: 10, CreatedAt: now},
				{ProviderID: "p1_2", Provider: "pagination-provider", Title: "Item 2", Type: domain.ContentTypeVideo, Views: 200, Likes: 20, CreatedAt: now},
				{ProviderID: "p1_3", Provider: "pagination-provider", Title: "Item 3", Type: domain.ContentTypeVideo, Views: 300, Likes: 30, CreatedAt: now},
			},
		}
		registry.Register("pagination-provider", mockAdapter)

		providerSvc := NewProviderService(registry, logger)
		service := NewContentService(repo, providerSvc, scoringService, cacheClient, logger)

		req := &domain.SearchRequest{
			Page:     1,
			PageSize: 2,
			SortBy:   "score",
		}

		response, err := service.Search(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, 3, response.Total)
		assert.Len(t, response.Items, 2)
		assert.Equal(t, 2, response.TotalPages)
	})

	t.Run("Search with content type filter", func(t *testing.T) {
		db := setupTestDB(t)
		repo := repository.NewContentRepository(db)
		cacheClient := cache.NewInMemory()
		defer cacheClient.Close()

		registry := adapter.NewAdapterRegistry()
		contentType := domain.ContentTypeVideo
		mockAdapter := &MockAdapter{
			name: "filter-provider",
			contents: []*domain.Content{
				{ProviderID: "p1_1", Provider: "filter-provider", Title: "Video 1", Type: domain.ContentTypeVideo, Views: 100, Likes: 10, CreatedAt: time.Now()},
				{ProviderID: "p1_2", Provider: "filter-provider", Title: "Article 1", Type: domain.ContentTypeText, ReadingTime: 5, Reactions: 25, CreatedAt: time.Now()},
			},
		}
		registry.Register("filter-provider", mockAdapter)

		providerSvc := NewProviderService(registry, logger)
		service := NewContentService(repo, providerSvc, scoringService, cacheClient, logger)

		req := &domain.SearchRequest{
			Query:       "filter",
			ContentType: &contentType,
			Page:        1,
			PageSize:    20,
			SortBy:      "score",
		}

		response, err := service.Search(context.Background(), req)

		assert.NoError(t, err)
		assert.Greater(t, response.Total, 0)
		for _, item := range response.Items {
			assert.Equal(t, domain.ContentTypeVideo, item.Type)
		}
		assert.Equal(t, 1, response.Total)
	})

	t.Run("Search when all providers fail", func(t *testing.T) {
		db := setupTestDB(t)
		repo := repository.NewContentRepository(db)
		cacheClient := cache.NewInMemory()
		defer cacheClient.Close()

		registry := adapter.NewAdapterRegistry()
		mockAdapter := &MockAdapter{
			name: "failing-provider",
			err:  assert.AnError,
		}
		registry.Register("failing-provider", mockAdapter)

		providerSvc := NewProviderService(registry, logger)
		service := NewContentService(repo, providerSvc, scoringService, cacheClient, logger)

		req := &domain.SearchRequest{
			Query:    "failing-query",
			Page:     1,
			PageSize: 20,
			SortBy:   "score",
		}

		response, err := service.Search(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "provider")
	})

	t.Run("Search normalizes pagination", func(t *testing.T) {
		db := setupTestDB(t)
		repo := repository.NewContentRepository(db)
		cacheClient := cache.NewInMemory()
		defer cacheClient.Close()

		registry := adapter.NewAdapterRegistry()
		providerSvc := NewProviderService(registry, logger)
		service := NewContentService(repo, providerSvc, scoringService, cacheClient, logger)

		req := &domain.SearchRequest{
			Page:     0,
			PageSize: 0,
			SortBy:   "score",
		}

		response, err := service.Search(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 20, response.PageSize)
	})
}

func TestContentService_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewContentRepository(db)
	cacheClient := cache.NewInMemory()
	defer cacheClient.Close()

	logger, _ := zap.NewDevelopment()
	registry := adapter.NewAdapterRegistry()
	providerSvc := NewProviderService(registry, logger)
	scoringService := NewScoringService()
	service := NewContentService(repo, providerSvc, scoringService, cacheClient, logger)

	t.Run("Get existing content by ID", func(t *testing.T) {
		content := &domain.Content{
			ProviderID: "provider1_1",
			Provider:   "provider1",
			Title:      "Test Content",
			Type:       domain.ContentTypeVideo,
			Views:      1000,
			Likes:      50,
			Score:      10.5,
			CreatedAt:  time.Now(),
		}

		err := db.Create(content).Error
		require.NoError(t, err)

		result, err := service.GetByID(context.Background(), content.ID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, content.ID, result.ID)
		assert.Equal(t, content.Title, result.Title)
	})

	t.Run("Get non-existing content by ID", func(t *testing.T) {
		result, err := service.GetByID(context.Background(), 99999)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestContentService_paginateCachedResults(t *testing.T) {
	cacheClient := cache.NewInMemory()
	defer cacheClient.Close()

	logger, _ := zap.NewDevelopment()
	registry := adapter.NewAdapterRegistry()
	providerSvc := NewProviderService(registry, logger)
	scoringService := NewScoringService()
	db := setupTestDB(t)
	repo := repository.NewContentRepository(db)
	service := NewContentService(repo, providerSvc, scoringService, cacheClient, logger)

	t.Run("Paginate within bounds", func(t *testing.T) {
		cached := []*domain.Content{
			{ID: 1, Title: "Item 1"},
			{ID: 2, Title: "Item 2"},
			{ID: 3, Title: "Item 3"},
		}

		result := service.paginateCachedResults(cached, 1, 2)

		assert.Len(t, result, 2)
		assert.Equal(t, "Item 1", result[0].Title)
		assert.Equal(t, "Item 2", result[1].Title)
	})

	t.Run("Paginate beyond bounds", func(t *testing.T) {
		cached := []*domain.Content{
			{ID: 1, Title: "Item 1"},
			{ID: 2, Title: "Item 2"},
		}

		result := service.paginateCachedResults(cached, 2, 2)

		assert.Len(t, result, 0)
	})

	t.Run("Paginate partial page", func(t *testing.T) {
		cached := []*domain.Content{
			{ID: 1, Title: "Item 1"},
			{ID: 2, Title: "Item 2"},
			{ID: 3, Title: "Item 3"},
		}

		result := service.paginateCachedResults(cached, 2, 2)

		assert.Len(t, result, 1)
		assert.Equal(t, "Item 3", result[0].Title)
	})
}
