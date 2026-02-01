-- Create content_types enum
CREATE TYPE content_type AS ENUM ('video', 'text');

-- Create contents table
CREATE TABLE contents (
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
    UNIQUE(provider_id, provider)
);

-- Create indexes for better query performance
CREATE INDEX idx_contents_type ON contents(type);
CREATE INDEX idx_contents_score ON contents(score DESC);
CREATE INDEX idx_contents_created_at ON contents(created_at DESC);
CREATE INDEX idx_contents_provider ON contents(provider);
CREATE INDEX idx_contents_title_search ON contents USING gin(to_tsvector('english', title));

-- Create composite index for common query patterns
CREATE INDEX idx_contents_type_score ON contents(type, score DESC);
CREATE INDEX idx_contents_type_created_at ON contents(type, created_at DESC);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_contents_updated_at BEFORE UPDATE ON contents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
