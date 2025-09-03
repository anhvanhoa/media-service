package usecase

import (
	"context"
	"fmt"
	"media-service/domain/entity"
	"media-service/domain/repository"
	"time"

	"github.com/anhvanhoa/service-core/domain/goid"
	"github.com/anhvanhoa/service-core/domain/log"
	"github.com/anhvanhoa/service-core/domain/processing"
	"github.com/anhvanhoa/service-core/domain/storage"
	"github.com/anhvanhoa/service-core/utils"
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
	file, err := uc.processing.CreateFileFromReader(req.FileData)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	defer uc.processing.DeleteFile(file.Name())
	req.FileData = file

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
		req.Ext = entity.ExtWebP
	}

	uc.processing.ResetReader(file)
	url, err := uc.uploadToStorage(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("storage upload failed: %w", err)
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
	if req.FileName == "" {
		req.FileName = req.ID
	}
	outFile := utils.ConvertToSlug(req.FileName) + req.Ext
	return uc.processing.ConvertWebPBufferToFile(ctx, req.FileData, outFile)
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
		Name:             req.FileName,
		Size:             req.Size,
		URL:              url,
		MimeType:         mimeType,
		Type:             req.Type,
		ProcessingStatus: entity.ProcessingStatusCompleted,
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
