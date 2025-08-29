package repo

import (
	"context"
	"fmt"
	"media-service/domain/entity"
	"media-service/domain/repository"

	"github.com/go-pg/pg/v10"
)

type mediaRepository struct {
	db *pg.DB
}

// NewMediaRepository creates a new media repository
func NewMediaRepository(db *pg.DB) repository.MediaRepository {
	return &mediaRepository{db: db}
}

func (r *mediaRepository) Create(ctx context.Context, media *entity.Media) error {
	_, err := r.db.ModelContext(ctx, media).Insert()
	return err
}

func (r *mediaRepository) GetByID(ctx context.Context, id string) (*entity.Media, error) {
	media := &entity.Media{}
	err := r.db.ModelContext(ctx, media).Where("id = ?", id).Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return media, nil
}

func (r *mediaRepository) GetByCreatedBy(ctx context.Context, createdBy string, limit, offset int) ([]*entity.Media, error) {
	var media []*entity.Media
	err := r.db.ModelContext(ctx, &media).
		Where("created_by = ?", createdBy).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Select()
	return media, err
}

func (r *mediaRepository) Update(ctx context.Context, media *entity.Media) error {
	_, err := r.db.ModelContext(ctx, media).WherePK().Update()
	return err
}

func (r *mediaRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ModelContext(ctx, (*entity.Media)(nil)).Where("id = ?", id).Delete()
	return err
}

func (r *mediaRepository) List(ctx context.Context, filters repository.MediaFilters) ([]*entity.Media, int, error) {
	var media []*entity.Media
	query := r.db.ModelContext(ctx, &media)

	// Apply filters
	if filters.CreatedBy != "" {
		query = query.Where("created_by = ?", filters.CreatedBy)
	}
	if filters.Type != "" {
		query = query.Where("type = ?", filters.Type)
	}
	if filters.MimeType != "" {
		query = query.Where("mime_type = ?", filters.MimeType)
	}

	// Apply sorting
	sortBy := "created_at"
	if filters.SortBy != "" {
		sortBy = filters.SortBy
	}
	sortOrder := "DESC"
	if filters.SortOrder == "asc" {
		sortOrder = "ASC"
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Get total count
	count, err := query.Count()
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	err = query.Select()
	return media, count, err
}

func (r *mediaRepository) UpdateProcessingStatus(ctx context.Context, id string, status entity.ProcessingStatus) error {
	_, err := r.db.ModelContext(ctx, (*entity.Media)(nil)).
		Set("processing_status = ?", status).
		Set("updated_at = NOW()").
		Where("id = ?", id).
		Update()
	return err
}

func (r *mediaRepository) GetPendingProcessing(ctx context.Context, limit int) ([]*entity.Media, error) {
	var media []*entity.Media
	err := r.db.ModelContext(ctx, &media).
		Where("processing_status = ?", entity.ProcessingStatusPending).
		Order("created_at ASC").
		Limit(limit).
		Select()
	return media, err
}
