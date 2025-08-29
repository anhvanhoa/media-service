package usecase

import (
	"context"
	"io"
	"media-service/domain/entity"
)

// ProcessingService interface defines methods for media processing operations
type ProcessingService interface {
	// ExtractMetadata extracts metadata from a media file
	ExtractMetadata(ctx context.Context, fileData io.Reader) (*MediaMetadata, error)

	// ProcessImage processes an image (resize, convert, create thumbnails)
	ProcessImage(ctx context.Context, media *entity.Media) error

	// ProcessVideo processes a video (transcode, create thumbnails)
	ProcessVideo(ctx context.Context, media *entity.Media) error

	// QueueProcessing adds a media to the processing queue
	QueueProcessing(ctx context.Context, mediaID string) error
}

// MediaMetadata represents extracted metadata from a media file
type MediaMetadata struct {
	Width    int     // Image/Video width
	Height   int     // Image/Video height
	Duration float64 // Video/Audio duration in seconds
	Bitrate  int     // Video/Audio bitrate
	Format   string  // File format
}
