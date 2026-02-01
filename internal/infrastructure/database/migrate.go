package database

import (
	"fmt"

	"search-engine-go/internal/domain"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := createEnumType(db); err != nil {
		return fmt.Errorf("failed to create enum type: %w", err)
	}

	var tableExists bool
	if err := db.Raw(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'contents'
		)
	`).Scan(&tableExists).Error; err != nil {
		return fmt.Errorf("failed to check if table exists: %w", err)
	}

	if !tableExists {
		if err := createTable(db); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	} else {
		var deletedAtExists bool
		if err := db.Raw(`
			SELECT EXISTS (
				SELECT FROM information_schema.columns 
				WHERE table_schema = 'public' 
				AND table_name = 'contents' 
				AND column_name = 'deleted_at'
			)
		`).Scan(&deletedAtExists).Error; err != nil {
			return fmt.Errorf("failed to check if deleted_at column exists: %w", err)
		}

		if !deletedAtExists {
			if err := db.Exec(`ALTER TABLE contents ADD COLUMN deleted_at TIMESTAMP`).Error; err != nil {
				return fmt.Errorf("failed to add deleted_at column: %w", err)
			}
		}

		_ = db.AutoMigrate(&domain.Content{})
	}

	if err := createCustomIndexes(db); err != nil {
		return fmt.Errorf("failed to create custom indexes: %w", err)
	}

	return nil
}

func createTable(db *gorm.DB) error {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS contents (
			id BIGSERIAL PRIMARY KEY,
			provider_id VARCHAR(255) NOT NULL,
			provider VARCHAR(100) NOT NULL,
			title VARCHAR(500) NOT NULL,
			type content_type NOT NULL,
			views INTEGER DEFAULT 0,
			likes INTEGER DEFAULT 0,
			reading_time INTEGER DEFAULT 0,
			reactions INTEGER DEFAULT 0,
			score DECIMAL(10, 4) DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP,
			UNIQUE(provider_id, provider)
		)
	`
	if err := db.Exec(createTableSQL).Error; err != nil {
		return fmt.Errorf("failed to create contents table: %w", err)
	}
	return nil
}

func createEnumType(db *gorm.DB) error {
	var exists bool
	if err := db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM pg_type WHERE typname = 'content_type'
		)
	`).Scan(&exists).Error; err != nil {
		return fmt.Errorf("failed to check enum type: %w", err)
	}

	if !exists {
		if err := db.Exec(`
			CREATE TYPE content_type AS ENUM ('video', 'text')
		`).Error; err != nil {
			return fmt.Errorf("failed to create enum type: %w", err)
		}
	}

	return nil
}

func createCustomIndexes(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_contents_provider 
		ON contents(provider)
	`).Error; err != nil {
		return fmt.Errorf("failed to create provider index: %w", err)
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_contents_title_search 
		ON contents USING gin(to_tsvector('english', title))
	`).Error; err != nil {
		return fmt.Errorf("failed to create title search index: %w", err)
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_contents_type_score 
		ON contents(type, score DESC)
	`).Error; err != nil {
		return fmt.Errorf("failed to create type_score index: %w", err)
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_contents_type_created_at 
		ON contents(type, created_at DESC)
	`).Error; err != nil {
		return fmt.Errorf("failed to create type_created_at index: %w", err)
	}

	return nil
}
