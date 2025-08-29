package usecase

import (
	"context"
	"fmt"
	"io"
	"media-service/domain/entity"
	"media-service/domain/repository"
	"mime"
	"strings"
	"time"

	"github.com/anhvanhoa/service-core/domain/log"
	"github.com/google/uuid"
)

// UploadMediaUsecase handles media file upload
type UploadMediaUsecase struct {
	mediaRepo repository.MediaRepository
	logger    *log.LogGRPCImpl
}

// NewUploadMediaUsecase creates a new upload media usecase
func NewUploadMediaUsecase(
	mediaRepo repository.MediaRepository,
	logger *log.LogGRPCImpl,
) *UploadMediaUsecase {
	return &UploadMediaUsecase{
		mediaRepo: mediaRepo,
		logger:    logger,
	}
}

// Execute performs the upload media operation
func (uc *UploadMediaUsecase) Execute(ctx context.Context, req *UploadMediaRequest) (*entity.Media, error) {
	uc.logger.Info(fmt.Sprintf("Starting media upload: %s", req.FileName))

	// Step 1: Validate input
	if err := uc.validateInput(req); err != nil {
		uc.logger.Error(fmt.Sprintf("Input validation failed: %v", err))
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Step 2: Generate unique media ID
	mediaID := uuid.New().String()
	uc.logger.Debug(fmt.Sprintf("Generated media ID: %s", mediaID))

	// Step 3: Determine media type from MIME type
	mediaType := uc.getMediaTypeFromMime(req.MimeType)
	uc.logger.Debug(fmt.Sprintf("Determined media type: %s", string(mediaType)))

	// Step 4: Upload file to storage
	url, err := uc.uploadToStorage(ctx, mediaID, req)
	if err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to upload to storage: %v", err))
		return nil, fmt.Errorf("storage upload failed: %w", err)
	}
	uc.logger.Info(fmt.Sprintf("File uploaded to storage: %s", url))

	// Step 5: Extract metadata from file
	metadata, err := uc.extractMetadata(ctx, req.FileData)
	if err != nil {
		uc.logger.Warn(fmt.Sprintf("Failed to extract metadata: %v", err))
		// Continue without metadata - not critical
	}

	// Step 6: Create media entity
	media := uc.createMediaEntity(mediaID, req, url, mediaType, metadata)

	// Step 7: Save to database
	if err := uc.saveToDatabase(ctx, media); err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to save to database: %v", err))
		// Cleanup uploaded file
		// _ = uc.storageService.Delete(ctx, url)
		return nil, fmt.Errorf("database save failed: %w", err)
	}

	// Step 8: Queue for background processing if needed
	if err := uc.queueForProcessing(ctx, mediaID, mediaType, req.MimeType); err != nil {
		uc.logger.Warn(fmt.Sprintf("Failed to queue for processing: %v", err))
		// Not critical - can be processed manually later
	}

	uc.logger.Info(fmt.Sprintf("Media upload completed successfully: %s", mediaID))

	return media, nil
}

// Step 1: Validate input
func (uc *UploadMediaUsecase) validateInput(req *UploadMediaRequest) error {
	if req.FileName == "" {
		return fmt.Errorf("file name is required")
	}

	if req.MimeType == "" {
		return fmt.Errorf("mime type is required")
	}

	if req.CreatedBy == "" {
		return fmt.Errorf("created_by is required")
	}

	if req.FileSize <= 0 {
		return fmt.Errorf("file size must be greater than 0")
	}

	if req.FileSize > 100*1024*1024 { // 100MB limit
		return fmt.Errorf("file too large: %d bytes (max: 100MB)", req.FileSize)
	}

	if !uc.isSupportedMimeType(req.MimeType) {
		return fmt.Errorf("unsupported file type: %s", req.MimeType)
	}

	return nil
}

// Step 3: Determine media type from MIME type
func (uc *UploadMediaUsecase) getMediaTypeFromMime(mimeType string) entity.MediaType {
	mainType, _, _ := mime.ParseMediaType(mimeType)
	parts := strings.Split(mainType, "/")
	if len(parts) < 2 {
		return entity.MediaTypeOther
	}

	switch parts[0] {
	case "image":
		return entity.MediaTypeImage
	case "video":
		return entity.MediaTypeVideo
	case "audio":
		return entity.MediaTypeAudio
	default:
		return entity.MediaTypeOther
	}
}

// Step 4: Upload file to storage
func (uc *UploadMediaUsecase) uploadToStorage(ctx context.Context, mediaID string, req *UploadMediaRequest) (string, error) {
	// return uc.storageService.Upload(ctx, &UploadRequest{
	// 	ID:       mediaID,
	// 	FileName: req.FileName,
	// 	FileData: req.FileData,
	// 	FileSize: req.FileSize,
	// 	MimeType: req.MimeType,
	// })
	return "", nil
}

// Step 5: Extract metadata from file
func (uc *UploadMediaUsecase) extractMetadata(ctx context.Context, fileData io.Reader) (*MediaMetadata, error) {
	// return uc.processingService.ExtractMetadata(ctx, fileData)
	return nil, nil
}

// Step 6: Create media entity
func (uc *UploadMediaUsecase) createMediaEntity(
	mediaID string,
	req *UploadMediaRequest,
	url string,
	mediaType entity.MediaType,
	metadata *MediaMetadata,
) *entity.Media {
	media := &entity.Media{
		ID:               mediaID,
		CreatedBy:        req.CreatedBy,
		Name:             req.FileName,
		Size:             req.FileSize,
		URL:              url,
		MimeType:         req.MimeType,
		Type:             mediaType,
		ProcessingStatus: entity.ProcessingStatusPending,
		Metadata:         req.Metadata,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Add extracted metadata if available
	if metadata != nil {
		if metadata.Width > 0 {
			width := int32(metadata.Width)
			media.Width = &width
		}
		if metadata.Height > 0 {
			height := int32(metadata.Height)
			media.Height = &height
		}
		if metadata.Duration > 0 {
			duration := int32(metadata.Duration)
			media.Duration = &duration
		}
	}

	return media
}

// Step 7: Save to database
func (uc *UploadMediaUsecase) saveToDatabase(ctx context.Context, media *entity.Media) error {
	return uc.mediaRepo.Create(ctx, media)
}

// Step 8: Queue for background processing
func (uc *UploadMediaUsecase) queueForProcessing(ctx context.Context, mediaID string, mediaType entity.MediaType, mimeType string) error {
	if uc.needsProcessing(mediaType, mimeType) {
		// return uc.processingService.QueueProcessing(ctx, mediaID)
		return nil
	}

	// Mark as completed if no processing needed
	return uc.mediaRepo.UpdateProcessingStatus(ctx, mediaID, entity.ProcessingStatusCompleted)
}

// Helper methods
func (uc *UploadMediaUsecase) isSupportedMimeType(mimeType string) bool {
	supportedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"image/gif":       true,
		"image/webp":      true,
		"video/mp4":       true,
		"video/webm":      true,
		"video/quicktime": true,
		"video/x-msvideo": true,
	}
	return supportedTypes[mimeType]
}

func (uc *UploadMediaUsecase) needsProcessing(mediaType entity.MediaType, mimeType string) bool {
	// Images need processing for thumbnails and format conversion
	if mediaType == entity.MediaTypeImage {
		return true
	}
	// Videos need processing for different resolutions and formats
	if mediaType == entity.MediaTypeVideo {
		return true
	}
	return false
}
