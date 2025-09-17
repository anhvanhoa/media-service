package usecase

import (
	"context"
	"media-service/domain/entity"
	"media-service/domain/repository"

	"github.com/anhvanhoa/service-core/domain/goid"
	"github.com/anhvanhoa/service-core/domain/log"
	"github.com/anhvanhoa/service-core/domain/processing"
	"github.com/anhvanhoa/service-core/domain/storage"
)

type MediaUsecases struct {
	UploadUC       *UploadMediaUsecase
	UploadStreamUC *UploadMediaStreamUsecase
	GetUC          *GetMediaUsecase
	ListUC         *ListMediaUsecase
	UpdateUC       *UpdateMediaUsecase
	DeleteUC       *DeleteMediaUsecase
}

type MediaUsecaseInterfaces interface {
	UploadMedia(ctx context.Context, req *UploadMediaRequest) (*entity.Media, error)

	UploadMediaStream(ctx context.Context, req *UploadMediaStreamRequest) (*entity.Media, error)

	GetByID(ctx context.Context, id string) (*entity.Media, error)

	List(ctx context.Context, req *ListMediaRequest) (*ListMediaResponse, error)

	Update(ctx context.Context, id, createdBy string, req *UpdateMediaRequest) (*entity.Media, error)

	Delete(ctx context.Context, id, createdBy string) error
}

func NewMediaUsecases(
	mediaRepo repository.MediaRepository,
	logger *log.LogGRPCImpl,
	processing processing.ProcessingI,
	storage storage.StorageI,
) MediaUsecaseInterfaces {
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
			logger,
			storage,
		),
	}
}

// Implementation of MediaUsecaseInterfaces
func (m *MediaUsecases) UploadMedia(ctx context.Context, req *UploadMediaRequest) (*entity.Media, error) {
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
