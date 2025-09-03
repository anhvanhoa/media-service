package usecase

import (
	"context"
	"fmt"
	"media-service/domain/entity"
	"media-service/domain/repository"

	"github.com/anhvanhoa/service-core/domain/goid"
	"github.com/anhvanhoa/service-core/domain/log"
	"github.com/anhvanhoa/service-core/domain/processing"
	"github.com/anhvanhoa/service-core/domain/storage"
)

// MediaUsecases aggregates all media-related use cases
type MediaUsecases struct {
	UploadUC       *UploadMediaUsecase       // Legacy streaming upload
	UploadStreamUC *UploadMediaStreamUsecase // Improved streaming upload
	GetUC          *GetMediaUsecase
	ListUC         *ListMediaUsecase
	UpdateUC       *UpdateMediaUsecase
	DeleteUC       *DeleteMediaUsecase
	ProcessUC      *ProcessMediaUsecase
	GetVariantsUC  *GetMediaVariantsUsecase
}

// MediaUsecaseInterfaces defines interfaces for all media operations
type MediaUsecaseInterfaces interface {
	// Upload uploads a new media file (legacy streaming)
	UploadMedia(ctx context.Context, req *UploadMediaRequest) (*entity.Media, error)

	// UploadStream uploads a media file with streaming
	UploadMediaStream(ctx context.Context, req *UploadMediaStreamRequest) (*entity.Media, error)

	// GetByID retrieves a media by ID
	GetByID(ctx context.Context, id string) (*entity.Media, error)

	// List retrieves media with filters and pagination
	List(ctx context.Context, req *ListMediaRequest) (*ListMediaResponse, error)

	// Update updates media metadata
	Update(ctx context.Context, id, createdBy string, req *UpdateMediaRequest) (*entity.Media, error)

	// Delete deletes a media file
	Delete(ctx context.Context, id, createdBy string) error

	// GetVariants gets all variants of a media
	GetVariants(ctx context.Context, mediaID string) ([]*entity.MediaVariant, error)

	// ProcessMedia processes a media file (resize, convert, etc.)
	ProcessMedia(ctx context.Context, mediaID string) error
}

// NewMediaUsecases creates all media use cases
func NewMediaUsecases(
	mediaRepo repository.MediaRepository,
	variantRepo repository.MediaVariantRepository,
	logger *log.LogGRPCImpl,
	processing processing.ProcessingI,
	storage storage.StorageI,
) *MediaUsecases {
	goid := goid.NewGoId().UUID()
	return &MediaUsecases{
		UploadUC: NewUploadMediaUsecase(
			mediaRepo,
			logger,
			goid,
			processing,
			storage,
		),
		UploadStreamUC: NewUploadMediaStreamUsecase(
			mediaRepo,
			logger,
			goid,
			processing,
			storage,
		),
		GetUC: NewGetMediaUsecase(
			mediaRepo,
			logger,
		),
		ListUC: NewListMediaUsecase(
			mediaRepo,
			logger,
		),
		UpdateUC: NewUpdateMediaUsecase(
			mediaRepo,
			logger,
		),
		DeleteUC: NewDeleteMediaUsecase(
			mediaRepo,
			variantRepo,
			logger,
		),
		ProcessUC: NewProcessMediaUsecase(
			mediaRepo,
			logger,
		),
		GetVariantsUC: NewGetMediaVariantsUsecase(
			mediaRepo,
			variantRepo,
			logger,
		),
	}
}

// Implementation of MediaUsecaseInterfaces
func (m *MediaUsecases) UploadMedia(ctx context.Context, req *UploadMediaRequest) (*entity.Media, error) {
	fmt.Println("UploadMedia (legacy)", req)
	return m.UploadUC.Execute(ctx, req)
}

func (m *MediaUsecases) UploadMediaStream(ctx context.Context, req *UploadMediaStreamRequest) (*entity.Media, error) {
	return m.UploadStreamUC.Execute(ctx, req)
}

func (m *MediaUsecases) GetByID(ctx context.Context, id string) (*entity.Media, error) {
	return m.GetUC.Execute(ctx, id)
}

func (m *MediaUsecases) List(ctx context.Context, req *ListMediaRequest) (*ListMediaResponse, error) {
	return m.ListUC.Execute(ctx, req)
}

func (m *MediaUsecases) Update(ctx context.Context, id, createdBy string, req *UpdateMediaRequest) (*entity.Media, error) {
	return m.UpdateUC.Execute(ctx, id, createdBy, req)
}

func (m *MediaUsecases) Delete(ctx context.Context, id, createdBy string) error {
	return m.DeleteUC.Execute(ctx, id, createdBy)
}

func (m *MediaUsecases) GetVariants(ctx context.Context, mediaID string) ([]*entity.MediaVariant, error) {
	return m.GetVariantsUC.Execute(ctx, mediaID)
}

func (m *MediaUsecases) ProcessMedia(ctx context.Context, mediaID string) error {
	return m.ProcessUC.Execute(ctx, mediaID)
}
