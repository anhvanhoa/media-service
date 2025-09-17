package usecase

import (
	"context"
	"fmt"
	"io"
	"media-service/domain/entity"
	"media-service/domain/repository"
	"os"
	"time"

	"github.com/anhvanhoa/service-core/domain/goid"
	"github.com/anhvanhoa/service-core/domain/log"
	"github.com/anhvanhoa/service-core/domain/processing"
	"github.com/anhvanhoa/service-core/domain/storage"
)

type UploadMediaStreamUsecase struct {
	mediaRepo      repository.MediaRepository
	logger         *log.LogGRPCImpl
	uuid           goid.GoUUID
	processing     processing.ProcessingI
	storageService storage.StorageI
}

func NewUploadMediaStreamUsecase(
	mediaRepo repository.MediaRepository,
	logger *log.LogGRPCImpl,
	uuid goid.GoUUID,
	processing processing.ProcessingI,
	storageService storage.StorageI,
) *UploadMediaStreamUsecase {
	return &UploadMediaStreamUsecase{
		mediaRepo:      mediaRepo,
		logger:         logger,
		uuid:           uuid,
		processing:     processing,
		storageService: storageService,
	}
}

type UploadMediaStreamRequest struct {
	ID        string
	FileName  string
	CreatedBy string
	Metadata  map[string]string
	FileData  io.Reader
	FileSize  int64
}

func (uc *UploadMediaStreamUsecase) Execute(ctx context.Context, req *UploadMediaStreamRequest) (*entity.Media, error) {
	uc.logger.Info(fmt.Sprintf("Starting streaming media upload: %s", req.FileName))

	tmpFile, err := uc.createTempFile()
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer uc.cleanupTempFile(tmpFile)

	bytesWritten, err := uc.bufferStreamToFile(req.FileData, tmpFile)
	if err != nil {
		return nil, fmt.Errorf("failed to buffer stream: %w", err)
	}
	uc.logger.Info(fmt.Sprintf("Buffered %d bytes from stream", bytesWritten))

	if req.FileSize > 0 && bytesWritten != req.FileSize {
		uc.logger.Warn(fmt.Sprintf("Expected %d bytes but received %d bytes", req.FileSize, bytesWritten))
	}

	if err := uc.resetFilePointer(tmpFile); err != nil {
		return nil, fmt.Errorf("failed to reset file pointer: %w", err)
	}

	url, err := uc.uploadToStorage(ctx, req, tmpFile)
	if err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to upload to storage: %v", err))
		return nil, fmt.Errorf("storage upload failed: %w", err)
	}
	uc.logger.Info(fmt.Sprintf("File uploaded to storage: %s", url))

	if err := uc.resetFilePointer(tmpFile); err != nil {
		return nil, fmt.Errorf("failed to reset file pointer: %w", err)
	}

	var (
		width, height int
		duration      float64
	)

	meta, err := uc.processing.ExtractImageMetadata(ctx, tmpFile)
	if err != nil {
		uc.logger.Warn(fmt.Sprintf("Could not extract metadata, using defaults: %v", err))
		return nil, fmt.Errorf("could not extract metadata: %w", err)
	} else {
		uc.logger.Info(fmt.Sprintf("Extracted metadata: %v", meta))
		width = meta.Width
		height = meta.Height
	}

	media := uc.createMediaEntity(
		req,
		url,
		string(entity.MimeTypeWebP),
		bytesWritten,
		width,
		height,
		duration,
	)

	if err := uc.saveToDatabase(ctx, media); err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to save to database: %v", err))
		_ = uc.storageService.Delete(ctx, url)
		return nil, fmt.Errorf("database save failed: %w", err)
	}

	uc.logger.Info(fmt.Sprintf("Streaming media upload completed successfully: %s", req.ID))
	return media, nil
}

func (uc *UploadMediaStreamUsecase) createTempFile() (*os.File, error) {
	tmpFile, err := os.CreateTemp("", "stream-upload-*")
	if err != nil {
		return nil, err
	}
	return tmpFile, nil
}

func (uc *UploadMediaStreamUsecase) cleanupTempFile(tmpFile *os.File) {
	fileName := tmpFile.Name()
	tmpFile.Close()
	if err := os.Remove(fileName); err != nil {
		uc.logger.Warn(fmt.Sprintf("Failed to cleanup temp file %s: %v", fileName, err))
	}
}

func (uc *UploadMediaStreamUsecase) bufferStreamToFile(reader io.Reader, tmpFile *os.File) (int64, error) {
	const bufferSize = 32 * 1024 // 32KB buffer
	buffer := make([]byte, bufferSize)
	var totalBytes int64
	for {
		bytesRead, err := reader.Read(buffer)
		if bytesRead > 0 {
			bytesWritten, writeErr := tmpFile.Write(buffer[:bytesRead])
			if writeErr != nil {
				return totalBytes, fmt.Errorf("failed to write to temp file: %w", writeErr)
			}
			totalBytes += int64(bytesWritten)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return totalBytes, fmt.Errorf("failed to read from stream: %w", err)
		}
	}

	return totalBytes, nil
}

func (uc *UploadMediaStreamUsecase) uploadToStorage(ctx context.Context, req *UploadMediaStreamRequest, tmpFile *os.File) (string, error) {
	outputFile := fmt.Sprintf("%s.webp", req.ID)
	uc.logger.Info(fmt.Sprintf("Converting and uploading to %s", outputFile))
	return uc.processing.ConvertWebPBufferToFile(ctx, tmpFile, outputFile)
}

func (uc *UploadMediaStreamUsecase) createMediaEntity(
	req *UploadMediaStreamRequest,
	url string,
	mimeType string,
	fileSize int64,
	width int,
	height int,
	duration float64,
) *entity.Media {
	media := &entity.Media{
		ID:               req.ID,
		Name:             req.FileName,
		Size:             fileSize,
		URL:              url,
		MimeType:         mimeType,
		Type:             entity.MediaTypeImage,
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

func (uc *UploadMediaStreamUsecase) saveToDatabase(ctx context.Context, media *entity.Media) error {
	return uc.mediaRepo.Create(ctx, media)
}

func (uc *UploadMediaStreamUsecase) resetFilePointer(tmpFile *os.File) error {
	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to reset file pointer: %w", err)
	}
	return nil
}
