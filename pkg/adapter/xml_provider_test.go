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

func TestXMLProviderAdapter_GetName(t *testing.T) {
	adapter := NewXMLProviderAdapter("test-provider", "http://example.com", 60, 5*time.Second)
	assert.Equal(t, "test-provider", adapter.GetName())
}

func TestXMLProviderAdapter_GetRateLimit(t *testing.T) {
	adapter := NewXMLProviderAdapter("test-provider", "http://example.com", 120, 5*time.Second)
	assert.Equal(t, 120, adapter.GetRateLimit())
}

func TestXMLProviderAdapter_FetchContent_FromFile(t *testing.T) {
	tmpDir := t.TempDir()
	xmlFile := filepath.Join(tmpDir, "test.xml")
	xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<feed>
	<items>
		<item>
			<id>v1</id>
			<headline>Test Video</headline>
			<type>video</type>
			<stats>
				<views>1000</views>
				<likes>50</likes>
			</stats>
			<publication_date>2024-03-15</publication_date>
		</item>
		<item>
			<id>a1</id>
			<headline>Test Article</headline>
			<type>article</type>
			<stats>
				<reading_time>5</reading_time>
				<reactions>25</reactions>
			</stats>
			<publication_date>2024-03-14</publication_date>
		</item>
	</items>
	<meta>
		<total_count>2</total_count>
		<current_page>1</current_page>
		<items_per_page>10</items_per_page>
	</meta>
</feed>`
	err := os.WriteFile(xmlFile, []byte(xmlContent), 0644)
	require.NoError(t, err)

	adapter := NewXMLProviderAdapter("test-provider", xmlFile, 60, 5*time.Second)

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

func TestXMLProviderAdapter_FetchContent_WithContentTypeFilter(t *testing.T) {
	tmpDir := t.TempDir()
	xmlFile := filepath.Join(tmpDir, "test.xml")
	xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<feed>
	<items>
		<item>
			<id>v1</id>
			<headline>Test Video</headline>
			<type>video</type>
			<stats>
				<views>1000</views>
				<likes>50</likes>
			</stats>
			<publication_date>2024-03-15</publication_date>
		</item>
		<item>
			<id>a1</id>
			<headline>Test Article</headline>
			<type>article</type>
			<stats>
				<reading_time>5</reading_time>
				<reactions>25</reactions>
			</stats>
			<publication_date>2024-03-14</publication_date>
		</item>
	</items>
	<meta>
		<total_count>2</total_count>
		<current_page>1</current_page>
		<items_per_page>10</items_per_page>
	</meta>
</feed>`
	err := os.WriteFile(xmlFile, []byte(xmlContent), 0644)
	require.NoError(t, err)

	adapter := NewXMLProviderAdapter("test-provider", xmlFile, 60, 5*time.Second)
	contentType := domain.ContentTypeVideo

	contents, err := adapter.FetchContent(context.Background(), "test", &contentType)

	assert.NoError(t, err)
	assert.Len(t, contents, 2)
}

func TestXMLProviderAdapter_FetchContent_InvalidXML(t *testing.T) {
	tmpDir := t.TempDir()
	xmlFile := filepath.Join(tmpDir, "invalid.xml")
	err := os.WriteFile(xmlFile, []byte("invalid xml"), 0644)
	require.NoError(t, err)

	adapter := NewXMLProviderAdapter("test-provider", xmlFile, 60, 5*time.Second)

	contents, err := adapter.FetchContent(context.Background(), "", nil)

	assert.Error(t, err)
	assert.Nil(t, contents)
	assert.Contains(t, err.Error(), "parse XML")
}

func TestXMLProviderAdapter_FetchContent_FileNotFound(t *testing.T) {
	adapter := NewXMLProviderAdapter("test-provider", "/nonexistent/file.xml", 60, 5*time.Second)

	contents, err := adapter.FetchContent(context.Background(), "", nil)

	assert.Error(t, err)
	assert.Nil(t, contents)
	assert.Contains(t, err.Error(), "read mock file")
}

func TestXMLProviderAdapter_convertToDomain(t *testing.T) {
	adapter := NewXMLProviderAdapter("test-provider", "http://example.com", 60, 5*time.Second)

	t.Run("Convert video content", func(t *testing.T) {
		item := XMLContentItem{
			ID:       "v1",
			Headline: "Test Video",
			Type:     "video",
			Stats: XMLStats{
				Views: 1000,
				Likes: 50,
			},
			PublicationDate: "2024-03-15",
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
		item := XMLContentItem{
			ID:       "a1",
			Headline: "Test Article",
			Type:     "article",
			Stats: XMLStats{
				ReadingTime: 5,
				Reactions:   25,
			},
			PublicationDate: "2024-03-14",
		}

		content := adapter.convertToDomain(item)

		assert.Equal(t, domain.ContentTypeText, content.Type)
		assert.Equal(t, 5, content.ReadingTime)
		assert.Equal(t, 25, content.Reactions)
	})

	t.Run("Convert with invalid date", func(t *testing.T) {
		item := XMLContentItem{
			ID:              "v1",
			Headline:        "Test",
			Type:            "video",
			PublicationDate: "invalid-date",
		}

		content := adapter.convertToDomain(item)

		assert.NotZero(t, content.CreatedAt)
	})
}

func TestXMLProviderAdapter_isFilePath(t *testing.T) {
	adapter := NewXMLProviderAdapter("test-provider", "http://example.com", 60, 5*time.Second)

	assert.True(t, adapter.isFilePath("/path/to/file.xml"))
	assert.True(t, adapter.isFilePath("relative/path.xml"))
	assert.False(t, adapter.isFilePath("http://example.com"))
	assert.False(t, adapter.isFilePath("https://example.com"))
}

func TestXMLProviderAdapter_WithRetry(t *testing.T) {
	adapter := NewXMLProviderAdapterWithRetry("test-provider", "http://example.com", 60, 5*time.Second, 3, 1*time.Second)
	assert.Equal(t, "test-provider", adapter.GetName())
	assert.Equal(t, 60, adapter.GetRateLimit())
}
