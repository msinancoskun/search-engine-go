package adapter

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"search-engine-go/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONProviderAdapter_GetName(t *testing.T) {
	adapter := NewJSONProviderAdapter("test-provider", "http://example.com", 60, 5*time.Second)
	assert.Equal(t, "test-provider", adapter.GetName())
}

func TestJSONProviderAdapter_GetRateLimit(t *testing.T) {
	adapter := NewJSONProviderAdapter("test-provider", "http://example.com", 120, 5*time.Second)
	assert.Equal(t, 120, adapter.GetRateLimit())
}

func TestJSONProviderAdapter_FetchContent_FromFile(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "test.json")
	jsonContent := `{
		"contents": [
			{
				"id": "v1",
				"title": "Test Video",
				"type": "video",
				"metrics": {
					"views": 1000,
					"likes": 50
				},
				"published_at": "2024-03-15T10:00:00Z",
				"tags": ["test"]
			},
			{
				"id": "a1",
				"title": "Test Article",
				"type": "article",
				"metrics": {
					"reading_time": 5,
					"reactions": 25
				},
				"published_at": "2024-03-14T15:30:00Z",
				"tags": ["test"]
			}
		],
		"pagination": {
			"total": 2,
			"page": 1,
			"per_page": 10
		}
	}`
	err := os.WriteFile(jsonFile, []byte(jsonContent), 0644)
	require.NoError(t, err)

	adapter := NewJSONProviderAdapter("test-provider", jsonFile, 60, 5*time.Second)

	contents, err := adapter.FetchContent(context.Background(), "", nil)

	assert.NoError(t, err)
	assert.Len(t, contents, 2)
	assert.Equal(t, "test-provider_v1", contents[0].ProviderID)
	assert.Equal(t, "Test Video", contents[0].Title)
	assert.Equal(t, domain.ContentTypeVideo, contents[0].Type)
	assert.Equal(t, 1000, contents[0].Views)
	assert.Equal(t, 50, contents[0].Likes)

	assert.Equal(t, "test-provider_a1", contents[1].ProviderID)
	assert.Equal(t, "Test Article", contents[1].Title)
	assert.Equal(t, domain.ContentTypeText, contents[1].Type)
	assert.Equal(t, 5, contents[1].ReadingTime)
	assert.Equal(t, 25, contents[1].Reactions)
}

func TestJSONProviderAdapter_FetchContent_WithContentTypeFilter(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "test.json")
	jsonContent := `{
		"contents": [
			{
				"id": "v1",
				"title": "Test Video",
				"type": "video",
				"metrics": {
					"views": 1000,
					"likes": 50
				},
				"published_at": "2024-03-15T10:00:00Z"
			},
			{
				"id": "a1",
				"title": "Test Article",
				"type": "article",
				"metrics": {
					"reading_time": 5,
					"reactions": 25
				},
				"published_at": "2024-03-14T15:30:00Z"
			}
		],
		"pagination": {
			"total": 2,
			"page": 1,
			"per_page": 10
		}
	}`
	err := os.WriteFile(jsonFile, []byte(jsonContent), 0644)
	require.NoError(t, err)

	adapter := NewJSONProviderAdapter("test-provider", jsonFile, 60, 5*time.Second)
	contentType := domain.ContentTypeVideo

	contents, err := adapter.FetchContent(context.Background(), "test", &contentType)

	assert.NoError(t, err)
	assert.Len(t, contents, 2)
}

func TestJSONProviderAdapter_FetchContent_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "invalid.json")
	err := os.WriteFile(jsonFile, []byte("invalid json"), 0644)
	require.NoError(t, err)

	adapter := NewJSONProviderAdapter("test-provider", jsonFile, 60, 5*time.Second)

	contents, err := adapter.FetchContent(context.Background(), "", nil)

	assert.Error(t, err)
	assert.Nil(t, contents)
	assert.Contains(t, err.Error(), "parse JSON")
}

func TestJSONProviderAdapter_FetchContent_FileNotFound(t *testing.T) {
	adapter := NewJSONProviderAdapter("test-provider", "/nonexistent/file.json", 60, 5*time.Second)

	contents, err := adapter.FetchContent(context.Background(), "", nil)

	assert.Error(t, err)
	assert.Nil(t, contents)
	assert.Contains(t, err.Error(), "read mock file")
}

func TestJSONProviderAdapter_convertToDomain(t *testing.T) {
	adapter := NewJSONProviderAdapter("test-provider", "http://example.com", 60, 5*time.Second)

	t.Run("Convert video content", func(t *testing.T) {
		item := JSONContentItem{
			ID:    "v1",
			Title: "Test Video",
			Type:  "video",
			Metrics: Metrics{
				Views: 1000,
				Likes: 50,
			},
			PublishedAt: "2024-03-15T10:00:00Z",
		}

		content := adapter.convertToDomain(item)

		assert.Equal(t, "test-provider_v1", content.ProviderID)
		assert.Equal(t, "test-provider", content.Provider)
		assert.Equal(t, "Test Video", content.Title)
		assert.Equal(t, domain.ContentTypeVideo, content.Type)
		assert.Equal(t, 1000, content.Views)
		assert.Equal(t, 50, content.Likes)
	})

	t.Run("Convert text content", func(t *testing.T) {
		item := JSONContentItem{
			ID:    "a1",
			Title: "Test Article",
			Type:  "article",
			Metrics: Metrics{
				ReadingTime: 5,
				Reactions:   25,
			},
			PublishedAt: "2024-03-14T15:30:00Z",
		}

		content := adapter.convertToDomain(item)

		assert.Equal(t, domain.ContentTypeText, content.Type)
		assert.Equal(t, 5, content.ReadingTime)
		assert.Equal(t, 25, content.Reactions)
	})

	t.Run("Convert with invalid date", func(t *testing.T) {
		item := JSONContentItem{
			ID:    "v1",
			Title: "Test",
			Type:  "video",
			PublishedAt: "invalid-date",
		}

		content := adapter.convertToDomain(item)

		assert.NotZero(t, content.CreatedAt)
	})
}

func TestJSONProviderAdapter_isFilePath(t *testing.T) {
	adapter := NewJSONProviderAdapter("test-provider", "http://example.com", 60, 5*time.Second)

	assert.True(t, adapter.isFilePath("/path/to/file.json"))
	assert.True(t, adapter.isFilePath("relative/path.json"))
	assert.False(t, adapter.isFilePath("http://example.com"))
	assert.False(t, adapter.isFilePath("https://example.com"))
}

func TestJSONProviderAdapter_WithRetry(t *testing.T) {
	adapter := NewJSONProviderAdapterWithRetry("test-provider", "http://example.com", 60, 5*time.Second, 3, 1*time.Second)
	assert.Equal(t, "test-provider", adapter.GetName())
	assert.Equal(t, 60, adapter.GetRateLimit())
}
