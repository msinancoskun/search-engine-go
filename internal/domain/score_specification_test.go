package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestContentRelevanceScoreSpecification_Calculate(t *testing.T) {
	now := time.Now()
	spec := NewContentRelevanceScoreSpecification(func() time.Time { return now })

	t.Run("Complete video relevance score calculation", func(t *testing.T) {
		content := &Content{
			Type:      ContentTypeVideo,
			Views:     10000,
			Likes:     500,
			CreatedAt: now.Add(-3 * 24 * time.Hour),
		}

		score := spec.Calculate(content)

		expected := 22.5 + 5.0 + 0.5
		assert.Equal(t, expected, score)
	})

	t.Run("Complete text relevance score calculation", func(t *testing.T) {
		content := &Content{
			Type:        ContentTypeText,
			ReadingTime: 10,
			Reactions:   50,
			CreatedAt:   now.Add(-15 * 24 * time.Hour),
		}

		score := spec.Calculate(content)

		expected := 11.0 + 3.0 + 25.0
		assert.Equal(t, expected, score)
	})
}

func TestCompositeScoreSpecification(t *testing.T) {
	now := time.Now()

	t.Run("Combines multiple specifications", func(t *testing.T) {
		popularitySpec := NewContentPopularityScoreSpecification()
		recencySpec := NewRecentContentBoostSpecification(now)
		qualitySpec := NewContentQualityRatioSpecification()

		composite := NewCompositeScoreSpecification(
			popularitySpec,
			recencySpec,
			qualitySpec,
		)

		content := &Content{
			Type:        ContentTypeText,
			ReadingTime: 5,
			Reactions:   25,
			CreatedAt:   now.Add(-2 * 24 * time.Hour),
		}

		score := composite.Calculate(content)

		popularityScore := popularitySpec.Calculate(content)
		recencyScore := recencySpec.Calculate(content)
		qualityScore := qualitySpec.Calculate(content)
		expected := popularityScore + recencyScore + qualityScore

		assert.Equal(t, expected, score)
	})
}
