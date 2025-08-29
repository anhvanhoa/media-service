package repository

import (
	"context"
	"media-service/domain/entity"
)

// MediaRepository interface defines methods for media data access
type MediaRepository interface {
	// Create creates a new media record
	Create(ctx context.Context, media *entity.Media) error

	// GetByID retrieves a media record by ID
	GetByID(ctx context.Context, id string) (*entity.Media, error)

	// GetByCreatedBy retrieves media records by creator ID with pagination
	GetByCreatedBy(ctx context.Context, createdBy string, limit, offset int) ([]*entity.Media, error)

	// Update updates a media record
	Update(ctx context.Context, media *entity.Media) error

	// Delete deletes a media record by ID
	Delete(ctx context.Context, id string) error

	// List retrieves media records with pagination and filters
	List(ctx context.Context, filters MediaFilters) ([]*entity.Media, int, error)

	// UpdateProcessingStatus updates the processing status of a media
	UpdateProcessingStatus(ctx context.Context, id string, status entity.ProcessingStatus) error

	// GetPendingProcessing retrieves media records that need processing
	GetPendingProcessing(ctx context.Context, limit int) ([]*entity.Media, error)
}

// MediaVariantRepository interface defines methods for media variant data access
type MediaVariantRepository interface {
	// Create creates a new media variant
	Create(ctx context.Context, variant *entity.MediaVariant) error

	// GetByMediaID retrieves all variants for a media
	GetByMediaID(ctx context.Context, mediaID string) ([]*entity.MediaVariant, error)

	// GetByMediaIDAndType retrieves a specific variant by media ID and type
	GetByMediaIDAndType(ctx context.Context, mediaID, variantType, size string) (*entity.MediaVariant, error)

	// Delete deletes a variant by ID
	Delete(ctx context.Context, id string) error

	// DeleteByMediaID deletes all variants for a media
	DeleteByMediaID(ctx context.Context, mediaID string) error
}

// MediaFilters represents filters for listing media
type MediaFilters struct {
	CreatedBy string
	Type      entity.MediaType
	MimeType  string
	Limit     int
	Offset    int
	SortBy    string // created_at, name, size
	SortOrder string // asc, desc
}
