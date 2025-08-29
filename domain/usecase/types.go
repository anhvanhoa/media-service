package usecase

import (
	"io"
)

// UploadMediaRequest represents an upload request
type UploadMediaRequest struct {
	FileName  string
	FileData  io.Reader
	FileSize  int64
	MimeType  string
	CreatedBy string
	Metadata  map[string]string
}

// UpdateMediaRequest represents an update request
type UpdateMediaRequest struct {
	Name     *string
	Metadata map[string]string
}
