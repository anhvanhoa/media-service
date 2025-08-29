package usecase

import (
	"context"
	"fmt"
	"media-service/domain/entity"
	"media-service/domain/repository"

	"github.com/anhvanhoa/service-core/domain/log"
)

// GetMediaUsecase handles retrieving a single media by ID
type GetMediaUsecase struct {
	mediaRepo repository.MediaRepository
	logger    *log.LogGRPCImpl
}

// NewGetMediaUsecase creates a new get media usecase
func NewGetMediaUsecase(
	mediaRepo repository.MediaRepository,
	logger *log.LogGRPCImpl,
) *GetMediaUsecase {
	return &GetMediaUsecase{
		mediaRepo: mediaRepo,
		logger:    logger,
	}
}

// Execute retrieves a media by ID
func (uc *GetMediaUsecase) Execute(ctx context.Context, mediaID string) (*entity.Media, error) {
	uc.logger.Info(fmt.Sprintf("Getting media by ID: %s", mediaID))

	// Step 1: Validate input
	if err := uc.validateInput(mediaID); err != nil {
		uc.logger.Error(fmt.Sprintf("Input validation failed: %v", err))
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Step 2: Retrieve media from database
	media, err := uc.retrieveFromDatabase(ctx, mediaID)
	if err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to retrieve media from database: %v", err))
		return nil, fmt.Errorf("database retrieval failed: %w", err)
	}

	// Step 3: Check if media exists
	if media == nil {
		uc.logger.Warn(fmt.Sprintf("Media not found: %s", mediaID))
		return nil, fmt.Errorf("media not found")
	}

	// Step 4: Log successful retrieval
	uc.logger.Info(fmt.Sprintf("Media retrieved successfully: %s", mediaID))

	return media, nil
}

// Step 1: Validate input
func (uc *GetMediaUsecase) validateInput(mediaID string) error {
	if mediaID == "" {
		return fmt.Errorf("media ID is required")
	}

	// Basic UUID format validation
	if len(mediaID) != 36 {
		return fmt.Errorf("invalid media ID format")
	}

	return nil
}

// Step 2: Retrieve media from database
func (uc *GetMediaUsecase) retrieveFromDatabase(ctx context.Context, mediaID string) (*entity.Media, error) {
	return uc.mediaRepo.GetByID(ctx, mediaID)
}
