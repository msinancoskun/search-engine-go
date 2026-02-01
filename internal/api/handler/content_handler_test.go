package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"search-engine-go/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockContentService struct {
	mock.Mock
}

func (m *MockContentService) Search(ctx context.Context, req *domain.SearchRequest) (*domain.SearchResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SearchResponse), args.Error(1)
}

func (m *MockContentService) GetByID(ctx context.Context, id int64) (*domain.Content, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Content), args.Error(1)
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

func TestContentHandler_Search(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("Successful search request", func(t *testing.T) {
		mockService := new(MockContentService)
		handler := NewContentHandler(mockService, logger)

		expectedResponse := &domain.SearchResponse{
			Items: []*domain.Content{
				{
					ID:    1,
					Title: "Test Video",
					Type:  domain.ContentTypeVideo,
				},
			},
			Total:      1,
			Page:       1,
			PageSize:   20,
			TotalPages: 1,
		}

		mockService.On("Search", mock.Anything, mock.MatchedBy(func(req *domain.SearchRequest) bool {
			return req.Query == "test" && req.Page == 1 && req.PageSize == 20
		})).Return(expectedResponse, nil)

		router := setupTestRouter(handler)
		req := httptest.NewRequest("GET", "/api/v1/search?query=test&page=1&page_size=20", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response domain.SearchResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse.Total, response.Total)
		assert.Equal(t, len(expectedResponse.Items), len(response.Items))

		mockService.AssertExpectations(t)
	})

	t.Run("Search with content type filter", func(t *testing.T) {
		mockService := new(MockContentService)
		handler := NewContentHandler(mockService, logger)

		contentType := domain.ContentTypeVideo
		expectedResponse := &domain.SearchResponse{
			Items:      []*domain.Content{},
			Total:      0,
			Page:       1,
			PageSize:   20,
			TotalPages: 0,
		}

		mockService.On("Search", mock.Anything, mock.MatchedBy(func(req *domain.SearchRequest) bool {
			return req.ContentType != nil && *req.ContentType == contentType
		})).Return(expectedResponse, nil)

		router := setupTestRouter(handler)
		req := httptest.NewRequest("GET", "/api/v1/search?query=test&content_type=video", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid request parameters", func(t *testing.T) {
		mockService := new(MockContentService)
		handler := NewContentHandler(mockService, logger)

		router := setupTestRouter(handler)
		req := httptest.NewRequest("GET", "/api/v1/search?page=invalid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"].(string), "invalid input for field")
		assert.Equal(t, "INVALID_INPUT", response["code"])

		mockService.AssertNotCalled(t, "Search")
	})

	t.Run("Service error", func(t *testing.T) {
		mockService := new(MockContentService)
		handler := NewContentHandler(mockService, logger)

		mockService.On("Search", mock.Anything, mock.Anything).Return(nil, assert.AnError)

		router := setupTestRouter(handler)
		req := httptest.NewRequest("GET", "/api/v1/search?query=test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Internal server error", response["error"])

		mockService.AssertExpectations(t)
	})

	t.Run("Pagination normalization", func(t *testing.T) {
		mockService := new(MockContentService)
		handler := NewContentHandler(mockService, logger)

		expectedResponse := &domain.SearchResponse{
			Items:      []*domain.Content{},
			Total:      0,
			Page:       1,
			PageSize:   20,
			TotalPages: 0,
		}

		mockService.On("Search", mock.Anything, mock.MatchedBy(func(req *domain.SearchRequest) bool {
			return req.Page == 1 && req.PageSize == 20
		})).Return(expectedResponse, nil)

		router := setupTestRouter(handler)
		req := httptest.NewRequest("GET", "/api/v1/search?query=test&page=0&page_size=0", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestContentHandler_GetByID(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("Successful get by ID", func(t *testing.T) {
		mockService := new(MockContentService)
		handler := NewContentHandler(mockService, logger)

		expectedContent := &domain.Content{
			ID:    1,
			Title: "Test Video",
			Type:  domain.ContentTypeVideo,
			Views: 1000,
			Likes: 50,
		}

		mockService.On("GetByID", mock.Anything, int64(1)).Return(expectedContent, nil)

		router := setupTestRouter(handler)
		req := httptest.NewRequest("GET", "/api/v1/content/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var content domain.Content
		err := json.Unmarshal(w.Body.Bytes(), &content)
		assert.NoError(t, err)
		assert.Equal(t, expectedContent.ID, content.ID)
		assert.Equal(t, expectedContent.Title, content.Title)

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid ID format", func(t *testing.T) {
		mockService := new(MockContentService)
		handler := NewContentHandler(mockService, logger)

		router := setupTestRouter(handler)
		req := httptest.NewRequest("GET", "/api/v1/content/invalid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"].(string), "invalid input for field 'id'")
		assert.Equal(t, "INVALID_INPUT", response["code"])

		mockService.AssertNotCalled(t, "GetByID")
	})

	t.Run("Content not found", func(t *testing.T) {
		mockService := new(MockContentService)
		handler := NewContentHandler(mockService, logger)

		notFoundErr := domain.NewNotFoundError("content", int64(999))
		mockService.On("GetByID", mock.Anything, int64(999)).Return(nil, notFoundErr)

		router := setupTestRouter(handler)
		req := httptest.NewRequest("GET", "/api/v1/content/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"].(string), "not found")
		assert.Equal(t, "NOT_FOUND", response["code"])

		mockService.AssertExpectations(t)
	})

	t.Run("Large ID value", func(t *testing.T) {
		mockService := new(MockContentService)
		handler := NewContentHandler(mockService, logger)

		expectedContent := &domain.Content{
			ID:    9223372036854775807,
			Title: "Test Content",
		}

		mockService.On("GetByID", mock.Anything, int64(9223372036854775807)).Return(expectedContent, nil)

		router := setupTestRouter(handler)
		req := httptest.NewRequest("GET", "/api/v1/content/9223372036854775807", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}
