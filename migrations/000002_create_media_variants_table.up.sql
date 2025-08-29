CREATE TABLE IF NOT EXISTS media_variants (
    id VARCHAR(255) PRIMARY KEY,
    media_id VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL, -- thumbnail, converted, etc.
    size VARCHAR(50), -- small, medium, large for thumbnails
    url VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL DEFAULT 0,
    width INTEGER,
    height INTEGER,
    quality INTEGER,
    format VARCHAR(20) NOT NULL, -- webp, jpeg, mp4, etc.
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_media_variants_media_id 
        FOREIGN KEY (media_id) REFERENCES media(id) ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX idx_media_variants_media_id ON media_variants(media_id);
CREATE INDEX idx_media_variants_type ON media_variants(type);
CREATE INDEX idx_media_variants_size ON media_variants(size);
CREATE INDEX idx_media_variants_format ON media_variants(format);
CREATE INDEX idx_media_variants_type_size ON media_variants(type, size);

-- Create unique constraint to prevent duplicate variants
CREATE UNIQUE INDEX idx_media_variants_unique ON media_variants(media_id, type, size, format);
