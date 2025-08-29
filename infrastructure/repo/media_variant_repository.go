package repo

import (
	"context"
	"media-service/domain/entity"
	"media-service/domain/repository"

	"github.com/go-pg/pg/v10"
)

type mediaVariantRepository struct {
	db *pg.DB
}

// NewMediaVariantRepository creates a new media variant repository
func NewMediaVariantRepository(db *pg.DB) repository.MediaVariantRepository {
	return &mediaVariantRepository{db: db}
}

func (r *mediaVariantRepository) Create(ctx context.Context, variant *entity.MediaVariant) error {
	_, err := r.db.ModelContext(ctx, variant).Insert()
	return err
}

func (r *mediaVariantRepository) GetByMediaID(ctx context.Context, mediaID string) ([]*entity.MediaVariant, error) {
	var variants []*entity.MediaVariant
	err := r.db.ModelContext(ctx, &variants).
		Where("media_id = ?", mediaID).
		Order("created_at ASC").
		Select()
	return variants, err
}

func (r *mediaVariantRepository) GetByMediaIDAndType(ctx context.Context, mediaID, variantType, size string) (*entity.MediaVariant, error) {
	variant := &entity.MediaVariant{}
	query := r.db.ModelContext(ctx, variant).
		Where("media_id = ?", mediaID).
		Where("type = ?", variantType)

	if size != "" {
		query = query.Where("size = ?", size)
	}

	err := query.Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return variant, nil
}

func (r *mediaVariantRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ModelContext(ctx, (*entity.MediaVariant)(nil)).Where("id = ?", id).Delete()
	return err
}

func (r *mediaVariantRepository) DeleteByMediaID(ctx context.Context, mediaID string) error {
	_, err := r.db.ModelContext(ctx, (*entity.MediaVariant)(nil)).Where("media_id = ?", mediaID).Delete()
	return err
}
