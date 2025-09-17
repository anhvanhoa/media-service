package usecase

import (
	"context"
	"io"
	"media-service/domain/entity"
	"media-service/domain/repository"
	"time"

	"github.com/anhvanhoa/service-core/domain/goid"
	"github.com/anhvanhoa/service-core/domain/queue"
	"github.com/anhvanhoa/service-core/domain/storage"
	"github.com/anhvanhoa/service-core/utils"
)

type UploadMediaQueueRequest struct {
	ID        string
	FileName  string
	FileData  io.Reader
	Type      entity.MediaType
	CreatedBy string
}

type UploadMediaQueue interface {
	Execute(ctx context.Context, req *UploadMediaQueueRequest) error
}
type UploadMediaQueueUsecase struct {
	mediaRepo      repository.MediaRepository
	queue          queue.QueueClient
	storageService storage.StorageI
	uuid           goid.GoUUID
}

func NewUploadMediaQueueUsecase(
	mediaRepo repository.MediaRepository,
	queue queue.QueueClient,
	storageService storage.StorageI,
	uuid goid.GoUUID,
) UploadMediaQueue {
	return &UploadMediaQueueUsecase{
		mediaRepo:      mediaRepo,
		queue:          queue,
		storageService: storageService,
		uuid:           uuid,
	}
}

func (uc *UploadMediaQueueUsecase) Execute(ctx context.Context, req *UploadMediaQueueRequest) error {
	req.ID = uc.uuid.Gen()
	url, err := uc.storageService.Upload(ctx, &storage.UploadRequest{
		FileData:   req.FileData,
		OutputPath: utils.ConvertToSlug(req.ID + "_" + req.FileName),
	})
	if err != nil {
		return err
	}
	media := uc.createMediaEntity(req, url)
	if err := uc.saveToDatabase(ctx, media); err != nil {
		return err
	}
	return nil
}

func (uc *UploadMediaQueueUsecase) TaskQueue(ctx context.Context, req *UploadMediaQueueRequest) (string, error) {
	return uc.queue.EnqueueAnyTask(queue.NewPayloadMediaProcess(req.ID))
}

func (uc *UploadMediaQueueUsecase) createMediaEntity(
	req *UploadMediaQueueRequest,
	url string,
) *entity.Media {
	media := &entity.Media{
		ID:               req.ID,
		Name:             req.FileName,
		URL:              url,
		Type:             req.Type,
		ProcessingStatus: entity.ProcessingStatusPending,
		CreatedBy:        req.CreatedBy,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	return media
}

func (uc *UploadMediaQueueUsecase) saveToDatabase(ctx context.Context, media *entity.Media) error {
	return uc.mediaRepo.Create(ctx, media)
}
