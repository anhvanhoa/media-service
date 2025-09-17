package usecase

import (
	"io"
	"media-service/domain/entity"
)

// UploadMediaRequest represents an upload request
type UploadMediaRequest struct {
	ID        string
	FileName  string
	Size      int64
	Type      entity.MediaType
	FileData  io.Reader
	CreatedBy string
	Metadata  map[string]string
	Ext       string
}

// UpdateMediaRequest represents an update request
type UpdateMediaRequest struct {
	Name     *string
	Metadata map[string]string
}

type MediaMetadata struct {
	Width    int     // Image/Video width
	Height   int     // Image/Video height
	Duration float64 // Video/Audio duration in seconds
	Bitrate  int     // Video/Audio bitrate
	Format   string  // File format
}
