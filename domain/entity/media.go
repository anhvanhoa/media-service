package entity

import (
	"time"
)

// MediaType represents the type of media
type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVideo MediaType = "video"
	MediaTypeAudio MediaType = "audio"
	MediaTypeOther MediaType = "other"
)

type MimeType string

const (
	MimeTypeImage MimeType = "image/jpeg"
	MimeTypeWebP  MimeType = "image/webp"
	MimeTypePNG   MimeType = "image/png"
	MimeTypeJPEG  MimeType = "image/jpeg"
	MimeTypeGIF   MimeType = "image/gif"
	MimeTypeVideo MimeType = "video/mp4"
	MimeTypeAudio MimeType = "audio/mpeg"
	MimeTypeOther MimeType = "application/octet-stream"
)

// ProcessingStatus represents the processing status of media
type ProcessingStatus string

const (
	ProcessingStatusPending    ProcessingStatus = "pending"
	ProcessingStatusProcessing ProcessingStatus = "processing"
	ProcessingStatusCompleted  ProcessingStatus = "completed"
	ProcessingStatusFailed     ProcessingStatus = "failed"
)

// Media represents a media file entity
type Media struct {
	ID               string            `json:"id" pg:"id,pk"`
	CreatedBy        string            `json:"created_by" pg:"created_by"`
	Name             string            `json:"name" pg:"name,notnull"`
	Size             int64             `json:"size" pg:"size"`
	URL              string            `json:"url" pg:"url"`
	MimeType         string            `json:"mime_type" pg:"mime_type"`
	Type             MediaType         `json:"type" pg:"type"`
	Width            *int              `json:"width,omitempty" pg:"width"`
	Height           *int              `json:"height,omitempty" pg:"height"`
	Duration         *float64          `json:"duration,omitempty" pg:"duration"` // For video/audio in seconds
	ProcessingStatus ProcessingStatus  `json:"processing_status" pg:"processing_status"`
	Metadata         map[string]string `json:"metadata,omitempty" pg:"metadata"`
	CreatedAt        time.Time         `json:"created_at" pg:"created_at,default:now()"`
	UpdatedAt        time.Time         `json:"updated_at" pg:"updated_at,default:now()"`
}

// MediaVariant represents different variants of a media file (thumbnails, different formats, etc.)
type MediaVariant struct {
	ID        string    `json:"id" pg:"id,pk"`
	MediaID   string    `json:"media_id" pg:"media_id,notnull"`
	Type      string    `json:"type" pg:"type"` // thumbnail, webp, mp4, etc.
	Size      string    `json:"size" pg:"size"` // small, medium, large for thumbnails
	URL       string    `json:"url" pg:"url"`
	FileSize  int64     `json:"file_size" pg:"file_size"`
	Width     *int32    `json:"width,omitempty" pg:"width"`
	Height    *int32    `json:"height,omitempty" pg:"height"`
	Quality   *int32    `json:"quality,omitempty" pg:"quality"`
	Format    string    `json:"format" pg:"format"` // webp, jpeg, mp4, webm
	CreatedAt time.Time `json:"created_at" pg:"created_at,default:now()"`
}

// UploadRequest represents a file upload request
type UploadRequest struct {
	FileName  string            `json:"file_name"`
	FileSize  int64             `json:"file_size"`
	MimeType  string            `json:"mime_type"`
	CreatedBy string            `json:"created_by"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// ProcessingJob represents a media processing job
type ProcessingJob struct {
	ID        string           `json:"id"`
	MediaID   string           `json:"media_id"`
	JobType   string           `json:"job_type"` // resize, convert, thumbnail, etc.
	Status    ProcessingStatus `json:"status"`
	Progress  int32            `json:"progress"` // 0-100
	Error     string           `json:"error,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// TableName returns the table name for Media
func (Media) TableName() string {
	return "media"
}

// TableName returns the table name for MediaVariant
func (MediaVariant) TableName() string {
	return "media_variants"
}
