package usecase

import (
	"context"
	"fmt"
	"io"
	"media-service/domain/entity"
	"media-service/domain/repository"
	"time"

	"github.com/anhvanhoa/service-core/domain/goid"
	"github.com/anhvanhoa/service-core/domain/log"
	"github.com/anhvanhoa/service-core/domain/processing"
	"github.com/anhvanhoa/service-core/domain/storage"
)

type UploadMediaUsecase struct {
	mediaRepo      repository.MediaRepository
	logger         *log.LogGRPCImpl
	uuid           goid.GoUUID
	processing     processing.ProcessingI
	storageService storage.StorageI
}

func NewUploadMediaUsecase(
	mediaRepo repository.MediaRepository,
	logger *log.LogGRPCImpl,
	uuid goid.GoUUID,
	processing processing.ProcessingI,
	storageService storage.StorageI,
) *UploadMediaUsecase {
	return &UploadMediaUsecase{
		mediaRepo:      mediaRepo,
		logger:         logger,
		uuid:           uuid,
		processing:     processing,
		storageService: storageService,
	}
}

func (uc *UploadMediaUsecase) Execute(ctx context.Context, req *UploadMediaRequest) (*entity.Media, error) {

	url, err := uc.uploadToStorage(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("storage upload failed: %w", err)
	}

	var (
		width, height int
		duration      float64
	)

	meta, err := uc.processing.ExtractImageMetadata(ctx, req.FileData)
	if err != nil {
		return nil, fmt.Errorf("could not extract metadata: %w", err)
	} else {
		width = meta.Width
		height = meta.Height
		duration = meta.Duration
	}

	media := uc.createMediaEntity(req, url, string(entity.MimeTypeWebP), width, height, duration)

	if err := uc.saveToDatabase(ctx, media); err != nil {
		_ = uc.storageService.Delete(ctx, url)
		return nil, fmt.Errorf("database save failed: %w", err)
	}

	uc.logger.Info(fmt.Sprintf("Media upload completed successfully: %s", req.ID))
	return media, nil
}

func (uc *UploadMediaUsecase) uploadToStorage(ctx context.Context, req *UploadMediaRequest) (string, error) {
	imageBytes, err := io.ReadAll(req.FileData)
	if err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to convert reader to bytes: %v", err))
		return "", err
	}

	uc.logger.Info(fmt.Sprintf("Read %d bytes from file", len(imageBytes)))
	return uc.processing.ConvertWebPBufferToFile(ctx, imageBytes, req.OutputFile)
}

func (uc *UploadMediaUsecase) createMediaEntity(
	req *UploadMediaRequest,
	url string,
	mimeType string,
	width int,
	height int,
	duration float64,
) *entity.Media {
	media := &entity.Media{
		ID:               req.ID,
		Name:             req.Name,
		Size:             req.Size,
		URL:              url,
		MimeType:         mimeType,
		Type:             req.Type,
		ProcessingStatus: entity.ProcessingStatusPending,
		CreatedBy:        req.CreatedBy,
		Width:            &width,
		Height:           &height,
		Duration:         &duration,
		Metadata:         req.Metadata,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	return media
}

func (uc *UploadMediaUsecase) saveToDatabase(ctx context.Context, media *entity.Media) error {
	return uc.mediaRepo.Create(ctx, media)
}
