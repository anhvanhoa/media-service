package constants

const (
	// Service name
	ServiceName = "media-service"

	// Error codes
	ErrCodeInvalidRequest    = "INVALID_REQUEST"
	ErrCodeUnauthorized      = "UNAUTHORIZED"
	ErrCodeNotFound          = "NOT_FOUND"
	ErrCodeInternalError     = "INTERNAL_ERROR"
	ErrCodeFileTooLarge      = "FILE_TOO_LARGE"
	ErrCodeUnsupportedFormat = "UNSUPPORTED_FORMAT"
	ErrCodeProcessingFailed  = "PROCESSING_FAILED"

	// Media processing
	MaxFileSize         = 100 * 1024 * 1024 // 100MB
	MaxImageWidth       = 2048
	MaxImageHeight      = 2048
	DefaultImageQuality = 85
	MaxVideoDuration    = 1800 // 30 minutes

	// Thumbnail sizes
	ThumbnailSmall  = "small"
	ThumbnailMedium = "medium"
	ThumbnailLarge  = "large"

	// Image formats
	FormatWebP = "webp"
	FormatJPEG = "jpeg"
	FormatPNG  = "png"

	// Video formats
	FormatMP4  = "mp4"
	FormatWebM = "webm"

	// Queue names
	QueueMediaProcessing = "media_processing"
	QueueMediaCleanup    = "media_cleanup"

	// Job types
	JobTypeImageResize     = "image_resize"
	JobTypeImageConvert    = "image_convert"
	JobTypeVideoTranscode  = "video_transcode"
	JobTypeCreateThumbnail = "create_thumbnail"
	JobTypeCleanupFiles    = "cleanup_files"
)
