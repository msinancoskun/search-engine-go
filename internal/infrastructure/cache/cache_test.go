package cache

import (
	"context"
	"testing"
	"time"

	"search-engine-go/internal/config"
	"search-engine-go/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestInMemoryCache_GetSet(t *testing.T) {
	cache := NewInMemory()
	ctx := context.Background()

	t.Run("Get non-existent key", func(t *testing.T) {
		value, found := cache.Get(ctx, "nonexistent")
		assert.False(t, found)
		assert.Nil(t, value)
	})

	t.Run("Set and Get", func(t *testing.T) {
		contents := []*domain.Content{
			{
				ID:    1,
				Title: "Test Content",
				Type:  domain.ContentTypeVideo,
			},
		}

		err := cache.Set(ctx, "test-key", contents, 5*time.Minute)
		assert.NoError(t, err)

		value, found := cache.Get(ctx, "test-key")
		assert.True(t, found)
		assert.NotNil(t, value)
		assert.Len(t, value, 1)
		assert.Equal(t, "Test Content", value[0].Title)
	})

	t.Run("Set with zero TTL uses default", func(t *testing.T) {
		contents := []*domain.Content{
			{ID: 2, Title: "Default TTL"},
		}

		err := cache.Set(ctx, "default-ttl", contents, 0)
		assert.NoError(t, err)

		value, found := cache.Get(ctx, "default-ttl")
		assert.True(t, found)
		assert.NotNil(t, value)
	})
}

func TestInMemoryCache_TTLExpiration(t *testing.T) {
	cache := NewInMemory()
	ctx := context.Background()

	contents := []*domain.Content{
		{ID: 1, Title: "Expiring Content"},
	}

		err := cache.Set(ctx, "expiring-key", contents, 100*time.Millisecond)
	require.NoError(t, err)

	value, found := cache.Get(ctx, "expiring-key")
	assert.True(t, found)
	assert.NotNil(t, value)

	time.Sleep(150 * time.Millisecond)

	value, found = cache.Get(ctx, "expiring-key")
	assert.False(t, found)
	assert.Nil(t, value)
}

func TestInMemoryCache_Delete(t *testing.T) {
	cache := NewInMemory()
	ctx := context.Background()

	contents := []*domain.Content{
		{ID: 1, Title: "To Delete"},
	}

	err := cache.Set(ctx, "delete-key", contents, 5*time.Minute)
	require.NoError(t, err)

	value, found := cache.Get(ctx, "delete-key")
	assert.True(t, found)
	assert.NotNil(t, value)

	err = cache.Delete(ctx, "delete-key")
	assert.NoError(t, err)

	value, found = cache.Get(ctx, "delete-key")
	assert.False(t, found)
	assert.Nil(t, value)
}

func TestInMemoryCache_Clear(t *testing.T) {
	cache := NewInMemory()
	ctx := context.Background()

	contents1 := []*domain.Content{{ID: 1, Title: "Content 1"}}
	contents2 := []*domain.Content{{ID: 2, Title: "Content 2"}}

	err := cache.Set(ctx, "key1", contents1, 5*time.Minute)
	require.NoError(t, err)
	err = cache.Set(ctx, "key2", contents2, 5*time.Minute)
	require.NoError(t, err)

	_, found1 := cache.Get(ctx, "key1")
	_, found2 := cache.Get(ctx, "key2")
	assert.True(t, found1)
	assert.True(t, found2)

	err = cache.Clear(ctx)
	assert.NoError(t, err)

	_, found1 = cache.Get(ctx, "key1")
	_, found2 = cache.Get(ctx, "key2")
	assert.False(t, found1)
	assert.False(t, found2)
}

func TestInMemoryCache_Eviction(t *testing.T) {
	logger, _ := zap.NewProduction()
	cache := &InMemoryCache{
		data:    make(map[string]cacheItem),
		ttl:     5 * time.Minute,
		log:     logger,
		maxSize: 3,
	}
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		contents := []*domain.Content{{ID: int64(i), Title: "Content"}}
		err := cache.Set(ctx, "key"+string(rune(i)), contents, 5*time.Minute)
		require.NoError(t, err)
	}

	oldestTime := time.Now().Add(-10 * time.Minute)
	cache.data["key0"] = cacheItem{
		value:     []*domain.Content{{ID: 0}},
		expiresAt: oldestTime,
	}

	newContents := []*domain.Content{{ID: 99, Title: "New Content"}}
	err := cache.Set(ctx, "key-new", newContents, 5*time.Minute)
	require.NoError(t, err)

	_, found := cache.Get(ctx, "key0")
	assert.False(t, found, "Oldest key should be evicted")

	value, found := cache.Get(ctx, "key-new")
	assert.True(t, found)
	assert.NotNil(t, value)
}

func TestInMemoryCache_Close(t *testing.T) {
	cache := NewInMemory()
	ctx := context.Background()

	contents := []*domain.Content{{ID: 1, Title: "Content"}}
	err := cache.Set(ctx, "key", contents, 5*time.Minute)
	require.NoError(t, err)

	err = cache.Close()
	assert.NoError(t, err)

	_, found := cache.Get(ctx, "key")
	assert.False(t, found)
}

func TestRedisCache_GetSet(t *testing.T) {
	cfg := config.CacheConfig{
		Host: "localhost",
		Port: 6379,
		DB:   1,
		TTL:  5 * time.Minute,
	}

	cache, err := NewRedis(cfg)
	if err != nil {
		t.Skip("Redis not available, skipping Redis cache tests")
		return
	}
	defer cache.Close()

	ctx := context.Background()

	err = cache.Clear(ctx)
	require.NoError(t, err)

	t.Run("Get non-existent key", func(t *testing.T) {
		value, found := cache.Get(ctx, "nonexistent")
		assert.False(t, found)
		assert.Nil(t, value)
	})

	t.Run("Set and Get", func(t *testing.T) {
		contents := []*domain.Content{
			{
				ID:    1,
				Title: "Test Content",
				Type:  domain.ContentTypeVideo,
			},
		}

		err := cache.Set(ctx, "test-key", contents, 5*time.Minute)
		assert.NoError(t, err)

		value, found := cache.Get(ctx, "test-key")
		assert.True(t, found)
		assert.NotNil(t, value)
		assert.Len(t, value, 1)
		assert.Equal(t, "Test Content", value[0].Title)
	})

	t.Run("Set with zero TTL uses default", func(t *testing.T) {
		contents := []*domain.Content{
			{ID: 2, Title: "Default TTL"},
		}

		err := cache.Set(ctx, "default-ttl", contents, 0)
		assert.NoError(t, err)

		value, found := cache.Get(ctx, "default-ttl")
		assert.True(t, found)
		assert.NotNil(t, value)
	})
}

func TestRedisCache_TTLExpiration(t *testing.T) {
	cfg := config.CacheConfig{
		Host: "localhost",
		Port: 6379,
		DB:   1,
		TTL:  5 * time.Minute,
	}

	cache, err := NewRedis(cfg)
	if err != nil {
		t.Skip("Redis not available, skipping Redis cache tests")
		return
	}
	defer cache.Close()

	ctx := context.Background()

	contents := []*domain.Content{
		{ID: 1, Title: "Expiring Content"},
	}

	err = cache.Set(ctx, "expiring-key", contents, 100*time.Millisecond)
	require.NoError(t, err)

	value, found := cache.Get(ctx, "expiring-key")
	assert.True(t, found)
	assert.NotNil(t, value)

	time.Sleep(150 * time.Millisecond)

	value, found = cache.Get(ctx, "expiring-key")
	assert.False(t, found)
	assert.Nil(t, value)
}

func TestRedisCache_Delete(t *testing.T) {
	cfg := config.CacheConfig{
		Host: "localhost",
		Port: 6379,
		DB:   1,
		TTL:  5 * time.Minute,
	}

	cache, err := NewRedis(cfg)
	if err != nil {
		t.Skip("Redis not available, skipping Redis cache tests")
		return
	}
	defer cache.Close()

	ctx := context.Background()

	contents := []*domain.Content{
		{ID: 1, Title: "To Delete"},
	}

	err = cache.Set(ctx, "delete-key", contents, 5*time.Minute)
	require.NoError(t, err)

	value, found := cache.Get(ctx, "delete-key")
	assert.True(t, found)
	assert.NotNil(t, value)

	err = cache.Delete(ctx, "delete-key")
	assert.NoError(t, err)

	value, found = cache.Get(ctx, "delete-key")
	assert.False(t, found)
	assert.Nil(t, value)
}

func TestRedisCache_Clear(t *testing.T) {
	cfg := config.CacheConfig{
		Host: "localhost",
		Port: 6379,
		DB:   1,
		TTL:  5 * time.Minute,
	}

	cache, err := NewRedis(cfg)
	if err != nil {
		t.Skip("Redis not available, skipping Redis cache tests")
		return
	}
	defer cache.Close()

	ctx := context.Background()

	contents1 := []*domain.Content{{ID: 1, Title: "Content 1"}}
	contents2 := []*domain.Content{{ID: 2, Title: "Content 2"}}

	err = cache.Set(ctx, "key1", contents1, 5*time.Minute)
	require.NoError(t, err)
	err = cache.Set(ctx, "key2", contents2, 5*time.Minute)
	require.NoError(t, err)

	_, found1 := cache.Get(ctx, "key1")
	_, found2 := cache.Get(ctx, "key2")
	assert.True(t, found1)
	assert.True(t, found2)

	err = cache.Clear(ctx)
	assert.NoError(t, err)

	_, found1 = cache.Get(ctx, "key1")
	_, found2 = cache.Get(ctx, "key2")
	assert.False(t, found1)
	assert.False(t, found2)
}
