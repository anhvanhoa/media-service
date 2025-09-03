CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS media (
    id VARCHAR(255) PRIMARY KEY,
    created_by VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    size BIGINT NOT NULL DEFAULT 0,
    url VARCHAR(500) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL,
    width INTEGER,
    height INTEGER,
    duration INTEGER,
    processing_status VARCHAR(50) NOT NULL DEFAULT 'pending',
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX idx_media_created_by ON media(created_by);
CREATE INDEX idx_media_type ON media(type);
CREATE INDEX idx_media_processing_status ON media(processing_status);
CREATE INDEX idx_media_created_at ON media(created_at DESC);
CREATE INDEX idx_media_mime_type ON media(mime_type);

CREATE TRIGGER update_media_updated_at BEFORE UPDATE ON media
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
