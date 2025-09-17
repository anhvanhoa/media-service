package usecase

import (
	"context"
	"fmt"
	"media-service/domain/entity"
	"media-service/domain/repository"

	"github.com/anhvanhoa/service-core/domain/log"
)

// DeleteMediaUsecase handles deleting media files
type DeleteMediaUsecase struct {
	mediaRepo repository.MediaRepository
	logger    *log.LogGRPCImpl
}

// NewDeleteMediaUsecase creates a new delete media usecase
func NewDeleteMediaUsecase(
	mediaRepo repository.MediaRepository,
	logger *log.LogGRPCImpl,
) *DeleteMediaUsecase {
	return &DeleteMediaUsecase{
		mediaRepo: mediaRepo,
		logger:    logger,
	}
}

// Execute deletes a media file and all its variants
func (uc *DeleteMediaUsecase) Execute(ctx context.Context, mediaID, createdBy string) error {
	uc.logger.Info(fmt.Sprintf("Deleting media: %s", mediaID))

	// Step 1: Validate input
	if err := uc.validateInput(mediaID, createdBy); err != nil {
		uc.logger.Error(fmt.Sprintf("Input validation failed: %v", err))
		return fmt.Errorf("validation failed: %w", err)
	}

	// Step 2: Retrieve existing media
	existingMedia, err := uc.retrieveExistingMedia(ctx, mediaID)
	if err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to retrieve existing media: %v", err))
		return fmt.Errorf("failed to retrieve media: %w", err)
	}

	// Step 3: Check ownership
	if err := uc.checkOwnership(existingMedia, createdBy); err != nil {
		uc.logger.Error(fmt.Sprintf("Ownership check failed: %v", err))
		return fmt.Errorf("unauthorized: %w", err)
	}

	// Step 5: Delete main media file from storage
	if err := uc.deleteFromStorage(ctx, existingMedia.URL); err != nil {
		uc.logger.Warn(fmt.Sprintf("Failed to delete file from storage: %v", err))
		// Continue with database deletion even if storage cleanup fails
	}

	// Step 6: Delete media record from database
	if err := uc.deleteFromDatabase(ctx, mediaID); err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to delete media from database: %v", err))
		return fmt.Errorf("failed to delete from database: %w", err)
	}

	// Step 7: Log successful deletion
	uc.logger.Info(fmt.Sprintf("Media deleted successfully: %s", mediaID))

	return nil
}

// Step 1: Validate input
func (uc *DeleteMediaUsecase) validateInput(mediaID, createdBy string) error {
	if mediaID == "" {
		return fmt.Errorf("media ID is required")
	}

	if createdBy == "" {
		return fmt.Errorf("created_by is required")
	}

	return nil
}

// Step 2: Retrieve existing media
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

// Step 3: Check ownership
func (uc *DeleteMediaUsecase) checkOwnership(media *entity.Media, createdBy string) error {
	if media.CreatedBy != createdBy {
		return fmt.Errorf("user %s does not own media %s", createdBy, media.ID)
	}
	return nil
}

// Step 5: Delete main media file from storage
func (uc *DeleteMediaUsecase) deleteFromStorage(ctx context.Context, url string) error {
	fmt.Println("Đang xóa file: ", url, " - ", "Chưa có xử lý")
	return nil
	// return uc.storageService.Delete(ctx, url)
}

// Step 6: Delete media record from database
func (uc *DeleteMediaUsecase) deleteFromDatabase(ctx context.Context, mediaID string) error {
	return uc.mediaRepo.Delete(ctx, mediaID)
}
