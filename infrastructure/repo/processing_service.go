package repo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"media-service/domain/entity"
	"media-service/domain/repository"
	"media-service/domain/usecase"
	"os"
	"path/filepath"

	"github.com/anhvanhoa/service-core/domain/log"
	"github.com/anhvanhoa/service-core/domain/queue"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

// MediaProcessingService implements ProcessingService
type MediaProcessingService struct {
	variantRepo    repository.MediaVariantRepository
	storageService usecase.StorageService
	queueClient    queue.QueueClient
	logger         *log.LogGRPCImpl
}

// NewMediaProcessingService creates a new media processing service
func NewMediaProcessingService(
	variantRepo repository.MediaVariantRepository,
	storageService usecase.StorageService,
	queueClient queue.QueueClient,
	logger *log.LogGRPCImpl,
) usecase.ProcessingService {
	return &MediaProcessingService{
		variantRepo:    variantRepo,
		storageService: storageService,
		queueClient:    queueClient,
		logger:         logger,
	}
}

func (s *MediaProcessingService) ExtractMetadata(ctx context.Context, fileData io.Reader) (*usecase.MediaMetadata, error) {
	// Create temporary file to read metadata
	tempFile, err := os.CreateTemp("", "media_*")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Copy data to temp file
	_, err = io.Copy(tempFile, fileData)
	if err != nil {
		return nil, err
	}

	// Reset file pointer
	_, err = tempFile.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	metadata := &usecase.MediaMetadata{}

	// Try to decode image to get dimensions
	img, err := imaging.Open(tempFile.Name())
	if err == nil {
		bounds := img.Bounds()
		metadata.Width = bounds.Dx()
		metadata.Height = bounds.Dy()
	}

	return metadata, nil
}

func (s *MediaProcessingService) ProcessImage(ctx context.Context, media *entity.Media) error {
	s.logger.Info(fmt.Sprintf("Processing image: %s", media.ID))
	// Download original image
	// For now, assume it's stored locally and we can access it directly
	fileName := filepath.Base(media.URL)
	originalPath := filepath.Join("./uploads", fileName)

	// Check if file exists
	if _, err := os.Stat(originalPath); os.IsNotExist(err) {
		return fmt.Errorf("original image file not found: %s", originalPath)
	}

	// Create WebP version (actually JPEG for now)
	err := s.createWebPVariant(ctx, media, originalPath)
	if err != nil {
		s.logger.Warn(fmt.Sprintf("Failed to create converted variant: %v", err))
	}

	// Create thumbnails
	thumbnailSizes := map[string][]int{
		"small":  {150, 150},
		"medium": {300, 300},
		"large":  {600, 600},
	}

	for size, dimensions := range thumbnailSizes {
		err = s.createThumbnail(ctx, media, originalPath, size, dimensions[0], dimensions[1])
		if err != nil {
			s.logger.Warn(fmt.Sprintf("Failed to create thumbnail: %s, error: %v", size, err))
		}
	}

	return nil
}

func (s *MediaProcessingService) ProcessVideo(ctx context.Context, media *entity.Media) error {
	s.logger.Info(fmt.Sprintf("Processing video: %s", media.ID))

	// For video processing, you would typically use FFmpeg
	// This is a placeholder implementation

	// Create video thumbnail
	err := s.createVideoThumbnail(ctx, media)
	if err != nil {
		s.logger.Warn(fmt.Sprintf("Failed to create video thumbnail: %v", err))
	}

	// In a real implementation, you would:
	// 1. Extract video frame for thumbnail
	// 2. Transcode to different resolutions (480p, 720p, 1080p)
	// 3. Convert to different formats (MP4, WebM)
	// 4. Generate HLS streams for adaptive bitrate

	return nil
}

func (s *MediaProcessingService) QueueProcessing(ctx context.Context, mediaID string) error {
	_, err := s.queueClient.EnqueueAnyTask("media:process", queue.NewPayloadMediaProcess(mediaID))
	return err
}

func (s *MediaProcessingService) createWebPVariant(ctx context.Context, media *entity.Media, originalPath string) error {
	// Open original image
	img, err := imaging.Open(originalPath)
	if err != nil {
		return err
	}

	// Save as JPEG (WebP not directly supported by imaging library)
	// In production, you'd use a library that supports WebP
	tempFile, err := os.CreateTemp("", "webp_*.jpg")
	if err != nil {
		return err
	}
	defer tempFile.Close()
	defer os.Remove(tempFile.Name())

	err = imaging.Save(img, tempFile.Name())
	if err != nil {
		return err
	}

	// Read the converted file
	fileBytes, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return err
	}

	// Upload converted version
	variantID := uuid.New().String()
	url, err := s.storageService.Upload(ctx, &usecase.UploadRequest{
		ID:       variantID,
		FileName: fmt.Sprintf("%s.jpg", media.ID),
		FileData: bytes.NewReader(fileBytes),
		FileSize: int64(len(fileBytes)),
		MimeType: "image/jpeg",
	})
	if err != nil {
		return err
	}

	// Create variant record
	variant := &entity.MediaVariant{
		ID:       variantID,
		MediaID:  media.ID,
		Type:     "converted",
		Size:     "",
		URL:      url,
		FileSize: int64(len(fileBytes)),
		Width:    media.Width,
		Height:   media.Height,
		Format:   "jpeg",
	}

	return s.variantRepo.Create(ctx, variant)
}

func (s *MediaProcessingService) createThumbnail(ctx context.Context, media *entity.Media, originalPath string, size string, width, height int) error {
	// Open original image
	img, err := imaging.Open(originalPath)
	if err != nil {
		return err
	}

	// Resize image
	resizedImg := imaging.Resize(img, width, height, imaging.Lanczos)

	// Save resized image as JPEG
	tempFile, err := os.CreateTemp("", fmt.Sprintf("thumb_%s_*.jpg", size))
	if err != nil {
		return err
	}
	defer tempFile.Close()
	defer os.Remove(tempFile.Name())

	err = imaging.Save(resizedImg, tempFile.Name())
	if err != nil {
		return err
	}

	// Read the thumbnail file
	thumbBytes, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return err
	}

	// Upload thumbnail
	variantID := uuid.New().String()
	url, err := s.storageService.Upload(ctx, &usecase.UploadRequest{
		ID:       variantID,
		FileName: fmt.Sprintf("%s_thumb_%s.jpg", media.ID, size),
		FileData: bytes.NewReader(thumbBytes),
		FileSize: int64(len(thumbBytes)),
		MimeType: "image/jpeg",
	})
	if err != nil {
		return err
	}

	// Create variant record
	w, h := int32(width), int32(height)
	variant := &entity.MediaVariant{
		ID:       variantID,
		MediaID:  media.ID,
		Type:     "thumbnail",
		Size:     size,
		URL:      url,
		FileSize: int64(len(thumbBytes)),
		Width:    &w,
		Height:   &h,
		Format:   "jpeg",
	}

	return s.variantRepo.Create(ctx, variant)
}

func (s *MediaProcessingService) createVideoThumbnail(ctx context.Context, media *entity.Media) error {
	// This would typically use FFmpeg to extract a frame from the video
	// For now, this is a placeholder
	s.logger.Info(fmt.Sprintf("Creating video thumbnail: %s", media.ID))

	// In a real implementation:
	// 1. Use FFmpeg to extract frame at specific timestamp (e.g., 1 second)
	// 2. Resize the extracted frame to thumbnail sizes
	// 3. Convert to WebP format
	// 4. Upload and create variant records

	return nil
}
