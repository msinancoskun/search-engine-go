package service

import (
	"testing"
	"time"

	"search-engine-go/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestScoringService_CalculateScore(t *testing.T) {
	now := time.Now()
	service := NewScoringServiceWithTime(now)

	tests := []struct {
		name     string
		content  *domain.Content
		expected float64
	}{
		{
			name: "Video with high engagement",
			content: &domain.Content{
				Type:      domain.ContentTypeVideo,
				Views:     10000,
				Likes:     500,
				CreatedAt: now.Add(-3 * 24 * time.Hour),
			},
			expected: 28.0,
		},
		{
			name: "Text content",
			content: &domain.Content{
				Type:        domain.ContentTypeText,
				ReadingTime: 10,
				Reactions:   50,
				CreatedAt:   now.Add(-10 * 24 * time.Hour),
			},
			expected: 39.0,
		},
		{
			name: "Old video content",
			content: &domain.Content{
				Type:      domain.ContentTypeVideo,
				Views:     5000,
				Likes:     100,
				CreatedAt: now.Add(-100 * 24 * time.Hour),
			},
			expected: 9.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := service.CalculateScore(tt.content)
			assert.Equal(t, tt.expected, score)
			assert.IsType(t, float64(0), score)
		})
	}
}

func TestContentPopularityScoreSpecification(t *testing.T) {
	spec := domain.NewContentPopularityScoreSpecification()

	t.Run("Video popularity score", func(t *testing.T) {
		content := &domain.Content{
			Type:  domain.ContentTypeVideo,
			Views: 10000,
			Likes: 500,
		}
		score := spec.Calculate(content)
		expected := float64(10000)/1000.0 + float64(500)/100.0
		assert.Equal(t, expected, score)
	})

	t.Run("Text popularity score", func(t *testing.T) {
		content := &domain.Content{
			Type:        domain.ContentTypeText,
			ReadingTime: 10,
			Reactions:   50,
		}
		score := spec.Calculate(content)
		expected := float64(10) + float64(50)/50.0
		assert.Equal(t, expected, score)
	})
}

func TestVideoTypeBoostSpecification(t *testing.T) {
	spec := domain.NewVideoTypeBoostSpecification()

	t.Run("Video gets boost", func(t *testing.T) {
		content := &domain.Content{
			Type:  domain.ContentTypeVideo,
			Views: 10000,
			Likes: 500,
		}
		score := spec.Calculate(content)
		popularityScore := float64(10000)/1000.0 + float64(500)/100.0
		expected := popularityScore * 1.5
		assert.Equal(t, expected, score)
	})

	t.Run("Text gets no boost", func(t *testing.T) {
		content := &domain.Content{
			Type:        domain.ContentTypeText,
			ReadingTime: 10,
			Reactions:   50,
		}
		score := spec.Calculate(content)
		popularityScore := float64(10) + float64(50)/50.0
		expected := popularityScore * 1.0
		assert.Equal(t, expected, score)
	})
}

func TestRecentContentBoostSpecification(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		createdAt time.Time
		expected  float64
	}{
		{
			name:      "1 week old gets maximum boost",
			createdAt: now.Add(-5 * 24 * time.Hour),
			expected:  5.0,
		},
		{
			name:      "2 weeks old gets medium boost",
			createdAt: now.Add(-15 * 24 * time.Hour),
			expected:  3.0,
		},
		{
			name:      "2 months old gets small boost",
			createdAt: now.Add(-60 * 24 * time.Hour),
			expected:  1.0,
		},
		{
			name:      "6 months old gets no boost",
			createdAt: now.Add(-180 * 24 * time.Hour),
			expected:  0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := domain.NewRecentContentBoostSpecification(now)
			content := &domain.Content{CreatedAt: tt.createdAt}
			score := spec.Calculate(content)
			assert.Equal(t, tt.expected, score)
		})
	}
}

func TestContentQualityRatioSpecification(t *testing.T) {
	spec := domain.NewContentQualityRatioSpecification()

	t.Run("Video quality ratio", func(t *testing.T) {
		content := &domain.Content{
			Type:  domain.ContentTypeVideo,
			Views: 1000,
			Likes: 100,
		}
		score := spec.Calculate(content)
		expected := (float64(100) / float64(1000)) * 10.0
		assert.Equal(t, expected, score)
	})

	t.Run("Text quality ratio", func(t *testing.T) {
		content := &domain.Content{
			Type:        domain.ContentTypeText,
			ReadingTime: 10,
			Reactions:   20,
		}
		score := spec.Calculate(content)
		expected := (float64(20) / float64(10)) * 5.0
		assert.Equal(t, expected, score)
	})

	t.Run("Zero views/reading time returns zero", func(t *testing.T) {
		content := &domain.Content{
			Type:  domain.ContentTypeVideo,
			Views: 0,
			Likes: 100,
		}
		score := spec.Calculate(content)
		assert.Equal(t, 0.0, score)
	})
}
