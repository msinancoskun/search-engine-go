-- Drop trigger
DROP TRIGGER IF EXISTS update_contents_updated_at ON contents;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_contents_type_score;
DROP INDEX IF EXISTS idx_contents_type_created_at;
DROP INDEX IF EXISTS idx_contents_title_search;
DROP INDEX IF EXISTS idx_contents_provider;
DROP INDEX IF EXISTS idx_contents_created_at;
DROP INDEX IF EXISTS idx_contents_score;
DROP INDEX IF EXISTS idx_contents_type;

-- Drop table
DROP TABLE IF EXISTS contents;

-- Drop enum
DROP TYPE IF EXISTS content_type;
