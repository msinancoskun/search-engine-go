package service

import (
	"time"

	"search-engine-go/internal/domain"
)

type ScoringService struct {
	specification domain.ScoreSpecification
}

func NewScoringService() *ScoringService {
	return &ScoringService{
		specification: domain.NewContentRelevanceScoreSpecification(time.Now),
	}
}

func NewScoringServiceWithTime(now time.Time) *ScoringService {
	return &ScoringService{
		specification: domain.NewContentRelevanceScoreSpecification(func() time.Time { return now }),
	}
}

func (s *ScoringService) CalculateScore(content *domain.Content) float64 {
	return s.specification.Calculate(content)
}
