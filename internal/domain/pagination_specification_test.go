package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaginationSpecification_NormalizePagination(t *testing.T) {
	spec := NewPaginationSpecification()

	t.Run("Normalizes invalid page to default", func(t *testing.T) {
		req := &SearchRequest{
			Page:     0,
			PageSize: 20,
		}

		spec.NormalizePagination(req)

		assert.Equal(t, DefaultPage, req.Page)
		assert.Equal(t, 20, req.PageSize)
	})

	t.Run("Normalizes negative page to default", func(t *testing.T) {
		req := &SearchRequest{
			Page:     -1,
			PageSize: 20,
		}

		spec.NormalizePagination(req)

		assert.Equal(t, DefaultPage, req.Page)
		assert.Equal(t, 20, req.PageSize)
	})

	t.Run("Normalizes invalid page size to default", func(t *testing.T) {
		req := &SearchRequest{
			Page:     1,
			PageSize: 0,
		}

		spec.NormalizePagination(req)

		assert.Equal(t, 1, req.Page)
		assert.Equal(t, DefaultPageSize, req.PageSize)
	})

	t.Run("Normalizes negative page size to default", func(t *testing.T) {
		req := &SearchRequest{
			Page:     1,
			PageSize: -5,
		}

		spec.NormalizePagination(req)

		assert.Equal(t, 1, req.Page)
		assert.Equal(t, DefaultPageSize, req.PageSize)
	})

	t.Run("Caps page size at maximum", func(t *testing.T) {
		req := &SearchRequest{
			Page:     1,
			PageSize: 150,
		}

		spec.NormalizePagination(req)

		assert.Equal(t, 1, req.Page)
		assert.Equal(t, MaxPageSize, req.PageSize)
	})

	t.Run("Normalizes multiple invalid values", func(t *testing.T) {
		req := &SearchRequest{
			Page:     -5,
			PageSize: 200,
		}

		spec.NormalizePagination(req)

		assert.Equal(t, DefaultPage, req.Page)
		assert.Equal(t, MaxPageSize, req.PageSize)
	})

	t.Run("Leaves valid values unchanged", func(t *testing.T) {
		req := &SearchRequest{
			Page:     5,
			PageSize: 50,
		}

		spec.NormalizePagination(req)

		assert.Equal(t, 5, req.Page)
		assert.Equal(t, 50, req.PageSize)
	})

	t.Run("Allows maximum page size", func(t *testing.T) {
		req := &SearchRequest{
			Page:     1,
			PageSize: MaxPageSize,
		}

		spec.NormalizePagination(req)

		assert.Equal(t, 1, req.Page)
		assert.Equal(t, MaxPageSize, req.PageSize)
	})
}

func TestPaginationSpecification_BooleanHelpers(t *testing.T) {
	spec := NewPaginationSpecification()

	t.Run("isInvalidPage returns true for invalid pages", func(t *testing.T) {
		assert.True(t, spec.isInvalidPage(0))
		assert.True(t, spec.isInvalidPage(-1))
		assert.False(t, spec.isInvalidPage(1))
		assert.False(t, spec.isInvalidPage(10))
	})

	t.Run("isInvalidPageSize returns true for invalid page sizes", func(t *testing.T) {
		assert.True(t, spec.isInvalidPageSize(0))
		assert.True(t, spec.isInvalidPageSize(-1))
		assert.False(t, spec.isInvalidPageSize(1))
		assert.False(t, spec.isInvalidPageSize(20))
	})

	t.Run("exceedsMaxPageSize returns true for values exceeding max", func(t *testing.T) {
		assert.True(t, spec.exceedsMaxPageSize(101))
		assert.True(t, spec.exceedsMaxPageSize(200))
		assert.False(t, spec.exceedsMaxPageSize(100))
		assert.False(t, spec.exceedsMaxPageSize(50))
		assert.False(t, spec.exceedsMaxPageSize(1))
	})
}
