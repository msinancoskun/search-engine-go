package adapter

import (
	"context"
	"search-engine-go/internal/domain"
)

type ProviderAdapter interface {
	FetchContent(ctx context.Context, query string, contentType *domain.ContentType) ([]*domain.Content, error)

	GetName() string

	GetRateLimit() int
}

type AdapterRegistry struct {
	adapters map[string]ProviderAdapter
}

func NewAdapterRegistry() *AdapterRegistry {
	return &AdapterRegistry{
		adapters: make(map[string]ProviderAdapter),
	}
}

func (r *AdapterRegistry) Register(name string, adapter ProviderAdapter) {
	r.adapters[name] = adapter
}

func (r *AdapterRegistry) Get(name string) (ProviderAdapter, bool) {
	adapter, exists := r.adapters[name]
	return adapter, exists
}

func (r *AdapterRegistry) GetAll() map[string]ProviderAdapter {
	return r.adapters
}

func (r *AdapterRegistry) List() []string {
	names := make([]string, 0, len(r.adapters))
	for name := range r.adapters {
		names = append(names, name)
	}
	return names
}
