package usecase

import (
	"context"
	"fmt"
	"media-service/domain/entity"
	"media-service/domain/repository"

	"github.com/anhvanhoa/service-core/domain/log"
)

// GetMediaVariantsUsecase handles retrieving all variants of a media
type GetMediaVariantsUsecase struct {
	mediaRepo   repository.MediaRepository
	variantRepo repository.MediaVariantRepository
	logger      *log.LogGRPCImpl
}

// NewGetMediaVariantsUsecase creates a new get media variants usecase
func NewGetMediaVariantsUsecase(
	mediaRepo repository.MediaRepository,
	variantRepo repository.MediaVariantRepository,
	logger *log.LogGRPCImpl,
) *GetMediaVariantsUsecase {
	return &GetMediaVariantsUsecase{
		mediaRepo:   mediaRepo,
		variantRepo: variantRepo,
		logger:      logger,
	}
}

// Execute retrieves all variants of a media
func (uc *GetMediaVariantsUsecase) Execute(ctx context.Context, mediaID string) ([]*entity.MediaVariant, error) {
	uc.logger.Info(fmt.Sprintf("Getting media variants: %s", mediaID))

	// Step 1: Validate input
	if err := uc.validateInput(mediaID); err != nil {
		uc.logger.Error(fmt.Sprintf("Input validation failed: %v", err))
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Step 2: Check if media exists
	if err := uc.checkMediaExists(ctx, mediaID); err != nil {
		uc.logger.Error(fmt.Sprintf("Media existence check failed: %v", err))
		return nil, fmt.Errorf("media check failed: %w", err)
	}

	// Step 3: Retrieve variants from database
	variants, err := uc.retrieveVariants(ctx, mediaID)
	if err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to retrieve variants: %v", err))
		return nil, fmt.Errorf("failed to retrieve variants: %w", err)
	}

	// Step 4: Log successful retrieval
	uc.logger.Info(fmt.Sprintf("Media variants retrieved successfully: %s", mediaID))

	return variants, nil
}

// Step 1: Validate input
func (uc *GetMediaVariantsUsecase) validateInput(mediaID string) error {
	if mediaID == "" {
		return fmt.Errorf("media ID is required")
	}
	return nil
}

// Step 2: Check if media exists
func (uc *GetMediaVariantsUsecase) checkMediaExists(ctx context.Context, mediaID string) error {
	media, err := uc.mediaRepo.GetByID(ctx, mediaID)
	if err != nil {
		return fmt.Errorf("failed to check media existence: %w", err)
	}
	if media == nil {
		return fmt.Errorf("media not found")
	}
	return nil
}

// Step 3: Retrieve variants from database
func (uc *GetMediaVariantsUsecase) retrieveVariants(ctx context.Context, mediaID string) ([]*entity.MediaVariant, error) {
	return uc.variantRepo.GetByMediaID(ctx, mediaID)
}
