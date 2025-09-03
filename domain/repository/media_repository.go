package repository

import (
	"context"
	"media-service/domain/entity"
)

type MediaRepository interface {
	Create(ctx context.Context, media *entity.Media) error

	GetByID(ctx context.Context, id string) (*entity.Media, error)

	GetByCreatedBy(ctx context.Context, createdBy string, limit, offset int) ([]*entity.Media, error)

	Update(ctx context.Context, media *entity.Media) error

	Delete(ctx context.Context, id string) error

	List(ctx context.Context, filters MediaFilters) ([]*entity.Media, int, error)

	UpdateProcessingStatus(ctx context.Context, id string, status entity.ProcessingStatus) error

	GetPendingProcessing(ctx context.Context, limit int) ([]*entity.Media, error)
}

type MediaFilters struct {
	CreatedBy string
	Type      entity.MediaType
	MimeType  string
	Limit     int
	Offset    int
	SortBy    string // created_at, name, size
	SortOrder string // asc, desc
}
