package usecase

import (
	"context"
	"fmt"
	"media-service/domain/entity"
	"media-service/domain/repository"
	"time"

	"github.com/anhvanhoa/service-core/domain/log"
)

// UpdateMediaUsecase handles updating media metadata
type UpdateMediaUsecase struct {
	mediaRepo repository.MediaRepository
	logger    *log.LogGRPCImpl
}

// NewUpdateMediaUsecase creates a new update media usecase
func NewUpdateMediaUsecase(
	mediaRepo repository.MediaRepository,
	logger *log.LogGRPCImpl,
) *UpdateMediaUsecase {
	return &UpdateMediaUsecase{
		mediaRepo: mediaRepo,
		logger:    logger,
	}
}

// Execute updates media metadata
func (uc *UpdateMediaUsecase) Execute(ctx context.Context, mediaID, createdBy string, req *UpdateMediaRequest) (*entity.Media, error) {
	uc.logger.Info(fmt.Sprintf("Updating media: %s", mediaID))

	// Step 1: Validate input
	if err := uc.validateInput(mediaID, createdBy, req); err != nil {
		uc.logger.Error(fmt.Sprintf("Input validation failed: %v", err))
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Step 2: Retrieve existing media
	existingMedia, err := uc.retrieveExistingMedia(ctx, mediaID)
	if err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to retrieve existing media: %v", err))
		return nil, fmt.Errorf("failed to retrieve media: %w", err)
	}

	// Step 3: Check ownership
	if err := uc.checkOwnership(existingMedia, createdBy); err != nil {
		uc.logger.Error(fmt.Sprintf("Ownership check failed: %v", err))
		return nil, fmt.Errorf("unauthorized: %w", err)
	}

	// Step 4: Update media fields
	updatedMedia := uc.updateMediaFields(existingMedia, req)

	// Step 5: Save changes to database
	if err := uc.saveToDatabase(ctx, updatedMedia); err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to save updated media: %v", err))
		return nil, fmt.Errorf("failed to save changes: %w", err)
	}

	// Step 6: Log successful update
	uc.logger.Info(fmt.Sprintf("Media updated successfully: %s", mediaID))

	return updatedMedia, nil
}

// Step 1: Validate input
func (uc *UpdateMediaUsecase) validateInput(mediaID, createdBy string, req *UpdateMediaRequest) error {
	if mediaID == "" {
		return fmt.Errorf("media ID is required")
	}

	if createdBy == "" {
		return fmt.Errorf("created_by is required")
	}

	if req == nil {
		return fmt.Errorf("update request is required")
	}

	// Validate name if provided
	if req.Name != nil && *req.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	// Validate metadata keys
	if req.Metadata != nil {
		for key := range req.Metadata {
			if key == "" {
				return fmt.Errorf("metadata key cannot be empty")
			}
		}
	}

	return nil
}

// Step 2: Retrieve existing media
func (uc *UpdateMediaUsecase) retrieveExistingMedia(ctx context.Context, mediaID string) (*entity.Media, error) {
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
func (uc *UpdateMediaUsecase) checkOwnership(media *entity.Media, createdBy string) error {
	if media.CreatedBy != createdBy {
		return fmt.Errorf("user %s does not own media %s", createdBy, media.ID)
	}
	return nil
}

// Step 4: Update media fields
func (uc *UpdateMediaUsecase) updateMediaFields(media *entity.Media, req *UpdateMediaRequest) *entity.Media {
	// Create a copy to avoid modifying the original
	updatedMedia := *media

	// Update name if provided
	if req.Name != nil {
		updatedMedia.Name = *req.Name
	}

	// Update metadata if provided
	if req.Metadata != nil {
		// Merge with existing metadata
		if updatedMedia.Metadata == nil {
			updatedMedia.Metadata = make(map[string]string)
		}
		for key, value := range req.Metadata {
			if value == "" {
				// Remove metadata key if value is empty
				delete(updatedMedia.Metadata, key)
			} else {
				updatedMedia.Metadata[key] = value
			}
		}
	}

	// Update timestamp
	updatedMedia.UpdatedAt = time.Now()

	return &updatedMedia
}

// Step 5: Save changes to database
func (uc *UpdateMediaUsecase) saveToDatabase(ctx context.Context, media *entity.Media) error {
	return uc.mediaRepo.Update(ctx, media)
}
