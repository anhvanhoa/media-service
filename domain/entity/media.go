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

const (
	ExtWebP  = ".webp"
	ExtJPEG  = ".jpeg"
	ExtPNG   = ".png"
	ExtGIF   = ".gif"
	ExtMP4   = ".mp4"
	ExtAudio = ".mp3"
	ExtOther = ".other"
)

type ProcessingStatus string

const (
	ProcessingStatusPending    ProcessingStatus = "pending"
	ProcessingStatusProcessing ProcessingStatus = "processing"
	ProcessingStatusCompleted  ProcessingStatus = "completed"
	ProcessingStatusFailed     ProcessingStatus = "failed"
)

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
type UploadRequest struct {
	FileName  string            `json:"file_name"`
	FileSize  int64             `json:"file_size"`
	MimeType  string            `json:"mime_type"`
	CreatedBy string            `json:"created_by"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

func (Media) TableName() string {
	return "media"
}
