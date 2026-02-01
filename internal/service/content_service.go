package service

import (
	"context"
	"fmt"
	"time"

	"search-engine-go/internal/domain"
	"search-engine-go/internal/infrastructure/cache"
	"search-engine-go/internal/repository"

	"go.uber.org/zap"
)

type ContentServiceInterface interface {
	Search(ctx context.Context, req *domain.SearchRequest) (*domain.SearchResponse, error)
	GetByID(ctx context.Context, id int64) (*domain.Content, error)
}

type ContentService struct {
	repo        *repository.ContentRepository
	providerSvc *ProviderService
	scoringSvc  *ScoringService
	cache       cache.Cache
	log         *zap.Logger
}

func NewContentService(
	repo *repository.ContentRepository,
	providerSvc *ProviderService,
	scoringSvc *ScoringService,
	cache cache.Cache,
	log *zap.Logger,
) *ContentService {
	return &ContentService{
		repo:        repo,
		providerSvc: providerSvc,
		scoringSvc:  scoringSvc,
		cache:       cache,
		log:         log,
	}
}

func (s *ContentService) Search(ctx context.Context, req *domain.SearchRequest) (*domain.SearchResponse, error) {
	paginationSpec := domain.NewPaginationSpecification()
	paginationSpec.NormalizePagination(req)

	if req.SortOrder != "asc" && req.SortOrder != "desc" {
		req.SortOrder = "desc"
	}

	cacheKey := s.generateCacheKey(req)

	if cached, found := s.cache.Get(ctx, cacheKey); found {
		s.log.Debug("Cache hit", zap.String("key", cacheKey))
		total := len(cached)
		totalPages := (total + req.PageSize - 1) / req.PageSize

		paginatedCached := s.paginateCachedResults(cached, req.Page, req.PageSize)

		return &domain.SearchResponse{
			Items:      paginatedCached,
			Total:      total,
			Page:       req.Page,
			PageSize:   req.PageSize,
			TotalPages: totalPages,
		}, nil
	}

	allContents, err := s.providerSvc.FetchFromAllProviders(ctx, req.Query, req.ContentType)
	if err != nil {
		s.log.Warn("Failed to fetch from some providers", zap.Error(err))
		if len(allContents) == 0 {
			return nil, domain.NewProviderError("all", "all providers failed", err)
		}
	}

	for _, content := range allContents {
		content.Score = s.scoringSvc.CalculateScore(content)
	}

	if err := s.repo.BatchCreateOrUpdate(ctx, allContents); err != nil {
		s.log.Error("Failed to save content to database", zap.Error(err))
		return nil, domain.NewDatabaseError("batch_create_or_update", err)
	}

	contents, total, err := s.repo.Search(ctx, req)
	if err != nil {
		return nil, domain.NewDatabaseError("search", err)
	}

	if err := s.cache.Set(ctx, cacheKey, contents, 5*time.Minute); err != nil {
		s.log.Warn("Failed to cache results", zap.Error(err))
	}

	totalPages := (total + req.PageSize - 1) / req.PageSize

	return &domain.SearchResponse{
		Items:      contents,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *ContentService) GetByID(ctx context.Context, id int64) (*domain.Content, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ContentService) paginateCachedResults(cached []*domain.Content, page, pageSize int) []*domain.Content {
	start := (page - 1) * pageSize
	end := start + pageSize
	total := len(cached)

	if s.isStartBeyondBounds(start, total) {
		return []*domain.Content{}
	}
	if s.isEndBeyondBounds(end, total) {
		return cached[start:]
	}
	return cached[start:end]
}

func (s *ContentService) isStartBeyondBounds(start, total int) bool {
	return start > total
}

func (s *ContentService) isEndBeyondBounds(end, total int) bool {
	return end > total
}

func (s *ContentService) generateCacheKey(req *domain.SearchRequest) string {
	contentType := "all"
	if req.ContentType != nil {
		contentType = string(*req.ContentType)
	}
	sortOrder := req.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}
	return fmt.Sprintf("search:%s:%s:%s:%s", req.Query, contentType, req.SortBy, sortOrder)
}
