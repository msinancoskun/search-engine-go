package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"search-engine-go/internal/config"
	"search-engine-go/internal/domain"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Cache interface {
	Get(ctx context.Context, key string) ([]*domain.Content, bool)
	Set(ctx context.Context, key string, value []*domain.Content, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	Close() error
}

func NewRedis(cfg config.CacheConfig) (Cache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger, _ := zap.NewProduction()
	return &RedisCache{
		client: rdb,
		ttl:    cfg.TTL,
		log:    logger,
	}, nil
}

func NewInMemory() Cache {
	logger, _ := zap.NewProduction()
	return &InMemoryCache{
		data:    make(map[string]cacheItem),
		ttl:     5 * time.Minute,
		log:     logger,
		maxSize: 1000,
	}
}

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
	log    *zap.Logger
}

func (c *RedisCache) Get(ctx context.Context, key string) ([]*domain.Content, bool) {
	data, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false
	}
	if err != nil {
		c.log.Warn("Failed to get from cache", zap.Error(err))
		return nil, false
	}

	var contents []*domain.Content
	if err := json.Unmarshal([]byte(data), &contents); err != nil {
		c.log.Warn("Failed to unmarshal cache data", zap.Error(err))
		return nil, false
	}

	return contents, true
}

func (c *RedisCache) Set(ctx context.Context, key string, value []*domain.Content, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.ttl
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *RedisCache) Clear(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}

type InMemoryCache struct {
	mu      sync.RWMutex
	data    map[string]cacheItem
	ttl     time.Duration
	log     *zap.Logger
	maxSize int
}

type cacheItem struct {
	value     []*domain.Content
	expiresAt time.Time
}

func (c *InMemoryCache) Get(ctx context.Context, key string) ([]*domain.Content, bool) {
	c.mu.RLock()
	item, exists := c.data[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	if c.hasExpiredItem(item) {
		c.mu.Lock()
		delete(c.data, key)
		c.mu.Unlock()
		return nil, false
	}

	return item.value, true
}

func (c *InMemoryCache) hasExpiredItem(item cacheItem) bool {
	return time.Now().After(item.expiresAt)
}

func (c *InMemoryCache) Set(ctx context.Context, key string, value []*domain.Content, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.ttl
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isCacheFullLocked() {
		c.evictOldestEntryLocked()
	}

	c.data[key] = cacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}

	return nil
}

func (c *InMemoryCache) isCacheFull() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data) >= c.maxSize
}

func (c *InMemoryCache) isCacheFullLocked() bool {
	return len(c.data) >= c.maxSize
}

func (c *InMemoryCache) evictOldestEntry() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.evictOldestEntryLocked()
}

func (c *InMemoryCache) evictOldestEntryLocked() {
	var oldestKey string
	var oldestTime time.Time
	for k, v := range c.data {
		if oldestTime.IsZero() || v.expiresAt.Before(oldestTime) {
			oldestKey = k
			oldestTime = v.expiresAt
		}
	}
	if oldestKey != "" {
		delete(c.data, oldestKey)
	}
}

func (c *InMemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
	return nil
}

func (c *InMemoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]cacheItem)
	return nil
}

func (c *InMemoryCache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]cacheItem)
	return nil
}
