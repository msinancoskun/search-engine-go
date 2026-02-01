package domain

import (
	"database/sql/driver"
	"errors"
	"time"

	"gorm.io/gorm"
)

type ContentType string

const (
	ContentTypeVideo ContentType = "video"
	ContentTypeText  ContentType = "text"
)

// Value implements the driver.Valuer interface for ContentType
func (ct ContentType) Value() (driver.Value, error) {
	return string(ct), nil
}

// Scan implements the sql.Scanner interface for ContentType
func (ct *ContentType) Scan(value interface{}) error {
	if value == nil {
		return errors.New("cannot scan nil into ContentType")
	}
	switch v := value.(type) {
	case string:
		*ct = ContentType(v)
	case []byte:
		*ct = ContentType(v)
	default:
		return errors.New("cannot scan non-string value into ContentType")
	}
	return nil
}

type Content struct {
	ID          int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	ProviderID  string         `json:"provider_id" gorm:"type:varchar(255);not null;uniqueIndex:idx_provider_content"`
	Provider    string         `json:"provider" gorm:"type:varchar(100);not null;uniqueIndex:idx_provider_content"`
	Title       string         `json:"title" gorm:"type:varchar(500);not null"`
	Type        ContentType    `json:"type" gorm:"type:content_type;not null;index"`
	Views       int            `json:"views" gorm:"default:0"`
	Likes       int            `json:"likes" gorm:"default:0"`
	ReadingTime int            `json:"reading_time" gorm:"default:0"`
	Reactions   int            `json:"reactions" gorm:"default:0"`
	Score       float64        `json:"score" gorm:"type:decimal(10,4);default:0;index"`
	CreatedAt   time.Time      `json:"created_at" gorm:"index"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Content) TableName() string {
	return "contents"
}

type SearchRequest struct {
	Query       string       `json:"query" form:"query"`
	ContentType *ContentType `json:"content_type,omitempty" form:"content_type"`
	Page        int          `json:"page" form:"page"`
	PageSize    int          `json:"page_size" form:"page_size"`
	SortBy      string       `json:"sort_by" form:"sort_by"`
	SortOrder   string       `json:"sort_order" form:"sort_order"`
}

type SearchResponse struct {
	Items      []*Content `json:"items"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalPages int        `json:"total_pages"`
}
