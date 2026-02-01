package adapter

import (
	"context"
	"testing"
	"time"

	"search-engine-go/internal/domain"

	"github.com/stretchr/testify/assert"
)

type MockAdapter struct {
	name string
}

func (m *MockAdapter) GetName() string {
	return m.name
}

func (m *MockAdapter) GetRateLimit() int {
	return 60
}

func (m *MockAdapter) FetchContent(ctx context.Context, query string, contentType *domain.ContentType) ([]*domain.Content, error) {
	return []*domain.Content{
		{ProviderID: "test_1", Provider: m.name, Title: "Test Content"},
	}, nil
}

func TestNewAdapterRegistry(t *testing.T) {
	registry := NewAdapterRegistry()
	assert.NotNil(t, registry)
	assert.NotNil(t, registry.adapters)
}

func TestAdapterRegistry_Register(t *testing.T) {
	registry := NewAdapterRegistry()
	adapter1 := &MockAdapter{name: "provider1"}
	adapter2 := &MockAdapter{name: "provider2"}

	registry.Register("provider1", adapter1)
	registry.Register("provider2", adapter2)

	all := registry.GetAll()
	assert.Len(t, all, 2)
	assert.Equal(t, adapter1, all["provider1"])
	assert.Equal(t, adapter2, all["provider2"])
}

func TestAdapterRegistry_Get(t *testing.T) {
	registry := NewAdapterRegistry()
	adapter1 := &MockAdapter{name: "provider1"}
	registry.Register("provider1", adapter1)

	t.Run("Get existing adapter", func(t *testing.T) {
		adapter, exists := registry.Get("provider1")
		assert.True(t, exists)
		assert.Equal(t, adapter1, adapter)
	})

	t.Run("Get non-existing adapter", func(t *testing.T) {
		adapter, exists := registry.Get("nonexistent")
		assert.False(t, exists)
		assert.Nil(t, adapter)
	})
}

func TestAdapterRegistry_GetAll(t *testing.T) {
	registry := NewAdapterRegistry()
	adapter1 := &MockAdapter{name: "provider1"}
	adapter2 := &MockAdapter{name: "provider2"}

	registry.Register("provider1", adapter1)
	registry.Register("provider2", adapter2)

	all := registry.GetAll()
	assert.Len(t, all, 2)
	assert.Contains(t, all, "provider1")
	assert.Contains(t, all, "provider2")
}

func TestAdapterRegistry_List(t *testing.T) {
	registry := NewAdapterRegistry()
	adapter1 := &MockAdapter{name: "provider1"}
	adapter2 := &MockAdapter{name: "provider2"}

	registry.Register("provider1", adapter1)
	registry.Register("provider2", adapter2)

	names := registry.List()
	assert.Len(t, names, 2)
	assert.Contains(t, names, "provider1")
	assert.Contains(t, names, "provider2")
}

func TestAdapterRegistry_Overwrite(t *testing.T) {
	registry := NewAdapterRegistry()
	adapter1 := &MockAdapter{name: "provider1"}
	adapter2 := &MockAdapter{name: "provider1"}

	registry.Register("provider1", adapter1)
	registry.Register("provider1", adapter2) // Overwrite

	adapter, exists := registry.Get("provider1")
	assert.True(t, exists)
	assert.Equal(t, adapter2, adapter)
}

func TestAdapterRegistry_Integration(t *testing.T) {
	registry := NewAdapterRegistry()

	jsonAdapter := NewJSONProviderAdapter("json-provider", "http://example.com/json", 60, 5*time.Second)
	registry.Register("json-provider", jsonAdapter)

	xmlAdapter := NewXMLProviderAdapter("xml-provider", "http://example.com/xml", 60, 5*time.Second)
	registry.Register("xml-provider", xmlAdapter)

	json, exists := registry.Get("json-provider")
	assert.True(t, exists)
	assert.Equal(t, "json-provider", json.GetName())

	xml, exists := registry.Get("xml-provider")
	assert.True(t, exists)
	assert.Equal(t, "xml-provider", xml.GetName())

	names := registry.List()
	assert.Len(t, names, 2)
	assert.Contains(t, names, "json-provider")
	assert.Contains(t, names, "xml-provider")

	all := registry.GetAll()
	assert.Len(t, all, 2)
}
