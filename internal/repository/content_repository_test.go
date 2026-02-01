package repository

import (
	"context"
	"testing"
	"time"

	"search-engine-go/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&domain.Content{})
	require.NoError(t, err)

	return db
}

func TestContentRepository_Search(t *testing.T) {
	db := setupTestDB(t)
	repo := NewContentRepository(db)
	ctx := context.Background()

	now := time.Now()
	contents := []*domain.Content{
		{
			ProviderID: "provider1_1",
			Provider:   "provider1",
			Title:      "Test Video 1",
			Type:       domain.ContentTypeVideo,
			Views:      1000,
			Likes:      50,
			Score:      10.5,
			CreatedAt:  now.Add(-1 * time.Hour),
		},
		{
			ProviderID: "provider1_2",
			Provider:   "provider1",
			Title:      "Test Text 1",
			Type:       domain.ContentTypeText,
			ReadingTime: 5,
			Reactions:  25,
			Score:      8.3,
			CreatedAt:  now.Add(-2 * time.Hour),
		},
		{
			ProviderID: "provider2_1",
			Provider:   "provider2",
			Title:      "Another Video",
			Type:       domain.ContentTypeVideo,
			Views:      2000,
			Likes:      100,
			Score:      15.2,
			CreatedAt:  now.Add(-3 * time.Hour),
		},
	}

	for _, content := range contents {
		err := db.Create(content).Error
		require.NoError(t, err)
	}

	t.Run("Search all content", func(t *testing.T) {
		req := &domain.SearchRequest{
			Page:     1,
			PageSize: 10,
			SortBy:   "score",
		}

		results, total, err := repo.Search(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, results, 3)
		assert.Equal(t, "Another Video", results[0].Title)
		assert.Equal(t, "Test Video 1", results[1].Title)
	})

	t.Run("Search with query", func(t *testing.T) {
		req := &domain.SearchRequest{
			Query:    "Test",
			Page:     1,
			PageSize: 10,
			SortBy:   "score",
		}

		results, total, err := repo.Search(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, results, 2)
	})

	t.Run("Search with content type filter", func(t *testing.T) {
		contentType := domain.ContentTypeVideo
		req := &domain.SearchRequest{
			ContentType: &contentType,
			Page:        1,
			PageSize:    10,
			SortBy:      "score",
		}

		results, total, err := repo.Search(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, results, 2)
		for _, result := range results {
			assert.Equal(t, domain.ContentTypeVideo, result.Type)
		}
	})

	t.Run("Search with pagination", func(t *testing.T) {
		req := &domain.SearchRequest{
			Page:     1,
			PageSize: 2,
			SortBy:   "score",
		}

		results, total, err := repo.Search(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, results, 2)
	})

	t.Run("Search with sort by created_at", func(t *testing.T) {
		req := &domain.SearchRequest{
			Page:     1,
			PageSize: 10,
			SortBy:   "created_at",
		}

		results, total, err := repo.Search(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Equal(t, "Test Video 1", results[0].Title)
	})

	t.Run("Search with sort by popularity", func(t *testing.T) {
		req := &domain.SearchRequest{
			Page:     1,
			PageSize: 10,
			SortBy:   "popularity",
		}

		results, total, err := repo.Search(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Equal(t, "Another Video", results[0].Title)
	})

	t.Run("Search with empty results", func(t *testing.T) {
		req := &domain.SearchRequest{
			Query:    "NonExistentContent",
			Page:     1,
			PageSize: 10,
			SortBy:   "score",
		}

		results, total, err := repo.Search(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.Len(t, results, 0)
	})
}

func TestContentRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewContentRepository(db)
	ctx := context.Background()

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

	t.Run("Get existing content by ID", func(t *testing.T) {
		result, err := repo.GetByID(ctx, content.ID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, content.ID, result.ID)
		assert.Equal(t, content.Title, result.Title)
		assert.Equal(t, content.Type, result.Type)
	})

	t.Run("Get non-existing content by ID", func(t *testing.T) {
		result, err := repo.GetByID(ctx, 99999)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestContentRepository_BatchCreateOrUpdate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewContentRepository(db)
	ctx := context.Background()

	t.Run("Batch create new content", func(t *testing.T) {
		contents := []*domain.Content{
			{
				ProviderID: "provider1_new1",
				Provider:   "provider1",
				Title:      "New Content 1",
				Type:       domain.ContentTypeVideo,
				Views:      100,
				Likes:      10,
				Score:      5.0,
			},
			{
				ProviderID: "provider1_new2",
				Provider:   "provider1",
				Title:      "New Content 2",
				Type:       domain.ContentTypeText,
				ReadingTime: 3,
				Reactions:  15,
				Score:      6.0,
			},
		}

		err := repo.BatchCreateOrUpdate(ctx, contents)

		assert.NoError(t, err)

		var count int64
		db.Model(&domain.Content{}).Where("provider_id IN ?", []string{"provider1_new1", "provider1_new2"}).Count(&count)
		assert.Equal(t, int64(2), count)

		assert.NotZero(t, contents[0].ID)
		assert.NotZero(t, contents[1].ID)
	})

	t.Run("Batch update existing content", func(t *testing.T) {
		existing := &domain.Content{
			ProviderID: "provider1_existing",
			Provider:   "provider1",
			Title:      "Original Title",
			Type:       domain.ContentTypeVideo,
			Views:      100,
			Likes:      10,
			Score:      5.0,
		}
		err := db.Create(existing).Error
		require.NoError(t, err)

		updated := []*domain.Content{
			{
				ProviderID: "provider1_existing",
				Provider:   "provider1",
				Title:      "Updated Title",
				Type:       domain.ContentTypeVideo,
				Views:      200,
				Likes:      20,
				Score:      10.0,
			},
		}

		err = repo.BatchCreateOrUpdate(ctx, updated)
		assert.NoError(t, err)

		var result domain.Content
		err = db.Where("provider_id = ?", "provider1_existing").First(&result).Error
		assert.NoError(t, err)
		assert.Equal(t, "Updated Title", result.Title)
		assert.Equal(t, 200, result.Views)
		assert.Equal(t, 20, result.Likes)
		assert.Equal(t, 10.0, result.Score)
		assert.Equal(t, existing.ID, result.ID)
	})

	t.Run("Batch create and update mixed", func(t *testing.T) {
		existing := &domain.Content{
			ProviderID: "provider1_mixed1",
			Provider:   "provider1",
			Title:      "Original",
			Type:       domain.ContentTypeVideo,
			Views:      100,
			Likes:      10,
			Score:      5.0,
		}
		err := db.Create(existing).Error
		require.NoError(t, err)

		contents := []*domain.Content{
			{
				ProviderID: "provider1_mixed1",
				Provider:   "provider1",
				Title:      "Updated",
				Type:       domain.ContentTypeVideo,
				Views:      200,
				Likes:      20,
				Score:      10.0,
			},
			{
				ProviderID: "provider1_mixed2",
				Provider:   "provider1",
				Title:      "New",
				Type:       domain.ContentTypeText,
				ReadingTime: 5,
				Reactions:  10,
				Score:      7.0,
			},
		}

		err = repo.BatchCreateOrUpdate(ctx, contents)
		assert.NoError(t, err)
		
		var updated domain.Content
		err = db.Where("provider_id = ?", "provider1_mixed1").First(&updated).Error
		assert.NoError(t, err)
		assert.Equal(t, "Updated", updated.Title)

		var created domain.Content
		err = db.Where("provider_id = ?", "provider1_mixed2").First(&created).Error
		assert.NoError(t, err)
		assert.Equal(t, "New", created.Title)
	})

	t.Run("Empty batch", func(t *testing.T) {
		err := repo.BatchCreateOrUpdate(ctx, []*domain.Content{})
		assert.NoError(t, err)
	})
}
