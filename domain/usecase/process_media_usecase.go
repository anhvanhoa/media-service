package usecase

import (
	"context"
	"fmt"
	"media-service/domain/entity"
	"media-service/domain/repository"

	"github.com/anhvanhoa/service-core/domain/log"
)

// ProcessMediaUsecase handles processing media files (thumbnails, format conversion, etc.)
type ProcessMediaUsecase struct {
	mediaRepo repository.MediaRepository
	logger    *log.LogGRPCImpl
}

// NewProcessMediaUsecase creates a new process media usecase
func NewProcessMediaUsecase(
	mediaRepo repository.MediaRepository,
	logger *log.LogGRPCImpl,
) *ProcessMediaUsecase {
	return &ProcessMediaUsecase{
		mediaRepo: mediaRepo,
		logger:    logger,
	}
}

// Execute processes a media file (resize, convert, create thumbnails)
func (uc *ProcessMediaUsecase) Execute(ctx context.Context, mediaID string) error {
	uc.logger.Info(fmt.Sprintf("Starting media processing: %s", mediaID))

	// Step 1: Validate input
	if err := uc.validateInput(mediaID); err != nil {
		uc.logger.Error(fmt.Sprintf("Input validation failed: %v", err))
		return fmt.Errorf("validation failed: %w", err)
	}

	// Step 2: Retrieve media from database
	media, err := uc.retrieveMedia(ctx, mediaID)
	if err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to retrieve media: %v", err))
		return fmt.Errorf("failed to retrieve media: %w", err)
	}

	// Step 3: Check if media needs processing
	if !uc.needsProcessing(media) {
		uc.logger.Info(fmt.Sprintf("Media does not need processing: %s", mediaID))
		return uc.markAsCompleted(ctx, mediaID)
	}

	// Step 4: Update status to processing
	if err := uc.updateProcessingStatus(ctx, mediaID, entity.ProcessingStatusProcessing); err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to update status to processing: %v", err))
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Step 5: Process media based on type
	if err := uc.processMediaByType(ctx, media); err != nil {
		uc.logger.Error(fmt.Sprintf("Media processing failed: %v", err))

		// Update status to failed
		_ = uc.updateProcessingStatus(ctx, mediaID, entity.ProcessingStatusFailed)
		return fmt.Errorf("processing failed: %w", err)
	}

	// Step 6: Mark as completed
	if err := uc.markAsCompleted(ctx, mediaID); err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to mark as completed: %v", err))
		return fmt.Errorf("failed to mark as completed: %w", err)
	}

	// Step 7: Log successful completion
	uc.logger.Info(fmt.Sprintf("Media processing completed successfully: %s", mediaID))

	return nil
}

// Step 1: Validate input
func (uc *ProcessMediaUsecase) validateInput(mediaID string) error {
	if mediaID == "" {
		return fmt.Errorf("media ID is required")
	}
	return nil
}

// Step 2: Retrieve media from database
func (uc *ProcessMediaUsecase) retrieveMedia(ctx context.Context, mediaID string) (*entity.Media, error) {
	media, err := uc.mediaRepo.GetByID(ctx, mediaID)
	if err != nil {
		return nil, err
	}
	if media == nil {
		return nil, fmt.Errorf("media not found")
	}
	return media, nil
}

// Step 3: Check if media needs processing
func (uc *ProcessMediaUsecase) needsProcessing(media *entity.Media) bool {
	// Images need processing for thumbnails and format conversion
	if media.Type == entity.MediaTypeImage {
		return true
	}

	// Videos need processing for different resolutions and formats
	if media.Type == entity.MediaTypeVideo {
		return true
	}

	// Other types don't need processing currently
	return false
}

// Step 4: Update processing status
func (uc *ProcessMediaUsecase) updateProcessingStatus(ctx context.Context, mediaID string, status entity.ProcessingStatus) error {
	return uc.mediaRepo.UpdateProcessingStatus(ctx, mediaID, status)
}

// Step 5: Process media based on type
func (uc *ProcessMediaUsecase) processMediaByType(ctx context.Context, media *entity.Media) error {
	switch media.Type {
	case entity.MediaTypeImage:
		uc.logger.Debug(fmt.Sprintf("Processing image: %s", media.ID))
		fmt.Println("Đang chờ xử lý ảnh: Chưa có xử lý")
		return nil

	case entity.MediaTypeVideo:
		uc.logger.Debug(fmt.Sprintf("Processing video: %s", media.ID))
		fmt.Println("Đang chờ xử lý video: Chưa có xử lý")
		return nil

	default:
		uc.logger.Debug(fmt.Sprintf("No processing needed for media type: %s", media.ID))
		return nil
	}
}

// Step 6: Mark as completed
func (uc *ProcessMediaUsecase) markAsCompleted(ctx context.Context, mediaID string) error {
	return uc.updateProcessingStatus(ctx, mediaID, entity.ProcessingStatusCompleted)
}
