//go:build integration

package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"search-engine-go/internal/domain"
	"search-engine-go/internal/infrastructure/cache"
	"search-engine-go/internal/repository"
	"search-engine-go/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupIntegrationTest(t *testing.T) (*ContentHandler, func()) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&domain.Content{})
	require.NoError(t, err)

	logger, _ := zap.NewDevelopment()

	cacheClient := cache.NewInMemory()

	contentRepo := repository.NewContentRepository(db)

	providerService := &MockProviderService{
		contents: []*domain.Content{
			{
				ProviderID: "provider1_1",
				Provider:   "provider1",
				Title:      "Integration Test Video",
				Type:       domain.ContentTypeVideo,
				Views:      1000,
				Likes:      50,
				CreatedAt:  time.Now(),
			},
			{
				ProviderID: "provider1_2",
				Provider:   "provider1",
				Title:      "Integration Test Text",
				Type:       domain.ContentTypeText,
				ReadingTime: 5,
				Reactions:  25,
				CreatedAt:  time.Now(),
			},
		},
	}

	scoringService := service.NewScoringService()
	contentService := service.NewContentService(contentRepo, providerService, scoringService, cacheClient, logger)

	handler := NewContentHandler(contentService, logger)

	cleanup := func() {
		cacheClient.Close()
	}

	return handler, cleanup
}

type MockProviderService struct {
	contents []*domain.Content
}

func (m *MockProviderService) FetchFromAllProviders(ctx context.Context, query string, contentType *domain.ContentType) ([]*domain.Content, error) {
	var results []*domain.Content
	for _, content := range m.contents {
		if query != "" && !contains(content.Title, query) {
			continue
		}
		if contentType != nil && content.Type != *contentType {
			continue
		}
		copy := *content
		results = append(results, &copy)
	}
	return results, nil
}

func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func setupTestRouter(handler *ContentHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	{
		v1.GET("/search", handler.Search)
		v1.GET("/content/:id", handler.GetByID)
	}
	return router
}

func TestContentHandler_Integration_Search(t *testing.T) {
	handler, cleanup := setupIntegrationTest(t)
	defer cleanup()

	router := setupTestRouter(handler)

	t.Run("End-to-end search flow", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/search?query=Integration&page=1&page_size=10", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response domain.SearchResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Greater(t, response.Total, 0)
		assert.Greater(t, len(response.Items), 0)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 10, response.PageSize)
		assert.Greater(t, response.TotalPages, 0)

		for _, item := range response.Items {
			assert.Greater(t, item.Score, 0.0)
		}
	})

	t.Run("Search with content type filter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/search?query=Integration&content_type=video&page=1&page_size=10", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response domain.SearchResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		for _, item := range response.Items {
			assert.Equal(t, domain.ContentTypeVideo, item.Type)
		}
	})

	t.Run("Search uses cache on second request", func(t *testing.T) {
		req1 := httptest.NewRequest("GET", "/api/v1/search?query=Test&page=1&page_size=10", nil)
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)

		assert.Equal(t, http.StatusOK, w1.Code)

		var response1 domain.SearchResponse
		err := json.Unmarshal(w1.Body.Bytes(), &response1)
		require.NoError(t, err)

		req2 := httptest.NewRequest("GET", "/api/v1/search?query=Test&page=1&page_size=10", nil)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusOK, w2.Code)

		var response2 domain.SearchResponse
		err = json.Unmarshal(w2.Body.Bytes(), &response2)
		require.NoError(t, err)

		assert.Equal(t, response1.Total, response2.Total)
		assert.Equal(t, len(response1.Items), len(response2.Items))
	})

	t.Run("Pagination works correctly", func(t *testing.T) {
		req1 := httptest.NewRequest("GET", "/api/v1/search?query=Integration&page=1&page_size=1", nil)
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)

		assert.Equal(t, http.StatusOK, w1.Code)

		var response1 domain.SearchResponse
		err := json.Unmarshal(w1.Body.Bytes(), &response1)
		require.NoError(t, err)

		assert.Equal(t, 1, response1.Page)
		assert.Equal(t, 1, response1.PageSize)
		assert.Len(t, response1.Items, 1)

		req2 := httptest.NewRequest("GET", "/api/v1/search?query=Integration&page=2&page_size=1", nil)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusOK, w2.Code)

		var response2 domain.SearchResponse
		err = json.Unmarshal(w2.Body.Bytes(), &response2)
		require.NoError(t, err)

		assert.Equal(t, 2, response2.Page)
		assert.Equal(t, 1, response2.PageSize)
		if len(response2.Items) > 0 && len(response1.Items) > 0 {
			assert.NotEqual(t, response1.Items[0].ID, response2.Items[0].ID)
		}
	})
}

func TestContentHandler_Integration_GetByID(t *testing.T) {
	handler, cleanup := setupIntegrationTest(t)
	defer cleanup()

	router := setupTestRouter(handler)

	req := httptest.NewRequest("GET", "/api/v1/search?query=Integration&page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var searchResponse domain.SearchResponse
	err := json.Unmarshal(w.Body.Bytes(), &searchResponse)
	require.NoError(t, err)
	require.Greater(t, len(searchResponse.Items), 0)

	contentID := searchResponse.Items[0].ID

	t.Run("Get content by ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/content/"+fmt.Sprintf("%d", contentID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var content domain.Content
		err := json.Unmarshal(w.Body.Bytes(), &content)
		require.NoError(t, err)

		assert.Equal(t, contentID, content.ID)
		assert.NotEmpty(t, content.Title)
	})
}
