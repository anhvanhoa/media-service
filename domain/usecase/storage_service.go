package usecase

import (
	"context"
	"io"
)

// StorageService interface defines methods for file storage operations
type StorageService interface {
	// Upload uploads a file and returns the URL
	Upload(ctx context.Context, req *UploadRequest) (string, error)

	// Delete deletes a file by URL
	Delete(ctx context.Context, url string) error

	// GetDownloadURL generates a signed download URL (for private storage)
	GetDownloadURL(ctx context.Context, url string, expirationMinutes int) (string, error)
}

// UploadRequest represents a file upload request to storage
type UploadRequest struct {
	ID       string
	FileName string
	FileData io.Reader
	FileSize int64
	MimeType string
}
