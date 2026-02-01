package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"search-engine-go/internal/domain"

	"gorm.io/gorm"
)

type ContentRepository struct {
	db *gorm.DB
}

func NewContentRepository(db *gorm.DB) *ContentRepository {
	return &ContentRepository{db: db}
}

func (r *ContentRepository) isRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func (r *ContentRepository) hasError(err error) bool {
	return err != nil
}

func (r *ContentRepository) Search(ctx context.Context, req *domain.SearchRequest) ([]*domain.Content, int, error) {
	offset := (req.Page - 1) * req.PageSize
	query := r.db.WithContext(ctx).Model(&domain.Content{})

	if req.Query != "" {
		if r.isPostgreSQL() {
			query = query.Where("to_tsvector('english', title) @@ plainto_tsquery('english', ?)", req.Query)
		} else {
			query = query.Where("title LIKE ?", "%"+req.Query+"%")
		}
	}

	if req.ContentType != nil {
		query = query.Where("type = ?", *req.ContentType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortOrder := "DESC"
	if req.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	switch req.SortBy {
	case "created_at":
		query = query.Order(fmt.Sprintf("created_at %s", sortOrder))
	case "popularity":
		if sortOrder == "ASC" {
			query = query.Order("views ASC, likes ASC")
		} else {
			query = query.Order("views DESC, likes DESC")
		}
	default:
		query = query.Order(fmt.Sprintf("score %s", sortOrder))
	}

	var contents []*domain.Content
	if err := query.Offset(offset).Limit(req.PageSize).Find(&contents).Error; err != nil {
		return nil, 0, err
	}

	return contents, int(total), nil
}

func (r *ContentRepository) GetByID(ctx context.Context, id int64) (*domain.Content, error) {
	var content domain.Content
	if err := r.db.WithContext(ctx).First(&content, id).Error; err != nil {
		if r.isRecordNotFound(err) {
			return nil, domain.NewNotFoundError("content", id)
		}
		return nil, domain.NewDatabaseError("get_by_id", err)
	}
	return &content, nil
}

func (r *ContentRepository) BatchCreateOrUpdate(ctx context.Context, contents []*domain.Content) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, content := range contents {
			var existing domain.Content
			result := tx.Where("provider_id = ? AND provider = ?", content.ProviderID, content.Provider).
				First(&existing)

			if r.isRecordFound(result.Error) {
				updateData := map[string]interface{}{
					"title":        content.Title,
					"type":         content.Type,
					"views":        content.Views,
					"likes":        content.Likes,
					"reading_time": content.ReadingTime,
					"reactions":    content.Reactions,
					"score":        content.Score,
				}
				if err := tx.Model(&existing).Updates(updateData).Error; err != nil {
					return fmt.Errorf("failed to update content: %w", err)
				}
				content.ID = existing.ID
			} else if r.isRecordNotFound(result.Error) {
				if err := tx.Create(content).Error; err != nil {
					return fmt.Errorf("failed to create content: %w", err)
				}
			} else {
				return fmt.Errorf("failed to check existing content: %w", result.Error)
			}
		}
		return nil
	})
}

func (r *ContentRepository) isRecordFound(err error) bool {
	return err == nil
}

func (r *ContentRepository) isPostgreSQL() bool {
	name := r.db.Dialector.Name()
	return strings.Contains(strings.ToLower(name), "postgres")
}
