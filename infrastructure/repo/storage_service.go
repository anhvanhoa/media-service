package repo

import (
	"context"
	"fmt"
	"io"
	"media-service/domain/usecase"
	"os"
	"path/filepath"

	"github.com/anhvanhoa/service-core/domain/log"
)

type LocalStorageService struct {
	uploadDir string
	publicURL string
	logger    *log.LogGRPCImpl
}

func NewLocalStorageService(uploadDir, publicURL string, logger *log.LogGRPCImpl) usecase.StorageService {
	return &LocalStorageService{
		uploadDir: uploadDir,
		publicURL: publicURL,
		logger:    logger,
	}
}

func (s *LocalStorageService) Upload(ctx context.Context, req *usecase.UploadRequest) (string, error) {
	// Ensure upload directory exists
	err := os.MkdirAll(s.uploadDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate file path
	ext := filepath.Ext(req.FileName)
	fileName := fmt.Sprintf("%s%s", req.ID, ext)
	filePath := filepath.Join(s.uploadDir, fileName)

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy data to file
	_, err = io.Copy(file, req.FileData)
	if err != nil {
		// Cleanup on error
		_ = os.Remove(filePath)
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Return public URL
	url := fmt.Sprintf("%s/%s", s.publicURL, fileName)
	return url, nil
}

func (s *LocalStorageService) Delete(ctx context.Context, url string) error {
	// Extract filename from URL
	fileName := filepath.Base(url)
	filePath := filepath.Join(s.uploadDir, fileName)

	// Delete file
	err := os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		s.logger.Warn(fmt.Sprintf("Failed to delete file: %s, error: %v", filePath, err))
		return err
	}

	return nil
}

func (s *LocalStorageService) GetDownloadURL(ctx context.Context, url string, expirationMinutes int) (string, error) {
	// For local storage, just return the URL as is
	// In production, you might want to implement signed URLs
	return url, nil
}
