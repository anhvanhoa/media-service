package usecase

import (
	"context"
	"fmt"
	"media-service/domain/entity"
	"media-service/domain/repository"

	"github.com/anhvanhoa/service-core/domain/log"
	"github.com/anhvanhoa/service-core/domain/storage"
)

type DeleteMediaUsecase struct {
	mediaRepo      repository.MediaRepository
	logger         *log.LogGRPCImpl
	storageService storage.StorageI
}

func NewDeleteMediaUsecase(
	mediaRepo repository.MediaRepository,
	logger *log.LogGRPCImpl,
	storageService storage.StorageI,
) *DeleteMediaUsecase {
	return &DeleteMediaUsecase{
		mediaRepo:      mediaRepo,
		logger:         logger,
		storageService: storageService,
	}
}

func (uc *DeleteMediaUsecase) Execute(ctx context.Context, mediaID, createdBy string) error {
	uc.logger.Info(fmt.Sprintf("Deleting media: %s", mediaID))

	if err := uc.validateInput(mediaID, createdBy); err != nil {
		uc.logger.Error(fmt.Sprintf("Input validation failed: %v", err))
		return fmt.Errorf("validation failed: %w", err)
	}

	existingMedia, err := uc.retrieveExistingMedia(ctx, mediaID)
	if err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to retrieve existing media: %v", err))
		return fmt.Errorf("failed to retrieve media: %w", err)
	}

	if err := uc.checkOwnership(existingMedia, createdBy); err != nil {
		uc.logger.Error(fmt.Sprintf("Ownership check failed: %v", err))
		return fmt.Errorf("unauthorized: %w", err)
	}

	if err := uc.deleteFromStorage(ctx, existingMedia.URL); err != nil {
		uc.logger.Warn(fmt.Sprintf("Failed to delete file from storage: %v", err))
	}

	if err := uc.deleteFromDatabase(ctx, mediaID); err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to delete media from database: %v", err))
		return fmt.Errorf("failed to delete from database: %w", err)
	}

	uc.logger.Info(fmt.Sprintf("Media deleted successfully: %s", mediaID))

	return nil
}

func (uc *DeleteMediaUsecase) validateInput(mediaID, createdBy string) error {
	if mediaID == "" {
		return fmt.Errorf("media ID is required")
	}

	if createdBy == "" {
		return fmt.Errorf("created_by is required")
	}

	return nil
}

func (uc *DeleteMediaUsecase) retrieveExistingMedia(ctx context.Context, mediaID string) (*entity.Media, error) {
	media, err := uc.mediaRepo.GetByID(ctx, mediaID)
	if err != nil {
		return nil, err
	}
	if media == nil {
		return nil, fmt.Errorf("media not found")
	}
	return media, nil
}

func (uc *DeleteMediaUsecase) checkOwnership(media *entity.Media, createdBy string) error {
	if media.CreatedBy != createdBy {
		return fmt.Errorf("user %s does not own media %s", createdBy, media.ID)
	}
	return nil
}

func (uc *DeleteMediaUsecase) deleteFromStorage(ctx context.Context, url string) error {
	fmt.Println("Đang xóa file: ", url, " - ", "Chưa có xử lý")
	return uc.storageService.Delete(ctx, url)
}

func (uc *DeleteMediaUsecase) deleteFromDatabase(ctx context.Context, mediaID string) error {
	return uc.mediaRepo.Delete(ctx, mediaID)
}
