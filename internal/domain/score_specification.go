package domain

import (
	"time"
)

type ScoreSpecification interface {
	Calculate(content *Content) float64
}

type ContentPopularityScoreSpecification struct{}

func NewContentPopularityScoreSpecification() *ContentPopularityScoreSpecification {
	return &ContentPopularityScoreSpecification{}
}

func (s *ContentPopularityScoreSpecification) Calculate(content *Content) float64 {
	if s.isVideoContent(content) {
		return float64(content.Views)/1000.0 + float64(content.Likes)/100.0
	}
	return float64(content.ReadingTime) + float64(content.Reactions)/50.0
}

func (s *ContentPopularityScoreSpecification) isVideoContent(content *Content) bool {
	return content.Type == ContentTypeVideo
}

type VideoTypeBoostSpecification struct{}

func NewVideoTypeBoostSpecification() *VideoTypeBoostSpecification {
	return &VideoTypeBoostSpecification{}
}

func (s *VideoTypeBoostSpecification) Calculate(content *Content) float64 {
	popularityScore := NewContentPopularityScoreSpecification().Calculate(content)

	if s.isVideoContent(content) {
		return popularityScore * 1.5
	}
	return popularityScore * 1.0
}

func (s *VideoTypeBoostSpecification) isVideoContent(content *Content) bool {
	return content.Type == ContentTypeVideo
}

type RecentContentBoostSpecification struct {
	now time.Time
}

func NewRecentContentBoostSpecification(now time.Time) *RecentContentBoostSpecification {
	return &RecentContentBoostSpecification{now: now}
}

func (s *RecentContentBoostSpecification) Calculate(content *Content) float64 {
	age := s.now.Sub(content.CreatedAt)

	if s.isWithinWeek(age) {
		return 5.0
	} else if s.isWithinMonth(age) {
		return 3.0
	} else if s.isWithinThreeMonths(age) {
		return 1.0
	}
	return 0.0
}

func (s *RecentContentBoostSpecification) isWithinWeek(age time.Duration) bool {
	oneWeek := 7 * 24 * time.Hour
	return age <= oneWeek
}

func (s *RecentContentBoostSpecification) isWithinMonth(age time.Duration) bool {
	oneMonth := 30 * 24 * time.Hour
	return age <= oneMonth
}

func (s *RecentContentBoostSpecification) isWithinThreeMonths(age time.Duration) bool {
	threeMonths := 90 * 24 * time.Hour
	return age <= threeMonths
}

type ContentQualityRatioSpecification struct{}

func NewContentQualityRatioSpecification() *ContentQualityRatioSpecification {
	return &ContentQualityRatioSpecification{}
}

func (s *ContentQualityRatioSpecification) Calculate(content *Content) float64 {
	if s.isVideoContent(content) {
		if s.hasNoViews(content) {
			return 0.0
		}
		return (float64(content.Likes) / float64(content.Views)) * 10.0
	}

	if s.hasNoReadingTime(content) {
		return 0.0
	}
	return (float64(content.Reactions) / float64(content.ReadingTime)) * 5.0
}

func (s *ContentQualityRatioSpecification) isVideoContent(content *Content) bool {
	return content.Type == ContentTypeVideo
}

func (s *ContentQualityRatioSpecification) hasNoViews(content *Content) bool {
	return content.Views == 0
}

func (s *ContentQualityRatioSpecification) hasNoReadingTime(content *Content) bool {
	return content.ReadingTime == 0
}

type CompositeScoreSpecification struct {
	specs []ScoreSpecification
}

func NewCompositeScoreSpecification(specs ...ScoreSpecification) *CompositeScoreSpecification {
	return &CompositeScoreSpecification{specs: specs}
}

func (s *CompositeScoreSpecification) Calculate(content *Content) float64 {
	var totalScore float64

	for _, spec := range s.specs {
		totalScore += spec.Calculate(content)
	}

	return totalScore
}

type ContentRelevanceScoreSpecification struct {
	nowProvider func() time.Time
}

func NewContentRelevanceScoreSpecification(nowProvider func() time.Time) *ContentRelevanceScoreSpecification {
	return &ContentRelevanceScoreSpecification{
		nowProvider: nowProvider,
	}
}

func (s *ContentRelevanceScoreSpecification) Calculate(content *Content) float64 {
	now := s.nowProvider()
	composite := NewCompositeScoreSpecification(
		NewVideoTypeBoostSpecification(),
		NewRecentContentBoostSpecification(now),
		NewContentQualityRatioSpecification(),
	)
	return composite.Calculate(content)
}
